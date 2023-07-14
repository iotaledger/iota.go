package iotago

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"runtime"
	"sort"
	"time"

	"golang.org/x/crypto/blake2b"

	hiveEd25519 "github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
	"github.com/iotaledger/iota.go/v4/hexutil"
	"github.com/iotaledger/iota.go/v4/pow"
	"github.com/iotaledger/iota.go/v4/util"
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

// StrongParentsIDs is a slice of BlockIDs the block strongly references.
type StrongParentsIDs = BlockIDs

// WeakParentsIDs is a slice of BlockIDs the block weakly references.
type WeakParentsIDs = BlockIDs

// ShallowLikeParentIDs is a slice of BlockIDs the block shallow like references.
type ShallowLikeParentIDs = BlockIDs

// Block represents a vertex in the Tangle.
type Block struct {
	// The protocol version under which this block operates.
	ProtocolVersion byte `serix:"0,mapKey=protocolVersion"`

	NetworkID NetworkID `serix:"1,mapKey=networkId"`

	// The parents the block references.
	StrongParents      StrongParentsIDs     `serix:"2,lengthPrefixType=uint8,mapKey=strongParents"`
	WeakParents        WeakParentsIDs       `serix:"3,lengthPrefixType=uint8,mapKey=weakParents"`
	ShallowLikeParents ShallowLikeParentIDs `serix:"4,lengthPrefixType=uint8,mapKey=shallowLikeParents"`

	IssuerID    AccountID `serix:"5,mapKey=issuerID"`
	IssuingTime time.Time `serix:"6,mapKey=issuingTime"`

	SlotCommitment      *Commitment `serix:"7,mapKey=slotCommitment"`
	LatestFinalizedSlot SlotIndex   `serix:"8,mapKey=latestFinalizedSlot"`

	// The inner payload of the block. Can be nil.
	Payload BlockPayload `serix:"9,optional,mapKey=payload,omitempty"`

	BurnedMana Mana `serix:"10,mapKey=burnedMana"`

	Signature Signature `serix:"11,mapKey=signature"`

	// The nonce which lets this block fulfill the PoW requirements.
	Nonce uint64 `serix:"12,mapKey=nonce"`
}

func (b *Block) ContentHash() (Identifier, error) {
	data, err := internalEncode(b)
	if err != nil {
		return Identifier{}, ierrors.Errorf("failed to encode block: %w", err)
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

	nonceBytes := blockBytes[len(blockBytes)-serializer.UInt64ByteSize:]

	return IdentifierFromData(byteutils.ConcatBytes(contentHash[:], signatureBytes[:], nonceBytes[:])), nil
}

func contentHashFromBlockBytes(blockBytes []byte) (Identifier, error) {
	if len(blockBytes) < Ed25519SignatureSerializedBytesSize+serializer.UInt64ByteSize {
		return Identifier{}, ierrors.New("not enough block bytes")
	}
	return blake2b.Sum256(blockBytes[:len(blockBytes)-Ed25519SignatureSerializedBytesSize-serializer.UInt64ByteSize]), nil
}

func signatureBytesFromBlockBytes(blockBytes []byte) ([Ed25519SignatureSerializedBytesSize]byte, error) {
	if len(blockBytes) < Ed25519SignatureSerializedBytesSize+serializer.UInt64ByteSize {
		return [Ed25519SignatureSerializedBytesSize]byte{}, ierrors.New("not enough block bytes")
	}
	return [Ed25519SignatureSerializedBytesSize]byte(blockBytes[len(blockBytes)-Ed25519SignatureSerializedBytesSize-serializer.UInt64ByteSize:]), nil
}

// SigningMessage returns the to be signed message.
// It is the 'encoded(IssuingTime)+encoded(SlotCommitment.ID()+contentHash'.
func (b *Block) SigningMessage() ([]byte, error) {
	contentHash, err := b.ContentHash()
	if err != nil {
		return nil, err
	}

	issuingTimeBytes, err := internalEncode(b.IssuingTime)
	if err != nil {
		return nil, ierrors.Errorf("failed to serialize block's issuing time: %w", err)
	}

	commitmentID, err := b.SlotCommitment.ID()
	if err != nil {
		return nil, ierrors.Errorf("failed to serialize block's commitment ID: %w", err)
	}

	return byteutils.ConcatBytes(issuingTimeBytes, commitmentID[:], contentHash[:]), nil
}

// Sign produces signatures signing the essence for every given AddressKeys.
// The produced signatures are in the same order as the AddressKeys.
func (b *Block) Sign(addrKey AddressKeys) (Signature, error) {
	signMsg, err := b.SigningMessage()
	if err != nil {
		return nil, err
	}

	signer := NewInMemoryAddressSigner(addrKey)

	return signer.Sign(addrKey.Address, signMsg)
}

// VerifySignature verifies the Signature of the block.
func (b *Block) VerifySignature() (valid bool, err error) {
	signingMessage, err := b.SigningMessage()
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
func (b *Block) ID(timeProvider *TimeProvider) (BlockID, error) {
	data, err := internalEncode(b)
	if err != nil {
		return BlockID{}, ierrors.Errorf("can't compute block ID: %w", err)
	}

	slotIndex := timeProvider.SlotFromTime(b.IssuingTime)

	blockIdentifier, err := BlockIdentifierFromBlockBytes(data)
	if err != nil {
		return BlockID{}, err
	}

	return NewSlotIdentifier(slotIndex, blockIdentifier), nil
}

// MustID works like ID but panics if the BlockID can't be computed.
func (b *Block) MustID(timeProvider *TimeProvider) BlockID {
	blockID, err := b.ID(timeProvider)
	if err != nil {
		panic(err)
	}
	return blockID
}

// POW computes the PoW score of the Block.
func (b *Block) POW() (float64, []byte, error) {
	data, err := internalEncode(b)
	if err != nil {
		return 0, nil, ierrors.Errorf("can't compute block PoW score: %w", err)
	}
	return pow.Score(data), data, nil
}

// DoPOW executes the proof-of-work required to fulfill the targetScore.
// Use the given context to cancel proof-of-work.
func (b *Block) DoPOW(ctx context.Context, targetScore float64) error {
	data, err := internalEncode(b)
	if err != nil {
		return ierrors.Errorf("can't compute block PoW score: %w", err)
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
