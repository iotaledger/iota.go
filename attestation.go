package iotago

import (
	"bytes"
	"context"

	hiveEd25519 "github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
)

// Attestations is a slice of Attestation.
type Attestations = []*Attestation

type Attestation struct {
	API       API
	Header    BlockHeader `serix:"0,nest"`
	BodyHash  Identifier  `serix:"1,mapKey=bodyHash"`
	Signature Signature   `serix:"2,mapKey=signature"`
}

func NewAttestation(api API, block *Block) *Attestation {
	return &Attestation{
		API:       api,
		Header:    block.Header,
		BodyHash:  lo.PanicOnErr(block.Body.Hash()),
		Signature: block.Signature,
	}
}

func AttestationFromBytes(apiProvider APIProvider) func(bytes []byte) (attestation *Attestation, consumedBytes int, err error) {
	return func(bytes []byte) (attestation *Attestation, consumedBytes int, err error) {
		attestation = new(Attestation)

		var version Version
		if version, consumedBytes, err = VersionFromBytes(bytes); err != nil {
			err = ierrors.Wrap(err, "failed to parse version")
		} else if attestation.API, err = apiProvider.APIForVersion(version); err != nil {
			err = ierrors.Wrapf(err, "failed to retrieve API for version %d", version)
		} else if consumedBytes, err = attestation.API.Decode(bytes, attestation, serix.WithValidation()); err != nil {
			err = ierrors.Wrap(err, "failed to deserialize attestation")
		}

		return attestation, consumedBytes, err
	}
}

func (a *Attestation) SetDeserializationContext(ctx context.Context) {
	a.API = APIFromContext(ctx)
}

func (a *Attestation) Compare(other *Attestation) int {
	switch {
	case a == nil && other == nil:
		return 0
	case a == nil:
		return -1
	case other == nil:
		return 1
	case a.Header.SlotCommitmentID.Slot() > other.Header.SlotCommitmentID.Slot():
		return 1
	case a.Header.SlotCommitmentID.Slot() < other.Header.SlotCommitmentID.Slot():
		return -1
	case a.Header.IssuingTime.After(other.Header.IssuingTime):
		return 1
	case a.Header.IssuingTime.Before(other.Header.IssuingTime):
		return -1
	default:
		return bytes.Compare(a.BodyHash[:], other.BodyHash[:])
	}
}

func (a *Attestation) BlockID() (BlockID, error) {
	signatureBytes, err := a.API.Encode(a.Signature)
	if err != nil {
		return EmptyBlockID, ierrors.Errorf("failed to create blockID: %w", err)
	}

	headerHash, err := a.Header.Hash(a.API)
	if err != nil {
		return EmptyBlockID, ierrors.Errorf("failed to create blockID: %w", err)
	}

	id := blockIdentifier(headerHash, a.BodyHash, signatureBytes)
	slot := a.API.TimeProvider().SlotFromTime(a.Header.IssuingTime)

	return NewBlockID(slot, id), nil
}

func (a *Attestation) signingMessage() ([]byte, error) {
	headerHash, err := a.Header.Hash(a.API)
	if err != nil {
		return nil, ierrors.Errorf("failed to create signing message: %w", err)
	}

	return blockSigningMessage(headerHash, a.BodyHash), nil
}

func (a *Attestation) VerifySignature() (valid bool, err error) {
	signingMessage, err := a.signingMessage()
	if err != nil {
		return false, err
	}

	edSig, isEdSig := a.Signature.(*Ed25519Signature)
	if !isEdSig {
		return false, ierrors.Errorf("only ed2519 signatures supported, got %s", a.Signature.Type())
	}

	return hiveEd25519.Verify(edSig.PublicKey[:], signingMessage, edSig.Signature[:]), nil
}

func (a *Attestation) Bytes() ([]byte, error) {
	return a.API.Encode(a)
}
