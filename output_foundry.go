package iotago

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/iotaledger/hive.go/serializer"
)

const (
	// FoundryIDLength is the byte length of a FoundryID consisting out of the alias address, serial number and token scheme.
	FoundryIDLength = AliasAddressSerializedBytesSize + serializer.UInt32ByteSize + serializer.OneByte
)

var (
	// ErrNonUniqueFoundryOutputs gets returned when multiple FoundryOutput(s) with the same FoundryID exist within an OutputsByType.
	ErrNonUniqueFoundryOutputs = errors.New("non unique foundries within outputs")
	// ErrMissingFoundryTransition gets returned when a FoundryDiff is missing for a computation.
	ErrMissingFoundryTransition = errors.New("missing foundry transition")
	// ErrInvalidFoundryTransition gets returned when a foundry transition is invalid.
	ErrInvalidFoundryTransition = errors.New("invalid foundry transition")
	// ErrInvalidFoundryState gets returned when the state between two FoundryOutput(s) is invalid.
	ErrInvalidFoundryState = errors.New("invalid foundry state")
)

// TokenTag is a tag holding some additional data which might be interpreted by higher layers.
type TokenTag = [TokenTagLength]byte

// FoundryID defines the identifier for a foundry consisting out of the address, serial number and TokenScheme.
type FoundryID [FoundryIDLength]byte

func (fID FoundryID) String() string {
	return hex.EncodeToString(fID[:])
}

// FoundryOutputs is a slice of FoundryOutput(s).
type FoundryOutputs []*FoundryOutput

// FoundryOutputsSet is a set of FoundryOutput(s).
type FoundryOutputsSet map[FoundryID]*FoundryOutput

// Diff returns the supply diff between the given sets of FoundryOutput(s).
func (set FoundryOutputsSet) Diff(other FoundryOutputsSet) (FoundryStateDiffs, error) {
	diffs := make(FoundryStateDiffs)
	seen := make(map[FoundryID]struct{})
	for foundryID, foundryOutput := range set {
		seen[foundryID] = struct{}{}
		nativeTokenID, err := foundryOutput.NativeTokenID()
		if err != nil {
			return nil, err
		}

		otherFoundryOutput, has := other[foundryID]
		if !has {
			negCircSupply := new(big.Int)
			diffs[foundryID] = &FoundryDiff{
				NativeTokenID: nativeTokenID,
				SupplyDiff:    negCircSupply.Neg(foundryOutput.CirculatingSupply),
				Transition:    FoundryTransitionDestroyed,
			}
			continue
		}

		if err := foundryOutput.StateChangeOk(otherFoundryOutput); err != nil {
			return nil, err
		}

		diff := new(big.Int)
		diff.Sub(otherFoundryOutput.CirculatingSupply, foundryOutput.CirculatingSupply)
		diffs[foundryID] = &FoundryDiff{
			NativeTokenID: nativeTokenID,
			Transition:    FoundryTransitionStateChange,
		}
	}

	for foundryID, foundryOutput := range other {
		if _, alreadySeen := seen[foundryID]; alreadySeen {
			continue
		}

		nativeTokenID, err := foundryOutput.NativeTokenID()
		if err != nil {
			return nil, err
		}

		diffs[foundryID] = &FoundryDiff{
			NativeTokenID: nativeTokenID,
			SupplyDiff:    foundryOutput.CirculatingSupply,
			Transition:    FoundryTransitionNew,
		}
	}

	return diffs, nil
}

// FoundryStateDiffs defines a map of diffs computed from the circulating supply for a given NativeToken
// when comparing the state transition of a certain foundry.
type FoundryStateDiffs map[FoundryID]*FoundryDiff

func (fsd FoundryStateDiffs) CheckFoundryOutputsSerialNumber(inputs Outputs, outputs Outputs) {
	// TODO
}

// FoundryDiffs is a slice of FoundryDiff(s).
type FoundryDiffs []*FoundryDiff

// FoundryDiff defines the circulating supply diff of the computation when looking at specific FoundryOutput(s) state transition.
type FoundryDiff struct {
	// The NativeTokenID to which this diff applies to.
	NativeTokenID NativeTokenID
	// The circulating supply diff between the FoundryOutput(s).
	// Positive if the supply increased, negative if tokens were burned.
	SupplyDiff *big.Int
	// The type of state transition.
	//	- If the transition is FoundryTransitionNew then SupplyDiff corresponds to the circulating supply.
	//	- If the transition is FoundryTransitionDestroyed then SupplyDiff equals nil.
	// 	- If the transition is FoundryTransitionStateChange, then SupplyDiff represents the actual delta between the states.
	Transition FoundryTransition
}

// FoundryTransition defines the transition type of a FoundryOutput.
type FoundryTransition byte

const (
	// FoundryTransitionNew defines a transition where a foundry is created.
	FoundryTransitionNew FoundryTransition = iota
	// FoundryTransitionStateChange defines a transition where a foundry's state is changed.
	FoundryTransitionStateChange
	// FoundryTransitionDestroyed defines a transition where a foundry is destroyed.
	FoundryTransitionDestroyed
)

// FoundryTransitionToString returns the name for the given FoundryTransition.
func FoundryTransitionToString(tr FoundryTransition) string {
	switch tr {
	case FoundryTransitionNew:
		return "FoundryTransitionNew"
	case FoundryTransitionStateChange:
		return "FoundryTransitionStateChange"
	case FoundryTransitionDestroyed:
		return "FoundryTransitionDestroyed"
	default:
		return "unknown foundry transition"
	}
}

// FoundryOutput is an output type which controls the supply of user defined native tokens.
type FoundryOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount uint64
	// The native tokens held by the output.
	NativeTokens NativeTokens
	// The alias controlling the foundry.
	Address Address
	// The serial number of the foundry.
	SerialNumber uint32
	// The tag which is always the last 12 bytes of the tokens generated by this foundry.
	TokenTag TokenTag
	// The circulating supply of tokens controlled by this foundry.
	CirculatingSupply *big.Int
	// The maximum supply of tokens controlled by this foundry.
	MaximumSupply *big.Int
	// The token scheme this foundry uses.
	TokenScheme TokenScheme
	// The feature blocks which modulate the constraints on the output.
	Blocks FeatureBlocks
}

func (f *FoundryOutput) StateChangeOk(other *FoundryOutput) error {
	srcID, err := f.ID()
	if err != nil {
		return err
	}
	otherID, err := other.ID()
	if err != nil {
		return err
	}

	switch {
	case srcID != otherID:
		return fmt.Errorf("%w: ID mismatch wanted %s but got %s", ErrInvalidFoundryState, srcID, otherID)
	case f.MaximumSupply.Cmp(other.MaximumSupply) != 0:
		return fmt.Errorf("%w: maximum supply mismatch wanted %s but got %s", ErrInvalidFoundryState, f.MaximumSupply, other.MaximumSupply)
	case f.TokenScheme.Type() != other.TokenScheme.Type():
		return fmt.Errorf("%w: token scheme mismatch wanted %s but got %s", ErrInvalidFoundryState, TokenSchemeTypeToString(f.TokenScheme.Type()), TokenSchemeTypeToString(other.TokenScheme.Type()))
	}
	return nil
}

// ID returns the FoundryID of this FoundryOutput.
func (f *FoundryOutput) ID() (FoundryID, error) {
	var foundryID FoundryID
	addrBytes, err := f.Address.Serialize(serializer.DeSeriModePerformValidation)
	if err != nil {
		return foundryID, err
	}
	copy(foundryID[:], addrBytes)
	binary.LittleEndian.PutUint32(foundryID[len(addrBytes):], f.SerialNumber)
	foundryID[len(foundryID)-1] = byte(f.TokenScheme.Type())
	return foundryID, nil
}

// NativeTokenID returns the NativeTokenID this FoundryOutput operates on.
func (f *FoundryOutput) NativeTokenID() (NativeTokenID, error) {
	var nativeTokenID NativeTokenID
	foundryID, err := f.ID()
	if err != nil {
		return nativeTokenID, err
	}
	copy(nativeTokenID[:], foundryID[:])
	copy(nativeTokenID[len(foundryID):], f.TokenTag[:])
	return nativeTokenID, nil
}

func (f *FoundryOutput) NativeTokenSet() NativeTokens {
	return f.NativeTokens
}

func (f *FoundryOutput) FeatureBlocks() FeatureBlocks {
	return f.Blocks
}

func (f *FoundryOutput) Deposit() (uint64, error) {
	return f.Amount, nil
}

func (f *FoundryOutput) Ident() (Address, error) {
	return f.Address, nil
}

func (f *FoundryOutput) Type() OutputType {
	return OutputFoundry
}

func (f *FoundryOutput) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(OutputAlias), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize foundry output: %w", err)
		}).
		ReadNum(&f.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for foundry output: %w", err)
		}).
		ReadObject(&f.Address, deSeriMode, serializer.TypeDenotationByte, AddressSelector, func(err error) error {
			return fmt.Errorf("unable to deserialize address for foundry output: %w", err)
		}).
		ReadSliceOfObjects(&f.NativeTokens, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationNone, func(ty uint32) (serializer.Serializable, error) {
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
		ReadObject(&f.TokenScheme, deSeriMode, serializer.TypeDenotationByte, TokenSchemeSelector, func(err error) error {
			return fmt.Errorf("unable to deserialize token scheme for foundry output: %w", err)
		}).
		ReadSliceOfObjects(&f.Blocks, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationByte, foundryOutputFeatureBlocksGuard, featBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize feature blocks for NFT output: %w", err)
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

func foundryOutputFeatureBlocksGuard(ty uint32) (serializer.Serializable, error) {
	if !featureBlocksSupportedByFoundryOutput(ty) {
		return nil, fmt.Errorf("%w: unable to deserialize foundry output, unsupported feature block type %s", ErrUnsupportedFeatureBlockType, FeatureBlockTypeToString(FeatureBlockType(ty)))
	}
	return FeatureBlockSelector(ty)
}

func featureBlocksSupportedByFoundryOutput(ty uint32) bool {
	switch ty {
	case uint32(FeatureBlockMetadata):
	default:
		return false
	}
	return true
}

func (f *FoundryOutput) Serialize(deSeriMode serializer.DeSerializationMode) ([]byte, error) {
	return serializer.NewSerializer().
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := outputAmountValidator(-1, f); err != nil {
					return fmt.Errorf("%w: unable to serialize foundry output", err)
				}

				if err := isValidAddrType(f.Address); err != nil {
					return fmt.Errorf("invalid address set in foundry output: %w", err)
				}

				if err := featureBlockSupported(f.FeatureBlocks(), featureBlocksSupportedByFoundryOutput); err != nil {
					return fmt.Errorf("invalid feature blocks set in foundry output: %w", err)
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
		WriteSliceOfObjects(&f.NativeTokens, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, nativeTokensArrayRules.ToWrittenObjectConsumer(deSeriMode), func(err error) error {
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
		WriteSliceOfObjects(&f.Blocks, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, featBlockArrayRules.ToWrittenObjectConsumer(deSeriMode), func(err error) error {
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

	jFoundryOutput.NativeTokens, err = serializablesToJSONRawMsgs(f.NativeTokens.ToSerializables())
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

	jFoundryOutput.Blocks, err = serializablesToJSONRawMsgs(f.Blocks.ToSerializables())
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

	e.NativeTokens, err = nativeTokensFromJSONRawMsg(j.NativeTokens)
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

	e.TokenScheme, err = tokenSchemeFromJSONRawMsg(j.TokenScheme)
	if err != nil {
		return nil, err
	}

	e.Blocks, err = featureBlocksFromJSONRawMsg(j.Blocks)
	if err != nil {
		return nil, err
	}

	return e, nil
}
