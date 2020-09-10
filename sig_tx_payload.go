package iota

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	// Defines the signed transaction's payload ID.
	SignedTransactionPayloadID uint32 = 0

	// Defines the maximum amount of inputs within a transaction.
	MaxInputsCount = 126
	// Defines the minimum amount of inputs within a transaction.
	MinInputsCount = 1
	// Defines the maximum amount of outputs within a transaction.
	MaxOutputsCount = 126
	// Defines the minimum amount of inputs within a transaction.
	MinOutputsCount = 1

	// Defines the minimum size of a signed transaction payload.
	SignedTransactionPayloadMinSize = UInt32ByteSize
)

var (
	// Returned if the count of inputs is too small.
	ErrMinInputsNotReached = fmt.Errorf("min %d input(s) are required within a transaction", MinInputsCount)
	// Returned if the count of inputs is too big.
	ErrMaxInputsExceeded = fmt.Errorf("max %d input(s) are allowed within a transaction", MaxInputsCount)
	// Returned if the count of outputs is too small.
	ErrMinOutputsNotReached = fmt.Errorf("min %d output(s) are required within a transaction", MinOutputsCount)
	// Returned if the count of outputs is too big.
	ErrMaxOutputsExceeded = fmt.Errorf("max %d output(s) are allowed within a transaction", MaxOutputsCount)
	// Returned if the count of unlock blocks doesn't match the count of inputs.
	ErrUnlockBlocksMustMatchInputCount = errors.New("the count of unlock blocks must match the inputs of the transaction")

	// restrictions around input within a transaction.
	inputsArrayBound = ArrayRules{
		Min:                         MinInputsCount,
		Max:                         MaxInputsCount,
		MinErr:                      ErrMinInputsNotReached,
		MaxErr:                      ErrMaxInputsExceeded,
		ElementBytesLexicalOrder:    true,
		ElementBytesLexicalOrderErr: ErrInputsOrderViolatesLexicalOrder,
	}

	// restrictions around outputs within a transaction.
	outputsArrayBound = ArrayRules{
		Min:                         MinInputsCount,
		Max:                         MaxInputsCount,
		MinErr:                      ErrMinOutputsNotReached,
		MaxErr:                      ErrMaxOutputsExceeded,
		ElementBytesLexicalOrder:    true,
		ElementBytesLexicalOrderErr: ErrOutputsOrderViolatesLexicalOrder,
	}
)

// SignedTransactionPayload is a transaction with its inputs, outputs and unlock blocks.
type SignedTransactionPayload struct {
	// The transaction respectively transfer part of a signed transaction payload.
	Transaction Serializable `json:"transaction"`
	// The unlock blocks defining the unlocking data for the inputs within the transaction.
	UnlockBlocks Serializables `json:"unlock_blocks"`
}

func (s *SignedTransactionPayload) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkMinByteLength(SignedTransactionPayloadMinSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid signed transaction payload bytes: %w", err)
		}
		if err := checkType(data, SignedTransactionPayloadID); err != nil {
			return 0, fmt.Errorf("unable to deserialize signed transaction payload: %w", err)
		}
	}

	// skip payload type
	bytesReadTotal := TypeDenotationByteSize
	data = data[TypeDenotationByteSize:]

	tx, txBytesRead, err := DeserializeObject(data, deSeriMode, TypeDenotationByte, TransactionSelector)
	if err != nil {
		return 0, fmt.Errorf("%w: unable to deserialize transaction within signed transaction payload", err)
	}
	bytesReadTotal += txBytesRead
	s.Transaction = tx

	// TODO: tx must be an unsigned tx but might be something else in the future
	inputCount := uint16(len(tx.(*UnsignedTransaction).Inputs))

	// advance to unlock blocks
	data = data[txBytesRead:]
	unlockBlocks, unlockBlocksByteRead, err := DeserializeArrayOfObjects(data, deSeriMode, TypeDenotationByte, UnlockBlockSelector, &ArrayRules{
		Min:    inputCount,
		Max:    inputCount,
		MinErr: ErrUnlockBlocksMustMatchInputCount,
		MaxErr: ErrUnlockBlocksMustMatchInputCount,
	})
	if err != nil {
		return 0, fmt.Errorf("%w: unable to deserialize unlock blocks", err)
	}
	bytesReadTotal += unlockBlocksByteRead

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := ValidateUnlockBlocks(unlockBlocks, UnlockBlocksSigUniqueAndRefValidator()); err != nil {
			return 0, err
		}
	}

	s.UnlockBlocks = unlockBlocks

	return bytesReadTotal, nil
}

func (s *SignedTransactionPayload) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := ValidateUnlockBlocks(s.UnlockBlocks, UnlockBlocksSigUniqueAndRefValidator()); err != nil {
			return nil, err
		}
	}

	var b bytes.Buffer
	if err := binary.Write(&b, binary.LittleEndian, SignedTransactionPayloadID); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize signed transaction payload ID", err)
	}

	// write transaction
	txBytes, err := s.Transaction.Serialize(deSeriMode)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to serialize signed transaction payload's inner transaction", err)
	}
	if _, err := b.Write(txBytes); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize signed transaction payload's inner transaction to buffer", err)
	}

	// write unlock blocks and count
	if err := binary.Write(&b, binary.LittleEndian, uint16(len(s.UnlockBlocks))); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize signed transaction payload's unlock block count", err)
	}
	for i := range s.UnlockBlocks {
		unlockBlockSer, err := s.UnlockBlocks[i].Serialize(deSeriMode)
		if err != nil {
			return nil, fmt.Errorf("%w: unable to serialize signed transaction payload's unlock block at pos %d", err, i)
		}
		if _, err := b.Write(unlockBlockSer); err != nil {
			return nil, fmt.Errorf("%w: unable to serialize signed transaction payload's unlock block at pos %d to buffer", err, i)
		}
	}

	return b.Bytes(), nil
}

func (s *SignedTransactionPayload) Validate() error {

	return nil
}
