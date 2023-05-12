package iotago_test

import (
	"errors"
	"math/big"
	"testing"

	"github.com/iotaledger/iota.go/v4/tpkg"

	"github.com/stretchr/testify/assert"

	iotago "github.com/iotaledger/iota.go/v4"
)

func TestTransactionEssenceSelector(t *testing.T) {
	_, err := iotago.TransactionEssenceSelector(100)
	assert.True(t, errors.Is(err, iotago.ErrUnknownTransactionEssenceType))
}

func TestTransactionEssence_DeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok",
			source: tpkg.RandTransactionEssence(),
			target: &iotago.TransactionEssence{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestChainConstrainedOutputUniqueness(t *testing.T) {
	ident1 := tpkg.RandEd25519Address()

	inputIDs := tpkg.RandOutputIDs(1)

	accountAddress := iotago.AccountAddressFromOutputID(inputIDs[0])
	accountID := accountAddress.AccountID()

	nftAddress := iotago.NFTAddressFromOutputID(inputIDs[0])
	nftID := nftAddress.NFTID()

	tests := []deSerializeTest{
		{
			// we transition the same Account twice
			name: "transition the same Account twice",
			source: tpkg.RandTransactionWithEssence(&iotago.TransactionEssence{
				NetworkID: tpkg.TestNetworkID,
				Inputs:    inputIDs.UTXOInputs(),
				Outputs: iotago.TxEssenceOutputs{
					&iotago.AccountOutput{
						Amount:    OneMi,
						AccountID: accountID,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident1},
						},
						Features: nil,
					},
					&iotago.AccountOutput{
						Amount:    OneMi,
						AccountID: accountID,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident1},
						},
						Features: nil,
					},
				},
			}),
			target:    &iotago.Transaction{},
			seriErr:   iotago.ErrNonUniqueChainOutputs,
			deSeriErr: nil,
		},
		{
			// we transition the same NFT twice
			name: "transition the same NFT twice",
			source: tpkg.RandTransactionWithEssence(&iotago.TransactionEssence{
				NetworkID: tpkg.TestNetworkID,
				Inputs:    inputIDs.UTXOInputs(),
				Outputs: iotago.TxEssenceOutputs{
					&iotago.NFTOutput{
						Amount: OneMi,
						NFTID:  nftID,
						Conditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
						Features: nil,
					},
					&iotago.NFTOutput{
						Amount: OneMi,
						NFTID:  nftID,
						Conditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
						Features: nil,
					},
				},
			}),
			target:    &iotago.Transaction{},
			seriErr:   iotago.ErrNonUniqueChainOutputs,
			deSeriErr: nil,
		},
		{
			// we transition the same Foundry twice
			name: "transition the same Foundry twice",
			source: tpkg.RandTransactionWithEssence(&iotago.TransactionEssence{
				NetworkID: tpkg.TestNetworkID,
				Inputs:    inputIDs.UTXOInputs(),
				Outputs: iotago.TxEssenceOutputs{
					&iotago.AccountOutput{
						Amount:    OneMi,
						AccountID: accountID,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident1},
						},
						Features: nil,
					},
					&iotago.FoundryOutput{
						Amount:       OneMi,
						NativeTokens: nil,
						SerialNumber: 1,
						TokenScheme: &iotago.SimpleTokenScheme{
							MintedTokens:  big.NewInt(50),
							MeltedTokens:  big.NewInt(0),
							MaximumSupply: big.NewInt(50),
						},
						Conditions: iotago.FoundryOutputUnlockConditions{
							&iotago.ImmutableAccountUnlockCondition{Address: &accountAddress},
						},
						Features: nil,
					},
					&iotago.FoundryOutput{
						Amount:       OneMi,
						NativeTokens: nil,
						SerialNumber: 1,
						TokenScheme: &iotago.SimpleTokenScheme{
							MintedTokens:  big.NewInt(50),
							MeltedTokens:  big.NewInt(0),
							MaximumSupply: big.NewInt(50),
						},
						Conditions: iotago.FoundryOutputUnlockConditions{
							&iotago.ImmutableAccountUnlockCondition{Address: &accountAddress},
						},
						Features: nil,
					},
				},
			}),
			target:    &iotago.Transaction{},
			seriErr:   iotago.ErrNonUniqueChainOutputs,
			deSeriErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}
