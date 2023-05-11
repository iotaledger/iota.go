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
		},
	}
}

type virtualMachine struct {
	execList []vm.ExecFunc
}

func (stardustVM *virtualMachine) Execute(t *iotago.Transaction, vmParams *vm.Params, inputs iotago.OutputSet, overrideFuncs ...vm.ExecFunc) error {
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

func (stardustVM *virtualMachine) ChainSTVF(transType iotago.ChainTransitionType, current iotago.ChainOutput, next iotago.ChainOutput, vmParams *vm.Params) error {
	var ok bool
	switch current.(type) {
	case *iotago.AliasOutput:
		var nextAlias *iotago.AliasOutput
		if next != nil {
			if nextAlias, ok = next.(*iotago.AliasOutput); !ok {
				return fmt.Errorf("can only state transition to another alias output")
			}
		}
		return aliasSTVF(current.(*iotago.AliasOutput), transType, nextAlias, vmParams)
	case *iotago.FoundryOutput:
		var nextFoundry *iotago.FoundryOutput
		if next != nil {
			if nextFoundry, ok = next.(*iotago.FoundryOutput); !ok {
				return fmt.Errorf("can only state transition to another foundry output")
			}
		}
		return foundrySTVF(current.(*iotago.FoundryOutput), transType, nextFoundry, vmParams)
	case *iotago.NFTOutput:
		var nextNFT *iotago.NFTOutput
		if next != nil {
			if nextNFT, ok = next.(*iotago.NFTOutput); !ok {
				return fmt.Errorf("can only state transition to another NFT output")
			}
		}
		return nftSTVF(current.(*iotago.NFTOutput), transType, nextNFT, vmParams)
	default:
		panic(fmt.Sprintf("invalid output type %s passed to Stardust virtual machine", current))
	}
}

// For output AliasOutput(s) with non-zeroed AliasID, there must be a corresponding input AliasOutput where either its
// AliasID is zeroed and StateIndex and FoundryCounter are zero or an input AliasOutput with the same AliasID.
//
// On alias state transitions: The StateIndex must be incremented by 1 and Only Amount, NativeTokens, StateIndex, StateMetadata and FoundryCounter can be mutated.
//
// On alias governance transition: Only StateController (must be mutated), GovernanceController and the MetadataBlock can be mutated.
func aliasSTVF(a *iotago.AliasOutput, transType iotago.ChainTransitionType, next *iotago.AliasOutput, vmParams *vm.Params) error {
	var err error
	switch transType {
	case iotago.ChainTransitionTypeGenesis:
		err = aliasGenesisValid(a, vmParams)
	case iotago.ChainTransitionTypeStateChange:
		err = aliasStateChangeValid(a, vmParams, next)
	case iotago.ChainTransitionTypeDestroy:
		return nil
	default:
		panic("unknown chain transition type in AliasOutput")
	}
	if err != nil {
		return &iotago.ChainTransitionError{Inner: err, Msg: fmt.Sprintf("alias %s", a.AliasID)}
	}
	return nil
}

func aliasGenesisValid(current *iotago.AliasOutput, vmParams *vm.Params) error {
	if !current.AliasID.Empty() {
		return fmt.Errorf("AliasOutput's ID is not zeroed even though it is new")
	}
	return vm.IsIssuerOnOutputUnlocked(current, vmParams.WorkingSet.UnlockedIdents)
}

func aliasStateChangeValid(current *iotago.AliasOutput, vmParams *vm.Params, next *iotago.AliasOutput) error {
	if !current.ImmutableFeatures.Equal(next.ImmutableFeatures) {
		return fmt.Errorf("old state %s, next state %s", current.ImmutableFeatures, next.ImmutableFeatures)
	}
	if current.StateIndex == next.StateIndex {
		return aliasGovernanceSTVF(current, next)
	}
	return aliasStateSTVF(current, next, vmParams)
}

func aliasGovernanceSTVF(current *iotago.AliasOutput, next *iotago.AliasOutput) error {
	switch {
	case current.Amount != next.Amount:
		return fmt.Errorf("%w: amount changed, in %d / out %d ", iotago.ErrInvalidAliasGovernanceTransition, current.Amount, next.Amount)
	case !current.NativeTokens.Equal(next.NativeTokens):
		return fmt.Errorf("%w: native tokens changed, in %v / out %v", iotago.ErrInvalidAliasGovernanceTransition, current.NativeTokens, next.NativeTokens)
	case current.StateIndex != next.StateIndex:
		return fmt.Errorf("%w: state index changed, in %d / out %d", iotago.ErrInvalidAliasGovernanceTransition, current.StateIndex, next.StateIndex)
	case !bytes.Equal(current.StateMetadata, next.StateMetadata):
		return fmt.Errorf("%w: state metadata changed, in %v / out %v", iotago.ErrInvalidAliasGovernanceTransition, current.StateMetadata, next.StateMetadata)
	case current.FoundryCounter != next.FoundryCounter:
		return fmt.Errorf("%w: foundry counter changed, in %d / out %d", iotago.ErrInvalidAliasGovernanceTransition, current.FoundryCounter, next.FoundryCounter)
	}
	return nil
}

func aliasStateSTVF(current *iotago.AliasOutput, next *iotago.AliasOutput, vmParams *vm.Params) error {
	switch {
	case !current.StateController().Equal(next.StateController()):
		return fmt.Errorf("%w: state controller changed, in %v / out %v", iotago.ErrInvalidAliasStateTransition, current.StateController(), next.StateController())
	case !current.GovernorAddress().Equal(next.GovernorAddress()):
		return fmt.Errorf("%w: governance controller changed, in %v / out %v", iotago.ErrInvalidAliasStateTransition, current.GovernorAddress(), next.GovernorAddress())
	case current.FoundryCounter > next.FoundryCounter:
		return fmt.Errorf("%w: foundry counter of next state is less than previous, in %d / out %d", iotago.ErrInvalidAliasStateTransition, current.FoundryCounter, next.FoundryCounter)
	case current.StateIndex+1 != next.StateIndex:
		return fmt.Errorf("%w: state index %d on the input side but %d on the output side", iotago.ErrInvalidAliasStateTransition, current.StateIndex, next.StateIndex)
	}

	if err := iotago.FeatureUnchanged(iotago.FeatureMetadata, current.Features.MustSet(), next.Features.MustSet()); err != nil {
		return fmt.Errorf("%w: %s", iotago.ErrInvalidAliasStateTransition, err)
	}

	// check that for a foundry counter change, X amount of foundries were actually created
	if current.FoundryCounter == next.FoundryCounter {
		return nil
	}

	var seenNewFoundriesOfAlias uint32
	for _, output := range vmParams.WorkingSet.Tx.Essence.Outputs {
		foundryOutput, is := output.(*iotago.FoundryOutput)
		if !is {
			continue
		}

		if _, notNew := vmParams.WorkingSet.InChains[foundryOutput.MustID()]; notNew {
			continue
		}

		foundryAliasID := foundryOutput.Ident().(*iotago.AliasAddress).Chain()
		if !foundryAliasID.Matches(next.AliasID) {
			continue
		}
		seenNewFoundriesOfAlias++
	}

	expectedNewFoundriesCount := next.FoundryCounter - current.FoundryCounter
	if expectedNewFoundriesCount != seenNewFoundriesOfAlias {
		return fmt.Errorf("%w: %d new foundries were created but the alias output's foundry counter changed by %d", iotago.ErrInvalidAliasStateTransition, seenNewFoundriesOfAlias, expectedNewFoundriesCount)
	}

	return aliasBlockIssuerSTVF(current, next, vmParams)
}

// If an alias output has a block issuer feature, the following conditions for its transition must be checked.
// The expiry time of the block issuer feature, if changed, must be set at least MaxCommitableSlotAge greater than the TX slot index.
// Check that at least one Block Issuer Key is present
// TODO: add block issuer credit check for account destruction.
func aliasBlockIssuerSTVF(current *iotago.AliasOutput, next *iotago.AliasOutput, vmParams *vm.Params) error {
	currentBIFeat := current.FeatureSet().BlockIssuer()
	nextBIFeat := next.FeatureSet().BlockIssuer()
	// if the account has no block issuer feature.
	if currentBIFeat == nil && nextBIFeat == nil {
		return nil
	}
	// else if the account has negative bic, this is invalid.
	// new block issuers may not have a bic registered yet.
	if bic, exists := vmParams.WorkingSet.BIC[current.AliasID]; exists {
		if bic < 0 {
			return fmt.Errorf("%w: Negative block issuer credit", iotago.ErrInvalidBlockIssuerTransition)
		}
	}
	txSlotIndex := vmParams.External.ProtocolParameters.SlotTimeProvider().IndexFromTime(vmParams.WorkingSet.Tx.Essence.CreationTime)

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
	return nil
}

func nftSTVF(current *iotago.NFTOutput, transType iotago.ChainTransitionType, next *iotago.NFTOutput, vmParams *vm.Params) error {
	var err error
	switch transType {
	case iotago.ChainTransitionTypeGenesis:
		err = nftGenesisValid(current, vmParams)
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

func foundrySTVF(current *iotago.FoundryOutput, transType iotago.ChainTransitionType, next *iotago.FoundryOutput, vmParams *vm.Params) error {
	inSums := vmParams.WorkingSet.InNativeTokens
	outSums := vmParams.WorkingSet.OutNativeTokens

	var err error
	switch transType {
	case iotago.ChainTransitionTypeGenesis:
		err = foundryGenesisValid(current, vmParams, current.MustID(), outSums)
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

	// grab foundry counter from transitioning AliasOutput
	aliasID := current.Ident().(*iotago.AliasAddress).AliasID()
	inAlias, ok := vmParams.WorkingSet.InChains[aliasID]
	if !ok {
		return fmt.Errorf("missing input transitioning alias output %s for new foundry output %s", aliasID, thisFoundryID)
	}

	outAlias, ok := vmParams.WorkingSet.OutChains[aliasID]
	if !ok {
		return fmt.Errorf("missing output transitioning alias output %s for new foundry output %s", aliasID, thisFoundryID)
	}

	if err := foundrySerialNumberValid(current, vmParams, inAlias.(*iotago.AliasOutput), outAlias.(*iotago.AliasOutput), thisFoundryID); err != nil {
		return err
	}

	return nil
}

func foundrySerialNumberValid(current *iotago.FoundryOutput, vmParams *vm.Params, inAlias *iotago.AliasOutput, outAlias *iotago.AliasOutput, thisFoundryID iotago.FoundryID) error {
	// this new foundry's serial number must be between the given foundry counter interval
	startSerial := inAlias.FoundryCounter
	endIncSerial := outAlias.FoundryCounter
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
