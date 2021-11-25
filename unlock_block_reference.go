package iotago

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	// ReferenceUnlockBlockSize defines the size of a ReferenceUnlockBlock.
	ReferenceUnlockBlockSize = serializer.SmallTypeDenotationByteSize + serializer.UInt16ByteSize
)

// ReferenceUnlockBlock is an UnlockBlock which references a previous unlock block.
type ReferenceUnlockBlock struct {
	// The other unlock block this ReferenceUnlockBlock references to.
	Reference uint16
}

func (r *ReferenceUnlockBlock) SourceAllowed(address Address) bool {
	_, ok := address.(ChainConstrainedAddress)
	return !ok
}

func (r *ReferenceUnlockBlock) Chainable() bool {
	return false
}

func (r *ReferenceUnlockBlock) Ref() uint16 {
	return r.Reference
}

func (r *ReferenceUnlockBlock) Type() UnlockBlockType {
	return UnlockBlockReference
}

func (r *ReferenceUnlockBlock) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
		if err := serializer.CheckMinByteLength(ReferenceUnlockBlockSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid reference unlock block bytes: %w", err)
		}
		if err := serializer.CheckTypeByte(data, byte(UnlockBlockReference)); err != nil {
			return 0, fmt.Errorf("unable to deserialize reference unlock block: %w", err)
		}
	}
	data = data[serializer.SmallTypeDenotationByteSize:]
	r.Reference = binary.LittleEndian.Uint16(data)
	return ReferenceUnlockBlockSize, nil
}

func (r *ReferenceUnlockBlock) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	var b [ReferenceUnlockBlockSize]byte
	b[0] = byte(UnlockBlockReference)
	binary.LittleEndian.PutUint16(b[serializer.SmallTypeDenotationByteSize:], r.Reference)
	return b[:], nil
}

func (r *ReferenceUnlockBlock) MarshalJSON() ([]byte, error) {
	jReferenceUnlockBlock := &jsonReferenceUnlockBlock{}
	jReferenceUnlockBlock.Type = int(UnlockBlockReference)
	jReferenceUnlockBlock.Reference = int(r.Reference)
	return json.Marshal(jReferenceUnlockBlock)
}

func (r *ReferenceUnlockBlock) UnmarshalJSON(bytes []byte) error {
	jReferenceUnlockBlock := &jsonReferenceUnlockBlock{}
	if err := json.Unmarshal(bytes, jReferenceUnlockBlock); err != nil {
		return err
	}
	seri, err := jReferenceUnlockBlock.ToSerializable()
	if err != nil {
		return err
	}
	*r = *seri.(*ReferenceUnlockBlock)
	return nil
}

// jsonReferenceUnlockBlock defines the json representation of a ReferenceUnlockBlock.
type jsonReferenceUnlockBlock struct {
	Type      int `json:"type"`
	Reference int `json:"reference"`
}

func (j *jsonReferenceUnlockBlock) ToSerializable() (serializer.Serializable, error) {
	block := &ReferenceUnlockBlock{Reference: uint16(j.Reference)}
	return block, nil
}
