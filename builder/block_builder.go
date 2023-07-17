package builder

import (
	"crypto/ed25519"
	"time"

	"github.com/iotaledger/hive.go/ierrors"
	iotago "github.com/iotaledger/iota.go/v4"
)

// NewBasicBlockBuilder creates a new BasicBlockBuilder.
func NewBasicBlockBuilder(api iotago.API) *BasicBlockBuilder {
	basicBlock := &iotago.BasicBlock{}

	protocolBlock := &iotago.ProtocolBlock{
		BlockHeader: iotago.BlockHeader{
			ProtocolVersion:  api.ProtocolParameters().Version(),
			SlotCommitmentID: iotago.EmptyCommitmentID,
		},
		Signature: &iotago.Ed25519Signature{},
		Block:     basicBlock,
	}

	return &BasicBlockBuilder{
		api:           api,
		protocolBlock: protocolBlock,
		basicBlock:    basicBlock,
	}
}

// BasicBlockBuilder is used to easily build up a Basic Block.
type BasicBlockBuilder struct {
	api iotago.API

	basicBlock *iotago.BasicBlock

	protocolBlock *iotago.ProtocolBlock
	err           error
}

// Build builds the ProtocolBlock or returns any error which occurred during the build steps.
func (b *BasicBlockBuilder) Build() (*iotago.ProtocolBlock, error) {
	if b.err != nil {
		return nil, b.err
	}

	return b.protocolBlock, nil
}

// ProtocolVersion sets the protocol version.
func (b *BasicBlockBuilder) ProtocolVersion(version iotago.Version) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.protocolBlock.ProtocolVersion = version

	return b
}

func (b *BasicBlockBuilder) IssuingTime(time time.Time) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.protocolBlock.IssuingTime = time

	return b
}

// SlotCommitmentID sets the slot commitment.
func (b *BasicBlockBuilder) SlotCommitmentID(commitment iotago.CommitmentID) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.protocolBlock.SlotCommitmentID = commitment

	return b
}

// LatestFinalizedSlot sets the latest finalized slot.
func (b *BasicBlockBuilder) LatestFinalizedSlot(slot iotago.SlotIndex) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.protocolBlock.LatestFinalizedSlot = slot

	return b
}

func (b *BasicBlockBuilder) Sign(accountID iotago.AccountID, prvKey ed25519.PrivateKey) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.protocolBlock.IssuerID = accountID

	signature, err := b.protocolBlock.Sign(b.api, iotago.NewAddressKeysForEd25519Address(iotago.Ed25519AddressFromPubKey(prvKey.Public().(ed25519.PublicKey)), prvKey))
	if err != nil {
		b.err = ierrors.Errorf("error signing block: %w", err)
		return b
	}

	edSig, isEdSig := signature.(*iotago.Ed25519Signature)
	if !isEdSig {
		panic("unsupported signature type")
	}

	b.protocolBlock.Signature = edSig

	return b
}

// StrongParents sets the strong parents.
func (b *BasicBlockBuilder) StrongParents(parents iotago.BlockIDs) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.basicBlock.StrongParents = parents.RemoveDupsAndSort()

	return b
}

// WeakParents sets the weak parents.
func (b *BasicBlockBuilder) WeakParents(parents iotago.BlockIDs) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.basicBlock.WeakParents = parents.RemoveDupsAndSort()

	return b
}

// ShallowLikeParents sets the shallow like parents.
func (b *BasicBlockBuilder) ShallowLikeParents(parents iotago.BlockIDs) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.basicBlock.ShallowLikeParents = parents.RemoveDupsAndSort()

	return b
}

// Payload sets the payload.
func (b *BasicBlockBuilder) Payload(payload iotago.Payload) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.basicBlock.Payload = payload

	return b
}

// BurnedMana sets the amount of mana burned by the block.
func (b *BasicBlockBuilder) BurnedMana(burnedMana iotago.Mana) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.basicBlock.BurnedMana = burnedMana

	return b
}

// NewValidationBlockBuilder creates a new ValidationBlockBuilder.
func NewValidationBlockBuilder(api iotago.API) *ValidationBlockBuilder {
	validationBlock := &iotago.ValidationBlock{}

	protocolBlock := &iotago.ProtocolBlock{
		BlockHeader: iotago.BlockHeader{
			ProtocolVersion:  api.ProtocolParameters().Version(),
			SlotCommitmentID: iotago.NewEmptyCommitment(api.ProtocolParameters().Version()).MustID(),
			IssuingTime:      time.Now(),
		},
		Signature: &iotago.Ed25519Signature{},
		Block:     validationBlock,
	}

	return &ValidationBlockBuilder{
		api:             api,
		protocolBlock:   protocolBlock,
		validationBlock: validationBlock,
	}
}

// ValidationBlockBuilder is used to easily build up a Validation Block.
type ValidationBlockBuilder struct {
	api iotago.API

	validationBlock *iotago.ValidationBlock

	protocolBlock *iotago.ProtocolBlock
	err           error
}

// Build builds the ProtocolBlock or returns any error which occurred during the build steps.
func (v *ValidationBlockBuilder) Build() (*iotago.ProtocolBlock, error) {
	if v.err != nil {
		return nil, v.err
	}

	return v.protocolBlock, nil
}

// ProtocolVersion sets the protocol version.
func (v *ValidationBlockBuilder) ProtocolVersion(version iotago.Version) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.protocolBlock.ProtocolVersion = version

	return v
}

func (v *ValidationBlockBuilder) IssuingTime(time time.Time) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.protocolBlock.IssuingTime = time

	return v
}

// SlotCommitmentID sets the slot commitment.
func (v *ValidationBlockBuilder) SlotCommitmentID(commitmentID iotago.CommitmentID) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.protocolBlock.SlotCommitmentID = commitmentID

	return v
}

// LatestFinalizedSlot sets the latest finalized slot.
func (v *ValidationBlockBuilder) LatestFinalizedSlot(slot iotago.SlotIndex) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.protocolBlock.LatestFinalizedSlot = slot

	return v
}

func (v *ValidationBlockBuilder) Sign(accountID iotago.AccountID, prvKey ed25519.PrivateKey) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.protocolBlock.IssuerID = accountID

	signature, err := v.protocolBlock.Sign(v.api, iotago.NewAddressKeysForEd25519Address(iotago.Ed25519AddressFromPubKey(prvKey.Public().(ed25519.PublicKey)), prvKey))
	if err != nil {
		v.err = ierrors.Errorf("error signing block: %w", err)
		return v
	}

	edSig, isEdSig := signature.(*iotago.Ed25519Signature)
	if !isEdSig {
		panic("unsupported signature type")
	}

	v.protocolBlock.Signature = edSig

	return v
}

// StrongParents sets the strong parents.
func (v *ValidationBlockBuilder) StrongParents(parents iotago.BlockIDs) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.validationBlock.StrongParents = parents.RemoveDupsAndSort()

	return v
}

// WeakParents sets the weak parents.
func (v *ValidationBlockBuilder) WeakParents(parents iotago.BlockIDs) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.validationBlock.WeakParents = parents.RemoveDupsAndSort()

	return v
}

// ShallowLikeParents sets the shallow like parents.
func (v *ValidationBlockBuilder) ShallowLikeParents(parents iotago.BlockIDs) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.validationBlock.ShallowLikeParents = parents.RemoveDupsAndSort()

	return v
}

// HighestSupportedVersion sets the highest supported version.
func (v *ValidationBlockBuilder) HighestSupportedVersion(highestSupportedVersion iotago.Version) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.validationBlock.HighestSupportedVersion = highestSupportedVersion

	return v
}

// ProtocolParametersHash sets the ProtocolParametersHash of the highest supported version.
func (v *ValidationBlockBuilder) ProtocolParametersHash(hash iotago.Identifier) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.validationBlock.ProtocolParametersHash = hash

	return v
}
