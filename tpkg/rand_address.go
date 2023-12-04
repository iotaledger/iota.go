package tpkg

import (
	"bytes"
	"slices"

	"github.com/iotaledger/hive.go/ierrors"
	iotago "github.com/iotaledger/iota.go/v4"
)

// RandEd25519Address returns a random Ed25519 address.
func RandEd25519Address() *iotago.Ed25519Address {
	edAddr := &iotago.Ed25519Address{}
	addr := RandBytes(iotago.Ed25519AddressBytesLength)
	copy(edAddr[:], addr)

	return edAddr
}

// RandAccountAddress returns a random AccountAddress.
func RandAccountAddress() *iotago.AccountAddress {
	addr := &iotago.AccountAddress{}
	accountID := RandBytes(iotago.AccountAddressBytesLength)
	copy(addr[:], accountID)

	return addr
}

// RandNFTAddress returns a random NFTAddress.
func RandNFTAddress() *iotago.NFTAddress {
	addr := &iotago.NFTAddress{}
	nftID := RandBytes(iotago.NFTAddressBytesLength)
	copy(addr[:], nftID)

	return addr
}

// RandAnchorAddress returns a random AnchorAddress.
func RandAnchorAddress() *iotago.AnchorAddress {
	addr := &iotago.AnchorAddress{}
	anchorID := RandBytes(iotago.AnchorAddressBytesLength)
	copy(addr[:], anchorID)

	return addr
}

// RandImplicitAccountCreationAddress returns a random ImplicitAccountCreationAddress.
func RandImplicitAccountCreationAddress() *iotago.ImplicitAccountCreationAddress {
	iacAddr := &iotago.ImplicitAccountCreationAddress{}
	addr := RandBytes(iotago.Ed25519AddressBytesLength)
	copy(iacAddr[:], addr)

	return iacAddr
}

// RandMultiAddress returns a random MultiAddress.
func RandMultiAddress() *iotago.MultiAddress {
	addrCnt := RandInt(10) + 1

	cumulativeWeight := 0
	addresses := make([]*iotago.AddressWithWeight, 0, addrCnt)
	for i := 0; i < addrCnt; i++ {
		weight := RandInt(8) + 1
		cumulativeWeight += weight
		addresses = append(addresses, &iotago.AddressWithWeight{
			Address: RandAddress(),
			Weight:  byte(weight),
		})
	}

	slices.SortFunc(addresses, func(a *iotago.AddressWithWeight, b *iotago.AddressWithWeight) int {
		return bytes.Compare(a.Address.ID(), b.Address.ID())
	})

	threshold := RandInt(cumulativeWeight) + 1

	return &iotago.MultiAddress{
		Addresses: addresses,
		Threshold: uint16(threshold),
	}
}

// RandRestrictedEd25519Address returns a random restricted Ed25519 address.
func RandRestrictedEd25519Address(capabilities iotago.AddressCapabilitiesBitMask) *iotago.RestrictedAddress {
	return &iotago.RestrictedAddress{
		Address:             RandEd25519Address(),
		AllowedCapabilities: capabilities,
	}
}

// RandRestrictedAccountAddress returns a random restricted account address.
func RandRestrictedAccountAddress(capabilities iotago.AddressCapabilitiesBitMask) *iotago.RestrictedAddress {
	return &iotago.RestrictedAddress{
		Address:             RandAccountAddress(),
		AllowedCapabilities: capabilities,
	}
}

// RandRestrictedNFTAddress returns a random restricted NFT address.
func RandRestrictedNFTAddress(capabilities iotago.AddressCapabilitiesBitMask) *iotago.RestrictedAddress {
	return &iotago.RestrictedAddress{
		Address:             RandNFTAddress(),
		AllowedCapabilities: capabilities,
	}
}

// RandRestrictedAnchorAddress returns a random restricted anchor address.
func RandRestrictedAnchorAddress(capabilities iotago.AddressCapabilitiesBitMask) *iotago.RestrictedAddress {
	return &iotago.RestrictedAddress{
		Address:             RandAnchorAddress(),
		AllowedCapabilities: capabilities,
	}
}

// RandRestrictedMultiAddress returns a random restricted multi address.
func RandRestrictedMultiAddress(capabilities iotago.AddressCapabilitiesBitMask) *iotago.RestrictedAddress {
	return &iotago.RestrictedAddress{
		Address:             RandMultiAddress(),
		AllowedCapabilities: capabilities,
	}
}

// RandAddress returns a random address (Ed25519, Account, NFT, Anchor).
func RandAddress(addressType ...iotago.AddressType) iotago.Address {
	var addrType iotago.AddressType
	if len(addressType) > 0 {
		addrType = addressType[0]
	} else {
		addressTypes := []iotago.AddressType{iotago.AddressEd25519, iotago.AddressAccount, iotago.AddressNFT, iotago.AddressAnchor}
		addrType = addressTypes[RandInt(len(addressTypes))]
	}

	//nolint:exhaustive
	switch addrType {
	case iotago.AddressEd25519:
		return RandEd25519Address()
	case iotago.AddressAccount:
		return RandAccountAddress()
	case iotago.AddressNFT:
		return RandNFTAddress()
	case iotago.AddressAnchor:
		return RandAnchorAddress()
	default:
		panic(ierrors.Wrapf(iotago.ErrUnknownAddrType, "type %d", addrType))
	}
}
