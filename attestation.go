package iotago

import (
	"bytes"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
	iotagoEd25519 "github.com/iotaledger/iota.go/v4/ed25519"
)

type Attestation struct {
	BlockID          BlockID      `serix:"0,mapKey=blockID"`
	IssuerID         AccountID    `serix:"1,mapKey=issuerID"`
	IssuingTime      time.Time    `serix:"2,mapKey=issuingTime"`
	SlotCommitmentID CommitmentID `serix:"3,mapKey=slotCommitmentID"`
	BlockContentHash Identifier   `serix:"4,mapKey=blockContentHash"`
	Signature        Signature    `serix:"5,mapKey=signature"`
}

func NewAttestation(block *Block, slotTimeProvider *SlotTimeProvider) *Attestation {
	return &Attestation{
		BlockID:          block.MustID(slotTimeProvider),
		IssuerID:         block.IssuerID,
		IssuingTime:      block.IssuingTime,
		SlotCommitmentID: block.SlotCommitment.MustID(),
		BlockContentHash: lo.PanicOnErr(block.ContentHash()),
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
	case a.IssuingTime.After(other.IssuingTime):
		return 1
	case other.IssuingTime.After(a.IssuingTime):
		return -1
	default:
		return bytes.Compare(a.BlockContentHash[:], other.BlockContentHash[:])
	}
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
