package iotago

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer"
)

const (
	// NFTUnlockBlockSize defines the size of an NFTUnlockBlock.
	NFTUnlockBlockSize = serializer.SmallTypeDenotationByteSize + serializer.UInt16ByteSize
)

// NFTUnlockBlock is an UnlockBlock which references a previous unlock block.
type NFTUnlockBlock struct {
	// The other unlock block this NFTUnlockBlock references to.
	Reference uint16
}

func (r *NFTUnlockBlock) SourceAllowed(address Address) bool {
	_, ok := address.(*NFTAddress)
	return ok
}

func (r *NFTUnlockBlock) Chainable() bool {
	return true
}

func (r *NFTUnlockBlock) Ref() uint16 {
	return r.Reference
}

func (r *NFTUnlockBlock) Type() UnlockBlockType {
	return UnlockBlockNFT
}

func (r *NFTUnlockBlock) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
		if err := serializer.CheckMinByteLength(NFTUnlockBlockSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid nft unlock block bytes: %w", err)
		}
		if err := serializer.CheckTypeByte(data, byte(UnlockBlockNFT)); err != nil {
			return 0, fmt.Errorf("unable to deserialize nft unlock block: %w", err)
		}
	}
	data = data[serializer.SmallTypeDenotationByteSize:]
	r.Reference = binary.LittleEndian.Uint16(data)
	return NFTUnlockBlockSize, nil
}

func (r *NFTUnlockBlock) Serialize(deSeriMode serializer.DeSerializationMode) ([]byte, error) {
	var b [NFTUnlockBlockSize]byte
	b[0] = byte(UnlockBlockNFT)
	binary.LittleEndian.PutUint16(b[serializer.SmallTypeDenotationByteSize:], r.Reference)
	return b[:], nil
}

func (r *NFTUnlockBlock) MarshalJSON() ([]byte, error) {
	jNFTUnlockBlock := &jsonNFTUnlockBlock{}
	jNFTUnlockBlock.Type = int(UnlockBlockNFT)
	jNFTUnlockBlock.Reference = int(r.Reference)
	return json.Marshal(jNFTUnlockBlock)
}

func (r *NFTUnlockBlock) UnmarshalJSON(bytes []byte) error {
	jNFTUnlockBlock := &jsonNFTUnlockBlock{}
	if err := json.Unmarshal(bytes, jNFTUnlockBlock); err != nil {
		return err
	}
	seri, err := jNFTUnlockBlock.ToSerializable()
	if err != nil {
		return err
	}
	*r = *seri.(*NFTUnlockBlock)
	return nil
}

// jsonNFTUnlockBlock defines the json representation of an NFTUnlockBlock.
type jsonNFTUnlockBlock struct {
	Type      int `json:"type"`
	Reference int `json:"reference"`
}

func (j *jsonNFTUnlockBlock) ToSerializable() (serializer.Serializable, error) {
	block := &NFTUnlockBlock{Reference: uint16(j.Reference)}
	return block, nil
}
