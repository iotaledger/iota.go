package iotago

import (
	"context"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
)

var (
	// the addresses need to be unique and lexically ordered to calculate a deterministic bech32 address for a MultiAddress.
	addressesWithWeightV3ArrRules = &serix.ArrayRules{
		Min: 1,
		Max: 10,
		UniquenessSliceFunc: func(next []byte) []byte {
			// we need to ignore the Weight of the AddressWithWeight to compare for address uniqueness
			return next[:len(next)-AddressWeightSerializedBytesSize]
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates | serializer.ArrayValidationModeLexicalOrdering,
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
		must(api.RegisterTypeSettings(RestrictedAccountAddress{},
			serix.TypeSettings{}.WithObjectType(uint8(AddressRestrictedAccount))),
		)
		must(api.RegisterTypeSettings(NFTAddress{},
			serix.TypeSettings{}.WithObjectType(uint8(AddressNFT)).WithMapKey("nftId")),
		)
		must(api.RegisterTypeSettings(RestrictedNFTAddress{},
			serix.TypeSettings{}.WithObjectType(uint8(AddressRestrictedNFT))),
		)
		must(api.RegisterTypeSettings(ImplicitAccountCreationAddress{},
			serix.TypeSettings{}.WithObjectType(uint8(AddressImplicitAccountCreation)).WithMapKey("pubKeyHash")),
		)
		must(api.RegisterTypeSettings(MultiAddress{},
			serix.TypeSettings{}.WithObjectType(uint8(AddressMulti))),
		)
		must(api.RegisterTypeSettings(RestrictedMultiAddress{},
			serix.TypeSettings{}.WithObjectType(uint8(AddressRestrictedMulti))),
		)
		must(api.RegisterTypeSettings(AddressesWithWeight{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(addressesWithWeightV3ArrRules),
		))

		must(api.RegisterInterfaceObjects((*Address)(nil), (*Ed25519Address)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*RestrictedEd25519Address)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*AccountAddress)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*RestrictedAccountAddress)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*NFTAddress)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*RestrictedNFTAddress)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*ImplicitAccountCreationAddress)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*MultiAddress)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*RestrictedMultiAddress)(nil)))

		must(api.RegisterValidators(MultiAddress{}, nil, func(ctx context.Context, addr MultiAddress) error {
			var cumulativeWeight uint16
			for i, address := range addr.Addresses {
				if address.Weight < 1 {
					return fmt.Errorf("%w: address with index %d needs to have at least weight=1", ErrMultiAddressThresholdInvalid, i)
				}
				cumulativeWeight += uint16(address.Weight)
			}

			if addr.Threshold > cumulativeWeight {
				return fmt.Errorf("%w: the threshold value exceeds the cumulative weight of all addresses (%d>%d)", ErrMultiAddressThresholdInvalid, addr.Threshold, cumulativeWeight)
			}

			return nil
		}))
		must(api.RegisterValidators(RestrictedMultiAddress{}, nil, func(ctx context.Context, addr RestrictedMultiAddress) error {
			var cumulativeWeight uint16
			for i, address := range addr.Addresses {
				if address.Weight < 1 {
					return fmt.Errorf("%w: address with index %d needs to have at least weight=1", ErrMultiAddressThresholdInvalid, i)
				}
				cumulativeWeight += uint16(address.Weight)
			}

			if addr.Threshold > cumulativeWeight {
				return fmt.Errorf("%w: the threshold value exceeds the cumulative weight of all addresses (%d>%d)", ErrMultiAddressThresholdInvalid, addr.Threshold, cumulativeWeight)
			}

			return nil
		}))

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
