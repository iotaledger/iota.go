package builder

import (
	"context"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/iota.go/v3/pow"
)

const (
	defaultProtocolVersion = 2
)

// NewBlockBuilder creates a new BlockBuilder.
func NewBlockBuilder() *BlockBuilder {
	return &BlockBuilder{
		block: &iotago.Block{ProtocolVersion: defaultProtocolVersion},
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

	return mb.Parents(parents)
}

// Parents sets the parents.
func (mb *BlockBuilder) Parents(parents iotago.BlockIDs) *BlockBuilder {
	if mb.err != nil {
		return mb
	}

	mb.block.Parents = parents.RemoveDupsAndSort()
	return mb
}

// ProofOfWork does the proof-of-work needed in order to satisfy the given target score.
// It can be canceled by canceling the given context. This function should appear
// as the last step before Build.
func (mb *BlockBuilder) ProofOfWork(ctx context.Context, targetScore float64, numWorkers ...int) *BlockBuilder {
	if mb.err != nil {
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
