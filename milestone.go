package iotago

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iotaledger/hive.go/serializer/v2"
	iotagoEd25519 "github.com/iotaledger/iota.go/v3/ed25519"
	"github.com/iotaledger/iota.go/v3/util"
	"golang.org/x/crypto/blake2b"
	"google.golang.org/grpc"
	"sort"

	"github.com/iotaledger/iota.go/v3/remotesigner"
)

const (
	// MilestoneInclusionMerkleProofLength defines the length of the inclusion merkle proof within a milestone payload.
	MilestoneInclusionMerkleProofLength = blake2b.Size256
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
	// ErrMilestoneInvalidMinPoWScoreValues gets returned when the min. PoW score fields are invalid.
	ErrMilestoneInvalidMinPoWScoreValues = fmt.Errorf("invalid milestone min pow score values")

	milestonePayloadGuard = serializer.SerializableGuard{
		ReadGuard: func(ty uint32) (serializer.Serializable, error) {
			if PayloadType(ty) != PayloadReceipt {
				return nil, ErrTypeIsNotSupportedPayload
			}
			return PayloadSelector(ty)
		},
		WriteGuard: func(seri serializer.Serializable) error {
			if seri == nil {
				return nil
			}
			if _, is := seri.(*Receipt); !is {
				return ErrTypeIsNotSupportedPayload
			}
			return nil
		},
	}

	// restrictions around parents within a Milestone.
	milestoneParentArrayRules = serializer.ArrayRules{
		Min:            MinParentsInAMessage,
		Max:            MaxParentsInAMessage,
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
	// MilestoneID is the ID of a Milestone.
	MilestoneID [MilestoneIDLength]byte
	// MilestonePublicKey is a public key within a Milestone.
	MilestonePublicKey = [MilestonePublicKeyLength]byte
	// MilestonePublicKeySet is a set of unique MilestonePublicKey.
	MilestonePublicKeySet = map[MilestonePublicKey]struct{}
	// MilestoneSignature is a signature within a Milestone.
	MilestoneSignature = [MilestoneSignatureLength]byte
	// MilestonePublicKeyMapping is a mapping from a public key to a private key.
	MilestonePublicKeyMapping = map[MilestonePublicKey]ed25519.PrivateKey
	// MilestoneParentMessageID is a reference to a parent message.
	MilestoneParentMessageID = MessageID
	// MilestoneParentMessageIDs are references to parent messages.
	MilestoneParentMessageIDs = []MilestoneParentMessageID
	// MilestoneInclusionMerkleProof is the inclusion merkle proof data of a milestone.
	MilestoneInclusionMerkleProof = [MilestoneInclusionMerkleProofLength]byte
)

// ToHex converts the MilestoneID to its hex representation.
func (milestoneID MilestoneID) ToHex() string {
	return EncodeHex(milestoneID[:])
}

// NewMilestone creates a new unsigned Milestone.
func NewMilestone(index uint32, timestamp uint64, parents MilestoneParentMessageIDs, inclMerkleProof MilestoneInclusionMerkleProof) (*Milestone, error) {
	ms := &Milestone{
		Index:                index,
		Timestamp:            timestamp,
		Parents:              parents,
		InclusionMerkleProof: inclMerkleProof,
	}
	return ms, nil
}

// Milestone represents a special payload which defines the inclusion set of other messages in the Tangle.
type Milestone struct {
	// The index of this milestone.
	Index uint32
	// The time at which this milestone was issued.
	Timestamp uint64
	// The parents where this milestone attaches to.
	Parents MilestoneParentMessageIDs
	// The inclusion merkle proof of included/newly confirmed transaction IDs.
	InclusionMerkleProof MilestoneInclusionMerkleProof
	// The next minimum PoW score to use after NextPoWScoreMilestoneIndex is hit.
	NextPoWScore uint32
	// The milestone index at which the PoW score changes to NextPoWScore.
	NextPoWScoreMilestoneIndex uint32
	// The inner payload of the milestone. Can be nil or a Receipt.
	Receipt serializer.Serializable
	// The signatures held by the milestone.
	Signatures Signatures
}

func (m *Milestone) PayloadType() PayloadType {
	return PayloadMilestone
}

// ID computes the ID of the Milestone.
func (m *Milestone) ID() (*MilestoneID, error) {
	data, err := m.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return nil, fmt.Errorf("can't compute milestone payload ID: %w", err)
	}
	h := blake2b.Sum256(data)
	mID := &MilestoneID{}
	copy(mID[:], h[:])
	return mID, nil
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
		Write32BytesArraySlice(m.Parents, serializer.DeSeriModePerformValidation, serializer.SeriLengthPrefixTypeAsByte, &milestoneParentArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize milestone parents for essence: %w", err)
		}).
		WriteBytes(m.InclusionMerkleProof[:], func(err error) error {
			return fmt.Errorf("unable to serialize milestone inclusion merkle proof for essence: %w", err)
		}).
		WriteNum(m.NextPoWScore, func(err error) error {
			return fmt.Errorf("unable to serialize milestone next pow score for essence: %w", err)
		}).
		WriteNum(m.NextPoWScoreMilestoneIndex, func(err error) error {
			return fmt.Errorf("unable to serialize milestone next pow score milestone index for essence: %w", err)
		}).
		WritePayload(m.Receipt, serializer.DeSeriModePerformValidation, nil, milestonePayloadGuard.WriteGuard, func(err error) error {
			return fmt.Errorf("unable to serialize milestone receipt for essence: %w", err)
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
		conn, err := grpc.Dial(remoteEndpoint, grpc.WithInsecure())
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

// Sign produces the signatures with the given envelope message and updates the Signatures field of the Milestone
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
		ReadSliceOfArraysOf32Bytes(&m.Parents, deSeriMode, serializer.SeriLengthPrefixTypeAsByte, &milestoneParentArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize milestone parents: %w", err)
		}).
		ReadArrayOf32Bytes(&m.InclusionMerkleProof, func(err error) error {
			return fmt.Errorf("unable to deserialize milestone inclusion merkle proof: %w", err)
		}).
		ReadNum(&m.NextPoWScore, func(err error) error {
			return fmt.Errorf("unable to deserialize milestone next pow score: %w", err)
		}).
		ReadNum(&m.NextPoWScoreMilestoneIndex, func(err error) error {
			return fmt.Errorf("unable to deserialize milestone next pow score milestone index: %w", err)
		}).
		WithValidation(deSeriMode, func(_ []byte, err error) error {
			switch {
			case m.NextPoWScore != 0 && m.NextPoWScoreMilestoneIndex == 0:
				return fmt.Errorf("%w: next-pow-score-milestone-index is zero but next-pow-score is not", ErrMilestoneInvalidMinPoWScoreValues)
			}
			return nil
		}).
		ReadPayload(&m.Receipt, deSeriMode, deSeriCtx, milestonePayloadGuard.ReadGuard, func(err error) error {
			return fmt.Errorf("unable to deserialize milestone receipt: %w", err)
		}).
		ReadSliceOfObjects(&m.Signatures, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationByte, &milestoneSignatureArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize milestone signatures: %w", err)
		}).
		Done()
}

func (m *Milestone) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WithValidation(deSeriMode, func(_ []byte, err error) error {
			switch {
			case m.NextPoWScore != 0 && m.NextPoWScoreMilestoneIndex == 0:
				return fmt.Errorf("%w: next-pow-score-milestone-index is zero but next-pow-score is not", ErrMilestoneInvalidMinPoWScoreValues)
			}
			return nil
		}).
		WriteNum(PayloadMilestone, func(err error) error {
			return fmt.Errorf("unable to serialize milestone payload ID: %w", err)
		}).
		WriteNum(m.Index, func(err error) error {
			return fmt.Errorf("unable to serialize milestone index: %w", err)
		}).
		WriteNum(m.Timestamp, func(err error) error {
			return fmt.Errorf("unable to serialize milestone timestamp: %w", err)
		}).
		Write32BytesArraySlice(m.Parents, deSeriMode, serializer.SeriLengthPrefixTypeAsByte, &milestoneParentArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize milestone parents: %w", err)
		}).
		WriteBytes(m.InclusionMerkleProof[:], func(err error) error {
			return fmt.Errorf("unable to serialize milestone inclusion merkle proof: %w", err)
		}).
		WriteNum(m.NextPoWScore, func(err error) error {
			return fmt.Errorf("unable to serialize milestone next pow score: %w", err)
		}).
		WriteNum(m.NextPoWScoreMilestoneIndex, func(err error) error {
			return fmt.Errorf("unable to serialize milestone next pow score milestone index: %w", err)
		}).
		WritePayload(m.Receipt, deSeriMode, deSeriCtx, milestonePayloadGuard.WriteGuard, func(err error) error {
			return fmt.Errorf("unable to serialize milestone receipt: %w", err)
		}).
		WriteSliceOfObjects(&m.Signatures, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, &milestoneSignatureArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize milestone signatures: %w", err)
		}).
		Serialize()
}

func (m *Milestone) Size() int {
	// 1 byte for length prefixes
	parentMessagesByteLen := serializer.OneByte + MessageIDLength*len(m.Parents)
	signatureByteLen := serializer.OneByte + ((&Ed25519Signature{}).Size())*len(m.Signatures)

	// in theory `Size()` should not serialize, just return the size - but Milestone.Size() probably won't be used, so it should be fine to be non-optimal
	receiptBytesLen := serializer.UInt32ByteSize // 4 bytes for length prefix if receipt is nil
	if m.Receipt != nil {
		receiptBytes, err := m.Receipt.Serialize(serializer.DeSeriModeNoValidation, ZeroRentParas)
		if err != nil {
			return 0
		}
		receiptBytesLen = len(receiptBytes)
	}
	return util.NumByteLen(uint32(PayloadMilestone)) + util.NumByteLen(m.Index) + util.NumByteLen(m.Timestamp) +
		parentMessagesByteLen + MilestoneInclusionMerkleProofLength + util.NumByteLen(m.NextPoWScore) +
		util.NumByteLen(m.NextPoWScoreMilestoneIndex) + receiptBytesLen + signatureByteLen
}

func (m *Milestone) MarshalJSON() ([]byte, error) {
	jMilestone := &jsonMilestone{}
	jMilestone.Type = int(PayloadMilestone)
	jMilestone.Index = int(m.Index)
	jMilestone.Timestamp = int(m.Timestamp)
	jMilestone.Parents = make([]string, len(m.Parents))
	for i, parent := range m.Parents {
		jMilestone.Parents[i] = EncodeHex(parent[:])
	}
	jMilestone.InclusionMerkleProof = EncodeHex(m.InclusionMerkleProof[:])
	jMilestone.NextPoWScore = int(m.NextPoWScore)
	jMilestone.NextPoWScoreMilestoneIndex = int(m.NextPoWScoreMilestoneIndex)

	if m.Receipt != nil {
		jsonReceipt, err := m.Receipt.MarshalJSON()
		if err != nil {
			return nil, err
		}
		rawReceiptJsonPayload := json.RawMessage(jsonReceipt)
		jMilestone.Receipt = &rawReceiptJsonPayload
	}

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
	Type                       int                `json:"type"`
	Index                      int                `json:"index"`
	Timestamp                  int                `json:"timestamp"`
	Parents                    []string           `json:"parentMessageIds"`
	InclusionMerkleProof       string             `json:"inclusionMerkleProof"`
	NextPoWScore               int                `json:"nextPoWScore"`
	NextPoWScoreMilestoneIndex int                `json:"nextPoWScoreMilestoneIndex"`
	Receipt                    *json.RawMessage   `json:"receipt,omitempty"`
	Signatures                 []*json.RawMessage `json:"signatures"`
}

func (j *jsonMilestone) ToSerializable() (serializer.Serializable, error) {
	var err error

	payload := &Milestone{}
	payload.Index = uint32(j.Index)
	payload.Timestamp = uint64(j.Timestamp)

	payload.Parents = make(MilestoneParentMessageIDs, len(j.Parents))
	for i, jparent := range j.Parents {
		parentBytes, err := DecodeHex(jparent)
		if err != nil {
			return nil, fmt.Errorf("unable to decode parent %d from JSON for milestone payload: %w", i+1, err)
		}
		copy(payload.Parents[i][:], parentBytes)
	}

	inclusionMerkleProofBytes, err := DecodeHex(j.InclusionMerkleProof)
	if err != nil {
		return nil, fmt.Errorf("unable to decode inlcusion merkle proof from JSON for milestone payload: %w", err)
	}
	copy(payload.InclusionMerkleProof[:], inclusionMerkleProofBytes)

	payload.NextPoWScore = uint32(j.NextPoWScore)
	payload.NextPoWScoreMilestoneIndex = uint32(j.NextPoWScoreMilestoneIndex)

	if j.Receipt != nil {
		jsonPayload, err := DeserializeObjectFromJSON(j.Receipt, func(ty int) (JSONSerializable, error) {
			return &jsonReceipt{}, nil
		})
		if err != nil {
			return nil, err
		}

		payload.Receipt, err = jsonPayload.ToSerializable()
		if err != nil {
			return nil, fmt.Errorf("unable to decode inner milestone receipt: %w", err)
		}
	}

	payload.Signatures, err = signaturesFromJSONRawMsg(j.Signatures)
	if err != nil {
		return nil, err
	}

	return payload, nil
}
