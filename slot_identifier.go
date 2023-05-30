package iotago

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sync"

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

// SlotIdentifierRepresentingData returns a new SlotIdentifier for the given data by hashing it with blake2b and associating it with the given slot index.
func SlotIdentifierRepresentingData(index SlotIndex, data []byte) SlotIdentifier {
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

	return SlotIdentifierFromBytes(bytes)
}

// SlotIdentifierFromBytes returns a new SlotIdentifier represented by the passed bytes.
func SlotIdentifierFromBytes(bytes []byte) (SlotIdentifier, error) {
	if len(bytes) != SlotIdentifierLength {
		return SlotIdentifier{}, ErrInvalidIdentifierLength
	}

	return SlotIdentifier(bytes), nil
}

// MustSlotIdentifierFromHexString converts the hex to a SlotIdentifier representation.
func MustSlotIdentifierFromHexString(hex string) SlotIdentifier {
	id, err := SlotIdentifierFromHexString(hex)
	if err != nil {
		panic(err)
	}

	return id
}

func (id SlotIdentifier) Bytes() ([]byte, error) {
	return id[:], nil
}

func (id *SlotIdentifier) FromBytes(bytes []byte) (int, error) {
	var err error
	*id, err = SlotIdentifierFromBytes(bytes)
	if err != nil {
		return 0, err
	}
	return SlotIdentifierLength, nil
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
	return id.Alias()
}

func (id SlotIdentifier) Index() SlotIndex {
	return SlotIndex(binary.LittleEndian.Uint64(id[IdentifierLength:]))
}

func (id SlotIdentifier) Identifier() Identifier {
	return Identifier(id[:IdentifierLength])
}

var (
	// slotIidentifierAliases contains a dictionary of identifiers associated to their human-readable alias.
	slotIidentifierAliases = make(map[SlotIdentifier]string)

	// slotIdentifierAliasesMutex is the mutex that is used to synchronize access to the previous map.
	slotIdentifierAliasesMutex = sync.RWMutex{}
)

// RegisterAlias allows to register a human-readable alias for the Identifier which will be used as a replacement for
// the String method.
func (id SlotIdentifier) RegisterAlias(alias string) {
	slotIdentifierAliasesMutex.Lock()
	defer slotIdentifierAliasesMutex.Unlock()

	slotIidentifierAliases[id] = alias
}

// Alias returns the human-readable alias of the Identifier (or the base58 encoded bytes of no alias was set).
func (id SlotIdentifier) Alias() (alias string) {
	slotIdentifierAliasesMutex.RLock()
	defer slotIdentifierAliasesMutex.RUnlock()

	if existingAlias, exists := slotIidentifierAliases[id]; exists {
		return existingAlias
	}

	return fmt.Sprintf("%s:%d", id.ToHex(), id.Index())
}

// UnregisterAlias allows to unregister a previously registered alias.
func (id SlotIdentifier) UnregisterAlias() {
	slotIdentifierAliasesMutex.Lock()
	defer slotIdentifierAliasesMutex.Unlock()

	delete(slotIidentifierAliases, id)
}
