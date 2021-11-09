package iotago

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/iotaledger/hive.go/serializer"
)

const (
	// 	NFTIDLength = 20 is the byte length of an NFTID.
	NFTIDLength = 20
)

var (
	emptyNFTID = [NFTIDLength]byte{}
)

// NFTID is the identifier for an NFT.
// It is computed as the Blake2b-160 hash of the OutputID of the output which created the NFT.
type NFTID = [NFTIDLength]byte

// NFTOutput is an output type used to implement non-fungible tokens.
type NFTOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount uint64
	// The native tokens held by the output.
	NativeTokens NativeTokens
	// The actual address.
	Address Address
	// The identifier of this NFT.
	NFTID NFTID
	// Arbitrary immutable binary data attached to this NFT.
	ImmutableMetadata []byte
	// The feature blocks which modulate the constraints on the output.
	Blocks FeatureBlocks
}

func (n *NFTOutput) NativeTokenSet() NativeTokens {
	return n.NativeTokens
}

func (n *NFTOutput) FeatureBlocks() FeatureBlocks {
	return n.Blocks
}

func (n *NFTOutput) Deposit() (uint64, error) {
	return n.Amount, nil
}

func (n *NFTOutput) Target() (serializer.Serializable, error) {
	return n.Address, nil
}

func (n *NFTOutput) Type() OutputType {
	return OutputNFT
}

func (n *NFTOutput) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(OutputNFT), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize NFT output: %w", err)
		}).
		ReadNum(&n.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for NFT output: %w", err)
		}).
		ReadSliceOfObjects(&n.NativeTokens, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationNone, func(ty uint32) (serializer.Serializable, error) {
			return &NativeToken{}, nil
		}, nativeTokensArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize native tokens for NFT output: %w", err)
		}).
		ReadObject(&n.Address, deSeriMode, serializer.TypeDenotationByte, AddressSelector, func(err error) error {
			return fmt.Errorf("unable to deserialize address for NFT output: %w", err)
		}).
		ReadArrayOf20Bytes(&n.NFTID, func(err error) error {
			return fmt.Errorf("unable to deserialize NFT ID for NFT output: %w", err)
		}).
		ReadVariableByteSlice(&n.ImmutableMetadata, serializer.SeriLengthPrefixTypeAsUint32, func(err error) error {
			return fmt.Errorf("unable to deserialize immutable metadata for NFT output: %w", err)
		}, MessageBinSerializedMaxSize).
		ReadSliceOfObjects(&n.Blocks, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationByte, nftOutputFeatureBlocksGuard, featBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize feature blocks for NFT output: %w", err)
		}).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := outputAmountValidator(-1, n); err != nil {
					return fmt.Errorf("%w: unable to deserialize NFT output", err)
				}
			}
			return nil
		}).
		Done()
}

func nftOutputFeatureBlocksGuard(ty uint32) (serializer.Serializable, error) {
	return FeatureBlockSelector(ty)
}

func (n *NFTOutput) Serialize(deSeriMode serializer.DeSerializationMode) ([]byte, error) {
	return serializer.NewSerializer().
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := outputAmountValidator(-1, n); err != nil {
					return fmt.Errorf("%w: unable to serialize NFT output", err)
				}

				if err := isValidAddrType(n.Address); err != nil {
					return fmt.Errorf("invalid address set in NFT output: %w", err)
				}
			}
			return nil
		}).
		Do(func() {
			if deSeriMode.HasMode(serializer.DeSeriModePerformLexicalOrdering) {
				seris := n.NativeTokens.ToSerializables()
				sort.Sort(serializer.SortedSerializables(seris))
				n.NativeTokens.FromSerializables(seris)
			}
		}).
		WriteNum(OutputNFT, func(err error) error {
			return fmt.Errorf("unable to serialize NFT output type ID: %w", err)
		}).
		WriteNum(n.Amount, func(err error) error {
			return fmt.Errorf("unable to serialize NFT output amount: %w", err)
		}).
		WriteSliceOfObjects(&n.NativeTokens, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, nativeTokensArrayRules.ToWrittenObjectConsumer(deSeriMode), func(err error) error {
			return fmt.Errorf("unable to serialize NFT output native tokens: %w", err)
		}).
		WriteObject(n.Address, deSeriMode, func(err error) error {
			return fmt.Errorf("unable to serialize NFT output address: %w", err)
		}).
		WriteBytes(n.NFTID[:], func(err error) error {
			return fmt.Errorf("unable to serialize NFT output NFT ID: %w", err)
		}).
		WriteVariableByteSlice(n.ImmutableMetadata, serializer.SeriLengthPrefixTypeAsUint32, func(err error) error {
			return fmt.Errorf("unable to serialize NFT output immutable metadata: %w", err)
		}).
		WriteSliceOfObjects(&n.Blocks, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, featBlockArrayRules.ToWrittenObjectConsumer(deSeriMode), func(err error) error {
			return fmt.Errorf("unable to serialize NFT output feature blocks: %w", err)
		}).
		Serialize()
}

func (n *NFTOutput) MarshalJSON() ([]byte, error) {
	var err error
	jNFTOutput := &jsonNFTOutput{
		Type:   int(OutputNFT),
		Amount: int(n.Amount),
	}

	jNFTOutput.NativeTokens, err = serializablesToJSONRawMsgs(n.NativeTokens.ToSerializables())
	if err != nil {
		return nil, err
	}

	jNFTOutput.Address, err = addressToJSONRawMsg(n.Address)
	if err != nil {
		return nil, err
	}

	jNFTOutput.NFTID = hex.EncodeToString(n.NFTID[:])
	jNFTOutput.ImmutableData = hex.EncodeToString(n.ImmutableMetadata[:])

	jNFTOutput.Blocks, err = serializablesToJSONRawMsgs(n.Blocks.ToSerializables())
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
	Type          int                `json:"type"`
	Amount        int                `json:"amount"`
	NativeTokens  []*json.RawMessage `json:"nativeTokens"`
	Address       *json.RawMessage   `json:"address"`
	NFTID         string             `json:"nftId"`
	ImmutableData string             `json:"immutableData"`
	Blocks        []*json.RawMessage `json:"blocks"`
}

func (j *jsonNFTOutput) ToSerializable() (serializer.Serializable, error) {
	var err error
	e := &NFTOutput{
		Amount: uint64(j.Amount),
	}

	e.NativeTokens, err = nativeTokensFromJSONRawMsg(j.NativeTokens)
	if err != nil {
		return nil, err
	}

	e.Address, err = addressFromJSONRawMsg(j.Address)
	if err != nil {
		return nil, err
	}

	nftIDBytes, err := hex.DecodeString(j.NFTID)
	if err != nil {
		return nil, err
	}
	copy(e.NFTID[:], nftIDBytes)

	immuDataBytes, err := hex.DecodeString(j.ImmutableData)
	if err != nil {
		return nil, err
	}
	copy(e.ImmutableMetadata[:], immuDataBytes)

	e.Blocks, err = featureBlocksFromJSONRawMsg(j.Blocks)
	if err != nil {
		return nil, err
	}

	return e, nil
}
