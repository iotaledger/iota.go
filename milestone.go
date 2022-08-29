package iotago

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"golang.org/x/crypto/blake2b"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotagoEd25519 "github.com/iotaledger/iota.go/v3/ed25519"
	"github.com/iotaledger/iota.go/v3/util"

	"github.com/iotaledger/iota.go/v3/remotesigner"
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
	ErrMilestoneInvalidMinSignatureThreshold = errors.New("min threshold must be at least 1")
	// ErrMilestoneNonApplicablePublicKey gets returned when a Milestone contains a public key which isn't in the applicable public key set.
	ErrMilestoneNonApplicablePublicKey = errors.New("non applicable public key found")
	// ErrMilestoneSignatureThresholdGreaterThanApplicablePublicKeySet gets returned when a min. signature threshold is greater than a given applicable public key set.
	ErrMilestoneSignatureThresholdGreaterThanApplicablePublicKeySet = errors.New("the min. signature threshold must be less or equal the applicable public key set")
	// ErrMilestoneInvalidSignature gets returned when a Milestone's signature is invalid.
	ErrMilestoneInvalidSignature = errors.New("invalid milestone signature")
	// ErrMilestoneInMemorySignerPrivateKeyMissing gets returned when an InMemoryEd25519MilestoneSigner is missing a private key.
	ErrMilestoneInMemorySignerPrivateKeyMissing = errors.New("private key missing")

	milestoneSupportedMsOptTypes = MilestoneOptTypeSet{MilestoneOptReceipt: struct{}{}, MilestoneOptProtocolParams: struct{}{}}
	milestoneOptsArrayRules      = serializer.ArrayRules{
		Min: 0,
		Max: 2,
		Guards: serializer.SerializableGuard{
			ReadGuard:  msOptReadGuard(milestoneSupportedMsOptTypes),
			WriteGuard: msOptWriteGuard(milestoneSupportedMsOptTypes),
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates | serializer.ArrayValidationModeLexicalOrdering,
	}

	// restrictions around parents within a Milestone.
	milestoneParentArrayRules = serializer.ArrayRules{
		Min:            BlockMinParents,
		Max:            BlockMaxParents,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates | serializer.ArrayValidationModeLexicalOrdering,
	}

	milestoneSupportedSigTypes = SignatureTypeSet{SignatureEd25519: struct{}{}}

	// restrictions around signatures within a Milestone.
	milestoneSignatureArrayRules = serializer.ArrayRules{
		Min: MinSignaturesInAMilestone,
		Max: MaxSignaturesInAMilestone,
		Guards: serializer.SerializableGuard{
			ReadGuard:  sigReadGuard(milestoneSupportedSigTypes),
			WriteGuard: sigWriteGuard(milestoneSupportedSigTypes),
		},
		UniquenessSliceFunc: func(next []byte) []byte { return next[:ed25519.PublicKeySize] },
		ValidationMode:      serializer.ArrayValidationModeNoDuplicates | serializer.ArrayValidationModeLexicalOrdering,
	}
)

// MilestoneParentArrayRules returns array rules defining the constraints on a slice of milestone parent references.
func MilestoneParentArrayRules() serializer.ArrayRules {
	return milestoneParentArrayRules
}

// MilestoneSignatureArrayRules returns array rules defining the constraints on a slice of signatures within a milestone.
func MilestoneSignatureArrayRules() serializer.ArrayRules {
	return milestoneSignatureArrayRules
}

type (
	// MilestonePublicKey is a public key within a Milestone.
	MilestonePublicKey = [MilestonePublicKeyLength]byte
	// MilestonePublicKeySet is a set of unique MilestonePublicKey.
	MilestonePublicKeySet = map[MilestonePublicKey]struct{}
	// MilestoneSignature is a signature within a Milestone.
	MilestoneSignature = [MilestoneSignatureLength]byte
	// MilestonePublicKeyMapping is a mapping from a public key to a private key.
	MilestonePublicKeyMapping = map[MilestonePublicKey]ed25519.PrivateKey
	// MilestoneMerkleProof is the merkle root within a milestone.
	MilestoneMerkleProof = [MilestoneMerkleProofLength]byte
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
func NewMilestone(index MilestoneIndex, timestamp uint32, protocolVersion byte, prevMsID MilestoneID, parents BlockIDs, inclMerkleProof MilestoneMerkleProof, appliedMerkleRoot MilestoneMerkleProof) *Milestone {
	return &Milestone{
		Index:               index,
		Timestamp:           timestamp,
		ProtocolVersion:     protocolVersion,
		PreviousMilestoneID: prevMsID,
		Parents:             parents,
		InclusionMerkleRoot: inclMerkleProof,
		AppliedMerkleRoot:   appliedMerkleRoot,
	}
}

// Milestone represents a special payload which defines the inclusion set of other blocks in the Tangle.
type Milestone struct {
	// The index of this milestone.
	Index MilestoneIndex
	// The time at which this milestone was issued.
	Timestamp uint32
	// The protocol version under which this milestone operates.
	ProtocolVersion byte
	// The pointer to the previous milestone.
	// Zeroed if there wasn't a previous milestone.
	PreviousMilestoneID MilestoneID
	// The parents where this milestone attaches to.
	Parents BlockIDs
	// The merkle root of all directly/indirectly referenced blocks (their IDs) which
	// were newly included by this milestone.
	InclusionMerkleRoot MilestoneMerkleProof
	// The merkle root of all blocks (their IDs) carrying ledger state mutating transactions.
	AppliedMerkleRoot MilestoneMerkleProof
	// The metadata associated with this milestone.
	Metadata []byte
	// The milestone options carried with this milestone.
	Opts MilestoneOpts
	// The signatures held by the milestone.
	Signatures Signatures
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
	essenceBytes, err := serializer.NewSerializer().
		WriteNum(m.Index, func(err error) error {
			return fmt.Errorf("unable to serialize milestone index for essence: %w", err)
		}).
		WriteNum(m.Timestamp, func(err error) error {
			return fmt.Errorf("unable to serialize milestone timestamp for essence: %w", err)
		}).
		WriteNum(m.ProtocolVersion, func(err error) error {
			return fmt.Errorf("unable to serialize milestone protocol version for essence: %w", err)
		}).
		WriteBytes(m.PreviousMilestoneID[:], func(err error) error {
			return fmt.Errorf("unable to serialize milestone last milestone ID for essence: %w", err)
		}).
		Write32BytesArraySlice(m.Parents.ToSerializerType(), serializer.DeSeriModePerformValidation, serializer.SeriLengthPrefixTypeAsByte, &milestoneParentArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize milestone parents for essence: %w", err)
		}).
		WriteBytes(m.InclusionMerkleRoot[:], func(err error) error {
			return fmt.Errorf("unable to serialize milestone inclusion merkle root for essence: %w", err)
		}).
		WriteBytes(m.AppliedMerkleRoot[:], func(err error) error {
			return fmt.Errorf("unable to serialize milestone applied merkle root for essence: %w", err)
		}).
		WriteVariableByteSlice(m.Metadata, serializer.SeriLengthPrefixTypeAsUint16, func(err error) error {
			return fmt.Errorf("unable to serialize milestone metadata for essence: %w", err)
		}, 0, MaxMetadataLength).
		WriteSliceOfObjects(&m.Opts, serializer.DeSeriModePerformValidation, nil, serializer.SeriLengthPrefixTypeAsByte, &milestoneOptsArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize milestone options for essence: %w", err)
		}).
		Serialize()
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
type MilestoneSigningFunc func(pubKeys []MilestonePublicKey, msEssence []byte) ([]MilestoneSignature, error)

// InMemoryEd25519MilestoneSigner is a function which uses the provided Ed25519 MilestonePublicKeyMapping to produce signatures for the Milestone essence data.
func InMemoryEd25519MilestoneSigner(prvKeys MilestonePublicKeyMapping) MilestoneSigningFunc {
	return func(pubKeys []MilestonePublicKey, msEssence []byte) ([]MilestoneSignature, error) {
		sigs := make([]MilestoneSignature, len(pubKeys))
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
	return func(pubKeys []MilestonePublicKey, msEssence []byte) ([]MilestoneSignature, error) {
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

		sigs64 := make([]MilestoneSignature, len(sigs))
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

	edSigs := make([]Signature, len(sigs))
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

func (m *Milestone) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	parentsSlice := serializer.SliceOfArraysOf32Bytes{}
	prevMsID := serializer.ArrayOf32Bytes{}
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(PayloadMilestone), serializer.TypeDenotationUint32, func(err error) error {
			return fmt.Errorf("unable to deserialize milestone: %w", err)
		}).
		ReadNum(&m.Index, func(err error) error {
			return fmt.Errorf("unable to deserialize milestone index: %w", err)
		}).
		ReadNum(&m.Timestamp, func(err error) error {
			return fmt.Errorf("unable to deserialize milestone timestamp: %w", err)
		}).
		ReadNum(&m.ProtocolVersion, func(err error) error {
			return fmt.Errorf("unable to deserialize milestone protocol version: %w", err)
		}).
		ReadArrayOf32Bytes(&prevMsID, func(err error) error {
			return fmt.Errorf("unable to deserialize milestone last milestone ID: %w", err)
		}).
		Do(func() {
			copy(m.PreviousMilestoneID[:], prevMsID[:])
		}).
		ReadSliceOfArraysOf32Bytes(&parentsSlice, deSeriMode, serializer.SeriLengthPrefixTypeAsByte, &milestoneParentArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize milestone parents: %w", err)
		}).
		Do(func() {
			m.Parents = make(BlockIDs, len(parentsSlice))
			for i, ele := range parentsSlice {
				m.Parents[i] = ele
			}
		}).
		ReadArrayOf32Bytes(&m.InclusionMerkleRoot, func(err error) error {
			return fmt.Errorf("unable to deserialize milestone inclusion merkle root: %w", err)
		}).
		ReadArrayOf32Bytes(&m.AppliedMerkleRoot, func(err error) error {
			return fmt.Errorf("unable to deserialize milestone applied merkle root: %w", err)
		}).
		ReadVariableByteSlice(&m.Metadata, serializer.SeriLengthPrefixTypeAsUint16, func(err error) error {
			return fmt.Errorf("unable to deserialize milestone metadata: %w", err)
		}, 0, MaxMetadataLength).
		ReadSliceOfObjects(&m.Opts, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationByte, &milestoneOptsArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize milestone options: %w", err)
		}).
		ReadSliceOfObjects(&m.Signatures, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationByte, &milestoneSignatureArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize milestone signatures: %w", err)
		}).
		Done()
}

func (m *Milestone) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(PayloadMilestone, func(err error) error {
			return fmt.Errorf("unable to serialize milestone payload ID: %w", err)
		}).
		WriteNum(m.Index, func(err error) error {
			return fmt.Errorf("unable to serialize milestone index: %w", err)
		}).
		WriteNum(m.Timestamp, func(err error) error {
			return fmt.Errorf("unable to serialize milestone timestamp: %w", err)
		}).
		WriteNum(m.ProtocolVersion, func(err error) error {
			return fmt.Errorf("unable to serialize milestone protocol version: %w", err)
		}).
		WriteBytes(m.PreviousMilestoneID[:], func(err error) error {
			return fmt.Errorf("unable to serialize milestone last milestone ID for essence: %w", err)
		}).
		Write32BytesArraySlice(m.Parents.ToSerializerType(), deSeriMode, serializer.SeriLengthPrefixTypeAsByte, &milestoneParentArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize milestone parents: %w", err)
		}).
		WriteBytes(m.InclusionMerkleRoot[:], func(err error) error {
			return fmt.Errorf("unable to serialize milestone inclusion merkle root: %w", err)
		}).
		WriteBytes(m.AppliedMerkleRoot[:], func(err error) error {
			return fmt.Errorf("unable to serialize milestone applied merkle root: %w", err)
		}).
		WriteVariableByteSlice(m.Metadata, serializer.SeriLengthPrefixTypeAsUint16, func(err error) error {
			return fmt.Errorf("unable to serialize milestone metadata: %w", err)
		}, 0, MaxMetadataLength).
		WriteSliceOfObjects(&m.Opts, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, &milestoneOptsArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize milestone options: %w", err)
		}).
		WriteSliceOfObjects(&m.Signatures, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, &milestoneSignatureArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize milestone signatures: %w", err)
		}).
		Serialize()
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

func (m *Milestone) MarshalJSON() ([]byte, error) {
	jMilestone := &jsonMilestone{}
	jMilestone.Type = int(PayloadMilestone)
	jMilestone.Index = int(m.Index)
	jMilestone.Timestamp = int(m.Timestamp)
	jMilestone.ProtocolVersion = int(m.ProtocolVersion)
	jMilestone.PreviousMilestoneID = EncodeHex(m.PreviousMilestoneID[:])
	jMilestone.Parents = make([]string, len(m.Parents))
	for i, parent := range m.Parents {
		jMilestone.Parents[i] = EncodeHex(parent[:])
	}
	jMilestone.InclusionMerkleRoot = EncodeHex(m.InclusionMerkleRoot[:])
	jMilestone.AppliedMerkleRoot = EncodeHex(m.AppliedMerkleRoot[:])

	jMilestone.Opts = make([]*json.RawMessage, len(m.Opts))
	for i, opt := range m.Opts {
		jsonOpt, err := opt.MarshalJSON()
		if err != nil {
			return nil, err
		}
		rawJsonSig := json.RawMessage(jsonOpt)
		jMilestone.Opts[i] = &rawJsonSig
	}

	jMilestone.Metadata = EncodeHex(m.Metadata[:])

	jMilestone.Signatures = make([]*json.RawMessage, len(m.Signatures))
	for i, sig := range m.Signatures {
		jsonSig, err := sig.MarshalJSON()
		if err != nil {
			return nil, err
		}
		rawJsonSig := json.RawMessage(jsonSig)
		jMilestone.Signatures[i] = &rawJsonSig
	}

	return json.Marshal(jMilestone)
}

func (m *Milestone) UnmarshalJSON(bytes []byte) error {
	jMilestone := &jsonMilestone{}
	if err := json.Unmarshal(bytes, jMilestone); err != nil {
		return err
	}
	seri, err := jMilestone.ToSerializable()
	if err != nil {
		return err
	}
	*m = *seri.(*Milestone)
	return nil
}

// jsonMilestone defines the json representation of a Milestone.
type jsonMilestone struct {
	Type                int                `json:"type"`
	Index               int                `json:"index"`
	Timestamp           int                `json:"timestamp"`
	ProtocolVersion     int                `json:"protocolVersion"`
	PreviousMilestoneID string             `json:"previousMilestoneId"`
	Parents             []string           `json:"parents"`
	InclusionMerkleRoot string             `json:"inclusionMerkleRoot"`
	AppliedMerkleRoot   string             `json:"appliedMerkleRoot"`
	Metadata            string             `json:"metadata,omitempty"`
	Opts                []*json.RawMessage `json:"options,omitempty"`
	Signatures          []*json.RawMessage `json:"signatures"`
}

func (j *jsonMilestone) ToSerializable() (serializer.Serializable, error) {
	var err error

	payload := &Milestone{}
	payload.Index = MilestoneIndex(j.Index)
	payload.Timestamp = uint32(j.Timestamp)
	payload.ProtocolVersion = byte(j.ProtocolVersion)
	prevMsID, err := DecodeHex(j.PreviousMilestoneID)
	if err != nil {
		return nil, fmt.Errorf("unable to decode milestone last milestone ID from JSON: %w", err)
	}
	copy(payload.PreviousMilestoneID[:], prevMsID)

	payload.Parents = make(BlockIDs, len(j.Parents))
	for i, jParent := range j.Parents {
		parentBytes, err := DecodeHex(jParent)
		if err != nil {
			return nil, fmt.Errorf("unable to decode parent %d from JSON: %w", i+1, err)
		}
		copy(payload.Parents[i][:], parentBytes)
	}

	inclusionMerkleRoot, err := DecodeHex(j.InclusionMerkleRoot)
	if err != nil {
		return nil, fmt.Errorf("unable to decode inclusion merkle root from JSON: %w", err)
	}
	copy(payload.InclusionMerkleRoot[:], inclusionMerkleRoot)

	appliedMerkleRoot, err := DecodeHex(j.AppliedMerkleRoot)
	if err != nil {
		return nil, fmt.Errorf("unable to decode applied merkle root from JSON: %w", err)
	}
	copy(payload.AppliedMerkleRoot[:], appliedMerkleRoot)

	payload.Metadata, err = DecodeHex(j.Metadata)
	if err != nil {
		return nil, err
	}

	payload.Opts, err = milestoneOptsFromJSONRawMsg(j.Opts)
	if err != nil {
		return nil, err
	}

	payload.Signatures, err = signaturesFromJSONRawMsg(j.Signatures)
	if err != nil {
		return nil, err
	}

	return payload, nil
}
