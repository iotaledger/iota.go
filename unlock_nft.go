package iotago

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	// NFTUnlockSize defines the size of an NFTUnlock.
	NFTUnlockSize = serializer.SmallTypeDenotationByteSize + serializer.UInt16ByteSize
)

// NFTUnlock is an Unlock which references a previous unlock.
type NFTUnlock struct {
	// The other unlock this NFTUnlock references to.
	Reference uint16
}

func (r *NFTUnlock) SourceAllowed(address Address) bool {
	_, ok := address.(*NFTAddress)

	return ok
}

func (r *NFTUnlock) Chainable() bool {
	return true
}

func (r *NFTUnlock) Ref() uint16 {
	return r.Reference
}

func (r *NFTUnlock) Type() UnlockType {
	return UnlockNFT
}

func (r *NFTUnlock) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
		if err := serializer.CheckMinByteLength(NFTUnlockSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid NFT unlock bytes: %w", err)
		}
		if err := serializer.CheckTypeByte(data, byte(UnlockNFT)); err != nil {
			return 0, fmt.Errorf("unable to deserialize NFT unlock: %w", err)
		}
	}
	data = data[serializer.SmallTypeDenotationByteSize:]
	r.Reference = binary.LittleEndian.Uint16(data)

	return NFTUnlockSize, nil
}

func (r *NFTUnlock) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	var b [NFTUnlockSize]byte
	b[0] = byte(UnlockNFT)
	binary.LittleEndian.PutUint16(b[serializer.SmallTypeDenotationByteSize:], r.Reference)

	return b[:], nil
}

func (r *NFTUnlock) Size() int {
	return NFTUnlockSize
}

func (r *NFTUnlock) MarshalJSON() ([]byte, error) {
	jNFTUnlock := &jsonNFTUnlock{}
	jNFTUnlock.Type = int(UnlockNFT)
	jNFTUnlock.Reference = int(r.Reference)

	return json.Marshal(jNFTUnlock)
}

func (r *NFTUnlock) UnmarshalJSON(bytes []byte) error {
	jNFTUnlock := &jsonNFTUnlock{}
	if err := json.Unmarshal(bytes, jNFTUnlock); err != nil {
		return err
	}
	seri, err := jNFTUnlock.ToSerializable()
	if err != nil {
		return err
	}
	*r = *seri.(*NFTUnlock)

	return nil
}

// jsonNFTUnlock defines the json representation of an NFTUnlock.
type jsonNFTUnlock struct {
	Type      int `json:"type"`
	Reference int `json:"reference"`
}

func (j *jsonNFTUnlock) ToSerializable() (serializer.Serializable, error) {
	unlock := &NFTUnlock{Reference: uint16(j.Reference)}

	return unlock, nil
}
