//nolint:scopelint
package iotago_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestTransactionEssence_DeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok",
			source: tpkg.RandTransaction(tpkg.ZeroCostTestAPI),
			target: &iotago.Transaction{API: tpkg.ZeroCostTestAPI},
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

	anchorAddress := iotago.AnchorAddressFromOutputID(inputIDs[0])
	anchorID := anchorAddress.AnchorID()

	nftAddress := iotago.NFTAddressFromOutputID(inputIDs[0])
	nftID := nftAddress.NFTID()

	tests := []deSerializeTest{
		{
			// we transition the same Account twice
			name: "transition the same Account twice",
			source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				&iotago.Transaction{
					API: tpkg.ZeroCostTestAPI,
					TransactionEssence: &iotago.TransactionEssence{
						NetworkID:     tpkg.TestNetworkID,
						ContextInputs: iotago.TxEssenceContextInputs{},
						Inputs:        inputIDs.UTXOInputs(),
						Allotments:    iotago.Allotments{},
						Capabilities:  iotago.TransactionCapabilitiesBitMask{},
					},
					Outputs: iotago.TxEssenceOutputs{
						&iotago.AccountOutput{
							Amount:    OneIOTA,
							AccountID: accountID,
							UnlockConditions: iotago.AccountOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: ident1},
							},
							Features: nil,
						},
						&iotago.AccountOutput{
							Amount:    OneIOTA,
							AccountID: accountID,
							UnlockConditions: iotago.AccountOutputUnlockConditions{
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
			// we transition the same Anchor twice
			name: "transition the same Anchor twice",
			source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				&iotago.Transaction{
					API: tpkg.ZeroCostTestAPI,
					TransactionEssence: &iotago.TransactionEssence{
						NetworkID:     tpkg.TestNetworkID,
						ContextInputs: iotago.TxEssenceContextInputs{},
						Inputs:        inputIDs.UTXOInputs(),
						Allotments:    iotago.Allotments{},
						Capabilities:  iotago.TransactionCapabilitiesBitMask{},
					},
					Outputs: iotago.TxEssenceOutputs{
						&iotago.AnchorOutput{
							Amount:   OneIOTA,
							AnchorID: anchorID,
							UnlockConditions: iotago.AnchorOutputUnlockConditions{
								&iotago.StateControllerAddressUnlockCondition{Address: ident1},
								&iotago.GovernorAddressUnlockCondition{Address: ident1},
							},
							Features: nil,
						},
						&iotago.AnchorOutput{
							Amount:   OneIOTA,
							AnchorID: anchorID,
							UnlockConditions: iotago.AnchorOutputUnlockConditions{
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
			source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				&iotago.Transaction{
					API: tpkg.ZeroCostTestAPI,
					TransactionEssence: &iotago.TransactionEssence{
						NetworkID:    tpkg.TestNetworkID,
						Inputs:       inputIDs.UTXOInputs(),
						Capabilities: iotago.TransactionCapabilitiesBitMask{},
					},
					Outputs: iotago.TxEssenceOutputs{
						&iotago.NFTOutput{
							Amount: OneIOTA,
							NFTID:  nftID,
							UnlockConditions: iotago.NFTOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: ident1},
							},
							Features: nil,
						},
						&iotago.NFTOutput{
							Amount: OneIOTA,
							NFTID:  nftID,
							UnlockConditions: iotago.NFTOutputUnlockConditions{
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
			source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				&iotago.Transaction{
					API: tpkg.ZeroCostTestAPI,
					TransactionEssence: &iotago.TransactionEssence{
						NetworkID:    tpkg.TestNetworkID,
						Inputs:       inputIDs.UTXOInputs(),
						Capabilities: iotago.TransactionCapabilitiesBitMask{},
					},
					Outputs: iotago.TxEssenceOutputs{
						&iotago.AccountOutput{
							Amount:    OneIOTA,
							AccountID: accountID,
							UnlockConditions: iotago.AccountOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: ident1},
							},
							Features: nil,
						},
						&iotago.FoundryOutput{
							Amount:       OneIOTA,
							SerialNumber: 1,
							TokenScheme: &iotago.SimpleTokenScheme{
								MintedTokens:  big.NewInt(50),
								MeltedTokens:  big.NewInt(0),
								MaximumSupply: big.NewInt(50),
							},
							UnlockConditions: iotago.FoundryOutputUnlockConditions{
								&iotago.ImmutableAccountUnlockCondition{Address: accountAddress},
							},
							Features: nil,
						},
						&iotago.FoundryOutput{
							Amount:       OneIOTA,
							SerialNumber: 1,
							TokenScheme: &iotago.SimpleTokenScheme{
								MintedTokens:  big.NewInt(50),
								MeltedTokens:  big.NewInt(0),
								MaximumSupply: big.NewInt(50),
							},
							UnlockConditions: iotago.FoundryOutputUnlockConditions{
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
			source: tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI,
				&iotago.Transaction{
					API: tpkg.ZeroCostTestAPI,
					TransactionEssence: &iotago.TransactionEssence{
						NetworkID:     tpkg.TestNetworkID,
						ContextInputs: iotago.TxEssenceContextInputs{},
						Inputs:        inputIDs.UTXOInputs(),
						Allotments: iotago.Allotments{
							&iotago.Allotment{
								AccountID: accountID,
								Mana:      0,
							},
							&iotago.Allotment{
								AccountID: accountID,
								Mana:      12,
							},
							&iotago.Allotment{
								AccountID: tpkg.RandAccountID(),
								Mana:      12,
							},
						},
						Capabilities: iotago.TransactionCapabilitiesBitMask{},
					},
					Outputs: iotago.TxEssenceOutputs{
						tpkg.RandBasicOutput(iotago.AddressEd25519),
					},
				}),
			target:    &iotago.SignedTransaction{},
			seriErr:   serix.ErrArrayValidationViolatesUniqueness,
			deSeriErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestTransactionEssenceCapabilitiesBitMask(t *testing.T) {

	type test struct {
		name    string
		tx      *iotago.Transaction
		wantErr error
	}

	randTransactionWithCapabilities := func(capabilities iotago.TransactionCapabilitiesBitMask) *iotago.Transaction {
		tx := tpkg.RandTransaction(tpkg.ZeroCostTestAPI)
		tx.Capabilities = capabilities
		return tx
	}

	tests := []*test{
		{
			name:    "ok - no trailing zero bytes",
			tx:      randTransactionWithCapabilities(iotago.TransactionCapabilitiesBitMask{0x01}),
			wantErr: nil,
		},
		{
			name:    "ok - empty capabilities",
			tx:      randTransactionWithCapabilities(iotago.TransactionCapabilitiesBitMask{}),
			wantErr: nil,
		},
		{
			name:    "fail - single zero byte",
			tx:      randTransactionWithCapabilities(iotago.TransactionCapabilitiesBitMask{0x00}),
			wantErr: iotago.ErrBitmaskTrailingZeroBytes,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.tx.SyntacticallyValidate(tpkg.ZeroCostTestAPI)
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)

				return
			}
			require.NoError(t, err)
		})
	}
}

func TestTransactionSyntacticMaxMana(t *testing.T) {
	type test struct {
		name    string
		tx      *iotago.Transaction
		wantErr error
	}

	basicOutputWithMana := func(mana iotago.Mana) *iotago.BasicOutput {
		return &iotago.BasicOutput{
			Amount: OneIOTA,
			Mana:   mana,
			UnlockConditions: iotago.BasicOutputUnlockConditions{
				&iotago.AddressUnlockCondition{
					Address: tpkg.RandEd25519Address(),
				},
			},
		}
	}

	allotmentWithMana := func(mana iotago.Mana) *iotago.Allotment {
		return &iotago.Allotment{
			Mana:      mana,
			AccountID: tpkg.RandAccountID(),
		}
	}

	var maxManaValue iotago.Mana = (1 << tpkg.ZeroCostTestAPI.ProtocolParameters().ManaParameters().BitsCount) - 1
	tests := []*test{
		{
			name: "ok - stored mana sum below max value",
			tx: tpkg.RandTransactionWithOptions(tpkg.ZeroCostTestAPI,
				func(tx *iotago.Transaction) {
					tx.Outputs = iotago.TxEssenceOutputs{basicOutputWithMana(1), basicOutputWithMana(maxManaValue - 1)}
				},
			),
			wantErr: nil,
		},
		{
			name: "fail - one output's stored mana exceeds max mana value",
			tx: tpkg.RandTransactionWithOptions(tpkg.ZeroCostTestAPI,
				func(tx *iotago.Transaction) {
					tx.Outputs = iotago.TxEssenceOutputs{basicOutputWithMana(maxManaValue + 1)}
				},
			),
			wantErr: iotago.ErrMaxManaExceeded,
		},
		{
			name: "fail - sum of stored mana exceeds max mana value",
			tx: tpkg.RandTransactionWithOptions(tpkg.ZeroCostTestAPI,
				func(tx *iotago.Transaction) {
					tx.Outputs = iotago.TxEssenceOutputs{basicOutputWithMana(maxManaValue - 1), basicOutputWithMana(maxManaValue - 1)}
				},
			),
			wantErr: iotago.ErrMaxManaExceeded,
		},
		{
			name: "ok - allotted mana sum below max value",
			tx: tpkg.RandTransactionWithOptions(tpkg.ZeroCostTestAPI,
				func(tx *iotago.Transaction) {
					tx.Allotments = iotago.Allotments{allotmentWithMana(1), allotmentWithMana(maxManaValue - 1)}
				},
			),
			wantErr: nil,
		},
		{
			name: "fail - one allotment's mana exceeds max value",
			tx: tpkg.RandTransactionWithOptions(tpkg.ZeroCostTestAPI,
				func(tx *iotago.Transaction) {
					tx.Allotments = iotago.Allotments{allotmentWithMana(maxManaValue + 1)}
				},
			),
			wantErr: iotago.ErrMaxManaExceeded,
		},
		{
			name: "fail - sum of allotted mana exceeds max value",
			tx: tpkg.RandTransactionWithOptions(tpkg.ZeroCostTestAPI,
				func(tx *iotago.Transaction) {
					tx.Allotments = iotago.Allotments{allotmentWithMana(maxManaValue - 1), allotmentWithMana(maxManaValue - 1)}
				},
			),
			wantErr: iotago.ErrMaxManaExceeded,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.tx.SyntacticallyValidate(tpkg.ZeroCostTestAPI)
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)

				return
			}
			require.NoError(t, err)
		})
	}
}

func TestTransactionInputUniqueness(t *testing.T) {
	type test struct {
		name    string
		inputs  iotago.TxEssenceInputs
		wantErr error
	}

	input1 := iotago.MustOutputIDFromHexString("0x2668778ef0362d601c36ea05c742185daa1740dfcdaee0acfde6a9806a1f2ed20d8566fd0800")
	input2 := iotago.MustOutputIDFromHexString("0x3f34a869f47f8454e7cb233943cd31a0e3bd8b9551b1390039ec582b0a196856eff185120400")
	input3 := iotago.MustOutputIDFromHexString("0xfdad2fee88cc4f1020848dce710124ac9060cdbee840a72b750c1f6901502576422f83b50500")
	// Differs from input3 only in the output index.
	input4 := iotago.MustOutputIDFromHexString("0xfdad2fee88cc4f1020848dce710124ac9060cdbee840a72b750c1f6901502576422f83b50600")

	tests := []test{
		{
			name: "ok - inputs unique",
			inputs: iotago.TxEssenceInputs{
				input3.UTXOInput(),
				input1.UTXOInput(),
				input4.UTXOInput(),
				input2.UTXOInput(),
			},
			wantErr: nil,
		},
		{
			name: "fail - duplicate inputs",
			inputs: iotago.TxEssenceInputs{
				input1.UTXOInput(),
				input2.UTXOInput(),
				input2.UTXOInput(),
			},
			wantErr: serix.ErrArrayValidationViolatesUniqueness,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			basicOutput := &iotago.BasicOutput{
				Amount: OneIOTA,
				UnlockConditions: iotago.BasicOutputUnlockConditions{
					&iotago.AddressUnlockCondition{
						Address: tpkg.RandEd25519Address(),
					},
				},
			}

			tx := &iotago.Transaction{
				API: tpkg.ZeroCostTestAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: test.inputs,
				},
				Outputs: iotago.TxEssenceOutputs{
					basicOutput,
				},
			}

			_, err := tpkg.ZeroCostTestAPI.Encode(tx, serix.WithValidation())
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)

				return
			}
			require.NoError(t, err)
		})
	}
}

func TestTransactionContextInputLexicalOrderAndUniqueness(t *testing.T) {
	type test struct {
		name          string
		contextInputs iotago.TxEssenceContextInputs
		wantErr       error
	}

	accountID1 := iotago.MustAccountIDFromHexString("0x2668778ef0362d601c36ea05c742185daa1740dfcdaee0acfde6a9806a1f2ed2")
	accountID2 := iotago.MustAccountIDFromHexString("0x4e7cb233943cd31a0e3bd8b92668778ef0362d601c36ea05c742039ec582b0af")
	commitmentID1 := iotago.MustCommitmentIDFromHexString("0x3f34a869f47f8454e7cb233943cd31a0e3bd8b9551b1390039ec582b0a196856e50500fd")
	commitmentID2 := iotago.MustCommitmentIDFromHexString("0x90039ec582b0a196856e50500fd3f34a869f47f8454e7cb233943cd31a0e3bd8b9551ac4")

	tests := []test{
		{
			name: "ok - context inputs lexically ordered and unique",
			contextInputs: iotago.TxEssenceContextInputs{
				&iotago.CommitmentInput{
					CommitmentID: commitmentID1,
				},
				&iotago.BlockIssuanceCreditInput{
					AccountID: accountID1,
				},
				&iotago.RewardInput{
					Index: 0,
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - context inputs lexically unordered",
			contextInputs: iotago.TxEssenceContextInputs{
				&iotago.BlockIssuanceCreditInput{
					AccountID: accountID1,
				},
				&iotago.CommitmentInput{
					CommitmentID: commitmentID1,
				},
				&iotago.RewardInput{
					Index: 0,
				},
			},
			wantErr: serix.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - block issuance credits inputs lexically unordered",
			contextInputs: iotago.TxEssenceContextInputs{
				&iotago.BlockIssuanceCreditInput{
					AccountID: accountID2,
				},
				&iotago.BlockIssuanceCreditInput{
					AccountID: accountID1,
				},
			},
			wantErr: serix.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - reward inputs lexically unordered",
			contextInputs: iotago.TxEssenceContextInputs{
				&iotago.RewardInput{
					Index: 5,
				},
				&iotago.RewardInput{
					Index: 3,
				},
			},
			wantErr: serix.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - commitment inputs lexically unordered",
			contextInputs: iotago.TxEssenceContextInputs{
				&iotago.CommitmentInput{
					CommitmentID: commitmentID2,
				},
				&iotago.CommitmentInput{
					CommitmentID: commitmentID1,
				},
			},
			wantErr: serix.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - duplicate block issuance credit inputs",
			contextInputs: iotago.TxEssenceContextInputs{
				&iotago.BlockIssuanceCreditInput{
					AccountID: accountID1,
				},
				&iotago.BlockIssuanceCreditInput{
					AccountID: accountID1,
				},
			},
			wantErr: serix.ErrArrayValidationViolatesUniqueness,
		},
		{
			name: "fail - duplicate reward inputs",
			contextInputs: iotago.TxEssenceContextInputs{
				&iotago.BlockIssuanceCreditInput{
					AccountID: accountID1,
				},
				&iotago.RewardInput{
					Index: 3,
				},
				&iotago.RewardInput{
					Index: 3,
				},
			},
			wantErr: serix.ErrArrayValidationViolatesUniqueness,
		},
		{
			name: "fail - duplicate commitment inputs",
			contextInputs: iotago.TxEssenceContextInputs{
				&iotago.CommitmentInput{
					CommitmentID: commitmentID1,
				},
				&iotago.CommitmentInput{
					CommitmentID: commitmentID1,
				},
				&iotago.RewardInput{
					Index: 3,
				},
			},
			wantErr: serix.ErrArrayValidationViolatesUniqueness,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			basicOutput := &iotago.BasicOutput{
				Amount: OneIOTA,
				UnlockConditions: iotago.BasicOutputUnlockConditions{
					&iotago.AddressUnlockCondition{
						Address: tpkg.RandEd25519Address(),
					},
				},
			}

			tx := &iotago.Transaction{
				API: tpkg.ZeroCostTestAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: iotago.TxEssenceInputs{
						tpkg.RandUTXOInput(),
						tpkg.RandUTXOInput(),
						tpkg.RandUTXOInput(),
					},
					ContextInputs: test.contextInputs,
				},
				Outputs: iotago.TxEssenceOutputs{
					basicOutput,
				},
			}

			_, err := tpkg.ZeroCostTestAPI.Encode(tx, serix.WithValidation())
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)

				return
			}
			require.NoError(t, err)
		})
	}
}
