package iotago

import (
	"bytes"

	hiveEd25519 "github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
)

// Attestations is a slice of Attestation.
type Attestations = []*Attestation

type Attestation struct {
	BlockHeader `serix:"0"`
	BlockHash   Identifier `serix:"1,mapKey=blockHash"`
	Signature   Signature  `serix:"2,mapKey=signature"`
}

func NewAttestation(api API, block *ProtocolBlock) *Attestation {
	return &Attestation{
		BlockHeader: block.BlockHeader,
		BlockHash:   lo.PanicOnErr(block.Block.Hash(api)),
		Signature:   block.Signature,
	}
}

func (a *Attestation) Compare(other *Attestation) int {
	switch {
	case a == nil && other == nil:
		return 0
	case a == nil:
		return -1
	case other == nil:
		return 1
	case a.SlotCommitmentID.Index() > other.SlotCommitmentID.Index():
		return 1
	case a.SlotCommitmentID.Index() < other.SlotCommitmentID.Index():
		return -1
	case a.IssuingTime.After(other.IssuingTime):
		return 1
	case a.IssuingTime.Before(other.IssuingTime):
		return -1
	default:
		return bytes.Compare(a.BlockHash[:], other.BlockHash[:])
	}
}

func (a Attestation) BlockID(api API) (BlockID, error) {
	signatureBytes, err := api.Encode(a.Signature)
	if err != nil {
		return EmptyBlockID(), ierrors.Errorf("failed to create blockID: %w", err)
	}

	headerHash, err := a.BlockHeader.Hash(api)
	if err != nil {
		return EmptyBlockID(), ierrors.Errorf("failed to create blockID: %w", err)
	}

	id := blockIdentifier(headerHash, a.BlockHash, signatureBytes)
	slotIndex := api.TimeProvider().SlotFromTime(a.IssuingTime)

	return NewSlotIdentifier(slotIndex, id), nil
}

func (a *Attestation) signingMessage(api API) ([]byte, error) {
	headerHash, err := a.BlockHeader.Hash(api)
	if err != nil {
		return nil, ierrors.Errorf("failed to create signing message: %w", err)
	}

	return blockSigningMessage(headerHash, a.BlockHash), nil
}

func (a *Attestation) VerifySignature(api API) (valid bool, err error) {
	signingMessage, err := a.signingMessage(api)
	if err != nil {
		return false, err
	}

	edSig, isEdSig := a.Signature.(*Ed25519Signature)
	if !isEdSig {
		return false, ierrors.Errorf("only ed2519 signatures supported, got %s", a.Signature.Type())
	}

	return hiveEd25519.Verify(edSig.PublicKey[:], signingMessage, edSig.Signature[:]), nil
}
