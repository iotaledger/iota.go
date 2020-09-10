package iota

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

// Defines the type of outputs.
type OutputType = byte

const (
	// Denotes a type of output which is locked by a signature and deposits onto a single address.
	OutputSigLockedSingleDeposit OutputType = iota

	// The size of a sig locked single deposit containing a WOTS address as its deposit address.
	SigLockedSingleDepositWOTSAddrBytesSize = SmallTypeDenotationByteSize + WOTSAddressSerializedBytesSize + UInt64ByteSize
	// The size of a sig locked single deposit containing an Ed25519 address as its deposit address.
	SigLockedSingleDepositEd25519AddrBytesSize = SmallTypeDenotationByteSize + Ed25519AddressSerializedBytesSize + UInt64ByteSize

	// Defines the minimum size a sig locked single deposit must be.
	SigLockedSingleDepositBytesMinSize = SigLockedSingleDepositEd25519AddrBytesSize
	// Defines the offset at which the address portion within a sig locked single deposit begins.
	SigLockedSingleDepositAddressOffset = SmallTypeDenotationByteSize
)

var (
	// Returned if the deposit amount of an output is less or equal zero.
	ErrDepositAmountMustBeGreaterThanZero = errors.New("deposit amount must be greater than zero")
)

// OutputSelector implements SerializableSelectorFunc for output types.
func OutputSelector(outputType uint32) (Serializable, error) {
	var seri Serializable
	switch byte(outputType) {
	case OutputSigLockedSingleDeposit:
		seri = &SigLockedSingleDeposit{}
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownOutputType, outputType)
	}
	return seri, nil
}

// SigLockedSingleDeposit is an output type which can be unlocked via a signature. It deposits onto one single address.
type SigLockedSingleDeposit struct {
	// The actual address.
	Address Serializable `json:"address"`
	// The amount to deposit.
	Amount uint64 `json:"amount"`
}

func (s *SigLockedSingleDeposit) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkMinByteLength(SigLockedSingleDepositBytesMinSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid signature locked single deposit bytes: %w", err)
		}
		if err := checkTypeByte(data, OutputSigLockedSingleDeposit); err != nil {
			return 0, fmt.Errorf("unable to deserialize signature locked single deposit: %w", err)
		}
	}

	data = data[SmallTypeDenotationByteSize:]
	addr, addrBytesRead, err := DeserializeObject(data, deSeriMode, TypeDenotationByte, AddressSelector)
	if err != nil {
		return 0, err
	}
	s.Address = addr

	data = data[addrBytesRead:]

	// read amount of the deposit
	s.Amount = binary.LittleEndian.Uint64(data)

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := outputAmountValidator(-1, s); err != nil {
			return 0, fmt.Errorf("%w: unable to deserialize signature locked single deposit", err)
		}
	}

	return SmallTypeDenotationByteSize + addrBytesRead + UInt64ByteSize, nil
}

func (s *SigLockedSingleDeposit) Serialize(deSeriMode DeSerializationMode) (data []byte, err error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := outputAmountValidator(-1, s); err != nil {
			return nil, fmt.Errorf("%w: unable to serialize signature locked single deposit", err)
		}
	}

	var b []byte
	switch s.Address.(type) {
	case *WOTSAddress:
		b = make([]byte, SigLockedSingleDepositWOTSAddrBytesSize)
	case *Ed25519Address:
		b = make([]byte, SigLockedSingleDepositEd25519AddrBytesSize)
	default:
		return nil, fmt.Errorf("%w: signature locked single deposit defines unknown address", ErrUnknownAddrType)
	}

	b[0] = OutputSigLockedSingleDeposit
	addrBytes, err := s.Address.Serialize(deSeriMode)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to serialize address within signature locked single deposit", err)
	}
	copy(b[SmallTypeDenotationByteSize:], addrBytes)
	binary.LittleEndian.PutUint64(b[len(b)-UInt64ByteSize:], s.Amount)
	return b, nil
}

// OutputsValidatorFunc which given the index of an output and the output itself, runs validations and returns an error if any should fail.
type OutputsValidatorFunc func(index int, output *SigLockedSingleDeposit) error

// OutputsAddrUniqueValidator returns a validator which checks that all addresses are unique.
func OutputsAddrUniqueValidator() OutputsValidatorFunc {
	set := map[string]int{}
	return func(index int, dep *SigLockedSingleDeposit) error {
		var b strings.Builder
		// can't be reduced to one b.Write()
		switch addr := dep.Address.(type) {
		case *WOTSAddress:
			if _, err := b.Write(addr[:]); err != nil {
				return fmt.Errorf("%w: unable to serialize WOTS address in addr unique validator", err)
			}
		case *Ed25519Address:
			if _, err := b.Write(addr[:]); err != nil {
				return fmt.Errorf("%w: unable to serialize Ed25519 address in addr unique validator", err)
			}
		}
		k := b.String()
		if j, has := set[k]; has {
			return fmt.Errorf("%w: output %d and %d share the same address", ErrOutputAddrNotUnique, j, index)
		}
		set[k] = index
		return nil
	}
}

// OutputsDepositAmountValidator returns a validator which checks that:
//	1. every output deposits more than zero
//	2. every output deposits less than the total supply
//	3. the sum of deposits does not exceed the total supply
// If -1 is passed to the validator func, then the sum is not aggregated over multiple calls.
func OutputsDepositAmountValidator() OutputsValidatorFunc {
	var sum uint64
	return func(index int, dep *SigLockedSingleDeposit) error {
		if dep.Amount == 0 {
			return fmt.Errorf("%w: output %d", ErrDepositAmountMustBeGreaterThanZero, index)
		}
		if dep.Amount > TokenSupply {
			return fmt.Errorf("%w: output %d", ErrOutputDepositsMoreThanTotalSupply, index)
		}
		if sum+dep.Amount > TokenSupply {
			return fmt.Errorf("%w: output %d", ErrOutputsSumExceedsTotalSupply, index)
		}
		if index != -1 {
			sum += dep.Amount
		}
		return nil
	}
}

// supposed to be called with -1 as input in order to be used over multiple calls.
var outputAmountValidator = OutputsDepositAmountValidator()

// ValidateOutputs validates the outputs by running them against the given OutputsValidatorFunc.
func ValidateOutputs(outputs Serializables, funcs ...OutputsValidatorFunc) error {
	for i, output := range outputs {
		dep, ok := output.(*SigLockedSingleDeposit)
		if !ok {
			return fmt.Errorf("%w: can only validate on signature locked single deposits", ErrUnknownOutputType)
		}
		for _, f := range funcs {
			if err := f(i, dep); err != nil {
				return err
			}
		}
	}
	return nil
}
