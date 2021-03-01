package iotago

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"golang.org/x/crypto/blake2b"
)

const (
	// The length of a TreasuryInput
	TreasuryInputBytesLength = blake2b.Size256
	// The size of a serialized TreasuryInput with its type denoting byte.
	TreasuryInputSerializedBytesSize = SmallTypeDenotationByteSize + TreasuryInputBytesLength
)

// TreasuryInput is an input which references a milestone which generated a TreasuryOutput.
type TreasuryInput [32]byte

func (ti *TreasuryInput) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkMinByteLength(TreasuryInputSerializedBytesSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid treasury input bytes: %w", err)
		}
		if err := checkTypeByte(data, InputTreasury); err != nil {
			return 0, fmt.Errorf("unable to deserialize treasury input: %w", err)
		}
	}
	copy(ti[:], data[SmallTypeDenotationByteSize:])
	return TreasuryInputSerializedBytesSize, nil
}

func (ti *TreasuryInput) Serialize(deSeriMode DeSerializationMode) (data []byte, err error) {
	var b [TreasuryInputSerializedBytesSize]byte
	b[0] = InputTreasury
	copy(b[SmallTypeDenotationByteSize:], ti[:])
	return b[:], nil
}

func (ti *TreasuryInput) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsontreasuryinput{
		Type:          int(InputTreasury),
		MilestoneHash: hex.EncodeToString(ti[:]),
	})
}

func (ti *TreasuryInput) UnmarshalJSON(bytes []byte) error {
	jsonTreasuryInput := &jsontreasuryinput{}
	if err := json.Unmarshal(bytes, jsonTreasuryInput); err != nil {
		return err
	}
	seri, err := jsonTreasuryInput.ToSerializable()
	if err != nil {
		return err
	}
	*ti = *seri.(*TreasuryInput)
	return nil
}

// jsontreasuryinput defines the json representation of a TreasuryInput.
type jsontreasuryinput struct {
	Type          int    `json:"type"`
	MilestoneHash string `json:"milestoneHash"`
}

func (j *jsontreasuryinput) ToSerializable() (Serializable, error) {
	msHash, err := hex.DecodeString(j.MilestoneHash)
	if err != nil {
		return nil, fmt.Errorf("unable to decode milestone hash from JSON for treasury input: %w", err)
	}
	if err := checkExactByteLength(len(msHash), MilestoneIDLength); err != nil {
		return nil, fmt.Errorf("unable to decode milestone hash from JSON for treasury input: %w", err)
	}
	input := &TreasuryInput{}
	copy(input[:], msHash)
	return input, nil
}
