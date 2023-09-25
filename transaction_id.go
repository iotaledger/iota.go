package iotago

import (
	"encoding/binary"
	"encoding/hex"
	"sync"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

const (
	// TransactionIDLength defines the length of a Transaction ID.
	TransactionIDLength = SlotIndexLength + IdentifierLength
)

var (
	EmptyTransactionID = TransactionID{}

	ErrInvalidTransactionIDLength = ierrors.New("Invalid transactionID length")
)

// TransactionID is the ID of a Transaction.
type TransactionID [TransactionIDLength]byte

// TransactionIDs are IDs of transactions.
type TransactionIDs []TransactionID

// TransactionIDFromData returns a new TransactionID for the given data by hashing it with blake2b.
func TransactionIDFromData(creationSlot SlotIndex, data []byte) TransactionID {
	dataHash := blake2b.Sum256(data)

	var txID TransactionID
	binary.LittleEndian.PutUint32(txID[:SlotIndexLength], uint32(creationSlot))
	copy(txID[SlotIndexLength:], dataHash[:])

	return txID
}

// TransactionIDFromHexString converts the hex to an TransactionID representation.
func TransactionIDFromHexString(hex string) (TransactionID, error) {
	bytes, err := hexutil.DecodeHex(hex)
	if err != nil {
		return TransactionID{}, err
	}

	id, _, err := TransactionIDFromBytes(bytes)

	return id, err
}

// MustTransactionIDFromHexString converts the hex to an TransactionID representation.
func MustTransactionIDFromHexString(hex string) TransactionID {
	id, err := TransactionIDFromHexString(hex)
	if err != nil {
		panic(err)
	}

	return id
}

func TransactionIDFromBytes(bytes []byte) (TransactionID, int, error) {
	var id TransactionID
	if len(bytes) < TransactionIDLength {
		return id, 0, ErrInvalidTransactionIDLength
	}
	copy(id[:], bytes)

	return id, len(bytes), nil
}

// CreationSlotIndex returns the SlotIndex the Transaction was created in.
func (id TransactionID) CreationSlotIndex() SlotIndex {
	return SlotIndex(binary.LittleEndian.Uint32(id[:OutputIndexLength]))
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

// Empty tells whether the TransactionID is empty.
func (id TransactionID) Empty() bool {
	return id == EmptyTransactionID
}

// ToHex converts the TransactionID to its hex representation.
func (id TransactionID) ToHex() string {
	return hexutil.EncodeHex(id[:])
}

func (id TransactionID) String() string {
	return id.Alias()
}

var (
	// TransactionIDAliases contains a dictionary of TransactionIDs associated to their human-readable alias.
	TransactionIDAliases = make(map[TransactionID]string)

	// TransactionIDAliasesMutex is the mutex that is used to synchronize access to the previous map.
	TransactionIDAliasesMutex = sync.RWMutex{}
)

// RegisterAlias allows to register a human-readable alias for the TransactionID which will be used as a replacement for
// the String method.
func (id TransactionID) RegisterAlias(alias string) {
	TransactionIDAliasesMutex.Lock()
	defer TransactionIDAliasesMutex.Unlock()

	TransactionIDAliases[id] = alias
}

// Alias returns the human-readable alias of the TransactionID (or the base58 encoded bytes of no alias was set).
func (id TransactionID) Alias() (alias string) {
	TransactionIDAliasesMutex.RLock()
	defer TransactionIDAliasesMutex.RUnlock()

	if existingAlias, exists := TransactionIDAliases[id]; exists {
		return existingAlias
	}

	return id.ToHex()
}

// UnregisterAlias allows to unregister a previously registered alias.
func (id TransactionID) UnregisterAlias() {
	TransactionIDAliasesMutex.Lock()
	defer TransactionIDAliasesMutex.Unlock()

	delete(TransactionIDAliases, id)
}

// UnregisterTransactionIDAliases allows to unregister all previously registered aliases.
func UnregisterTransactionIDAliases() {
	TransactionIDAliasesMutex.Lock()
	defer TransactionIDAliasesMutex.Unlock()

	TransactionIDAliases = make(map[TransactionID]string)
}
