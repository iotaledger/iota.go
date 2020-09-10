package iota

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	SignedTransactionPayloadID uint32 = 0

	MaxInputsCount  = 126
	MinInputsCount  = 1
	MaxOutputsCount = 126
	MinOutputsCount = 1

	SignedTransactionPayloadMinSize = UInt32ByteSize
)

var (
	ErrMinInputsNotReached             = errors.New(fmt.Sprintf("min %d input(s) are required within a transaction", MinInputsCount))
	ErrMaxInputsExceeded               = errors.New(fmt.Sprintf("max %d input(s) are allowed within a transaction", MaxInputsCount))
	ErrMinOutputsNotReached            = errors.New(fmt.Sprintf("min %d output(s) are required within a transaction", MinOutputsCount))
	ErrMaxOutputsExceeded              = errors.New(fmt.Sprintf("max %d output(s) are allowed within a transaction", MaxOutputsCount))
	ErrUnlockBlocksMustMatchInputCount = errors.New("the count of unlock blocks must match the inputs of the transaction")

	inputsArrayBound = ArrayRules{
		Min:                         MinInputsCount,
		Max:                         MaxInputsCount,
		MinErr:                      ErrMinInputsNotReached,
		MaxErr:                      ErrMaxInputsExceeded,
		ElementBytesLexicalOrder:    true,
		ElementBytesLexicalOrderErr: ErrInputsOrderViolatesLexicalOrder,
	}

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
	Transaction  Serializable  `json:"transaction"`
	UnlockBlocks Serializables `json:"unlock_blocks"`
}

func (s *SignedTransactionPayload) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkMinByteLength(SignedTransactionPayloadMinSize, len(data)); err != nil {
			return 0, err
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
		return 0, err
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
		return 0, err
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
		return nil, err
	}

	// write transaction
	txBytes, err := s.Transaction.Serialize(deSeriMode)
	if err != nil {
		return nil, err
	}
	if _, err := b.Write(txBytes); err != nil {
		return nil, err
	}

	// write unlock blocks and count
	if err := binary.Write(&b, binary.LittleEndian, uint16(len(s.UnlockBlocks))); err != nil {
		return nil, err
	}
	for i := range s.UnlockBlocks {
		unlockBlockSer, err := s.UnlockBlocks[i].Serialize(deSeriMode)
		if err != nil {
			return nil, err
		}
		if _, err := b.Write(unlockBlockSer); err != nil {
			return nil, err
		}
	}

	return b.Bytes(), nil
}

func (s *SignedTransactionPayload) Validate() error {

	return nil
}
