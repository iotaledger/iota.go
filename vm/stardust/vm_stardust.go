package stardust

import (
	"bytes"
	"fmt"

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
			vm.ExecFuncBalancedDeposit(),
			vm.ExecFuncBalancedNativeTokens(),
			vm.ExecFuncChainTransitions(),
			vm.ExecFuncBalancedMana(),
		},
	}
}

type virtualMachine struct {
	execList []vm.ExecFunc
}

func (stardustVM *virtualMachine) Execute(t *iotago.Transaction, vmParams *vm.Params, inputs vm.ResolvedInputs, overrideFuncs ...vm.ExecFunc) error {
	var err error
	vmParams.WorkingSet, err = vm.NewVMParamsWorkingSet(t, inputs)
	if err != nil {
		return err
	}

	if len(overrideFuncs) > 0 {
		return vm.RunVMFuncs(stardustVM, vmParams, overrideFuncs...)
	}

	return vm.RunVMFuncs(stardustVM, vmParams, stardustVM.execList...)
}

func (stardustVM *virtualMachine) ChainSTVF(transType iotago.ChainTransitionType, input *vm.ChainOutputWithCreationTime, next iotago.ChainOutput, vmParams *vm.Params) error {
	transitionState := next
	if transType != iotago.ChainTransitionTypeGenesis {
		transitionState = input.Output
	}

	var ok bool
	switch transitionState.(type) {
	case *iotago.AccountOutput:
		var nextAccount *iotago.AccountOutput
		if next != nil {
			if nextAccount, ok = next.(*iotago.AccountOutput); !ok {
				return fmt.Errorf("can only state transition to another account output")
			}
		}

		return accountSTVF(input, transType, nextAccount, vmParams)
	case *iotago.FoundryOutput:
		var nextFoundry *iotago.FoundryOutput
		if next != nil {
			if nextFoundry, ok = next.(*iotago.FoundryOutput); !ok {
				return fmt.Errorf("can only state transition to another foundry output")
			}
		}

		return foundrySTVF(input, transType, nextFoundry, vmParams)
	case *iotago.NFTOutput:
		var nextNFT *iotago.NFTOutput
		if next != nil {
			if nextNFT, ok = next.(*iotago.NFTOutput); !ok {
				return fmt.Errorf("can only state transition to another NFT output")
			}
		}

		return nftSTVF(input, transType, nextNFT, vmParams)
	case *iotago.DelegationOutput:
		var nextDelegationOutput *iotago.DelegationOutput
		if next != nil {
			if nextDelegationOutput, ok = next.(*iotago.DelegationOutput); !ok {
				return fmt.Errorf("can only state transition to another Delegation output")
			}
		}

		return delegationSTVF(input, transType, nextDelegationOutput, vmParams)
	default:
		panic(fmt.Sprintf("invalid output type %v passed to Stardust virtual machine", input.Output))
	}
}

// For output AccountOutput(s) with non-zeroed AccountID, there must be a corresponding input AccountOutput where either its
// AccountID is zeroed and StateIndex and FoundryCounter are zero or an input AccountOutput with the same AccountID.
//
// On account state transitions: The StateIndex must be incremented by 1 and Only Amount, NativeTokens, StateIndex, StateMetadata and FoundryCounter can be mutated.
//
// On account governance transition: Only StateController (must be mutated), GovernanceController and the MetadataBlock can be mutated.
func accountSTVF(input *vm.ChainOutputWithCreationTime, transType iotago.ChainTransitionType, next *iotago.AccountOutput, vmParams *vm.Params) error {
	switch transType {
	case iotago.ChainTransitionTypeGenesis:
		if err := accountGenesisValid(next, vmParams); err != nil {
			return fmt.Errorf("%w:  account %s", err, next.AccountID)
		}
	case iotago.ChainTransitionTypeStateChange:
		if err := accountStateChangeValid(input, vmParams, next); err != nil {
			a := input.Output.(*iotago.AccountOutput)

			return fmt.Errorf("%w: account %s", err, a.AccountID)
		}
	case iotago.ChainTransitionTypeDestroy:
		if err := accountDestructionValid(input, vmParams); err != nil {
			a := input.Output.(*iotago.AccountOutput)

			return fmt.Errorf("%w: account %s", err, a.AccountID)
		}
	default:
		panic("unknown chain transition type in AccountOutput")
	}

	return nil
}

func accountGenesisValid(current *iotago.AccountOutput, vmParams *vm.Params) error {
	if !current.AccountID.Empty() {
		return fmt.Errorf("%w: AccountOutput's ID is not zeroed even though it is new", iotago.ErrInvalidAccountStateTransition)
	}

	if nextBIFeat := current.FeatureSet().BlockIssuer(); nextBIFeat != nil && nextBIFeat.ExpirySlot != 0 && nextBIFeat.ExpirySlot < vmParams.WorkingSet.Tx.Essence.CreationTime+vmParams.External.ProtocolParameters.MaxCommittableAge {
		return fmt.Errorf("%w: block issuer feature expiry set too soon", iotago.ErrInvalidBlockIssuerTransition)
	}

	if stakingFeat := current.FeatureSet().Staking(); stakingFeat != nil {
		if err := accountStakingGenesisValidation(current, stakingFeat, vmParams); err != nil {
			return err
		}
	}

	return vm.IsIssuerOnOutputUnlocked(current, vmParams.WorkingSet.UnlockedIdents)
}

func accountStateChangeValid(input *vm.ChainOutputWithCreationTime, vmParams *vm.Params, next *iotago.AccountOutput) error {
	current := input.Output.(*iotago.AccountOutput)
	if !current.ImmutableFeatures.Equal(next.ImmutableFeatures) {
		return fmt.Errorf("%w: old state %s, next state %s", iotago.ErrInvalidAccountStateTransition, current.ImmutableFeatures, next.ImmutableFeatures)
	}

	if current.StateIndex == next.StateIndex {
		return accountGovernanceSTVF(current, next)
	}

	return accountStateSTVF(input, next, vmParams)
}

func accountGovernanceSTVF(current *iotago.AccountOutput, next *iotago.AccountOutput) error {
	switch {
	case current.Amount != next.Amount:
		return fmt.Errorf("%w: amount changed, in %d / out %d ", iotago.ErrInvalidAccountGovernanceTransition, current.Amount, next.Amount)
	case !current.NativeTokens.Equal(next.NativeTokens):
		return fmt.Errorf("%w: native tokens changed, in %v / out %v", iotago.ErrInvalidAccountGovernanceTransition, current.NativeTokens, next.NativeTokens)
	case current.StateIndex != next.StateIndex:
		return fmt.Errorf("%w: state index changed, in %d / out %d", iotago.ErrInvalidAccountGovernanceTransition, current.StateIndex, next.StateIndex)
	case !bytes.Equal(current.StateMetadata, next.StateMetadata):
		return fmt.Errorf("%w: state metadata changed, in %v / out %v", iotago.ErrInvalidAccountGovernanceTransition, current.StateMetadata, next.StateMetadata)
	case current.FoundryCounter != next.FoundryCounter:
		return fmt.Errorf("%w: foundry counter changed, in %d / out %d", iotago.ErrInvalidAccountGovernanceTransition, current.FoundryCounter, next.FoundryCounter)
	}
	return nil
}

func accountStateSTVF(input *vm.ChainOutputWithCreationTime, next *iotago.AccountOutput, vmParams *vm.Params) error {
	current := input.Output.(*iotago.AccountOutput)
	switch {
	case !current.StateController().Equal(next.StateController()):
		return fmt.Errorf("%w: state controller changed, in %v / out %v", iotago.ErrInvalidAccountStateTransition, current.StateController(), next.StateController())
	case !current.GovernorAddress().Equal(next.GovernorAddress()):
		return fmt.Errorf("%w: governance controller changed, in %v / out %v", iotago.ErrInvalidAccountStateTransition, current.GovernorAddress(), next.GovernorAddress())
	case current.FoundryCounter > next.FoundryCounter:
		return fmt.Errorf("%w: foundry counter of next state is less than previous, in %d / out %d", iotago.ErrInvalidAccountStateTransition, current.FoundryCounter, next.FoundryCounter)
	case current.StateIndex+1 != next.StateIndex:
		return fmt.Errorf("%w: state index %d on the input side but %d on the output side", iotago.ErrInvalidAccountStateTransition, current.StateIndex, next.StateIndex)
	}

	if err := iotago.FeatureUnchanged(iotago.FeatureMetadata, current.Features.MustSet(), next.Features.MustSet()); err != nil {
		return fmt.Errorf("%w: %s", iotago.ErrInvalidAccountStateTransition, err)
	}

	if err := accountBlockIssuerSTVF(input, next, vmParams); err != nil {
		return err
	}

	if err := accountStakingSTVF(current, next, vmParams); err != nil {
		return err
	}

	// check that for a foundry counter change, X amount of foundries were actually created
	if current.FoundryCounter == next.FoundryCounter {
		return nil
	}

	var seenNewFoundriesOfAccount uint32
	for _, output := range vmParams.WorkingSet.Tx.Essence.Outputs {
		foundryOutput, is := output.(*iotago.FoundryOutput)
		if !is {
			continue
		}

		if _, notNew := vmParams.WorkingSet.InChains[foundryOutput.MustID()]; notNew {
			continue
		}

		foundryAccountID := foundryOutput.Ident().(*iotago.AccountAddress).Chain()
		if !foundryAccountID.Matches(next.AccountID) {
			continue
		}
		seenNewFoundriesOfAccount++
	}

	expectedNewFoundriesCount := next.FoundryCounter - current.FoundryCounter
	if expectedNewFoundriesCount != seenNewFoundriesOfAccount {
		return fmt.Errorf("%w: %d new foundries were created but the account output's foundry counter changed by %d", iotago.ErrInvalidAccountStateTransition, seenNewFoundriesOfAccount, expectedNewFoundriesCount)
	}

	return nil
}

// If an account output has a block issuer feature, the following conditions for its transition must be checked.
// The block issuer credit must be non-negative.
// The expiry time of the block issuer feature, if creating new account or expired already, must be set at least MaxCommittableSlotAge greater than the TX slot index.
// Check that at least one Block Issuer Key is present.
func accountBlockIssuerSTVF(input *vm.ChainOutputWithCreationTime, next *iotago.AccountOutput, vmParams *vm.Params) error {
	current := input.Output.(*iotago.AccountOutput)
	currentBIFeat := current.FeatureSet().BlockIssuer()
	nextBIFeat := next.FeatureSet().BlockIssuer()
	// if the account has no block issuer feature.
	if currentBIFeat == nil && nextBIFeat == nil {
		return nil
	}
	// else if the account has negative bic, this is invalid.
	// new block issuers may not have a bic registered yet.
	if bic, exists := vmParams.WorkingSet.BIC[current.AccountID]; exists {
		if bic.Negative() {
			return fmt.Errorf("%w: negative block issuer credit", iotago.ErrInvalidBlockIssuerTransition)
		}
	} else {
		return fmt.Errorf("%w: no BIC provided for block issuer", iotago.ErrInvalidBlockIssuerTransition)
	}

	txSlotIndex := vmParams.WorkingSet.Tx.Essence.CreationTime
	if currentBIFeat.ExpirySlot >= txSlotIndex {
		// if the block issuer feature has not expired, it can not be removed.
		if nextBIFeat == nil {
			return fmt.Errorf("%w: cannot remove block issuer feature until it expires", iotago.ErrInvalidBlockIssuerTransition)
		}
		if nextBIFeat.ExpirySlot != 0 && nextBIFeat.ExpirySlot != currentBIFeat.ExpirySlot && nextBIFeat.ExpirySlot < txSlotIndex+vmParams.External.ProtocolParameters.MaxCommittableAge {
			return fmt.Errorf("%w: block issuer feature expiry set too soon", iotago.ErrInvalidBlockIssuerTransition)
		}

	} else if nextBIFeat != nil {
		// if the block issuer feature has expired, it must either be removed or expiry extended.
		if nextBIFeat.ExpirySlot != 0 && nextBIFeat.ExpirySlot < txSlotIndex+vmParams.External.ProtocolParameters.MaxCommittableAge {
			return fmt.Errorf("%w: block issuer feature expiry set too soon", iotago.ErrInvalidBlockIssuerTransition)
		}
	}

	// the Mana on the account on the input side must not be moved to any other outputs or accounts.
	manaDecayProvider := vmParams.External.ProtocolParameters.ManaDecayProvider()
	manaIn := vm.TotalManaIn(
		manaDecayProvider,
		vmParams.WorkingSet.Tx.Essence.CreationTime,
		vmParams.WorkingSet.UTXOInputsWithCreationTime,
	)
	manaOut := vm.TotalManaOut(
		vmParams.WorkingSet.Tx.Essence.Outputs,
		vmParams.WorkingSet.Tx.Essence.Allotments,
	)

	manaIn -= manaDecayProvider.StoredManaWithDecay(current.Mana, input.CreationTime, vmParams.WorkingSet.Tx.Essence.CreationTime)      // AccountInStored
	manaIn -= manaDecayProvider.PotentialManaWithDecay(current.Amount, input.CreationTime, vmParams.WorkingSet.Tx.Essence.CreationTime) // AccountInPotential
	manaOut -= next.Mana                                                                                                                // AccountOutStored
	manaOut -= vmParams.WorkingSet.Tx.Essence.Allotments.Get(current.AccountID)                                                         // AccountOutAllotted

	// subtract AccountOutLocked - we only consider basic and NFT outputs because only these output types can include a timelock and address unlock condition.
	for _, output := range vmParams.WorkingSet.OutputsByType[iotago.OutputBasic] {
		basicOutput, is := output.(*iotago.BasicOutput)
		if !is {
			continue
		}
		if basicOutput.UnlockConditionSet().HasManalockCondition(current.AccountID, txSlotIndex+vmParams.External.ProtocolParameters.MaxCommittableAge) {
			manaOut -= basicOutput.StoredMana()
		}
	}
	for _, output := range vmParams.WorkingSet.OutputsByType[iotago.OutputNFT] {
		nftOutput, is := output.(*iotago.NFTOutput)
		if !is {
			continue
		}
		if nftOutput.UnlockConditionSet().HasManalockCondition(current.AccountID, txSlotIndex+vmParams.External.ProtocolParameters.MaxCommittableAge) {
			manaOut -= nftOutput.StoredMana()
		}
	}

	if manaIn > manaOut {
		return fmt.Errorf("%w: cannot move Mana off an account", iotago.ErrInvalidBlockIssuerTransition)
	}

	return nil
}

func accountStakingSTVF(current *iotago.AccountOutput, next *iotago.AccountOutput, vmParams *vm.Params) error {
	currentStakingFeat := current.FeatureSet().Staking()
	nextStakingFeat := next.FeatureSet().Staking()

	if nextStakingFeat != nil && next.FeatureSet().BlockIssuer() == nil {
		return fmt.Errorf("%w: a staking feature can only be present if a block issuer feature is present", iotago.ErrInvalidStakingTransition)
	}

	if currentStakingFeat != nil {
		timeProvider := vmParams.External.ProtocolParameters.TimeProvider()
		creationEpoch := timeProvider.EpochsFromSlot(vmParams.WorkingSet.Tx.Essence.CreationTime)

		if creationEpoch < currentStakingFeat.EndEpoch {
			if nextStakingFeat == nil {
				return fmt.Errorf("%w: the staking feature cannot be removed before the end epoch", iotago.ErrInvalidStakingTransition)
			} else {
				if next.FeatureSet().BlockIssuer() == nil {
					return fmt.Errorf("%w: a staking feature can only be present if a block issuer feature is present", iotago.ErrInvalidStakingTransition)
				}
			}

			if currentStakingFeat.StakedAmount != nextStakingFeat.StakedAmount ||
				currentStakingFeat.FixedCost != nextStakingFeat.FixedCost ||
				currentStakingFeat.StartEpoch != nextStakingFeat.StartEpoch {
				return fmt.Errorf("%w: staked amount, fixed cost and start epoch must match on the input and output", iotago.ErrInvalidStakingTransition)
			}

			unbondingEpoch := creationEpoch + vmParams.External.ProtocolParameters.StakingUnbondingPeriod
			if currentStakingFeat.EndEpoch != nextStakingFeat.EndEpoch &&
				nextStakingFeat.EndEpoch < unbondingEpoch {
				return fmt.Errorf("%w: the end epoch must be in the future by at least the unbonding period (i.e. end epoch %d should be >= %d) or the end epoch must match on input and output side", iotago.ErrInvalidStakingTransition, nextStakingFeat.EndEpoch, unbondingEpoch)
			}
		} else {
			// Current epoch index is past the end epoch.
			_, isClaiming := vmParams.WorkingSet.Rewards[current.AccountID]

			// Mana Claiming by either removing the Feature or changing the feature's epoch range.
			if nextStakingFeat == nil {
				if !isClaiming {
					return fmt.Errorf("%w: cannot remove the staking feature without claiming rewards", iotago.ErrInvalidStakingTransition)
				}
			} else {
				if isClaiming {
					// When claiming with a feature on the output side, it must be transitioned as if it was newly added,
					// so the new epoch range is different.
					return accountStakingGenesisValidation(current, nextStakingFeat, vmParams)
				} else {
					// If not claiming, the feature must be unchanged.
					if !currentStakingFeat.Equal(nextStakingFeat) {
						return fmt.Errorf("%w: cannot change the staking feature without claiming rewards", iotago.ErrInvalidStakingTransition)
					}
				}
			}
		}
	}

	return nil
}

// Validates the rules for a newly added Staking Feature in an account,
// or one which was effectively removed and added within the same transaction.
// This is allowed as long as the epoch range of the old and new feature are disjoint.
func accountStakingGenesisValidation(acc *iotago.AccountOutput, stakingFeat *iotago.StakingFeature, vmParams *vm.Params) error {
	if acc.Amount < stakingFeat.StakedAmount {
		return fmt.Errorf("%w: the account's amount is less than the staked amount in the staking feature", iotago.ErrInvalidStakingTransition)
	}

	timeProvider := vmParams.External.ProtocolParameters.TimeProvider()
	creationEpoch := timeProvider.EpochsFromSlot(vmParams.WorkingSet.Tx.Essence.CreationTime)

	if stakingFeat.StartEpoch != creationEpoch {
		return fmt.Errorf("%w: the start epoch must be set to the epoch index of the transaction (%d)", iotago.ErrInvalidStakingTransition, creationEpoch)
	}

	unbondingEpoch := creationEpoch + vmParams.External.ProtocolParameters.StakingUnbondingPeriod
	if stakingFeat.EndEpoch < unbondingEpoch {
		return fmt.Errorf("%w: the end epoch must be in the future by at least the unbonding period (i.e. end epoch %d should be >= %d)", iotago.ErrInvalidStakingTransition, stakingFeat.EndEpoch, unbondingEpoch)
	}

	if acc.FeatureSet().BlockIssuer() == nil {
		return fmt.Errorf("%w: a staking feature can only be added if a block issuer feature is present", iotago.ErrInvalidStakingTransition)
	}

	return nil
}

func accountDestructionValid(input *vm.ChainOutputWithCreationTime, vmParams *vm.Params) error {
	outputToDestroy := input.Output.(*iotago.AccountOutput)
	BIFeat := outputToDestroy.FeatureSet().BlockIssuer()
	if BIFeat != nil {
		if BIFeat.ExpirySlot == 0 || BIFeat.ExpirySlot >= vmParams.WorkingSet.Tx.Essence.CreationTime {
			// TODO: better error
			return fmt.Errorf("%w: cannot destroy output until the block issuer feature expires", iotago.ErrInvalidBlockIssuerTransition)
		}
		if bic, exists := vmParams.WorkingSet.BIC[outputToDestroy.AccountID]; exists {
			if bic.Negative() {
				return fmt.Errorf("%w: negative block issuer credit", iotago.ErrInvalidBlockIssuerTransition)
			}
		} else {
			// TODO: better error
			return fmt.Errorf("%w: no BIC provided for block issuer", iotago.ErrInvalidBlockIssuerTransition)
		}
	}

	stakingFeat := outputToDestroy.FeatureSet().Staking()
	if stakingFeat != nil {
		timeProvider := vmParams.External.ProtocolParameters.TimeProvider()
		creationEpoch := timeProvider.EpochsFromSlot(vmParams.WorkingSet.Tx.Essence.CreationTime)

		if creationEpoch < stakingFeat.EndEpoch {
			return fmt.Errorf("%w: cannot destroy output until the staking feature is unbonded", iotago.ErrInvalidAccountStateTransition)
		}
	} else {
		// TODO: Mana Rewards Claiming.
		fmt.Println("todo")
	}

	return nil
}

func nftSTVF(input *vm.ChainOutputWithCreationTime, transType iotago.ChainTransitionType, next *iotago.NFTOutput, vmParams *vm.Params) error {
	switch transType {
	case iotago.ChainTransitionTypeGenesis:
		if err := nftGenesisValid(next, vmParams); err != nil {
			return &iotago.ChainTransitionError{Inner: err, Msg: fmt.Sprintf("NFT %s", next.NFTID)}
		}
	case iotago.ChainTransitionTypeStateChange:
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
		return fmt.Errorf("NFTOutput's ID is not zeroed even though it is new")
	}

	return vm.IsIssuerOnOutputUnlocked(current, vmParams.WorkingSet.UnlockedIdents)
}

func nftStateChangeValid(current *iotago.NFTOutput, next *iotago.NFTOutput) error {
	if !current.ImmutableFeatures.Equal(next.ImmutableFeatures) {
		return fmt.Errorf("old state %s, next state %s", current.ImmutableFeatures, next.ImmutableFeatures)
	}

	return nil
}

func foundrySTVF(input *vm.ChainOutputWithCreationTime, transType iotago.ChainTransitionType, next *iotago.FoundryOutput, vmParams *vm.Params) error {
	inSums := vmParams.WorkingSet.InNativeTokens
	outSums := vmParams.WorkingSet.OutNativeTokens

	switch transType {
	case iotago.ChainTransitionTypeGenesis:
		if err := foundryGenesisValid(next, vmParams, next.MustID(), outSums); err != nil {
			return fmt.Errorf("%w: foundry %s, token %s", err, next.MustID(), next.MustNativeTokenID())
		}
	case iotago.ChainTransitionTypeStateChange:
		current := input.Output.(*iotago.FoundryOutput)
		if err := foundryStateChangeValid(current, next, inSums, outSums); err != nil {
			return fmt.Errorf("%w: foundry %s, token %s", err, current.MustID(), current.MustNativeTokenID())
		}
	case iotago.ChainTransitionTypeDestroy:
		current := input.Output.(*iotago.FoundryOutput)
		if err := foundryDestructionValid(current, inSums, outSums); err != nil {
			return fmt.Errorf("%w: foundry %s, token %s", err, current.MustID(), current.MustNativeTokenID())
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
	accountID := current.Ident().(*iotago.AccountAddress).AccountID()
	inAccount, ok := vmParams.WorkingSet.InChains[accountID]
	if !ok {
		return fmt.Errorf("%w: missing input transitioning account output %s for new foundry output %s", iotago.ErrInvalidFoundryStateTransition, accountID, thisFoundryID)
	}

	outAccount, ok := vmParams.WorkingSet.OutChains[accountID]
	if !ok {
		return fmt.Errorf("%w: missing output transitioning account output %s for new foundry output %s", iotago.ErrInvalidFoundryStateTransition, accountID, thisFoundryID)
	}

	if err := foundrySerialNumberValid(current, vmParams, inAccount.Output.(*iotago.AccountOutput), outAccount.(*iotago.AccountOutput), thisFoundryID); err != nil {
		return err
	}

	return nil
}

func foundrySerialNumberValid(current *iotago.FoundryOutput, vmParams *vm.Params, inAccount *iotago.AccountOutput, outAccount *iotago.AccountOutput, thisFoundryID iotago.FoundryID) error {
	// this new foundry's serial number must be between the given foundry counter interval
	startSerial := inAccount.FoundryCounter
	endIncSerial := outAccount.FoundryCounter
	if startSerial >= current.SerialNumber || current.SerialNumber > endIncSerial {
		return fmt.Errorf("%w: new foundry output %s's serial number is not between the foundry counter interval of [%d,%d)", iotago.ErrInvalidFoundryStateTransition, thisFoundryID, startSerial, endIncSerial)
	}

	// OPTIMIZE: this loop happens on every STVF of every new foundry output
	// check order of serial number
	for outputIndex, output := range vmParams.WorkingSet.Tx.Essence.Outputs {
		otherFoundryOutput, is := output.(*iotago.FoundryOutput)
		if !is {
			continue
		}

		if !otherFoundryOutput.Ident().Equal(current.Ident()) {
			continue
		}

		otherFoundryID, err := otherFoundryOutput.ID()
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
			return fmt.Errorf("%w: new foundry output %s at index %d has bigger equal serial number than this foundry %s", iotago.ErrInvalidFoundryStateTransition, otherFoundryID, outputIndex, thisFoundryID)
		}
	}

	return nil
}

func foundryStateChangeValid(current *iotago.FoundryOutput, next *iotago.FoundryOutput, inSums iotago.NativeTokenSum, outSums iotago.NativeTokenSum) error {
	if !current.ImmutableFeatures.Equal(next.ImmutableFeatures) {
		return fmt.Errorf("%w: old state %s, next state %s", iotago.ErrInvalidFoundryStateTransition, current.ImmutableFeatures, next.ImmutableFeatures)
	}

	// the check for the serial number and token scheme not being mutated is implicit
	// as a change would cause the foundry ID to be different, which would result in
	// no matching foundry to be found to validate the state transition against
	switch {
	case current.MustID() != next.MustID():
		// impossible invariant as the STVF should be called via the matching next foundry output
		panic(fmt.Sprintf("foundry IDs mismatch in state transition validation function: have %v got %v", current.MustID(), next.MustID()))
	}

	nativeTokenID := current.MustNativeTokenID()

	return current.TokenScheme.StateTransition(iotago.ChainTransitionTypeStateChange, next.TokenScheme, inSums.ValueOrBigInt0(nativeTokenID), outSums.ValueOrBigInt0(nativeTokenID))
}

func foundryDestructionValid(current *iotago.FoundryOutput, inSums iotago.NativeTokenSum, outSums iotago.NativeTokenSum) error {
	nativeTokenID := current.MustNativeTokenID()

	return current.TokenScheme.StateTransition(iotago.ChainTransitionTypeDestroy, nil, inSums.ValueOrBigInt0(nativeTokenID), outSums.ValueOrBigInt0(nativeTokenID))
}

func delegationSTVF(input *vm.ChainOutputWithCreationTime, transType iotago.ChainTransitionType, next *iotago.DelegationOutput, vmParams *vm.Params) error {
	switch transType {
	case iotago.ChainTransitionTypeGenesis:
		if err := delegationGenesisValid(next, vmParams); err != nil {
			return &iotago.ChainTransitionError{Inner: err, Msg: fmt.Sprintf("Delegation %s", next.DelegationID)}
		}
	case iotago.ChainTransitionTypeStateChange:
		current := input.Output.(*iotago.DelegationOutput)
		if err := delegationStateChangeValid(current, next, vmParams); err != nil {
			return &iotago.ChainTransitionError{Inner: err, Msg: fmt.Sprintf("Delegation %s", current.DelegationID)}
		}
	case iotago.ChainTransitionTypeDestroy:
		// TODO: Mana Rewards Claiming?
		return nil
	default:
		panic("unknown chain transition type in DelegationOutput")
	}

	return nil
}

func delegationGenesisValid(current *iotago.DelegationOutput, vmParams *vm.Params) error {
	if !current.DelegationID.Empty() {
		return fmt.Errorf("%w: DelegationOutput's ID is not zeroed even though it is new", iotago.ErrInvalidDelegationGenesis)
	}

	timeProvider := vmParams.External.ProtocolParameters.TimeProvider()
	creationSlot := vmParams.WorkingSet.Tx.Essence.CreationTime
	creationEpoch := timeProvider.EpochsFromSlot(creationSlot)
	votingPowerSlot := votingPowerCalculationSlot(creationSlot, timeProvider)

	var expectedStartEpoch iotago.EpochIndex
	if creationSlot <= votingPowerSlot {
		expectedStartEpoch = creationEpoch + 1
	} else {
		expectedStartEpoch = creationEpoch + 2
	}

	if current.StartEpoch != uint64(expectedStartEpoch) {
		return fmt.Errorf("%w: DelegationOutput's start epoch is expected to be %d", iotago.ErrInvalidDelegationGenesis, expectedStartEpoch)
	}

	if current.DelegatedAmount != current.Amount {
		return fmt.Errorf("%w: DelegationOutput's delegated amount is not equal to amount", iotago.ErrInvalidDelegationGenesis)
	}

	if current.EndEpoch != 0 {
		return fmt.Errorf("%w: DelegationOutput's end epoch is not set to zero", iotago.ErrInvalidDelegationGenesis)
	}

	return vm.IsIssuerOnOutputUnlocked(current, vmParams.WorkingSet.UnlockedIdents)
}

func delegationStateChangeValid(current *iotago.DelegationOutput, next *iotago.DelegationOutput, vmParams *vm.Params) error {
	// State transitioning a Delegation Output is always a transition to the delayed claiming state.
	// Since they can only be transitioned once, the input will always need to have a zeroed ID.
	if !current.DelegationID.Empty() {
		return fmt.Errorf("%w: consumed DelegationOutput's ID is not zeroed", iotago.ErrInvalidDelegationTransition)
	}

	if !current.ImmutableFeatures.Equal(next.ImmutableFeatures) {
		return fmt.Errorf("immutable features mismatch: old state %s, next state %s", current.ImmutableFeatures, next.ImmutableFeatures)
	}

	if current.DelegatedAmount != next.DelegatedAmount ||
		current.ValidatorID != next.ValidatorID ||
		current.StartEpoch != next.StartEpoch {
		return fmt.Errorf("%w: delegated amount, validator ID and start epoch must match on the input and output", iotago.ErrInvalidDelegationTransition)
	}

	timeProvider := vmParams.External.ProtocolParameters.TimeProvider()
	creationSlot := vmParams.WorkingSet.Tx.Essence.CreationTime
	creationEpoch := timeProvider.EpochsFromSlot(creationSlot)
	votingPowerSlot := votingPowerCalculationSlot(creationSlot, timeProvider)

	var expectedEndEpoch iotago.EpochIndex
	if creationSlot <= votingPowerSlot {
		expectedEndEpoch = creationEpoch
	} else {
		expectedEndEpoch = creationEpoch + 1
	}

	if current.EndEpoch != uint64(expectedEndEpoch) {
		return fmt.Errorf("%w: DelegationOutput's end epoch is expected to be %d", iotago.ErrInvalidDelegationGenesis, expectedEndEpoch)
	}

	return nil
}

// votingPowerCalculationSlot returns the slot at the end of which the voting power for the next epoch is calculated.
func votingPowerCalculationSlot(currentSlotIndex iotago.SlotIndex, timeProvider *iotago.TimeProvider) iotago.SlotIndex {
	// currentEpoch := timeProvider.EpochsFromSlot(currentSlotIndex)
	// startSlotNextEpoch := timeProvider.EpochStart(currentEpoch)
	// votingPowerCalcSlotNextEpoch := startSlotNextEpoch - iotago.SlotIndex(vmParams.External.ProtocolParameters.MaxCommitableAge)
	// TODO: Finalize when committee selection is finalized.
	return currentSlotIndex
}
