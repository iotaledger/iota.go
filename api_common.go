package iotago

import (
	"context"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
)

var (
	// the addresses need to be unique and lexically ordered to calculate a deterministic bech32 address for a MultiAddress.
	// HINT: the uniqueness is checked within a custom validator function, which is on MultiAddress level.
	addressesWithWeightArrRules = &serix.ArrayRules{
		Min:            1,
		Max:            10,
		ValidationMode: serializer.ArrayValidationModeLexicalOrdering,
	}

	// multiAddressValidatorFunc is a validator which checks that:
	//  1. ImplicitAccountCreationAddress, MultiAddresses, RestrictedAddress are not nested inside the MultiAddress.
	//  2. "raw address part" of all addresses are unique (without type byte and capabilities).
	//  3. The weight of each address is at least 1.
	//  4. The threshold is smaller or equal to the cumulative weight of all addresses.
	multiAddressValidatorFunc = func(ctx context.Context, addr MultiAddress) error {
		addrSet := map[string]int{}

		var cumulativeWeight uint16
		for idx, address := range addr.Addresses {
			var addrWithoutTypeAndCapabilities []byte

			switch addr := address.Address.(type) {
			case *Ed25519Address:
				addrWithoutTypeAndCapabilities = addr[:]
			case *AccountAddress:
				addrWithoutTypeAndCapabilities = addr[:]
			case *NFTAddress:
				addrWithoutTypeAndCapabilities = addr[:]
			case *ImplicitAccountCreationAddress:
				return ierrors.Wrapf(ErrInvalidNestedAddressType, "address with index %d is an implicit account creation address inside a multi address", idx)
			case *MultiAddress:
				return ierrors.Wrapf(ErrInvalidNestedAddressType, "address with index %d is a multi address inside a multi address", idx)
			case *RestrictedAddress:
				return ierrors.Wrapf(ErrInvalidNestedAddressType, "address with index %d is a restricted address inside a multi address", idx)
			default:
				return ierrors.Wrapf(ErrUnknownAddrType, "address with index %d has an unknown address type (%T) inside a multi address", idx, addr)
			}

			// we need to check for uniqueness of the address, but instead of the whole serialized address, we need to ignore
			// different address types or capabilities, that might result in the same signature.
			addrString := string(addrWithoutTypeAndCapabilities)
			if j, has := addrSet[addrString]; has {
				return ierrors.Wrapf(serializer.ErrArrayValidationViolatesUniqueness, "addresses with indices %d and %d in multi address are duplicates", j, idx)
			}
			addrSet[addrString] = idx

			// check for minimum address weight
			if address.Weight == 0 {
				return ierrors.Wrapf(ErrMultiAddressWeightInvalid, "address with index %d needs to have at least weight=1", idx)
			}

			cumulativeWeight += uint16(address.Weight)
		}

		// check for valid threshold
		if addr.Threshold > cumulativeWeight {
			return ierrors.Wrapf(ErrMultiAddressThresholdInvalid, "the threshold value exceeds the cumulative weight of all addresses (%d>%d)", addr.Threshold, cumulativeWeight)
		}
		if addr.Threshold < 1 {
			return ierrors.Wrap(ErrMultiAddressThresholdInvalid, "multi addresses need to have at least threshold=1")
		}

		return nil
	}

	// restrictedAddressValidatorFunc is a validator which checks that:
	//  1. ImplicitAccountCreationAddress are not nested inside the RestrictedAddress.
	//  2. RestrictedAddresses are not nested inside the RestrictedAddress.
	restrictedAddressValidatorFunc = func(ctx context.Context, addr RestrictedAddress) error {
		switch addr.Address.(type) {
		case *Ed25519Address, *AccountAddress, *NFTAddress, *MultiAddress:
			// allowed address types
		case *ImplicitAccountCreationAddress:
			return ierrors.Wrap(ErrInvalidNestedAddressType, "underlying address is an implicit account creation address inside a restricted address")
		case *RestrictedAddress:
			return ierrors.Wrap(ErrInvalidNestedAddressType, "underlying address is a restricted address inside a restricted address")
		default:
			return ierrors.Wrapf(ErrUnknownAddrType, "underlying address has an unknown address type (%T) inside a restricted address", addr)
		}

		return nil
	}
)

func CommonSerixAPI() *serix.API {
	api := serix.NewAPI()

	{
		must(api.RegisterTypeSettings(Ed25519Address{},
			serix.TypeSettings{}.WithObjectType(uint8(AddressEd25519)).WithMapKey("pubKeyHash")),
		)
		must(api.RegisterTypeSettings(AccountAddress{},
			serix.TypeSettings{}.WithObjectType(uint8(AddressAccount)).WithMapKey("accountId")),
		)
		must(api.RegisterTypeSettings(NFTAddress{},
			serix.TypeSettings{}.WithObjectType(uint8(AddressNFT)).WithMapKey("nftId")),
		)
		must(api.RegisterTypeSettings(ImplicitAccountCreationAddress{},
			serix.TypeSettings{}.WithObjectType(uint8(AddressImplicitAccountCreation)).WithMapKey("pubKeyHash")),
		)
		must(api.RegisterTypeSettings(MultiAddress{},
			serix.TypeSettings{}.WithObjectType(uint8(AddressMulti))),
		)
		must(api.RegisterValidators(MultiAddress{}, nil, multiAddressValidatorFunc))
		must(api.RegisterTypeSettings(AddressesWithWeight{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(addressesWithWeightArrRules),
		))
		must(api.RegisterTypeSettings(RestrictedAddress{},
			serix.TypeSettings{}.WithObjectType(uint8(AddressRestricted))),
		)
		must(api.RegisterValidators(RestrictedAddress{}, nil, restrictedAddressValidatorFunc))
		must(api.RegisterTypeSettings(AddressCapabilitiesBitMask{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithMaxLen(1),
		))

		must(api.RegisterInterfaceObjects((*Address)(nil), (*Ed25519Address)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*AccountAddress)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*NFTAddress)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*ImplicitAccountCreationAddress)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*MultiAddress)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*RestrictedAddress)(nil)))

		// All versions of the protocol need to be able to parse older protocol parameter versions.
		{
			must(api.RegisterTypeSettings(RentStructure{}, serix.TypeSettings{}))

			must(api.RegisterTypeSettings(V3ProtocolParameters{},
				serix.TypeSettings{}.WithObjectType(uint8(ProtocolParametersV3))),
			)
			must(api.RegisterInterfaceObjects((*ProtocolParameters)(nil), (*V3ProtocolParameters)(nil)))
		}

		{
			must(api.RegisterTypeSettings(Ed25519PublicKeyBlockIssuerKey{},
				serix.TypeSettings{}.WithObjectType(byte(BlockIssuerKeyEd25519PublicKey)),
			))
			must(api.RegisterTypeSettings(Ed25519AddressBlockIssuerKey{},
				serix.TypeSettings{}.WithObjectType(byte(BlockIssuerKeyEd25519Address)),
			))
			must(api.RegisterInterfaceObjects((*BlockIssuerKey)(nil), (*Ed25519PublicKeyBlockIssuerKey)(nil)))
			must(api.RegisterInterfaceObjects((*BlockIssuerKey)(nil), (*Ed25519AddressBlockIssuerKey)(nil)))

			must(api.RegisterTypeSettings(BlockIssuerKeys{},
				serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(accountOutputV3BlockIssuerKeysArrRules),
			))

		}
	}

	return api
}

func ProtocolParametersFromBytes(bytes []byte) (params ProtocolParameters, bytesRead int, err error) {
	var protocolParameters ProtocolParameters
	n, err := CommonSerixAPI().Decode(context.TODO(), bytes, &protocolParameters, serix.WithValidation())
	if err != nil {
		return nil, 0, err
	}

	return protocolParameters, n, nil
}
