package iotago

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/bech32"
)

// AddressType defines the type of addresses.
type AddressType = byte

const (
	// AddressEd25519 denotes an Ed25519 address.
	AddressEd25519 = 0
	// AddressAlias denotes an Alias address.
	AddressAlias = 8
	// AddressNFT denotes an NFT address.
	AddressNFT = 16
)

// AddressTypeSet is a set of AddressType.
type AddressTypeSet map[AddressType]struct{}

var (
	// ErrTypeIsNotSupportedAddress gets returned when a serializable was found to not be a supported Address.
	ErrTypeIsNotSupportedAddress = errors.New("serializable is not a supported address")

	allAddressTypeSet = AddressTypeSet{
		AddressEd25519: struct{}{},
		AddressAlias:   struct{}{},
		AddressNFT:     struct{}{},
	}
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
	NonEphemeralObject
	fmt.Stringer

	// Type returns the type of the address.
	Type() AddressType

	// Bech32 encodes the address as a bech32 string.
	Bech32(hrp NetworkPrefix) string

	// Equal checks whether other is equal to this Address.
	Equal(other Address) bool

	// Key returns a string which can be used to index the Address in a map.
	Key() string

	// Clone clones the Address.
	Clone() Address
}

// DirectUnlockableAddress is a type of Address which can be directly unlocked.
type DirectUnlockableAddress interface {
	Address
	// Unlock unlocks this DirectUnlockableAddress given the Signature.
	Unlock(msg []byte, sig Signature) error
}

// ChainConstrainedAddress is a type of Address representing ownership of an output by a ChainConstrainedOutput.
type ChainConstrainedAddress interface {
	Address
	Chain() ChainID
}

// ChainID represents the chain ID of a chain created by a ChainConstrainedOutput.
type ChainID interface {
	// Matches checks whether other matches this ChainID.
	Matches(other ChainID) bool
	// Addressable tells whether this ChainID can be converted into a ChainConstrainedAddress.
	Addressable() bool
	// ToAddress converts this ChainID into an ChainConstrainedAddress.
	ToAddress() ChainConstrainedAddress
	// Empty tells whether the ChainID is empty.
	Empty() bool
	// Key returns a key to use to index this ChainID.
	Key() interface{}
}

// UTXOIDChainID is a ChainID which gets produced by taking an OutputID.
type UTXOIDChainID interface {
	FromOutputID(id OutputID) ChainID
}

// AddressSelector implements SerializableSelectorFunc for address types.
func AddressSelector(addressType uint32) (Address, error) {
	return newAddress(byte(addressType))
}

// AddressTypeToString returns the name for the given AddressType.
func AddressTypeToString(ty AddressType) string {
	switch ty {
	case AddressEd25519:
		return "Ed25519Address"
	case AddressAlias:
		return "AliasAddress"
	case AddressNFT:
		return "NFTAddress"
	default:
		return "unknown address typ"
	}
}

func newAddress(addressType byte) (address Address, err error) {
	switch addressType {
	case AddressEd25519:
		return &Ed25519Address{}, nil
	case AddressAlias:
		return &AliasAddress{}, nil
	case AddressNFT:
		return &NFTAddress{}, nil
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownAddrType, addressType)
	}
}

func jsonAddressToAddress(jAddr JSONSerializable) (Address, error) {
	addr, err := jAddr.ToSerializable()
	if err != nil {
		return nil, err
	}
	return addr.(Address), nil
}

// checks whether the given Serializable is an Address and also supported AddressType.
func addrWriteGuard(supportedAddr AddressTypeSet) serializer.SerializableWriteGuardFunc {
	return func(seri serializer.Serializable) error {
		if seri == nil {
			return fmt.Errorf("%w: because nil", ErrTypeIsNotSupportedAddress)
		}
		addr, is := seri.(Address)
		if !is {
			return fmt.Errorf("%w: because not address", ErrTypeIsNotSupportedAddress)
		}

		if _, supported := supportedAddr[addr.Type()]; !supported {
			return fmt.Errorf("%w: because not in set %v", ErrTypeIsNotSupportedAddress, supported)
		}

		return nil
	}
}

func addrReadGuard(supportedAddr AddressTypeSet) serializer.SerializableReadGuardFunc {
	return func(ty uint32) (serializer.Serializable, error) {
		if _, supported := supportedAddr[byte(ty)]; !supported {
			return nil, fmt.Errorf("%w: because not in set %v (%d)", ErrTypeIsNotSupportedAddress, supportedAddr, ty)
		}
		return AddressSelector(ty)
	}
}

func addressFromJSONRawMsg(jRawMsg *json.RawMessage) (Address, error) {
	jsonAddr, err := DeserializeObjectFromJSON(jRawMsg, jsonAddressSelector)
	if err != nil {
		return nil, fmt.Errorf("can't decode address type from JSON: %w", err)
	}

	addr, err := jsonAddr.ToSerializable()
	if err != nil {
		return nil, err
	}
	return addr.(Address), nil
}

func addressToJSONRawMsg(addr serializer.Serializable) (*json.RawMessage, error) {
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
	bytes, _ := addr.Serialize(serializer.DeSeriModeNoValidation, nil)
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

	n, err := addr.Deserialize(addrData, serializer.DeSeriModePerformValidation, nil)
	if err != nil {
		return "", nil, err
	}

	if n != len(addrData) {
		return "", nil, serializer.ErrDeserializationNotAllConsumed
	}

	return NetworkPrefix(hrp), addr, nil
}
