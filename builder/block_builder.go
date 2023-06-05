package builder

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/nodeclient"
	"github.com/iotaledger/iota.go/v4/pow"
)

const (
	defaultProtocolVersion = 3
)

// NewBlockBuilder creates a new BlockBuilder.
func NewBlockBuilder() *BlockBuilder {
	return &BlockBuilder{
		block: &iotago.Block{
			ProtocolVersion:  defaultProtocolVersion,
			SlotCommitment:   iotago.NewEmptyCommitment(),
			IssuingTimestamp: uint64(time.Now().UnixNano()),
			Signature:        &iotago.Ed25519Signature{},
		},
	}
}

// BlockBuilder is used to easily build up a Block.
type BlockBuilder struct {
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

	mb.block.SetIssuingTime(time)

	return mb
}

// Tips queries the node API for tips/parents and sets them accordingly.
func (mb *BlockBuilder) Tips(ctx context.Context, nodeAPI *nodeclient.Client) *BlockBuilder {
	if mb.err != nil {
		return mb
	}

	res, err := nodeAPI.Tips(ctx)
	if err != nil {
		mb.err = fmt.Errorf("unable to fetch tips from node API: %w", err)
		return mb
	}

	parents, err := res.Tips()
	if err != nil {
		mb.err = fmt.Errorf("unable to fetch tips: %w", err)
		return mb
	}

	return mb.StrongParents(parents)
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

// LatestFinalizedSlot sets the latest finalized slot.
func (mb *BlockBuilder) LatestFinalizedSlot(index iotago.SlotIndex) *BlockBuilder {
	if mb.err != nil {
		return mb
	}

	mb.block.LatestFinalizedSlot = index

	return mb
}

// ProofOfWork does the proof-of-work needed in order to satisfy the given target score.
// It can be canceled by canceling the given context. This function should appear
// as the last step before Build.
func (mb *BlockBuilder) ProofOfWork(ctx context.Context, targetScore float64, numWorkers ...int) *BlockBuilder {
	if mb.err != nil {
		return mb
	}

	if targetScore == 0 {
		return mb
	}

	// cut out the nonce
	_, blockData, err := mb.block.POW()
	if err != nil {
		mb.err = fmt.Errorf("unable to compute pow relevant data: %w", err)
		return mb
	}
	powRelevantData := blockData[:len(blockData)-serializer.UInt64ByteSize]
	worker := pow.New(numWorkers...)
	nonce, err := worker.Mine(ctx, powRelevantData, targetScore)
	if err != nil {
		mb.err = fmt.Errorf("unable to complete proof-of-work: %w", err)
		return mb
	}
	mb.block.Nonce = nonce

	return mb
}

func (mb *BlockBuilder) Sign(accountID iotago.AccountID, prvKey ed25519.PrivateKey) *BlockBuilder {
	if mb.err != nil {
		return mb
	}

	mb.block.IssuerID = accountID

	signature, err := mb.block.Sign(iotago.NewAddressKeysForEd25519Address(iotago.Ed25519AddressFromPubKey(prvKey.Public().(ed25519.PublicKey)), prvKey))
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
