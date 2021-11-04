package iotago

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/iotaledger/hive.go/serializer"
)

// TokenTag is a tag holding some additional data which might be interpreted by higher layers.
type TokenTag = [TokenTagLength]byte

// FoundryOutput is an output type which controls the supply of user defined native tokens.
type FoundryOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount uint64
	// The native tokens held by the output.
	NativeTokens serializer.Serializables
	// The alias controlling the foundry.
	Address serializer.Serializable
	// The serial number of the foundry.
	SerialNumber uint32
	// The tag which is always the last 12 bytes of the tokens generated by this foundry.
	TokenTag TokenTag
	// The circulating supply of tokens controlled by this foundry.
	CirculatingSupply *big.Int
	// The maximum supply of tokens controlled by this foundry.
	MaximumSupply *big.Int
	// The token scheme this foundry uses.
	TokenScheme serializer.Serializable
	// The feature blocks which modulate the constraints on the output.
	Blocks serializer.Serializables
}

func (f *FoundryOutput) NativeTokenSet() serializer.Serializables {
	return f.NativeTokens
}

func (f *FoundryOutput) Deposit() (uint64, error) {
	return f.Amount, nil
}

func (f *FoundryOutput) Target() (serializer.Serializable, error) {
	return f.Address, nil
}

func (f *FoundryOutput) Type() OutputType {
	return OutputFoundry
}

func (f *FoundryOutput) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	return serializer.NewDeserializer(data).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := serializer.CheckTypeByte(data, OutputAlias); err != nil {
					return fmt.Errorf("unable to deserialize foundry output: %w", err)
				}
			}
			return nil
		}).
		Skip(serializer.SmallTypeDenotationByteSize, func(err error) error {
			return fmt.Errorf("unable to skip foundry output type during deserialization: %w", err)
		}).
		ReadNum(&f.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for foundry output: %w", err)
		}).
		ReadSliceOfObjects(func(seri serializer.Serializables) { f.NativeTokens = seri }, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationNone, func(ty uint32) (serializer.Serializable, error) {
			return &NativeToken{}, nil
		}, nativeTokensArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize native tokens for foundry output: %w", err)
		}).
		ReadNum(&f.SerialNumber, func(err error) error {
			return fmt.Errorf("unable to deserialize serial number for foundry output: %w", err)
		}).
		ReadArrayOf12Bytes(&f.TokenTag, func(err error) error {
			return fmt.Errorf("unable to deserialize token tag for foundry output: %w", err)
		}).
		ReadUint256(f.CirculatingSupply, func(err error) error {
			return fmt.Errorf("unable to deserialize circulating supply for foundry output: %w", err)
		}).
		ReadUint256(f.MaximumSupply, func(err error) error {
			return fmt.Errorf("unable to deserialize maximum supply for foundry output: %w", err)
		}).
		ReadObject(func(seri serializer.Serializable) { f.TokenScheme = seri }, deSeriMode, serializer.TypeDenotationByte, TokenSchemeSelector, func(err error) error {
			return fmt.Errorf("unable to deserialize token scheme for foundry output: %w", err)
		}).
		ReadSliceOfObjects(func(seri serializer.Serializables) { f.Blocks = seri }, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationByte, FeatureBlockSelector, featBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize feature blocks for foundry output: %w", err)
		}).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := outputAmountValidator(-1, f); err != nil {
					return fmt.Errorf("%w: unable to deserialize foundry output", err)
				}
			}
			return nil
		}).
		Done()
}

func (f *FoundryOutput) Serialize(deSeriMode serializer.DeSerializationMode) ([]byte, error) {
	var nativeTokensWrittenConsumer serializer.WrittenObjectConsumer
	return serializer.NewSerializer().
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := outputAmountValidator(-1, f); err != nil {
					return fmt.Errorf("%w: unable to serialize foundry output", err)
				}

				if err := isValidAddrType(f.Address); err != nil {
					return fmt.Errorf("invalid address set in foundry output: %w", err)
				}

				nativeTokensLexicalNoDupsValidator := nativeTokensArrayRules.LexicalOrderWithoutDupsValidator()
				nativeTokensWrittenConsumer = func(index int, written []byte) error {
					if err := nativeTokensLexicalNoDupsValidator(index, written); err != nil {
						return fmt.Errorf("%w: unable to serialize native tokens of alias output since inputs are not lexically sorted or contain duplicates", err)
					}
					return nil
				}
			}
			return nil
		}).
		WriteNum(OutputFoundry, func(err error) error {
			return fmt.Errorf("unable to serialize foundry output type ID: %w", err)
		}).
		WriteNum(f.Amount, func(err error) error {
			return fmt.Errorf("unable to serialize foundry output amount: %w", err)
		}).
		WriteSliceOfObjects(f.NativeTokens, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, nativeTokensWrittenConsumer, func(err error) error {
			return fmt.Errorf("unable to serialize foundry output native tokens: %w", err)
		}).
		WriteObject(f.Address, deSeriMode, func(err error) error {
			return fmt.Errorf("unable to serialize foundry output address: %w", err)
		}).
		WriteNum(f.SerialNumber, func(err error) error {
			return fmt.Errorf("unable to serialize foundry output serial number: %w", err)
		}).
		WriteBytes(f.TokenTag[:], func(err error) error {
			return fmt.Errorf("unable to serialize foundry output token tag: %w", err)
		}).
		WriteUint256(f.CirculatingSupply, func(err error) error {
			return fmt.Errorf("unable to serialize foundry output circulating supply: %w", err)
		}).
		WriteUint256(f.MaximumSupply, func(err error) error {
			return fmt.Errorf("unable to serialize foundry output maximum supply: %w", err)
		}).
		WriteObject(f.TokenScheme, deSeriMode, func(err error) error {
			return fmt.Errorf("unable to serialize foundry output token scheme: %w", err)
		}).
		WriteSliceOfObjects(f.Blocks, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, nil, func(err error) error {
			return fmt.Errorf("unable to serialize foundry output feature blocks: %w", err)
		}).
		Serialize()
}

func (f *FoundryOutput) MarshalJSON() ([]byte, error) {
	var err error
	jFoundryOutput := &jsonFoundryOutput{
		Type:         int(OutputFoundry),
		Amount:       int(f.Amount),
		SerialNumber: int(f.SerialNumber),
	}

	jFoundryOutput.NativeTokens, err = serializablesToJSONRawMsgs(f.NativeTokens)
	if err != nil {
		return nil, err
	}

	jFoundryOutput.Address, err = addressToJSONRawMsg(f.Address)
	if err != nil {
		return nil, err
	}

	jFoundryOutput.TokenTag = hex.EncodeToString(f.TokenTag[:])

	jFoundryOutput.CirculatingSupply = f.CirculatingSupply.String()
	jFoundryOutput.MaximumSupply = f.MaximumSupply.String()

	jTokenSchemeBytes, err := f.TokenScheme.MarshalJSON()
	if err != nil {
		return nil, err
	}
	jsonRawMsgTokenScheme := json.RawMessage(jTokenSchemeBytes)
	jFoundryOutput.TokenScheme = &jsonRawMsgTokenScheme

	jFoundryOutput.Blocks, err = serializablesToJSONRawMsgs(f.Blocks)
	if err != nil {
		return nil, err
	}

	return json.Marshal(jFoundryOutput)
}

func (f *FoundryOutput) UnmarshalJSON(bytes []byte) error {
	jNFTOutput := &jsonNFTOutput{}
	if err := json.Unmarshal(bytes, jNFTOutput); err != nil {
		return err
	}
	seri, err := jNFTOutput.ToSerializable()
	if err != nil {
		return err
	}
	*f = *seri.(*FoundryOutput)
	return nil
}

// jsonFoundryOutput defines the json representation of a FoundryOutput.
type jsonFoundryOutput struct {
	Type              int                `json:"type"`
	Amount            int                `json:"amount"`
	NativeTokens      []*json.RawMessage `json:"nativeTokens"`
	Address           *json.RawMessage   `json:"address"`
	SerialNumber      int                `json:"serialNumber"`
	TokenTag          string             `json:"tokenTag"`
	CirculatingSupply string             `json:"circulatingSupply"`
	MaximumSupply     string             `json:"maximumSupply"`
	TokenScheme       *json.RawMessage   `json:"tokenScheme"`
	Blocks            []*json.RawMessage `json:"blocks"`
}

func (j *jsonFoundryOutput) ToSerializable() (serializer.Serializable, error) {
	var err error
	e := &FoundryOutput{
		Amount:       uint64(j.Amount),
		SerialNumber: uint32(j.SerialNumber),
	}

	e.NativeTokens, err = jsonRawMsgsToSerializables(j.NativeTokens, func(ty int) (JSONSerializable, error) {
		return &jsonNativeToken{}, nil
	})
	if err != nil {
		return nil, err
	}

	e.Address, err = addressFromJSONRawMsg(j.Address)
	if err != nil {
		return nil, err
	}

	tokenTagBytes, err := hex.DecodeString(j.TokenTag)
	if err != nil {
		return nil, err
	}
	copy(e.TokenTag[:], tokenTagBytes)

	var ok bool
	e.CirculatingSupply, ok = new(big.Int).SetString(j.CirculatingSupply, 10)
	if !ok {
		return nil, fmt.Errorf("%w: circluating supply field of foundry output '%s'", ErrDecodeJSONUint256Str, j.CirculatingSupply)
	}

	e.MaximumSupply, ok = new(big.Int).SetString(j.MaximumSupply, 10)
	if !ok {
		return nil, fmt.Errorf("%w: maximum supply field of foundry output '%s'", ErrDecodeJSONUint256Str, j.MaximumSupply)
	}

	jsonTokenScheme, err := DeserializeObjectFromJSON(j.TokenScheme, jsonTokenSchemeSelector)
	if err != nil {
		return nil, fmt.Errorf("unable to decode token scheme from JSON: %w", err)
	}
	e.TokenScheme, err = jsonTokenScheme.ToSerializable()
	if err != nil {
		return nil, err
	}

	e.Blocks, err = jsonRawMsgsToSerializables(j.Blocks, jsonFeatureBlockSelector)
	if err != nil {
		return nil, err
	}

	return e, nil
}
