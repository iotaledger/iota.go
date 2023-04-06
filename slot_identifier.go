package iotago

import (
	"encoding/binary"
	"encoding/hex"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	SlotIdentifierLength = IdentifierLength + serializer.Int64ByteSize
)

var (
	emptySlotIdentifier = SlotIdentifier{}
)

// SlotIdentifier is a 32 byte hash value that can be used to uniquely identify some blob of data together with an 8 byte slot index.
type SlotIdentifier [SlotIdentifierLength]byte

// SlotIdentifierFromData returns a new SlotIdentifier for the given data by hashing it with blake2b and associating it with the given slot index.
func SlotIdentifierFromData(index SlotIndex, data []byte) SlotIdentifier {
	return NewSlotIdentifier(index, blake2b.Sum256(data))
}

func NewSlotIdentifier(index SlotIndex, idBytes Identifier) SlotIdentifier {
	id := SlotIdentifier{}
	copy(id[:], idBytes[:])
	binary.LittleEndian.PutUint64(id[IdentifierLength:], uint64(index))

	return id
}

// SlotIdentifierFromHexString converts the hex to a SlotIdentifier representation.
func SlotIdentifierFromHexString(hex string) (SlotIdentifier, error) {
	bytes, err := DecodeHex(hex)
	if err != nil {
		return SlotIdentifier{}, err
	}

	if len(bytes) != SlotIdentifierLength {
		return SlotIdentifier{}, ErrInvalidIdentifierLength
	}

	var id SlotIdentifier
	copy(id[:], bytes)
	return id, nil
}

// MustSlotIdentifierFromHexString converts the hex to a SlotIdentifier representation.
func MustSlotIdentifierFromHexString(hex string) SlotIdentifier {
	id, err := SlotIdentifierFromHexString(hex)
	if err != nil {
		panic(err)
	}

	return id
}

func (id SlotIdentifier) MarshalText() (text []byte, err error) {
	dst := make([]byte, hex.EncodedLen(len(SlotIdentifier{})))
	hex.Encode(dst, id[:])
	return dst, nil
}

func (id *SlotIdentifier) UnmarshalText(text []byte) error {
	_, err := hex.Decode(id[:], text)
	return err
}

// Empty tells whether the Identifier is empty.
func (id SlotIdentifier) Empty() bool {
	return id == emptySlotIdentifier
}

// ToHex converts the Identifier to its hex representation.
func (id SlotIdentifier) ToHex() string {
	return EncodeHex(id[:])
}

func (id SlotIdentifier) String() string {
	return id.ToHex()
}

func (id SlotIdentifier) Slot() SlotIndex {
	return SlotIndex(binary.LittleEndian.Uint64(id[IdentifierLength:]))
}

func (id SlotIdentifier) Identifier() Identifier {
	return Identifier(id[:IdentifierLength])
}
