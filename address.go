package iotago

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/iotaledger/iota.go/v2/bech32"
	"github.com/iotaledger/iota.go/v2/ed25519"
	"golang.org/x/crypto/blake2b"
)

// Defines the type of addresses.
type AddressType = byte

const (
	// Denotes an Ed25519 address.
	AddressEd25519 AddressType = iota
)

// NetworkPrefix denotes the different network prefixes.
type NetworkPrefix string

// Network prefixes.
const (
	PrefixMainnet NetworkPrefix = "iota"
	PrefixTestnet NetworkPrefix = "atoi"
)

const (
	// The length of an Ed25519 address
	Ed25519AddressBytesLength = blake2b.Size256
	// The size of a serialized Ed25519 address with its type denoting byte.
	Ed25519AddressSerializedBytesSize = SmallTypeDenotationByteSize + Ed25519AddressBytesLength
)

// Address describes a general address.
type Address interface {
	Serializable
	fmt.Stringer

	// Type returns the type of the address.
	Type() AddressType

	// Bech32 encodes the address as a bech32 string.
	Bech32(hrp NetworkPrefix) string
}

// AddressSelector implements SerializableSelectorFunc for address types.
func AddressSelector(addressType uint32) (Serializable, error) {
	return newAddress(byte(addressType))
}

func newAddress(addressType byte) (address Address, err error) {
	switch addressType {
	case AddressEd25519:
		return &Ed25519Address{}, nil
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownAddrType, addressType)
	}
}

func bech32String(hrp NetworkPrefix, addr Address) string {
	bytes, _ := addr.Serialize(DeSeriModeNoValidation)
	s, err := bech32.Encode(string(hrp), bytes)
	if err != nil {
		panic(err)
	}
	return s
}

// ParseBech32 decodes a bech32 encoded string.
func ParseBech32(s string) (NetworkPrefix, Address, error) {
	hrp, addrData, err := bech32.Decode(s)
	if err != nil {
		return "", nil, fmt.Errorf("invalid bech32 encoding: %w", err)
	}

	if len(addrData) == 0 {
		return "", nil, ErrDeserializationNotEnoughData
	}

	addr, err := newAddress(addrData[0])
	if err != nil {
		return "", nil, err
	}

	n, err := addr.Deserialize(addrData, DeSeriModePerformValidation)
	if err != nil {
		return "", nil, err
	}

	if n != len(addrData) {
		return "", nil, ErrDeserializationNotAllConsumed
	}

	return NetworkPrefix(hrp), addr, nil
}

// Defines an Ed25519 address.
type Ed25519Address [Ed25519AddressBytesLength]byte

func (edAddr *Ed25519Address) Type() AddressType {
	return AddressEd25519
}

func (edAddr *Ed25519Address) Bech32(hrp NetworkPrefix) string {
	return bech32String(hrp, edAddr)
}

func (edAddr *Ed25519Address) String() string {
	return hex.EncodeToString(edAddr[:])
}

func (edAddr *Ed25519Address) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkMinByteLength(Ed25519AddressSerializedBytesSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid Ed25519 address bytes: %w", err)
		}
		if err := checkTypeByte(data, AddressEd25519); err != nil {
			return 0, fmt.Errorf("unable to deserialize Ed25519 address: %w", err)
		}
	}
	copy(edAddr[:], data[SmallTypeDenotationByteSize:])
	return Ed25519AddressSerializedBytesSize, nil
}

func (edAddr *Ed25519Address) Serialize(deSeriMode DeSerializationMode) (data []byte, err error) {
	var b [Ed25519AddressSerializedBytesSize]byte
	b[0] = AddressEd25519
	copy(b[SmallTypeDenotationByteSize:], edAddr[:])
	return b[:], nil
}

func (edAddr *Ed25519Address) MarshalJSON() ([]byte, error) {
	jsonAddr := &jsonEd25519Address{}
	jsonAddr.Address = hex.EncodeToString(edAddr[:])
	jsonAddr.Type = int(AddressEd25519)
	return json.Marshal(jsonAddr)
}

func (edAddr *Ed25519Address) UnmarshalJSON(bytes []byte) error {
	jsonAddr := &jsonEd25519Address{}
	if err := json.Unmarshal(bytes, jsonAddr); err != nil {
		return err
	}
	seri, err := jsonAddr.ToSerializable()
	if err != nil {
		return err
	}
	*edAddr = *seri.(*Ed25519Address)
	return nil
}

// AddressFromEd25519PubKey returns the address belonging to the given Ed25519 public key.
func AddressFromEd25519PubKey(pubKey ed25519.PublicKey) Ed25519Address {
	return blake2b.Sum256(pubKey[:])
}

// selects the json object for the given type.
func jsonAddressSelector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch byte(ty) {
	case AddressEd25519:
		obj = &jsonEd25519Address{}
	default:
		return nil, fmt.Errorf("unable to decode address type from JSON: %w", ErrUnknownAddrType)
	}
	return obj, nil
}

// jsonEd25519Address defines the json representation of an Ed25519Address.
type jsonEd25519Address struct {
	Type    int    `json:"type"`
	Address string `json:"address"`
}

func (j *jsonEd25519Address) ToSerializable() (Serializable, error) {
	addrBytes, err := hex.DecodeString(j.Address)
	if err != nil {
		return nil, fmt.Errorf("unable to decode address from JSON for Ed25519 address: %w", err)
	}
	if err := checkExactByteLength(len(addrBytes), Ed25519AddressBytesLength); err != nil {
		return nil, fmt.Errorf("unable to decode address from JSON for Ed25519 address: %w", err)
	}
	addr := &Ed25519Address{}
	copy(addr[:], addrBytes)
	return addr, nil
}
