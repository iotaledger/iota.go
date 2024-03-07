package iotago

import (
	"context"
	"fmt"
	"io"

	"github.com/iotaledger/hive.go/constraints"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
	"github.com/iotaledger/hive.go/serializer/v2/stream"
	"github.com/iotaledger/iota.go/v4/bech32"
)

var (
	// ErrInvalidAddressType gets returned when an address type is invalid.
	ErrInvalidAddressType = ierrors.New("invalid address type")
	// ErrInvalidRestrictedAddress gets returned when a RestrictedAddress is invalid.
	ErrInvalidRestrictedAddress = ierrors.New("invalid restricted address")
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
	// ErrAddressCannotReceiveAnchorOutput gets returned if an AnchorOutput is sent to an address without that capability.
	ErrAddressCannotReceiveAnchorOutput = ierrors.New("address cannot receive anchor outputs")
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
	// AddressAnchor denotes an Anchor address.
	AddressAnchor AddressType = 24
	// AddressImplicitAccountCreation denotes an Ed25519 address that can only be used to create an implicit account.
	AddressImplicitAccountCreation AddressType = 32
	// AddressMulti denotes a multi address.
	AddressMulti AddressType = 40
	// AddressRestricted denotes a restricted address that has a capability bitmask.
	AddressRestricted AddressType = 48
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
		"AnchorAddress", "", "", "", "", "", "", "",
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
	PrefixShimmer NetworkPrefix = "smr"
	PrefixTestnet NetworkPrefix = "rms"
)

// Address describes a general address.
type Address interface {
	Sizer
	NonEphemeralObject
	fmt.Stringer
	constraints.Cloneable[Address]
	constraints.Equalable[Address]

	// Type returns the type of the address.
	Type() AddressType

	// ID returns the address ID, which is the concatenation of type prefix
	// and the unique identifier of the address.
	ID() []byte

	// Bech32 encodes the address as a bech32 string.
	Bech32(hrp NetworkPrefix) string

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
	CannotReceiveAnchorOutputs() bool
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
	ChainID() ChainID
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
	case AddressAnchor:
		return &AnchorAddress{}, nil
	case AddressImplicitAccountCreation:
		return &ImplicitAccountCreationAddress{}, nil
	case AddressMulti:
		return &MultiAddress{}, nil
	case AddressRestricted:
		return &RestrictedAddress{}, nil
	default:
		panic(fmt.Sprintf("unknown address type %d", addressType))
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
		multiAddrRef, _, err := MultiAddressReferenceFromBytes(addrData)
		if err != nil {
			return "", nil, ierrors.Errorf("invalid multi address: %w", err)
		}

		return NetworkPrefix(hrp), multiAddrRef, nil

	case AddressRestricted:
		if len(addrData) == 1 {
			return "", nil, serializer.ErrDeserializationNotEnoughData
		}
		underlyingAddrType := AddressType(addrData[1])
		if underlyingAddrType == AddressMulti {
			multiAddrRef, consumed, err := MultiAddressReferenceFromBytes(addrData[1:])
			if err != nil {
				return "", nil, ierrors.Errorf("invalid multi address: %w", err)
			}

			// get the address capabilities from the remaining bytes
			capabilities, _, err := AddressCapabilitiesBitMaskFromBytes(addrData[1+consumed:])
			if err != nil {
				return "", nil, ierrors.Errorf("invalid address capabilities: %w", err)
			}

			return NetworkPrefix(hrp), &RestrictedAddress{
				Address:             multiAddrRef,
				AllowedCapabilities: capabilities,
			}, nil
		}
	}

	addr, err := newAddress(addrType)
	if err != nil {
		return "", nil, err
	}

	serixAPI := CommonSerixAPI()
	n, err := serixAPI.Decode(context.TODO(), addrData, addr)
	if err != nil {
		return "", nil, err
	}

	if n != len(addrData) {
		return "", nil, serializer.ErrDeserializationNotAllConsumed
	}

	return NetworkPrefix(hrp), addr, nil
}

func AddressFromBytes(bytes []byte) (Address, int, error) {
	var addr Address

	n, err := CommonSerixAPI().Decode(context.TODO(), bytes, &addr, serix.WithValidation())
	if err != nil {
		return nil, 0, ierrors.Wrap(err, "unable to decode address")
	}

	return addr, n, nil
}

func AddressFromReader(reader io.ReadSeeker) (Address, error) {
	addressType, err := stream.PeekSize(reader, serializer.SeriLengthPrefixTypeAsByte)
	if err != nil {
		return nil, ierrors.Wrap(err, "unable to read address type")
	}

	switch AddressType(addressType) {
	case AddressEd25519:
		return Ed25519AddressFromReader(reader)

	case AddressAccount:
		return AccountAddressFromReader(reader)

	case AddressNFT:
		return NFTAddressFromReader(reader)

	case AddressAnchor:
		return AnchorAddressFromReader(reader)

	case AddressImplicitAccountCreation:
		return ImplicitAccountCreationAddressFromReader(reader)

	case AddressMulti:
		return MultiAddressFromReader(reader)

	case AddressRestricted:
		return RestrictedAddressFromReader(reader)

	default:
		panic(fmt.Sprintf("unknown address type %d", addressType))
	}
}
