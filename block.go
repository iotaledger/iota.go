package iotago

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"time"

	"golang.org/x/crypto/blake2b"

	hiveEd25519 "github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
)

const (
	// MaxBlockSize defines the maximum size of a block.
	MaxBlockSize = 32768
	// BlockMaxParents defines the maximum amount of parents in a block.
	BlockMaxParents = 8
	// BlockTypeValidationMaxParents defines the maximum amount of parents in a ValidationBlock. TODO: replace number with committee size.
	BlockTypeValidationMaxParents = BlockMaxParents + 42

	// block type + strong parents count + weak parents count + shallow like parents count + payload type + mana.
	BasicBlockSizeEmptyParentsAndEmptyPayload = serializer.OneByte + serializer.OneByte + serializer.OneByte + serializer.OneByte + serializer.TypeDenotationByteSize + ManaSize
)

var (
	ErrWeakParentsInvalid                 = ierrors.New("weak parents must be disjunct to the rest of the parents")
	ErrTransactionCreationSlotTooRecent   = ierrors.New("a block cannot contain a transaction with creation slot more recent than the block's issuing time")
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
	BlockTypeBasic      BlockType = 0
	BlockTypeValidation BlockType = 1
)

type BlockPayload interface {
	Payload
}

// version + networkID + time + commitmentID + slot + accountID.
const BlockHeaderLength = serializer.OneByte + serializer.UInt64ByteSize + serializer.UInt64ByteSize + CommitmentIDLength + SlotIndexLength + AccountIDLength

type BlockHeader struct {
	ProtocolVersion Version   `serix:"0,mapKey=protocolVersion"`
	NetworkID       NetworkID `serix:"1,mapKey=networkId"`

	IssuingTime         time.Time    `serix:"2,mapKey=issuingTime"`
	SlotCommitmentID    CommitmentID `serix:"3,mapKey=slotCommitmentId"`
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

func (b *BlockHeader) WorkScore(_ *WorkScoreParameters) (WorkScore, error) {
	return 0, nil
}

func (b *BlockHeader) Size() int {
	return BlockHeaderLength
}

type ProtocolBlock struct {
	API         API
	BlockHeader `serix:"0"`
	Block       Block     `serix:"1,mapKey=block"`
	Signature   Signature `serix:"2,mapKey=signature"`
}

func ProtocolBlockFromBytes(apiProvider APIProvider) func(bytes []byte) (protocolBlock *ProtocolBlock, consumedBytes int, err error) {
	return func(bytes []byte) (protocolBlock *ProtocolBlock, consumedBytes int, err error) {
		protocolBlock = new(ProtocolBlock)

		var version Version
		if version, consumedBytes, err = VersionFromBytes(bytes); err != nil {
			err = ierrors.Wrap(err, "failed to parse version")
		} else if protocolBlock.API, err = apiProvider.APIForVersion(version); err != nil {
			err = ierrors.Wrapf(err, "failed to retrieve API for version %d", version)
		} else if consumedBytes, err = protocolBlock.API.Decode(bytes, protocolBlock, serix.WithValidation()); err != nil {
			err = ierrors.Wrap(err, "failed to deserialize ProtocolBlock")
		}

		return protocolBlock, consumedBytes, err
	}
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

func (b *ProtocolBlock) SetDeserializationContext(ctx context.Context) {
	b.API = APIFromContext(ctx)
}

// SigningMessage returns the to be signed message.
// The BlockHeader and Block are separately hashed and concatenated to enable the verification of the signature for
// an Attestation where only the BlockHeader and the hash of Block is known.
func (b *ProtocolBlock) SigningMessage() ([]byte, error) {
	headerHash, err := b.BlockHeader.Hash(b.API)
	if err != nil {
		return nil, err
	}

	blockHash, err := b.Block.Hash()
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
func (b *ProtocolBlock) Sign(addrKey AddressKeys) (Signature, error) {
	signMsg, err := b.SigningMessage()
	if err != nil {
		return nil, err
	}

	signer := NewInMemoryAddressSigner(addrKey)

	return signer.Sign(addrKey.Address, signMsg)
}

// VerifySignature verifies the Signature of the block.
func (b *ProtocolBlock) VerifySignature() (valid bool, err error) {
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
func (b *ProtocolBlock) ID() (BlockID, error) {
	data, err := b.API.Encode(b)
	if err != nil {
		return BlockID{}, ierrors.Errorf("can't compute block ID: %w", err)
	}

	id, err := BlockIdentifierFromBlockBytes(data)
	if err != nil {
		return BlockID{}, err
	}

	slot := b.API.TimeProvider().SlotFromTime(b.IssuingTime)

	return NewBlockID(slot, id), nil
}

// MustID works like ID but panics if the BlockID can't be computed.
func (b *ProtocolBlock) MustID() BlockID {
	blockID, err := b.ID()
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

func (b *ProtocolBlock) WorkScore() (WorkScore, error) {
	workScoreParameters := b.API.ProtocolParameters().WorkScoreParameters()

	workScoreHeader, err := b.BlockHeader.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	workScoreBlock, err := b.Block.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	workScoreSignature, err := b.Signature.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	return workScoreHeader.Add(workScoreHeader, workScoreBlock, workScoreSignature)
}

// Size returns the size of the block in bytes.
func (b *ProtocolBlock) Size() int {
	return b.BlockHeader.Size() + b.Block.Size() + b.Signature.Size()
}

// syntacticallyValidate syntactically validates the ProtocolBlock.
func (b *ProtocolBlock) syntacticallyValidate() error {
	if b.API.ProtocolParameters().Version() != b.ProtocolVersion {
		return ierrors.Wrapf(ErrInvalidBlockVersion, "mismatched protocol version: wanted %d, got %d in block", b.API.ProtocolParameters().Version(), b.ProtocolVersion)
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

	minCommittableAge := b.API.ProtocolParameters().MinCommittableAge()
	maxCommittableAge := b.API.ProtocolParameters().MaxCommittableAge()
	commitmentSlot := b.SlotCommitmentID.Slot()
	blockID, err := b.ID()
	if err != nil {
		return ierrors.Wrapf(err, "failed to syntactically validate block")
	}
	blockSlot := blockID.Slot()

	// check that commitment is not too recent.
	if commitmentSlot > 0 && // Don't filter commitments to genesis based on being too recent.
		blockSlot < commitmentSlot+minCommittableAge {
		return ierrors.Wrapf(ErrCommitmentTooRecent, "block at slot %d committing to slot %d", blockSlot, b.SlotCommitmentID.Slot())
	}

	// Check that commitment is not too old.
	if blockSlot > commitmentSlot+maxCommittableAge {
		return ierrors.Wrapf(ErrCommitmentTooOld, "block at slot %d committing to slot %d, max committable age %d", blockSlot, b.SlotCommitmentID.Slot(), maxCommittableAge)
	}

	return b.Block.syntacticallyValidate(b)
}

type Block interface {
	Type() BlockType

	StrongParentIDs() BlockIDs
	WeakParentIDs() BlockIDs
	ShallowLikeParentIDs() BlockIDs

	Hash() (Identifier, error)

	syntacticallyValidate(protocolBlock *ProtocolBlock) error

	ProcessableObject
	Sizer
}

// BasicBlock represents a basic vertex in the Tangle/BlockDAG.
type BasicBlock struct {
	API API

	// The parents the block references.
	StrongParents      BlockIDs `serix:"0,lengthPrefixType=uint8,mapKey=strongParents,minLen=1,maxLen=8"`
	WeakParents        BlockIDs `serix:"1,lengthPrefixType=uint8,mapKey=weakParents,minLen=0,maxLen=8"`
	ShallowLikeParents BlockIDs `serix:"2,lengthPrefixType=uint8,mapKey=shallowLikeParents,minLen=0,maxLen=8"`

	// The inner payload of the block. Can be nil.
	Payload BlockPayload `serix:"3,optional,mapKey=payload,omitempty"`

	MaxBurnedMana Mana `serix:"4,mapKey=maxBurnedMana"`
}

func (b *BasicBlock) SetDeserializationContext(ctx context.Context) {
	b.API = APIFromContext(ctx)
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

func (b *BasicBlock) Hash() (Identifier, error) {
	blockBytes, err := b.API.Encode(b)
	if err != nil {
		return Identifier{}, ierrors.Errorf("failed to serialize basic block: %w", err)
	}

	return blake2b.Sum256(blockBytes), nil
}

func (b *BasicBlock) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	var err error
	var workScorePayload WorkScore
	if b.Payload != nil {
		workScorePayload, err = b.Payload.WorkScore(workScoreParameters)
		if err != nil {
			return 0, err
		}
	}

	// offset for block plus payload.
	return workScoreParameters.Block.Add(workScorePayload)
}

func (b *BasicBlock) ManaCost(rmc Mana, workScoreParameters *WorkScoreParameters) (Mana, error) {
	workScore, err := b.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	return ManaCost(rmc, workScore)
}

func (b *BasicBlock) Size() int {
	var payloadSize int
	if b.Payload != nil {
		payloadSize = b.Payload.Size()
	}

	return BasicBlockSizeEmptyParentsAndEmptyPayload +
		len(b.StrongParents)*BlockIDLength +
		len(b.WeakParents)*BlockIDLength +
		len(b.ShallowLikeParents)*BlockIDLength +
		payloadSize
}

// syntacticallyValidate syntactically validates the BasicBlock.
func (b *BasicBlock) syntacticallyValidate(protocolBlock *ProtocolBlock) error {
	if b.Payload != nil && b.Payload.PayloadType() == PayloadSignedTransaction {
		blockID, err := protocolBlock.ID()
		if err != nil {
			return ierrors.Wrap(err, "error while calculating block ID during syntactical validation")
		}
		blockSlot := blockID.Slot()

		minCommittableAge := protocolBlock.API.ProtocolParameters().MinCommittableAge()
		maxCommittableAge := protocolBlock.API.ProtocolParameters().MaxCommittableAge()

		signedTransaction, _ := b.Payload.(*SignedTransaction)

		// check that transaction CreationSlot is smaller or equal than the block that contains it
		if blockSlot < signedTransaction.Transaction.CreationSlot {
			return ierrors.Wrapf(ErrTransactionCreationSlotTooRecent, "block at slot %d with commitment input to slot %d", blockSlot, signedTransaction.Transaction.CreationSlot)
		}

		if cInput := signedTransaction.Transaction.CommitmentInput(); cInput != nil {
			cInputSlot := cInput.CommitmentID.Slot()
			// check that commitment input is not too recent.
			if cInputSlot > 0 && // Don't filter commitments to genesis based on being too recent.
				blockSlot < cInputSlot+minCommittableAge { // filter commitments to future slots.
				return ierrors.Wrapf(ErrCommitmentInputTooRecent, "block at slot %d with commitment input to slot %d", blockSlot, cInput.CommitmentID.Slot())
			}
			// Check that commitment input is not too old.
			if blockSlot > cInputSlot+maxCommittableAge {
				return ierrors.Wrapf(ErrCommitmentInputTooOld, "block at slot %d committing to slot %d, max committable age %d", blockSlot, cInput.CommitmentID.Slot(), maxCommittableAge)
			}

			if cInputSlot > protocolBlock.SlotCommitmentID.Slot() {
				return ierrors.Wrapf(ErrCommitmentInputNewerThanCommitment, "transaction in a block contains CommitmentInput to slot %d while max allowed is %d", cInput.CommitmentID.Slot(), protocolBlock.SlotCommitmentID.Slot())
			}
		}
	}

	return nil
}

// ValidationBlock represents a validation vertex in the Tangle/BlockDAG.
type ValidationBlock struct {
	API API
	// The parents the block references.
	StrongParents      BlockIDs `serix:"0,lengthPrefixType=uint8,mapKey=strongParents,minLen=1,maxLen=50"`
	WeakParents        BlockIDs `serix:"1,lengthPrefixType=uint8,mapKey=weakParents,minLen=0,maxLen=50"`
	ShallowLikeParents BlockIDs `serix:"2,lengthPrefixType=uint8,mapKey=shallowLikeParents,minLen=0,maxLen=50"`

	HighestSupportedVersion Version `serix:"3,mapKey=highestSupportedVersion"`
	// ProtocolParametersHash is the hash of the protocol parameters for the HighestSupportedVersion.
	ProtocolParametersHash Identifier `serix:"4,mapKey=protocolParametersHash"`
}

func (b *ValidationBlock) SetDeserializationContext(ctx context.Context) {
	b.API = APIFromContext(ctx)
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

func (b *ValidationBlock) Hash() (Identifier, error) {
	blockBytes, err := b.API.Encode(b)
	if err != nil {
		return Identifier{}, ierrors.Errorf("failed to serialize validation block: %w", err)
	}

	return IdentifierFromData(blockBytes), nil
}

func (b *ValidationBlock) WorkScore(_ *WorkScoreParameters) (WorkScore, error) {
	// Validator blocks do not incur any work score as they do not burn mana
	return 0, nil
}

func (b *ValidationBlock) Size() int {
	return serializer.OneByte + // block type
		serializer.OneByte + len(b.StrongParents)*BlockIDLength + // StrongParents count
		serializer.OneByte + len(b.WeakParents)*BlockIDLength + // WeakParents count
		serializer.OneByte + len(b.ShallowLikeParents)*BlockIDLength + // ShallowLikeParents count
		serializer.OneByte + // highest supported version
		IdentifierLength // protocol parameters hash
}

// syntacticallyValidate syntactically validates the ValidationBlock.
func (b *ValidationBlock) syntacticallyValidate(protocolBlock *ProtocolBlock) error {
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
