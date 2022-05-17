package builder

import (
	"context"
	"fmt"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/iota.go/v3/pow"
)

// NewBlockBuilder creates a new BlockBuilder.
func NewBlockBuilder(protoVersion byte) *BlockBuilder {
	return &BlockBuilder{
		block: &iotago.Block{
			ProtocolVersion: protoVersion,
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

// Payload sets the payload to embed within the block.
func (mb *BlockBuilder) Payload(payload iotago.Payload) *BlockBuilder {
	if mb.err != nil {
		return mb
	}
	mb.block.Payload = payload
	return mb
}

// Tips uses the given Client to query for parents to use.
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

	mb.ParentsBlockIDs(parents)

	return mb
}

// Parents sets the parents of the block.
func (mb *BlockBuilder) Parents(parents [][]byte) *BlockBuilder {
	if mb.err != nil {
		return mb
	}

	pars := make(iotago.BlockIDs, len(parents))
	for i, parentBytes := range parents {
		parent := iotago.BlockID{}
		copy(parent[:], parentBytes)
		pars[i] = parent
	}
	mb.block.Parents = pars.RemoveDupsAndSort()
	return mb
}

// ParentsBlockIDs sets the parents of the block.
func (mb *BlockBuilder) ParentsBlockIDs(parents iotago.BlockIDs) *BlockBuilder {
	if mb.err != nil {
		return mb
	}

	mb.block.Parents = parents.RemoveDupsAndSort()
	return mb
}

// ProofOfWork does the proof-of-work needed in order to satisfy the given target score.
// It can be cancelled by cancelling the given context. This function should appear
// as the last step before Build.
func (mb *BlockBuilder) ProofOfWork(ctx context.Context, protoParas *iotago.ProtocolParameters, targetScore float64, numWorkers ...int) *BlockBuilder {
	if mb.err != nil {
		return mb
	}

	blockData, err := mb.block.Serialize(serializer.DeSeriModePerformValidation, protoParas)
	if err != nil {
		mb.err = err
		return mb
	}

	// cut out the nonce
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
