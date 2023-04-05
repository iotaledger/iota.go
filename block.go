package iotago

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"runtime"
	"sort"
	"time"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/pow"
	"github.com/iotaledger/iota.go/v4/slot"
)

const (
	// BlockIDLength defines the length of a block ID.
	BlockIDLength = blake2b.Size256 + serializer.UInt64ByteSize
	// MaxBlockSize defines the maximum size of a block.
	MaxBlockSize = 32768
	// BlockMinParents defines the minimum amount of parents in a block.
	BlockMinParents = 1
	// BlockMaxParents defines the maximum amount of parents in a block.
	BlockMaxParents = 8
)

var (
	// is an empty block ID.
	emptyBlockID = BlockID{}

	// restrictions around parents within a block.
	blockParentArrayRules = serializer.ArrayRules{
		Min:            BlockMinParents,
		Max:            BlockMaxParents,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates | serializer.ArrayValidationModeLexicalOrdering,
	}

	/*
		non-strong min parents = 0
		max = 8
	*/

	//TODO: weak parents must be disjunct to the rest of the parents
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

func (id BlockID) MarshalText() (text []byte, err error) {
	dst := make([]byte, hex.EncodedLen(len(BlockID{})))
	hex.Encode(dst, id[:])
	return dst, nil
}

func (id *BlockID) UnmarshalText(text []byte) error {
	_, err := hex.Decode(id[:], text)
	return err
}

// ToHex converts the given block ID to their hex representation.
func (id BlockID) ToHex() string {
	return EncodeHex(id[:])
}

// Empty tells whether the BlockID is empty.
func (id BlockID) Empty() bool {
	return id == emptyBlockID
}

func (id *BlockID) String() string {
	return id.ToHex()
}

func (id *BlockID) Slot() slot.Index {
	return slot.Index(binary.LittleEndian.Uint64(id[32:]))
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

// RemoveDupsAndSort removes duplicated BlockIDs and sorts the slice by the lexical ordering.
func (ids BlockIDs) RemoveDupsAndSort() BlockIDs {
	sorted := append(BlockIDs{}, ids...)
	sort.Slice(sorted, func(i, j int) bool {
		return bytes.Compare(sorted[i][:], sorted[j][:]) == -1
	})

	var result BlockIDs
	var prev BlockID
	for i, id := range sorted {
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

type BlockPayload interface {
	Payload
}

// Block represents a vertex in the Tangle.
type Block struct {
	// The protocol version under which this block operates.
	ProtocolVersion byte `serix:"0,mapKey=protocolVersion"`

	NetworkID NetworkID `serix:"1,mapKey=networkId"`

	// The parents the block references.
	StrongParents      BlockIDs `serix:"2,lengthPrefixType=uint8,mapKey=strongParents"`
	WeakParents        BlockIDs `serix:"3,lengthPrefixType=uint8,mapKey=weakParents"`
	ShallowLikeParents BlockIDs `serix:"4,lengthPrefixType=uint8,mapKey=shallowLikeParents"`

	IssuerID        types.Identifier  `serix:"5,mapKey=issuerID"`
	IssuerPublicKey ed25519.PublicKey `serix:"6,mapKey=issuerPublicKey"`
	IssuingTime     time.Time         `serix:"7,mapKey=issuingTime"`

	SlotCommitment      *Commitment `serix:"8,mapKey=slotCommitment"`
	LatestConfirmedSlot slot.Index  `serix:"9,mapKey=LatestConfirmedSlot"`

	// The inner payload of the block. Can be nil.
	Payload BlockPayload `serix:"10,optional,mapKey=payload,omitempty"`

	// The nonce which lets this block fulfill the PoW requirements.
	Nonce uint64 `serix:"11,mapKey=nonce"`

	Signature Ed25519Signature `serix:"12,mapKey=signature"`
}

//BLock id == 0x + hex(hash) + hex(slotIndex) // 40 bytes

// ID computes the ID of the Block.
func (b *Block) ID(slotTimeProvider *slot.TimeProvider) (BlockID, error) {
	data, err := internalEncode(b)
	if err != nil {
		return BlockID{}, fmt.Errorf("can't compute block ID: %w", err)
	}
	h := blake2b.Sum256(data)

	slotIndex := slotTimeProvider.IndexFromTime(b.IssuingTime)

	id := BlockID{}
	copy(id[:], h[:])
	binary.LittleEndian.PutUint64(id[32:], uint64(slotIndex))

	return id, nil
}

// MustID works like ID but panics if the BlockID can't be computed.
func (b *Block) MustID(slotTimeProvider *slot.TimeProvider) BlockID {
	blockID, err := b.ID(slotTimeProvider)
	if err != nil {
		panic(err)
	}
	return blockID
}

// POW computes the PoW score of the Block.
func (b *Block) POW() (float64, []byte, error) {
	data, err := internalEncode(b)
	if err != nil {
		return 0, nil, fmt.Errorf("can't compute block PoW score: %w", err)
	}
	return pow.Score(data), data, nil
}

// DoPOW executes the proof-of-work required to fulfill the targetScore.
// Use the given context to cancel proof-of-work.
func (b *Block) DoPOW(ctx context.Context, targetScore float64) error {
	data, err := internalEncode(b)
	if err != nil {
		return fmt.Errorf("can't compute block PoW score: %w", err)
	}
	powRelevantData := data[:len(data)-8]
	worker := pow.New(runtime.NumCPU())
	nonce, err := worker.Mine(ctx, powRelevantData, targetScore)
	if err != nil {
		return err
	}
	b.Nonce = nonce
	return nil
}
