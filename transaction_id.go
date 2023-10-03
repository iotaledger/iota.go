//nolint:dupl
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
	TransactionIDLength = IdentifierLength + SlotIndexLength
)

var (
	emptyTransactionID = TransactionID{}
)

// TransactionID is a 32 byte hash value that can be used to uniquely identify some blob of data together with an 4 byte slot index.
type TransactionID [TransactionIDLength]byte

// TransactionIDRepresentingData returns a new TransactionID for the given data by hashing it with blake2b and associating it with the given slot index.
func TransactionIDRepresentingData(slot SlotIndex, data []byte) TransactionID {
	return NewTransactionID(slot, blake2b.Sum256(data))
}

func NewTransactionID(slot SlotIndex, idBytes Identifier) TransactionID {
	id := TransactionID{}
	copy(id[:], idBytes[:])
	binary.LittleEndian.PutUint32(id[IdentifierLength:], uint32(slot))

	return id
}

// TransactionIDFromHexString converts the hex to a TransactionID representation.
func TransactionIDFromHexString(hex string) (TransactionID, error) {
	bytes, err := hexutil.DecodeHex(hex)
	if err != nil {
		return TransactionID{}, err
	}

	s, _, err := TransactionIDFromBytes(bytes)

	return s, err
}

// TransactionIDFromBytes returns a new TransactionID represented by the passed bytes.
func TransactionIDFromBytes(bytes []byte) (TransactionID, int, error) {
	if len(bytes) < TransactionIDLength {
		return TransactionID{}, 0, ErrInvalidIdentifierLength
	}

	return TransactionID(bytes), TransactionIDLength, nil
}

// MustTransactionIDFromHexString converts the hex to a TransactionID representation.
func MustTransactionIDFromHexString(hex string) TransactionID {
	id, err := TransactionIDFromHexString(hex)
	if err != nil {
		panic(err)
	}

	return id
}

func (id TransactionID) Bytes() ([]byte, error) {
	return id[:], nil
}

func (id TransactionID) MarshalText() (text []byte, err error) {
	dst := make([]byte, hex.EncodedLen(len(TransactionID{})))
	hex.Encode(dst, id[:])

	return dst, nil
}

func (id *TransactionID) UnmarshalText(text []byte) error {
	_, err := hex.Decode(id[:], text)

	return err
}

// Empty tells whether the Identifier is empty.
func (id TransactionID) Empty() bool {
	return id == emptyTransactionID
}

// ToHex converts the Identifier to its hex representation.
func (id TransactionID) ToHex() string {
	return hexutil.EncodeHex(id[:])
}

func (id TransactionID) String() string {
	return fmt.Sprintf("%s:%d", id.Alias(), id.Slot())
}

func (id TransactionID) Slot() SlotIndex {
	return SlotIndex(binary.LittleEndian.Uint32(id[IdentifierLength:]))
}

// Index returns a slot index to conform with hive's IndexedID interface.
func (id TransactionID) Index() SlotIndex {
	return id.Slot()
}

func (id TransactionID) Identifier() Identifier {
	return Identifier(id[:IdentifierLength])
}

var (
	// transactionIDAliases contains a dictionary of identifiers associated to their human-readable alias.
	transactionIDAliases = make(map[TransactionID]string)

	// transactionIDAliasesMutex is the mutex that is used to synchronize access to the previous map.
	transactionIDAliasesMutex = sync.RWMutex{}
)

// RegisterAlias allows to register a human-readable alias for the Identifier which will be used as a replacement for
// the String method.
func (id TransactionID) RegisterAlias(alias string) {
	transactionIDAliasesMutex.Lock()
	defer transactionIDAliasesMutex.Unlock()

	transactionIDAliases[id] = alias
}

// Alias returns the human-readable alias of the Identifier (or the base58 encoded bytes of no alias was set).
func (id TransactionID) Alias() (alias string) {
	transactionIDAliasesMutex.RLock()
	defer transactionIDAliasesMutex.RUnlock()

	if existingAlias, exists := transactionIDAliases[id]; exists {
		return existingAlias
	}

	return id.ToHex()
}

// UnregisterAlias allows to unregister a previously registered alias.
func (id TransactionID) UnregisterAlias() {
	transactionIDAliasesMutex.Lock()
	defer transactionIDAliasesMutex.Unlock()

	delete(transactionIDAliases, id)
}

var EmptyTransactionID = TransactionID{}

// TransactionIDs are IDs of transactions.
type TransactionIDs []TransactionID
