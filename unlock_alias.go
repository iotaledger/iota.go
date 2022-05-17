package iotago

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	// AliasUnlockSize defines the size of an AliasUnlock.
	AliasUnlockSize = serializer.SmallTypeDenotationByteSize + serializer.UInt16ByteSize
)

// AliasUnlock is an Unlock which references a previous unlock.
type AliasUnlock struct {
	// The other unlock this AliasUnlock references to.
	Reference uint16
}

func (r *AliasUnlock) SourceAllowed(address Address) bool {
	_, ok := address.(*AliasAddress)
	return ok
}

func (r *AliasUnlock) Chainable() bool {
	return true
}

func (r *AliasUnlock) Ref() uint16 {
	return r.Reference
}

func (r *AliasUnlock) Type() UnlockType {
	return UnlockAlias
}

func (r *AliasUnlock) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
		if err := serializer.CheckMinByteLength(AliasUnlockSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid alias unlock bytes: %w", err)
		}
		if err := serializer.CheckTypeByte(data, byte(UnlockAlias)); err != nil {
			return 0, fmt.Errorf("unable to deserialize alias unlock: %w", err)
		}
	}
	data = data[serializer.SmallTypeDenotationByteSize:]
	r.Reference = binary.LittleEndian.Uint16(data)
	return AliasUnlockSize, nil
}

func (r *AliasUnlock) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	var b [AliasUnlockSize]byte
	b[0] = byte(UnlockAlias)
	binary.LittleEndian.PutUint16(b[serializer.SmallTypeDenotationByteSize:], r.Reference)
	return b[:], nil
}

func (r *AliasUnlock) Size() int {
	return AliasUnlockSize
}

func (r *AliasUnlock) MarshalJSON() ([]byte, error) {
	jAliasUnlock := &jsonAliasUnlock{}
	jAliasUnlock.Type = int(UnlockAlias)
	jAliasUnlock.Reference = int(r.Reference)
	return json.Marshal(jAliasUnlock)
}

func (r *AliasUnlock) UnmarshalJSON(bytes []byte) error {
	jAliasUnlock := &jsonAliasUnlock{}
	if err := json.Unmarshal(bytes, jAliasUnlock); err != nil {
		return err
	}
	seri, err := jAliasUnlock.ToSerializable()
	if err != nil {
		return err
	}
	*r = *seri.(*AliasUnlock)
	return nil
}

// jsonAliasUnlock defines the json representation of an AliasUnlock.
type jsonAliasUnlock struct {
	Type      int `json:"type"`
	Reference int `json:"reference"`
}

func (j *jsonAliasUnlock) ToSerializable() (serializer.Serializable, error) {
	unlock := &AliasUnlock{Reference: uint16(j.Reference)}
	return unlock, nil
}
