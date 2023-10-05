package iotago

// Code generated by go generate; DO NOT EDIT. Check gen/ directory instead.

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

const (

	// OutputIndexLength defines the length of an OutputIndex.
	OutputIndexLength = serializer.UInt16ByteSize
	// OutputIDLength defines the length of an OutputID.
	OutputIDLength = TransactionIDLength + OutputIndexLength
)

var (
	ErrInvalidOutputIDLength = ierrors.New("invalid outputID length")

	EmptyOutputID = OutputID{}
)

// OutputID is a 32 byte hash value together with an 4 byte slot index.
type OutputID [OutputIDLength]byte

// OutputIDRepresentingData returns a new OutputID for the given data by hashing it with blake2b and associating it with the given slot index.
func OutputIDRepresentingData(slot SlotIndex, data []byte) OutputID {
	return NewOutputID(slot, blake2b.Sum256(data))
}

func NewOutputID(slot SlotIndex, idBytes Identifier) OutputID {
	o := OutputID{}
	copy(o[:], idBytes[:])
	binary.LittleEndian.PutUint32(o[IdentifierLength:], uint32(slot))

	return o
}

// OutputIDFromHexString converts the hex to a OutputID representation.
func OutputIDFromHexString(hex string) (OutputID, error) {
	b, err := hexutil.DecodeHex(hex)
	if err != nil {
		return OutputID{}, err
	}

	s, _, err := OutputIDFromBytes(b)

	return s, err
}

// OutputIDFromBytes returns a new OutputID represented by the passed bytes.
func OutputIDFromBytes(b []byte) (OutputID, int, error) {
	if len(b) < OutputIDLength {
		return OutputID{}, 0, ErrInvalidOutputIDLength
	}

	return OutputID(b), OutputIDLength, nil
}

// MustOutputIDFromHexString converts the hex to a OutputID representation.
func MustOutputIDFromHexString(hex string) OutputID {
	o, err := OutputIDFromHexString(hex)
	if err != nil {
		panic(err)
	}

	return o
}

func (o OutputID) Bytes() ([]byte, error) {
	return o[:], nil
}

func (o OutputID) MarshalText() (text []byte, err error) {
	dst := make([]byte, hex.EncodedLen(len(OutputID{})))
	hex.Encode(dst, o[:])

	return dst, nil
}

func (o *OutputID) UnmarshalText(text []byte) error {
	_, err := hex.Decode(o[:], text)

	return err
}

// Empty tells whether the OutputID is empty.
func (o OutputID) Empty() bool {
	return o == EmptyOutputID
}

// ToHex converts the Identifier to its hex representation.
func (o OutputID) ToHex() string {
	return hexutil.EncodeHex(o[:])
}

func (o OutputID) String() string {
	return fmt.Sprintf("OutputID(%s:%d)", o.Alias(), o.Slot())
}

func (o OutputID) Slot() SlotIndex {
	return SlotIndex(binary.LittleEndian.Uint32(o[IdentifierLength:]))
}

// Index returns the index of the Output this OutputID references.
func (outputID OutputID) Index() uint16 {
	return binary.LittleEndian.Uint16(outputID[TransactionIDLength:])
}

func (o OutputID) Identifier() Identifier {
	return Identifier(o[:IdentifierLength])
}

var (
	// OutputIDAliases contains a dictionary of identifiers associated to their human-readable alias.
	OutputIDAliases = make(map[OutputID]string)

	// OutputIDAliasesMutex is the mutex that is used to synchronize access to the previous map.
	OutputIDAliasesMutex = sync.RWMutex{}
)

// RegisterAlias allows to register a human-readable alias for the Identifier which will be used as a replacement for
// the String method.
func (o OutputID) RegisterAlias(alias string) {
	OutputIDAliasesMutex.Lock()
	defer OutputIDAliasesMutex.Unlock()

	OutputIDAliases[o] = alias
}

// Alias returns the human-readable alias of the Identifier (or the base58 encoded bytes of no alias was set).
func (o OutputID) Alias() (alias string) {
	OutputIDAliasesMutex.RLock()
	defer OutputIDAliasesMutex.RUnlock()

	if existingAlias, exists := OutputIDAliases[o]; exists {
		return existingAlias
	}

	return o.ToHex()
}

// UnregisterAlias allows to unregister a previously registered alias.
func (o OutputID) UnregisterAlias() {
	OutputIDAliasesMutex.Lock()
	defer OutputIDAliasesMutex.Unlock()

	delete(OutputIDAliases, o)
}

type OutputIDs []OutputID

// ToHex converts the OutputIDs to their hex representation.
func (ids OutputIDs) ToHex() []string {
	hexIDs := make([]string, len(ids))
	for i, o := range ids {
		hexIDs[i] = hexutil.EncodeHex(o[:])
	}

	return hexIDs
}

// RemoveDupsAndSort removes duplicated OutputIDs and sorts the slice by the lexical ordering.
func (ids OutputIDs) RemoveDupsAndSort() OutputIDs {
	sorted := append(OutputIDs{}, ids...)
	sort.Slice(sorted, func(i, j int) bool {
		return bytes.Compare(sorted[i][:], sorted[j][:]) == -1
	})

	var result OutputIDs
	var prev OutputID
	for i, o := range sorted {
		if i == 0 || !bytes.Equal(prev[:], o[:]) {
			result = append(result, o)
		}
		prev = o
	}

	return result
}

// OutputIDsFromHexString converts the given block IDs from their hex to OutputID representation.
func OutputIDsFromHexString(OutputIDsHex []string) (OutputIDs, error) {
	result := make(OutputIDs, len(OutputIDsHex))

	for i, hexString := range OutputIDsHex {
		OutputID, err := OutputIDFromHexString(hexString)
		if err != nil {
			return nil, err
		}
		result[i] = OutputID
	}

	return result, nil
}
