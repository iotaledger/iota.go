package iotago

import (
	"bytes"
	"crypto/ed25519"
	"fmt"
	"sort"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/blake2b"

	hiveEd25519 "github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

const (
	// BlockIDLength defines the length of a block ID.
	BlockIDLength = SlotIdentifierLength
	// MaxBlockSize defines the maximum size of a block.
	MaxBlockSize = 32768
	// BlockMinStrongParents defines the minimum amount of strong parents in a block.
	BlockMinStrongParents = 1
	// BlockMinParents defines the minimum amount of non-strong parents in a block.
	BlockMinParents = 0
	// BlockMaxParents defines the maximum amount of parents in a block.
	BlockMaxParents = 8
)

// BlockType denotes a type of Block.
type BlockType byte

const (
	BlockTypeBasic     BlockType = 1
	BlockTypeValidator BlockType = 2
)

// EmptyBlockID returns an empty BlockID.
func EmptyBlockID() BlockID {
	return emptySlotIdentifier
}

// BlockID is the ID of a Block.
type BlockID = SlotIdentifier

// BlockIDs are IDs of blocks.
type BlockIDs []BlockID

// ToHex converts the BlockIDs to their hex representation.
func (ids BlockIDs) ToHex() []string {
	hexIDs := make([]string, len(ids))
	for i, id := range ids {
		hexIDs[i] = hexutil.EncodeHex(id[:])
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
		blockID, err := SlotIdentifierFromHexString(hexString)
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

type ProtocolBlock struct {
	ProtocolVersion byte      `serix:"0,mapKey=protocolVersion"`
	NetworkID       NetworkID `serix:"1,mapKey=networkId"`

	IssuingTime         time.Time   `serix:"2,mapKey=issuingTime"`
	SlotCommitment      *Commitment `serix:"3,mapKey=slotCommitment"`
	LatestFinalizedSlot SlotIndex   `serix:"4,mapKey=latestFinalizedSlot"`

	IssuerID AccountID `serix:"5,mapKey=issuerID"`

	Block Block `serix:"6,mapKey=block"`

	Signature Signature `serix:"7,mapKey=signature"`
}

func (b *ProtocolBlock) ContentHash(api API) (Identifier, error) {
	data, err := api.Encode(b)
	if err != nil {
		return Identifier{}, fmt.Errorf("failed to encode block: %w", err)
	}

	return contentHashFromBlockBytes(data)
}

func BlockIdentifierFromBlockBytes(blockBytes []byte) (Identifier, error) {
	contentHash, err := contentHashFromBlockBytes(blockBytes)
	if err != nil {
		return emptyIdentifier, err
	}

	signatureBytes, err := signatureBytesFromBlockBytes(blockBytes)
	if err != nil {
		return emptyIdentifier, err
	}

	return IdentifierFromData(byteutils.ConcatBytes(contentHash[:], signatureBytes[:])), nil
}

func contentHashFromBlockBytes(blockBytes []byte) (Identifier, error) {
	if len(blockBytes) < Ed25519SignatureSerializedBytesSize {
		return Identifier{}, errors.New("not enough block bytes")
	}
	return blake2b.Sum256(blockBytes[:len(blockBytes)-Ed25519SignatureSerializedBytesSize]), nil
}

func signatureBytesFromBlockBytes(blockBytes []byte) ([Ed25519SignatureSerializedBytesSize]byte, error) {
	if len(blockBytes) < Ed25519SignatureSerializedBytesSize {
		return [Ed25519SignatureSerializedBytesSize]byte{}, errors.New("not enough block bytes")
	}
	return [Ed25519SignatureSerializedBytesSize]byte(blockBytes[len(blockBytes)-Ed25519SignatureSerializedBytesSize:]), nil
}

// SigningMessage returns the to be signed message.
// It is the 'encoded(IssuingTime)+encoded(SlotCommitment.ID()+contentHash'.
func (b *ProtocolBlock) SigningMessage(api API) ([]byte, error) {
	contentHash, err := b.ContentHash(api)
	if err != nil {
		return nil, err
	}

	issuingTimeBytes, err := api.Encode(b.IssuingTime)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize block's issuing time: %w", err)
	}

	commitmentID, err := b.SlotCommitment.ID(api)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize block's commitment ID: %w", err)
	}

	return byteutils.ConcatBytes([]byte{b.ProtocolVersion}, b.IssuerID[:], issuingTimeBytes, commitmentID[:], contentHash[:]), nil
}

// Sign produces signatures signing the essence for every given AddressKeys.
// The produced signatures are in the same order as the AddressKeys.
func (b *ProtocolBlock) Sign(api API, addrKey AddressKeys) (Signature, error) {
	signMsg, err := b.SigningMessage(api)
	if err != nil {
		return nil, err
	}

	signer := NewInMemoryAddressSigner(addrKey)

	return signer.Sign(addrKey.Address, signMsg)
}

// VerifySignature verifies the Signature of the block.
func (b *ProtocolBlock) VerifySignature(api API) (valid bool, err error) {
	signingMessage, err := b.SigningMessage(api)
	if err != nil {
		return false, err
	}

	edSig, isEdSig := b.Signature.(*Ed25519Signature)
	if !isEdSig {
		return false, fmt.Errorf("only ed2519 signatures supported, got %s", b.Signature.Type())
	}

	if edSig.PublicKey == [ed25519.PublicKeySize]byte{} {
		return false, fmt.Errorf("empty publicKeys are invalid")
	}

	return hiveEd25519.Verify(edSig.PublicKey[:], signingMessage, edSig.Signature[:]), nil
}

// ID computes the ID of the Block.
func (b *ProtocolBlock) ID(api API) (BlockID, error) {
	data, err := api.Encode(b)
	if err != nil {
		return BlockID{}, fmt.Errorf("can't compute block ID: %w", err)
	}

	slotIndex := api.TimeProvider().SlotFromTime(b.IssuingTime)

	blockIdentifier, err := BlockIdentifierFromBlockBytes(data)
	if err != nil {
		return BlockID{}, err
	}

	return NewSlotIdentifier(slotIndex, blockIdentifier), nil
}

// MustID works like ID but panics if the BlockID can't be computed.
func (b *ProtocolBlock) MustID(api API) BlockID {
	blockID, err := b.ID(api)
	if err != nil {
		panic(err)
	}
	return blockID
}

type Block interface {
	Type() BlockType

	StrongParentIDs() BlockIDs

	WeakParentIDs() BlockIDs

	ShallowLikeParentIDs() BlockIDs
}

// StrongParentsIDs is a slice of BlockIDs the block strongly references.
type strongParentsIDs = BlockIDs

// WeakParentsIDs is a slice of BlockIDs the block weakly references.
type WeakParentsIDs = BlockIDs

// ShallowLikeParentIDs is a slice of BlockIDs the block shallow like references.
type ShallowLikeParentIDs = BlockIDs

// BasicBlock represents a basic vertex in the Tangle/BlockDAG.
type BasicBlock struct {
	// The parents the block references.
	StrongParents      strongParentsIDs     `serix:"0,lengthPrefixType=uint8,mapKey=strongParents"`
	WeakParents        WeakParentsIDs       `serix:"1,lengthPrefixType=uint8,mapKey=weakParents"`
	ShallowLikeParents ShallowLikeParentIDs `serix:"2,lengthPrefixType=uint8,mapKey=shallowLikeParents"`

	// The inner payload of the block. Can be nil.
	Payload BlockPayload `serix:"3,optional,mapKey=payload,omitempty"`

	BurnedMana Mana `serix:"4,mapKey=burnedMana"`
}

func (b *BasicBlock) Type() BlockType {
	return BlockTypeBasic
}

func (b *BasicBlock) StrongParentIDs() BlockIDs {
	return b.StrongParents
}

func (b *BasicBlock) WeakParentIDs() BlockIDs {
	return b.WeakParents
}

func (b *BasicBlock) ShallowLikeParentIDs() BlockIDs {
	return b.ShallowLikeParents
}

// ValidatorBlock represents a validator vertex in the Tangle/BlockDAG.
type ValidatorBlock struct {
	// The parents the block references.
	StrongParents      strongParentsIDs     `serix:"0,lengthPrefixType=uint8,mapKey=strongParents"`
	WeakParents        WeakParentsIDs       `serix:"1,lengthPrefixType=uint8,mapKey=weakParents"`
	ShallowLikeParents ShallowLikeParentIDs `serix:"2,lengthPrefixType=uint8,mapKey=shallowLikeParents"`

	HighestSupportedVersion byte `serix:"3,mapKey=latestFinalizedSlot"`
}

func (b *ValidatorBlock) Type() BlockType {
	return BlockTypeValidator
}

func (b *ValidatorBlock) StrongParentIDs() BlockIDs {
	return b.StrongParents
}

func (b *ValidatorBlock) WeakParentIDs() BlockIDs {
	return b.WeakParents
}

func (b *ValidatorBlock) ShallowLikeParentIDs() BlockIDs {
	return b.ShallowLikeParents
}
