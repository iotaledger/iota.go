package iota

import (
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
	MilestoneInclusionMerkleProofLength = 64
	// Defines the length of the milestone signature.
	MilestoneSignatureLength = 64
	// Defines the length of a Milestone ID.
	MilestoneIDLength = blake2b.Size256
	// Defines the serialized size of a milestone payload.
	MilestoneBinSerializedSize = TypeDenotationByteSize + UInt64ByteSize + UInt64ByteSize + MilestoneInclusionMerkleProofLength + MilestoneSignatureLength
	// Defines the size of a milestone payload without the signature.
	MilestoneBinSerializedSizeWithoutSignature = MilestoneBinSerializedSize - MilestoneSignatureLength
	// Defines the size of the to be signed data of a milestone. Consists of: message version,  message parents, payload length,
	// payload type, milestone index+timestamp+inclusion merkle proof
	MilestoneSignatureInputSize = OneByte + MessageIDLength*2 + PayloadLengthByteSize + MilestoneBinSerializedSizeWithoutSignature
)

var (
	// Returned if the signature of a milestone is invalid.
	ErrInvalidMilestoneSignature = errors.New("invalid milestone signature")
)

// MilestoneID is the ID of a Milestone.
type MilestoneID = [MilestoneIDLength]byte

// Milestone holds the inclusion merkle proof and milestone signature.
type Milestone struct {
	// The index of this milestone.
	Index uint64
	// The time at which this milestone was issued.
	Timestamp uint64
	// The inclusion merkle proof of included/newly confirmed transaction hashes.
	InclusionMerkleProof [MilestoneInclusionMerkleProofLength]byte
	// The signature of the milestone.
	Signature [MilestoneSignatureLength]byte
}

// Hash computes the hash of the Milestone.
func (m *Milestone) Hash() (*MilestoneID, error) {
	data, err := m.Serialize(DeSeriModeNoValidation)
	if err != nil {
		return nil, fmt.Errorf("can't compute milestone payload hash: %w", err)
	}
	h := blake2b.Sum256(data)
	return &h, nil
}

// VerifySignature verifies the given milestone signature in conjunction with the given message.
func (m *Milestone) VerifySignature(msg *Message, pubKey ed25519.PublicKey) error {
	sigInput, sig, err := m.SignatureInput(msg)
	if err != nil {
		return fmt.Errorf("can't compute milestone signature input for signature verification: %w", err)
	}
	if ok := ed25519.Verify(pubKey, sigInput[:], sig[:]); !ok {
		return ErrInvalidMilestoneSignature
	}
	return nil
}

// Sign produces the signature with the given envelope message and updates the Signature field of the milestone payload.
func (m *Milestone) Sign(msg *Message, prvKey ed25519.PrivateKey) error {
	sigInput, _, err := m.SignatureInput(msg)
	if err != nil {
		return fmt.Errorf("can't compute milestone signature input for signing: %w", err)
	}
	copy(m.Signature[:], ed25519.Sign(prvKey, sigInput[:]))
	return nil
}

// SignatureInput returns the input data to be signed and the current signature.
func (m *Milestone) SignatureInput(msg *Message) (sigInput [MilestoneSignatureInputSize]byte, sig [MilestoneSignatureLength]byte, err error) {
	var msData []byte
	msData, err = m.Serialize(DeSeriModeNoValidation)
	if err != nil {
		return
	}
	msSigRelevantData := msData[:MilestoneBinSerializedSizeWithoutSignature]

	var offset int
	sigInput[0] = msg.Version
	offset += OneByte
	copy(sigInput[offset:offset+MessageIDLength], msg.Parent1[:])
	offset += MessageIDLength
	copy(sigInput[offset:offset+MessageIDLength], msg.Parent2[:])
	offset += MessageIDLength
	// note this is the size of a complete ms payload with the signature
	binary.LittleEndian.PutUint32(sigInput[offset:offset+UInt32ByteSize], uint32(MilestoneBinSerializedSize))
	offset += UInt32ByteSize
	// copy milestone payload data (without signature)
	copy(sigInput[offset:], msSigRelevantData)

	// copy sig
	copy(sig[:], msData[MilestoneBinSerializedSizeWithoutSignature:])
	return
}

func (m *Milestone) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkMinByteLength(MilestoneBinSerializedSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid milestone payload bytes: %w", err)
		}
		if err := checkType(data, MilestonePayloadTypeID); err != nil {
			return 0, fmt.Errorf("unable to deserialize milestone payload: %w", err)
		}
	}
	data = data[TypeDenotationByteSize:]

	// read index
	m.Index = binary.LittleEndian.Uint64(data)
	data = data[UInt64ByteSize:]

	// read timestamp
	m.Timestamp = binary.LittleEndian.Uint64(data)
	data = data[UInt64ByteSize:]

	// read merkle proof and signature
	copy(m.InclusionMerkleProof[:], data[:MilestoneInclusionMerkleProofLength])
	data = data[MilestoneInclusionMerkleProofLength:]
	copy(m.Signature[:], data[:MilestoneSignatureLength])

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		// TODO: validation
	}

	return MilestoneBinSerializedSize, nil
}

func (m *Milestone) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		// TODO: validation
	}
	var b [MilestoneBinSerializedSize]byte
	binary.LittleEndian.PutUint32(b[:], MilestonePayloadTypeID)
	binary.LittleEndian.PutUint64(b[TypeDenotationByteSize:], m.Index)
	binary.LittleEndian.PutUint64(b[TypeDenotationByteSize+UInt64ByteSize:], m.Timestamp)
	offset := TypeDenotationByteSize + UInt64ByteSize + UInt64ByteSize
	copy(b[offset:], m.InclusionMerkleProof[:])
	copy(b[offset+MilestoneInclusionMerkleProofLength:], m.Signature[:])
	return b[:], nil
}

func (m *Milestone) MarshalJSON() ([]byte, error) {
	jsonMilestonePayload := &jsonmilestonepayload{}
	jsonMilestonePayload.Type = int(MilestonePayloadTypeID)
	jsonMilestonePayload.Index = int(m.Index)
	jsonMilestonePayload.Signature = hex.EncodeToString(m.Signature[:])
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
	Type                 int    `json:"type"`
	Index                int    `json:"index"`
	Timestamp            int    `json:"timestamp"`
	InclusionMerkleProof string `json:"inclusionMerkleProof"`
	Signature            string `json:"signature"`
}

func (j *jsonmilestonepayload) ToSerializable() (Serializable, error) {
	inclusionMerkleProofBytes, err := hex.DecodeString(j.InclusionMerkleProof)
	if err != nil {
		return nil, fmt.Errorf("unable to decode inlcusion merkle proof from JSON for milestone payload: %w", err)
	}

	signatureBytes, err := hex.DecodeString(j.Signature)
	if err != nil {
		return nil, fmt.Errorf("unable to decode signature from JSON for milestone payload: %w", err)
	}

	payload := &Milestone{}
	copy(payload.InclusionMerkleProof[:], inclusionMerkleProofBytes)
	copy(payload.Signature[:], signatureBytes)

	payload.Index = uint64(j.Index)
	payload.Timestamp = uint64(j.Timestamp)
	return payload, nil
}
