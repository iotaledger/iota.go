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
func NewBlockBuilder(api iotago.API) *BlockBuilder {
	return &BlockBuilder{
		block: &iotago.Block{
			ProtocolVersion: defaultProtocolVersion,
			SlotCommitment:  iotago.NewEmptyCommitment(),
			IssuingTime:     time.Now(),
			Signature:       &iotago.Ed25519Signature{},
		},
	}
}

// BlockBuilder is used to easily build up a Block.
type BlockBuilder struct {
	api   iotago.API
	block *iotago.Block
	err   error
}

// Build builds the Block or returns any error which occurred during the build steps.
func (mb *BlockBuilder) Build() (*iotago.Block, error) {
	if mb.err != nil {
		return nil, mb.err
	}

	return mb.block, nil
}

// Payload sets the payload.
func (mb *BlockBuilder) Payload(payload iotago.Payload) *BlockBuilder {
	if mb.err != nil {
		return mb
	}

	mb.block.Payload = payload

	return mb
}

// ProtocolVersion sets the protocol version.
func (mb *BlockBuilder) ProtocolVersion(version byte) *BlockBuilder {
	if mb.err != nil {
		return mb
	}

	mb.block.ProtocolVersion = version

	return mb
}

func (mb *BlockBuilder) IssuingTime(time time.Time) *BlockBuilder {
	if mb.err != nil {
		return mb
	}

	mb.block.IssuingTime = time

	return mb
}

// StrongParents sets the strong parents.
func (mb *BlockBuilder) StrongParents(parents iotago.BlockIDs) *BlockBuilder {
	if mb.err != nil {
		return mb
	}

	mb.block.StrongParents = parents.RemoveDupsAndSort()

	return mb
}

// WeakParents sets the weak parents.
func (mb *BlockBuilder) WeakParents(parents iotago.BlockIDs) *BlockBuilder {
	if mb.err != nil {
		return mb
	}

	mb.block.WeakParents = parents.RemoveDupsAndSort()

	return mb
}

// ShallowLikeParents sets the shallow like parents.
func (mb *BlockBuilder) ShallowLikeParents(parents iotago.BlockIDs) *BlockBuilder {
	if mb.err != nil {
		return mb
	}

	mb.block.ShallowLikeParents = parents.RemoveDupsAndSort()

	return mb
}

// SlotCommitment sets the slot commitment.
func (mb *BlockBuilder) SlotCommitment(commitment *iotago.Commitment) *BlockBuilder {
	if mb.err != nil {
		return mb
	}

	mb.block.SlotCommitment = commitment

	return mb
}

// BurnedMana sets the amount of mana burned by the block.
func (mb *BlockBuilder) BurnedMana(burnedMana iotago.Mana) *BlockBuilder {
	if mb.err != nil {
		return mb
	}

	mb.block.BurnedMana = burnedMana

	return mb
}

// LatestFinalizedSlot sets the latest finalized slot.
func (mb *BlockBuilder) LatestFinalizedSlot(index iotago.SlotIndex) *BlockBuilder {
	if mb.err != nil {
		return mb
	}

	mb.block.LatestFinalizedSlot = index

	return mb
}

func (mb *BlockBuilder) Sign(accountID iotago.AccountID, prvKey ed25519.PrivateKey) *BlockBuilder {
	if mb.err != nil {
		return mb
	}

	mb.block.IssuerID = accountID

	signature, err := mb.block.Sign(mb.api, iotago.NewAddressKeysForEd25519Address(iotago.Ed25519AddressFromPubKey(prvKey.Public().(ed25519.PublicKey)), prvKey))
	if err != nil {
		mb.err = fmt.Errorf("error signing block: %w", err)
		return mb
	}

	edSig, isEdSig := signature.(*iotago.Ed25519Signature)
	if !isEdSig {
		panic("unsupported signature type")
	}

	mb.block.Signature = edSig

	return mb
}
