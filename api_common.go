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
		ValidationMode: serializer.ArrayValidationModeLexicalOrdering | serializer.ArrayValidationModeNoDuplicates,
	}

	// multiAddressValidatorFunc is a validator which checks that:
	//  1. ImplicitAccountCreationAddress, MultiAddresses, RestrictedAddress are not nested inside the MultiAddress.
	//  2. The weight of each address is at least 1.
	//  3. The threshold is smaller or equal to the cumulative weight of all addresses.
	multiAddressValidatorFunc = func(ctx context.Context, addr MultiAddress) error {
		var cumulativeWeight uint16
		for idx, address := range addr.Addresses {
			switch addr := address.Address.(type) {
			case *Ed25519Address:
			case *AccountAddress:
			case *NFTAddress:
			case *AnchorAddress:
			case *ImplicitAccountCreationAddress:
				return ierrors.Wrapf(ErrInvalidNestedAddressType, "address with index %d is an implicit account creation address inside a multi address", idx)
			case *MultiAddress:
				return ierrors.Wrapf(ErrInvalidNestedAddressType, "address with index %d is a multi address inside a multi address", idx)
			case *RestrictedAddress:
				return ierrors.Wrapf(ErrInvalidNestedAddressType, "address with index %d is a restricted address inside a multi address", idx)
			default:
				return ierrors.Wrapf(ErrUnknownAddrType, "address with index %d has an unknown address type (%T) inside a multi address", idx, addr)
			}

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
		case *Ed25519Address, *AccountAddress, *NFTAddress, *AnchorAddress, *MultiAddress:
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

	blockIssuerKeysArrRules = &serix.ArrayRules{
		Min: MinBlockIssuerKeysCount,
		Max: MaxBlockIssuerKeysCount,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering,
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
		must(api.RegisterTypeSettings(AnchorAddress{},
			serix.TypeSettings{}.WithObjectType(uint8(AddressAnchor)).WithMapKey("anchorId")),
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
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithMaxLen(2),
		))

		must(api.RegisterInterfaceObjects((*Address)(nil), (*Ed25519Address)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*AccountAddress)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*NFTAddress)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*AnchorAddress)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*ImplicitAccountCreationAddress)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*MultiAddress)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*RestrictedAddress)(nil)))

		// All versions of the protocol need to be able to parse older protocol parameter versions.
		{
			must(api.RegisterTypeSettings(StorageScoreStructure{}, serix.TypeSettings{}))

			must(api.RegisterTypeSettings(V3ProtocolParameters{},
				serix.TypeSettings{}.WithObjectType(uint8(ProtocolParametersV3))),
			)
			must(api.RegisterInterfaceObjects((*ProtocolParameters)(nil), (*V3ProtocolParameters)(nil)))
		}

		{
			must(api.RegisterTypeSettings(Ed25519PublicKeyBlockIssuerKey{},
				serix.TypeSettings{}.WithObjectType(byte(BlockIssuerKeyEd25519PublicKey)),
			))
			must(api.RegisterTypeSettings(Ed25519PublicKeyHashBlockIssuerKey{},
				serix.TypeSettings{}.WithObjectType(byte(BlockIssuerKeyPublicKeyHash)),
			))
			must(api.RegisterInterfaceObjects((*BlockIssuerKey)(nil), (*Ed25519PublicKeyBlockIssuerKey)(nil)))
			must(api.RegisterInterfaceObjects((*BlockIssuerKey)(nil), (*Ed25519PublicKeyHashBlockIssuerKey)(nil)))

			must(api.RegisterTypeSettings(BlockIssuerKeys{},
				serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(blockIssuerKeysArrRules),
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
