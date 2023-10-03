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
	SignedTransactionIDLength = IdentifierLength + SlotIndexLength
)

var (
	emptySignedTransactionID = SignedTransactionID{}
)

// SignedTransactionID is a 32 byte hash value that can be used to uniquely identify some blob of data together with an 4 byte slot index.
type SignedTransactionID [SignedTransactionIDLength]byte

// SignedTransactionIDRepresentingData returns a new SignedTransactionID for the given data by hashing it with blake2b and associating it with the given slot index.
func SignedTransactionIDRepresentingData(slot SlotIndex, data []byte) SignedTransactionID {
	return NewSignedTransactionID(slot, blake2b.Sum256(data))
}

func NewSignedTransactionID(slot SlotIndex, idBytes Identifier) SignedTransactionID {
	id := SignedTransactionID{}
	copy(id[:], idBytes[:])
	binary.LittleEndian.PutUint32(id[IdentifierLength:], uint32(slot))

	return id
}

// SignedTransactionIDFromHexString converts the hex to a SignedTransactionID representation.
func SignedTransactionIDFromHexString(hex string) (SignedTransactionID, error) {
	bytes, err := hexutil.DecodeHex(hex)
	if err != nil {
		return SignedTransactionID{}, err
	}

	s, _, err := SignedTransactionIDFromBytes(bytes)

	return s, err
}

// SignedTransactionIDFromBytes returns a new SignedTransactionID represented by the passed bytes.
func SignedTransactionIDFromBytes(bytes []byte) (SignedTransactionID, int, error) {
	if len(bytes) < SignedTransactionIDLength {
		return SignedTransactionID{}, 0, ErrInvalidIdentifierLength
	}

	return SignedTransactionID(bytes), SignedTransactionIDLength, nil
}

// MustSignedTransactionIDFromHexString converts the hex to a SignedTransactionID representation.
func MustSignedTransactionIDFromHexString(hex string) SignedTransactionID {
	id, err := SignedTransactionIDFromHexString(hex)
	if err != nil {
		panic(err)
	}

	return id
}

func (id SignedTransactionID) Bytes() ([]byte, error) {
	return id[:], nil
}

func (id SignedTransactionID) MarshalText() (text []byte, err error) {
	dst := make([]byte, hex.EncodedLen(len(SignedTransactionID{})))
	hex.Encode(dst, id[:])

	return dst, nil
}

func (id *SignedTransactionID) UnmarshalText(text []byte) error {
	_, err := hex.Decode(id[:], text)

	return err
}

// Empty tells whether the Identifier is empty.
func (id SignedTransactionID) Empty() bool {
	return id == emptySignedTransactionID
}

// ToHex converts the Identifier to its hex representation.
func (id SignedTransactionID) ToHex() string {
	return hexutil.EncodeHex(id[:])
}

func (id SignedTransactionID) String() string {
	return fmt.Sprintf("%s:%d", id.Alias(), id.Slot())
}

func (id SignedTransactionID) Slot() SlotIndex {
	return SlotIndex(binary.LittleEndian.Uint32(id[IdentifierLength:]))
}

// Index returns a slot index to conform with hive's IndexedID interface.
func (id SignedTransactionID) Index() SlotIndex {
	return id.Slot()
}

func (id SignedTransactionID) Identifier() Identifier {
	return Identifier(id[:IdentifierLength])
}

var (
	// signedTransactionIDAliases contains a dictionary of identifiers associated to their human-readable alias.
	signedTransactionIDAliases = make(map[SignedTransactionID]string)

	// signedTransactionIDAliasesMutex is the mutex that is used to synchronize access to the previous map.
	signedTransactionIDAliasesMutex = sync.RWMutex{}
)

// RegisterAlias allows to register a human-readable alias for the Identifier which will be used as a replacement for
// the String method.
func (id SignedTransactionID) RegisterAlias(alias string) {
	signedTransactionIDAliasesMutex.Lock()
	defer signedTransactionIDAliasesMutex.Unlock()

	signedTransactionIDAliases[id] = alias
}

// Alias returns the human-readable alias of the Identifier (or the base58 encoded bytes of no alias was set).
func (id SignedTransactionID) Alias() (alias string) {
	signedTransactionIDAliasesMutex.RLock()
	defer signedTransactionIDAliasesMutex.RUnlock()

	if existingAlias, exists := signedTransactionIDAliases[id]; exists {
		return existingAlias
	}

	return id.ToHex()
}

// UnregisterAlias allows to unregister a previously registered alias.
func (id SignedTransactionID) UnregisterAlias() {
	signedTransactionIDAliasesMutex.Lock()
	defer signedTransactionIDAliasesMutex.Unlock()

	delete(signedTransactionIDAliases, id)
}

var EmptySignedTransactionID = SignedTransactionID{}

// SignedTransactionIDs are IDs of signed transactions.
type SignedTransactionIDs []SignedTransactionID
