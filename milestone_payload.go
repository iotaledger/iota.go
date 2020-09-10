package iota

import (
	"encoding/binary"
	"fmt"
)

const (
	// Defines the milestone payload's ID.
	MilestonePayloadID uint32 = 1
	// Defines the length of the inclusion merkle proof within a milestone payload.
	MilestoneInclusionMerkleProofLength = 64
	// Defines the length of the milestone signature.
	MilestoneSignatureLength = 64
	// Defines the length of a milestone hash.
	MilestoneHashLength = 32
	// Defines the size of a milestone payload.
	MilestonePayloadSize = TypeDenotationByteSize + UInt64ByteSize + UInt64ByteSize + MilestoneInclusionMerkleProofLength + MilestoneSignatureLength
)

// MilestonePayload holds the inclusion merkle proof and milestone signature.
type MilestonePayload struct {
	// The index of this milestone.
	Index uint64 `json:"index"`
	// The time at which this milestone was issued.
	Timestamp uint64 `json:"timestamp"`
	// The inclusion merkle proof of included/newly confirmed transaction hashes.
	InclusionMerkleProof [MilestoneInclusionMerkleProofLength]byte `json:"inclusion_merkle_proof"`
	// The signature of the milestone.
	Signature [MilestoneSignatureLength]byte `json:"signature"`
}

func (m *MilestonePayload) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkMinByteLength(MilestonePayloadSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid milestone payload bytes: %w", err)
		}
		if err := checkType(data, MilestonePayloadID); err != nil {
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

	return MilestonePayloadSize, nil
}

func (m *MilestonePayload) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		// TODO: validation
	}
	var b [MilestonePayloadSize]byte
	binary.LittleEndian.PutUint32(b[:], MilestonePayloadID)
	binary.LittleEndian.PutUint64(b[TypeDenotationByteSize:], m.Index)
	binary.LittleEndian.PutUint64(b[TypeDenotationByteSize+UInt64ByteSize:], m.Timestamp)
	offset := TypeDenotationByteSize + UInt64ByteSize + UInt64ByteSize
	copy(b[offset:], m.InclusionMerkleProof[:])
	copy(b[offset+MilestoneInclusionMerkleProofLength:], m.Signature[:])
	return b[:], nil
}
