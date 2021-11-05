package iotago

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/iotaledger/hive.go/serializer"
)

const (
	// OutputIDLength defines the length of an OutputID.
	OutputIDLength = TransactionIDLength + serializer.UInt16ByteSize
)

// OutputID defines an identifier of an Output which consists of the hash of the transaction
// containing the output and the positioning index of that output within that transaction.
type OutputID [OutputIDLength]byte

// OutputType defines the type of outputs.
type OutputType = byte

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

var (
	// ErrDepositAmountMustBeGreaterThanZero returned if the deposit amount of an output is less or equal zero.
	ErrDepositAmountMustBeGreaterThanZero = errors.New("deposit amount must be greater than zero")
)

// Outputs is a slice of Output.
type Outputs []Output

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
	NativeTokenSet() serializer.Serializables
}

// FeatureBlockOutput is a type of Output which can hold FeatureBlock.
type FeatureBlockOutput interface {
	// FeatureBlocks returns the feature blocks this output defines.
	FeatureBlocks() serializer.Serializables
}

// OutputSelector implements SerializableSelectorFunc for output types.
func OutputSelector(outputType uint32) (serializer.Serializable, error) {
	var seri serializer.Serializable
	switch byte(outputType) {
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

// supposed to be called with -1 as input in order to be used over multiple calls.
var outputAmountValidator = OutputsPredicateDepositAmount()

// ValidateOutputs validates the outputs by running them against the given OutputsPredicateFunc(s).
func ValidateOutputs(outputs serializer.Serializables, funcs ...OutputsPredicateFunc) error {
	for i, output := range outputs {
		if _, isOutput := output.(Output); !isOutput {
			return fmt.Errorf("%w: can only validate outputs but got %T instead", ErrUnknownOutputType, output)
		}
		for _, f := range funcs {
			if err := f(i, output.(Output)); err != nil {
				return err
			}
		}
	}
	return nil
}

// jsonOutputSelector selects the json output implementation for the given type.
func jsonOutputSelector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch byte(ty) {
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
