package iotago

import (
	"fmt"

	"github.com/iotaledger/hive.go/core/serix"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/bech32"
)

// AddressType defines the type of addresses.
type AddressType byte

const (
	// AddressEd25519 denotes an Ed25519 address.
	AddressEd25519 AddressType = 0
	// AddressAlias denotes an Alias address.
	AddressAlias AddressType = 8
	// AddressNFT denotes an NFT address.
	AddressNFT AddressType = 16
)

func (addrType AddressType) String() string {
	if int(addrType) >= len(addressNames) {
		return fmt.Sprintf("unknown address type: %d", addrType)
	}

	return addressNames[addrType]
}

// AddressTypeSet is a set of AddressType.
type AddressTypeSet map[AddressType]struct{}

var (
	addressNames = [AddressNFT + 1]string{
		"Ed25519Address", "", "", "", "", "", "", "",
		"AliasAddress", "", "", "", "", "", "", "",
		"NFTAddress",
	}
)

// NetworkPrefix denotes the different network prefixes.
type NetworkPrefix string

// Network prefixes.
const (
	PrefixMainnet NetworkPrefix = "iota"
	PrefixDevnet  NetworkPrefix = "atoi"
	PrefixShimmer NetworkPrefix = "smr"
	PrefixTestnet NetworkPrefix = "rms"
)

// Address describes a general address.
type Address interface {
	serix.Serializable
	serix.Deserializable
	Sizer
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

// ChainAddress is a type of Address representing ownership of an output by a ChainOutput.
type ChainAddress interface {
	Address
	Chain() ChainID
}

// ChainID represents the chain ID of a chain created by a ChainOutput.
type ChainID interface {
	// Matches checks whether other matches this ChainID.
	Matches(other ChainID) bool
	// Addressable tells whether this ChainID can be converted into a ChainAddress.
	Addressable() bool
	// ToAddress converts this ChainID into an ChainAddress.
	ToAddress() ChainAddress
	// Empty tells whether the ChainID is empty.
	Empty() bool
	// Key returns a key to use to index this ChainID.
	Key() interface{}
	// ToHex returns the hex representation of the ChainID.
	ToHex() string
}

// UTXOIDChainID is a ChainID which gets produced by taking an OutputID.
type UTXOIDChainID interface {
	FromOutputID(id OutputID) ChainID
}

func newAddress(addressType byte) (address Address, err error) {
	switch AddressType(addressType) {
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

func bech32String(hrp NetworkPrefix, addr Address) string {
	bytes, err := _internalAPI.Encode(addr)
	if err != nil {
		panic(err)
	}
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

	n, err := _internalAPI.Decode(addrData, addr)
	if err != nil {
		return "", nil, err
	}

	if n != len(addrData) {
		return "", nil, serializer.ErrDeserializationNotAllConsumed
	}

	return NetworkPrefix(hrp), addr, nil
}
