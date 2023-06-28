package iotago

import (
	"bytes"
	"fmt"
	"time"

	hiveEd25519 "github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
)

// Attestations is a slice of Attestation.
type Attestations = []*Attestation

type Attestation struct {
	IssuerID         AccountID    `serix:"0,mapKey=issuerID"`
	IssuingTime      time.Time    `serix:"1,mapKey=issuingTime"`
	SlotCommitmentID CommitmentID `serix:"2,mapKey=slotCommitmentID"`
	BlockContentHash Identifier   `serix:"3,mapKey=blockContentHash"`
	Signature        Signature    `serix:"4,mapKey=signature"`
}

func NewAttestation(api API, block *Block) *Attestation {
	return &Attestation{
		IssuerID:         block.IssuerID,
		IssuingTime:      block.IssuingTime,
		SlotCommitmentID: block.SlotCommitment.MustID(api),
		BlockContentHash: lo.PanicOnErr(block.ContentHash(api)),
		Signature:        block.Signature,
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
		return bytes.Compare(a.BlockContentHash[:], other.BlockContentHash[:])
	}
}

func (a Attestation) BlockID(api API) (BlockID, error) {
	signatureBytes, err := api.Encode(a.Signature)
	if err != nil {
		return EmptyBlockID(), fmt.Errorf("failed to serialize block's signature: %w", err)
	}

	blockIdentifier := IdentifierFromData(byteutils.ConcatBytes(a.BlockContentHash[:], signatureBytes[:]))
	slotIndex := api.TimeProvider().SlotFromTime(a.IssuingTime)

	return NewSlotIdentifier(slotIndex, blockIdentifier), nil
}

func (a Attestation) Bytes(api API) (bytes []byte, err error) {
	return api.Encode(a)
}

func (a *Attestation) FromBytes(api API, bytes []byte) (consumedBytes int, err error) {
	return api.Decode(bytes, a)
}

func (a *Attestation) signingMessage(api API) ([]byte, error) {
	issuingTimeBytes, err := api.Encode(a.IssuingTime)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize block's issuing time: %w", err)
	}

	return byteutils.ConcatBytes(issuingTimeBytes, a.SlotCommitmentID[:], a.BlockContentHash[:]), nil
}

func (a *Attestation) VerifySignature(api API) (valid bool, err error) {
	signingMessage, err := a.signingMessage(api)
	if err != nil {
		return false, err
	}

	edSig, isEdSig := a.Signature.(*Ed25519Signature)
	if !isEdSig {
		return false, fmt.Errorf("only ed2519 signatures supported, got %s", a.Signature.Type())
	}

	return hiveEd25519.Verify(edSig.PublicKey[:], signingMessage, edSig.Signature[:]), nil
}
