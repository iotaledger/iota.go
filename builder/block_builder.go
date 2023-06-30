package builder

import (
	"crypto/ed25519"
	"fmt"
	"time"

	iotago "github.com/iotaledger/iota.go/v4"
)

// NewBasicBlockBuilder creates a new BasicBlockBuilder.
func NewBasicBlockBuilder(api iotago.API) *BasicBlockBuilder {
	basicBlock := &iotago.BasicBlock{}

	protocolBlock := &iotago.ProtocolBlock{
		BlockHeader: iotago.BlockHeader{
			ProtocolVersion:  api.ProtocolParameters().Version(),
			SlotCommitmentID: iotago.NewEmptyCommitment(api.ProtocolParameters().Version()).MustID(),
			IssuingTime:      time.Now(),
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
func (b *BasicBlockBuilder) ProtocolVersion(version byte) *BasicBlockBuilder {
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
		b.err = fmt.Errorf("error signing block: %w", err)
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

// NewValidatorBlockBuilder creates a new ValidatorBlockBuilder.
func NewValidatorBlockBuilder(api iotago.API) *ValidatorBlockBuilder {
	validatorBlock := &iotago.ValidatorBlock{}

	protocolBlock := &iotago.ProtocolBlock{
		BlockHeader: iotago.BlockHeader{
			ProtocolVersion:  api.ProtocolParameters().Version(),
			SlotCommitmentID: iotago.NewEmptyCommitment(api.ProtocolParameters().Version()).MustID(),
			IssuingTime:      time.Now(),
		},
		Signature: &iotago.Ed25519Signature{},
		Block:     validatorBlock,
	}

	return &ValidatorBlockBuilder{
		api:            api,
		protocolBlock:  protocolBlock,
		validatorBlock: validatorBlock,
	}
}

// ValidatorBlockBuilder is used to easily build up a Validator Block.
type ValidatorBlockBuilder struct {
	api iotago.API

	validatorBlock *iotago.ValidatorBlock

	protocolBlock *iotago.ProtocolBlock
	err           error
}

// Build builds the ProtocolBlock or returns any error which occurred during the build steps.
func (v *ValidatorBlockBuilder) Build() (*iotago.ProtocolBlock, error) {
	if v.err != nil {
		return nil, v.err
	}

	return v.protocolBlock, nil
}

// ProtocolVersion sets the protocol version.
func (v *ValidatorBlockBuilder) ProtocolVersion(version byte) *ValidatorBlockBuilder {
	if v.err != nil {
		return v
	}

	v.protocolBlock.ProtocolVersion = version

	return v
}

func (v *ValidatorBlockBuilder) IssuingTime(time time.Time) *ValidatorBlockBuilder {
	if v.err != nil {
		return v
	}

	v.protocolBlock.IssuingTime = time

	return v
}

// SlotCommitmentID sets the slot commitment.
func (v *ValidatorBlockBuilder) SlotCommitmentID(commitmentID iotago.CommitmentID) *ValidatorBlockBuilder {
	if v.err != nil {
		return v
	}

	v.protocolBlock.SlotCommitmentID = commitmentID

	return v
}

// LatestFinalizedSlot sets the latest finalized slot.
func (v *ValidatorBlockBuilder) LatestFinalizedSlot(slot iotago.SlotIndex) *ValidatorBlockBuilder {
	if v.err != nil {
		return v
	}

	v.protocolBlock.LatestFinalizedSlot = slot

	return v
}

func (v *ValidatorBlockBuilder) Sign(accountID iotago.AccountID, prvKey ed25519.PrivateKey) *ValidatorBlockBuilder {
	if v.err != nil {
		return v
	}

	v.protocolBlock.IssuerID = accountID

	signature, err := v.protocolBlock.Sign(v.api, iotago.NewAddressKeysForEd25519Address(iotago.Ed25519AddressFromPubKey(prvKey.Public().(ed25519.PublicKey)), prvKey))
	if err != nil {
		v.err = fmt.Errorf("error signing block: %w", err)
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
func (v *ValidatorBlockBuilder) StrongParents(parents iotago.BlockIDs) *ValidatorBlockBuilder {
	if v.err != nil {
		return v
	}

	v.validatorBlock.StrongParents = parents.RemoveDupsAndSort()

	return v
}

// WeakParents sets the weak parents.
func (v *ValidatorBlockBuilder) WeakParents(parents iotago.BlockIDs) *ValidatorBlockBuilder {
	if v.err != nil {
		return v
	}

	v.validatorBlock.WeakParents = parents.RemoveDupsAndSort()

	return v
}

// ShallowLikeParents sets the shallow like parents.
func (v *ValidatorBlockBuilder) ShallowLikeParents(parents iotago.BlockIDs) *ValidatorBlockBuilder {
	if v.err != nil {
		return v
	}

	v.validatorBlock.ShallowLikeParents = parents.RemoveDupsAndSort()

	return v
}

// HighestSupportedVersion sets the highest supported version.
func (v *ValidatorBlockBuilder) HighestSupportedVersion(highestSupportedVersion byte) *ValidatorBlockBuilder {
	if v.err != nil {
		return v
	}

	v.validatorBlock.HighestSupportedVersion = highestSupportedVersion

	return v
}
