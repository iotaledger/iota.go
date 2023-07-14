package iotago

import (
	"bytes"
	"crypto/ed25519"
	"fmt"
	"sort"
	"time"

	"golang.org/x/crypto/blake2b"

	hiveEd25519 "github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
	"github.com/iotaledger/iota.go/v4/hexutil"
	"github.com/iotaledger/iota.go/v4/util"
)

const (
	// BlockIDLength defines the length of a block ID.
	BlockIDLength = SlotIdentifierLength
	// MaxBlockSize defines the maximum size of a block.
	MaxBlockSize = 32768
	// BlockMaxParents defines the maximum amount of parents in a block.
	BlockMaxParents = 8
	// BlockTypeValidationMaxParents defines the maximum amount of parents in a ValidationBlock. TODO: replace number with committee size.
	BlockTypeValidationMaxParents = BlockMaxParents + 42
)

// BlockType denotes a type of Block.
type BlockType byte

const (
	BlockTypeBasic      BlockType = 1
	BlockTypeValidation BlockType = 2
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

const BlockHeaderLength = 1 + serializer.UInt64ByteSize + serializer.UInt64ByteSize + CommitmentIDLength + serializer.UInt64ByteSize + AccountIDLength

type BlockHeader struct {
	ProtocolVersion Version   `serix:"0,mapKey=protocolVersion"`
	NetworkID       NetworkID `serix:"1,mapKey=networkId"`

	IssuingTime         time.Time    `serix:"2,mapKey=issuingTime"`
	SlotCommitmentID    CommitmentID `serix:"3,mapKey=slotCommitment"`
	LatestFinalizedSlot SlotIndex    `serix:"4,mapKey=latestFinalizedSlot"`

	IssuerID AccountID `serix:"5,mapKey=issuerID"`
}

func (b *BlockHeader) Hash(api API) (Identifier, error) {
	headerBytes, err := api.Encode(b)
	if err != nil {
		return Identifier{}, ierrors.Errorf("failed to serialize block header: %w", err)
	}

	return blake2b.Sum256(headerBytes), nil
}

type ProtocolBlock struct {
	BlockHeader `serix:"0"`

	Block Block `serix:"1,mapKey=block"`

	Signature Signature `serix:"2,mapKey=signature"`
}

func BlockIdentifierFromBlockBytes(blockBytes []byte) (Identifier, error) {
	if len(blockBytes) < BlockHeaderLength+Ed25519SignatureSerializedBytesSize {
		return Identifier{}, ierrors.New("not enough block bytes")
	}

	length := len(blockBytes)
	// Separate into header hash, block hash and signature bytes so that we are able to recompute the BlockID from an Attestation.
	headerHash := blake2b.Sum256(blockBytes[:BlockHeaderLength])
	blockHash := blake2b.Sum256(blockBytes[BlockHeaderLength : length-Ed25519SignatureSerializedBytesSize])
	signatureBytes := [Ed25519SignatureSerializedBytesSize]byte(blockBytes[length-Ed25519SignatureSerializedBytesSize:])

	return blockIdentifier(headerHash, blockHash, signatureBytes[:]), nil
}

func blockIdentifier(headerHash Identifier, blockHash Identifier, signatureBytes []byte) Identifier {
	return IdentifierFromData(byteutils.ConcatBytes(headerHash[:], blockHash[:], signatureBytes))
}

// SigningMessage returns the to be signed message.
// The BlockHeader and Block are separately hashed and concatenated to enable the verification of the signature for
// an Attestation where only the BlockHeader and the hash of Block is known.
func (b *ProtocolBlock) SigningMessage(api API) ([]byte, error) {
	headerHash, err := b.BlockHeader.Hash(api)
	if err != nil {
		return nil, err
	}

	blockHash, err := b.Block.Hash(api)
	if err != nil {
		return nil, err
	}

	return blockSigningMessage(headerHash, blockHash), nil
}

func blockSigningMessage(headerHash Identifier, blockHash Identifier) []byte {
	return byteutils.ConcatBytes(headerHash[:], blockHash[:])
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
		return false, ierrors.Errorf("only ed2519 signatures supported, got %s", b.Signature.Type())
	}

	if edSig.PublicKey == [ed25519.PublicKeySize]byte{} {
		return false, ierrors.New("empty publicKeys are invalid")
	}

	return hiveEd25519.Verify(edSig.PublicKey[:], signingMessage, edSig.Signature[:]), nil
}

// ID computes the ID of the Block.
func (b *ProtocolBlock) ID(api API) (BlockID, error) {
	data, err := api.Encode(b)
	if err != nil {
		return BlockID{}, ierrors.Errorf("can't compute block ID: %w", err)
	}

	id, err := BlockIdentifierFromBlockBytes(data)
	if err != nil {
		return BlockID{}, err
	}

	slotIndex := api.TimeProvider().SlotFromTime(b.IssuingTime)

	return NewSlotIdentifier(slotIndex, id), nil
}

// MustID works like ID but panics if the BlockID can't be computed.
func (b *ProtocolBlock) MustID(api API) BlockID {
	blockID, err := b.ID(api)
	if err != nil {
		panic(err)
	}
	return blockID
}

func (b *ProtocolBlock) Parents() (parents []BlockID) {
	parents = make([]BlockID, 0)

	parents = append(parents, b.Block.StrongParentIDs()...)
	parents = append(parents, b.Block.WeakParentIDs()...)
	parents = append(parents, b.Block.ShallowLikeParentIDs()...)

	return parents
}

func (b *ProtocolBlock) ParentsWithType() (parents []Parent) {
	parents = make([]Parent, 0)

	for _, parentBlockID := range b.Block.StrongParentIDs() {
		parents = append(parents, Parent{parentBlockID, StrongParentType})
	}

	for _, parentBlockID := range b.Block.WeakParentIDs() {
		parents = append(parents, Parent{parentBlockID, WeakParentType})
	}

	for _, parentBlockID := range b.Block.ShallowLikeParentIDs() {
		parents = append(parents, Parent{parentBlockID, ShallowLikeParentType})
	}

	return parents
}

// ForEachParent executes a consumer func for each parent.
func (b *ProtocolBlock) ForEachParent(consumer func(parent Parent)) {
	for _, parent := range b.ParentsWithType() {
		consumer(parent)
	}
}

type Block interface {
	Type() BlockType

	StrongParentIDs() BlockIDs
	WeakParentIDs() BlockIDs
	ShallowLikeParentIDs() BlockIDs

	Hash(api API) (Identifier, error)
}

// BasicBlock represents a basic vertex in the Tangle/BlockDAG.
type BasicBlock struct {
	// The parents the block references.
	StrongParents      BlockIDs `serix:"0,lengthPrefixType=uint8,mapKey=strongParents,minLen=1,maxLen=8"`
	WeakParents        BlockIDs `serix:"1,lengthPrefixType=uint8,mapKey=weakParents,minLen=0,maxLen=8"`
	ShallowLikeParents BlockIDs `serix:"2,lengthPrefixType=uint8,mapKey=shallowLikeParents,minLen=0,maxLen=8"`

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

func (b *BasicBlock) Hash(api API) (Identifier, error) {
	blockBytes, err := api.Encode(b)
	if err != nil {
		return Identifier{}, ierrors.Errorf("failed to serialize basic block: %w", err)
	}

	return blake2b.Sum256(blockBytes), nil
}

// ValidationBlock represents a validation vertex in the Tangle/BlockDAG.
type ValidationBlock struct {
	// The parents the block references.
	StrongParents      BlockIDs `serix:"0,lengthPrefixType=uint8,mapKey=strongParents,minLen=1,maxLen=50"`
	WeakParents        BlockIDs `serix:"1,lengthPrefixType=uint8,mapKey=weakParents,minLen=0,maxLen=50"`
	ShallowLikeParents BlockIDs `serix:"2,lengthPrefixType=uint8,mapKey=shallowLikeParents,minLen=0,maxLen=50"`

	HighestSupportedVersion Version `serix:"3,mapKey=highestSupportedVersion"`
	// ProtocolParametersHash is the hash of the protocol parameters for the HighestSupportedVersion.
	ProtocolParametersHash Identifier `serix:"4,mapKey=protocolParametersHash"`
}

func (b *ValidationBlock) Type() BlockType {
	return BlockTypeValidation
}

func (b *ValidationBlock) StrongParentIDs() BlockIDs {
	return b.StrongParents
}

func (b *ValidationBlock) WeakParentIDs() BlockIDs {
	return b.WeakParents
}

func (b *ValidationBlock) ShallowLikeParentIDs() BlockIDs {
	return b.ShallowLikeParents
}

func (b *ValidationBlock) Hash(api API) (Identifier, error) {
	blockBytes, err := api.Encode(b)
	if err != nil {
		return Identifier{}, ierrors.Errorf("failed to serialize validation block: %w", err)
	}

	return IdentifierFromData(blockBytes), nil
}

// ParentsType is a type that defines the type of the parent.
type ParentsType uint8

const (
	// UndefinedParentType is the undefined parent.
	UndefinedParentType ParentsType = iota
	// StrongParentType is the ParentsType for a strong parent.
	StrongParentType
	// WeakParentType is the ParentsType for a weak parent.
	WeakParentType
	// ShallowLikeParentType is the ParentsType for the shallow like parent.
	ShallowLikeParentType
)

// String returns string representation of ParentsType.
func (p ParentsType) String() string {
	return fmt.Sprintf("ParentType(%s)", []string{"Undefined", "Strong", "Weak", "Shallow Like"}[p])
}

// Parent is a parent that can be either strong or weak.
type Parent struct {
	ID   BlockID
	Type ParentsType
}

func (b *Block) WorkScore(workScoreStructure *WorkScoreStructure) WorkScore {
	// Work Score for parents is a penalty for each missing strong parent below MinStrongParentsThreshold.
	var parentWorkScore WorkScore
	if byte(len(b.StrongParents)) < workScoreStructure.MinStrongParentsThreshold {
		parentWorkScore += workScoreStructure.Factors.MissingParent.Multiply(int(workScoreStructure.MinStrongParentsThreshold - byte(len(b.StrongParents))))
	}
	return parentWorkScore +
		// ProtocolVersion and NetworkID
		workScoreStructure.Factors.Data.Multiply(serializer.OneByte+serializer.UInt64ByteSize) +
		// IssuerID and IssuingTime
		workScoreStructure.Factors.Data.Multiply(AccountIDLength+serializer.UInt64ByteSize) +
		// SlotCommitment and LatestFinalizedSlot
		workScoreStructure.Factors.Data.Multiply(util.NumByteLen(b.SlotCommitment)+serializer.UInt64ByteSize) +
		b.Payload.WorkScore(workScoreStructure) +
		// BurnedMana
		workScoreStructure.Factors.Data.Multiply(ManaSize) +
		// Signature
		b.Signature.WorkScore(workScoreStructure)
}
