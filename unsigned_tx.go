package iota

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// Defines the type of transaction.
type TransactionType = uint32

const (
	// Denotes an unsigned transaction.
	TransactionUnsigned TransactionType = iota

	// Defines the length of a transaction hash.
	TransactionIDLength = 32

	// Defines the minimum size of an unsigned transaction.
	UnsignedTransactionMinByteSize = TypeDenotationByteSize + StructArrayLengthByteSize + StructArrayLengthByteSize + PayloadLengthByteSize
)

var (
	// Returned if the inputs are not in lexical order when serialized.
	ErrInputsOrderViolatesLexicalOrder = errors.New("inputs must be in their lexical order (byte wise) when serialized")
	// Returned if the outputs are not in lexical order when serialized.
	ErrOutputsOrderViolatesLexicalOrder = errors.New("outputs must be in their lexical order (byte wise) when serialized")
	// Returned if multiple inputs reference the same UTXO.
	ErrInputUTXORefsNotUnique = errors.New("inputs must each reference a unique UTXO")
	// Returned if multiple outputs deposit to the same address.
	ErrOutputAddrNotUnique = errors.New("outputs must each deposit to a unique address")
	// Returned if the sum of the output deposits exceeds the total supply of tokens.
	ErrOutputsSumExceedsTotalSupply = errors.New("accumulated output balance exceeds total supply")
	// Returned if an output deposits more than the total supply.
	ErrOutputDepositsMoreThanTotalSupply = errors.New("an output can not deposit more than the total supply")
)

// TransactionSelector implements SerializableSelectorFunc for transaction types.
func TransactionSelector(txType uint32) (Serializable, error) {
	var seri Serializable
	switch txType {
	case TransactionUnsigned:
		seri = &UnsignedTransaction{}
	default:
		return nil, fmt.Errorf("%w: type byte %d", ErrUnknownTransactionType, txType)
	}
	return seri, nil
}

// UnsignedTransaction is the unsigned part of a transaction.
type UnsignedTransaction struct {
	// The inputs of this transaction.
	Inputs Serializables `json:"inputs"`
	// The outputs of this transaction.
	Outputs Serializables `json:"outputs"`
	// The optional embedded payload.
	Payload Serializable `json:"payload"`
}

func (u *UnsignedTransaction) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkMinByteLength(UnsignedTransactionMinByteSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid unsigned trnasaction bytes: %w", err)
		}
		if err := checkType(data, TransactionUnsigned); err != nil {
			return 0, fmt.Errorf("unable to deserialize unsigned transaction: %w", err)
		}
	}

	// skip type byte
	bytesReadTotal := TypeDenotationByteSize
	data = data[TypeDenotationByteSize:]

	inputs, inputBytesRead, err := DeserializeArrayOfObjects(data, deSeriMode, TypeDenotationByte, InputSelector, &inputsArrayBound)
	if err != nil {
		return 0, fmt.Errorf("%w: unable to deserialize inputs of unsigned transaction", err)
	}
	bytesReadTotal += inputBytesRead

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := ValidateInputs(inputs, InputsUTXORefsUniqueValidator()); err != nil {
			return 0, fmt.Errorf("%w: unable to deserialize inputs of unsigned transaction since they are invalid", err)
		}
	}
	u.Inputs = inputs

	// advance to outputs
	data = data[inputBytesRead:]
	outputs, outputBytesRead, err := DeserializeArrayOfObjects(data, deSeriMode, TypeDenotationByte, OutputSelector, &outputsArrayBound)
	if err != nil {
		return 0, fmt.Errorf("%w: unable to deserialize outputs of unsigned transaction", err)
	}
	bytesReadTotal += outputBytesRead

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := ValidateOutputs(outputs, OutputsAddrUniqueValidator()); err != nil {
			return 0, fmt.Errorf("%w: unable to deserialize outputs of unsigned transaction since they are invalid", err)
		}
	}
	u.Outputs = outputs

	// advance to payload
	data = data[outputBytesRead:]

	payload, payloadBytesRead, err := ParsePayload(data, deSeriMode)
	if err != nil {
		return 0, fmt.Errorf("%w: can't parse payload within unsigned transaction", err)
	}
	bytesReadTotal += payloadBytesRead

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if payload != nil {
			// supports only indexation payloads
			if _, isIndexationPayload := payload.(*IndexationPayload); !isIndexationPayload {
				return 0, fmt.Errorf("%w: unsigned transactions only allow embedded indexation payloads but got %T instead", ErrInvalidBytes, payload)
			}
		}
	}

	u.Payload = payload

	return bytesReadTotal, nil
}

func (u *UnsignedTransaction) Serialize(deSeriMode DeSerializationMode) (data []byte, err error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := ValidateInputs(u.Inputs, InputsUTXORefsUniqueValidator()); err != nil {
			return nil, fmt.Errorf("%w: unable to serialize unsigned transaction since inputs are invalid", err)
		}
		if err := ValidateOutputs(u.Outputs, OutputsAddrUniqueValidator()); err != nil {
			return nil, fmt.Errorf("%w: unable to serialize unsigned transaction since outputs are invalid", err)
		}
	}

	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, TransactionUnsigned); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize unsigned transaction type ID", err)
	}

	var inputsLexicalOrderValidator LexicalOrderFunc
	if deSeriMode.HasMode(DeSeriModePerformValidation) && inputsArrayBound.ElementBytesLexicalOrder {
		inputsLexicalOrderValidator = inputsArrayBound.LexicalOrderValidator()
	}

	// write inputs
	if err := binary.Write(&buf, binary.LittleEndian, uint16(len(u.Inputs))); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize unsigned transaction's input count", err)
	}
	for i := range u.Inputs {
		inputSer, err := u.Inputs[i].Serialize(deSeriMode)
		if err != nil {
			return nil, fmt.Errorf("%w: unable to serialize input of unsigned transaction at index %d", err, i)
		}
		if _, err := buf.Write(inputSer); err != nil {
			return nil, fmt.Errorf("%w: unable to serialize input of unsigned transaction at index %d to buffer", err, i)
		}
		if inputsLexicalOrderValidator != nil {
			if err := inputsLexicalOrderValidator(i, inputSer); err != nil {
				return nil, fmt.Errorf("%w: unable to serialize inputs of unsigned transaction since inputs are not in lexical order", err)
			}
		}
	}

	var outputsLexicalOrderValidator LexicalOrderFunc
	if deSeriMode.HasMode(DeSeriModePerformValidation) && outputsArrayBound.ElementBytesLexicalOrder {
		outputsLexicalOrderValidator = outputsArrayBound.LexicalOrderValidator()
	}

	// write outputs
	if err := binary.Write(&buf, binary.LittleEndian, uint16(len(u.Outputs))); err != nil {
		return nil, err
	}
	for i := range u.Outputs {
		outputSer, err := u.Outputs[i].Serialize(deSeriMode)
		if err != nil {
			return nil, fmt.Errorf("%w: unable to serialize output of unsigned transaction at index %d", err, i)
		}
		if _, err := buf.Write(outputSer); err != nil {
			return nil, fmt.Errorf("%w: unable to serialize output of unsigned transaction at index %d to buffer", err, i)
		}
		if outputsLexicalOrderValidator != nil {
			if err := outputsLexicalOrderValidator(i, outputSer); err != nil {
				return nil, fmt.Errorf("%w: unable to serialize outputs of unsigned transaction since outputs are not in lexical order", err)
			}
		}
	}

	// no payload
	if u.Payload == nil {
		if err := binary.Write(&buf, binary.LittleEndian, uint32(0)); err != nil {
			return nil, fmt.Errorf("%w: unable to serialize unsigned transaction's inner zero payload length", err)
		}
		return buf.Bytes(), nil
	}

	// write payload
	payloadSer, err := u.Payload.Serialize(deSeriMode)
	if _, err := buf.Write(payloadSer); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize unsigned transaction's payload", err)
	}

	if err := binary.Write(&buf, binary.LittleEndian, uint32(len(payloadSer))); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize unsigned transaction's payload length", err)
	}

	if _, err := buf.Write(payloadSer); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize unsigned transaction's payload to buffer", err)
	}

	return buf.Bytes(), nil
}

// SyntacticallyValid checks whether the unsigned transaction is syntactically valid by checking whether:
//	1. every input references a unique UTXO and has valid UTXO index bounds
//	2. every output deposits to a unique address and deposits more than zero
//	3. the accumulated deposit output is not over the total supply
// The function does not syntactically validate the input or outputs themselves.
func (u *UnsignedTransaction) SyntacticallyValid() error {
	if err := ValidateInputs(u.Inputs,
		InputsUTXORefIndexBoundsValidator(),
		InputsUTXORefsUniqueValidator(),
	); err != nil {
		return err
	}

	if err := ValidateOutputs(u.Outputs,
		OutputsAddrUniqueValidator(),
		OutputsDepositAmountValidator(),
	); err != nil {
		return err
	}

	return nil
}
