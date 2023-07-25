package iotago

import (
	"encoding/hex"
	"sync"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

const (
	// IdentifierLength defines the length of an Identifier.
	IdentifierLength = blake2b.Size256
)

var (
	EmptyIdentifier = Identifier{}

	ErrInvalidIdentifierLength = ierrors.New("Invalid identifier length")
)

// Identifier is a 32 byte hash value that can be used to uniquely identify some blob of data.
type Identifier [IdentifierLength]byte

// IdentifierFromData returns a new Identifier for the given data by hashing it with blake2b.
func IdentifierFromData(data []byte) Identifier {
	return blake2b.Sum256(data)
}

// IdentifierFromHexString converts the hex to an Identifier representation.
func IdentifierFromHexString(hex string) (Identifier, error) {
	bytes, err := hexutil.DecodeHex(hex)
	if err != nil {
		return Identifier{}, err
	}

	id, _, err := IdentifierFromBytes(bytes)
	return id, err
}

// MustIdentifierFromHexString converts the hex to an Identifier representation.
func MustIdentifierFromHexString(hex string) Identifier {
	id, err := IdentifierFromHexString(hex)
	if err != nil {
		panic(err)
	}

	return id
}

func IdentifierFromBytes(bytes []byte) (Identifier, int, error) {
	var id Identifier
	if len(bytes) < IdentifierLength {
		return id, 0, ErrInvalidIdentifierLength
	}
	copy(id[:], bytes)
	return id, len(bytes), nil
}

func (id Identifier) Bytes() ([]byte, error) {
	return id[:], nil
}

func (id Identifier) MarshalText() (text []byte, err error) {
	dst := make([]byte, hex.EncodedLen(len(Identifier{})))
	hex.Encode(dst, id[:])
	return dst, nil
}

func (id *Identifier) UnmarshalText(text []byte) error {
	_, err := hex.Decode(id[:], text)
	return err
}

// Empty tells whether the Identifier is empty.
func (id Identifier) Empty() bool {
	return id == EmptyIdentifier
}

// ToHex converts the Identifier to its hex representation.
func (id Identifier) ToHex() string {
	return hexutil.EncodeHex(id[:])
}

func (id Identifier) String() string {
	return id.Alias()
}

var (
	// identifierAliases contains a dictionary of identifiers associated to their human-readable alias.
	identifierAliases = make(map[Identifier]string)

	// identifierAliasesMutex is the mutex that is used to synchronize access to the previous map.
	identifierAliasesMutex = sync.RWMutex{}
)

// RegisterAlias allows to register a human-readable alias for the Identifier which will be used as a replacement for
// the String method.
func (id Identifier) RegisterAlias(alias string) {
	identifierAliasesMutex.Lock()
	defer identifierAliasesMutex.Unlock()

	identifierAliases[id] = alias
}

// Alias returns the human-readable alias of the Identifier (or the base58 encoded bytes of no alias was set).
func (id Identifier) Alias() (alias string) {
	identifierAliasesMutex.RLock()
	defer identifierAliasesMutex.RUnlock()

	if existingAlias, exists := identifierAliases[id]; exists {
		return existingAlias
	}

	return id.ToHex()
}

// UnregisterAlias allows to unregister a previously registered alias.
func (id Identifier) UnregisterAlias() {
	identifierAliasesMutex.Lock()
	defer identifierAliasesMutex.Unlock()

	delete(identifierAliases, id)
}

// UnregisterIdentifierAliases allows to unregister all previously registered aliases.
func UnregisterIdentifierAliases() {
	identifierAliasesMutex.Lock()
	defer identifierAliasesMutex.Unlock()

	identifierAliases = make(map[Identifier]string)
}
