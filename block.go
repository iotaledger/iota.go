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
	// MaxBlockSize defines the maximum size of a block in bytes.
	MaxBlockSize = 32768
	// BasicBlockMaxParents defines the maximum number of parents in a basic block.
	BasicBlockMaxParents = 8
	// ValidationBlockMaxParents defines the maximum number of parents in a ValidationBlock.
	ValidationBlockMaxParents = 50

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
	ErrBlockMaxSizeExceeded               = ierrors.New("block exceeds the max size")
	ErrCommitmentInputNewerThanCommitment = ierrors.New("a block cannot contain a commitment input with index newer than the commitment index")
	ErrBlockNetworkIDInvalid              = ierrors.New("invalid network ID in block header")
	ErrHighestSupportedVersionTooSmall    = ierrors.New("highest supported version must be greater equal the block's protocol version")
)

// BlockBodyType denotes a type of Block Body.
type BlockBodyType byte

const (
	BlockBodyTypeBasic      BlockBodyType = 0
	BlockBodyTypeValidation BlockBodyType = 1
)

type ApplicationPayload interface {
	Payload
}

// version + networkID + time + commitmentID + slot + accountID.
const BlockHeaderLength = serializer.OneByte + serializer.UInt64ByteSize + serializer.UInt64ByteSize + CommitmentIDLength + SlotIndexLength + AccountIDLength

type BlockHeader struct {
	ProtocolVersion Version   `serix:""`
	NetworkID       NetworkID `serix:""`

	IssuingTime         time.Time    `serix:""`
	SlotCommitmentID    CommitmentID `serix:""`
	LatestFinalizedSlot SlotIndex    `serix:""`

	IssuerID AccountID `serix:""`
}

func (b *BlockHeader) Hash(api API) (Identifier, error) {
	headerBytes, err := api.Encode(b)
	if err != nil {
		return Identifier{}, ierrors.Wrap(err, "failed to serialize block header")
	}

	return blake2b.Sum256(headerBytes), nil
}

func (b *BlockHeader) Size() int {
	return BlockHeaderLength
}

type Block struct {
	API       API
	Header    BlockHeader `serix:""`
	Body      BlockBody   `serix:""`
	Signature Signature   `serix:""`
}

func BlockFromBytes(apiProvider APIProvider) func(bytes []byte) (block *Block, consumedBytes int, err error) {
	return func(bytes []byte) (block *Block, consumedBytes int, err error) {
		block = new(Block)

		var version Version
		if version, consumedBytes, err = VersionFromBytes(bytes); err != nil {
			err = ierrors.Wrap(err, "failed to parse version")
		} else if block.API, err = apiProvider.APIForVersion(version); err != nil {
			err = ierrors.Wrapf(err, "failed to retrieve API for version %d", version)
		} else if consumedBytes, err = block.API.Decode(bytes, block, serix.WithValidation()); err != nil {
			err = ierrors.Wrap(err, "failed to deserialize Block")
		}

		return block, consumedBytes, err
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

func (b *Block) SetDeserializationContext(ctx context.Context) {
	b.API = APIFromContext(ctx)
}

// SigningMessage returns the to be signed message.
// The BlockHeader and Block are separately hashed and concatenated to enable the verification of the signature for
// an Attestation where only the BlockHeader and the hash of Block is known.
func (b *Block) SigningMessage() ([]byte, error) {
	headerHash, err := b.Header.Hash(b.API)
	if err != nil {
		return nil, err
	}

	blockHash, err := b.Body.Hash()
	if err != nil {
		return nil, err
	}

	return blockSigningMessage(headerHash, blockHash), nil
}

func blockSigningMessage(headerHash Identifier, blockHash Identifier) []byte {
	return byteutils.ConcatBytes(headerHash[:], blockHash[:])
}

// Sign produces a signature by signing the signing message of the block.
func (b *Block) Sign(signer AddressSigner, addr Address) (Signature, error) {
	signMsg, err := b.SigningMessage()
	if err != nil {
		return nil, err
	}

	return signer.Sign(addr, signMsg)
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
		return false, ierrors.New("ed25519 public key must not be empty")
	}

	return hiveEd25519.Verify(edSig.PublicKey[:], signingMessage, edSig.Signature[:]), nil
}

// Slot returns the SlotIndex of the Block.
func (b *Block) Slot() SlotIndex {
	return b.API.TimeProvider().SlotFromTime(b.Header.IssuingTime)
}

// ID computes the ID of the Block.
func (b *Block) ID() (BlockID, error) {
	data, err := b.API.Encode(b)
	if err != nil {
		return BlockID{}, ierrors.Wrap(err, "failed to compute blockID")
	}

	id, err := BlockIdentifierFromBlockBytes(data)
	if err != nil {
		return BlockID{}, ierrors.Wrap(err, "failed to compute blockID")
	}

	return b.IDWithBlockIdentifier(id), nil
}

func (b *Block) IDWithBlockIdentifier(blockIdentifier Identifier) BlockID {
	return NewBlockID(b.Slot(), blockIdentifier)
}

// MustID works like ID but panics if the BlockID can't be computed.
func (b *Block) MustID() BlockID {
	blockID, err := b.ID()
	if err != nil {
		panic(err)
	}

	return blockID
}

func (b *Block) Parents() (parents []BlockID) {
	parents = make([]BlockID, 0)

	parents = append(parents, b.Body.StrongParentIDs()...)
	parents = append(parents, b.Body.WeakParentIDs()...)
	parents = append(parents, b.Body.ShallowLikeParentIDs()...)

	return parents
}

func (b *Block) ParentsWithType() (parents []Parent) {
	parents = make([]Parent, 0)

	for _, parentBlockID := range b.Body.StrongParentIDs() {
		parents = append(parents, Parent{parentBlockID, StrongParentType})
	}

	for _, parentBlockID := range b.Body.WeakParentIDs() {
		parents = append(parents, Parent{parentBlockID, WeakParentType})
	}

	for _, parentBlockID := range b.Body.ShallowLikeParentIDs() {
		parents = append(parents, Parent{parentBlockID, ShallowLikeParentType})
	}

	return parents
}

// ForEachParent executes a consumer func for each parent.
func (b *Block) ForEachParent(consumer func(parent Parent)) {
	for _, parent := range b.ParentsWithType() {
		consumer(parent)
	}
}

func (b *Block) WorkScore() (WorkScore, error) {
	workScoreParameters := b.API.ProtocolParameters().WorkScoreParameters()

	// the workscore of the block only consists of the workscore of the block body
	// because the body should already include an offset for the "block",
	// which accounts for the signature check of the block as well.
	return b.Body.WorkScore(workScoreParameters)
}

// Size returns the size of the block in bytes.
func (b *Block) Size() int {
	return b.Header.Size() + b.Body.Size() + b.Signature.Size()
}

// ManaCost returns the cost of the block in Mana, which is the given rmc multiplied by the block's workscore.
func (b *Block) ManaCost(rmc Mana) (Mana, error) {
	workScore, err := b.WorkScore()
	if err != nil {
		return 0, ierrors.Wrap(err, "failed to calculate block workscore")
	}

	manaCost, err := ManaCost(rmc, workScore)
	if err != nil {
		return 0, ierrors.Wrap(err, "failed to calculate block mana cost")
	}

	return manaCost, nil
}

// syntacticallyValidate syntactically validates the Block.
func (b *Block) syntacticallyValidate() error {
	if b.Size() > MaxBlockSize {
		return ierrors.WithMessagef(ErrBlockMaxSizeExceeded, "max size of a block is %d but got %d bytes", MaxBlockSize, b.Size())
	}

	if b.API.ProtocolParameters().Version() != b.Header.ProtocolVersion {
		return ierrors.WithMessagef(ErrInvalidBlockVersion, "mismatched protocol version: expected %d, got %d in block", b.API.ProtocolParameters().Version(), b.Header.ProtocolVersion)
	}

	expectedNetworkID := b.API.ProtocolParameters().NetworkID()
	if b.Header.NetworkID != expectedNetworkID {
		return ierrors.WithMessagef(ErrBlockNetworkIDInvalid, "expected %d (%s), got %d", expectedNetworkID, b.API.ProtocolParameters().NetworkName(), b.Header.NetworkID)
	}

	block := b.Body
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

	protocolParams := b.API.ProtocolParameters()
	genesisSlot := protocolParams.GenesisSlot()
	minCommittableAge := protocolParams.MinCommittableAge()
	maxCommittableAge := protocolParams.MaxCommittableAge()
	commitmentSlot := b.Header.SlotCommitmentID.Slot()
	blockSlot := b.Slot()

	// check that commitment is not too recent.
	if commitmentSlot > genesisSlot && // Don't filter commitments to genesis based on being too recent.
		blockSlot < commitmentSlot+minCommittableAge {
		return ierrors.WithMessagef(ErrCommitmentTooRecent, "block at slot %d committing to slot %d, min committable age %d", blockSlot, b.Header.SlotCommitmentID.Slot(), minCommittableAge)
	}

	// Check that commitment is not too old.
	if blockSlot > commitmentSlot+maxCommittableAge {
		return ierrors.WithMessagef(ErrCommitmentTooOld, "block at slot %d committing to slot %d, max committable age %d", blockSlot, b.Header.SlotCommitmentID.Slot(), maxCommittableAge)
	}

	return b.Body.syntacticallyValidate(b)
}

type BlockBody interface {
	Type() BlockBodyType

	StrongParentIDs() BlockIDs
	WeakParentIDs() BlockIDs
	ShallowLikeParentIDs() BlockIDs

	Hash() (Identifier, error)

	syntacticallyValidate(block *Block) error

	ProcessableObject
	Sizer
}

// BasicBlockBody represents a basic vertex in the Tangle/BlockDAG.
type BasicBlockBody struct {
	API API

	// The parents the block references.
	StrongParents      BlockIDs `serix:",lenPrefix=uint8,minLen=1,maxLen=8"`
	WeakParents        BlockIDs `serix:",lenPrefix=uint8,minLen=0,maxLen=8,omitempty"`
	ShallowLikeParents BlockIDs `serix:",lenPrefix=uint8,minLen=0,maxLen=8,omitempty"`

	// The inner payload of the block. Can be nil.
	Payload ApplicationPayload `serix:",optional,omitempty"`

	MaxBurnedMana Mana `serix:""`
}

func (b *BasicBlockBody) SetDeserializationContext(ctx context.Context) {
	b.API = APIFromContext(ctx)
}

func (b *BasicBlockBody) Type() BlockBodyType {
	return BlockBodyTypeBasic
}

func (b *BasicBlockBody) StrongParentIDs() BlockIDs {
	return b.StrongParents
}

func (b *BasicBlockBody) WeakParentIDs() BlockIDs {
	return b.WeakParents
}

func (b *BasicBlockBody) ShallowLikeParentIDs() BlockIDs {
	return b.ShallowLikeParents
}

func (b *BasicBlockBody) Hash() (Identifier, error) {
	blockBytes, err := b.API.Encode(b)
	if err != nil {
		return Identifier{}, ierrors.Wrap(err, "failed to serialize basic block")
	}

	return blake2b.Sum256(blockBytes), nil
}

func (b *BasicBlockBody) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	if b.Payload != nil {
		// offset for the block is included in the payload WorkScore.
		return b.Payload.WorkScore(workScoreParameters)
	}

	// if the payload is nil, just return the offset for block.
	return workScoreParameters.Block, nil
}

func (b *BasicBlockBody) Size() int {
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
func (b *BasicBlockBody) syntacticallyValidate(block *Block) error {
	if b.Payload != nil && b.Payload.PayloadType() == PayloadSignedTransaction {
		blockSlot := block.Slot()

		minCommittableAge := block.API.ProtocolParameters().MinCommittableAge()
		maxCommittableAge := block.API.ProtocolParameters().MaxCommittableAge()

		//nolint:forcetypeassert // we can safely assume that this is a SignedTransaction
		signedTransaction := b.Payload.(*SignedTransaction)

		// check that transaction CreationSlot is smaller or equal than the block that contains it
		if blockSlot < signedTransaction.Transaction.CreationSlot {
			return ierrors.WithMessagef(ErrTransactionCreationSlotTooRecent, "transaction creation slot %d exceeds block slot %d", signedTransaction.Transaction.CreationSlot, blockSlot)
		}

		if cInput := signedTransaction.Transaction.CommitmentInput(); cInput != nil {
			cInputSlot := cInput.CommitmentID.Slot()
			// check that commitment input is not too recent.
			if cInputSlot > 0 && // Don't filter commitments to genesis based on being too recent.
				blockSlot < cInputSlot+minCommittableAge { // filter commitments to future slots.
				return ierrors.WithMessagef(ErrCommitmentInputTooRecent, "block at slot %d with commitment input to slot %d, min committable age %d", blockSlot, cInput.CommitmentID.Slot(), minCommittableAge)
			}
			// Check that commitment input is not too old.
			if blockSlot > cInputSlot+maxCommittableAge {
				return ierrors.WithMessagef(ErrCommitmentInputTooOld, "block at slot %d with commitment input to slot %d, max committable age %d", blockSlot, cInput.CommitmentID.Slot(), maxCommittableAge)
			}

			if cInputSlot > block.Header.SlotCommitmentID.Slot() {
				return ierrors.WithMessagef(ErrCommitmentInputNewerThanCommitment, "transaction in a block contains CommitmentInput to slot %d while max allowed is %d", cInput.CommitmentID.Slot(), block.Header.SlotCommitmentID.Slot())
			}
		}
	}

	return nil
}

// ValidationBlockBody represents a validation vertex in the Tangle/BlockDAG.
type ValidationBlockBody struct {
	API API
	// The parents the block references.
	StrongParents      BlockIDs `serix:",lenPrefix=uint8,minLen=1,maxLen=50"`
	WeakParents        BlockIDs `serix:",lenPrefix=uint8,minLen=0,maxLen=50,omitempty"`
	ShallowLikeParents BlockIDs `serix:",lenPrefix=uint8,minLen=0,maxLen=50,omitempty"`

	HighestSupportedVersion Version `serix:""`
	// ProtocolParametersHash is the hash of the protocol parameters for the HighestSupportedVersion.
	ProtocolParametersHash Identifier `serix:""`
}

func (b *ValidationBlockBody) SetDeserializationContext(ctx context.Context) {
	b.API = APIFromContext(ctx)
}

func (b *ValidationBlockBody) Type() BlockBodyType {
	return BlockBodyTypeValidation
}

func (b *ValidationBlockBody) StrongParentIDs() BlockIDs {
	return b.StrongParents
}

func (b *ValidationBlockBody) WeakParentIDs() BlockIDs {
	return b.WeakParents
}

func (b *ValidationBlockBody) ShallowLikeParentIDs() BlockIDs {
	return b.ShallowLikeParents
}

func (b *ValidationBlockBody) Hash() (Identifier, error) {
	blockBytes, err := b.API.Encode(b)
	if err != nil {
		return Identifier{}, ierrors.Wrap(err, "failed to serialize validation block")
	}

	return IdentifierFromData(blockBytes), nil
}

func (b *ValidationBlockBody) WorkScore(_ *WorkScoreParameters) (WorkScore, error) {
	// Validator blocks do not incur any work score as they do not burn mana
	return 0, nil
}

func (b *ValidationBlockBody) Size() int {
	return serializer.OneByte + // block type
		serializer.OneByte + len(b.StrongParents)*BlockIDLength + // StrongParents count
		serializer.OneByte + len(b.WeakParents)*BlockIDLength + // WeakParents count
		serializer.OneByte + len(b.ShallowLikeParents)*BlockIDLength + // ShallowLikeParents count
		serializer.OneByte + // highest supported version
		IdentifierLength // protocol parameters hash
}

// syntacticallyValidate syntactically validates the ValidationBlock.
func (b *ValidationBlockBody) syntacticallyValidate(block *Block) error {
	if b.HighestSupportedVersion < block.Header.ProtocolVersion {
		return ierrors.WithMessagef(ErrHighestSupportedVersionTooSmall, "highest supported version %d, block header protocol version %d", b.HighestSupportedVersion, block.Header.ProtocolVersion)
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
