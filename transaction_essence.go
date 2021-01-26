package iota

import (
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
	MaxInputsCount = 127
	// Defines the minimum amount of inputs within a TransactionEssence.
	MinInputsCount = 1
	// Defines the maximum amount of outputs within a TransactionEssence.
	MaxOutputsCount = 127
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
	// Returned if a SigLockedDustAllowanceOutput deposits less than OutputSigLockedDustAllowanceOutputMinDeposit.
	ErrOutputDustAllowanceLessThanMinDeposit = errors.New("dust allowance output deposits less than the minimum required amount")

	// restrictions around input within a transaction.
	inputsArrayBound = ArrayRules{
		Min:            MinInputsCount,
		Max:            MaxInputsCount,
		ValidationMode: ArrayValidationModeLexicalOrdering,
	}

	// restrictions around outputs within a transaction.
	outputsArrayBound = ArrayRules{
		Min:            MinInputsCount,
		Max:            MaxInputsCount,
		ValidationMode: ArrayValidationModeLexicalOrdering,
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
	return NewDeserializer(data).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(DeSeriModePerformValidation) {
				if err := checkMinByteLength(TransactionEssenceMinByteSize, len(data)); err != nil {
					return fmt.Errorf("invalid transaction essence bytes: %w", err)
				}
				if err := checkTypeByte(data, TransactionEssenceNormal); err != nil {
					return fmt.Errorf("unable to deserialize transaction essence: %w", err)
				}
			}
			return nil
		}).
		Skip(SmallTypeDenotationByteSize, func(err error) error {
			return fmt.Errorf("unable to skip transaction essence ID during deserialization: %w", err)
		}).
		ReadSliceOfObjects(func(seri Serializables) { u.Inputs = seri }, deSeriMode, TypeDenotationByte, InputSelector, &inputsArrayBound, func(err error) error {
			return fmt.Errorf("unable to deserialize inputs of transaction essence: %w", err)
		}).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(DeSeriModePerformValidation) {
				if err := ValidateInputs(u.Inputs, InputsUTXORefsUniqueValidator()); err != nil {
					return fmt.Errorf("%w: unable to deserialize inputs of transaction essence since they are invalid", err)
				}
			}
			return nil
		}).
		ReadSliceOfObjects(func(seri Serializables) { u.Outputs = seri }, deSeriMode, TypeDenotationByte, OutputSelector, &inputsArrayBound, func(err error) error {
			return fmt.Errorf("unable to deserialize outputs of transaction essence: %w", err)
		}).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(DeSeriModePerformValidation) {
				if err := ValidateOutputs(u.Outputs, OutputsAddrUniqueValidator()); err != nil {
					return fmt.Errorf("%w: unable to deserialize outputs of transaction essence since they are invalid", err)
				}
			}
			return nil
		}).
		ReadPayload(func(seri Serializable) { u.Payload = seri }, deSeriMode, func(err error) error {
			return fmt.Errorf("unable to deserialize outputs of transaction essence: %w", err)
		}).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(DeSeriModePerformValidation) {
				if u.Payload != nil {
					// supports only indexation payloads
					if _, isIndexationPayload := u.Payload.(*Indexation); !isIndexationPayload {
						return fmt.Errorf("%w: transaction essences only allow embedded indexation payloads but got %T instead", ErrInvalidBytes, u.Payload)
					}
				}
			}
			return nil
		}).
		Done()
}

func (u *TransactionEssence) Serialize(deSeriMode DeSerializationMode) (data []byte, err error) {
	var inputsWrittenConsumer, outputsWrittenConsumer WrittenObjectConsumer
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if inputsArrayBound.ValidationMode.HasMode(ArrayValidationModeLexicalOrdering) {
			inputsLexicalOrderValidator := inputsArrayBound.LexicalOrderValidator()
			inputsWrittenConsumer = func(index int, written []byte) error {
				if err := inputsLexicalOrderValidator(index, written); err != nil {
					return fmt.Errorf("%w: unable to serialize inputs of transaction essence since inputs are not in lexical order", err)
				}
				return nil
			}
		}
		if outputsArrayBound.ValidationMode.HasMode(ArrayValidationModeLexicalOrdering) {
			outputsLexicalOrderValidator := outputsArrayBound.LexicalOrderValidator()
			outputsWrittenConsumer = func(index int, written []byte) error {
				if err := outputsLexicalOrderValidator(index, written); err != nil {
					return fmt.Errorf("%w: unable to serialize outputs of transaction essence since outputs are not in lexical order", err)
				}
				return nil
			}
		}
	}

	return NewSerializer().
		Do(func() {
			if deSeriMode.HasMode(DeSeriModePerformLexicalOrdering) {
				u.SortInputsOutputs()
			}
		}).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(DeSeriModePerformValidation) {
				if err := u.SyntacticallyValidate(); err != nil {
					return err
				}
			}
			return nil
		}).
		WriteNum(TransactionEssenceNormal, func(err error) error {
			return fmt.Errorf("unable to serialize transaction essence type ID: %w", err)
		}).
		WriteSliceOfObjects(u.Inputs, deSeriMode, inputsWrittenConsumer, func(err error) error {
			return fmt.Errorf("unable to serialize transaction essence inputs: %w", err)
		}).
		WriteSliceOfObjects(u.Outputs, deSeriMode, outputsWrittenConsumer, func(err error) error {
			return fmt.Errorf("unable to serialize transaction essence outputs: %w", err)
		}).
		WritePayload(u.Payload, deSeriMode, func(err error) error {
			return fmt.Errorf("unable to serialize transaction essence's embedded output: %w", err)
		}).
		Serialize()
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
//	2. every output (per type) deposits to a unique address and deposits more than zero
//	3. the accumulated deposit output is not over the total supply
//	4. SigLockedDustAllowanceOutput deposits at least OutputSigLockedDustAllowanceOutputMinDeposit.
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
