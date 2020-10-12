package iota

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/iotaledger/iota.go/bech32"
	"github.com/iotaledger/iota.go/encoding/t5b1"
	"github.com/iotaledger/iota.go/legacy/trinary"
	"golang.org/x/crypto/blake2b"
)

// Defines the type of addresses.
type AddressType = byte

const (
	// Denotes a WOTS address.
	AddressWOTS AddressType = iota
	// Denotes a Ed25510 address.
	AddressEd25519
)

// NetworkPrefix denotes the different network prefixes.
type NetworkPrefix int

// Network prefix options
const (
	PrefixMainnet NetworkPrefix = iota
	PrefixTestnet
)

const (
	// The length of a WOTS address.
	WOTSAddressBytesLength = 49 // t5b1.EncodedLen(legacy.HashTrytesSize)
	// The size of a serialized WOTS address with its type denoting byte.
	WOTSAddressSerializedBytesSize = SmallTypeDenotationByteSize + WOTSAddressBytesLength

	// The length of a Ed25519 address
	Ed25519AddressBytesLength = blake2b.Size256
	// The size of a serialized Ed25519 address with its type denoting byte.
	Ed25519AddressSerializedBytesSize = SmallTypeDenotationByteSize + Ed25519AddressBytesLength
)

func (p NetworkPrefix) String() string {
	return hrpStrings[p]
}

// ParsePrefix parses the string and returns the corresponding NetworkPrefix.
func ParsePrefix(s string) (NetworkPrefix, error) {
	for i := range hrpStrings {
		if s == hrpStrings[i] {
			return NetworkPrefix(i), nil
		}
	}
	return 0, fmt.Errorf("%w: prefix %s", ErrUnknownNetworkPrefix, s)
}

var (
	hrpStrings = [...]string{"iot", "tio"}
)

// Address describes a general address.
type Address interface {
	Serializable

	// Type returns the type of the address.
	Type() AddressType
	// Bech32 encodes the address as a bech32 string.
	Bech32(hrp NetworkPrefix) string

	String() string
}

// AddressSelector implements SerializableSelectorFunc for address types.
func AddressSelector(addressType uint32) (Serializable, error) {
	return newAddress(byte(addressType))
}

func newAddress(addressType byte) (address Address, err error) {
	switch addressType {
	case AddressWOTS:
		return &WOTSAddress{}, nil
	case AddressEd25519:
		return &Ed25519Address{}, nil
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownAddrType, addressType)
	}
}

func bech32String(hrp NetworkPrefix, addr Address) string {
	bytes, _ := addr.Serialize(DeSeriModeNoValidation)
	s, err := bech32.Encode(hrp.String(), bytes)
	if err != nil {
		panic(err)
	}
	return s
}

// ParseBech32 decodes a bech32 encoded string.
func ParseBech32(s string) (NetworkPrefix, Address, error) {
	hrp, addrData, err := bech32.Decode(s)
	if err != nil {
		return 0, nil, fmt.Errorf("invalid bech32 encoding: %w", err)
	}
	prefix, err := ParsePrefix(hrp)
	if err != nil {
		return 0, nil, fmt.Errorf("invalid human-readable prefix: %w", err)
	}
	if len(addrData) == 0 {
		return 0, nil, ErrDeserializationNotEnoughData
	}

	addr, err := newAddress(addrData[0])
	if err != nil {
		return 0, nil, err
	}
	n, err := addr.Deserialize(addrData, DeSeriModePerformValidation)
	if err != nil {
		return 0, nil, err
	}
	if n != len(addrData) {
		return 0, nil, ErrDeserializationNotAllConsumed
	}
	return prefix, addr, nil
}

// Defines a WOTS address.
// TODO: it could make more sense to store WOTS addresses as tryte arrays.
type WOTSAddress [WOTSAddressBytesLength]byte

// WOTSAddressFromTrytes creates a new WOTSAddress from trytes.
func WOTSAddressFromTrytes(trytes trinary.Trytes) *WOTSAddress {
	addrBytes := t5b1.EncodeTrytes(trytes)
	addr := &WOTSAddress{}
	copy(addr[:], addrBytes)
	return addr
}

func (wotsAddr *WOTSAddress) Type() AddressType {
	return AddressWOTS
}

func (wotsAddr *WOTSAddress) Bech32(hrp NetworkPrefix) string {
	return bech32String(hrp, wotsAddr)
}

func (wotsAddr *WOTSAddress) String() string {
	return hex.EncodeToString(wotsAddr[:])
}

func (wotsAddr *WOTSAddress) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkMinByteLength(WOTSAddressSerializedBytesSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid WOTS address bytes: %w", err)
		}
		if err := checkTypeByte(data, AddressWOTS); err != nil {
			return 0, fmt.Errorf("unable to deserialize WOTS address: %w", err)
		}
		if _, err := t5b1.DecodeToTrytes(data[SmallTypeDenotationByteSize:]); err != nil {
			return 0, fmt.Errorf("invalid WOTS address bytes: %w", err)
		}
	}
	copy(wotsAddr[:], data[SmallTypeDenotationByteSize:])
	return WOTSAddressSerializedBytesSize, nil
}

func (wotsAddr *WOTSAddress) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		_, err := t5b1.DecodeToTrytes(wotsAddr[:])
		if err != nil {
			return nil, fmt.Errorf("invalid WOTS address bytes: %w", err)
		}
	}
	return append([]byte{wotsAddr.Type()}, wotsAddr[:]...), nil
}

func (wotsAddr *WOTSAddress) MarshalJSON() ([]byte, error) {
	jsonAddr := &jsonwotsaddress{}
	jsonAddr.Address = hex.EncodeToString(wotsAddr[:])
	jsonAddr.Type = int(AddressWOTS)
	return json.Marshal(jsonAddr)
}

func (wotsAddr *WOTSAddress) UnmarshalJSON(bytes []byte) error {
	jsonAddr := &jsonwotsaddress{}
	if err := json.Unmarshal(bytes, jsonAddr); err != nil {
		return err
	}
	seri, err := jsonAddr.ToSerializable()
	if err != nil {
		return err
	}
	*wotsAddr = *seri.(*WOTSAddress)
	return nil
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
	jsonAddr := &jsoned25519{}
	jsonAddr.Address = hex.EncodeToString(edAddr[:])
	jsonAddr.Type = int(AddressEd25519)
	return json.Marshal(jsonAddr)
}

func (edAddr *Ed25519Address) UnmarshalJSON(bytes []byte) error {
	jsonAddr := &jsoned25519{}
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
func jsonaddressselector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch byte(ty) {
	case AddressEd25519:
		obj = &jsoned25519{}
	case AddressWOTS:
		obj = &jsonwotsaddress{}
	default:
		return nil, fmt.Errorf("unable to decode address type from JSON: %w", ErrUnknownAddrType)
	}
	return obj, nil
}

// jsoned25519 defines the json representation of an Ed25519Address.
type jsoned25519 struct {
	Type    int    `json:"type"`
	Address string `json:"address"`
}

func (j *jsoned25519) ToSerializable() (Serializable, error) {
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

// jsonwotsaddress defines the json representation of a WOTSAddress.
type jsonwotsaddress struct {
	Type    int    `json:"type"`
	Address string `json:"address"`
}

func (j *jsonwotsaddress) ToSerializable() (Serializable, error) {
	addrBytes, err := hex.DecodeString(j.Address)
	if err != nil {
		return nil, fmt.Errorf("unable to decode address from JSON for WOTS address: %w", err)
	}
	if err := checkExactByteLength(len(addrBytes), WOTSAddressBytesLength); err != nil {
		return nil, fmt.Errorf("unable to decode address from JSON for WOTS address: %w", err)
	}
	addr := &WOTSAddress{}
	copy(addr[:], addrBytes)
	return addr, nil
}
