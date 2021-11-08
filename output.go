package iotago

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/iotaledger/hive.go/serializer"
)

const (
	// OutputIDLength defines the length of an OutputID.
	OutputIDLength = TransactionIDLength + serializer.UInt16ByteSize
)

// OutputType defines the type of outputs.
type OutputType byte

const (
	// OutputSimple denotes a type of output which is locked by a signature and deposits onto a single address.
	OutputSimple OutputType = iota
	// OutputExtended denotes a type of output which can also hold native tokens and feature blocks.
	OutputExtended
	// OutputTreasury denotes the type of the TreasuryOutput.
	OutputTreasury
	// OutputAlias denotes the type of an AliasOutput.
	OutputAlias
	// OutputFoundry denotes the type of a FoundryOutput.
	OutputFoundry
	// OutputNFT denotes the type of an NFTOutput.
	OutputNFT
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

// Output defines the deposit of funds.
type Output interface {
	serializer.Serializable

	// Deposit returns the amount this Output deposits.
	Deposit() (uint64, error)
	// Target returns the target of the deposit.
	// If the type of output does not have/support a target, nil is returned.
	Target() (serializer.Serializable, error)
	// Type returns the type of the output.
	Type() OutputType
}

// NativeTokenOutput is a type of Output which also can hold NativeToken.
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

// OutputsPredicateFunc which given the index of an output and the output itself, runs validations and returns an error if any should fail.
type OutputsPredicateFunc func(index int, output Output) error

// OutputsPredicateAddrUnique returns an OutputsPredicateFunc which checks that all addresses are unique per OutputType.
// Deprecated: an output set no longer needs to hold unique addresses per output.
func OutputsPredicateAddrUnique() OutputsPredicateFunc {
	set := map[OutputType]map[string]int{}
	return func(index int, dep Output) error {
		var b strings.Builder

		target, err := dep.Target()
		if err != nil {
			return fmt.Errorf("unable to get target of output: %w", err)
		}

		if target == nil {
			return nil
		}

		// can't be reduced to one b.Write()
		switch addr := target.(type) {
		case *Ed25519Address:
			if _, err := b.Write(addr[:]); err != nil {
				return fmt.Errorf("%w: unable to serialize Ed25519 address in addr unique validator", err)
			}
		}

		k := b.String()

		m, ok := set[dep.Type()]
		if !ok {
			m = make(map[string]int)
			set[dep.Type()] = m
		}

		if j, has := m[k]; has {
			return fmt.Errorf("%w: output %d and %d share the same address", ErrOutputAddrNotUnique, j, index)
		}
		m[k] = index
		return nil
	}
}

// OutputsPredicateDepositAmount returns an OutputsPredicateFunc which checks that:
//	- every output deposits more than zero
//	- every output deposits less than the total supply
//	- the sum of deposits does not exceed the total supply
// If -1 is passed to the validator func, then the sum is not aggregated over multiple calls.
func OutputsPredicateDepositAmount() OutputsPredicateFunc {
	var sum uint64
	return func(index int, dep Output) error {
		deposit, err := dep.Deposit()
		if err != nil {
			return fmt.Errorf("unable to get deposit of output: %w", err)
		}
		if deposit == 0 {
			return fmt.Errorf("%w: output %d", ErrDepositAmountMustBeGreaterThanZero, index)
		}
		if deposit > TokenSupply {
			return fmt.Errorf("%w: output %d", ErrOutputDepositsMoreThanTotalSupply, index)
		}
		if sum+deposit > TokenSupply {
			return fmt.Errorf("%w: output %d", ErrOutputsSumExceedsTotalSupply, index)
		}
		if index != -1 {
			sum += deposit
		}
		return nil
	}
}

// OutputsPredicateNativeTokensCount returns an OutputsPredicateFunc which checks that:
//	- the sum of native tokens count across all outputs does not exceed MaxNativeTokensCount
func OutputsPredicateNativeTokensCount() OutputsPredicateFunc {
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

// OutputsPredicateSenderFeatureBlockRequirement returns an OutputsPredicateFunc which checks that:
//	- if an output contains a SenderFeatureBlock if another FeatureBlock (example ReturnFeatureBlock) requires it
func OutputsPredicateSenderFeatureBlockRequirement() OutputsPredicateFunc {
	return func(index int, output Output) error {
		featureBlockOutput, is := output.(FeatureBlockOutput)
		if !is {
			return nil
		}
		var hasReturnFeatBlock, hasExpMsFeatBlock, hasExpUnixFeatBlock, hasSenderFeatBlock bool
		for _, featureBlock := range featureBlockOutput.FeatureBlocks() {
			switch featureBlock.(type) {
			case *ReturnFeatureBlock:
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

// OutputsPredicateAlias returns an OutputsPredicateFunc which checks that AliasOutput(s)':
//	- StateIndex/FoundryCounter are zero if the AliasID is zeroed
//	- StateController and GovernanceController must be different from AliasAddress derived from AliasID
func OutputsPredicateAlias(txID *TransactionID) OutputsPredicateFunc {
	return func(index int, output Output) error {
		aliasOutput, is := output.(*AliasOutput)
		if !is {
			return nil
		}

		var outputAliasAddr AliasAddress
		if aliasOutput.AliasID == emptyAliasID {
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

// OutputsPredicateFoundry returns an OutputsPredicateFunc which checks that FoundryOutput(s)':
//	- CirculatingSupply is less equal MaximumSupply
//	- MaximumSupply is not zero
func OutputsPredicateFoundry() OutputsPredicateFunc {
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

// OutputsPredicateNFT returns an OutputsPredicateFunc which checks that NFTOutput(s)':
//	- Address must be different from NFTAddress derived from NFTID
func OutputsPredicateNFT(txID *TransactionID) OutputsPredicateFunc {
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

// supposed to be called with -1 as input in order to be used over multiple calls.
var outputAmountValidator = OutputsPredicateDepositAmount()

// ValidateOutputs validates the outputs by running them against the given OutputsPredicateFunc(s).
func ValidateOutputs(outputs Outputs, funcs ...OutputsPredicateFunc) error {
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
