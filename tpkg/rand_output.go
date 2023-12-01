package tpkg

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"

	iotago "github.com/iotaledger/iota.go/v4"
)

func RandOutputIDWithCreationSlot(slot iotago.SlotIndex, index ...uint16) iotago.OutputID {
	txID := RandTransactionIDWithCreationSlot(slot)

	idx := RandUint16(126)
	if len(index) > 0 {
		idx = index[0]
	}

	var outputID iotago.OutputID
	copy(outputID[:], txID[:])
	binary.LittleEndian.PutUint16(outputID[iotago.TransactionIDLength:], idx)

	return outputID
}

func RandOutputID(index ...uint16) iotago.OutputID {
	return RandOutputIDWithCreationSlot(0, index...)
}

func RandOutputIDsWithCreationSlot(slot iotago.SlotIndex, count uint16) iotago.OutputIDs {
	outputIDs := make(iotago.OutputIDs, int(count))
	for i := 0; i < int(count); i++ {
		outputIDs[i] = RandOutputIDWithCreationSlot(slot, count)
	}

	return outputIDs
}

func RandOutputIDs(count uint16) iotago.OutputIDs {
	outputIDs := make(iotago.OutputIDs, int(count))
	for i := 0; i < int(count); i++ {
		outputIDs[i] = RandOutputID(count)
	}

	return outputIDs
}

// RandBasicOutput returns a random basic output (with no features).
func RandBasicOutput(addressType ...iotago.AddressType) *iotago.BasicOutput {
	dep := &iotago.BasicOutput{
		Amount:           0,
		UnlockConditions: iotago.BasicOutputUnlockConditions{},
		Features:         iotago.BasicOutputFeatures{},
	}

	addrType := iotago.AddressEd25519
	if len(addressType) > 0 {
		addrType = addressType[0]
	}

	//nolint:exhaustive
	switch addrType {
	case iotago.AddressEd25519:
		dep.UnlockConditions = iotago.BasicOutputUnlockConditions{&iotago.AddressUnlockCondition{Address: RandEd25519Address()}}
	default:
		panic(fmt.Sprintf("invalid addr type: %d", addrType))
	}

	dep.Amount = RandBaseToken(10000) + 1

	return dep
}

func RandOutputType() iotago.OutputType {
	outputTypes := []iotago.OutputType{iotago.OutputBasic, iotago.OutputAccount, iotago.OutputAnchor, iotago.OutputFoundry, iotago.OutputNFT, iotago.OutputDelegation}

	return outputTypes[RandInt(len(outputTypes)-1)]
}

func RandOutput(outputType iotago.OutputType) iotago.Output {
	var addr iotago.Address
	if outputType == iotago.OutputFoundry {
		addr = RandAddress(iotago.AddressAccount)
	} else {
		addr = RandAddress(iotago.AddressEd25519)
	}

	return RandOutputOnAddress(outputType, addr)
}

func RandOutputOnAddress(outputType iotago.OutputType, address iotago.Address) iotago.Output {
	return RandOutputOnAddressWithAmount(outputType, address, RandBaseToken(iotago.MaxBaseToken))
}

func RandOutputOnAddressWithAmount(outputType iotago.OutputType, address iotago.Address, amount iotago.BaseToken) iotago.Output {
	var iotaOutput iotago.Output

	switch outputType {
	case iotago.OutputBasic:
		//nolint:forcetypeassert // we already checked the type
		iotaOutput = &iotago.BasicOutput{
			Amount: amount,
			UnlockConditions: iotago.BasicOutputUnlockConditions{
				&iotago.AddressUnlockCondition{
					Address: address,
				},
			},
			Features: iotago.BasicOutputFeatures{},
		}

	case iotago.OutputAccount:
		//nolint:forcetypeassert // we already checked the type
		iotaOutput = &iotago.AccountOutput{
			Amount:    amount,
			AccountID: RandAccountID(),
			UnlockConditions: iotago.AccountOutputUnlockConditions{
				&iotago.AddressUnlockCondition{
					Address: address,
				},
			},
			Features:          iotago.AccountOutputFeatures{},
			ImmutableFeatures: iotago.AccountOutputImmFeatures{},
		}

	case iotago.OutputAnchor:
		//nolint:forcetypeassert // we already checked the type
		iotaOutput = &iotago.AnchorOutput{
			Amount:   amount,
			AnchorID: RandAnchorID(),
			UnlockConditions: iotago.AnchorOutputUnlockConditions{
				&iotago.StateControllerAddressUnlockCondition{
					Address: address,
				},
				&iotago.GovernorAddressUnlockCondition{
					Address: address,
				},
			},
			Features:          iotago.AnchorOutputFeatures{},
			ImmutableFeatures: iotago.AnchorOutputImmFeatures{},
		}

	case iotago.OutputFoundry:
		if address.Type() != iotago.AddressAccount {
			panic("not an alias address")
		}
		supply := new(big.Int).SetUint64(RandUint64(math.MaxUint64))

		//nolint:forcetypeassert // we already checked the type
		iotaOutput = &iotago.FoundryOutput{
			Amount:       amount,
			SerialNumber: 0,
			TokenScheme: &iotago.SimpleTokenScheme{
				MintedTokens:  supply,
				MeltedTokens:  new(big.Int).SetBytes([]byte{0}),
				MaximumSupply: supply,
			},
			UnlockConditions: iotago.FoundryOutputUnlockConditions{
				&iotago.ImmutableAccountUnlockCondition{
					Address: address.(*iotago.AccountAddress),
				},
			},
			Features:          iotago.FoundryOutputFeatures{},
			ImmutableFeatures: iotago.FoundryOutputImmFeatures{},
		}

	case iotago.OutputNFT:
		//nolint:forcetypeassert // we already checked the type
		iotaOutput = &iotago.NFTOutput{
			Amount: amount,
			NFTID:  RandNFTID(),
			UnlockConditions: iotago.NFTOutputUnlockConditions{
				&iotago.AddressUnlockCondition{
					Address: address,
				},
			},
			Features:          iotago.NFTOutputFeatures{},
			ImmutableFeatures: iotago.NFTOutputImmFeatures{},
		}

	case iotago.OutputDelegation:
		//nolint:forcetypeassert // we already checked the type
		iotaOutput = &iotago.DelegationOutput{
			Amount:           amount,
			DelegatedAmount:  amount,
			DelegationID:     RandDelegationID(),
			ValidatorAddress: RandAccountAddress(),
			StartEpoch:       RandEpoch(),
			EndEpoch:         iotago.MaxEpochIndex,
			UnlockConditions: iotago.DelegationOutputUnlockConditions{
				&iotago.AddressUnlockCondition{
					Address: address,
				},
			},
		}

	default:
		panic("unhandled output type")
	}

	return iotaOutput
}
