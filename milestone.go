package iotago

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"

	"golang.org/x/crypto/blake2b"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serix"
	iotagoEd25519 "github.com/iotaledger/iota.go/v3/ed25519"
	"github.com/iotaledger/iota.go/v3/remotesigner"
	"github.com/iotaledger/iota.go/v3/util"
)

const (
	// MilestoneMerkleProofLength defines the length of a merkle proof within a milestone payload.
	MilestoneMerkleProofLength = blake2b.Size256
	// MilestoneSignatureLength defines the length of the milestone signature.
	MilestoneSignatureLength = ed25519.SignatureSize
	// MilestoneIDLength defines the length of a Milestone ID.
	MilestoneIDLength = blake2b.Size256
	// MilestonePublicKeyLength defines the length of a public key within a milestone.
	MilestonePublicKeyLength = ed25519.PublicKeySize
	// MaxSignaturesInAMilestone is the maximum amount of signatures in a milestone.
	MaxSignaturesInAMilestone = 255
	// MinSignaturesInAMilestone is the minimum amount of signatures in a milestone.
	MinSignaturesInAMilestone = 1
)

var (
	// ErrMilestoneTooFewSignatures gets returned if a to be deserialized Milestone does not contain at least one signature.
	ErrMilestoneTooFewSignatures = errors.New("a milestone must hold at least one signature")
	// ErrMilestoneTooFewSignaturesForVerificationThreshold gets returned if there are less signatures within a Milestone than the min. threshold.
	ErrMilestoneTooFewSignaturesForVerificationThreshold = errors.New("too few signatures for verification")
	// ErrMilestoneProducedSignaturesCountMismatch gets returned when a MilestoneSigningFunc produces less signatures than expected.
	ErrMilestoneProducedSignaturesCountMismatch = errors.New("produced and wanted signature count mismatch")
	// ErrMilestoneTooManySignatures gets returned when a Milestone holds more than 255 signatures.
	ErrMilestoneTooManySignatures = fmt.Errorf("a milestone can hold max %d signatures", MaxSignaturesInAMilestone)
	// ErrMilestoneInvalidMinSignatureThreshold gets returned when an invalid min signatures threshold is given to the verification function.
	ErrMilestoneInvalidMinSignatureThreshold = fmt.Errorf("min threshold must be at least 1")
	// ErrMilestoneNonApplicablePublicKey gets returned when a Milestone contains a public key which isn't in the applicable public key set.
	ErrMilestoneNonApplicablePublicKey = fmt.Errorf("non applicable public key found")
	// ErrMilestoneSignatureThresholdGreaterThanApplicablePublicKeySet gets returned when a min. signature threshold is greater than a given applicable public key set.
	ErrMilestoneSignatureThresholdGreaterThanApplicablePublicKeySet = fmt.Errorf("the min. signature threshold must be less or equal the applicable public key set")
	// ErrMilestoneInvalidSignature gets returned when a Milestone's signature is invalid.
	ErrMilestoneInvalidSignature = fmt.Errorf("invalid milestone signature")
	// ErrMilestoneInMemorySignerPrivateKeyMissing gets returned when an InMemoryEd25519MilestoneSigner is missing a private key.
	ErrMilestoneInMemorySignerPrivateKeyMissing = fmt.Errorf("private key missing")
)

type (
	// MilestonePublicKey is a public key within a Milestone.
	MilestonePublicKey [MilestonePublicKeyLength]byte
	// MilestonePublicKeySet is a set of unique MilestonePublicKey.
	MilestonePublicKeySet = map[MilestonePublicKey]struct{}
	// MilestonePublicKeyMapping is a mapping from a public key to a private key.
	MilestonePublicKeyMapping = map[MilestonePublicKey]ed25519.PrivateKey
	// MilestoneMerkleProof is the merkle root within a milestone.
	MilestoneMerkleProof [MilestoneMerkleProofLength]byte
)

// MilestoneIndex is the index of a Milestone.
type MilestoneIndex = uint32

// MilestoneID is the ID of a Milestone.
type MilestoneID [MilestoneIDLength]byte

func (id MilestoneID) MarshalText() (text []byte, err error) {
	dst := make([]byte, hex.EncodedLen(len(MilestoneID{})))
	hex.Encode(dst, id[:])
	return dst, nil
}

func (id *MilestoneID) UnmarshalText(text []byte) error {
	_, err := hex.Decode(id[:], text)
	return err
}

// ToHex converts the given milestone ID to their hex representation.
func (id MilestoneID) ToHex() string {
	return EncodeHex(id[:])
}

// Empty tells whether the MilestoneID is empty.
func (id MilestoneID) Empty() bool {
	return id == MilestoneID(emptyBlockID)
}

func (id *MilestoneID) String() string {
	return id.ToHex()
}

// NewMilestone creates a new unsigned Milestone.
func NewMilestone(index MilestoneIndex, timestamp uint32, protocolVersion byte, prevMsID MilestoneID, parents MilestoneParentIDs, inclMerkleProof MilestoneMerkleProof, appliedMerkleRoot MilestoneMerkleProof) *Milestone {
	return &Milestone{
		MilestoneEssence: MilestoneEssence{
			Index:               index,
			Timestamp:           timestamp,
			ProtocolVersion:     protocolVersion,
			PreviousMilestoneID: prevMsID,
			Parents:             parents,
			InclusionMerkleRoot: inclMerkleProof,
			AppliedMerkleRoot:   appliedMerkleRoot,
		},
	}
}

// MilestoneSignature is a signature of a milestone.
type MilestoneSignature interface {
	Signature
}

// MilestoneParentIDs is a slice of BlockIDs the milestone references.
type MilestoneParentIDs BlockIDs

// Milestone represents a special payload which defines the inclusion set of other messages in the Tangle.
type Milestone struct {
	MilestoneEssence `serix:"0"`
	// The signatures held by the milestone.
	Signatures Signatures[MilestoneSignature] `serix:"1,mapKey=signatures"`
}

// MilestoneEssence is the essence part of a Milestone.
type MilestoneEssence struct {
	// The index of this milestone.
	Index MilestoneIndex `serix:"0,mapKey=index"`
	// The time at which this milestone was issued.
	Timestamp uint32 `serix:"1,mapKey=timestamp"`
	// The protocol version under which this milestone operates.
	ProtocolVersion byte `serix:"2,mapKey=protocolVersion"`
	// The pointer to the previous milestone.
	// Zeroed if there wasn't a previous milestone.
	PreviousMilestoneID MilestoneID `serix:"3,mapKey=previousMilestoneId"`
	// The parents where this milestone attaches to.
	Parents MilestoneParentIDs `serix:"4,mapKey=parents"`
	// The merkle root of all directly/indirectly referenced blocks (their IDs) which
	// were newly included by this milestone.
	InclusionMerkleRoot MilestoneMerkleProof `serix:"5,mapKey=inclusionMerkleRoot"`
	// The merkle root of all blocks (their IDs) carrying ledger state mutating transactions.
	AppliedMerkleRoot MilestoneMerkleProof `serix:"6,mapKey=appliedMerkleRoot"`
	// The metadata associated with this milestone.
	Metadata []byte `serix:"7,lengthPrefixType=uint16,mapKey=metadata,omitempty,maxLen=8192"`
	// The milestone options carried with this milestone.
	Opts MilestoneOpts `serix:"8,mapKey=options,omitempty"`
}

func (m *Milestone) PayloadType() PayloadType {
	return PayloadMilestone
}

// ID computes the ID of the Milestone.
func (m *Milestone) ID() (MilestoneID, error) {
	var msID MilestoneID
	data, err := m.Essence()
	if err != nil {
		return MilestoneID{}, err
	}
	copy(msID[:], data)
	return msID, nil
}

// MustID works like ID but panics if there is an error.
func (m *Milestone) MustID() MilestoneID {
	id, err := m.ID()
	if err != nil {
		panic(err)
	}
	return id
}

// Essence returns the essence bytes (the bytes to be signed) of the Milestone.
func (m *Milestone) Essence() ([]byte, error) {
	essenceBytes, err := internalEncode(m.MilestoneEssence, serix.WithValidation())
	if err != nil {
		return nil, err
	}
	essenceHash := blake2b.Sum256(essenceBytes)
	return essenceHash[:], nil
}

// VerifySignatures verifies that min. minSigThreshold signatures occur in the Milestone and that all
// signatures within it are valid with respect to the given applicable public key set.
// The public key set must only contain keys applicable for the given Milestone index.
// The caller must only call this function on a Milestone which was deserialized with validation.
func (m *Milestone) VerifySignatures(minSigThreshold int, applicablePubKeys MilestonePublicKeySet) error {
	switch {
	case minSigThreshold == 0:
		return ErrMilestoneInvalidMinSignatureThreshold
	case len(m.Signatures) == 0:
		return ErrMilestoneTooFewSignatures
	case len(m.Signatures) < minSigThreshold:
		return fmt.Errorf("%w: wanted min. %d but only had %d", ErrMilestoneTooFewSignaturesForVerificationThreshold, minSigThreshold, len(m.Signatures))
	case len(applicablePubKeys) < minSigThreshold:
		return ErrMilestoneSignatureThresholdGreaterThanApplicablePublicKeySet
	}

	msEssence, err := m.Essence()
	if err != nil {
		return fmt.Errorf("unable to compute milestone essence for signature verification: %w", err)
	}

	// it is guaranteed by the serialization logic that this milestone does not contain duplicated public keys.
	for msSigIndex, msSig := range m.Signatures {
		// guaranteed by deserialization
		edSig := msSig.(*Ed25519Signature)

		if _, has := applicablePubKeys[edSig.PublicKey]; !has {
			return fmt.Errorf("%w: public key %s is not applicable", ErrMilestoneNonApplicablePublicKey, EncodeHex(edSig.PublicKey[:]))
		}

		if ok := iotagoEd25519.Verify(edSig.PublicKey[:], msEssence[:], edSig.Signature[:]); !ok {
			return fmt.Errorf("%w: at index %d, %s", ErrMilestoneInvalidSignature, msSigIndex, edSig)
		}
	}

	return nil
}

// MilestoneSigningFunc is a function which produces a set of signatures for the given Milestone essence data.
// The given public keys dictate in which order the returned signatures must occur.
type MilestoneSigningFunc func(pubKeys []MilestonePublicKey, msEssence []byte) ([][MilestoneSignatureLength]byte, error)

// InMemoryEd25519MilestoneSigner is a function which uses the provided Ed25519 MilestonePublicKeyMapping to produce signatures for the Milestone essence data.
func InMemoryEd25519MilestoneSigner(prvKeys MilestonePublicKeyMapping) MilestoneSigningFunc {
	return func(pubKeys []MilestonePublicKey, msEssence []byte) ([][MilestoneSignatureLength]byte, error) {
		sigs := make([][MilestoneSignatureLength]byte, len(pubKeys))
		for i, pubKey := range pubKeys {
			prvKey, ok := prvKeys[pubKey]
			if !ok {
				return nil, fmt.Errorf("%w: needed for public key %s", ErrMilestoneInMemorySignerPrivateKeyMissing, EncodeHex(pubKey[:]))
			}
			sig := ed25519.Sign(prvKey, msEssence)
			copy(sigs[i][:], sig)
		}
		return sigs, nil
	}
}

// InsecureRemoteEd25519MilestoneSigner is a function which uses a remote RPC server via an insecure connection
// to produce signatures for the Milestone essence data.
// You must only use this function if the remote lives on the same host as the caller.
func InsecureRemoteEd25519MilestoneSigner(remoteEndpoint string) MilestoneSigningFunc {
	return func(pubKeys []MilestonePublicKey, msEssence []byte) ([][MilestoneSignatureLength]byte, error) {
		pubKeysUnbound := make([][]byte, len(pubKeys))
		for i := range pubKeys {
			pubKeysUnbound[i] = make([]byte, 32)
			copy(pubKeysUnbound[i][:], pubKeys[i][:32])
		}

		// insecure because this RPC remote should be local; in turns, it employs TLS mutual authentication to reach the actual signers.
		conn, err := grpc.Dial(remoteEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		defer conn.Close()

		client := remotesigner.NewSignatureDispatcherClient(conn)
		response, err := client.SignMilestone(context.Background(), &remotesigner.SignMilestoneRequest{
			PubKeys:   pubKeysUnbound,
			MsEssence: msEssence,
		})
		if err != nil {
			return nil, err
		}

		sigs := response.GetSignatures()
		if len(sigs) != len(pubKeys) {
			return nil, fmt.Errorf("%w: remote did not provide the correct count of signatures", ErrMilestoneProducedSignaturesCountMismatch)
		}

		sigs64 := make([][MilestoneSignatureLength]byte, len(sigs))
		for i := range sigs {
			copy(sigs64[i][:], sigs[i][:64])
		}
		return sigs64, nil
	}
}

// Sign produces the signatures with the given envelope block and updates the Signatures field of the Milestone
// with the resulting signatures of the given MilestoneSigningFunc. pubKeys are passed to the given MilestoneSigningFunc
// so it can determine which signatures to produce.
func (m *Milestone) Sign(pubKeys []MilestonePublicKey, signingFunc MilestoneSigningFunc) error {
	msEssence, err := m.Essence()
	if err != nil {
		return fmt.Errorf("unable to compute milestone essence for signing: %w", err)
	}

	sigs, err := signingFunc(pubKeys, msEssence)
	if err != nil {
		return fmt.Errorf("unable to produce milestone signatures: %w", err)
	}

	switch {
	case len(pubKeys) != len(sigs):
		return fmt.Errorf("%w: wanted %d signatures but only produced %d", ErrMilestoneProducedSignaturesCountMismatch, len(pubKeys), len(sigs))
	case len(sigs) < MinSignaturesInAMilestone:
		return fmt.Errorf("%w: not enough signatures were produced during signing", ErrMilestoneTooFewSignatures)
	case len(sigs) > MaxSignaturesInAMilestone:
		return fmt.Errorf("%w: too many signatures were produced during signing", ErrMilestoneTooManySignatures)
	}

	edSigs := make(Signatures[MilestoneSignature], len(sigs))
	for i, sig := range sigs {
		edSig := &Ed25519Signature{}
		copy(edSig.PublicKey[:], pubKeys[i][:])
		copy(edSig.Signature[:], sig[:])
		edSigs[i] = edSig
	}

	sort.Slice(edSigs, func(i, j int) bool {
		return bytes.Compare(edSigs[i].(*Ed25519Signature).PublicKey[:], edSigs[j].(*Ed25519Signature).PublicKey[:]) == -1
	})

	m.Signatures = edSigs
	return nil
}

func (m *Milestone) Size() int {
	// 1 byte for length prefixes
	parentBlocksByteLen := serializer.OneByte + BlockIDLength*len(m.Parents)
	signatureByteLen := serializer.OneByte + ((&Ed25519Signature{}).Size())*len(m.Signatures)
	metadataLen := serializer.UInt16ByteSize + len(m.Metadata)

	return util.NumByteLen(uint32(PayloadMilestone)) + util.NumByteLen(m.Index) + util.NumByteLen(m.Timestamp) +
		util.NumByteLen(m.ProtocolVersion) + MilestoneIDLength + parentBlocksByteLen + MilestoneMerkleProofLength +
		MilestoneMerkleProofLength + metadataLen + m.Opts.Size() + signatureByteLen
}
