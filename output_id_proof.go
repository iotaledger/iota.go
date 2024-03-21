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
	Slot                  SlotIndex                                       `serix:""`
	OutputIndex           uint16                                          `serix:""`
	TransactionCommitment Identifier                                      `serix:""`
	OutputCommitmentProof *merklehasher.Proof[*APIByter[TxEssenceOutput]] `serix:""`
}

func OutputIDProofFromTransaction(tx *Transaction, outputIndex uint16) (*OutputIDProof, error) {
	if tx.API == nil {
		panic("API on transaction not set")
	}

	transactionCommitment, err := tx.TransactionCommitment()
	if err != nil {
		return nil, err
	}

	return NewOutputIDProof(tx.API, transactionCommitment, tx.CreationSlot, tx.Outputs, outputIndex)
}

func NewOutputIDProof(api API, txCommitment Identifier, txCreationSlot SlotIndex, outputs TxEssenceOutputs, outputIndex uint16) (*OutputIDProof, error) {
	if int(outputIndex) >= len(outputs) {
		return nil, ierrors.Errorf("index %d out of bounds for outputs slice of len %d", outputIndex, len(outputs))
	}

	outputHasher := merklehasher.NewHasher[*APIByter[TxEssenceOutput]](crypto.BLAKE2b_256)
	wrappedOutputs := lo.Map(outputs, APIByterFactory[TxEssenceOutput](api))

	proof, err := outputHasher.ComputeProofForIndex(wrappedOutputs, int(outputIndex))
	if err != nil {
		return nil, ierrors.Wrapf(err, "failed to compute proof for index %d", outputIndex)
	}

	return &OutputIDProof{
		API:                   api,
		Slot:                  txCreationSlot,
		OutputIndex:           outputIndex,
		TransactionCommitment: txCommitment,
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
		panic("API on OutputIDProof not set")
	}

	outputHasher := merklehasher.NewHasher[*APIByter[TxEssenceOutput]](crypto.BLAKE2b_256)

	contains, err := p.OutputCommitmentProof.ContainsValue(APIByterFactory[TxEssenceOutput](p.API)(output), outputHasher)
	if err != nil {
		return EmptyOutputID, ierrors.Wrapf(err, "failed to check if proof contains output")
	}

	// The proof does not contain a hash of the output
	if !contains {
		return EmptyOutputID, ierrors.New("proof does not contain the given output")
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
	Value T `serix:",inlined"`
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
