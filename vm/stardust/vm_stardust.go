package stardust

import (
	"bytes"
	"fmt"

	"github.com/iotaledger/hive.go/core/safemath"
	"github.com/iotaledger/hive.go/ierrors"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/vm"
)

// NewVirtualMachine returns an VirtualMachine adhering to the Stardust protocol.
func NewVirtualMachine() vm.VirtualMachine {
	return &virtualMachine{
		execList: []vm.ExecFunc{
			vm.ExecFuncTimelocks(),
			vm.ExecFuncInputUnlocks(),
			vm.ExecFuncSenderUnlocked(),
			vm.ExecFuncBalancedBaseTokens(),
			vm.ExecFuncBalancedNativeTokens(),
			vm.ExecFuncChainTransitions(),
			vm.ExecFuncBalancedMana(),
			vm.ExecFuncAddressRestrictions(),
		},
	}
}

type virtualMachine struct {
	execList []vm.ExecFunc
}

func NewVMParamsWorkingSet(api iotago.API, t *iotago.SignedTransaction, inputs vm.ResolvedInputs) (*vm.WorkingSet, error) {
	var err error
	utxoInputsSet := constructInputSet(inputs.InputSet)
	workingSet := &vm.WorkingSet{}
	workingSet.Tx = t
	workingSet.UnlockedIdents = make(vm.UnlockedIdentities)
	workingSet.UTXOInputsSet = utxoInputsSet
	workingSet.InputIDToIndex = make(map[iotago.OutputID]uint16)
	for inputIndex, inputRef := range workingSet.Tx.Transaction.TransactionEssence.Inputs {
		//nolint:forcetypeassert // we can safely assume that this is an UTXOInput
		ref := inputRef.(*iotago.UTXOInput).OutputID()
		workingSet.InputIDToIndex[ref] = uint16(inputIndex)
		input, ok := workingSet.UTXOInputsSet[ref]
		if !ok {
			return nil, ierrors.Wrapf(iotago.ErrMissingUTXO, "utxo for input %d not supplied", inputIndex)
		}
		workingSet.UTXOInputs = append(workingSet.UTXOInputs, input)
	}

	workingSet.EssenceMsgToSign, err = t.Transaction.SigningMessage(api)
	if err != nil {
		return nil, err
	}

	txID, err := workingSet.Tx.ID()
	if err != nil {
		return nil, err
	}

	workingSet.InChains = utxoInputsSet.ChainInputSet()
	workingSet.OutChains = workingSet.Tx.Transaction.Outputs.ChainOutputSet(txID)

	workingSet.BIC = inputs.BlockIssuanceCreditInputSet
	workingSet.Commitment = inputs.CommitmentInput
	workingSet.Rewards = inputs.RewardsInputSet

	workingSet.TotalManaIn, err = vm.TotalManaIn(
		api.ProtocolParameters().ManaDecayProvider(),
		api.RentStructure(),
		workingSet.Tx.Transaction.CreationSlot,
		workingSet.UTXOInputsSet,
	)
	if err != nil {
		return nil, ierrors.Join(iotago.ErrManaAmountInvalid, err)
	}

	workingSet.TotalManaOut, err = vm.TotalManaOut(
		workingSet.Tx.Transaction.Outputs,
		workingSet.Tx.Transaction.Allotments,
	)
	if err != nil {
		return nil, ierrors.Join(iotago.ErrManaAmountInvalid, err)
	}

	return workingSet, nil
}

func constructInputSet(inputSet vm.InputSet) vm.InputSet {
	utxoInputsSet := vm.InputSet{}
	for outputID, outputWithCreationSlot := range inputSet {
		if basicOutput, isBasic := outputWithCreationSlot.(*iotago.BasicOutput); isBasic {
			if addressUnlock := basicOutput.UnlockConditionSet().Address(); addressUnlock != nil {
				if addressUnlock.Address.Type() == iotago.AddressImplicitAccountCreation {
					utxoInputsSet[outputID] = &vm.ImplicitAccountOutput{BasicOutput: basicOutput}

					continue
				}
			}
		}
		utxoInputsSet[outputID] = outputWithCreationSlot
	}

	return utxoInputsSet
}

func (stardustVM *virtualMachine) Execute(t *iotago.SignedTransaction, vmParams *vm.Params, inputs vm.ResolvedInputs, overrideFuncs ...vm.ExecFunc) error {
	if vmParams.API == nil {
		return ierrors.New("no API provided")
	}

	var err error
	vmParams.WorkingSet, err = NewVMParamsWorkingSet(vmParams.API, t, inputs)
	if err != nil {
		return err
	}

	if len(overrideFuncs) > 0 {
		return vm.RunVMFuncs(stardustVM, vmParams, overrideFuncs...)
	}

	return vm.RunVMFuncs(stardustVM, vmParams, stardustVM.execList...)
}

func (stardustVM *virtualMachine) ChainSTVF(transType iotago.ChainTransitionType, input *vm.ChainOutputWithIDs, next iotago.ChainOutput, vmParams *vm.Params) error {
	transitionState := next
	if transType != iotago.ChainTransitionTypeGenesis {
		transitionState = input.Output
	}

	var ok bool
	switch castedInput := transitionState.(type) {
	case *iotago.AccountOutput:
		var nextAccount *iotago.AccountOutput
		if next != nil {
			if nextAccount, ok = next.(*iotago.AccountOutput); !ok {
				return ierrors.New("can only state transition to another account output")
			}
		}

		return accountSTVF(input, transType, nextAccount, vmParams)
	case *vm.ImplicitAccountOutput:
		var nextAccount *iotago.AccountOutput
		if next != nil {
			if nextAccount, ok = next.(*iotago.AccountOutput); !ok {
				return ierrors.New("can only state transition implicit account to an account output")
			}
		}

		return implicitAccountSTVF(castedInput, input.OutputID, nextAccount, vmParams, transType)
	case *iotago.FoundryOutput:
		var nextFoundry *iotago.FoundryOutput
		if next != nil {
			if nextFoundry, ok = next.(*iotago.FoundryOutput); !ok {
				return ierrors.New("can only state transition to another foundry output")
			}
		}

		return foundrySTVF(input, transType, nextFoundry, vmParams)
	case *iotago.NFTOutput:
		var nextNFT *iotago.NFTOutput
		if next != nil {
			if nextNFT, ok = next.(*iotago.NFTOutput); !ok {
				return ierrors.New("can only state transition to another NFT output")
			}
		}

		return nftSTVF(input, transType, nextNFT, vmParams)
	case *iotago.DelegationOutput:
		var nextDelegationOutput *iotago.DelegationOutput
		if next != nil {
			if nextDelegationOutput, ok = next.(*iotago.DelegationOutput); !ok {
				return ierrors.New("can only state transition to another Delegation output")
			}
		}

		return delegationSTVF(input, transType, nextDelegationOutput, vmParams)
	default:
		panic(fmt.Sprintf("invalid output type %v passed to Stardust virtual machine", input.Output))
	}
}

// For implicit account conversion, there must be a basic output as input, and an account output as output with an AccountID matching the input.
func implicitAccountSTVF(implicitAccount *vm.ImplicitAccountOutput, outputID iotago.OutputID, next *iotago.AccountOutput, vmParams *vm.Params, transType iotago.ChainTransitionType) error {
	if transType == iotago.ChainTransitionTypeDestroy {
		return iotago.ErrImplicitAccountDestructionDisallowed
	}

	// Create a wrapper around the implicit account.
	implicitAccountChainOutput := &vm.ChainOutputWithIDs{
		ChainID:  next.AccountID,
		OutputID: outputID,
		Output:   implicitAccount,
	}

	implicitAccountBlockIssuerFeature := &iotago.BlockIssuerFeature{
		BlockIssuerKeys: iotago.NewBlockIssuerKeys(),
		// Setting MaxSlotIndex means one cannot remove the block issuer feature in the transition, but it does allow for setting
		// the expiry slot to a lower value, which is the behavior we want.
		ExpirySlot: iotago.MaxSlotIndex,
	}

	if err := accountBlockIssuerSTVF(implicitAccountChainOutput, implicitAccountBlockIssuerFeature, next, vmParams); err != nil {
		return err
	}

	return accountGenesisValid(next, vmParams, false)
}

// For output AccountOutput(s) with non-zeroed AccountID, there must be a corresponding input AccountOutput where either its
// AccountID is zeroed and StateIndex and FoundryCounter are zero or an input AccountOutput with the same AccountID.
//
// On account state transitions: The StateIndex must be incremented by 1 and Only Amount, NativeTokens, StateIndex, StateMetadata and FoundryCounter can be mutated.
//
// On account governance transition: Only StateController (must be mutated), GovernanceController and the MetadataBlock can be mutated.
func accountSTVF(input *vm.ChainOutputWithIDs, transType iotago.ChainTransitionType, next *iotago.AccountOutput, vmParams *vm.Params) error {
	switch transType {
	case iotago.ChainTransitionTypeGenesis:
		if err := accountGenesisValid(next, vmParams, true); err != nil {
			return ierrors.Wrapf(err, " account %s", next.AccountID)
		}
	case iotago.ChainTransitionTypeStateChange:
		if err := accountStateChangeValid(input, vmParams, next); err != nil {
			//nolint:forcetypeassert // we can safely assume that this is an AccountOutput
			a := input.Output.(*iotago.AccountOutput)

			return ierrors.Wrapf(err, "account %s", a.AccountID)
		}
	case iotago.ChainTransitionTypeDestroy:
		if err := accountDestructionValid(input, vmParams); err != nil {
			//nolint:forcetypeassert // we can safely assume that this is an AccountOutput
			a := input.Output.(*iotago.AccountOutput)

			return ierrors.Wrapf(err, "account %s", a.AccountID)
		}
	default:
		panic("unknown chain transition type in AccountOutput")
	}

	return nil
}

func accountGenesisValid(current *iotago.AccountOutput, vmParams *vm.Params, accountIDMustBeZeroed bool) error {
	if accountIDMustBeZeroed && !current.AccountID.Empty() {
		return ierrors.Wrap(iotago.ErrInvalidAccountStateTransition, "AccountOutput's ID is not zeroed even though it is new")
	}

	if nextBlockIssuerFeat := current.FeatureSet().BlockIssuer(); nextBlockIssuerFeat != nil {
		if vmParams.WorkingSet.Commitment == nil {
			return ierrors.Wrap(iotago.ErrInvalidBlockIssuerTransition, "block issuer feature validation requires a commitment input")
		}

		if nextBlockIssuerFeat.ExpirySlot < vmParams.PastBoundedSlotIndex(vmParams.WorkingSet.Commitment.Slot) {
			return ierrors.Wrap(iotago.ErrInvalidBlockIssuerTransition, "block issuer feature expiry set too soon")
		}
	}

	if stakingFeat := current.FeatureSet().Staking(); stakingFeat != nil {
		if err := accountStakingGenesisValidation(current, stakingFeat, vmParams); err != nil {
			return ierrors.Join(iotago.ErrInvalidStakingTransition, err)
		}
	}

	return vm.IsIssuerOnOutputUnlocked(current, vmParams.WorkingSet.UnlockedIdents)
}

func accountStateChangeValid(input *vm.ChainOutputWithIDs, vmParams *vm.Params, next *iotago.AccountOutput) error {
	//nolint:forcetypeassert // we can safely assume that this is an AccountOutput
	current := input.Output.(*iotago.AccountOutput)
	if !current.ImmutableFeatures.Equal(next.ImmutableFeatures) {
		return ierrors.Wrapf(iotago.ErrInvalidAccountStateTransition, "old state %s, next state %s", current.ImmutableFeatures, next.ImmutableFeatures)
	}

	// If a Block Issuer Feature is present on the input side of the transaction,
	// and the BIC is negative, the account is locked.
	if current.FeatureSet().BlockIssuer() != nil {
		if bic, exists := vmParams.WorkingSet.BIC[current.AccountID]; exists {
			if bic < 0 {
				return ierrors.Wrap(iotago.ErrAccountLocked, "negative block issuer credit")
			}
		} else {
			return iotago.ErrBlockIssuanceCreditInputRequired
		}
	}

	if current.StateIndex == next.StateIndex {
		return accountGovernanceSTVF(input, next, vmParams)
	}

	return accountStateSTVF(input, next, vmParams)
}

func accountGovernanceSTVF(input *vm.ChainOutputWithIDs, next *iotago.AccountOutput, vmParams *vm.Params) error {
	//nolint:forcetypeassert // we can safely assume that this is an AccountOutput
	current := input.Output.(*iotago.AccountOutput)

	switch {
	case current.Amount != next.Amount:
		return ierrors.Wrapf(iotago.ErrInvalidAccountGovernanceTransition, "amount changed, in %d / out %d ", current.Amount, next.Amount)
	case !current.NativeTokens.Equal(next.NativeTokens):
		return ierrors.Wrapf(iotago.ErrInvalidAccountGovernanceTransition, "native tokens changed, in %v / out %v", current.NativeTokens, next.NativeTokens)
	case current.StateIndex != next.StateIndex:
		return ierrors.Wrapf(iotago.ErrInvalidAccountGovernanceTransition, "state index changed, in %d / out %d", current.StateIndex, next.StateIndex)
	case !bytes.Equal(current.StateMetadata, next.StateMetadata):
		return ierrors.Wrapf(iotago.ErrInvalidAccountGovernanceTransition, "state metadata changed, in %v / out %v", current.StateMetadata, next.StateMetadata)
	case current.FoundryCounter != next.FoundryCounter:
		return ierrors.Wrapf(iotago.ErrInvalidAccountGovernanceTransition, "foundry counter changed, in %d / out %d", current.FoundryCounter, next.FoundryCounter)
	}

	// staking feature cannot change during account governance transition
	if err := iotago.FeatureUnchanged(iotago.FeatureStaking, current.Features.MustSet(), next.Features.MustSet()); err != nil {
		return ierrors.Join(iotago.ErrInvalidAccountGovernanceTransition, err)
	}

	if current.FeatureSet().Staking() != nil && next.FeatureSet().BlockIssuer() == nil {
		return ierrors.Wrapf(iotago.ErrInvalidAccountGovernanceTransition, "%w", iotago.ErrInvalidStakingBlockIssuerRequired)
	}

	return accountBlockIssuerSTVF(input, input.Output.FeatureSet().BlockIssuer(), next, vmParams)
}

func accountStateSTVF(input *vm.ChainOutputWithIDs, next *iotago.AccountOutput, vmParams *vm.Params) error {
	//nolint:forcetypeassert // we can safely assume that this is an AccountOutput
	current := input.Output.(*iotago.AccountOutput)
	switch {
	case !current.StateController().Equal(next.StateController()):
		return ierrors.Wrapf(iotago.ErrInvalidAccountStateTransition, "state controller changed, in %v / out %v", current.StateController(), next.StateController())
	case !current.GovernorAddress().Equal(next.GovernorAddress()):
		return ierrors.Wrapf(iotago.ErrInvalidAccountStateTransition, "governance controller changed, in %v / out %v", current.GovernorAddress(), next.GovernorAddress())
	case current.FoundryCounter > next.FoundryCounter:
		return ierrors.Wrapf(iotago.ErrInvalidAccountStateTransition, "foundry counter of next state is less than previous, in %d / out %d", current.FoundryCounter, next.FoundryCounter)
	case current.StateIndex+1 != next.StateIndex:
		return ierrors.Wrapf(iotago.ErrInvalidAccountStateTransition, "state index %d on the input side but %d on the output side", current.StateIndex, next.StateIndex)
	}

	if err := iotago.FeatureUnchanged(iotago.FeatureMetadata, current.Features.MustSet(), next.Features.MustSet()); err != nil {
		return ierrors.Wrapf(iotago.ErrInvalidAccountStateTransition, "%w", err)
	}

	// block issuer feature cannot change during account state transition
	if err := iotago.FeatureUnchanged(iotago.FeatureBlockIssuer, current.Features.MustSet(), next.Features.MustSet()); err != nil {
		return ierrors.Wrapf(iotago.ErrInvalidAccountStateTransition, "%w", err)
	}

	if err := accountStakingSTVF(input.ChainID, current, next, vmParams); err != nil {
		return err
	}

	// check that for a foundry counter change, X amount of foundries were actually created
	if current.FoundryCounter == next.FoundryCounter {
		return nil
	}

	var seenNewFoundriesOfAccount uint32
	for _, output := range vmParams.WorkingSet.Tx.Transaction.Outputs {
		foundryOutput, is := output.(*iotago.FoundryOutput)
		if !is {
			continue
		}

		if _, notNew := vmParams.WorkingSet.InChains[foundryOutput.MustFoundryID()]; notNew {
			continue
		}

		//nolint:forcetypeassert // we can safely assume that this is an AccountAddress
		foundryAccountID := foundryOutput.Ident().(*iotago.AccountAddress).ChainID()
		if !foundryAccountID.Matches(next.AccountID) {
			continue
		}
		seenNewFoundriesOfAccount++
	}

	expectedNewFoundriesCount := next.FoundryCounter - current.FoundryCounter
	if expectedNewFoundriesCount != seenNewFoundriesOfAccount {
		return ierrors.Wrapf(iotago.ErrInvalidAccountStateTransition, "%d new foundries were created but the account output's foundry counter changed by %d", seenNewFoundriesOfAccount, expectedNewFoundriesCount)
	}

	return nil
}

// If an account output has a block issuer feature, the following conditions for its transition must be checked.
// The block issuer credit must be non-negative.
// The expiry time of the block issuer feature, if creating new account or expired already, must be set at least MaxCommittableAge greater than the Commitment Input.
// Check that at least one Block Issuer Key is present.
func accountBlockIssuerSTVF(input *vm.ChainOutputWithIDs, currentBlockIssuerFeat *iotago.BlockIssuerFeature, next *iotago.AccountOutput, vmParams *vm.Params) error {
	current := input.Output
	nextBlockIssuerFeat := next.FeatureSet().BlockIssuer()
	// if the account has no block issuer feature.
	if currentBlockIssuerFeat == nil && nextBlockIssuerFeat == nil {
		return nil
	}

	// else if the account has negative bic, this is invalid.
	// new block issuers may not have a bic registered yet.
	if bic, exists := vmParams.WorkingSet.BIC[next.AccountID]; exists {
		if bic < 0 {
			return ierrors.Wrap(iotago.ErrInvalidBlockIssuerTransition, "negative block issuer credit")
		}
	} else {
		return ierrors.Wrap(iotago.ErrInvalidBlockIssuerTransition, "no BIC provided for block issuer")
	}

	if vmParams.WorkingSet.Commitment == nil {
		return ierrors.Wrap(iotago.ErrInvalidBlockIssuerTransition, "block issuer feature validation requires a commitment input")
	}

	commitmentInputSlot := vmParams.WorkingSet.Commitment.Slot
	pastBoundedSlot := vmParams.PastBoundedSlotIndex(commitmentInputSlot)

	if currentBlockIssuerFeat != nil && currentBlockIssuerFeat.ExpirySlot >= commitmentInputSlot {
		// if the block issuer feature has not expired, it can not be removed.
		if nextBlockIssuerFeat == nil {
			return ierrors.Wrap(iotago.ErrInvalidBlockIssuerTransition, "cannot remove block issuer feature until it expires")
		}
		if nextBlockIssuerFeat.ExpirySlot != currentBlockIssuerFeat.ExpirySlot && nextBlockIssuerFeat.ExpirySlot < pastBoundedSlot {
			return ierrors.Wrap(iotago.ErrInvalidBlockIssuerTransition, "block issuer feature expiry set too soon")
		}
	} else if nextBlockIssuerFeat != nil {
		// The block issuer feature was newly added,
		// or the current feature has expired but it was not removed.
		// In both cases the expiry slot must be set sufficiently far in the future.
		if nextBlockIssuerFeat.ExpirySlot < pastBoundedSlot {
			return ierrors.Wrapf(iotago.ErrInvalidBlockIssuerTransition, "block issuer feature expiry set too soon (is %d, must be >= %d)", nextBlockIssuerFeat.ExpirySlot, pastBoundedSlot)
		}
	}

	// the Mana on the account on the input side must not be moved to any other outputs or accounts.
	manaDecayProvider := vmParams.API.ProtocolParameters().ManaDecayProvider()
	rentStructure := vmParams.API.RentStructure()

	manaIn := vmParams.WorkingSet.TotalManaIn
	manaOut := vmParams.WorkingSet.TotalManaOut

	// AccountInStored
	manaStoredAccount, err := manaDecayProvider.ManaWithDecay(current.StoredMana(), input.OutputID.CreationSlot(), vmParams.WorkingSet.Tx.Transaction.CreationSlot)
	if err != nil {
		return ierrors.Wrapf(err, "account %s stored mana calculation failed", next.AccountID)
	}
	manaIn, err = safemath.SafeSub(manaIn, manaStoredAccount)
	if err != nil {
		return ierrors.Wrapf(err, "account %s stored mana in exceeds total remaining mana in", next.AccountID)
	}

	// AccountInPotential - the potential mana from the input side of the account in question
	// the storage deposit does not generate potential mana, so we only use the excess base tokens to calculate the potential mana
	minDeposit, err := rentStructure.MinDeposit(current)
	if err != nil {
		return err
	}
	excessBaseTokensAccount, err := safemath.SafeSub(current.BaseTokenAmount(), minDeposit)
	if err != nil {
		excessBaseTokensAccount = 0
	}
	manaPotentialAccount, err := manaDecayProvider.ManaGenerationWithDecay(excessBaseTokensAccount, input.OutputID.CreationSlot(), vmParams.WorkingSet.Tx.Transaction.CreationSlot)
	if err != nil {
		return ierrors.Wrapf(err, "account %s potential mana calculation failed", next.AccountID)
	}
	manaIn, err = safemath.SafeSub(manaIn, manaPotentialAccount)
	if err != nil {
		return ierrors.Wrapf(err, "account %s potential mana in exceeds total remaining mana in", next.AccountID)
	}
	// AccountOutStored - stored Mana on the output side of the account in question
	manaOut, err = safemath.SafeSub(manaOut, next.Mana)
	if err != nil {
		return ierrors.Wrapf(err, "account %s stored mana out exceeds total remaining mana out", next.AccountID)
	}
	// AccountOutAllotted - allotments to the account in question
	accountOutAllotted := vmParams.WorkingSet.Tx.Transaction.Allotments.Get(next.AccountID)
	manaOut, err = safemath.SafeSub(manaOut, accountOutAllotted)
	if err != nil {
		return ierrors.Wrapf(err, "account %s allotment exceeds total remaining mana out", next.AccountID)
	}

	// AccountOutLocked - outputs with manalock conditions
	minManalockedSlot := pastBoundedSlot + vmParams.API.ProtocolParameters().MaxCommittableAge()
	for _, output := range vmParams.WorkingSet.Tx.Transaction.Outputs {
		if output.UnlockConditionSet().HasManalockCondition(next.AccountID, minManalockedSlot) {
			manaOut, err = safemath.SafeSub(manaOut, output.StoredMana())
			if err != nil {
				return ierrors.Wrapf(err, "account %s manalocked output mana exceeds total remaining mana out", next.AccountID)
			}
		}
	}

	if manaIn < manaOut {
		return ierrors.Wrapf(iotago.ErrInvalidBlockIssuerTransition, "cannot move Mana off an account: mana in %d, mana out %d", manaIn, manaOut)
	}

	return nil
}

func accountStakingSTVF(chainID iotago.ChainID, current *iotago.AccountOutput, next *iotago.AccountOutput, vmParams *vm.Params) error {
	currentStakingFeat := current.FeatureSet().Staking()
	nextStakingFeat := next.FeatureSet().Staking()

	_, isClaiming := vmParams.WorkingSet.Rewards[chainID]

	if currentStakingFeat != nil {
		commitment := vmParams.WorkingSet.Commitment
		if commitment == nil {
			return ierrors.Join(iotago.ErrInvalidStakingTransition, iotago.ErrInvalidStakingCommitmentInput)
		}

		timeProvider := vmParams.API.TimeProvider()
		pastBoundedSlot := vmParams.PastBoundedSlotIndex(commitment.Slot)
		pastBoundedEpoch := timeProvider.EpochFromSlot(pastBoundedSlot)
		futureBoundedSlot := vmParams.FutureBoundedSlotIndex(commitment.Slot)
		futureBoundedEpoch := timeProvider.EpochFromSlot(futureBoundedSlot)

		if futureBoundedEpoch <= currentStakingFeat.EndEpoch {
			earliestUnbondingEpoch := pastBoundedEpoch + vmParams.API.ProtocolParameters().StakingUnbondingPeriod()

			return accountStakingNonExpiredValidation(
				currentStakingFeat, nextStakingFeat, earliestUnbondingEpoch, isClaiming,
			)
		}

		return accountStakingExpiredValidation(
			next, currentStakingFeat, nextStakingFeat, vmParams, isClaiming,
		)
	} else if nextStakingFeat != nil {
		return accountStakingGenesisValidation(next, nextStakingFeat, vmParams)
	}

	return nil
}

// Validates the rules for a newly added Staking Feature in an account,
// or one which was effectively removed and added within the same transaction.
// This is allowed as long as the epoch range of the old and new feature are disjoint.
func accountStakingGenesisValidation(acc *iotago.AccountOutput, stakingFeat *iotago.StakingFeature, vmParams *vm.Params) error {
	if acc.Amount < stakingFeat.StakedAmount {
		return iotago.ErrInvalidStakingAmountMismatch
	}

	// It should already never be nil here, but for 100% safety, we'll check again.
	commitment := vmParams.WorkingSet.Commitment
	if commitment == nil {
		return iotago.ErrInvalidStakingCommitmentInput
	}

	pastBoundedSlot := vmParams.PastBoundedSlotIndex(commitment.Slot)
	timeProvider := vmParams.API.TimeProvider()
	pastBoundedEpoch := timeProvider.EpochFromSlot(pastBoundedSlot)

	if stakingFeat.StartEpoch != pastBoundedEpoch {
		return iotago.ErrInvalidStakingStartEpoch
	}

	unbondingEpoch := pastBoundedEpoch + vmParams.API.ProtocolParameters().StakingUnbondingPeriod()
	if stakingFeat.EndEpoch < unbondingEpoch {
		return ierrors.Wrapf(iotago.ErrInvalidStakingEndEpochTooEarly, "(i.e. end epoch %d should be >= %d)", stakingFeat.EndEpoch, unbondingEpoch)
	}

	if acc.FeatureSet().BlockIssuer() == nil {
		return iotago.ErrInvalidStakingBlockIssuerRequired
	}

	return nil
}

// Validates a staking feature's transition if the feature is not expired,
// i.e. the current epoch is before the end epoch.
func accountStakingNonExpiredValidation(
	currentStakingFeat *iotago.StakingFeature,
	nextStakingFeat *iotago.StakingFeature,
	earliestUnbondingEpoch iotago.EpochIndex,
	isClaiming bool,
) error {
	if nextStakingFeat == nil {
		return ierrors.Wrapf(iotago.ErrInvalidStakingTransition, "%w", iotago.ErrInvalidStakingBondedRemoval)
	}

	if isClaiming {
		return ierrors.Wrapf(iotago.ErrInvalidStakingTransition, "%w", iotago.ErrInvalidStakingRewardClaim)
	}

	if currentStakingFeat.StakedAmount != nextStakingFeat.StakedAmount ||
		currentStakingFeat.FixedCost != nextStakingFeat.FixedCost ||
		currentStakingFeat.StartEpoch != nextStakingFeat.StartEpoch {
		return ierrors.Wrapf(iotago.ErrInvalidStakingTransition, "%w", iotago.ErrInvalidStakingBondedModified)
	}

	if currentStakingFeat.EndEpoch != nextStakingFeat.EndEpoch &&
		nextStakingFeat.EndEpoch < earliestUnbondingEpoch {
		return ierrors.Wrapf(iotago.ErrInvalidStakingTransition, "%w (i.e. end epoch %d should be >= %d) or the end epoch must match on input and output side", iotago.ErrInvalidStakingEndEpochTooEarly, nextStakingFeat.EndEpoch, earliestUnbondingEpoch)
	}

	return nil
}

// Validates a staking feature's transition if the feature is expired,
// i.e. the current epoch is equal or after the end epoch.
func accountStakingExpiredValidation(
	next *iotago.AccountOutput,
	currentStakingFeat *iotago.StakingFeature,
	nextStakingFeat *iotago.StakingFeature,
	vmParams *vm.Params,
	isClaiming bool,
) error {
	// Mana Claiming by either removing the Feature or changing the feature's epoch range.
	if nextStakingFeat == nil {
		if !isClaiming {
			return ierrors.Wrapf(iotago.ErrInvalidStakingTransition, "%w", iotago.ErrInvalidStakingRewardInputRequired)
		}
	} else {
		if isClaiming {
			// When claiming with a feature on the output side, it must be transitioned as if it was newly added,
			// so that the new epoch range is disjoint from the current staking feature.
			if err := accountStakingGenesisValidation(next, nextStakingFeat, vmParams); err != nil {
				return ierrors.Wrapf(iotago.ErrInvalidStakingTransition, "%w: rewards claiming without removing the feature requires updating the feature", err)
			}
		} else {
			// If not claiming, the feature must be unchanged.
			if !currentStakingFeat.Equal(nextStakingFeat) {
				return ierrors.Wrapf(iotago.ErrInvalidStakingTransition, "%w", iotago.ErrInvalidStakingRewardInputRequired)
			}
		}
	}

	return nil
}

func accountDestructionValid(input *vm.ChainOutputWithIDs, vmParams *vm.Params) error {
	//nolint:forcetypeassert // we can safely assume that this is an AccountOutput
	outputToDestroy := input.Output.(*iotago.AccountOutput)

	blockIssuerFeat := outputToDestroy.FeatureSet().BlockIssuer()
	if blockIssuerFeat != nil {
		if vmParams.WorkingSet.Commitment == nil {
			return ierrors.Wrap(iotago.ErrInvalidBlockIssuerTransition, "block issuer feature validation requires a commitment input")
		}

		if blockIssuerFeat.ExpirySlot >= vmParams.WorkingSet.Commitment.Slot {
			return ierrors.Wrap(iotago.ErrInvalidBlockIssuerTransition, "cannot destroy output until the block issuer feature expires")
		}
		if bic, exists := vmParams.WorkingSet.BIC[outputToDestroy.AccountID]; exists {
			if bic < 0 {
				return ierrors.Wrap(iotago.ErrAccountLocked, "cannot destroy locked account")
			}
		} else {
			return iotago.ErrBlockIssuanceCreditInputRequired
		}
	}

	stakingFeat := outputToDestroy.FeatureSet().Staking()
	if stakingFeat != nil {
		// This case should never occur as the staking feature requires the presence of a block issuer feature,
		// which also requires a commitment input.
		commitment := vmParams.WorkingSet.Commitment
		if commitment == nil {
			return ierrors.Join(iotago.ErrInvalidStakingTransition, iotago.ErrInvalidStakingCommitmentInput)
		}

		timeProvider := vmParams.API.TimeProvider()
		futureBoundedSlot := vmParams.FutureBoundedSlotIndex(commitment.Slot)
		futureBoundedEpoch := timeProvider.EpochFromSlot(futureBoundedSlot)

		if futureBoundedEpoch <= stakingFeat.EndEpoch {
			return ierrors.Wrapf(iotago.ErrInvalidAccountStateTransition, "%w: cannot destroy account until the staking feature is unbonded", iotago.ErrInvalidStakingBondedRemoval)
		}

		_, isClaiming := vmParams.WorkingSet.Rewards[input.ChainID]
		if !isClaiming {
			return ierrors.Wrapf(iotago.ErrInvalidAccountStateTransition, "%w: cannot destroy account with a staking feature without reward input", iotago.ErrInvalidStakingRewardInputRequired)
		}
	}

	return nil
}

func nftSTVF(input *vm.ChainOutputWithIDs, transType iotago.ChainTransitionType, next *iotago.NFTOutput, vmParams *vm.Params) error {
	switch transType {
	case iotago.ChainTransitionTypeGenesis:
		if err := nftGenesisValid(next, vmParams); err != nil {
			return &iotago.ChainTransitionError{Inner: err, Msg: fmt.Sprintf("NFT %s", next.NFTID)}
		}
	case iotago.ChainTransitionTypeStateChange:
		//nolint:forcetypeassert // we can safely assume that this is an NFTOutput
		current := input.Output.(*iotago.NFTOutput)
		if err := nftStateChangeValid(current, next); err != nil {
			return &iotago.ChainTransitionError{Inner: err, Msg: fmt.Sprintf("NFT %s", current.NFTID)}
		}
	case iotago.ChainTransitionTypeDestroy:
		return nil
	default:
		panic("unknown chain transition type in NFTOutput")
	}

	return nil
}

func nftGenesisValid(current *iotago.NFTOutput, vmParams *vm.Params) error {
	if !current.NFTID.Empty() {
		return ierrors.New("NFTOutput's ID is not zeroed even though it is new")
	}

	return vm.IsIssuerOnOutputUnlocked(current, vmParams.WorkingSet.UnlockedIdents)
}

func nftStateChangeValid(current *iotago.NFTOutput, next *iotago.NFTOutput) error {
	if !current.ImmutableFeatures.Equal(next.ImmutableFeatures) {
		return ierrors.Errorf("old state %s, next state %s", current.ImmutableFeatures, next.ImmutableFeatures)
	}

	return nil
}

func foundrySTVF(input *vm.ChainOutputWithIDs, transType iotago.ChainTransitionType, next *iotago.FoundryOutput, vmParams *vm.Params) error {
	inSums := vmParams.WorkingSet.InNativeTokens
	outSums := vmParams.WorkingSet.OutNativeTokens

	switch transType {
	case iotago.ChainTransitionTypeGenesis:
		if err := foundryGenesisValid(next, vmParams, next.MustFoundryID(), outSums); err != nil {
			return ierrors.Wrapf(err, "foundry %s, token %s", next.MustFoundryID(), next.MustNativeTokenID())
		}
	case iotago.ChainTransitionTypeStateChange:
		//nolint:forcetypeassert // we can safely assume that this is a FoundryOutput
		current := input.Output.(*iotago.FoundryOutput)
		if err := foundryStateChangeValid(current, next, inSums, outSums); err != nil {
			return ierrors.Wrapf(err, "foundry %s, token %s", current.MustFoundryID(), current.MustNativeTokenID())
		}
	case iotago.ChainTransitionTypeDestroy:
		//nolint:forcetypeassert // we can safely assume that this is a FoundryOutput
		current := input.Output.(*iotago.FoundryOutput)
		if err := foundryDestructionValid(current, inSums, outSums); err != nil {
			return ierrors.Wrapf(err, "foundry %s, token %s", current.MustFoundryID(), current.MustNativeTokenID())
		}
	default:
		panic("unknown chain transition type in FoundryOutput")
	}

	return nil
}

func foundryGenesisValid(current *iotago.FoundryOutput, vmParams *vm.Params, thisFoundryID iotago.FoundryID, outSums iotago.NativeTokenSum) error {
	nativeTokenID := current.MustNativeTokenID()
	if err := current.TokenScheme.StateTransition(iotago.ChainTransitionTypeGenesis, nil, nil, outSums.ValueOrBigInt0(nativeTokenID)); err != nil {
		return err
	}

	// grab foundry counter from transitioning AccountOutput
	//nolint:forcetypeassert // we can safely assume that this is an AccountAddress
	accountID := current.Ident().(*iotago.AccountAddress).AccountID()
	inAccount, ok := vmParams.WorkingSet.InChains[accountID]
	if !ok {
		return ierrors.Wrapf(iotago.ErrInvalidFoundryStateTransition, "missing input transitioning account output %s for new foundry output %s", accountID, thisFoundryID)
	}

	outAccount, ok := vmParams.WorkingSet.OutChains[accountID]
	if !ok {
		return ierrors.Wrapf(iotago.ErrInvalidFoundryStateTransition, "missing output transitioning account output %s for new foundry output %s", accountID, thisFoundryID)
	}

	//nolint:forcetypeassert // we can safely assume that this is an AccountOutput
	return foundrySerialNumberValid(current, vmParams, inAccount.Output.(*iotago.AccountOutput), outAccount.(*iotago.AccountOutput), thisFoundryID)
}

func foundrySerialNumberValid(current *iotago.FoundryOutput, vmParams *vm.Params, inAccount *iotago.AccountOutput, outAccount *iotago.AccountOutput, thisFoundryID iotago.FoundryID) error {
	// this new foundry's serial number must be between the given foundry counter interval
	startSerial := inAccount.FoundryCounter
	endIncSerial := outAccount.FoundryCounter
	if startSerial >= current.SerialNumber || current.SerialNumber > endIncSerial {
		return ierrors.Wrapf(iotago.ErrInvalidFoundryStateTransition, "new foundry output %s's serial number is not between the foundry counter interval of [%d,%d)", thisFoundryID, startSerial, endIncSerial)
	}

	// OPTIMIZE: this loop happens on every STVF of every new foundry output
	// check order of serial number
	for outputIndex, output := range vmParams.WorkingSet.Tx.Transaction.Outputs {
		otherFoundryOutput, is := output.(*iotago.FoundryOutput)
		if !is {
			continue
		}

		if !otherFoundryOutput.Ident().Equal(current.Ident()) {
			continue
		}

		otherFoundryID, err := otherFoundryOutput.FoundryID()
		if err != nil {
			return err
		}

		if _, isNotNew := vmParams.WorkingSet.InChains[otherFoundryID]; isNotNew {
			continue
		}

		// only check up to own foundry whether it is ordered
		if otherFoundryID == thisFoundryID {
			break
		}

		if otherFoundryOutput.SerialNumber >= current.SerialNumber {
			return ierrors.Wrapf(iotago.ErrInvalidFoundryStateTransition, "new foundry output %s at index %d has bigger equal serial number than this foundry %s", otherFoundryID, outputIndex, thisFoundryID)
		}
	}

	return nil
}

func foundryStateChangeValid(current *iotago.FoundryOutput, next *iotago.FoundryOutput, inSums iotago.NativeTokenSum, outSums iotago.NativeTokenSum) error {
	if !current.ImmutableFeatures.Equal(next.ImmutableFeatures) {
		return ierrors.Wrapf(iotago.ErrInvalidFoundryStateTransition, "old state %s, next state %s", current.ImmutableFeatures, next.ImmutableFeatures)
	}

	// the check for the serial number and token scheme not being mutated is implicit
	// as a change would cause the foundry ID to be different, which would result in
	// no matching foundry to be found to validate the state transition against
	if current.MustFoundryID() != next.MustFoundryID() {
		// impossible invariant as the STVF should be called via the matching next foundry output
		panic(fmt.Sprintf("foundry IDs mismatch in state transition validation function: have %v got %v", current.MustFoundryID(), next.MustFoundryID()))
	}

	nativeTokenID := current.MustNativeTokenID()

	return current.TokenScheme.StateTransition(iotago.ChainTransitionTypeStateChange, next.TokenScheme, inSums.ValueOrBigInt0(nativeTokenID), outSums.ValueOrBigInt0(nativeTokenID))
}

func foundryDestructionValid(current *iotago.FoundryOutput, inSums iotago.NativeTokenSum, outSums iotago.NativeTokenSum) error {
	nativeTokenID := current.MustNativeTokenID()

	return current.TokenScheme.StateTransition(iotago.ChainTransitionTypeDestroy, nil, inSums.ValueOrBigInt0(nativeTokenID), outSums.ValueOrBigInt0(nativeTokenID))
}

func delegationSTVF(input *vm.ChainOutputWithIDs, transType iotago.ChainTransitionType, next *iotago.DelegationOutput, vmParams *vm.Params) error {
	switch transType {
	case iotago.ChainTransitionTypeGenesis:
		if err := delegationGenesisValid(next, vmParams); err != nil {
			return &iotago.ChainTransitionError{Inner: err, Msg: fmt.Sprintf("Delegation %s", next.DelegationID)}
		}
	case iotago.ChainTransitionTypeStateChange:
		_, isClaiming := vmParams.WorkingSet.Rewards[input.ChainID]
		if isClaiming {
			return ierrors.Wrapf(iotago.ErrInvalidDelegationTransition, "%w: cannot claim rewards during delegation output transition", iotago.ErrInvalidDelegationRewardsClaiming)
		}
		//nolint:forcetypeassert // we can safely assume that this is an DelegationOutput
		current := input.Output.(*iotago.DelegationOutput)
		if err := delegationStateChangeValid(current, next, vmParams); err != nil {
			return &iotago.ChainTransitionError{Inner: err, Msg: fmt.Sprintf("Delegation %s", current.DelegationID)}
		}
	case iotago.ChainTransitionTypeDestroy:
		_, isClaiming := vmParams.WorkingSet.Rewards[input.ChainID]
		if !isClaiming {
			return ierrors.Wrapf(iotago.ErrInvalidDelegationTransition, "%w: cannot destroy delegation output without a rewards input", iotago.ErrInvalidDelegationRewardsClaiming)
		}

		return nil
	default:
		panic("unknown chain transition type in DelegationOutput")
	}

	return nil
}

func delegationGenesisValid(current *iotago.DelegationOutput, vmParams *vm.Params) error {
	if !current.DelegationID.Empty() {
		return ierrors.Wrapf(iotago.ErrInvalidDelegationNonZeroedID, "%w", iotago.ErrInvalidDelegationTransition)
	}

	timeProvider := vmParams.API.TimeProvider()
	commitment := vmParams.WorkingSet.Commitment
	if commitment == nil {
		return iotago.ErrDelegationCommitmentInputRequired
	}
	pastBoundedSlot := vmParams.PastBoundedSlotIndex(commitment.Slot)
	pastBoundedEpoch := timeProvider.EpochFromSlot(pastBoundedSlot)
	votingPowerSlot := votingPowerSlot(pastBoundedEpoch, vmParams)

	var expectedStartEpoch iotago.EpochIndex
	if pastBoundedSlot <= votingPowerSlot {
		expectedStartEpoch = pastBoundedEpoch + 1
	} else {
		expectedStartEpoch = pastBoundedEpoch + 2
	}

	if current.StartEpoch != expectedStartEpoch {
		return ierrors.Wrapf(iotago.ErrInvalidDelegationTransition, "%w (is %d, expected %d)", iotago.ErrInvalidDelegationStartEpoch, current.StartEpoch, expectedStartEpoch)
	}

	if current.DelegatedAmount != current.Amount {
		return ierrors.Wrapf(iotago.ErrInvalidDelegationTransition, "%w", iotago.ErrInvalidDelegationAmount)
	}

	if current.EndEpoch != 0 {
		return ierrors.Wrapf(iotago.ErrInvalidDelegationTransition, "%w", iotago.ErrInvalidDelegationNonZeroEndEpoch)
	}

	return nil
}

func delegationStateChangeValid(current *iotago.DelegationOutput, next *iotago.DelegationOutput, vmParams *vm.Params) error {
	// State transitioning a Delegation Output is always a transition to the delayed claiming state.
	// Since they can only be transitioned once, the input will always need to have a zeroed ID.
	if !current.DelegationID.Empty() {
		return ierrors.Wrapf(iotago.ErrInvalidDelegationTransition, "%w: delegation output can only be transitioned if it has a zeroed ID", iotago.ErrInvalidDelegationNonZeroedID)
	}

	if current.DelegatedAmount != next.DelegatedAmount ||
		!current.ValidatorAddress.Equal(next.ValidatorAddress) ||
		current.StartEpoch != next.StartEpoch {
		return ierrors.Wrapf(iotago.ErrInvalidDelegationTransition, "%w", iotago.ErrInvalidDelegationModified)
	}

	timeProvider := vmParams.API.TimeProvider()
	commitment := vmParams.WorkingSet.Commitment
	if commitment == nil {
		return iotago.ErrDelegationCommitmentInputRequired
	}
	futureBoundedSlot := vmParams.FutureBoundedSlotIndex(commitment.Slot)
	futureBoundedEpoch := timeProvider.EpochFromSlot(futureBoundedSlot)
	votingPowerSlot := votingPowerSlot(futureBoundedEpoch, vmParams)

	var expectedEndEpoch iotago.EpochIndex
	if futureBoundedSlot <= votingPowerSlot {
		expectedEndEpoch = futureBoundedEpoch
	} else {
		expectedEndEpoch = futureBoundedEpoch + 1
	}

	if next.EndEpoch != expectedEndEpoch {
		return ierrors.Wrapf(iotago.ErrInvalidDelegationTransition, "%w (is %d, expected %d)", iotago.ErrInvalidDelegationEndEpoch, next.EndEpoch, expectedEndEpoch)
	}

	return nil
}

// votingPowerSlot returns the slot at the end of which the voting power
// for the epoch with index epochIndex is calculated.
func votingPowerSlot(epoch iotago.EpochIndex, vmParams *vm.Params) iotago.SlotIndex {
	// TODO: Activity Window Duration missing.
	return vmParams.API.TimeProvider().EpochEnd(epoch) - vmParams.API.ProtocolParameters().EpochNearingThreshold()
}
