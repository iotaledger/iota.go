package iotago

import (
	"context"
	"crypto"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/iota.go/v4/merklehasher"
)

type OutputIDProof struct {
	API                   API
	Slot                  SlotIndex                                       `serix:"0,mapKey=slot"`
	OutputIndex           uint16                                          `serix:"1,mapKey=outputIndex"`
	TransactionCommitment Identifier                                      `serix:"2,mapKey=transactionCommitment"`
	OutputCommitmentProof *merklehasher.Proof[*APIByter[TxEssenceOutput]] `serix:"3,mapKey=outputCommitmentProof"`
}

func OutputIDProofForOutputAtIndex(tx *Transaction, index uint16) (*OutputIDProof, error) {
	if tx.API == nil {
		return nil, ierrors.New("API not set")
	}

	if int(index) >= len(tx.Outputs) {
		return nil, ierrors.Errorf("index %d out of bounds len=%d", index, len(tx.Outputs))
	}

	outputHasher := merklehasher.NewHasher[*APIByter[TxEssenceOutput]](crypto.BLAKE2b_256)
	wrappedOutputs := lo.Map(tx.Outputs, APIByterFactory[TxEssenceOutput](tx.API))

	proof, err := outputHasher.ComputeProofForIndex(wrappedOutputs, int(index))
	if err != nil {
		return nil, ierrors.Wrapf(err, "failed to compute proof for index %d", index)
	}

	transactionCommitment, err := tx.TransactionCommitment()
	if err != nil {
		return nil, err
	}

	return &OutputIDProof{
		API:                   tx.API,
		Slot:                  tx.CreationSlot,
		OutputIndex:           index,
		TransactionCommitment: transactionCommitment,
		OutputCommitmentProof: proof,
	}, nil
}

func OutputIDProofFromBytes(api API) func([]byte) (*OutputIDProof, int, error) {
	return func(b []byte) (proof *OutputIDProof, consumedBytes int, err error) {
		proof = new(OutputIDProof)
		consumedBytes, err = api.Decode(b, proof)

		return proof, consumedBytes, err
	}
}

func (p *OutputIDProof) Bytes() ([]byte, error) {
	return p.API.Encode(p)
}

func (p *OutputIDProof) SetDeserializationContext(ctx context.Context) {
	p.API = APIFromContext(ctx)
}

func (p *OutputIDProof) OutputID(output Output) (OutputID, error) {
	if p.API == nil {
		return EmptyOutputID, ierrors.New("API not set")
	}

	outputHasher := merklehasher.NewHasher[*APIByter[TxEssenceOutput]](crypto.BLAKE2b_256)

	contains, err := p.OutputCommitmentProof.ContainsValue(APIByterFactory[TxEssenceOutput](p.API)(output), outputHasher)
	if err != nil {
		return EmptyOutputID, ierrors.Wrapf(err, "failed to check if proof contains output")
	}

	// The proof does not contain a hash of the output
	if !contains {
		return EmptyOutputID, ierrors.Errorf("proof does not contain the given output")
	}

	// Hash the proof to get the root
	outputCommitment := Identifier(p.OutputCommitmentProof.Hash(outputHasher))

	// Compute the output ID from the contents of the proof
	utxoInput := &UTXOInput{
		TransactionID:          TransactionIDFromTransactionCommitmentAndOutputCommitment(p.Slot, p.TransactionCommitment, outputCommitment),
		TransactionOutputIndex: p.OutputIndex,
	}
	computedOutputID := utxoInput.OutputID()

	return computedOutputID, nil
}

type APIByter[T any] struct {
	API   API
	Value T `serix:"0"`
}

func APIByterFactory[T any](api API) func(value T) *APIByter[T] {
	return func(value T) *APIByter[T] {
		return &APIByter[T]{
			API:   api,
			Value: value,
		}
	}
}

func (a *APIByter[T]) SetDeserializationContext(ctx context.Context) {
	a.API = APIFromContext(ctx)
}

func (a *APIByter[T]) Bytes() ([]byte, error) {
	return a.API.Encode(a.Value)
}
