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
	addressesWithWeightV3ArrRules = &serix.ArrayRules{
		Min:            1,
		Max:            10,
		ValidationMode: serializer.ArrayValidationModeLexicalOrdering,
	}

	// multiAddressValidatorFunc is a validator which checks that:
	//  1. MultiAddresses are not nested inside the MultiAddress.
	//  2. "raw address part" of all addresses are unique (without type byte and capabilities).
	//  3. The weight of each address is at least 1.
	//  4. The threshold is smaller or equal to the cumulative weight of all addresses.
	multiAddressValidatorFunc = func(ctx context.Context, addr MultiAddress) error {
		addrSet := map[string]int{}

		var cumulativeWeight uint16
		for idx, address := range addr.Addresses {
			var addrWithoutTypeAndCapabilities []byte
			switch addr := address.Address.(type) {
			case *AccountAddress:
				addrWithoutTypeAndCapabilities = addr[:]
			case *Ed25519Address:
				addrWithoutTypeAndCapabilities = addr[:]
			case *ImplicitAccountCreationAddress:
				addrWithoutTypeAndCapabilities = addr[:]
			case *MultiAddress:
				return ierrors.Wrapf(ErrNestedMultiAddress, "address with index %d is a multi address inside a multi address", idx)
			case *NFTAddress:
				addrWithoutTypeAndCapabilities = addr[:]
			case *RestrictedEd25519Address:
				addrWithoutTypeAndCapabilities = addr.PubKeyHash[:]
			default:
				return ierrors.Wrapf(ErrUnknownAddrType, "address with index %d has an unknown address type (%T) inside a multi address", idx, addr)
			}

			// we need to check for uniqueness of the address, but instead of the whole serialized address, we need to ignore
			// different address types or capabilities, that might result in the same signature.
			addrString := string(addrWithoutTypeAndCapabilities)
			if j, has := addrSet[addrString]; has {
				return ierrors.Wrapf(serializer.ErrArrayValidationViolatesUniqueness, "element %d and %d are duplicates", j, idx)
			}
			addrSet[addrString] = idx

			// check for minimum address weight
			if address.Weight < 1 {
				return ierrors.Wrapf(ErrMultiAddressThresholdInvalid, "address with index %d needs to have at least weight=1", idx)
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
)

func CommonSerixAPI() *serix.API {
	api := serix.NewAPI()

	{
		must(api.RegisterTypeSettings(Ed25519Address{},
			serix.TypeSettings{}.WithObjectType(uint8(AddressEd25519)).WithMapKey("pubKeyHash")),
		)
		must(api.RegisterTypeSettings(RestrictedEd25519Address{},
			serix.TypeSettings{}.WithObjectType(uint8(AddressRestrictedEd25519))),
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
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(addressesWithWeightV3ArrRules),
		))

		must(api.RegisterInterfaceObjects((*Address)(nil), (*Ed25519Address)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*RestrictedEd25519Address)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*AccountAddress)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*NFTAddress)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*ImplicitAccountCreationAddress)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*MultiAddress)(nil)))

		// All versions of the protocol need to be able to parse older protocol parameter versions.
		{
			must(api.RegisterTypeSettings(RentStructure{}, serix.TypeSettings{}))

			must(api.RegisterTypeSettings(V3ProtocolParameters{},
				serix.TypeSettings{}.WithObjectType(uint8(ProtocolParametersV3))),
			)
			must(api.RegisterInterfaceObjects((*ProtocolParameters)(nil), (*V3ProtocolParameters)(nil)))
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
