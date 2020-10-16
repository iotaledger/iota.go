package iota

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

// Defines the type of transaction.
type TransactionEssenceType = byte

const (
	// Denotes a standard transaction essence.
	// TODO: find a better name for this
	TransactionEssenceNormal TransactionEssenceType = iota

	// Defines the minimum size of a TransactionEssence.
	TransactionEssenceMinByteSize = TypeDenotationByteSize + StructArrayLengthByteSize + StructArrayLengthByteSize + PayloadLengthByteSize

	// Defines the maximum amount of inputs within a TransactionEssence.
	MaxInputsCount = 126
	// Defines the minimum amount of inputs within a TransactionEssence.
	MinInputsCount = 1
	// Defines the maximum amount of outputs within a TransactionEssence.
	MaxOutputsCount = 126
	// Defines the minimum amount of inputs within a TransactionEssence.
	MinOutputsCount = 1
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

// TransactionEssenceSelector implements SerializableSelectorFunc for transaction essence types.
func TransactionEssenceSelector(txType uint32) (Serializable, error) {
	var seri Serializable
	switch byte(txType) {
	case TransactionEssenceNormal:
		seri = &TransactionEssence{}
	default:
		return nil, fmt.Errorf("%w: type byte %d", ErrUnknownTransactionEssenceType, txType)
	}
	return seri, nil
}

// TransactionEssence is the essence part of a Transaction.
type TransactionEssence struct {
	// The inputs of this transaction.
	Inputs Serializables `json:"inputs"`
	// The outputs of this transaction.
	Outputs Serializables `json:"outputs"`
	// The optional embedded payload.
	Payload Serializable `json:"payload"`
}

// SortInputsOuputs sorts the inputs and outputs according to their serialized lexical representation.
// Usually an implicit call to SortInputsOutputs() should be done by instructing serialization to use DeSeriModePerformLexicalOrdering.
func (u *TransactionEssence) SortInputsOutputs() {
	sort.Sort(SortedSerializables(u.Inputs))
	sort.Sort(SortedSerializables(u.Outputs))
}

func (u *TransactionEssence) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkMinByteLength(TransactionEssenceMinByteSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid transaction essence bytes: %w", err)
		}
		if err := checkTypeByte(data, TransactionEssenceNormal); err != nil {
			return 0, fmt.Errorf("unable to deserialize transaction essence: %w", err)
		}
	}

	// skip type byte
	bytesReadTotal := SmallTypeDenotationByteSize
	data = data[SmallTypeDenotationByteSize:]

	inputs, inputBytesRead, err := DeserializeArrayOfObjects(data, deSeriMode, TypeDenotationByte, InputSelector, &inputsArrayBound)
	if err != nil {
		return 0, fmt.Errorf("%w: unable to deserialize inputs of transaction essence", err)
	}
	bytesReadTotal += inputBytesRead

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := ValidateInputs(inputs, InputsUTXORefsUniqueValidator()); err != nil {
			return 0, fmt.Errorf("%w: unable to deserialize inputs of transaction essence since they are invalid", err)
		}
	}
	u.Inputs = inputs

	// advance to outputs
	data = data[inputBytesRead:]
	outputs, outputBytesRead, err := DeserializeArrayOfObjects(data, deSeriMode, TypeDenotationByte, OutputSelector, &outputsArrayBound)
	if err != nil {
		return 0, fmt.Errorf("%w: unable to deserialize outputs of transaction essence", err)
	}
	bytesReadTotal += outputBytesRead

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := ValidateOutputs(outputs, OutputsAddrUniqueValidator()); err != nil {
			return 0, fmt.Errorf("%w: unable to deserialize outputs of transaction essence since they are invalid", err)
		}
	}
	u.Outputs = outputs

	// advance to payload
	data = data[outputBytesRead:]

	payload, payloadBytesRead, err := ParsePayload(data, deSeriMode)
	if err != nil {
		return 0, fmt.Errorf("%w: can't parse payload within transaction essence", err)
	}
	bytesReadTotal += payloadBytesRead

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if payload != nil {
			// supports only indexation payloads
			if _, isIndexationPayload := payload.(*Indexation); !isIndexationPayload {
				return 0, fmt.Errorf("%w: transaction essences only allow embedded indexation payloads but got %T instead", ErrInvalidBytes, payload)
			}
		}
	}

	u.Payload = payload

	return bytesReadTotal, nil
}

func (u *TransactionEssence) Serialize(deSeriMode DeSerializationMode) (data []byte, err error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := u.SyntacticallyValidate(); err != nil {
			return nil, err
		}
	}

	if deSeriMode.HasMode(DeSeriModePerformLexicalOrdering) {
		u.SortInputsOutputs()
	}

	var buf bytes.Buffer
	if err := buf.WriteByte(TransactionEssenceNormal); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize transaction essence type ID", err)
	}

	var inputsLexicalOrderValidator LexicalOrderFunc
	if deSeriMode.HasMode(DeSeriModePerformValidation) && inputsArrayBound.ElementBytesLexicalOrder {
		inputsLexicalOrderValidator = inputsArrayBound.LexicalOrderValidator()
	}

	// write inputs
	if err := binary.Write(&buf, binary.LittleEndian, uint16(len(u.Inputs))); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize transaction essence's input count", err)
	}
	for i := range u.Inputs {
		inputSer, err := u.Inputs[i].Serialize(deSeriMode)
		if err != nil {
			return nil, fmt.Errorf("%w: unable to serialize input of transaction essence at index %d", err, i)
		}
		if _, err := buf.Write(inputSer); err != nil {
			return nil, fmt.Errorf("%w: unable to serialize input of transaction essence at index %d to buffer", err, i)
		}
		if inputsLexicalOrderValidator != nil {
			if err := inputsLexicalOrderValidator(i, inputSer); err != nil {
				return nil, fmt.Errorf("%w: unable to serialize inputs of transaction essence since inputs are not in lexical order", err)
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
			return nil, fmt.Errorf("%w: unable to serialize output of transaction essence at index %d", err, i)
		}
		if _, err := buf.Write(outputSer); err != nil {
			return nil, fmt.Errorf("%w: unable to serialize output of transaction essence at index %d to buffer", err, i)
		}
		if outputsLexicalOrderValidator != nil {
			if err := outputsLexicalOrderValidator(i, outputSer); err != nil {
				return nil, fmt.Errorf("%w: unable to serialize outputs of transaction essence since outputs are not in lexical order", err)
			}
		}
	}

	// no payload
	if u.Payload == nil {
		if err := binary.Write(&buf, binary.LittleEndian, uint32(0)); err != nil {
			return nil, fmt.Errorf("%w: unable to serialize transaction essence's inner zero payload length", err)
		}
		return buf.Bytes(), nil
	}

	// write payload
	payloadSer, err := u.Payload.Serialize(deSeriMode)
	if err := binary.Write(&buf, binary.LittleEndian, uint32(len(payloadSer))); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize transaction essence's payload length", err)
	}

	if _, err := buf.Write(payloadSer); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize transaction essence's payload to buffer", err)
	}

	return buf.Bytes(), nil
}

func (u *TransactionEssence) MarshalJSON() ([]byte, error) {
	jsonTx := &jsontransactionessence{
		Inputs:  make([]*json.RawMessage, len(u.Inputs)),
		Outputs: make([]*json.RawMessage, len(u.Outputs)),
		Payload: nil,
	}
	jsonTx.Type = int(TransactionEssenceNormal)

	for i, input := range u.Inputs {
		inputJson, err := input.MarshalJSON()
		if err != nil {
			return nil, err
		}
		rawMsgInputJson := json.RawMessage(inputJson)
		jsonTx.Inputs[i] = &rawMsgInputJson

	}
	for i, output := range u.Outputs {
		outputJson, err := output.MarshalJSON()
		if err != nil {
			return nil, err
		}
		rawMsgOutputJson := json.RawMessage(outputJson)
		jsonTx.Outputs[i] = &rawMsgOutputJson
	}

	if u.Payload != nil {
		jsonPayload, err := u.Payload.MarshalJSON()
		if err != nil {
			return nil, err
		}
		rawMsgJsonPayload := json.RawMessage(jsonPayload)
		jsonTx.Payload = &rawMsgJsonPayload
	}
	return json.Marshal(jsonTx)
}

func (u *TransactionEssence) UnmarshalJSON(bytes []byte) error {
	jsonTx := &jsontransactionessence{}
	if err := json.Unmarshal(bytes, jsonTx); err != nil {
		return err
	}
	seri, err := jsonTx.ToSerializable()
	if err != nil {
		return err
	}
	*u = *seri.(*TransactionEssence)
	return nil
}

// SyntacticallyValidate checks whether the transaction essence is syntactically valid by checking whether:
//	1. every input references a unique UTXO and has valid UTXO index bounds
//	2. every output deposits to a unique address and deposits more than zero
//	3. the accumulated deposit output is not over the total supply
// The function does not syntactically validate the input or outputs themselves.
func (u *TransactionEssence) SyntacticallyValidate() error {

	if len(u.Inputs) == 0 {
		return ErrMinInputsNotReached
	}

	if len(u.Outputs) == 0 {
		return ErrMinOutputsNotReached
	}

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

// jsontransactionessenceselector selects the json transaction essence object for the given type.
func jsontransactionessenceselector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch byte(ty) {
	case TransactionEssenceNormal:
		obj = &jsontransactionessence{}
	default:
		return nil, fmt.Errorf("unable to decode transaction essence type from JSON: %w", ErrUnknownTransactionEssenceType)
	}

	return obj, nil
}

// jsontransactionessence defines the json representation of a TransactionEssence.
type jsontransactionessence struct {
	Type    int                `json:"type"`
	Inputs  []*json.RawMessage `json:"inputs"`
	Outputs []*json.RawMessage `json:"outputs"`
	Payload *json.RawMessage   `json:"payload"`
}

func (j *jsontransactionessence) ToSerializable() (Serializable, error) {
	unsigTx := &TransactionEssence{
		Inputs:  make(Serializables, len(j.Inputs)),
		Outputs: make(Serializables, len(j.Outputs)),
		Payload: nil,
	}

	for i, input := range j.Inputs {
		jsonInput, err := DeserializeObjectFromJSON(input, jsoninputselector)
		if err != nil {
			return nil, fmt.Errorf("unable to decode input type from JSON, pos %d: %w", i, err)
		}
		input, err := jsonInput.ToSerializable()
		if err != nil {
			return nil, fmt.Errorf("pos %d: %w", i, err)
		}
		unsigTx.Inputs[i] = input
	}

	for i, output := range j.Outputs {
		jsonOutput, err := DeserializeObjectFromJSON(output, jsonoutputselector)
		if err != nil {
			return nil, fmt.Errorf("unable to decode output type from JSON, pos %d: %w", i, err)
		}
		output, err := jsonOutput.ToSerializable()
		if err != nil {
			return nil, fmt.Errorf("pos %d: %w", i, err)
		}
		unsigTx.Outputs[i] = output
	}

	if j.Payload == nil {
		return unsigTx, nil
	}

	jsonPayload, err := DeserializeObjectFromJSON(j.Payload, jsonpayloadselector)
	if err != nil {
		return nil, err
	}

	if _, isJSONIndexationPayload := jsonPayload.(*jsonindexation); !isJSONIndexationPayload {
		return nil, fmt.Errorf("%w: transaction essences only allow embedded indexation payloads but got type %T instead", ErrInvalidJSON, jsonPayload)
	}

	unsigTx.Payload, err = jsonPayload.ToSerializable()
	if err != nil {
		return nil, fmt.Errorf("unable to decode inner transaction essence payload: %w", err)
	}

	return unsigTx, nil
}
