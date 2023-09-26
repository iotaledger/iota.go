package iotago

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sync"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/iota.go/v4/hexutil"
)

const (
	SlotIdentifierLength = IdentifierLength + SlotIndexLength
)

var (
	emptySlotIdentifier = SlotIdentifier{}
)

// SlotIdentifier is a 32 byte hash value that can be used to uniquely identify some blob of data together with an 4 byte slot index.
type SlotIdentifier [SlotIdentifierLength]byte

// SlotIdentifierRepresentingData returns a new SlotIdentifier for the given data by hashing it with blake2b and associating it with the given slot index.
func SlotIdentifierRepresentingData(slot SlotIndex, data []byte) SlotIdentifier {
	return NewSlotIdentifier(slot, blake2b.Sum256(data))
}

func NewSlotIdentifier(slot SlotIndex, idBytes Identifier) SlotIdentifier {
	id := SlotIdentifier{}
	copy(id[:], idBytes[:])
	binary.LittleEndian.PutUint32(id[IdentifierLength:], uint32(slot))

	return id
}

// SlotIdentifierFromHexString converts the hex to a SlotIdentifier representation.
func SlotIdentifierFromHexString(hex string) (SlotIdentifier, error) {
	bytes, err := hexutil.DecodeHex(hex)
	if err != nil {
		return SlotIdentifier{}, err
	}

	s, _, err := SlotIdentifierFromBytes(bytes)

	return s, err
}

// SlotIdentifierFromBytes returns a new SlotIdentifier represented by the passed bytes.
func SlotIdentifierFromBytes(bytes []byte) (SlotIdentifier, int, error) {
	if len(bytes) < SlotIdentifierLength {
		return SlotIdentifier{}, 0, ErrInvalidIdentifierLength
	}

	return SlotIdentifier(bytes), SlotIdentifierLength, nil
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
	return hexutil.EncodeHex(id[:])
}

func (id SlotIdentifier) String() string {
	return fmt.Sprintf("%s:%d", id.Alias(), id.Index())
}

// TODO: rename to Slot?
func (id SlotIdentifier) Index() SlotIndex {
	return SlotIndex(binary.LittleEndian.Uint32(id[IdentifierLength:]))
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

	return id.ToHex()
}

// UnregisterAlias allows to unregister a previously registered alias.
func (id SlotIdentifier) UnregisterAlias() {
	slotIdentifierAliasesMutex.Lock()
	defer slotIdentifierAliasesMutex.Unlock()

	delete(slotIidentifierAliases, id)
}
