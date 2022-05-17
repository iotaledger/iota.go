package iotago

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/pow"
)

const (
	// BlockIDLength defines the length of a block ID.
	BlockIDLength = blake2b.Size256
	// BlockBinSerializedMaxSize defines the maximum size of a block.
	BlockBinSerializedMaxSize = 32768
	// BlockMinParents defines the minimum amount of parents in a block.
	BlockMinParents = 1
	// BlockMaxParents defines the maximum amount of parents in a block.
	BlockMaxParents = 8
)

var (
	// ErrBlockExceedsMaxSize gets returned when a serialized block exceeds BlockBinSerializedMaxSize.
	ErrBlockExceedsMaxSize = errors.New("block exceeds max size")

	// is an empty block ID
	emptyBlockID = BlockID{}

	blockPayloadGuard = serializer.SerializableGuard{
		ReadGuard: func(ty uint32) (serializer.Serializable, error) {
			switch PayloadType(ty) {
			case PayloadTransaction:
			case PayloadTaggedData:
			case PayloadMilestone:
			default:
				return nil, fmt.Errorf("a block can only contain a transaction/tagged data/milestone but got type ID %d: %w", ty, ErrUnsupportedPayloadType)
			}
			return PayloadSelector(ty)
		},
		WriteGuard: func(seri serializer.Serializable) error {
			if seri == nil {
				return nil
			}
			switch seri.(type) {
			case *Transaction:
			case *TaggedData:
			case *Milestone:
			default:
				return ErrUnsupportedPayloadType
			}
			return nil
		},
	}

	// restrictions around parents within a block.
	blockParentArrayRules = serializer.ArrayRules{
		Min:            BlockMinParents,
		Max:            BlockMaxParents,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates | serializer.ArrayValidationModeLexicalOrdering,
	}
)

// BlockParentArrayRules returns array rules defining the constraints on a slice of block parent references.
func BlockParentArrayRules() serializer.ArrayRules {
	return blockParentArrayRules
}

// EmptyBlockID returns an empty BlockID.
func EmptyBlockID() BlockID {
	return emptyBlockID
}

// BlockID is the ID of a Block.
type BlockID [BlockIDLength]byte

func (b BlockID) MarshalBinary() (data []byte, err error) {
	// copy
	return b[:], nil
}

// ToHex converts the given block ID to their hex representation.
func (b *BlockID) ToHex() string {
	return EncodeHex(b[:])
}

// Empty tells whether the BlockID is empty.
func (b *BlockID) Empty() bool {
	return *b == emptyBlockID
}

// BlockIDFromHexString converts the given block ID from its hex to BlockID representation.
func BlockIDFromHexString(blockIDHex string) (BlockID, error) {
	blockIDBytes, err := DecodeHex(blockIDHex)
	if err != nil {
		return BlockID{}, err
	}

	var blockID BlockID
	copy(blockID[:], blockIDBytes)
	return blockID, nil
}

// MustBlockIDFromHexString converts the given block ID from its hex to BlockID representation.
func MustBlockIDFromHexString(blockIDHex string) BlockID {
	blockID, err := BlockIDFromHexString(blockIDHex)
	if err != nil {
		panic(err)
	}
	return blockID
}

// BlockIDs are IDs of blocks.
type BlockIDs []BlockID

// ToHex converts the BlockIDs to their hex representation.
func (ids BlockIDs) ToHex() []string {
	hexIDs := make([]string, len(ids))
	for i, id := range ids {
		hexIDs[i] = EncodeHex(id[:])
	}
	return hexIDs
}

func (ids BlockIDs) ToSerializerType() serializer.SliceOfArraysOf32Bytes {
	result := make(serializer.SliceOfArraysOf32Bytes, len(ids))
	for i, ele := range ids {
		result[i] = ele
	}
	return result
}

// RemoveDupsAndSort removes duplicated BlockIDs and sorts the slice by the lexical ordering.
func (ids BlockIDs) RemoveDupsAndSort() BlockIDs {
	sorted := make(serializer.LexicalOrdered32ByteArrays, len(ids))
	for i, id := range ids {
		sorted[i] = id
	}
	sort.Sort(sorted)

	var result BlockIDs
	var prev BlockID
	for i, id := range sorted {
		// only add to the result, if it its different from its predecessor
		if i == 0 || !bytes.Equal(prev[:], id[:]) {
			result = append(result, id)
		}
		prev = id
	}
	return result
}

// BlockIDsFromHexString converts the given block IDs from their hex to BlockID representation.
func BlockIDsFromHexString(blockIDsHex []string) (BlockIDs, error) {
	result := make(BlockIDs, len(blockIDsHex))

	for i, hexString := range blockIDsHex {
		blockID, err := BlockIDFromHexString(hexString)
		if err != nil {
			return nil, err
		}
		result[i] = blockID
	}

	return result, nil
}

// Block represents a vertex in the Tangle.
type Block struct {
	// The protocol version under which this block operates.
	ProtocolVersion byte
	// The parents the block references.
	Parents BlockIDs
	// The inner payload of the block. Can be nil.
	Payload Payload
	// The nonce which lets this block fulfill the PoW requirements.
	Nonce uint64
}

// ID computes the ID of the Block.
func (m *Block) ID() (BlockID, error) {
	data, err := m.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return BlockID{}, fmt.Errorf("can't compute block ID: %w", err)
	}
	h := blake2b.Sum256(data)
	return h, nil
}

// MustID works like ID but panics if the BlockID can't be computed.
func (m *Block) MustID() BlockID {
	blockID, err := m.ID()
	if err != nil {
		panic(err)
	}
	return blockID
}

// POW computes the PoW score of the Block.
func (m *Block) POW() (float64, error) {
	data, err := m.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return 0, fmt.Errorf("can't compute block PoW score: %w", err)
	}
	return pow.Score(data), nil
}

func (m *Block) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	if len(data) > BlockBinSerializedMaxSize {
		return 0, fmt.Errorf("%w: size %d bytes", ErrBlockExceedsMaxSize, len(data))
	}
	parentsSlice := serializer.SliceOfArraysOf32Bytes{}
	return serializer.NewDeserializer(data).
		ReadNum(&m.ProtocolVersion, func(err error) error {
			return fmt.Errorf("unable to deserialize block protocol version: %w", err)
		}).
		ReadSliceOfArraysOf32Bytes(&parentsSlice, deSeriMode, serializer.SeriLengthPrefixTypeAsByte, &blockParentArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize block parents: %w", err)
		}).
		Do(func() {
			m.Parents = make(BlockIDs, len(parentsSlice))
			for i, ele := range parentsSlice {
				m.Parents[i] = ele
			}
		}).
		ReadPayload(&m.Payload, deSeriMode, deSeriCtx, blockPayloadGuard.ReadGuard, func(err error) error {
			return fmt.Errorf("unable to deserialize block's inner payload: %w", err)
		}).
		ReadNum(&m.Nonce, func(err error) error {
			return fmt.Errorf("unable to deserialize block nonce: %w", err)
		}).
		ConsumedAll(func(leftOver int, err error) error {
			return fmt.Errorf("%w: unable to deserialize block: %d bytes are still available", err, leftOver)
		}).
		Done()
}

func (m *Block) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	data, err := serializer.NewSerializer().
		Do(func() {
			if deSeriMode.HasMode(serializer.DeSeriModePerformLexicalOrdering) {
				m.Parents = m.Parents.RemoveDupsAndSort()
			}
		}).
		WriteNum(m.ProtocolVersion, func(err error) error {
			return fmt.Errorf("unable to serialize block protocol version: %w", err)
		}).
		Write32BytesArraySlice(m.Parents.ToSerializerType(), deSeriMode, serializer.SeriLengthPrefixTypeAsByte, &blockParentArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize block parents: %w", err)
		}).
		WritePayload(m.Payload, deSeriMode, deSeriCtx, blockPayloadGuard.WriteGuard, func(err error) error {
			return fmt.Errorf("unable to serialize block inner payload: %w", err)
		}).
		WriteNum(m.Nonce, func(err error) error {
			return fmt.Errorf("unable to serialize block nonce: %w", err)
		}).
		Serialize()
	if err != nil {
		return nil, err
	}
	if len(data) > BlockBinSerializedMaxSize {
		return nil, fmt.Errorf("%w: size %d bytes", ErrBlockExceedsMaxSize, len(data))
	}
	return data, nil
}

func (m *Block) MarshalJSON() ([]byte, error) {
	jBlock := &jsonBlock{
		ProtocolVersion: int(m.ProtocolVersion),
	}
	jBlock.Parents = make([]string, len(m.Parents))
	for i, parent := range m.Parents {
		jBlock.Parents[i] = EncodeHex(parent[:])
	}
	jBlock.Nonce = EncodeUint64(m.Nonce)
	if m.Payload != nil {
		jsonPayload, err := m.Payload.MarshalJSON()
		if err != nil {
			return nil, err
		}
		rawMsgJsonPayload := json.RawMessage(jsonPayload)
		jBlock.Payload = &rawMsgJsonPayload
	}
	return json.Marshal(jBlock)
}

func (m *Block) UnmarshalJSON(bytes []byte) error {
	jBlock := &jsonBlock{}
	if err := json.Unmarshal(bytes, jBlock); err != nil {
		return err
	}
	seri, err := jBlock.ToSerializable()
	if err != nil {
		return err
	}
	*m = *seri.(*Block)
	return nil
}

// jsonBlock defines the JSON representation of a Block.
type jsonBlock struct {
	ProtocolVersion int `json:"protocolVersion"`
	// The hex encoded IDs of the referenced parent blocks.
	Parents []string `json:"parents"`
	// The payload within the block.
	Payload *json.RawMessage `json:"payload,omitempty"`
	// The nonce the block used to fulfill the PoW requirement.
	Nonce string `json:"nonce"`
}

func (jm *jsonBlock) ToSerializable() (serializer.Serializable, error) {
	var err error

	m := &Block{
		ProtocolVersion: byte(jm.ProtocolVersion),
	}

	var parsedNonce uint64
	if len(jm.Nonce) != 0 {
		parsedNonce, err = DecodeUint64(jm.Nonce)
		if err != nil {
			return nil, fmt.Errorf("unable to parse block nonce from JSON: %w", err)
		}
	}
	m.Nonce = parsedNonce

	m.Parents = make(BlockIDs, len(jm.Parents))
	for i, jparent := range jm.Parents {
		parentBytes, err := DecodeHex(jparent)
		if err != nil {
			return nil, fmt.Errorf("unable to decode hex parent %d from JSON: %w", i+1, err)
		}

		copy(m.Parents[i][:], parentBytes)
	}

	if jm.Payload != nil {
		m.Payload, err = payloadFromJSONRawMsg(jm.Payload)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}
