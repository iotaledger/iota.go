package iotago

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/iotaledger/hive.go/serializer"
)

const (
	// OutputIDLength defines the length of an OutputID.
	OutputIDLength = TransactionIDLength + serializer.UInt16ByteSize
)

// OutputType defines the type of outputs.
type OutputType byte

const (
	// OutputSimple denotes a SimpleOutput
	OutputSimple OutputType = 0
	// OutputTreasury denotes the type of the TreasuryOutput.
	OutputTreasury OutputType = 2
	// OutputExtended denotes an ExtendedOutput.
	OutputExtended OutputType = 3
	// OutputAlias denotes an AliasOutput.
	OutputAlias OutputType = 4
	// OutputFoundry denotes a FoundryOutput.
	OutputFoundry OutputType = 5
	// OutputNFT denotes an NFTOutput.
	OutputNFT OutputType = 6
)

// OutputTypeToString returns the name of an Output given the type.
func OutputTypeToString(ty OutputType) string {
	switch ty {
	case OutputSimple:
		return "SimpleOutput"
	case OutputExtended:
		return "ExtendedOutput"
	case OutputTreasury:
		return "TreasuryOutput"
	case OutputAlias:
		return "AliasOutput"
	case OutputFoundry:
		return "FoundryOutput"
	case OutputNFT:
		return "NFTOutput"
	}
	return "unknown output"
}

var (
	// ErrDepositAmountMustBeGreaterThanZero returned if the deposit amount of an output is less or equal zero.
	ErrDepositAmountMustBeGreaterThanZero = errors.New("deposit amount must be greater than zero")
	// ErrMultiIdentOutputMismatch gets returned when MultiIdentOutput(s) aren't compatible.
	ErrMultiIdentOutputMismatch = errors.New("multi ident output mismatch")
	// ErrNonUniqueMultiIdentOutputs gets returned when multiple MultiIdentOutput(s) with the same ChainID exist within an OutputsByType.
	ErrNonUniqueMultiIdentOutputs = errors.New("non unique multi ident within outputs")
	// ErrChainMissing gets returned when a chain is missing.
	ErrChainMissing = errors.New("chain missing")
	// ErrNonUniqueChainConstrainedOutputs gets returned when multiple ChainConstrainedOutputs(s) with the same ChainID exist within sets.
	ErrNonUniqueChainConstrainedOutputs = errors.New("non unique chain constrained outputs")
	// ErrInvalidChainStateTransition gets returned when a state transition validation fails for a ChainConstrainedOutput.
	ErrInvalidChainStateTransition = errors.New("invalid chain state transition")
	// ErrTypeIsNotSupportedOutput gets returned when a serializable was found to not be a supported Output.
	ErrTypeIsNotSupportedOutput = errors.New("serializable is not a supported output")
)

// Outputs is a slice of Output.
type Outputs []Output

func (o Outputs) ToSerializables() serializer.Serializables {
	seris := make(serializer.Serializables, len(o))
	for i, x := range o {
		seris[i] = x.(serializer.Serializable)
	}
	return seris
}

func (o *Outputs) FromSerializables(seris serializer.Serializables) {
	*o = make(Outputs, len(seris))
	for i, seri := range seris {
		(*o)[i] = seri.(Output)
	}
}

// ChainConstrainedOutputSet returns a ChainConstrainedOutputsSet for all ChainConstrainedOutputs in Outputs.
func (o Outputs) ChainConstrainedOutputSet(txID TransactionID) ChainConstrainedOutputsSet {
	set := make(ChainConstrainedOutputsSet)
	for outputIndex, output := range o {
		chainConstrainedOutput, is := output.(ChainConstrainedOutput)
		if !is {
			continue
		}

		chainID := chainConstrainedOutput.Chain()
		if chainID.Empty() {
			if utxoIDChainID, is := chainConstrainedOutput.Chain().(UTXOIDChainID); is {
				chainID = utxoIDChainID.FromUTXOInputID(UTXOIDFromTransactionIDAndIndex(txID, uint16(outputIndex)))
			}
		}

		if chainID.Empty() {
			panic(fmt.Sprintf("output of type %s has empty chain ID but is not utxo dependable", OutputTypeToString(output.Type())))
		}

		set[chainID] = chainConstrainedOutput
	}
	return set
}

// ToOutputsByType converts the Outputs slice to OutputsByType.
func (o Outputs) ToOutputsByType() OutputsByType {
	outputsByType := make(OutputsByType)
	for _, output := range o {
		slice, has := outputsByType[output.Type()]
		if !has {
			slice = make(Outputs, 0)
		}
		outputsByType[output.Type()] = append(slice, output)
	}
	return outputsByType
}

// OutputsFilterFunc is a predicate function operating on an Output.
type OutputsFilterFunc func(output Output) bool

// OutputsFilterByType is an OutputsFilterFunc which filters Outputs by OutputType.
func OutputsFilterByType(ty OutputType) OutputsFilterFunc {
	return func(output Output) bool { return output.Type() == ty }
}

// Filter returns Outputs (retained order) passing the given OutputsFilterFunc.
func (o Outputs) Filter(f OutputsFilterFunc) Outputs {
	filtered := make(Outputs, 0)
	for _, output := range o {
		if !f(output) {
			continue
		}
		filtered = append(filtered, output)
	}
	return filtered
}

// OutputsByType is a map of OutputType(s) to slice of Output(s).
type OutputsByType map[OutputType][]Output

// NativeTokenOutputs returns a slice of Outputs which are NativeTokenOutput.
func (outputs OutputsByType) NativeTokenOutputs() NativeTokenOutputs {
	nativeTokenOutputs := make(NativeTokenOutputs, 0)
	for _, slice := range outputs {
		for _, output := range slice {
			nativeTokenOutput, is := output.(NativeTokenOutput)
			if !is {
				continue
			}
			nativeTokenOutputs = append(nativeTokenOutputs, nativeTokenOutput)
		}
	}
	return nativeTokenOutputs
}

// MultiIdentOutputs returns a slice of Outputs which are MultiIdentOutput.
func (outputs OutputsByType) MultiIdentOutputs() MultiIdentOutputs {
	multiIdentOutputs := make(MultiIdentOutputs, 0)
	for _, slice := range outputs {
		for _, output := range slice {
			multiIdentOutput, is := output.(MultiIdentOutput)
			if !is {
				continue
			}
			multiIdentOutputs = append(multiIdentOutputs, multiIdentOutput)
		}
	}
	return multiIdentOutputs
}

// MultiIdentOutputsSet returns a map of ChainID to MultiIdentOutput.
// If multiple MultiIdentOutput(s) exist for a given ChainID, an error is returned.
// Empty AccountIDs are ignored.
func (outputs OutputsByType) MultiIdentOutputsSet() (MultiIdentOutputsSet, error) {
	multiIdentOutputsSet := make(MultiIdentOutputsSet, 0)
	for _, output := range outputs.MultiIdentOutputs() {
		if output.Chain().Empty() {
			continue
		}
		if _, has := multiIdentOutputsSet[output.Chain()]; has {
			return nil, ErrNonUniqueMultiIdentOutputs
		}
		multiIdentOutputsSet[output.Chain()] = output
	}
	return multiIdentOutputsSet, nil
}

// FoundryOutputs returns a slice of Outputs which are FoundryOutput.
func (outputs OutputsByType) FoundryOutputs() FoundryOutputs {
	foundryOutputs := make(FoundryOutputs, 0)
	for _, output := range outputs[OutputFoundry] {
		foundryOutput, is := output.(*FoundryOutput)
		if !is {
			continue
		}
		foundryOutputs = append(foundryOutputs, foundryOutput)
	}
	return foundryOutputs
}

// FoundryOutputsSet returns a map of FoundryID to FoundryOutput.
// If multiple FoundryOutput(s) exist for a given FoundryID, an error is returned.
func (outputs OutputsByType) FoundryOutputsSet() (FoundryOutputsSet, error) {
	foundryOutputsSet := make(FoundryOutputsSet, 0)
	for _, output := range outputs[OutputFoundry] {
		foundryOutput, is := output.(*FoundryOutput)
		if !is {
			continue
		}
		foundryID, err := foundryOutput.ID()
		if err != nil {
			return nil, err
		}
		if _, has := foundryOutputsSet[foundryID]; has {
			return nil, ErrNonUniqueFoundryOutputs
		}
		foundryOutputsSet[foundryID] = foundryOutput
	}
	return foundryOutputsSet, nil
}

// AliasOutputs returns a slice of Outputs which are AliasOutput.
func (outputs OutputsByType) AliasOutputs() AliasOutputs {
	aliasOutputs := make(AliasOutputs, 0)
	for _, output := range outputs[OutputFoundry] {
		aliasOutput, is := output.(*AliasOutput)
		if !is {
			continue
		}
		aliasOutputs = append(aliasOutputs, aliasOutput)
	}
	return aliasOutputs
}

// NonNewAliasOutputsSet returns a map of AliasID to AliasOutput.
// If multiple AliasOutput(s) exist for a given AliasID, an error is returned.
// The produced set does not include AliasOutputs of which their AliasID are zeroed.
func (outputs OutputsByType) NonNewAliasOutputsSet() (AliasOutputsSet, error) {
	aliasOutputsSet := make(AliasOutputsSet, 0)
	for _, output := range outputs[OutputAlias] {
		aliasOutput, is := output.(*AliasOutput)
		if !is || aliasOutput.AliasEmpty() {
			continue
		}
		if _, has := aliasOutputsSet[aliasOutput.AliasID]; has {
			return nil, ErrNonUniqueAliasOutputs
		}
		aliasOutputsSet[aliasOutput.AliasID] = aliasOutput
	}
	return aliasOutputsSet, nil
}

// ChainConstrainedOutputsSet returns a map of ChainID to ChainConstrainedOutput.
// If multiple ChainConstrainedOutput(s) exist for a given ChainID, an error is returned.
func (outputs OutputsByType) ChainConstrainedOutputsSet() (ChainConstrainedOutputsSet, error) {
	chainConstrainedOutputs := make(ChainConstrainedOutputsSet, 0)
	for _, ty := range []OutputType{OutputAlias, OutputFoundry, OutputNFT} {
		for _, output := range outputs[ty] {
			chainConstrainedOutput, is := output.(ChainConstrainedOutput)
			if !is || chainConstrainedOutput.Chain().Empty() {
				continue
			}
			if _, has := chainConstrainedOutputs[chainConstrainedOutput.Chain()]; has {
				return nil, ErrNonUniqueChainConstrainedOutputs
			}
			chainConstrainedOutputs[chainConstrainedOutput.Chain()] = chainConstrainedOutput
		}
	}
	return chainConstrainedOutputs, nil
}

// ChainConstrainedOutputs returns a slice of Outputs which are ChainConstrainedOutput.
func (outputs OutputsByType) ChainConstrainedOutputs() ChainConstrainedOutputs {
	chainConstrainedOutputs := make(ChainConstrainedOutputs, 0)
	for _, ty := range []OutputType{OutputAlias, OutputFoundry, OutputNFT} {
		for _, output := range outputs[ty] {
			chainConstrainedOutput, is := output.(ChainConstrainedOutput)
			if !is {
				continue
			}
			chainConstrainedOutputs = append(chainConstrainedOutputs, chainConstrainedOutput)
		}
	}
	return chainConstrainedOutputs
}

// NativeTokenOutputs is a slice of NativeTokenOutput(s).
type NativeTokenOutputs []NativeTokenOutput

// Sum sums up the different NativeTokens occurring within the given outputs.
func (ntOutputs NativeTokenOutputs) Sum() (NativeTokenSum, error) {
	sum := make(map[NativeTokenID]*big.Int)
	for _, output := range ntOutputs {
		for _, nativeToken := range output.NativeTokenSet() {
			if sign := nativeToken.Amount.Sign(); sign == -1 || sign == 0 {
				return nil, ErrNativeTokenAmountLessThanEqualZero
			}

			val := sum[nativeToken.ID]
			if val == nil {
				val = new(big.Int)
			}

			if val.Add(val, nativeToken.Amount).Cmp(abi.MaxUint256) == 1 {
				return nil, ErrNativeTokenSumExceedsUint256
			}
			sum[nativeToken.ID] = val
		}
	}
	return sum, nil
}

// InputSet maps inputs to their origin UTXOs.
type InputSet map[UTXOInputID]Output

// NewAliases returns an AliasOutputsSet for all AliasOutputs which are new.
func (inputSet InputSet) NewAliases() AliasOutputsSet {
	set := make(AliasOutputsSet)
	for utxoInputID, output := range inputSet {
		aliasOutput, is := output.(*AliasOutput)
		if !is || !aliasOutput.AliasEmpty() {
			continue
		}
		set[AliasIDFromOutputID(utxoInputID)] = aliasOutput
	}
	return set
}

// ChainConstrainedOutputSet returns a ChainConstrainedOutputsSet for all ChainConstrainedOutputs in the InputSet.
func (inputSet InputSet) ChainConstrainedOutputSet() ChainConstrainedOutputsSet {
	set := make(ChainConstrainedOutputsSet)
	for utxoInputID, output := range inputSet {
		chainConstrainedOutput, is := output.(ChainConstrainedOutput)
		if !is {
			continue
		}

		chainID := chainConstrainedOutput.Chain()
		if chainID.Empty() {
			if utxoIDChainID, is := chainConstrainedOutput.Chain().(UTXOIDChainID); is {
				chainID = utxoIDChainID.FromUTXOInputID(utxoInputID)
			}
		}

		if chainID.Empty() {
			panic(fmt.Sprintf("output of type %s has empty chain ID but is not utxo dependable", OutputTypeToString(output.Type())))
		}

		set[chainID] = chainConstrainedOutput
	}
	return set
}

// Output defines a unit of output of a transaction.
type Output interface {
	serializer.Serializable
	NonEphemeralObject

	// Deposit returns the amount this Output deposits.
	Deposit() (uint64, error)
	// Type returns the type of the output.
	Type() OutputType
}

// SingleIdentOutput is a type of Output where without considering its FeatureBlocks,
// only one identity needs to be unlocked.
type SingleIdentOutput interface {
	Output
	// Ident returns the identity to which this output is locked to.
	Ident() (Address, error)
}

// MultiIdentOutputsSet is a set of MultiIdentOutput(s).
type MultiIdentOutputsSet map[ChainID]MultiIdentOutput

// MultiIdentOutputs is a slice of MultiIdentOutput(s).
type MultiIdentOutputs []MultiIdentOutput

// MultiIdentOutput is a type of Output which multiple identities can control/modify.
// Unlike the SingleIdentOutput, the MultiIdentOutput's to unlock identity is dependent
// on the transition the output does between inputs and outputs.
type MultiIdentOutput interface {
	ChainConstrainedOutput
	// Ident computes the identity to which this output is locked to by examining
	// the transition to the next output state.
	// Note that it is the caller's job to ensure that the given other MultiIdentOutput
	// corresponds to this MultiIdentOutput.
	// If this MultiIdentOutput is not dependent on a transition to compute the ident,
	// nil can be passed as an argument.
	Ident(nextState MultiIdentOutput) (Address, error)
}

// NativeTokenOutput is a type of Output which can hold NativeToken.
type NativeTokenOutput interface {
	Output
	// NativeTokenSet returns the NativeToken this output defines.
	NativeTokenSet() NativeTokens
}

// FeatureBlockOutput is a type of Output which can hold FeatureBlock.
type FeatureBlockOutput interface {
	// FeatureBlocks returns the feature blocks this output defines.
	FeatureBlocks() FeatureBlocks
}

// OutputSelector implements SerializableSelectorFunc for output types.
func OutputSelector(outputType uint32) (serializer.Serializable, error) {
	var seri serializer.Serializable
	switch OutputType(outputType) {
	case OutputSimple:
		seri = &SimpleOutput{}
	case OutputExtended:
		seri = &ExtendedOutput{}
	case OutputTreasury:
		seri = &TreasuryOutput{}
	case OutputAlias:
		seri = &AliasOutput{}
	case OutputFoundry:
		seri = &FoundryOutput{}
	case OutputNFT:
		seri = &NFTOutput{}
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownOutputType, outputType)
	}
	return seri, nil
}

// OutputIDHex is the hex representation of an output ID.
type OutputIDHex string

// MustSplitParts returns the transaction ID and output index parts of the hex output ID.
// It panics if the hex output ID is invalid.
func (oih OutputIDHex) MustSplitParts() (*TransactionID, uint16) {
	txID, outputIndex, err := oih.SplitParts()
	if err != nil {
		panic(err)
	}
	return txID, outputIndex
}

// SplitParts returns the transaction ID and output index parts of the hex output ID.
func (oih OutputIDHex) SplitParts() (*TransactionID, uint16, error) {
	outputIDBytes, err := hex.DecodeString(string(oih))
	if err != nil {
		return nil, 0, err
	}
	var txID TransactionID
	copy(txID[:], outputIDBytes[:TransactionIDLength])
	outputIndex := binary.LittleEndian.Uint16(outputIDBytes[TransactionIDLength : TransactionIDLength+serializer.UInt16ByteSize])
	return &txID, outputIndex, nil
}

// MustAsUTXOInput converts the hex output ID to a UTXOInput.
// It panics if the hex output ID is invalid.
func (oih OutputIDHex) MustAsUTXOInput() *UTXOInput {
	utxoInput, err := oih.AsUTXOInput()
	if err != nil {
		panic(err)
	}
	return utxoInput
}

// AsUTXOInput converts the hex output ID to a UTXOInput.
func (oih OutputIDHex) AsUTXOInput() (*UTXOInput, error) {
	var utxoInput UTXOInput
	txID, outputIndex, err := oih.SplitParts()
	if err != nil {
		return nil, err
	}
	copy(utxoInput.TransactionID[:], txID[:])
	utxoInput.TransactionOutputIndex = outputIndex
	return &utxoInput, nil
}

// OutputsSyntacticalValidationFunc which given the index of an output and the output itself, runs syntactical validations and returns an error if any should fail.
type OutputsSyntacticalValidationFunc func(index int, output Output) error

// OutputsSyntacticalDepositAmount returns an OutputsSyntacticalValidationFunc which checks that:
//	- every output deposits more than zero
//	- every output deposits less than the total supply
//	- the sum of deposits does not exceed the total supply
//	- the deposit fulfils the minimum deposit as calculated from the virtual byte cost of the output
//	- if the output contains a DustDepositReturnFeatureBlock, it must "return" bigger equal than the minimum dust deposit
//	  and must be less equal the minimum virtual byte rent cost for the output.
// If -1 is passed to the validator func, then the sum is not aggregated over multiple calls.
func OutputsSyntacticalDepositAmount(minDustDep uint64, costStruct *RentStructure) OutputsSyntacticalValidationFunc {
	var sum uint64
	return func(index int, output Output) error {
		deposit, err := output.Deposit()
		if err != nil {
			return fmt.Errorf("unable to get deposit of output: %w", err)
		}

		switch {
		case deposit == 0:
			return fmt.Errorf("%w: output %d", ErrDepositAmountMustBeGreaterThanZero, index)
		case deposit > TokenSupply:
			return fmt.Errorf("%w: output %d", ErrOutputDepositsMoreThanTotalSupply, index)
		case sum+deposit > TokenSupply:
			return fmt.Errorf("%w: output %d", ErrOutputsSumExceedsTotalSupply, index)
		}

		minRent, err := costStruct.CoversStateRent(output, deposit)
		if err != nil {
			return fmt.Errorf("%w: output %d", err, index)
		}

		if featureBlockOutput, is := output.(FeatureBlockOutput); is {
			featBlockSet, err := featureBlockOutput.FeatureBlocks().Set()
			if err != nil {
				return fmt.Errorf("unable to compute feature block set in deposit syntactic checks for output %d: %w", index, err)
			}
			if returnFeatBlock := featBlockSet[FeatureBlockDustDepositReturn]; returnFeatBlock != nil {
				returnAmount := returnFeatBlock.(*DustDepositReturnFeatureBlock).Amount
				switch {
				case returnAmount > minRent:
					return fmt.Errorf("%w: output %d", ErrOutputReturnBlockIsMoreThanVBRent, index)
				case returnAmount < minDustDep:
					return fmt.Errorf("%w: output %d", ErrOutputReturnBlockIsLessThanMinDust, index)
				}
			}
		}

		if index != -1 {
			sum += deposit
		}
		return nil
	}
}

// OutputsSyntacticalNativeTokensCount returns an OutputsSyntacticalValidationFunc which checks that:
//	- the sum of native tokens count across all outputs does not exceed MaxNativeTokensCount
func OutputsSyntacticalNativeTokensCount() OutputsSyntacticalValidationFunc {
	var nativeTokensCount int
	return func(index int, output Output) error {
		if nativeTokenOutput, is := output.(NativeTokenOutput); is {
			nativeTokensCount += len(nativeTokenOutput.NativeTokenSet())
			if nativeTokensCount > MaxNativeTokensCount {
				return ErrOutputsExceedMaxNativeTokensCount
			}
		}
		return nil
	}
}

// OutputsSyntacticalSenderFeatureBlockRequirement returns an OutputsSyntacticalValidationFunc which checks that:
//	- if an output contains a SenderFeatureBlock if another FeatureBlock (example DustDepositReturnFeatureBlock) requires it
func OutputsSyntacticalSenderFeatureBlockRequirement() OutputsSyntacticalValidationFunc {
	return func(index int, output Output) error {
		featureBlockOutput, is := output.(FeatureBlockOutput)
		if !is {
			return nil
		}
		var hasReturnFeatBlock, hasExpMsFeatBlock, hasExpUnixFeatBlock, hasSenderFeatBlock bool
		for _, featureBlock := range featureBlockOutput.FeatureBlocks() {
			switch featureBlock.(type) {
			case *DustDepositReturnFeatureBlock:
				hasReturnFeatBlock = true
			case *ExpirationMilestoneIndexFeatureBlock:
				hasExpMsFeatBlock = true
			case *ExpirationUnixFeatureBlock:
				hasExpUnixFeatBlock = true
			case *SenderFeatureBlock:
				hasSenderFeatBlock = true
			}
		}
		if (hasReturnFeatBlock || hasExpMsFeatBlock || hasExpUnixFeatBlock) && !hasSenderFeatBlock {
			return fmt.Errorf("%w: output %d", ErrOutputRequiresSenderFeatureBlock, index)
		}
		return nil
	}
}

// OutputsSyntacticalAlias returns an OutputsSyntacticalValidationFunc which checks that AliasOutput(s)':
//	- StateIndex/FoundryCounter are zero if the AliasID is zeroed
//	- StateController and GovernanceController must be different from AliasAddress derived from AliasID
func OutputsSyntacticalAlias(txID *TransactionID) OutputsSyntacticalValidationFunc {
	return func(index int, output Output) error {
		aliasOutput, is := output.(*AliasOutput)
		if !is {
			return nil
		}

		var outputAliasAddr AliasAddress
		if aliasOutput.AliasEmpty() {
			switch {
			case aliasOutput.StateIndex != 0:
				return fmt.Errorf("%w: output %d, state index not zero", ErrAliasOutputNonEmptyState, index)
			case aliasOutput.FoundryCounter != 0:
				return fmt.Errorf("%w: output %d, foundry counter not zero", ErrAliasOutputNonEmptyState, index)
			}

			// build AliasID using the transaction ID
			outputAliasAddr = AliasAddressFromOutputID(UTXOIDFromTransactionIDAndIndex(*txID, uint16(index)))
		}

		if outputAliasAddr == emptyAliasAddress {
			copy(outputAliasAddr[:], aliasOutput.AliasID[:])
		}

		if stateCtrlAddr, ok := aliasOutput.StateController.(*AliasAddress); ok && outputAliasAddr == *stateCtrlAddr {
			return fmt.Errorf("%w: output %d, AliasID=StateController", ErrAliasOutputCyclicAddress, index)
		}
		if govCtrlAddr, ok := aliasOutput.GovernanceController.(*AliasAddress); ok && outputAliasAddr == *govCtrlAddr {
			return fmt.Errorf("%w: output %d, AliasID=GovernanceController", ErrAliasOutputCyclicAddress, index)
		}

		return nil
	}
}

// OutputsSyntacticalFoundry returns an OutputsSyntacticalValidationFunc which checks that FoundryOutput(s)':
//	- CirculatingSupply is less equal MaximumSupply
//	- MaximumSupply is not zero
func OutputsSyntacticalFoundry() OutputsSyntacticalValidationFunc {
	return func(index int, output Output) error {
		foundryOutput, is := output.(*FoundryOutput)
		if !is {
			return nil
		}

		if r := foundryOutput.MaximumSupply.Cmp(new(big.Int).SetInt64(0)); r == -1 || r == 0 {
			return fmt.Errorf("%w: output %d, less than equal zero", ErrFoundryOutputInvalidMaximumSupply, index)
		}

		if r := foundryOutput.CirculatingSupply.Cmp(foundryOutput.MaximumSupply); r == 1 {
			return fmt.Errorf("%w: output %d, bigger than maximum supply", ErrFoundryOutputInvalidCirculatingSupply, index)
		}

		return nil
	}
}

// OutputsSyntacticalNFT returns an OutputsSyntacticalValidationFunc which checks that NFTOutput(s)':
//	- Address must be different from NFTAddress derived from NFTID
func OutputsSyntacticalNFT(txID *TransactionID) OutputsSyntacticalValidationFunc {
	return func(index int, output Output) error {
		nftOutput, is := output.(*NFTOutput)
		if !is {
			return nil
		}

		var outputNFTAddr NFTAddress
		if nftOutput.NFTID == emptyNFTID {
			outputNFTAddr = NFTAddressFromOutputID(UTXOIDFromTransactionIDAndIndex(*txID, uint16(index)))
		}

		if outputNFTAddr == emptyNFTAddress {
			copy(outputNFTAddr[:], nftOutput.NFTID[:])
		}

		if addr, ok := nftOutput.Address.(*NFTAddress); ok && outputNFTAddr == *addr {
			return fmt.Errorf("%w: output %d, AliasID=StateController", ErrNFTOutputCyclicAddress, index)
		}

		return nil
	}
}

// ValidateOutputs validates the outputs by running them against the given OutputsSyntacticalValidationFunc(s).
func ValidateOutputs(outputs Outputs, funcs ...OutputsSyntacticalValidationFunc) error {
	for i, output := range outputs {
		for _, f := range funcs {
			if err := f(i, output); err != nil {
				return err
			}
		}
	}
	return nil
}

// jsonOutputSelector selects the json output implementation for the given type.
func jsonOutputSelector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch OutputType(ty) {
	case OutputSimple:
		obj = &jsonSimpleOutput{}
	case OutputExtended:
		obj = &jsonExtendedOutput{}
	case OutputTreasury:
		obj = &jsonTreasuryOutput{}
	case OutputAlias:
		obj = &jsonAliasOutput{}
	case OutputFoundry:
		obj = &jsonFoundryOutput{}
	case OutputNFT:
		obj = &jsonNFTOutput{}
	default:
		return nil, fmt.Errorf("unable to decode output type from JSON: %w", ErrUnknownOutputType)
	}
	return obj, nil
}
