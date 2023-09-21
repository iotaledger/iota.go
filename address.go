package iotago

import (
	"context"
	"fmt"

	"github.com/iotaledger/hive.go/constraints"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/bech32"
)

var (
	// ErrUnknownAddrType gets returned for unknown address types.
	ErrUnknownAddrType = ierrors.New("unknown address type")
	// ErrInvalidNestedAddressType gets returned when a nested address inside a MultiAddress or RestrictedAddress is invalid.
	ErrInvalidNestedAddressType = ierrors.New("invalid nested address type")
	// ErrImplicitAccountCreationAddressInInvalidUnlockCondition gets returned when a Implicit Account Creation Address
	// is placed in an unlock condition where it is disallowed.
	ErrImplicitAccountCreationAddressInInvalidUnlockCondition = ierrors.New("implicit account creation address in unlock condition where it is disallowed")
	// ErrImplicitAccountCreationAddressInInvalidOutput gets returned when a ImplicitAccountCreationAddress
	// is placed in an output where it is disallowed.
	ErrImplicitAccountCreationAddressInInvalidOutput = ierrors.New("implicit account creation address in output where it is disallowed")
	// ErrAddressCannotReceiveNativeTokens gets returned if Native Tokens are sent to an address without that capability.
	ErrAddressCannotReceiveNativeTokens = ierrors.New("address cannot receive native tokens")
	// ErrAddressCannotReceiveMana gets returned if Mana is sent to an address without that capability.
	ErrAddressCannotReceiveMana = ierrors.New("address cannot receive mana")
	// ErrAddressCannotReceiveTimelockUnlockCondition gets returned if an output with a
	// TimelockUnlockCondition is sent to an address without that capability.
	ErrAddressCannotReceiveTimelockUnlockCondition = ierrors.New("address cannot receive outputs with timelock unlock condition")
	// ErrAddressCannotReceiveExpirationUnlockCondition gets returned if an output with a
	// ExpirationUnlockCondition is sent to an address without that capability.
	ErrAddressCannotReceiveExpirationUnlockCondition = ierrors.New("address cannot receive outputs with expiration unlock condition")
	// ErrAddressCannotReceiveStorageDepositReturnUnlockCondition gets returned if an output with a
	// StorageDepositReturnUnlockCondition is sent to an address without that capability.
	ErrAddressCannotReceiveStorageDepositReturnUnlockCondition = ierrors.New("address cannot receive outputs with storage deposit return unlock condition")
	// ErrAddressCannotReceiveAccountOutput gets returned if an AccountOutput is sent to an address without that capability.
	ErrAddressCannotReceiveAccountOutput = ierrors.New("address cannot receive account outputs")
	// ErrAddressCannotReceiveNFTOutput gets returned if an NFTOutput is sent to an address without that capability.
	ErrAddressCannotReceiveNFTOutput = ierrors.New("address cannot receive nft outputs")
	// ErrAddressCannotReceiveDelegationOutput gets returned if a DelegationOutput is sent to an address without that capability.
	ErrAddressCannotReceiveDelegationOutput = ierrors.New("address cannot receive delegation outputs")
)

// AddressType defines the type of addresses.
type AddressType byte

const (
	// AddressEd25519 denotes an Ed25519 address.
	AddressEd25519 AddressType = 0
	// AddressAccount denotes an Account address.
	AddressAccount AddressType = 8
	// AddressNFT denotes an NFT address.
	AddressNFT AddressType = 16
	// AddressImplicitAccountCreation denotes an Ed25519 address that can only be used to create an implicit account.
	AddressImplicitAccountCreation AddressType = 24
	// AddressMulti denotes a multi address.
	AddressMulti AddressType = 32
	// AddressRestricted denotes a restricted address that has a capability bitmask.
	AddressRestricted AddressType = 40
)

func (addrType AddressType) String() string {
	if int(addrType) >= len(addressNames) {
		return fmt.Sprintf("unknown address type: %d", addrType)
	}

	addressName := addressNames[addrType]
	if addressName == "" {
		return fmt.Sprintf("unknown address type: %d", addrType)
	}

	return addressName
}

// AddressTypeSet is a set of AddressType.
type AddressTypeSet map[AddressType]struct{}

var (
	addressNames = [AddressRestricted + 1]string{
		"Ed25519Address", "", "", "", "", "", "", "",
		"AccountAddress", "", "", "", "", "", "", "",
		"NFTAddress", "", "", "", "", "", "", "",
		"ImplicitAccountCreationAddress", "", "", "", "", "", "", "",
		"MultiAddress", "", "", "", "", "", "", "",
		"RestrictedAddress",
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
	Sizer
	NonEphemeralObject
	fmt.Stringer
	constraints.Cloneable[Address]

	// Type returns the type of the address.
	Type() AddressType

	// ID returns the address ID, which is the concatenation of type prefix
	// and the unique identifier of the address.
	ID() []byte

	// Bech32 encodes the address as a bech32 string.
	Bech32(hrp NetworkPrefix) string

	// Equal checks whether other is equal to this Address.
	Equal(other Address) bool

	// Key returns a string which can be used to index the Address in a map.
	Key() string
}

type AddressCapabilities interface {
	CannotReceiveNativeTokens() bool
	CannotReceiveMana() bool
	CannotReceiveOutputsWithTimelockUnlockCondition() bool
	CannotReceiveOutputsWithExpirationUnlockCondition() bool
	CannotReceiveOutputsWithStorageDepositReturnUnlockCondition() bool
	CannotReceiveAccountOutputs() bool
	CannotReceiveNFTOutputs() bool
	CannotReceiveDelegationOutputs() bool
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

func newAddress(addressType AddressType) (address Address, err error) {
	switch addressType {
	case AddressEd25519:
		return &Ed25519Address{}, nil
	case AddressAccount:
		return &AccountAddress{}, nil
	case AddressNFT:
		return &NFTAddress{}, nil
	case AddressImplicitAccountCreation:
		return &ImplicitAccountCreationAddress{}, nil
	case AddressMulti:
		return &MultiAddress{}, nil
	case AddressRestricted:
		return &RestrictedAddress{}, nil
	default:
		return nil, ierrors.Wrapf(ErrUnknownAddrType, "type %d", addressType)
	}
}

func bech32StringBytes(hrp NetworkPrefix, bytes []byte) string {
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
		return "", nil, ierrors.Errorf("invalid bech32 encoding: %w", err)
	}

	if len(addrData) == 0 {
		return "", nil, serializer.ErrDeserializationNotEnoughData
	}

	addrType := AddressType(addrData[0])

	// check for invalid MultiAddresses in bech32 string
	// MultiAddresses are hashed and can't be reconstructed via bech32
	//nolint:exhaustive
	switch addrType {
	case AddressMulti:
		// return the HRP so we can at least check for correct network
		return NetworkPrefix(hrp), nil, ErrMultiAddrCannotBeReconstructedViaBech32
	case AddressRestricted:
		if len(addrData) == 1 {
			return "", nil, serializer.ErrDeserializationNotEnoughData
		}
		underlyingAddrType := AddressType(addrData[1])
		if underlyingAddrType == AddressMulti {
			// return the HRP so we can at least check for correct network
			return NetworkPrefix(hrp), nil, ErrMultiAddrCannotBeReconstructedViaBech32
		}
	}

	addr, err := newAddress(addrType)
	if err != nil {
		return "", nil, err
	}

	serixAPI := CommonSerixAPI()
	n, err := serixAPI.Decode(context.Background(), addrData, addr)
	if err != nil {
		return "", nil, err
	}

	if n != len(addrData) {
		return "", nil, serializer.ErrDeserializationNotAllConsumed
	}

	return NetworkPrefix(hrp), addr, nil
}
