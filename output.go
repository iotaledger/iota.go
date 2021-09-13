package iotago

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/iotaledger/hive.go/serializer"
	"strings"
)

// OutputType defines the type of outputs.
type OutputType = byte

const (
	// OutputSigLockedSingleOutput denotes a type of output which is locked by a signature and deposits onto a single address.
	OutputSigLockedSingleOutput OutputType = iota
	// OutputSigLockedDustAllowanceOutput is like OutputSigLockedSingleOutput but it is used to increase the allowance/amount of dust outputs on a given address.
	OutputSigLockedDustAllowanceOutput
	// OutputTreasuryOutput denotes the type of the TreasuryOutput.
	OutputTreasuryOutput
	// OutputSigLockedDustAllowanceOutputMinDeposit defines the minimum deposit amount of a SigLockedDustAllowanceOutput.
	OutputSigLockedDustAllowanceOutputMinDeposit uint64 = 1_000_000
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

// OutputSelector implements SerializableSelectorFunc for output types.
func OutputSelector(outputType uint32) (serializer.Serializable, error) {
	var seri serializer.Serializable
	switch byte(outputType) {
	case OutputSigLockedSingleOutput:
		seri = &SigLockedSingleOutput{}
	case OutputSigLockedDustAllowanceOutput:
		seri = &SigLockedDustAllowanceOutput{}
	case OutputTreasuryOutput:
		seri = &TreasuryOutput{}
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

// OutputsValidatorFunc which given the index of an output and the output itself, runs validations and returns an error if any should fail.
type OutputsValidatorFunc func(index int, output Output) error

// OutputsAddrUniqueValidator returns a validator which checks that all addresses are unique per OutputType.
func OutputsAddrUniqueValidator() OutputsValidatorFunc {
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

// OutputsDepositAmountValidator returns a validator which checks that:
//	1. every output deposits more than zero
//	2. every output deposits less than the total supply
//	3. the sum of deposits does not exceed the total supply
//	4. SigLockedDustAllowanceOutput deposits at least OutputSigLockedDustAllowanceOutputMinDeposit.
// If -1 is passed to the validator func, then the sum is not aggregated over multiple calls.
func OutputsDepositAmountValidator() OutputsValidatorFunc {
	var sum uint64
	return func(index int, dep Output) error {
		deposit, err := dep.Deposit()
		if err != nil {
			return fmt.Errorf("unable to get deposit of output: %w", err)
		}
		if deposit == 0 {
			return fmt.Errorf("%w: output %d", ErrDepositAmountMustBeGreaterThanZero, index)
		}
		if _, isAllowanceOutput := dep.(*SigLockedDustAllowanceOutput); isAllowanceOutput {
			if deposit < OutputSigLockedDustAllowanceOutputMinDeposit {
				return fmt.Errorf("%w: output %d", ErrOutputDustAllowanceLessThanMinDeposit, index)
			}
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

// supposed to be called with -1 as input in order to be used over multiple calls.
var outputAmountValidator = OutputsDepositAmountValidator()

// ValidateOutputs validates the outputs by running them against the given OutputsValidatorFunc.
func ValidateOutputs(outputs serializer.Serializables, funcs ...OutputsValidatorFunc) error {
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
	case OutputSigLockedSingleOutput:
		obj = &jsonSigLockedSingleOutput{}
	case OutputSigLockedDustAllowanceOutput:
		obj = &jsonSigLockedDustAllowanceOutput{}
	case OutputTreasuryOutput:
		obj = &jsonTreasuryOutput{}
	default:
		return nil, fmt.Errorf("unable to decode output type from JSON: %w", ErrUnknownOutputType)
	}
	return obj, nil
}
