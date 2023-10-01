package iotago_test

import (
	"math/big"
	"testing"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestTransactionEssence_DeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok",
			source: tpkg.RandTransaction(),
			target: &iotago.Transaction{},
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
			source: tpkg.RandSignedTransactionWithTransaction(tpkg.TestAPI,
				&iotago.Transaction{
					TransactionEssence: &iotago.TransactionEssence{
						NetworkID:     tpkg.TestNetworkID,
						ContextInputs: iotago.TxEssenceContextInputs{},
						Inputs:        inputIDs.UTXOInputs(),
						Allotments:    iotago.Allotments{},
						Capabilities:  iotago.TransactionCapabilitiesBitMask{},
					},
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
			target:    &iotago.SignedTransaction{},
			seriErr:   iotago.ErrNonUniqueChainOutputs,
			deSeriErr: nil,
		},
		{
			// we transition the same NFT twice
			name: "transition the same NFT twice",
			source: tpkg.RandSignedTransactionWithTransaction(tpkg.TestAPI,
				&iotago.Transaction{
					TransactionEssence: &iotago.TransactionEssence{
						NetworkID:    tpkg.TestNetworkID,
						Inputs:       inputIDs.UTXOInputs(),
						Capabilities: iotago.TransactionCapabilitiesBitMask{},
					},
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
			target:    &iotago.SignedTransaction{},
			seriErr:   iotago.ErrNonUniqueChainOutputs,
			deSeriErr: nil,
		},
		{
			// we transition the same Foundry twice
			name: "transition the same Foundry twice",
			source: tpkg.RandSignedTransactionWithTransaction(tpkg.TestAPI,
				&iotago.Transaction{
					TransactionEssence: &iotago.TransactionEssence{
						NetworkID:    tpkg.TestNetworkID,
						Inputs:       inputIDs.UTXOInputs(),
						Capabilities: iotago.TransactionCapabilitiesBitMask{},
					},
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
							SerialNumber: 1,
							TokenScheme: &iotago.SimpleTokenScheme{
								MintedTokens:  big.NewInt(50),
								MeltedTokens:  big.NewInt(0),
								MaximumSupply: big.NewInt(50),
							},
							Conditions: iotago.FoundryOutputUnlockConditions{
								&iotago.ImmutableAccountUnlockCondition{Address: accountAddress},
							},
							Features: nil,
						},
						&iotago.FoundryOutput{
							Amount:       OneMi,
							SerialNumber: 1,
							TokenScheme: &iotago.SimpleTokenScheme{
								MintedTokens:  big.NewInt(50),
								MeltedTokens:  big.NewInt(0),
								MaximumSupply: big.NewInt(50),
							},
							Conditions: iotago.FoundryOutputUnlockConditions{
								&iotago.ImmutableAccountUnlockCondition{Address: accountAddress},
							},
							Features: nil,
						},
					},
				}),
			target:    &iotago.SignedTransaction{},
			seriErr:   iotago.ErrNonUniqueChainOutputs,
			deSeriErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestAllotmentUniqueness(t *testing.T) {
	inputIDs := tpkg.RandOutputIDs(1)

	accountAddress := iotago.AccountAddressFromOutputID(inputIDs[0])
	accountID := accountAddress.AccountID()

	tests := []deSerializeTest{
		{
			name: "allot to the same account twice",
			source: tpkg.RandSignedTransactionWithTransaction(tpkg.TestAPI,
				&iotago.Transaction{
					TransactionEssence: &iotago.TransactionEssence{
						NetworkID:     tpkg.TestNetworkID,
						ContextInputs: iotago.TxEssenceContextInputs{},
						Inputs:        inputIDs.UTXOInputs(),
						Allotments: iotago.Allotments{
							&iotago.Allotment{
								AccountID: accountID,
								Value:     0,
							},
							&iotago.Allotment{
								AccountID: tpkg.RandAccountID(),
								Value:     12,
							},
							&iotago.Allotment{
								AccountID: accountID,
								Value:     12,
							},
						},
						Capabilities: iotago.TransactionCapabilitiesBitMask{},
					},
					Outputs: iotago.TxEssenceOutputs{
						tpkg.RandBasicOutput(iotago.AddressEd25519),
					},
				}),
			target:    &iotago.SignedTransaction{},
			seriErr:   serializer.ErrArrayValidationOrderViolatesLexicalOrder,
			deSeriErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}
