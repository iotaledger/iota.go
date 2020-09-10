package iota

import (
	"encoding/binary"
	"fmt"
)

const (
	MilestonePayloadID                  uint32 = 1
	MilestoneInclusionMerkleProofLength        = 64
	MilestoneSignatureLength                   = 64
	MilestoneHashLength                        = 32
	MilestonePayloadSize                       = TypeDenotationByteSize + UInt64ByteSize + UInt64ByteSize + MilestoneInclusionMerkleProofLength + MilestoneSignatureLength
)

// MilestonePayload holds the inclusion merkle proof and milestone signature.
type MilestonePayload struct {
	Index                uint64                                    `json:"index"`
	Timestamp            uint64                                    `json:"timestamp"`
	InclusionMerkleProof [MilestoneInclusionMerkleProofLength]byte `json:"inclusion_merkle_proof"`
	Signature            [MilestoneSignatureLength]byte            `json:"signature"`
}

func (m *MilestonePayload) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkMinByteLength(MilestonePayloadSize, len(data)); err != nil {
			return 0, err
		}
		if err := checkType(data, MilestonePayloadID); err != nil {
			return 0, fmt.Errorf("unable to deserialize milestone payload: %w", err)
		}
	}
	data = data[TypeDenotationByteSize:]

	// read inex
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
	var b [MilestonePayloadSize]byte
	binary.LittleEndian.PutUint32(b[:], MilestonePayloadID)
	binary.LittleEndian.PutUint64(b[TypeDenotationByteSize:], m.Index)
	binary.LittleEndian.PutUint64(b[TypeDenotationByteSize+UInt64ByteSize:], m.Timestamp)
	offset := TypeDenotationByteSize + UInt64ByteSize + UInt64ByteSize
	copy(b[offset:], m.InclusionMerkleProof[:])
	copy(b[offset+MilestoneInclusionMerkleProofLength:], m.Signature[:])
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		// TODO: validation
	}
	return b[:], nil
}
