package iotago

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	// ReferenceUnlockSize defines the size of a ReferenceUnlock.
	ReferenceUnlockSize = serializer.SmallTypeDenotationByteSize + serializer.UInt16ByteSize
)

// ReferenceUnlock is an Unlock which references a previous unlock.
type ReferenceUnlock struct {
	// The other unlock this ReferenceUnlock references to.
	Reference uint16
}

func (r *ReferenceUnlock) SourceAllowed(address Address) bool {
	_, ok := address.(ChainConstrainedAddress)
	return !ok
}

func (r *ReferenceUnlock) Chainable() bool {
	return false
}

func (r *ReferenceUnlock) Ref() uint16 {
	return r.Reference
}

func (r *ReferenceUnlock) Type() UnlockType {
	return UnlockReference
}

func (r *ReferenceUnlock) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
		if err := serializer.CheckMinByteLength(ReferenceUnlockSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid reference unlock bytes: %w", err)
		}
		if err := serializer.CheckTypeByte(data, byte(UnlockReference)); err != nil {
			return 0, fmt.Errorf("unable to deserialize reference unlock: %w", err)
		}
	}
	data = data[serializer.SmallTypeDenotationByteSize:]
	r.Reference = binary.LittleEndian.Uint16(data)
	return ReferenceUnlockSize, nil
}

func (r *ReferenceUnlock) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	var b [ReferenceUnlockSize]byte
	b[0] = byte(UnlockReference)
	binary.LittleEndian.PutUint16(b[serializer.SmallTypeDenotationByteSize:], r.Reference)
	return b[:], nil
}

func (r *ReferenceUnlock) Size() int {
	return ReferenceUnlockSize
}

func (r *ReferenceUnlock) MarshalJSON() ([]byte, error) {
	jReferenceUnlock := &jsonReferenceUnlock{}
	jReferenceUnlock.Type = int(UnlockReference)
	jReferenceUnlock.Reference = int(r.Reference)
	return json.Marshal(jReferenceUnlock)
}

func (r *ReferenceUnlock) UnmarshalJSON(bytes []byte) error {
	jReferenceUnlock := &jsonReferenceUnlock{}
	if err := json.Unmarshal(bytes, jReferenceUnlock); err != nil {
		return err
	}
	seri, err := jReferenceUnlock.ToSerializable()
	if err != nil {
		return err
	}
	*r = *seri.(*ReferenceUnlock)
	return nil
}

// jsonReferenceUnlock defines the json representation of a ReferenceUnlock.
type jsonReferenceUnlock struct {
	Type      int `json:"type"`
	Reference int `json:"reference"`
}

func (j *jsonReferenceUnlock) ToSerializable() (serializer.Serializable, error) {
	unlock := &ReferenceUnlock{Reference: uint16(j.Reference)}
	return unlock, nil
}
