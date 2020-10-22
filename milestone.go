package iota

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"golang.org/x/crypto/blake2b"
)

const (
	// Defines the milestone payload's ID.
	MilestonePayloadTypeID uint32 = 1
	// Defines the length of the inclusion merkle proof within a milestone payload.
	MilestoneInclusionMerkleProofLength = blake2b.Size
	// Defines the length of the milestone signature.
	MilestoneSignatureLength = ed25519.SignatureSize
	// Defines the length of a Milestone ID.
	MilestoneIDLength = blake2b.Size256
	// Defines the length of a public key within a milestone.
	MilestonePublicKeyLength = ed25519.PublicKeySize
	// Defines the serialized size of a milestone payload.
	// payload type+index+timestamp+parent1+parent2+inclusion-merkle-proof+pubkeys-length+pubkey+signature
	MilestoneBinSerializedMinSize = TypeDenotationByteSize + UInt32ByteSize + UInt64ByteSize + MessageIDLength + MessageIDLength +
		MilestoneInclusionMerkleProofLength + OneByte + ed25519.PublicKeySize + OneByte + MilestoneSignatureLength
	// MaxSignaturesInAMilestone is the maximum amount of signatures in a milestone.
	MaxSignaturesInAMilestone = 255
	// MinSignaturesInAMilestone is the minimum amount of signatures in a milestone.
	MinSignaturesInAMilestone = 1
	// MaxPublicKeysInAMilestone is the maximum amount of public keys in a milestone.
	MaxPublicKeysInAMilestone = 255
	// MinPublicKeysInAMilestone is the minimum amount of public keys in a milestone.
	MinPublicKeysInAMilestone = 1
)

var (
	// Returned if a to be deserialized Milestone does not contain at least one signature.
	ErrMilestoneTooFewSignatures = errors.New("a milestone must hold at least one signature")
	// Returned if there are less signatures within a Milestone than the min. threshold.
	ErrMilestoneTooFewSignaturesForVerificationThreshold = errors.New("too few signatures for verification")
	// Returned if a to be deserialized Milestone does not contain at least one public key.
	ErrMilestoneTooFewPublicKeys = errors.New("a milestone must hold at least one public key")
	// Returned if the Milestone public keys are not in lexical order when serialized.
	ErrMilestonePublicKeyOrderViolatesLexicalOrder = errors.New("public keys must be in their lexical order (byte wise) when serialized")
	// Returned when a MilestoneSigningFunc produces less signatures than expected.
	ErrMilestoneProducedSignaturesCountMismatch = errors.New("produced and wanted signature count mismatch")
	// Returned when the count of signatures and public keys within a Milestone don't match.
	ErrMilestoneSignaturesPublicKeyCountMismatch = errors.New("milestone signatures and public keys count must be equal")
	// Returned when a Milestone holds more than 255 signatures.
	ErrMilestoneTooManySignatures = fmt.Errorf("a milestone can hold max %d signatures", MaxSignaturesInAMilestone)
	// Returned when a Milestone holds more than 255 public keys.
	ErrMilestoneTooManyPublicKeys = fmt.Errorf("a milestone can hold max %d public keys", MaxPublicKeysInAMilestone)
	// Returned when an invalid min signatures threshold is given the the verification function.
	ErrMilestoneInvalidMinSignatureThreshold = fmt.Errorf("min threshold must be at least 1")
	// Returned when the same public key occur within a Milestone.
	ErrMilestoneDupedPublicKeys = fmt.Errorf("milestone public keys must be unique")
	// Returned when a Milestone contains a public key which isn't in the applicable public key set.
	ErrMilestoneNonApplicablePublicKey = fmt.Errorf("non applicable public key found")
	// Returned when a min. signature threshold is greater than a given applicable public key set.
	ErrMilestoneSignatureThresholdGreaterThanApplicablePublicKeySet = fmt.Errorf("the min. signature threshold must be less or equal the applicable public key set")
	// Returned when a Milestone's signature is invalid.
	ErrMilestoneInvalidSignature = fmt.Errorf("invalid milestone signature")
	// Returned when a InMemoryEd25519MilestoneSigner is missing a private key.
	ErrMilestoneInMemorySignerPrivateKeyMissing = fmt.Errorf("private key missing")

	// restrictions around public keys within a Milestone.
	milestonePublicKeyArrayRules = ArrayRules{
		ElementBytesLexicalOrderErr: ErrMilestonePublicKeyOrderViolatesLexicalOrder,
	}
)

type (
	// MilestoneID is the ID of a Milestone.
	MilestoneID = [MilestoneIDLength]byte
	// MilestonePublicKey is a public key within a Milestone.
	MilestonePublicKey = [MilestonePublicKeyLength]byte
	// MilestonePublicKeySet is a set of unique MilestonePublicKey.
	MilestonePublicKeySet = map[MilestonePublicKey]struct{}
	// MilestoneSignature is a signature within a Milestone.
	MilestoneSignature = [MilestoneSignatureLength]byte
	// MilestonePublicKeyMapping is a mapping from a public key to a private key.
	MilestonePublicKeyMapping = map[MilestonePublicKey]ed25519.PrivateKey
)

// Milestone represents a special payload which defines the inclusion set of other messages in the Tangle.
type Milestone struct {
	// The index of this milestone.
	Index uint32
	// The time at which this milestone was issued.
	Timestamp uint64
	// The 1st parent where this milestone attaches to.
	Parent1 [MessageIDLength]byte
	// The 2nd parent where this milestone attaches to.
	Parent2 [MessageIDLength]byte
	// The inclusion merkle proof of included/newly confirmed transaction IDs.
	InclusionMerkleProof [MilestoneInclusionMerkleProofLength]byte
	// The public keys validating the signatures of the milestone.
	PublicKeys []MilestonePublicKey
	// The signatures held by the milestone.
	Signatures []MilestoneSignature
}

// ID computes the ID of the Milestone.
func (m *Milestone) ID() (*MilestoneID, error) {
	data, err := m.Serialize(DeSeriModeNoValidation)
	if err != nil {
		return nil, fmt.Errorf("can't compute milestone payload ID: %w", err)
	}
	h := blake2b.Sum256(data)
	return &h, nil
}

// Essence returns the essence bytes (the bytes to be signed) of the Milestone.
func (m *Milestone) Essence() ([]byte, error) {
	if len(m.PublicKeys) < MinPublicKeysInAMilestone {
		return nil, fmt.Errorf("unable to serialize milestone as essence: %w", ErrMilestoneTooFewPublicKeys)

	}

	return NewSerializer().
		WriteNum(m.Index, func(err error) error {
			return fmt.Errorf("unable to serialize milestone as essence: %w", err)
		}).
		WriteNum(m.Timestamp, func(err error) error {
			return fmt.Errorf("unable to serialize milestone as essence: %w", err)
		}).
		WriteBytes(m.Parent1[:], func(err error) error {
			return fmt.Errorf("unable to serialize milestone parent 1 for essence: %w", err)
		}).
		WriteBytes(m.Parent2[:], func(err error) error {
			return fmt.Errorf("unable to serialize milestone parent 2 for essence: %w", err)
		}).
		WriteBytes(m.InclusionMerkleProof[:], func(err error) error {
			return fmt.Errorf("unable to serialize milestone inclusion merkle proof for essence: %w", err)
		}).
		Write32BytesArraySlice(m.PublicKeys, SeriSliceLengthAsByte, func(err error) error {
			return fmt.Errorf("unable to serialize milestone public keys for essence: %w", err)
		}).
		Serialize()
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
	case len(m.Signatures) != len(m.PublicKeys):
		return ErrMilestoneSignaturesPublicKeyCountMismatch
	case len(m.Signatures) < minSigThreshold:
		return fmt.Errorf("%w: wanted min. %d but only had %d", ErrMilestoneTooFewSignaturesForVerificationThreshold, minSigThreshold, len(m.Signatures))
	case len(applicablePubKeys) < minSigThreshold:
		return ErrMilestoneSignatureThresholdGreaterThanApplicablePublicKeySet
	}

	msEssence, err := m.Essence()
	if err != nil {
		return fmt.Errorf("unable to compute milestone essence for signature verification: %w", err)
	}

	for msPubKeyIndex, msPubKey := range m.PublicKeys {
		if _, has := applicablePubKeys[msPubKey]; !has {
			return fmt.Errorf("%w: public key %s is not applicable", ErrMilestoneNonApplicablePublicKey, hex.EncodeToString(msPubKey[:]))
		}

		if ok := ed25519.Verify(msPubKey[:], msEssence[:], m.Signatures[msPubKeyIndex][:]); !ok {
			return fmt.Errorf("%w: at index %d, checked against public key %s", ErrMilestoneInvalidSignature, msPubKeyIndex, hex.EncodeToString(msPubKey[:]))
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
				return nil, fmt.Errorf("%w: needed for public key %s", ErrMilestoneInMemorySignerPrivateKeyMissing, hex.EncodeToString(pubKey[:]))
			}
			sig := ed25519.Sign(prvKey, msEssence)
			copy(sigs[i][:], sig)
		}
		return sigs, nil
	}
}

// Sign produces the signatures with the given envelope message and updates the Signatures field of the Milestone
// with the resulting signatures of the given MilestoneSigningFunc.
func (m *Milestone) Sign(signingFunc MilestoneSigningFunc) error {
	msEssence, err := m.Essence()
	if err != nil {
		return fmt.Errorf("unable to compute milestone essence for signing: %w", err)
	}

	sigs, err := signingFunc(m.PublicKeys, msEssence)
	if err != nil {
		return fmt.Errorf("unable to produce milestone signatures: %w", err)
	}

	switch {
	case len(m.PublicKeys) != len(sigs):
		return fmt.Errorf("%w: wanted %d signatures but only produced %d", ErrMilestoneProducedSignaturesCountMismatch, len(m.PublicKeys), len(sigs))
	case len(sigs) < MinSignaturesInAMilestone:
		return fmt.Errorf("%w: not enough signatures were produced during signing", ErrMilestoneTooFewSignatures)
	case len(sigs) > MaxSignaturesInAMilestone:
		return fmt.Errorf("%w: too many signatures were produced during signing", ErrMilestoneTooManySignatures)
	}

	m.Signatures = sigs
	return nil
}

func (m *Milestone) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkMinByteLength(MilestoneBinSerializedMinSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid milestone bytes: %w", err)
		}
		if err := checkType(data, MilestonePayloadTypeID); err != nil {
			return 0, fmt.Errorf("unable to deserialize milestone: %w", err)
		}
	}

	bytesRead, err := NewDeserializer(data).
		Skip(TypeDenotationByteSize, func(err error) error {
			return fmt.Errorf("unable to skip milestone payload ID during deserialization: %w", err)
		}).
		ReadNum(&m.Index, func(err error) error {
			return fmt.Errorf("unable to deserialize milestone index: %w", err)
		}).
		ReadNum(&m.Timestamp, func(err error) error {
			return fmt.Errorf("unable to deserialize milestone timestamp: %w", err)
		}).
		ReadArrayOf32Bytes(&m.Parent1, func(err error) error {
			return fmt.Errorf("unable to deserialize milestone parent 1: %w", err)
		}).
		ReadArrayOf32Bytes(&m.Parent2, func(err error) error {
			return fmt.Errorf("unable to deserialize milestone parent 2: %w", err)
		}).
		ReadArrayOf64Bytes(&m.InclusionMerkleProof, func(err error) error {
			return fmt.Errorf("unable to deserialize milestone inclusion merkle proof: %w", err)
		}).
		ReadSliceOfArraysOf32Bytes(&m.PublicKeys, SeriSliceLengthAsByte, func(err error) error {
			return fmt.Errorf("unable to deserialize milestone public keys: %w", err)
		}).
		ReadSliceOfArraysOf64Bytes(&m.Signatures, SeriSliceLengthAsByte, func(err error) error {
			return fmt.Errorf("unable to deserialize milestone public keys: %w", err)
		}).
		Done()

	if err != nil {
		return bytesRead, err
	}

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		pubKeyLexicalOrderValidator := milestonePublicKeyArrayRules.LexicalOrderValidator()
		for i := range m.PublicKeys {
			if err := pubKeyLexicalOrderValidator(i, m.PublicKeys[i][:]); err != nil {
				return 0, err
			}
		}
	}

	return bytesRead, nil
}

func (m *Milestone) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		pubKeyLexicalOrderValidator := milestonePublicKeyArrayRules.LexicalOrderValidator()
		for i := range m.PublicKeys {
			if err := pubKeyLexicalOrderValidator(i, m.PublicKeys[i][:]); err != nil {
				return nil, err
			}
		}

		switch {
		case len(m.PublicKeys) > MaxPublicKeysInAMilestone:
			return nil, fmt.Errorf("unable to serialize milestone: %w", ErrMilestoneTooManyPublicKeys)
		case len(m.PublicKeys) < MinPublicKeysInAMilestone:
			return nil, fmt.Errorf("unable to serialize milestone: %w", ErrMilestoneTooFewPublicKeys)
		case len(m.Signatures) > MaxSignaturesInAMilestone:
			return nil, fmt.Errorf("unable to serialize milestone: %w", ErrMilestoneTooManySignatures)
		case len(m.Signatures) < MinSignaturesInAMilestone:
			return nil, fmt.Errorf("unable to serialize milestone: %w", ErrMilestoneTooFewSignatures)
		}
	}

	return NewSerializer().
		WriteNum(MilestonePayloadTypeID, func(err error) error {
			return fmt.Errorf("unable to serialize milestone payload ID: %w", err)
		}).
		WriteNum(m.Index, func(err error) error {
			return fmt.Errorf("unable to serialize milestone index: %w", err)
		}).
		WriteNum(m.Timestamp, func(err error) error {
			return fmt.Errorf("unable to serialize milestone timestamp: %w", err)
		}).
		WriteBytes(m.Parent1[:], func(err error) error {
			return fmt.Errorf("unable to serialize milestone parent 1: %w", err)
		}).
		WriteBytes(m.Parent2[:], func(err error) error {
			return fmt.Errorf("unable to serialize milestone parent 2: %w", err)
		}).
		WriteBytes(m.InclusionMerkleProof[:], func(err error) error {
			return fmt.Errorf("unable to serialize milestone inclusion merkle proof: %w", err)
		}).
		Write32BytesArraySlice(m.PublicKeys, SeriSliceLengthAsByte, func(err error) error {
			return fmt.Errorf("unable to serialize milestone public keys: %w", err)
		}).
		Write64BytesArraySlice(m.Signatures, SeriSliceLengthAsByte, func(err error) error {
			return fmt.Errorf("unable to serialize milestone signatures: %w", err)
		}).
		Serialize()
}

func (m *Milestone) MarshalJSON() ([]byte, error) {
	jsonMilestonePayload := &jsonmilestonepayload{}
	jsonMilestonePayload.Type = int(MilestonePayloadTypeID)
	jsonMilestonePayload.Index = int(m.Index)
	jsonMilestonePayload.Signatures = make([]string, len(m.Signatures))
	for i, sig := range m.Signatures {
		jsonMilestonePayload.Signatures[i] = hex.EncodeToString(sig[:])
	}
	jsonMilestonePayload.Timestamp = int(m.Timestamp)
	jsonMilestonePayload.InclusionMerkleProof = hex.EncodeToString(m.InclusionMerkleProof[:])
	return json.Marshal(jsonMilestonePayload)
}

func (m *Milestone) UnmarshalJSON(bytes []byte) error {
	jsonMilestonePayload := &jsonmilestonepayload{}
	if err := json.Unmarshal(bytes, jsonMilestonePayload); err != nil {
		return err
	}
	seri, err := jsonMilestonePayload.ToSerializable()
	if err != nil {
		return err
	}
	*m = *seri.(*Milestone)
	return nil
}

// jsonmilestonepayload defines the json representation of a Milestone.
type jsonmilestonepayload struct {
	Type                 int      `json:"type"`
	Index                int      `json:"index"`
	Timestamp            int      `json:"timestamp"`
	InclusionMerkleProof string   `json:"inclusionMerkleProof"`
	Signatures           []string `json:"signatures"`
}

func (j *jsonmilestonepayload) ToSerializable() (Serializable, error) {
	inclusionMerkleProofBytes, err := hex.DecodeString(j.InclusionMerkleProof)
	if err != nil {
		return nil, fmt.Errorf("unable to decode inlcusion merkle proof from JSON for milestone payload: %w", err)
	}

	payload := &Milestone{}
	payload.Index = uint32(j.Index)
	payload.Timestamp = uint64(j.Timestamp)
	copy(payload.InclusionMerkleProof[:], inclusionMerkleProofBytes)

	payload.Signatures = make([]MilestoneSignature, len(j.Signatures))
	for i, sigHex := range j.Signatures {
		sigBytes, err := hex.DecodeString(sigHex)
		if err != nil {
			return nil, fmt.Errorf("unable to decode signature from JSON for milestone payload at pos %d: %w", i, err)
		}
		copy(payload.Signatures[i][:], sigBytes)
	}
	return payload, nil
}
