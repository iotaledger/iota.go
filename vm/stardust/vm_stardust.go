package stardust

import (
	"bytes"
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/vm"
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

func (stardustVM *virtualMachine) Execute(t *iotago.Transaction, vmParas *vm.Paras, inputs iotago.OutputSet, overrideFuncs ...vm.ExecFunc) error {
	var err error
	vmParas.WorkingSet, err = vm.NewVMParasWorkingSet(t, inputs)
	if err != nil {
		return err
	}

	if len(overrideFuncs) > 0 {
		return vm.RunVMFuncs(stardustVM, vmParas, overrideFuncs...)
	}

	return vm.RunVMFuncs(stardustVM, vmParas, stardustVM.execList...)
}

func (stardustVM *virtualMachine) ChainSTVF(transType iotago.ChainTransitionType, current iotago.ChainOutput, next iotago.ChainOutput, vmParas *vm.Paras) error {
	var ok bool
	switch current.(type) {
	case *iotago.AliasOutput:
		var nextAlias *iotago.AliasOutput
		if next != nil {
			if nextAlias, ok = next.(*iotago.AliasOutput); !ok {
				return fmt.Errorf("can only state transition to another alias output")
			}
		}
		return aliasSTVF(current.(*iotago.AliasOutput), transType, nextAlias, vmParas)
	case *iotago.FoundryOutput:
		var nextFoundry *iotago.FoundryOutput
		if next != nil {
			if nextFoundry, ok = next.(*iotago.FoundryOutput); !ok {
				return fmt.Errorf("can only state transition to another foundry output")
			}
		}
		return foundrySTVF(current.(*iotago.FoundryOutput), transType, nextFoundry, vmParas)
	case *iotago.NFTOutput:
		var nextNFT *iotago.NFTOutput
		if next != nil {
			if nextNFT, ok = next.(*iotago.NFTOutput); !ok {
				return fmt.Errorf("can only state transition to another NFT output")
			}
		}
		return nftSTVF(current.(*iotago.NFTOutput), transType, nextNFT, vmParas)
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
func aliasSTVF(a *iotago.AliasOutput, transType iotago.ChainTransitionType, next *iotago.AliasOutput, vmParas *vm.Paras) error {
	var err error
	switch transType {
	case iotago.ChainTransitionTypeGenesis:
		err = aliasGenesisValid(a, vmParas)
	case iotago.ChainTransitionTypeStateChange:
		err = aliasStateChangeValid(a, vmParas, next)
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

func aliasGenesisValid(current *iotago.AliasOutput, vmParas *vm.Paras) error {
	if !current.AliasID.Empty() {
		return fmt.Errorf("AliasOutput's ID is not zeroed even though it is new")
	}
	return vm.IsIssuerOnOutputUnlocked(current, vmParas.WorkingSet.UnlockedIdents)
}

func aliasStateChangeValid(current *iotago.AliasOutput, vmParas *vm.Paras, next *iotago.AliasOutput) error {
	if !current.ImmutableFeatures.Equal(next.ImmutableFeatures) {
		return fmt.Errorf("old state %s, next state %s", current.ImmutableFeatures, next.ImmutableFeatures)
	}
	if current.StateIndex == next.StateIndex {
		return aliasGovernanceSTVF(current, next)
	}
	return aliasStateSTVF(current, next, vmParas)
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

func aliasStateSTVF(current *iotago.AliasOutput, next *iotago.AliasOutput, vmParas *vm.Paras) error {
	switch {
	case !current.StateController().Equal(next.StateController()):
		return fmt.Errorf("%w: state controller changed, in %v / out %v", iotago.ErrInvalidAliasStateTransition, current.StateController(), next.StateController())
	case !current.GovernorAddress().Equal(next.GovernorAddress()):
		return fmt.Errorf("%w: governance controller changed, in %v / out %v", iotago.ErrInvalidAliasStateTransition, current.StateController(), next.StateController())
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
	for _, output := range vmParas.WorkingSet.Tx.Essence.Outputs {
		foundryOutput, is := output.(*iotago.FoundryOutput)
		if !is {
			continue
		}

		if _, notNew := vmParas.WorkingSet.InChains[foundryOutput.MustID()]; notNew {
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

	return nil
}

func nftSTVF(current *iotago.NFTOutput, transType iotago.ChainTransitionType, next *iotago.NFTOutput, vmParas *vm.Paras) error {
	var err error
	switch transType {
	case iotago.ChainTransitionTypeGenesis:
		err = nftGenesisValid(current, vmParas)
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

func nftGenesisValid(current *iotago.NFTOutput, vmParas *vm.Paras) error {
	if !current.NFTID.Empty() {
		return fmt.Errorf("NFTOutput's ID is not zeroed even though it is new")
	}
	return vm.IsIssuerOnOutputUnlocked(current, vmParas.WorkingSet.UnlockedIdents)
}

func nftStateChangeValid(current *iotago.NFTOutput, next *iotago.NFTOutput) error {
	if !current.ImmutableFeatures.Equal(next.ImmutableFeatures) {
		return fmt.Errorf("old state %s, next state %s", current.ImmutableFeatures, next.ImmutableFeatures)
	}
	return nil
}

func foundrySTVF(current *iotago.FoundryOutput, transType iotago.ChainTransitionType, next *iotago.FoundryOutput, vmParas *vm.Paras) error {
	inSums := vmParas.WorkingSet.InNativeTokens
	outSums := vmParas.WorkingSet.OutNativeTokens

	var err error
	switch transType {
	case iotago.ChainTransitionTypeGenesis:
		err = foundryGenesisValid(current, vmParas, current.MustID(), outSums)
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

func foundryGenesisValid(current *iotago.FoundryOutput, vmParas *vm.Paras, thisFoundryID iotago.FoundryID, outSums iotago.NativeTokenSum) error {
	nativeTokenID := current.MustNativeTokenID()
	if err := current.TokenScheme.StateTransition(iotago.ChainTransitionTypeGenesis, nil, nil, outSums.ValueOrBigInt0(nativeTokenID)); err != nil {
		return err
	}

	// grab foundry counter from transitioning AliasOutput
	aliasID := current.Ident().(*iotago.AliasAddress).AliasID()
	inAlias, ok := vmParas.WorkingSet.InChains[aliasID]
	if !ok {
		return fmt.Errorf("missing input transitioning alias output %s for new foundry output %s", aliasID, thisFoundryID)
	}

	outAlias, ok := vmParas.WorkingSet.OutChains[aliasID]
	if !ok {
		return fmt.Errorf("missing output transitioning alias output %s for new foundry output %s", aliasID, thisFoundryID)
	}

	if err := foundrySerialNumberValid(current, vmParas, inAlias.(*iotago.AliasOutput), outAlias.(*iotago.AliasOutput), thisFoundryID); err != nil {
		return err
	}

	return nil
}

func foundrySerialNumberValid(current *iotago.FoundryOutput, vmParas *vm.Paras, inAlias *iotago.AliasOutput, outAlias *iotago.AliasOutput, thisFoundryID iotago.FoundryID) error {
	// this new foundry's serial number must be between the given foundry counter interval
	startSerial := inAlias.FoundryCounter
	endIncSerial := outAlias.FoundryCounter
	if startSerial >= current.SerialNumber || current.SerialNumber > endIncSerial {
		return fmt.Errorf("new foundry output %s's serial number is not between the foundry counter interval of [%d,%d)", thisFoundryID, startSerial, endIncSerial)
	}

	// OPTIMIZE: this loop happens on every STVF of every new foundry output
	// check order of serial number
	for outputIndex, output := range vmParas.WorkingSet.Tx.Essence.Outputs {
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

		if _, isNotNew := vmParas.WorkingSet.InChains[otherFoundryID]; isNotNew {
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
