package iotago

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer"
	"github.com/iotaledger/iota.go/v3/bech32"
)

// AddressType defines the type of addresses.
type AddressType = byte

const (
	// AddressEd25519 denotes an Ed25519 address.
	AddressEd25519 = 0
	// AddressBLS denotes a BLS address.
	AddressBLS = 1
	// AddressAlias denotes an Alias address.
	AddressAlias = 8
	// AddressNFT denotes an NFT address.
	AddressNFT = 16
)

var (
	// ErrTypeIsNotAddress gets returned when a serializable was found to not be an Address.
	ErrTypeIsNotAddress = errors.New("serializable is not an address")
)

// NetworkPrefix denotes the different network prefixes.
type NetworkPrefix string

// Network prefixes.
const (
	PrefixMainnet NetworkPrefix = "iota"
	PrefixTestnet NetworkPrefix = "atoi"
)

// Address describes a general address.
type Address interface {
	serializer.Serializable
	fmt.Stringer

	// Type returns the type of the address.
	Type() AddressType

	// Bech32 encodes the address as a bech32 string.
	Bech32(hrp NetworkPrefix) string
}

// AddressSelector implements SerializableSelectorFunc for address types.
func AddressSelector(addressType uint32) (serializer.Serializable, error) {
	return newAddress(byte(addressType))
}

func newAddress(addressType byte) (address Address, err error) {
	switch addressType {
	case AddressEd25519:
		return &Ed25519Address{}, nil
	case AddressBLS:
		return &BLSAddress{}, nil
	case AddressAlias:
		return &AliasAddress{}, nil
	case AddressNFT:
		return &NFTAddress{}, nil
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownAddrType, addressType)
	}
}

// checks whether the given serializable is an address and also an existing type of address.
func isValidAddrType(seri serializer.Serializable) error {
	if seri == nil {
		return ErrTypeIsNotAddress
	}
	addr, isAddress := seri.(Address)
	if !isAddress {
		return ErrTypeIsNotAddress
	}
	if _, err := AddressSelector(uint32(addr.Type())); err != nil {
		return err
	}
	return nil
}

func addressFromJSONRawMsg(jRawMsg *json.RawMessage) (serializer.Serializable, error) {
	jsonAddr, err := DeserializeObjectFromJSON(jRawMsg, jsonAddressSelector)
	if err != nil {
		return nil, fmt.Errorf("can't decode address type from JSON: %w", err)
	}

	addr, err := jsonAddr.ToSerializable()
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func addressToJSONRawMsg(addr serializer.Serializable) (*json.RawMessage, error) {
	if err := isValidAddrType(addr); err != nil {
		return nil, err
	}
	addrJsonBytes, err := addr.MarshalJSON()
	if err != nil {
		return nil, err
	}
	jsonRawMsgAddr := json.RawMessage(addrJsonBytes)
	return &jsonRawMsgAddr, nil
}

// selects the json object for the given type.
func jsonAddressSelector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch byte(ty) {
	case AddressEd25519:
		obj = &jsonEd25519Address{}
	case AddressBLS:
		obj = &jsonBLSAddress{}
	case AddressAlias:
		obj = &jsonAliasAddress{}
	case AddressNFT:
		obj = &jsonNFTAddress{}
	default:
		return nil, fmt.Errorf("unable to decode address type from JSON: %w", ErrUnknownAddrType)
	}
	return obj, nil
}

func bech32String(hrp NetworkPrefix, addr Address) string {
	bytes, _ := addr.Serialize(serializer.DeSeriModeNoValidation)
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
		return "", nil, serializer.ErrDeserializationNotEnoughData
	}

	addr, err := newAddress(addrData[0])
	if err != nil {
		return "", nil, err
	}

	n, err := addr.Deserialize(addrData, serializer.DeSeriModePerformValidation)
	if err != nil {
		return "", nil, err
	}

	if n != len(addrData) {
		return "", nil, serializer.ErrDeserializationNotAllConsumed
	}

	return NetworkPrefix(hrp), addr, nil
}
