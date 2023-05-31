package iotago

import (
	"bytes"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
	iotagoEd25519 "github.com/iotaledger/iota.go/v4/ed25519"
)

// Attestations is a slice of Attestation.
type Attestations = []*Attestation

type Attestation struct {
	IssuerID         AccountID    `serix:"0,mapKey=issuerID"`
	IssuingTime      time.Time    `serix:"1,mapKey=issuingTime"`
	SlotCommitmentID CommitmentID `serix:"2,mapKey=slotCommitmentID"`
	BlockContentHash Identifier   `serix:"3,mapKey=blockContentHash"`
	Signature        Signature    `serix:"4,mapKey=signature"`
	Nonce            uint64       `serix:"5,mapKey=nonce"`
}

func NewAttestation(block *Block) *Attestation {
	return &Attestation{
		IssuerID:         block.IssuerID,
		IssuingTime:      block.IssuingTime,
		SlotCommitmentID: block.SlotCommitment.MustID(),
		BlockContentHash: lo.PanicOnErr(block.ContentHash()),
		Signature:        block.Signature,
		Nonce:            block.Nonce,
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

func (a Attestation) BlockID(slotTimeProvider *SlotTimeProvider) (BlockID, error) {
	signatureBytes, err := internalEncode(a.Signature)
	if err != nil {
		return EmptyBlockID(), fmt.Errorf("failed to serialize block's signature: %w", err)
	}

	nonceBytes, err := internalEncode(a.Nonce)
	if err != nil {
		return EmptyBlockID(), fmt.Errorf("failed to serialize block's nonce: %w", err)
	}

	blockIdentifier := IdentifierFromData(byteutils.ConcatBytes(a.BlockContentHash[:], signatureBytes[:], nonceBytes[:]))
	slotIndex := slotTimeProvider.IndexFromTime(a.IssuingTime)

	return NewSlotIdentifier(slotIndex, blockIdentifier), nil
}

func (a Attestation) Bytes() (bytes []byte, err error) {
	return internalEncode(a)
}

func (a *Attestation) FromBytes(bytes []byte) (consumedBytes int, err error) {
	return internalDecode(bytes, a)
}

func (a *Attestation) signingMessage() ([]byte, error) {
	issuingTimeBytes, err := internalEncode(a.IssuingTime)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize block's issuing time: %w", err)
	}

	return byteutils.ConcatBytes(issuingTimeBytes, a.SlotCommitmentID[:], a.BlockContentHash[:]), nil
}

func (a *Attestation) VerifySignature() (valid bool, err error) {
	signingMessage, err := a.signingMessage()
	if err != nil {
		return false, err
	}

	edSig, isEdSig := a.Signature.(*Ed25519Signature)
	if !isEdSig {
		return false, fmt.Errorf("only ed2519 signatures supported, got %s", a.Signature.Type())
	}

	return iotagoEd25519.Verify(edSig.PublicKey[:], signingMessage, edSig.Signature[:]), nil
}
