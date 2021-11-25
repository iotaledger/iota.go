package iotago

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/v2/serializer"
)

const (
	// AliasUnlockBlockSize defines the size of an AliasUnlockBlock.
	AliasUnlockBlockSize = serializer.SmallTypeDenotationByteSize + serializer.UInt16ByteSize
)

// AliasUnlockBlock is an UnlockBlock which references a previous unlock block.
type AliasUnlockBlock struct {
	// The other unlock block this AliasUnlockBlock references to.
	Reference uint16
}

func (r *AliasUnlockBlock) SourceAllowed(address Address) bool {
	_, ok := address.(*AliasAddress)
	return ok
}

func (r *AliasUnlockBlock) Chainable() bool {
	return true
}

func (r *AliasUnlockBlock) Ref() uint16 {
	return r.Reference
}

func (r *AliasUnlockBlock) Type() UnlockBlockType {
	return UnlockBlockAlias
}

func (r *AliasUnlockBlock) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
		if err := serializer.CheckMinByteLength(AliasUnlockBlockSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid alias unlock block bytes: %w", err)
		}
		if err := serializer.CheckTypeByte(data, byte(UnlockBlockAlias)); err != nil {
			return 0, fmt.Errorf("unable to deserialize alias unlock block: %w", err)
		}
	}
	data = data[serializer.SmallTypeDenotationByteSize:]
	r.Reference = binary.LittleEndian.Uint16(data)
	return AliasUnlockBlockSize, nil
}

func (r *AliasUnlockBlock) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	var b [AliasUnlockBlockSize]byte
	b[0] = byte(UnlockBlockAlias)
	binary.LittleEndian.PutUint16(b[serializer.SmallTypeDenotationByteSize:], r.Reference)
	return b[:], nil
}

func (r *AliasUnlockBlock) MarshalJSON() ([]byte, error) {
	jAliasUnlockBlock := &jsonAliasUnlockBlock{}
	jAliasUnlockBlock.Type = int(UnlockBlockAlias)
	jAliasUnlockBlock.Reference = int(r.Reference)
	return json.Marshal(jAliasUnlockBlock)
}

func (r *AliasUnlockBlock) UnmarshalJSON(bytes []byte) error {
	jAliasUnlockBlock := &jsonAliasUnlockBlock{}
	if err := json.Unmarshal(bytes, jAliasUnlockBlock); err != nil {
		return err
	}
	seri, err := jAliasUnlockBlock.ToSerializable()
	if err != nil {
		return err
	}
	*r = *seri.(*AliasUnlockBlock)
	return nil
}

// jsonAliasUnlockBlock defines the json representation of an AliasUnlockBlock.
type jsonAliasUnlockBlock struct {
	Type      int `json:"type"`
	Reference int `json:"reference"`
}

func (j *jsonAliasUnlockBlock) ToSerializable() (serializer.Serializable, error) {
	block := &AliasUnlockBlock{Reference: uint16(j.Reference)}
	return block, nil
}
