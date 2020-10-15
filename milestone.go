package iota

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
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
	// Defines the serialized size of a milestone payload.
	MilestoneBinSerializedMinSize = TypeDenotationByteSize + UInt64ByteSize + UInt64ByteSize + MilestoneInclusionMerkleProofLength + OneByte + MilestoneSignatureLength
	// Defines the size of a milestone payload without the signatures.
	MilestoneBinSerializedSizeWithoutSignatures = MilestoneBinSerializedMinSize - MilestoneSignatureLength
	// Defines the size of the to be signed data of a milestone. Consists of: message version,  message parents, payload length,
	// payload type, milestone index+timestamp+inclusion merkle proof+signatures count
	MilestoneSignatureInputSize = OneByte + MessageIDLength*2 + PayloadLengthByteSize + MilestoneBinSerializedSizeWithoutSignatures
	// MaxSignaturesInAMilestone is the maximum amount of signatures in a milestone.
	MaxSignaturesInAMilestone = 255
)

var (
	// Returned if the minimum amount of valid signatures on a milestone isn't reached.
	ErrMilestoneMinSigsThresholdNotReached = errors.New("did not reach minimum amount of valid signatures")
	// Returned if a to be deserialized milestone does not contain at least one signature.
	ErrMilestoneInvalidSignatureCount = errors.New("a milestone must hold at least one signature")
	// Returned when a MilestoneSigningFunc produces less signatures than expected.
	ErrMilestoneProducedSigsCountMismatch = errors.New("produced and wanted signature count mismatch")
	// Returned when a milestone holds more than 255 signatures.
	ErrMilestoneTooManySignatures = fmt.Errorf("a milestone can hold max %d signatures", MaxSignaturesInAMilestone)
	// Returned when an invalid min signatures threshold is given the the verification function.
	ErrMilestoneInvalidMinSigsThreshold = fmt.Errorf("min threshold must be at least 1")
	// Returned when the same public key has been seen multiple times during milestone signatures verification.
	ErrMilestoneSigVerificationDupPublicKeys = fmt.Errorf("duplicated public key encountered during milestone signatures verification")
)

// MilestoneID is the ID of a Milestone.
type MilestoneID = [MilestoneIDLength]byte

// Milestone represents a special payload which defines the inclusion set
// of other messages in the Tangle. It holds the inclusion merkle proof defining which
// messages have been applied to the ledger. A Milestone holds multiple signatures
// of which a minimum amount have to be valid in order to deem the entire Milestone valid.
type Milestone struct {
	// The index of this milestone.
	Index uint64
	// The time at which this milestone was issued.
	Timestamp uint64
	// The inclusion merkle proof of included/newly confirmed transaction IDs.
	InclusionMerkleProof [MilestoneInclusionMerkleProofLength]byte
	// The signatures held by the milestone.
	Signatures [][MilestoneSignatureLength]byte
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

// VerifySignatures verifies that at least minValid signatures are valid via the given public keys.
// Returns the public keys which successfully verified a signature, this can be used to determine
// whether a public key needs to be updated. pubKeys must not contain duplicated public keys.
func (m *Milestone) VerifySignatures(msg *Message, minValid int, pubKeys []ed25519.PublicKey) ([]ed25519.PublicKey, error) {
	if minValid == 0 {
		return nil, ErrMilestoneInvalidMinSigsThreshold
	}

	if len(m.Signatures) == 0 {
		return nil, ErrMilestoneInvalidSignatureCount
	}

	msSigInput, err := m.SignatureInput(msg, byte(len(m.Signatures)))
	if err != nil {
		return nil, fmt.Errorf("unable to compute milestone signature input for signature verification: %w", err)
	}

	// TODO: should len(m.Signatures) == len(pubKeys)?
	sigs := make([][MilestoneSignatureLength]byte, len(m.Signatures))
	copy(sigs, m.Signatures)

	valid := make([]ed25519.PublicKey, 0)
	seenPubKeys := map[string]int{}
	for pubKeyIndex, pubKey := range pubKeys {

		// sanity check that the caller isn't supplying the same
		// public key multiple times
		seenPubKeyMapKey := string(pubKey)
		seenPubKeyIndex, seen := seenPubKeys[seenPubKeyMapKey]
		if seen {
			return nil, fmt.Errorf("%w: index %d and %d", ErrMilestoneSigVerificationDupPublicKeys, seenPubKeyIndex, pubKeyIndex)
		}
		seenPubKeys[seenPubKeyMapKey] = pubKeyIndex

		for sigIndex, sig := range sigs {
			if ok := ed25519.Verify(pubKey, msSigInput[:], sig[:]); ok {
				// remove the verified signature since each signature should only match one public key
				sigs = append(sigs[:sigIndex], sigs[sigIndex+1:]...)
				valid = append(valid, pubKey)
				break
			}
		}
	}

	if len(valid) < minValid {
		return nil, fmt.Errorf("%w: %d out of %d", ErrMilestoneMinSigsThresholdNotReached, len(valid), minValid)
	}

	return valid, nil
}

// MilestoneSigningFunc is a function which produces a set of signatures for the given Milestone essence data.
type MilestoneSigningFunc func(msSigInput [MilestoneSignatureInputSize]byte) ([][MilestoneSignatureLength]byte, error)

// InMemoryEd25519MilestoneSigner is a function which uses the provided Ed25519 private keys to produce signatures for the Milestone essence data.
func InMemoryEd25519MilestoneSigner(prvKeys ...ed25519.PrivateKey) MilestoneSigningFunc {
	return func(msSigInput [MilestoneSignatureInputSize]byte) ([][MilestoneSignatureLength]byte, error) {
		sigs := make([][MilestoneSignatureLength]byte, len(prvKeys))
		for i, prvKey := range prvKeys {
			sig := ed25519.Sign(prvKey, msSigInput[:])
			copy(sigs[i][:], sig)
		}
		return sigs, nil
	}
}

// Sign produces the signatures with the given envelope message and updates the Signatures field of the Milestone
// with the resulting signatures of the given MilestoneSigningFunc.
func (m *Milestone) Sign(msg *Message, sigsCount byte, signingFunc MilestoneSigningFunc) error {
	msSigInput, err := m.SignatureInput(msg, sigsCount)
	if err != nil {
		return fmt.Errorf("unable to compute milestone signature input for signing: %w", err)
	}

	sigs, err := signingFunc(msSigInput)
	if err != nil {
		return fmt.Errorf("unable to produce milestone signatures: %w", err)
	}

	if int(sigsCount) != len(sigs) {
		return fmt.Errorf("%w: wanted %d signatures but only produced %d", ErrMilestoneProducedSigsCountMismatch, sigsCount, len(sigs))
	}

	if len(sigs) == 0 {
		return fmt.Errorf("%w: not enough signatures were produced during signing", ErrMilestoneInvalidSignatureCount)
	}

	if len(sigs) > MaxSignaturesInAMilestone {
		return fmt.Errorf("%w: too many signatures were produced during signing", ErrMilestoneTooManySignatures)
	}

	m.Signatures = sigs
	return nil
}

// SignatureInput returns the input data to be signed.
func (m *Milestone) SignatureInput(msg *Message, sigsCount byte) (sigInput [MilestoneSignatureInputSize]byte, err error) {
	var msData []byte
	msData, err = m.Serialize(DeSeriModeNoValidation)
	if err != nil {
		return
	}

	msSigRelevantData := msData[:MilestoneBinSerializedSizeWithoutSignatures]

	var offset int
	sigInput[0] = msg.Version
	offset += OneByte
	// parents
	copy(sigInput[offset:offset+MessageIDLength], msg.Parent1[:])
	offset += MessageIDLength
	copy(sigInput[offset:offset+MessageIDLength], msg.Parent2[:])
	offset += MessageIDLength
	// payload length
	// we need to calculate how big the milestone will be in its serialized form
	msSerializedSize := MilestoneBinSerializedSizeWithoutSignatures + MilestoneSignatureLength*int(sigsCount)
	binary.LittleEndian.PutUint32(sigInput[offset:offset+UInt32ByteSize], uint32(msSerializedSize))
	offset += UInt32ByteSize
	// milestone data without signatures
	copy(sigInput[offset:], msSigRelevantData)

	// override signature count
	sigInput[len(sigInput)-1] = sigsCount

	return
}

func (m *Milestone) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkMinByteLength(MilestoneBinSerializedMinSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid milestone payload bytes: %w", err)
		}
		if err := checkType(data, MilestonePayloadTypeID); err != nil {
			return 0, fmt.Errorf("unable to deserialize milestone payload: %w", err)
		}
	}

	bytesReadTotal := TypeDenotationByteSize
	data = data[TypeDenotationByteSize:]

	// read index
	m.Index = binary.LittleEndian.Uint64(data)
	data = data[UInt64ByteSize:]
	bytesReadTotal += UInt64ByteSize

	m.Timestamp = binary.LittleEndian.Uint64(data)
	data = data[UInt64ByteSize:]
	// read timestamp
	bytesReadTotal += UInt64ByteSize

	// read inclusion set merkle proof
	copy(m.InclusionMerkleProof[:], data[:MilestoneInclusionMerkleProofLength])
	data = data[MilestoneInclusionMerkleProofLength:]
	bytesReadTotal += MilestoneInclusionMerkleProofLength

	// read signature count
	sigCount := data[0]
	data = data[OneByte:]
	bytesReadTotal += OneByte

	if sigCount == 0 {
		return 0, fmt.Errorf("unable to deserialize milestone payload: %w", ErrMilestoneInvalidSignatureCount)
	}

	sigs := make([][MilestoneSignatureLength]byte, sigCount)
	for i := 0; i < int(sigCount); i++ {
		if len(data) < MilestoneSignatureLength {
			return 0, fmt.Errorf("unable to deserialize signature in milestone payload: %w", ErrDeserializationNotEnoughData)
		}
		copy(sigs[i][:], data[:MilestoneSignatureLength])
		data = data[MilestoneSignatureLength:]
	}
	bytesReadTotal += int(sigCount) * MilestoneSignatureLength

	m.Signatures = sigs

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		// TODO: validation
	}

	return bytesReadTotal, nil
}

func (m *Milestone) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		// TODO: validation
	}

	var b bytes.Buffer
	if err := binary.Write(&b, binary.LittleEndian, MilestonePayloadTypeID); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize milestone payload ID", err)
	}
	if err := binary.Write(&b, binary.LittleEndian, m.Index); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize milestone index", err)
	}
	if err := binary.Write(&b, binary.LittleEndian, m.Timestamp); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize milestone timestamp", err)
	}
	if _, err := b.Write(m.InclusionMerkleProof[:]); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize milestone inclusion merkle proof", err)
	}
	if len(m.Signatures) > MaxSignaturesInAMilestone {
		return nil, fmt.Errorf("%w: unable to serialize milestone", ErrMilestoneTooManySignatures)
	}
	if err := b.WriteByte(byte(len(m.Signatures))); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize milestone signatures count", err)
	}
	for i, sig := range m.Signatures {
		if _, err := b.Write(sig[:]); err != nil {
			return nil, fmt.Errorf("%w: unable to serialize milestone signature at pos %d", err, i)
		}
	}

	return b.Bytes(), nil
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
	payload.Index = uint64(j.Index)
	payload.Timestamp = uint64(j.Timestamp)
	copy(payload.InclusionMerkleProof[:], inclusionMerkleProofBytes)

	payload.Signatures = make([][MilestoneSignatureLength]byte, len(j.Signatures))
	for i, sigHex := range j.Signatures {
		sigBytes, err := hex.DecodeString(sigHex)
		if err != nil {
			return nil, fmt.Errorf("unable to decode signature from JSON for milestone payload at pos %d: %w", i, err)
		}
		copy(payload.Signatures[i][:], sigBytes)
	}
	return payload, nil
}
