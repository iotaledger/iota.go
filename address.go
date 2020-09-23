package iota

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"golang.org/x/crypto/blake2b"
)

// Defines the type of addresses.
type AddressType = byte

const (
	// Denotes a WOTS address.
	AddressWOTS AddressType = iota
	// Denotes a Ed25510 address.
	AddressEd25519

	// The length of a WOTS address.
	WOTSAddressBytesLength = 49
	// The size of a serialized WOTS address with its type denoting byte.
	WOTSAddressSerializedBytesSize = SmallTypeDenotationByteSize + WOTSAddressBytesLength

	// The length of a Ed25519 address
	Ed25519AddressBytesLength = blake2b.Size256
	// The size of a serialized Ed25519 address with its type denoting byte.
	Ed25519AddressSerializedBytesSize = SmallTypeDenotationByteSize + Ed25519AddressBytesLength
)

// AddressSelector implements SerializableSelectorFunc for address types.
func AddressSelector(typeByte uint32) (Serializable, error) {
	var seri Serializable
	switch byte(typeByte) {
	case AddressWOTS:
		seri = &WOTSAddress{}
	case AddressEd25519:
		seri = &Ed25519Address{}
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownAddrType, typeByte)
	}
	return seri, nil
}

// Defines a WOTS address.
type WOTSAddress [WOTSAddressBytesLength]byte

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
		// TODO: check T5B1 encoding
	}
	copy(wotsAddr[:], data[SmallTypeDenotationByteSize:])
	return WOTSAddressSerializedBytesSize, nil
}

func (wotsAddr *WOTSAddress) Serialize(deSeriMode DeSerializationMode) (data []byte, err error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		// TODO: check T5B1 encoding
	}
	var b [WOTSAddressSerializedBytesSize]byte
	b[0] = AddressWOTS
	copy(b[SmallTypeDenotationByteSize:], wotsAddr[:])
	return b[:], nil
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
	addr := &Ed25519Address{}
	addrBytes, err := hex.DecodeString(j.Address)
	if err != nil {
		return nil, fmt.Errorf("unable to decode address from JSON for Ed25519 address: %w", err)
	}
	copy(addr[:], addrBytes)
	return addr, nil
}

// jsonwotsaddress defines the json representation of a WOTSAddress.
type jsonwotsaddress struct {
	Type    int    `json:"type"`
	Address string `json:"address"`
}

func (j *jsonwotsaddress) ToSerializable() (Serializable, error) {
	addr := &WOTSAddress{}
	addrBytes, err := hex.DecodeString(j.Address)
	if err != nil {
		return nil, fmt.Errorf("unable to decode address from JSON for WOTS address: %w", err)
	}
	copy(addr[:], addrBytes)
	return addr, nil
}
