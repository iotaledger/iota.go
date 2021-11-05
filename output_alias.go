package iotago

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/iotaledger/hive.go/serializer"
)

const (
	// AliasIDLength is the byte length of an AliasID.
	AliasIDLength = 20
)

var (
	emptyAliasID = [AliasIDLength]byte{}
)

// AliasID is the identifier for an alias account.
// It is computed as the Blake2b-160 hash of the OutputID of the output which created the account.
type AliasID = [AliasIDLength]byte

// AliasOutput is an output type which represents an alias account.
type AliasOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount uint64
	// The native tokens held by the output.
	NativeTokens serializer.Serializables
	// The identifier for this alias account.
	AliasID AliasID
	// The entity which is allowed to control this alias account state.
	StateController serializer.Serializable
	// The entity which is allowed to govern this alias account.
	GovernanceController serializer.Serializable
	// The index of the state.
	StateIndex uint32
	// The state of the alias account which can only be mutated by the state controller.
	StateMetadata []byte
	// The counter that denotes the number of foundries created by this alias account.
	FoundryCounter uint32
	// The feature blocks which modulate the constraints on the output.
	Blocks serializer.Serializables
}

func (a *AliasOutput) NativeTokenSet() serializer.Serializables {
	return a.NativeTokens
}

func (a *AliasOutput) FeatureBlocks() serializer.Serializables {
	return a.Blocks
}

func (a *AliasOutput) Deposit() (uint64, error) {
	return a.Amount, nil
}

func (a *AliasOutput) Target() (serializer.Serializable, error) {
	addr := new(AliasAddress)
	copy(addr[:], a.AliasID[:])
	return addr, nil
}

func (a *AliasOutput) Type() OutputType {
	return OutputAlias
}

func (a *AliasOutput) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	return serializer.NewDeserializer(data).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := serializer.CheckTypeByte(data, OutputAlias); err != nil {
					return fmt.Errorf("unable to deserialize alias output: %w", err)
				}
			}
			return nil
		}).
		Skip(serializer.SmallTypeDenotationByteSize, func(err error) error {
			return fmt.Errorf("unable to skip alias output type during deserialization: %w", err)
		}).
		ReadNum(&a.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for alias output: %w", err)
		}).
		ReadSliceOfObjects(func(seri serializer.Serializables) { a.NativeTokens = seri }, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationNone, func(ty uint32) (serializer.Serializable, error) {
			return &NativeToken{}, nil
		}, nativeTokensArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize native tokens for alias output: %w", err)
		}).
		ReadArrayOf20Bytes(&a.AliasID, func(err error) error {
			return fmt.Errorf("unable to deserialize alias ID for alias output: %w", err)
		}).
		ReadObject(func(seri serializer.Serializable) { a.StateController = seri }, deSeriMode, serializer.TypeDenotationByte, AddressSelector, func(err error) error {
			return fmt.Errorf("unable to deserialize state controller for alias output: %w", err)
		}).
		ReadObject(func(seri serializer.Serializable) { a.GovernanceController = seri }, deSeriMode, serializer.TypeDenotationByte, AddressSelector, func(err error) error {
			return fmt.Errorf("unable to deserialize governance controller for alias output: %w", err)
		}).
		ReadNum(&a.StateIndex, func(err error) error {
			return fmt.Errorf("unable to deserialize state index for alias output: %w", err)
		}).
		ReadVariableByteSlice(&a.StateMetadata, serializer.SeriLengthPrefixTypeAsUint32, func(err error) error {
			// TODO: replace max read with actual variable
			return fmt.Errorf("unable to deserialize state metadata for alias output: %w", err)
		}, MaxMetadataLength).
		ReadNum(&a.FoundryCounter, func(err error) error {
			return fmt.Errorf("unable to deserialize foundry counter for alias output: %w", err)
		}).
		ReadSliceOfObjects(func(seri serializer.Serializables) { a.Blocks = seri }, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationByte, func(ty uint32) (serializer.Serializable, error) {
			if !featureBlocksSupportedByAliasOutput(ty) {
				return nil, fmt.Errorf("%w: unable to deserialize alias output, unsupported feature block type %s", ErrUnsupportedFeatureBlockType, FeatureBlockTypeToString(ty))
			}
			return FeatureBlockSelector(ty)
		}, featBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize feature blocks for NFT output: %w", err)
		}).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := outputAmountValidator(-1, a); err != nil {
					return fmt.Errorf("%w: unable to deserialize alias output", err)
				}
			}
			return nil
		}).
		Done()
}

func featureBlocksSupportedByAliasOutput(ty uint32) bool {
	switch ty {
	case uint32(FeatureBlockIssuer):
	case uint32(FeatureBlockMetadata):
	default:
		return false
	}
	return true
}

func (a *AliasOutput) Serialize(deSeriMode serializer.DeSerializationMode) ([]byte, error) {
	return serializer.NewSerializer().
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := outputAmountValidator(-1, a); err != nil {
					return fmt.Errorf("%w: unable to serialize alias output", err)
				}

				if err := isValidAddrType(a.StateController); err != nil {
					return fmt.Errorf("invalid state controller set in alias output: %w", err)
				}
				if err := isValidAddrType(a.GovernanceController); err != nil {
					return fmt.Errorf("invalid governance controller set in alias output: %w", err)
				}

				if err := featureBlockSupported(a.FeatureBlocks(), featureBlocksSupportedByAliasOutput); err != nil {
					return fmt.Errorf("invalid feature blocks set in alias output: %w", err)
				}
			}
			return nil
		}).
		Do(func() {
			if deSeriMode.HasMode(serializer.DeSeriModePerformLexicalOrdering) {
				sort.Sort(serializer.SortedSerializables(a.NativeTokens))
			}
		}).
		WriteNum(OutputAlias, func(err error) error {
			return fmt.Errorf("unable to serialize alias output type ID: %w", err)
		}).
		WriteNum(a.Amount, func(err error) error {
			return fmt.Errorf("unable to serialize alias output amount: %w", err)
		}).
		WriteSliceOfObjects(a.NativeTokens, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, nativeTokensArrayRules.ToWrittenObjectConsumer(deSeriMode), func(err error) error {
			return fmt.Errorf("unable to serialize alias output native tokens: %w", err)
		}).
		WriteBytes(a.AliasID[:], func(err error) error {
			return fmt.Errorf("unable to serialize alias output alias ID: %w", err)
		}).
		WriteObject(a.StateController, deSeriMode, func(err error) error {
			return fmt.Errorf("unable to serialize alias output state controller: %w", err)
		}).
		WriteObject(a.GovernanceController, deSeriMode, func(err error) error {
			return fmt.Errorf("unable to serialize alias output governance controller: %w", err)
		}).
		WriteNum(a.StateIndex, func(err error) error {
			return fmt.Errorf("unable to serialize alias output state index: %w", err)
		}).
		WriteVariableByteSlice(a.StateMetadata, serializer.SeriLengthPrefixTypeAsUint32, func(err error) error {
			return fmt.Errorf("unable to serialize alias output state metadata: %w", err)
		}).
		WriteNum(a.FoundryCounter, func(err error) error {
			return fmt.Errorf("unable to serialize alias output foundry counter: %w", err)
		}).
		WriteSliceOfObjects(a.Blocks, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, featBlockArrayRules.ToWrittenObjectConsumer(deSeriMode), func(err error) error {
			return fmt.Errorf("unable to serialize alias output feature blocks: %w", err)
		}).
		Serialize()
}

func (a *AliasOutput) MarshalJSON() ([]byte, error) {
	var err error
	jAliasOutput := &jsonAliasOutput{
		Type:           int(OutputAlias),
		Amount:         int(a.Amount),
		StateIndex:     int(a.StateIndex),
		FoundryCounter: int(a.FoundryCounter),
	}

	jAliasOutput.NativeTokens, err = serializablesToJSONRawMsgs(a.NativeTokens)
	if err != nil {
		return nil, err
	}

	jAliasOutput.AliasID = hex.EncodeToString(a.AliasID[:])

	jAliasOutput.StateController, err = addressToJSONRawMsg(a.StateController)
	if err != nil {
		return nil, err
	}

	jAliasOutput.GovernanceController, err = addressToJSONRawMsg(a.GovernanceController)
	if err != nil {
		return nil, err
	}

	jAliasOutput.StateMetadata = hex.EncodeToString(a.StateMetadata)

	jAliasOutput.Blocks, err = serializablesToJSONRawMsgs(a.Blocks)
	if err != nil {
		return nil, err
	}

	return json.Marshal(jAliasOutput)
}

func (a *AliasOutput) UnmarshalJSON(bytes []byte) error {
	jAliasOutput := &jsonAliasOutput{}
	if err := json.Unmarshal(bytes, jAliasOutput); err != nil {
		return err
	}
	seri, err := jAliasOutput.ToSerializable()
	if err != nil {
		return err
	}
	*a = *seri.(*AliasOutput)
	return nil
}

// jsonAliasOutput defines the json representation of an AliasOutput.
type jsonAliasOutput struct {
	Type                 int                `json:"type"`
	Amount               int                `json:"amount"`
	NativeTokens         []*json.RawMessage `json:"nativeTokens"`
	AliasID              string             `json:"aliasId"`
	StateController      *json.RawMessage   `json:"stateController"`
	GovernanceController *json.RawMessage   `json:"governanceController"`
	StateIndex           int                `json:"stateIndex"`
	StateMetadata        string             `json:"stateMetadata"`
	FoundryCounter       int                `json:"foundryCounter"`
	Blocks               []*json.RawMessage `json:"blocks"`
}

func (j *jsonAliasOutput) ToSerializable() (serializer.Serializable, error) {
	var err error
	e := &AliasOutput{
		Amount:         uint64(j.Amount),
		StateIndex:     uint32(j.StateIndex),
		FoundryCounter: uint32(j.FoundryCounter),
	}

	e.NativeTokens, err = jsonRawMsgsToSerializables(j.NativeTokens, func(ty int) (JSONSerializable, error) {
		return &jsonNativeToken{}, nil
	})
	if err != nil {
		return nil, err
	}

	aliasIDSlice, err := hex.DecodeString(j.AliasID)
	if err != nil {
		return nil, err
	}
	copy(e.AliasID[:], aliasIDSlice)

	e.StateController, err = addressFromJSONRawMsg(j.StateController)
	if err != nil {
		return nil, err
	}

	e.GovernanceController, err = addressFromJSONRawMsg(j.GovernanceController)
	if err != nil {
		return nil, err
	}

	e.StateMetadata, err = hex.DecodeString(j.StateMetadata)
	if err != nil {
		return nil, err
	}

	e.Blocks, err = jsonRawMsgsToSerializables(j.Blocks, jsonFeatureBlockSelector)
	if err != nil {
		return nil, err
	}

	return e, nil
}
