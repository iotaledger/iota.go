package iotago

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
	"golang.org/x/crypto/blake2b"
)

// TransactionEssenceType defines the type of transaction.
type TransactionEssenceType = byte

const (
	// TransactionEssenceNormal denotes a standard transaction essence.
	TransactionEssenceNormal TransactionEssenceType = 1

	// MaxInputsCount defines the maximum amount of inputs within a TransactionEssence.
	MaxInputsCount = 128
	// MinInputsCount defines the minimum amount of inputs within a TransactionEssence.
	MinInputsCount = 1
	// MaxOutputsCount defines the maximum amount of outputs within a TransactionEssence.
	MaxOutputsCount = 128
	// MinOutputsCount defines the minimum amount of inputs within a TransactionEssence.
	MinOutputsCount = 1

	// InputsCommitmentLength defines the length of the inputs commitment hash.
	InputsCommitmentLength = blake2b.Size256
)

var (
	// ErrInvalidInputsCommitment gets returned when the inputs commitment is invalid.
	ErrInvalidInputsCommitment = errors.New("invalid inputs commitment")
	// ErrInputUTXORefsNotUnique gets returned if multiple inputs reference the same UTXO.
	ErrInputUTXORefsNotUnique = errors.New("inputs must each reference a unique UTXO")
	// ErrAliasOutputNonEmptyState gets returned if an AliasOutput with zeroed AliasID contains state (counters non-zero etc.).
	ErrAliasOutputNonEmptyState = errors.New("alias output is not empty state")
	// ErrAliasOutputCyclicAddress gets returned if an AliasOutput's AliasID results into the same address as the State/Governance controller.
	ErrAliasOutputCyclicAddress = errors.New("alias output's AliasID corresponds to state and/or governance controller")
	// ErrNFTOutputCyclicAddress gets returned if an NFTOutput's NFTID results into the same address as the address field within the output.
	ErrNFTOutputCyclicAddress = errors.New("nft output's NFTID corresponds to address field")
	// ErrFoundryOutputInvalidMaximumSupply gets returned when a FoundryOutput's MaximumSupply is invalid.
	ErrFoundryOutputInvalidMaximumSupply = errors.New("foundry output's maximum supply is invalid")
	// ErrFoundryOutputInvalidCirculatingSupply gets returned when a FoundryOutput's CirculatingSupply is invalid.
	ErrFoundryOutputInvalidCirculatingSupply = errors.New("foundry output's circulating supply is invalid")
	// ErrOutputsSumExceedsTotalSupply gets returned if the sum of the output deposits exceeds the total supply of tokens.
	ErrOutputsSumExceedsTotalSupply = errors.New("accumulated output balance exceeds total supply")
	// ErrOutputDepositsMoreThanTotalSupply gets returned if an output deposits more than the total supply.
	ErrOutputDepositsMoreThanTotalSupply = errors.New("an output can not deposit more than the total supply")
	// ErrStorageDepositLessThanMinReturnOutputStorageDeposit gets returned when the storage deposit condition's amount is less than the min storage deposit for the return output.
	ErrStorageDepositLessThanMinReturnOutputStorageDeposit = errors.New("storage deposit return amount is less than the min storage deposit needed for the return output")
	// ErrStorageDepositExceedsTargetOutputCost gets returned when the storage deposit condition's amount exceeds the needed amount for covering the target output's storage deposit.
	ErrStorageDepositExceedsTargetOutputCost = errors.New("storage deposit return amount exceeds needed amount for target output's storage deposit")
	// ErrMaxNativeTokensCountExceeded gets returned if outputs or transactions exceed the MaxNativeTokensCount.
	ErrMaxNativeTokensCountExceeded = errors.New("max native tokens count exceeded")

	essencePayloadGuard = serializer.SerializableGuard{
		ReadGuard: func(ty uint32) (serializer.Serializable, error) {
			if PayloadType(ty) != PayloadTaggedData {
				return nil, fmt.Errorf("transaction essence can only contain a tagged data payload: %w", ErrTypeIsNotSupportedPayload)
			}
			return PayloadSelector(ty)
		},
		WriteGuard: func(seri serializer.Serializable) error {
			if seri == nil {
				// can be nil
				return nil
			}
			if _, is := seri.(*TaggedData); !is {
				return ErrTypeIsNotSupportedPayload
			}
			return nil
		},
	}

	// restrictions around input within a transaction.
	essenceInputsArrayRules = &serializer.ArrayRules{
		Min: MinInputsCount,
		Max: MaxInputsCount,
		Guards: serializer.SerializableGuard{
			ReadGuard: func(ty uint32) (serializer.Serializable, error) {
				switch ty {
				case uint32(InputUTXO):
				default:
					return nil, fmt.Errorf("transaction essence can only contain UTXO input as inputs but got type ID %d: %w", ty, ErrUnsupportedObjectType)
				}
				return InputSelector(ty)
			},
			WriteGuard: func(seri serializer.Serializable) error {
				if seri == nil {
					return fmt.Errorf("%w: because nil", ErrTypeIsNotSupportedInput)
				}
				switch seri.(type) {
				case *UTXOInput:
				default:
					return ErrTypeIsNotSupportedInput
				}
				return nil
			},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates,
	}

	// restrictions around outputs within a transaction.
	essenceOutputsArrayRules = &serializer.ArrayRules{
		Min: MinOutputsCount,
		Max: MaxOutputsCount,
		Guards: serializer.SerializableGuard{
			ReadGuard: func(ty uint32) (serializer.Serializable, error) {
				switch ty {
				case uint32(OutputBasic):
				case uint32(OutputAlias):
				case uint32(OutputFoundry):
				case uint32(OutputNFT):
				default:
					return nil, fmt.Errorf("transaction essence can only contain basic/alias/foundry/nft outputs types but got type ID %d: %w", ty, ErrUnsupportedObjectType)
				}
				return OutputSelector(ty)
			},
			WriteGuard: func(seri serializer.Serializable) error {
				if seri == nil {
					return fmt.Errorf("%w: because nil", ErrTypeIsNotSupportedOutput)
				}
				switch seri.(type) {
				case *BasicOutput:
				case *AliasOutput:
				case *FoundryOutput:
				case *NFTOutput:
				default:
					return ErrTypeIsNotSupportedOutput
				}
				return nil
			},
		},
		ValidationMode: serializer.ArrayValidationModeNone,
	}
)

// TransactionEssenceInputsArrayRules returns array rules defining the constraints on Inputs within a TransactionEssence.
func TransactionEssenceInputsArrayRules() serializer.ArrayRules {
	return *essenceInputsArrayRules
}

// TransactionEssenceOutputsArrayRules returns array rules defining the constraints on Outputs within a TransactionEssence.
func TransactionEssenceOutputsArrayRules() serializer.ArrayRules {
	return *essenceOutputsArrayRules
}

// TransactionEssenceSelector implements SerializableSelectorFunc for transaction essence types.
func TransactionEssenceSelector(txType uint32) (*TransactionEssence, error) {
	var seri *TransactionEssence
	switch byte(txType) {
	case TransactionEssenceNormal:
		seri = &TransactionEssence{}
	default:
		return nil, fmt.Errorf("%w: type byte %d", ErrUnknownTransactionEssenceType, txType)
	}
	return seri, nil
}

// InputsCommitment is a commitment to the inputs of a transaction.
type InputsCommitment = [InputsCommitmentLength]byte

// TransactionEssence is the essence part of a Transaction.
type TransactionEssence struct {
	// The network ID for which this essence is valid for.
	NetworkID NetworkID
	// The inputs of this transaction.
	Inputs Inputs `json:"inputs"`
	// The commitment to the referenced inputs.
	InputsCommitment InputsCommitment `json:"inputsCommitment"`
	// The outputs of this transaction.
	Outputs Outputs `json:"outputs"`
	// The optional embedded payload.
	Payload Payload `json:"payload"`
}

// SigningMessage returns the to be signed message.
func (u *TransactionEssence) SigningMessage() ([]byte, error) {
	essenceBytes, err := u.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return nil, err
	}
	essenceBytesHash := blake2b.Sum256(essenceBytes)
	return essenceBytesHash[:], nil
}

// Sign produces signatures signing the essence for every given AddressKeys.
// The produced signatures are in the same order as the AddressKeys.
func (u *TransactionEssence) Sign(inputsCommitment []byte, addrKeys ...AddressKeys) ([]Signature, error) {
	if inputsCommitment == nil || len(inputsCommitment) != InputsCommitmentLength {
		return nil, ErrInvalidInputsCommitment
	}

	copy(u.InputsCommitment[:], inputsCommitment)

	signMsg, err := u.SigningMessage()
	if err != nil {
		return nil, err
	}

	sigs := make([]Signature, len(addrKeys))
	signer := NewInMemoryAddressSigner(addrKeys...)
	for i, v := range addrKeys {
		sig, err := signer.Sign(v.Address, signMsg)
		if err != nil {
			return nil, err
		}
		sigs[i] = sig
	}

	return sigs, nil
}

func (u *TransactionEssence) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(TransactionEssenceNormal), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize transaction essence: %w", err)
		}).
		ReadNum(&u.NetworkID, func(err error) error {
			return fmt.Errorf("unable to deserialize network ID of transaction essence: %w", err)
		}).
		ReadSliceOfObjects(&u.Inputs, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationByte, essenceInputsArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize inputs of transaction essence: %w", err)
		}).
		ReadArrayOf32Bytes(&u.InputsCommitment, func(err error) error {
			return fmt.Errorf("unable to deserialize inputs commitment of transaction essence: %w", err)
		}).
		ReadSliceOfObjects(&u.Outputs, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationByte, essenceOutputsArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize outputs of transaction essence: %w", err)
		}).
		ReadPayload(&u.Payload, deSeriMode, deSeriCtx, essencePayloadGuard.ReadGuard, func(err error) error {
			return fmt.Errorf("unable to deserialize outputs of transaction essence: %w", err)
		}).
		Done()
}

func (u *TransactionEssence) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (data []byte, err error) {
	return serializer.NewSerializer().
		WriteNum(byte(TransactionEssenceNormal), func(err error) error {
			return fmt.Errorf("unable to serialize transaction essence type ID: %w", err)
		}).
		WriteNum(u.NetworkID, func(err error) error {
			return fmt.Errorf("unable to serialize transaction essence network ID: %w", err)
		}).
		WriteSliceOfObjects(&u.Inputs, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsUint16, essenceInputsArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize transaction essence inputs: %w", err)
		}).
		WriteBytes(u.InputsCommitment[:], func(err error) error {
			return fmt.Errorf("unable to serialize transaction essence inputs commitment: %w", err)
		}).
		WriteSliceOfObjects(&u.Outputs, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsUint16, essenceOutputsArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize transaction essence outputs: %w", err)
		}).
		WritePayload(u.Payload, deSeriMode, deSeriCtx, essencePayloadGuard.WriteGuard, func(err error) error {
			return fmt.Errorf("unable to serialize transaction essence's embedded output: %w", err)
		}).
		Serialize()
}

func (u *TransactionEssence) Size() int {
	payloadSize := util.NumByteLen(uint32(0))
	if u.Payload != nil {
		payloadSize = u.Payload.Size()
	}
	return util.NumByteLen(byte(TransactionEssenceNormal)) +
		util.NumByteLen(u.NetworkID) +
		u.Inputs.Size() +
		InputsCommitmentLength +
		u.Outputs.Size() +
		payloadSize
}

func (u *TransactionEssence) MarshalJSON() ([]byte, error) {
	jTransactionEssence := &jsonTransactionEssence{
		NetworkID:        EncodeUint64(u.NetworkID),
		Inputs:           make([]*json.RawMessage, len(u.Inputs)),
		InputsCommitment: EncodeHex(u.InputsCommitment[:]),
		Outputs:          make([]*json.RawMessage, len(u.Outputs)),
		Payload:          nil,
	}
	jTransactionEssence.Type = int(TransactionEssenceNormal)

	for i, input := range u.Inputs {
		inputJson, err := input.MarshalJSON()
		if err != nil {
			return nil, err
		}
		rawMsgInputJson := json.RawMessage(inputJson)
		jTransactionEssence.Inputs[i] = &rawMsgInputJson
	}

	for i, output := range u.Outputs {
		outputJson, err := output.MarshalJSON()
		if err != nil {
			return nil, err
		}
		rawMsgOutputJson := json.RawMessage(outputJson)
		jTransactionEssence.Outputs[i] = &rawMsgOutputJson
	}

	if u.Payload != nil {
		jsonPayload, err := u.Payload.MarshalJSON()
		if err != nil {
			return nil, err
		}
		rawMsgJsonPayload := json.RawMessage(jsonPayload)
		jTransactionEssence.Payload = &rawMsgJsonPayload
	}
	return json.Marshal(jTransactionEssence)
}

func (u *TransactionEssence) UnmarshalJSON(bytes []byte) error {
	jTransactionEssence := &jsonTransactionEssence{}
	if err := json.Unmarshal(bytes, jTransactionEssence); err != nil {
		return err
	}
	seri, err := jTransactionEssence.ToSerializable()
	if err != nil {
		return err
	}
	*u = *seri.(*TransactionEssence)
	return nil
}

// syntacticallyValidate checks whether the transaction essence is syntactically valid.
// The function does not syntactically validate the input or outputs themselves.
func (u *TransactionEssence) syntacticallyValidate(rentStruct *RentStructure) error {
	if err := ValidateInputs(u.Inputs,
		InputsSyntacticalUnique(),
		InputsSyntacticalIndicesWithinBounds(),
	); err != nil {
		return err
	}

	if err := ValidateOutputs(u.Outputs,
		OutputsSyntacticalExpirationAndTimelock(),
		OutputsSyntacticalDepositAmount(rentStruct),
		OutputsSyntacticalNativeTokensCount(),
		OutputsSyntacticalFoundry(),
	); err != nil {
		return err
	}

	return nil
}

// jsonTransactionEssenceSelector selects the json transaction essence object for the given type.
func jsonTransactionEssenceSelector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch byte(ty) {
	case TransactionEssenceNormal:
		obj = &jsonTransactionEssence{}
	default:
		return nil, fmt.Errorf("unable to decode transaction essence type from JSON: %w", ErrUnknownTransactionEssenceType)
	}

	return obj, nil
}

// jsonTransactionEssence defines the json representation of a TransactionEssence.
type jsonTransactionEssence struct {
	Type             int                `json:"type"`
	NetworkID        string             `json:"networkId"`
	Inputs           []*json.RawMessage `json:"inputs"`
	InputsCommitment string             `json:"inputsCommitment"`
	Outputs          []*json.RawMessage `json:"outputs"`
	Payload          *json.RawMessage   `json:"payload"`
}

func (j *jsonTransactionEssence) ToSerializable() (serializer.Serializable, error) {
	unsigTx := &TransactionEssence{
		Inputs:  make(Inputs, len(j.Inputs)),
		Outputs: make(Outputs, len(j.Outputs)),
		Payload: nil,
	}

	var err error
	unsigTx.NetworkID, err = DecodeUint64(j.NetworkID)
	if err != nil {
		return nil, err
	}

	for i, jInput := range j.Inputs {
		jsonInput, err := DeserializeObjectFromJSON(jInput, jsonInputSelector)
		if err != nil {
			return nil, fmt.Errorf("unable to decode input type from JSON, pos %d: %w", i, err)
		}
		input, err := jsonInput.ToSerializable()
		if err != nil {
			return nil, fmt.Errorf("pos %d: %w", i, err)
		}
		unsigTx.Inputs[i] = input.(Input)
	}

	inputsCommitmentSlice, err := DecodeHex(j.InputsCommitment)
	if err != nil {
		return unsigTx, fmt.Errorf("unable to decode JSON inputs commitment: %w", err)
	}
	copy(unsigTx.InputsCommitment[:], inputsCommitmentSlice)

	for i, jOutput := range j.Outputs {
		jsonOutput, err := DeserializeObjectFromJSON(jOutput, JsonOutputSelector)
		if err != nil {
			return nil, fmt.Errorf("unable to decode output type from JSON, pos %d: %w", i, err)
		}
		output, err := jsonOutput.ToSerializable()
		if err != nil {
			return nil, fmt.Errorf("pos %d: %w", i, err)
		}
		unsigTx.Outputs[i] = output.(Output)
	}

	if j.Payload == nil {
		return unsigTx, nil
	}

	unsigTx.Payload, err = payloadFromJSONRawMsg(j.Payload)
	if err != nil {
		return nil, fmt.Errorf("unable to decode inner transaction essence payload: %w", err)
	}

	if _, isTaggedDataPayload := unsigTx.Payload.(*TaggedData); !isTaggedDataPayload {
		return nil, fmt.Errorf("%w: transaction essences only allow embedded tagged data payloads but got type %T instead", ErrInvalidJSON, unsigTx.Payload)
	}

	return unsigTx, nil
}
