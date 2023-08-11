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
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
	"github.com/iotaledger/iota.go/v4/hexutil"
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

var (
	ErrWeakParentsInvalid                 = ierrors.New("weak parents must be disjunct to the rest of the parents")
	ErrCommitmentTooOld                   = ierrors.New("a block cannot commit to a slot that is older than the block's slot minus maxCommittableAge")
	ErrCommitmentTooRecent                = ierrors.New("a block cannot commit to a slot that is more recent than the block's slot minus minCommittableAge")
	ErrCommitmentInputTooOld              = ierrors.New("a block cannot contain a commitment input with index older than the block's slot minus maxCommittableAge")
	ErrCommitmentInputTooRecent           = ierrors.New("a block cannot contain a commitment input with index more recent than the block's slot minus minCommittableAge")
	ErrInvalidBlockVersion                = ierrors.New("block has invalid protocol version")
	ErrCommitmentInputNewerThanCommitment = ierrors.New("a block cannot contain a commitment input with index newer than the commitment index")
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

const BlockHeaderLength = serializer.OneByte + serializer.UInt64ByteSize + serializer.UInt64ByteSize + CommitmentIDLength + SlotIndexLength + AccountIDLength

type BlockHeader struct {
	ProtocolVersion Version   `serix:"0,mapKey=protocolVersion"`
	NetworkID       NetworkID `serix:"1,mapKey=networkId"`

	IssuingTime         time.Time    `serix:"2,mapKey=issuingTime"`
	SlotCommitmentID    CommitmentID `serix:"3,mapKey=slotCommitment"`
	LatestFinalizedSlot SlotIndex    `serix:"4,mapKey=latestFinalizedSlot"`

	IssuerID AccountID `serix:"5,mapKey=issuerId"`
}

func (b *BlockHeader) Hash(api API) (Identifier, error) {
	headerBytes, err := api.Encode(b)
	if err != nil {
		return Identifier{}, ierrors.Errorf("failed to serialize block header: %w", err)
	}

	return blake2b.Sum256(headerBytes), nil
}

func (b *BlockHeader) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
	// Version, NetworkID, IssuingTime, SlotCommitmentID, LatestFinalisedSlot and IssuerID
	return workScoreStructure.DataByte.Multiply(serializer.OneByte +
		serializer.UInt64ByteSize +
		serializer.UInt64ByteSize +
		CommitmentIDLength +
		SlotIndexLength +
		AccountIDLength)
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

func (b *ProtocolBlock) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
	workScoreHeader, err := b.BlockHeader.WorkScore(workScoreStructure)
	if err != nil {
		return 0, err
	}

	workScoreBlock, err := b.Block.WorkScore(workScoreStructure)
	if err != nil {
		return 0, err
	}

	workScoreSignature, err := b.Signature.WorkScore(workScoreStructure)
	if err != nil {
		return 0, err
	}

	return workScoreHeader.Add(workScoreBlock, workScoreSignature)
}

// syntacticallyValidate syntactically validates the ProtocolBlock.
func (b *ProtocolBlock) syntacticallyValidate(api API) error {
	if api.ProtocolParameters().Version() != b.ProtocolVersion {
		return ierrors.Wrapf(ErrInvalidBlockVersion, "mismatched protocol version: wanted %d, got %d in block", api.ProtocolParameters().Version(), b.ProtocolVersion)
	}

	block := b.Block
	if len(block.WeakParentIDs()) > 0 {
		// weak parents must be disjunct to the rest of the parents
		nonWeakParents := lo.KeyOnlyBy(append(block.StrongParentIDs(), block.ShallowLikeParentIDs()...), func(v BlockID) BlockID {
			return v
		})

		for _, parent := range block.WeakParentIDs() {
			if _, contains := nonWeakParents[parent]; contains {
				return ierrors.Wrapf(ErrWeakParentsInvalid, "weak parents (%s) cannot have common elements with strong parents (%s) or shallow likes (%s)", block.WeakParentIDs(), block.StrongParentIDs(), block.ShallowLikeParentIDs())
			}
		}
	}

	minCommittableAge := api.ProtocolParameters().MinCommittableAge()
	maxCommittableAge := api.ProtocolParameters().MaxCommittableAge()
	commitmentIndex := b.SlotCommitmentID.Index()
	blockID, err := b.ID(api)
	if err != nil {
		return ierrors.Wrapf(err, "failed to syntactically validate block")
	}
	blockIndex := blockID.Index()

	// check that commitment is not too recent.
	if commitmentIndex > 0 && // Don't filter commitments to genesis based on being too recent.
		blockIndex < commitmentIndex+minCommittableAge {
		return ierrors.Wrapf(ErrCommitmentTooRecent, "block at slot %d committing to slot %d", blockIndex, b.SlotCommitmentID.Index())
	}

	// Check that commitment is not too old.
	if blockIndex > commitmentIndex+maxCommittableAge {
		return ierrors.Wrapf(ErrCommitmentTooOld, "block at slot %d committing to slot %d, max committable age %d", blockIndex, b.SlotCommitmentID.Index(), maxCommittableAge)
	}

	return b.Block.syntacticallyValidate(api, b)
}

type Block interface {
	Type() BlockType

	StrongParentIDs() BlockIDs
	WeakParentIDs() BlockIDs
	ShallowLikeParentIDs() BlockIDs

	Hash(api API) (Identifier, error)

	syntacticallyValidate(api API, protocolBlock *ProtocolBlock) error

	ProcessableObject
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

func (b *BasicBlock) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
	// BlockType + BurnedMana
	workScoreBytes, err := workScoreStructure.DataByte.Multiply(serializer.OneByte + ManaSize)
	if err != nil {
		return 0, err
	}

	// work score for parents is a penalty for each missing strong parent below MinStrongParentsThreshold
	var workScoreMissingParents WorkScore
	if len(b.StrongParents) < int(workScoreStructure.MinStrongParentsThreshold) {
		var err error
		workScoreMissingParents, err = workScoreStructure.MissingParent.Multiply(int(workScoreStructure.MinStrongParentsThreshold) - len(b.StrongParents))
		if err != nil {
			return 0, err
		}
	}

	workScorePayload, err := b.Payload.WorkScore(workScoreStructure)
	if err != nil {
		return 0, err
	}

	// data bytes, plus missing parents, plus payload, plus block offset.
	return workScoreBytes.Add(workScoreMissingParents, workScorePayload, workScoreStructure.Block)
}

func (b *BasicBlock) ManaCost(rmc Mana, workScoreStructure *WorkScoreStructure) (Mana, error) {
	workScore, err := b.WorkScore(workScoreStructure)
	if err != nil {
		return 0, err
	}

	return Mana(workScore) * rmc, nil
}

// syntacticallyValidate syntactically validates the BasicBlock.
func (b *BasicBlock) syntacticallyValidate(api API, protocolBlock *ProtocolBlock) error {
	if b.Payload != nil && b.Payload.PayloadType() == PayloadTransaction {
		blockID, err := protocolBlock.ID(api)
		if err != nil {
			// TODO: wrap error
			return err
		}
		blockIndex := blockID.Index()

		minCommittableAge := api.ProtocolParameters().MinCommittableAge()
		maxCommittableAge := api.ProtocolParameters().MaxCommittableAge()

		tx, _ := b.Payload.(*Transaction)
		if cInput := tx.CommitmentInput(); cInput != nil {
			cInputIndex := cInput.CommitmentID.Index()
			// check that commitment input is not too recent.
			if cInputIndex > 0 && // Don't filter commitments to genesis based on being too recent.
				blockIndex < cInputIndex+minCommittableAge { // filter commitments to future slots.
				return ierrors.Wrapf(ErrCommitmentInputTooRecent, "block at slot %d with commitment input to slot %d", blockIndex, cInput.CommitmentID.Index())
			}
			// Check that commitment input is not too old.
			if blockIndex > cInputIndex+maxCommittableAge {
				return ierrors.Wrapf(ErrCommitmentInputTooOld, "block at slot %d committing to slot %d, max committable age %d", blockIndex, cInput.CommitmentID.Index(), maxCommittableAge)
			}

			if cInputIndex > protocolBlock.SlotCommitmentID.Index() {
				return ierrors.Wrapf(ErrCommitmentInputNewerThanCommitment, "transaction in a block contains CommitmentInput to slot %d while max allowed is %d", cInput.CommitmentID.Index(), protocolBlock.SlotCommitmentID.Index())
			}

		}
	}

	return nil
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

func (b *ValidationBlock) WorkScore(_ *WorkScoreStructure) (WorkScore, error) {
	// Validator blocks do not incur any work score as they do not burn mana
	return 0, nil
}

// syntacticallyValidate syntactically validates the ValidationBlock.
func (b *ValidationBlock) syntacticallyValidate(_ API, protocolBlock *ProtocolBlock) error {
	if b.HighestSupportedVersion < protocolBlock.ProtocolVersion {
		return ierrors.Errorf("highest supported version %d must be greater equal protocol version %d", b.HighestSupportedVersion, protocolBlock.ProtocolVersion)
	}

	return nil
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
