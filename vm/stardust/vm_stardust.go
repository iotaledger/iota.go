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

func (stardustVM *virtualMachine) Execute(t *iotago.Transaction, vmParams *vm.Params, inputs iotago.InputSet, bic iotago.BlockIssuanceCredit, overrideFuncs ...vm.ExecFunc) error {
	var err error
	vmParams.WorkingSet, err = vm.NewVMParamsWorkingSet(t, inputs, bic)
	if err != nil {
		return err
	}

	if len(overrideFuncs) > 0 {
		return vm.RunVMFuncs(stardustVM, vmParams, overrideFuncs...)
	}

	return vm.RunVMFuncs(stardustVM, vmParams, stardustVM.execList...)
}

func (stardustVM *virtualMachine) ChainSTVF(transType iotago.ChainTransitionType, input iotago.ChainOutputWithCreationTime, next iotago.ChainOutput, vmParams *vm.Params) error {
	var ok bool
	switch input.ChainOutput.(type) {
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
	default:
		panic(fmt.Sprintf("invalid output type %s passed to Stardust virtual machine", input.ChainOutput))
	}
}

// For output AccountOutput(s) with non-zeroed AccountID, there must be a corresponding input AccountOutput where either its
// AccountID is zeroed and StateIndex and FoundryCounter are zero or an input AccountOutput with the same AccountID.
//
// On account state transitions: The StateIndex must be incremented by 1 and Only Amount, NativeTokens, StateIndex, StateMetadata and FoundryCounter can be mutated.
//
// On account governance transition: Only StateController (must be mutated), GovernanceController and the MetadataBlock can be mutated.
func accountSTVF(input iotago.ChainOutputWithCreationTime, transType iotago.ChainTransitionType, next *iotago.AccountOutput, vmParams *vm.Params) error {
	a := input.ChainOutput.(*iotago.AccountOutput)
	var err error
	switch transType {
	case iotago.ChainTransitionTypeGenesis:
		err = accountGenesisValid(next, vmParams)
	case iotago.ChainTransitionTypeStateChange:
		err = accountStateChangeValid(input, vmParams, next)
	case iotago.ChainTransitionTypeDestroy:
		err = accountDestructionValid(input, vmParams)
	default:
		panic("unknown chain transition type in AccountOutput")
	}
	if err != nil {
		return &iotago.ChainTransitionError{Inner: err, Msg: fmt.Sprintf("account %s", a.AccountID)}
	}
	return nil
}

func accountGenesisValid(current *iotago.AccountOutput, vmParams *vm.Params) error {
	if !current.AccountID.Empty() {
		return fmt.Errorf("AccountOutput's ID is not zeroed even though it is new")
	}
	return vm.IsIssuerOnOutputUnlocked(current, vmParams.WorkingSet.UnlockedIdents)
}

func accountStateChangeValid(input iotago.ChainOutputWithCreationTime, vmParams *vm.Params, next *iotago.AccountOutput) error {
	current := input.ChainOutput.(*iotago.AccountOutput)
	if !current.ImmutableFeatures.Equal(next.ImmutableFeatures) {
		return fmt.Errorf("old state %s, next state %s", current.ImmutableFeatures, next.ImmutableFeatures)
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

func accountStateSTVF(input iotago.ChainOutputWithCreationTime, next *iotago.AccountOutput, vmParams *vm.Params) error {
	current := input.ChainOutput.(*iotago.AccountOutput)
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

	return accountBlockIssuerSTVF(input, next, vmParams)
}

// If an account output has a block issuer feature, the following conditions for its transition must be checked.
// The block issuer credit must be non negative.
// The expiry time of the block issuer feature, if changed, must be set at least MaxCommitableSlotAge greater than the TX slot index.
// Check that at least one Block Issuer Key is present
func accountBlockIssuerSTVF(input iotago.ChainOutputWithCreationTime, next *iotago.AccountOutput, vmParams *vm.Params) error {
	current := input.ChainOutput.(*iotago.AccountOutput)
	currentBIFeat := current.FeatureSet().BlockIssuer()
	nextBIFeat := next.FeatureSet().BlockIssuer()
	// if the account has no block issuer feature.
	if currentBIFeat == nil && nextBIFeat == nil {
		return nil
	}
	// else if the account has negative bic, this is invalid.
	// new block issuers may not have a bic registered yet.
	if bic, exists := vmParams.WorkingSet.BIC[current.AccountID]; exists {
		if bic < 0 {
			return fmt.Errorf("%w: Negative block issuer credit", iotago.ErrInvalidBlockIssuerTransition)
		}
	}
	txSlotIndex := vmParams.WorkingSet.Tx.Essence.CreationTime

	if currentBIFeat.ExpirySlot >= txSlotIndex {
		// if the block issuer feature has not expired, it can not be removed.
		if nextBIFeat == nil {
			return fmt.Errorf("%w: Cannot remove block issuer feature until it expires", iotago.ErrInvalidBlockIssuerTransition)
		}
		if nextBIFeat.ExpirySlot != currentBIFeat.ExpirySlot && nextBIFeat.ExpirySlot < txSlotIndex+iotago.MaxCommitableSlotAge {
			return fmt.Errorf("%w: Block issuer feature expiry set too soon", iotago.ErrInvalidBlockIssuerTransition)
		}

	} else if nextBIFeat != nil {
		// if the block issuer feature has expired, it must either be removed or expiry extended.
		if nextBIFeat.ExpirySlot < txSlotIndex+iotago.MaxCommitableSlotAge {
			return fmt.Errorf("%w: Block issuer feature expiry set too soon", iotago.ErrInvalidBlockIssuerTransition)
		}
	}

	// the Mana on the account on the input side must be less than or equal to the total on the output side.
	// TODO: apply decay to input side
	// TODO: include outputs with ManaLock Conditions in this check
	timeHeld := uint64(vmParams.WorkingSet.Tx.Essence.CreationTime - input.CreationTime)
	accountManaIn := current.Mana + current.Amount*uint64(vmParams.External.ProtocolParameters.ManaGenerationRate)*timeHeld
	accountManaOut := next.Mana + vmParams.WorkingSet.Tx.Essence.Allotments.Get(current.AccountID)
	if accountManaIn < accountManaOut {
		return fmt.Errorf("%w: Cannot move Mana off an account", iotago.ErrInvalidBlockIssuerTransition)
	}
	return nil
}

func accountDestructionValid(input iotago.ChainOutputWithCreationTime, vmParams *vm.Params) error {
	outputToDestroy := input.ChainOutput.(*iotago.AccountOutput)
	BIFeat := outputToDestroy.FeatureSet().BlockIssuer()
	if BIFeat != nil {
		if BIFeat.ExpirySlot >= vmParams.WorkingSet.Tx.Essence.CreationTime {
			// TODO: better error
			return fmt.Errorf("%w: Cannot destroy output until the block issuer feature expires", iotago.ErrInvalidBlockIssuerTransition)
		}
		if bic, exists := vmParams.WorkingSet.BIC[outputToDestroy.AccountID]; exists {
			if bic < 0 {
				return fmt.Errorf("%w: Negative block issuer credit", iotago.ErrInvalidBlockIssuerTransition)
			}
		} else {
			// TODO: better error
			panic("No BIC provided for block issuer")
		}
	}
	return nil
}

func nftSTVF(input iotago.ChainOutputWithCreationTime, transType iotago.ChainTransitionType, next *iotago.NFTOutput, vmParams *vm.Params) error {
	current := input.ChainOutput.(*iotago.NFTOutput)
	var err error
	switch transType {
	case iotago.ChainTransitionTypeGenesis:
		err = nftGenesisValid(next, vmParams)
	case iotago.ChainTransitionTypeStateChange:
		err = nftStateChangeValid(current, next)
	case iotago.ChainTransitionTypeDestroy:
		return nil
	default:
		panic("unknown chain transition type in NFTOutput")
	}
	if err != nil {
		return &iotago.ChainTransitionError{Inner: err, Msg: fmt.Sprintf("NFT %s", current.NFTID)}
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

func foundrySTVF(input iotago.ChainOutputWithCreationTime, transType iotago.ChainTransitionType, next *iotago.FoundryOutput, vmParams *vm.Params) error {
	current := input.ChainOutput.(*iotago.FoundryOutput)
	inSums := vmParams.WorkingSet.InNativeTokens
	outSums := vmParams.WorkingSet.OutNativeTokens

	var err error
	switch transType {
	case iotago.ChainTransitionTypeGenesis:
		err = foundryGenesisValid(next, vmParams, next.MustID(), outSums)
	case iotago.ChainTransitionTypeStateChange:
		err = foundryStateChangeValid(current, next, inSums, outSums)
	case iotago.ChainTransitionTypeDestroy:
		err = foundryDestructionValid(current, inSums, outSums)
	default:
		panic("unknown chain transition type in FoundryOutput")
	}
	if err != nil {
		return &iotago.ChainTransitionError{Inner: err, Msg: fmt.Sprintf("foundry %s, token %s", current.MustID(), current.MustNativeTokenID())}
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
		return fmt.Errorf("missing input transitioning account output %s for new foundry output %s", accountID, thisFoundryID)
	}

	outAccount, ok := vmParams.WorkingSet.OutChains[accountID]
	if !ok {
		return fmt.Errorf("missing output transitioning account output %s for new foundry output %s", accountID, thisFoundryID)
	}

	if err := foundrySerialNumberValid(current, vmParams, inAccount.ChainOutput.(*iotago.AccountOutput), outAccount.(*iotago.AccountOutput), thisFoundryID); err != nil {
		return err
	}

	return nil
}

func foundrySerialNumberValid(current *iotago.FoundryOutput, vmParams *vm.Params, inAccount *iotago.AccountOutput, outAccount *iotago.AccountOutput, thisFoundryID iotago.FoundryID) error {
	// this new foundry's serial number must be between the given foundry counter interval
	startSerial := inAccount.FoundryCounter
	endIncSerial := outAccount.FoundryCounter
	if startSerial >= current.SerialNumber || current.SerialNumber > endIncSerial {
		return fmt.Errorf("new foundry output %s's serial number is not between the foundry counter interval of [%d,%d)", thisFoundryID, startSerial, endIncSerial)
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
			return fmt.Errorf("new foundry output %s at index %d has bigger equal serial number than this foundry %s", otherFoundryID, outputIndex, thisFoundryID)
		}
	}
	return nil
}

func foundryStateChangeValid(current *iotago.FoundryOutput, next *iotago.FoundryOutput, inSums iotago.NativeTokenSum, outSums iotago.NativeTokenSum) error {
	if !current.ImmutableFeatures.Equal(next.ImmutableFeatures) {
		return fmt.Errorf("old state %s, next state %s", current.ImmutableFeatures, next.ImmutableFeatures)
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
	if err := current.TokenScheme.StateTransition(iotago.ChainTransitionTypeStateChange, next.TokenScheme, inSums.ValueOrBigInt0(nativeTokenID), outSums.ValueOrBigInt0(nativeTokenID)); err != nil {
		return err
	}

	return nil
}

func foundryDestructionValid(current *iotago.FoundryOutput, inSums iotago.NativeTokenSum, outSums iotago.NativeTokenSum) error {
	nativeTokenID := current.MustNativeTokenID()
	if err := current.TokenScheme.StateTransition(iotago.ChainTransitionTypeDestroy, nil, inSums.ValueOrBigInt0(nativeTokenID), outSums.ValueOrBigInt0(nativeTokenID)); err != nil {
		return err
	}
	return nil
}
