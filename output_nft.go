package iotago

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

const (
	// 	NFTIDLength is the byte length of an NFTID.
	NFTIDLength = blake2b.Size256
)

var (
	emptyNFTID = [NFTIDLength]byte{}

	nftOutputUnlockCondsArrayRules = &serializer.ArrayRules{
		Min: 1, Max: 4,
		MustOccur: serializer.TypePrefixes{
			uint32(UnlockConditionAddress): struct{}{},
		},
		Guards: serializer.SerializableGuard{
			ReadGuard: func(ty uint32) (serializer.Serializable, error) {
				switch ty {
				case uint32(UnlockConditionAddress):
				case uint32(UnlockConditionStorageDepositReturn):
				case uint32(UnlockConditionTimelock):
				case uint32(UnlockConditionExpiration):
				default:
					return nil, fmt.Errorf("%w: unable to deserialize NFT output, unsupported unlock condition type %s", ErrUnsupportedUnlockConditionType, UnlockConditionType(ty))
				}
				return UnlockConditionSelector(ty)
			},
			WriteGuard: func(seri serializer.Serializable) error {
				switch seri.(type) {
				case *AddressUnlockCondition:
				case *StorageDepositReturnUnlockCondition:
				case *TimelockUnlockCondition:
				case *ExpirationUnlockCondition:
				default:
					return fmt.Errorf("%w: in NFT output", ErrUnsupportedUnlockConditionType)
				}
				return nil
			},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	nftOutputFeatBlockArrayRules = &serializer.ArrayRules{
		Min: 0,
		Max: 3,
		Guards: serializer.SerializableGuard{
			ReadGuard: func(ty uint32) (serializer.Serializable, error) {
				switch ty {
				case uint32(FeatureSender):
				case uint32(FeatureMetadata):
				case uint32(FeatureTag):
				default:
					return nil, fmt.Errorf("%w: unable to deserialize NFT output, unsupported feature type %s", ErrUnsupportedFeatureType, FeatureType(ty))
				}
				return FeatureSelector(ty)
			},
			WriteGuard: func(seri serializer.Serializable) error {
				switch seri.(type) {
				case *SenderFeature:
				case *MetadataFeature:
				case *TagFeature:
				default:
					return fmt.Errorf("%w: in NFT output", ErrUnsupportedFeatureType)
				}
				return nil
			},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	nftOutputImmFeatBlockArrayRules = &serializer.ArrayRules{
		Min: 0,
		Max: 2,
		Guards: serializer.SerializableGuard{
			ReadGuard: func(ty uint32) (serializer.Serializable, error) {
				switch ty {
				case uint32(FeatureIssuer):
				case uint32(FeatureMetadata):
				default:
					return nil, fmt.Errorf("%w: unable to deserialize NFT output, unsupported immutable feature type %s", ErrUnsupportedFeatureType, FeatureType(ty))
				}
				return FeatureSelector(ty)
			},
			WriteGuard: func(seri serializer.Serializable) error {
				switch seri.(type) {
				case *IssuerFeature:
				case *MetadataFeature:
				default:
					return fmt.Errorf("%w: in NFT output", ErrUnsupportedFeatureType)
				}
				return nil
			},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}
)

// NFTOutputFeaturesArrayRules returns array rules defining the constraints on Features within an NFTOutput.
func NFTOutputFeaturesArrayRules() serializer.ArrayRules {
	return *nftOutputFeatBlockArrayRules
}

// NFTOutputImmutableFeaturesArrayRules returns array rules defining the constraints on immutable Features within an NFTOutput.
func NFTOutputImmutableFeaturesArrayRules() serializer.ArrayRules {
	return *nftOutputImmFeatBlockArrayRules
}

// NFTID is the identifier for an NFT.
// It is computed as the Blake2b-256 hash of the OutputID of the output which created the NFT.
type NFTID [NFTIDLength]byte

func (nftID NFTID) ToHex() string {
	return EncodeHex(nftID[:])
}

// NFTIDs are NFTID(s).
type NFTIDs []NFTID

func (nftID NFTID) Addressable() bool {
	return true
}

func (nftID NFTID) Key() interface{} {
	return nftID.String()
}

func (nftID NFTID) FromOutputID(id OutputID) ChainID {
	addr := NFTAddressFromOutputID(id)
	return addr.Chain()
}

func (nftID NFTID) Empty() bool {
	return nftID == emptyNFTID
}

func (nftID NFTID) Matches(other ChainID) bool {
	otherNFTID, isNFTID := other.(NFTID)
	if !isNFTID {
		return false
	}
	return nftID == otherNFTID
}

func (nftID NFTID) ToAddress() ChainConstrainedAddress {
	var addr NFTAddress
	copy(addr[:], nftID[:])
	return &addr
}

func (nftID NFTID) String() string {
	return EncodeHex(nftID[:])
}

func NFTIDFromOutputID(o OutputID) NFTID {
	ret := NFTID{}
	addr := NFTAddressFromOutputID(o)
	copy(ret[:], addr[:])
	return ret
}

// NFTOutput is an output type used to implement non-fungible tokens.
type NFTOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount uint64
	// The native tokens held by the output.
	NativeTokens NativeTokens
	// The identifier of this NFT.
	NFTID NFTID
	// The unlock conditions on this output.
	Conditions UnlockConditions
	// The feature on the output.
	Features Features
	// The immutable feature on the output.
	ImmutableFeatures Features
}

func (n *NFTOutput) Clone() Output {
	return &NFTOutput{
		Amount:            n.Amount,
		NativeTokens:      n.NativeTokens.Clone(),
		NFTID:             n.NFTID,
		Conditions:        n.Conditions.Clone(),
		Features:          n.Features.Clone(),
		ImmutableFeatures: n.ImmutableFeatures.Clone(),
	}
}

func (n *NFTOutput) Ident() Address {
	return n.Conditions.MustSet().Address().Address
}

func (n *NFTOutput) UnlockableBy(ident Address, extParas *ExternalUnlockParameters) bool {
	ok, _ := outputUnlockable(n, nil, ident, extParas)
	return ok
}

func (n *NFTOutput) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	return outputOffsetVByteCost(rentStruct) +
		// prefix + amount
		rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+serializer.UInt64ByteSize) +
		n.NativeTokens.VBytes(rentStruct, nil) +
		rentStruct.VBFactorData.Multiply(NFTIDLength) +
		n.Conditions.VBytes(rentStruct, nil) +
		n.Features.VBytes(rentStruct, nil) +
		n.ImmutableFeatures.VBytes(rentStruct, nil)
}

func (n *NFTOutput) ValidateStateTransition(transType ChainTransitionType, next ChainConstrainedOutput, semValCtx *SemanticValidationContext) error {
	var err error
	switch transType {
	case ChainTransitionTypeGenesis:
		err = n.genesisValid(semValCtx)
	case ChainTransitionTypeStateChange:
		err = n.stateChangeValid(next)
	case ChainTransitionTypeDestroy:
		return nil
	default:
		panic("unknown chain transition type in NFTOutput")
	}
	if err != nil {
		return &ChainTransitionError{Inner: err, Msg: fmt.Sprintf("NFT %s", n.NFTID)}
	}
	return nil
}

func (n *NFTOutput) genesisValid(semValCtx *SemanticValidationContext) error {
	if !n.NFTID.Empty() {
		return fmt.Errorf("NFTOutput's ID is not zeroed even though it is new")
	}
	return IsIssuerOnOutputUnlocked(n, semValCtx.WorkingSet.UnlockedIdents)
}

func (n *NFTOutput) stateChangeValid(next ChainConstrainedOutput) error {
	nextState, is := next.(*NFTOutput)
	if !is {
		return fmt.Errorf("NFTOutput can only state transition to another NFTOutput")
	}
	if !n.ImmutableFeatures.Equal(nextState.ImmutableFeatures) {
		return fmt.Errorf("old state %s, next state %s", n.ImmutableFeatures, nextState.ImmutableFeatures)
	}
	return nil
}

func (n *NFTOutput) Chain() ChainID {
	return n.NFTID
}

func (n *NFTOutput) NativeTokenSet() NativeTokens {
	return n.NativeTokens
}

func (n *NFTOutput) FeaturesSet() FeaturesSet {
	return n.Features.MustSet()
}

func (n *NFTOutput) UnlockConditionsSet() UnlockConditionsSet {
	return n.Conditions.MustSet()
}

func (n *NFTOutput) ImmutableFeaturesSet() FeaturesSet {
	return n.ImmutableFeatures.MustSet()
}

func (n *NFTOutput) Deposit() uint64 {
	return n.Amount
}

func (n *NFTOutput) Type() OutputType {
	return OutputNFT
}

func (n *NFTOutput) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(OutputNFT), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize NFT output: %w", err)
		}).
		ReadNum(&n.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for NFT output: %w", err)
		}).
		ReadSliceOfObjects(&n.NativeTokens, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationNone, nativeTokensArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize native tokens for NFT output: %w", err)
		}).
		ReadBytesInPlace(n.NFTID[:], func(err error) error {
			return fmt.Errorf("unable to deserialize NFT ID for NFT output: %w", err)
		}).
		ReadSliceOfObjects(&n.Conditions, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationByte, nftOutputUnlockCondsArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize unlock conditions for NFT output: %w", err)
		}).
		ReadSliceOfObjects(&n.Features, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationByte, nftOutputFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize features for NFT output: %w", err)
		}).
		ReadSliceOfObjects(&n.ImmutableFeatures, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationByte, nftOutputImmFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize immutable features for NFT output: %w", err)
		}).
		Done()
}

func (n *NFTOutput) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(byte(OutputNFT), func(err error) error {
			return fmt.Errorf("unable to serialize NFT output type ID: %w", err)
		}).
		WriteNum(n.Amount, func(err error) error {
			return fmt.Errorf("unable to serialize NFT output amount: %w", err)
		}).
		WriteSliceOfObjects(&n.NativeTokens, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, nativeTokensArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize NFT output native tokens: %w", err)
		}).
		WriteBytes(n.NFTID[:], func(err error) error {
			return fmt.Errorf("unable to serialize NFT output NFT ID: %w", err)
		}).
		WriteSliceOfObjects(&n.Conditions, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, nftOutputUnlockCondsArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize NFT output unlock conditions: %w", err)
		}).
		WriteSliceOfObjects(&n.Features, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, nftOutputFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize NFT output features: %w", err)
		}).
		WriteSliceOfObjects(&n.ImmutableFeatures, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, nftOutputImmFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize NFT output immutable features: %w", err)
		}).
		Serialize()
}

func (n *NFTOutput) Size() int {
	return util.NumByteLen(byte(OutputNFT)) +
		util.NumByteLen(n.Amount) +
		n.NativeTokens.Size() +
		NFTIDLength +
		n.Conditions.Size() +
		n.Features.Size() +
		n.ImmutableFeatures.Size()
}

func (n *NFTOutput) MarshalJSON() ([]byte, error) {
	var err error
	jNFTOutput := &jsonNFTOutput{
		Type:   int(OutputNFT),
		Amount: EncodeUint64(n.Amount),
	}

	jNFTOutput.NativeTokens, err = serializablesToJSONRawMsgs(n.NativeTokens.ToSerializables())
	if err != nil {
		return nil, err
	}

	jNFTOutput.NFTID = EncodeHex(n.NFTID[:])

	jNFTOutput.Conditions, err = serializablesToJSONRawMsgs(n.Conditions.ToSerializables())
	if err != nil {
		return nil, err
	}

	jNFTOutput.Features, err = serializablesToJSONRawMsgs(n.Features.ToSerializables())
	if err != nil {
		return nil, err
	}

	jNFTOutput.ImmutableFeatures, err = serializablesToJSONRawMsgs(n.ImmutableFeatures.ToSerializables())
	if err != nil {
		return nil, err
	}

	return json.Marshal(jNFTOutput)
}

func (n *NFTOutput) UnmarshalJSON(bytes []byte) error {
	jNFTOutput := &jsonNFTOutput{}
	if err := json.Unmarshal(bytes, jNFTOutput); err != nil {
		return err
	}
	seri, err := jNFTOutput.ToSerializable()
	if err != nil {
		return err
	}
	*n = *seri.(*NFTOutput)
	return nil
}

// jsonNFTOutput defines the json representation of a NFTOutput.
type jsonNFTOutput struct {
	Type              int                `json:"type"`
	Amount            string             `json:"amount"`
	NativeTokens      []*json.RawMessage `json:"nativeTokens,omitempty"`
	NFTID             string             `json:"nftId"`
	Conditions        []*json.RawMessage `json:"unlockConditions,omitempty"`
	Features          []*json.RawMessage `json:"features,omitempty"`
	ImmutableFeatures []*json.RawMessage `json:"immutableFeatures,omitempty"`
}

func (j *jsonNFTOutput) ToSerializable() (serializer.Serializable, error) {
	var err error
	e := &NFTOutput{}

	e.Amount, err = DecodeUint64(j.Amount)
	if err != nil {
		return nil, err
	}

	e.NativeTokens, err = nativeTokensFromJSONRawMsg(j.NativeTokens)
	if err != nil {
		return nil, err
	}

	nftIDBytes, err := DecodeHex(j.NFTID)
	if err != nil {
		return nil, err
	}
	copy(e.NFTID[:], nftIDBytes)

	e.Conditions, err = unlockConditionsFromJSONRawMsg(j.Conditions)
	if err != nil {
		return nil, err
	}

	e.Features, err = featuresFromJSONRawMsg(j.Features)
	if err != nil {
		return nil, err
	}

	e.ImmutableFeatures, err = featuresFromJSONRawMsg(j.ImmutableFeatures)
	if err != nil {
		return nil, err
	}

	return e, nil
}
