package builder

import (
	"crypto/ed25519"
	"fmt"
	"time"

	iotago "github.com/iotaledger/iota.go/v4"
)

const (
	defaultProtocolVersion = 3
)

// NewBlockBuilder creates a new BlockBuilder.
func NewBlockBuilder(blockType iotago.BlockType) *BlockBuilder {
	b := &BlockBuilder{
		protocolBlock: &iotago.ProtocolBlock{
			ProtocolVersion: defaultProtocolVersion,
			SlotCommitment:  iotago.NewEmptyCommitment(),
			IssuingTime:     time.Now(),
			Signature:       &iotago.Ed25519Signature{},
		},
	}

	var block iotago.Block
	switch blockType {
	case iotago.BlockTypeBasic:
		b.basicBlock = &iotago.BasicBlock{}
		block = b.basicBlock
	case iotago.BlockTypeValidator:
		b.validatorBlock = &iotago.ValidatorBlock{}
		block = b.validatorBlock
	default:
		panic("unknown block type")
	}

	b.protocolBlock.Block = block

	return b
}

// BlockBuilder is used to easily build up a Block.
type BlockBuilder struct {
	basicBlock     *iotago.BasicBlock
	validatorBlock *iotago.ValidatorBlock

	protocolBlock *iotago.ProtocolBlock
	err           error
}

// Build builds the ProtocolBlock or returns any error which occurred during the build steps.
func (b *BlockBuilder) Build() (*iotago.ProtocolBlock, error) {
	if b.err != nil {
		return nil, b.err
	}

	return b.protocolBlock, nil
}

// Payload sets the payload.
func (b *BlockBuilder) Payload(payload iotago.Payload) *BlockBuilder {
	if b.err != nil {
		return b
	}

	if b.basicBlock == nil {
		panic("can't set payload on non-basic block")
	}
	b.basicBlock.Payload = payload

	return b
}

// ProtocolVersion sets the protocol version.
func (b *BlockBuilder) ProtocolVersion(version byte) *BlockBuilder {
	if b.err != nil {
		return b
	}

	b.protocolBlock.ProtocolVersion = version

	return b
}

func (b *BlockBuilder) IssuingTime(time time.Time) *BlockBuilder {
	if b.err != nil {
		return b
	}

	b.protocolBlock.IssuingTime = time

	return b
}

// StrongParents sets the strong parents.
func (b *BlockBuilder) StrongParents(parents iotago.BlockIDs) *BlockBuilder {
	if b.err != nil {
		return b
	}

	b.protocolBlock.Block.SetStrongParentIDs(parents.RemoveDupsAndSort())

	return b
}

// WeakParents sets the weak parents.
func (b *BlockBuilder) WeakParents(parents iotago.BlockIDs) *BlockBuilder {
	if b.err != nil {
		return b
	}

	b.protocolBlock.Block.SetWeakParentIDs(parents.RemoveDupsAndSort())

	return b
}

// ShallowLikeParents sets the shallow like parents.
func (b *BlockBuilder) ShallowLikeParents(parents iotago.BlockIDs) *BlockBuilder {
	if b.err != nil {
		return b
	}

	b.protocolBlock.Block.SetShallowLikeParentIDs(parents.RemoveDupsAndSort())

	return b
}

// SlotCommitment sets the slot commitment.
func (b *BlockBuilder) SlotCommitment(commitment *iotago.Commitment) *BlockBuilder {
	if b.err != nil {
		return b
	}

	b.protocolBlock.SlotCommitment = commitment

	return b
}

// BurnedMana sets the amount of mana burned by the block.
func (b *BlockBuilder) BurnedMana(burnedMana iotago.Mana) *BlockBuilder {
	if b.err != nil {
		return b
	}

	if b.basicBlock == nil {
		panic("can't set burned mana on non-basic block")
	}
	b.basicBlock.BurnedMana = burnedMana

	return b
}

// LatestFinalizedSlot sets the latest finalized slot.
func (b *BlockBuilder) LatestFinalizedSlot(index iotago.SlotIndex) *BlockBuilder {
	if b.err != nil {
		return b
	}

	b.protocolBlock.LatestFinalizedSlot = index

	return b
}

func (b *BlockBuilder) Sign(accountID iotago.AccountID, prvKey ed25519.PrivateKey) *BlockBuilder {
	if b.err != nil {
		return b
	}

	b.protocolBlock.IssuerID = accountID

	signature, err := b.protocolBlock.Sign(iotago.NewAddressKeysForEd25519Address(iotago.Ed25519AddressFromPubKey(prvKey.Public().(ed25519.PublicKey)), prvKey))
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
