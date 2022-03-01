package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

const (
	// 	NFTIDLength = 20 is the byte length of an NFTID.
	NFTIDLength = 20
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
				case uint32(FeatureBlockSender):
				case uint32(FeatureBlockMetadata):
				case uint32(FeatureBlockTag):
				default:
					return nil, fmt.Errorf("%w: unable to deserialize nft output, unsupported feature block type %s", ErrUnsupportedFeatureBlockType, FeatureBlockType(ty))
				}
				return FeatureBlockSelector(ty)
			},
			WriteGuard: func(seri serializer.Serializable) error {
				switch seri.(type) {
				case *SenderFeatureBlock:
				case *IssuerFeatureBlock:
				case *MetadataFeatureBlock:
				case *TagFeatureBlock:
				default:
					return fmt.Errorf("%w: in nft output", ErrUnsupportedFeatureBlockType)
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
				case uint32(FeatureBlockIssuer):
				case uint32(FeatureBlockMetadata):
				default:
					return nil, fmt.Errorf("%w: unable to deserialize nft output, unsupported immutable feature block type %s", ErrUnsupportedFeatureBlockType, FeatureBlockType(ty))
				}
				return FeatureBlockSelector(ty)
			},
			WriteGuard: func(seri serializer.Serializable) error {
				switch seri.(type) {
				case *IssuerFeatureBlock:
				case *MetadataFeatureBlock:
				default:
					return fmt.Errorf("%w: in nft output", ErrUnsupportedFeatureBlockType)
				}
				return nil
			},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}
)

// NFTOutputFeatureBlocksArrayRules returns array rules defining the constraints on FeatureBlocks within an NFTOutput.
func NFTOutputFeatureBlocksArrayRules() serializer.ArrayRules {
	return *nftOutputFeatBlockArrayRules
}

// NFTOutputImmutableFeatureBlocksArrayRules returns array rules defining the constraints on immutable FeatureBlocks within an NFTOutput.
func NFTOutputImmutableFeatureBlocksArrayRules() serializer.ArrayRules {
	return *nftOutputImmFeatBlockArrayRules
}

// NFTID is the identifier for an NFT.
// It is computed as the Blake2b-160 hash of the OutputID of the output which created the NFT.
type NFTID [NFTIDLength]byte

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
	// The feature blocks on the output.
	Blocks FeatureBlocks
	// The immutable feature blocks on the output.
	ImmutableBlocks FeatureBlocks
}

func (n *NFTOutput) Clone() Output {
	return &NFTOutput{
		Amount:          n.Amount,
		NativeTokens:    n.NativeTokens.Clone(),
		NFTID:           n.NFTID,
		Conditions:      n.Conditions.Clone(),
		Blocks:          n.Blocks.Clone(),
		ImmutableBlocks: n.ImmutableBlocks.Clone(),
	}
}

func (n *NFTOutput) Ident() Address {
	return n.Conditions.MustSet().Address().Address
}

func (n *NFTOutput) UnlockableBy(ident Address, extParas *ExternalUnlockParameters) bool {
	ok, _ := outputUnlockable(n, nil, ident, extParas)
	return ok
}

func (n *NFTOutput) VByteCost(costStruct *RentStructure, _ VByteCostFunc) uint64 {
	return outputOffsetVByteCost(costStruct) +
		// prefix + amount
		costStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+serializer.UInt64ByteSize) +
		n.NativeTokens.VByteCost(costStruct, nil) +
		costStruct.VBFactorData.Multiply(NFTIDLength) +
		n.Conditions.VByteCost(costStruct, nil) +
		n.Blocks.VByteCost(costStruct, nil) +
		n.ImmutableBlocks.VByteCost(costStruct, nil)
}

func (n *NFTOutput) ValidateStateTransition(transType ChainTransitionType, next ChainConstrainedOutput, semValCtx *SemanticValidationContext) error {
	switch transType {
	case ChainTransitionTypeGenesis:
		if !n.NFTID.Empty() {
			return fmt.Errorf("%w: NFTOutput's ID is not zeroed even though it is new", ErrInvalidChainStateTransition)
		}
		return IsIssuerOnOutputUnlocked(n, semValCtx.WorkingSet.UnlockedIdents)
	case ChainTransitionTypeStateChange:
		nextState, is := next.(*NFTOutput)
		if !is {
			return fmt.Errorf("%w: NFTOutput can only state transition to another NFTOutput", ErrInvalidChainStateTransition)
		}
		if !n.ImmutableBlocks.Equal(nextState.ImmutableBlocks) {
			return fmt.Errorf("%w: old state %s, next state %s", ErrInvalidChainStateTransition, n.ImmutableBlocks, nextState.ImmutableBlocks)
		}
		return nil
	case ChainTransitionTypeDestroy:
		return nil
	default:
		panic("unknown chain transition type in NFTOutput")
	}
}

func (n *NFTOutput) Chain() ChainID {
	return n.NFTID
}

func (n *NFTOutput) NativeTokenSet() NativeTokens {
	return n.NativeTokens
}

func (n *NFTOutput) UnlockConditions() UnlockConditions {
	return n.Conditions
}

func (n *NFTOutput) FeatureBlocks() FeatureBlocks {
	return n.Blocks
}

func (n *NFTOutput) ImmutableFeatureBlocks() FeatureBlocks {
	return n.ImmutableBlocks
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
		ReadSliceOfObjects(&n.Blocks, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationByte, nftOutputFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize feature blocks for NFT output: %w", err)
		}).
		ReadSliceOfObjects(&n.ImmutableBlocks, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationByte, nftOutputImmFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize immutable feature blocks for NFT output: %w", err)
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
		WriteSliceOfObjects(&n.Blocks, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, nftOutputFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize NFT output feature blocks: %w", err)
		}).
		WriteSliceOfObjects(&n.ImmutableBlocks, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, nftOutputImmFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize NFT output immutable feature blocks: %w", err)
		}).
		Serialize()
}

func (n *NFTOutput) Size() int {
	return util.NumByteLen(byte(OutputNFT)) +
		util.NumByteLen(n.Amount) +
		n.NativeTokens.Size() +
		NFTIDLength +
		n.Conditions.Size() +
		n.Blocks.Size() +
		n.ImmutableBlocks.Size()
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

	jNFTOutput.Blocks, err = serializablesToJSONRawMsgs(n.Blocks.ToSerializables())
	if err != nil {
		return nil, err
	}

	jNFTOutput.ImmutableBlocks, err = serializablesToJSONRawMsgs(n.ImmutableBlocks.ToSerializables())
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
	Type            int                `json:"type"`
	Amount          string             `json:"amount"`
	NativeTokens    []*json.RawMessage `json:"nativeTokens"`
	NFTID           string             `json:"nftId"`
	Conditions      []*json.RawMessage `json:"unlockConditions"`
	Blocks          []*json.RawMessage `json:"featureBlocks"`
	ImmutableBlocks []*json.RawMessage `json:"immutableFeatureBlocks"`
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

	e.Blocks, err = featureBlocksFromJSONRawMsg(j.Blocks)
	if err != nil {
		return nil, err
	}

	e.ImmutableBlocks, err = featureBlocksFromJSONRawMsg(j.ImmutableBlocks)
	if err != nil {
		return nil, err
	}

	return e, nil
}
