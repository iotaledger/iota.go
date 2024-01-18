//nolint:scopelint
package iotago_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/builder"
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
			deSeriErr: iotago.ErrNonUniqueChainOutputs,
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
			deSeriErr: iotago.ErrNonUniqueChainOutputs,
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
			deSeriErr: iotago.ErrNonUniqueChainOutputs,
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
			deSeriErr: iotago.ErrNonUniqueChainOutputs,
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
			seriErr:   iotago.ErrArrayValidationViolatesUniqueness,
			deSeriErr: iotago.ErrArrayValidationViolatesUniqueness,
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
					tx.Allotments.Sort()
				},
			),
			wantErr: nil,
		},
		{
			name: "fail - one allotment's mana exceeds max value",
			tx: tpkg.RandTransactionWithOptions(tpkg.ZeroCostTestAPI,
				func(tx *iotago.Transaction) {
					tx.Allotments = iotago.Allotments{allotmentWithMana(maxManaValue + 1)}
					tx.Allotments.Sort()
				},
			),
			wantErr: iotago.ErrMaxManaExceeded,
		},
		{
			name: "fail - sum of allotted mana exceeds max value",
			tx: tpkg.RandTransactionWithOptions(tpkg.ZeroCostTestAPI,
				func(tx *iotago.Transaction) {
					tx.Allotments = iotago.Allotments{allotmentWithMana(maxManaValue - 1), allotmentWithMana(maxManaValue - 1)}
					tx.Allotments.Sort()
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

type syntacticalSerializeTest struct {
	name        string
	transaction *iotago.SignedTransaction
	seriErr     error
	deseriErr   error
}

func (test *syntacticalSerializeTest) Run(t *testing.T) {
	serixData, err := tpkg.ZeroCostTestAPI.Encode(test.transaction, serix.WithValidation())
	if test.seriErr != nil {
		require.ErrorIs(t, err, test.seriErr, "serialization failed")

		serixData, err = tpkg.ZeroCostTestAPI.Encode(test.transaction)
		require.NoError(t, err)
	} else {
		require.NoError(t, err)
	}

	serixTarget := &iotago.SignedTransaction{}
	_, err = tpkg.ZeroCostTestAPI.Decode(serixData, serixTarget, serix.WithValidation())

	if test.deseriErr != nil {
		require.ErrorIs(t, err, test.deseriErr, "deserialization failed")
	} else {
		require.NoError(t, err)
	}
}

type transactionSerializeTest struct {
	name      string
	output    iotago.Output
	seriErr   error
	deseriErr error
}

func (test *transactionSerializeTest) Run(t *testing.T) {
	txBuilder := builder.NewTransactionBuilder(testAPI)
	txBuilder.WithTransactionCapabilities(
		iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanBurnNativeTokens(true)),
	)
	_, ident, addrKeys := tpkg.RandEd25519Identity()
	txBuilder.AddInput(&builder.TxInput{
		UnlockTarget: ident,
		InputID:      tpkg.RandUTXOInput().OutputID(),
		Input:        tpkg.RandBasicOutput(),
	})
	txBuilder.AddOutput(test.output)
	tx := lo.PanicOnErr(txBuilder.Build(iotago.NewInMemoryAddressSigner(addrKeys)))

	(&syntacticalSerializeTest{
		name:        test.name,
		transaction: tx,
		seriErr:     test.seriErr,
		deseriErr:   test.deseriErr,
	}).Run(t)
}

func TestTransactionInputUniqueness(t *testing.T) {
	type test struct {
		name      string
		inputs    iotago.TxEssenceInputs
		seriErr   error
		deseriErr error
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
			seriErr: nil,
		},
		{
			name: "fail - duplicate inputs",
			inputs: iotago.TxEssenceInputs{
				input1.UTXOInput(),
				input2.UTXOInput(),
				input2.UTXOInput(),
			},
			seriErr:   iotago.ErrInputUTXORefsNotUnique,
			deseriErr: iotago.ErrInputUTXORefsNotUnique,
		},
	}

	for _, test := range tests {

		stx := tpkg.RandSignedTransactionWithTransaction(tpkg.ZeroCostTestAPI, &iotago.Transaction{
			API: tpkg.ZeroCostTestAPI,
			TransactionEssence: &iotago.TransactionEssence{
				NetworkID: tpkg.ZeroCostTestAPI.ProtocolParameters().NetworkID(),
				Inputs:    test.inputs,
			},
			Outputs: iotago.TxEssenceOutputs{
				tpkg.RandBasicOutput(),
			},
		})

		tst := syntacticalSerializeTest{
			name:        test.name,
			transaction: stx,
			seriErr:     test.seriErr,
			deseriErr:   test.deseriErr,
		}

		t.Run(tst.name, tst.Run)
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
				&iotago.BlockIssuanceCreditInput{
					AccountID: accountID2,
				},
				&iotago.RewardInput{
					Index: 0,
				},
				&iotago.RewardInput{
					Index: 1,
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
			wantErr: iotago.ErrArrayValidationOrderViolatesLexicalOrder,
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
			wantErr: iotago.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - reward inputs lexically unordered",
			contextInputs: iotago.TxEssenceContextInputs{
				&iotago.RewardInput{
					Index: 1,
				},
				&iotago.RewardInput{
					Index: 0,
				},
			},
			wantErr: iotago.ErrArrayValidationOrderViolatesLexicalOrder,
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
			wantErr: iotago.ErrArrayValidationViolatesUniqueness,
		},
		{
			name: "fail - duplicate reward inputs",
			contextInputs: iotago.TxEssenceContextInputs{
				&iotago.BlockIssuanceCreditInput{
					AccountID: accountID1,
				},
				&iotago.RewardInput{
					Index: 0,
				},
				&iotago.RewardInput{
					Index: 0,
				},
			},
			wantErr: iotago.ErrArrayValidationViolatesUniqueness,
		},
		{
			// At most one commitment input is allowed.
			name: "fail - duplicate commitment inputs",
			contextInputs: iotago.TxEssenceContextInputs{
				&iotago.CommitmentInput{
					CommitmentID: commitmentID2,
				},
				&iotago.CommitmentInput{
					CommitmentID: commitmentID1,
				},
				&iotago.RewardInput{
					Index: 1,
				},
			},
			wantErr: iotago.ErrArrayValidationViolatesUniqueness,
		},
	}

	// We need to build the transaction manually, since the builder would sort the context inputs.
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
					NetworkID:    tpkg.ZeroCostTestAPI.ProtocolParameters().NetworkID(),
					CreationSlot: 5,
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

			stx := &iotago.SignedTransaction{
				API:         tpkg.ZeroCostTestAPI,
				Transaction: tx,
				Unlocks: iotago.Unlocks{
					tpkg.RandEd25519SignatureUnlock(),
					tpkg.RandEd25519SignatureUnlock(),
					tpkg.RandEd25519SignatureUnlock(),
				},
			}

			_, err := tpkg.ZeroCostTestAPI.Encode(stx, serix.WithValidation())
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)

				return
			}
			require.NoError(t, err)
		})
	}
}

// Tests that lexical order & uniqueness are checked for unlock conditions across all relevant outputs.
func TestTransactionOutputUnlockConditionsLexicalOrderAndUniqueness(t *testing.T) {
	// Unlock Cond Type 0
	addressUnlockCond := &iotago.AddressUnlockCondition{
		Address: tpkg.RandEd25519Address(),
	}
	// Unlock Cond Type 4
	stateCtrlUnlockCond := &iotago.StateControllerAddressUnlockCondition{
		Address: tpkg.RandEd25519Address(),
	}
	// Unlock Cond Type 5
	govUnlockCond := &iotago.GovernorAddressUnlockCondition{
		Address: tpkg.RandEd25519Address(),
	}

	// Unlock Cond Type 2
	timelockUnlockCond := &iotago.TimelockUnlockCondition{Slot: 1337}
	timelockUnlockCond2 := &iotago.TimelockUnlockCondition{Slot: 1000}

	// Unlock Cond Type 3
	expirationUnlockCond := &iotago.ExpirationUnlockCondition{
		ReturnAddress: tpkg.RandEd25519Address(),
		Slot:          1000,
	}

	tests := []transactionSerializeTest{
		{
			name: "fail - BasicOutput contains lexically unordered unlock conditions",
			output: &iotago.BasicOutput{
				Amount: 10_000_000,
				UnlockConditions: iotago.BasicOutputUnlockConditions{
					addressUnlockCond,
					expirationUnlockCond,
					timelockUnlockCond,
				},
				Features: iotago.BasicOutputFeatures{},
			},
			seriErr:   iotago.ErrArrayValidationOrderViolatesLexicalOrder,
			deseriErr: iotago.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - AnchorOutput contains lexically unordered unlock conditions",
			output: &iotago.AnchorOutput{
				Amount: 10_000_000,
				UnlockConditions: iotago.AnchorOutputUnlockConditions{
					govUnlockCond, stateCtrlUnlockCond,
				},
				Features: iotago.AnchorOutputFeatures{},
			},
			seriErr:   iotago.ErrArrayValidationOrderViolatesLexicalOrder,
			deseriErr: iotago.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - NFTOutput contains lexically unordered unlock conditions",
			output: &iotago.NFTOutput{
				Amount: 10_000_000,
				UnlockConditions: iotago.NFTOutputUnlockConditions{
					addressUnlockCond,
					expirationUnlockCond,
					timelockUnlockCond,
				},
				Features: iotago.NFTOutputFeatures{},
			},
			seriErr:   iotago.ErrArrayValidationOrderViolatesLexicalOrder,
			deseriErr: iotago.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - BasicOutput contains duplicate unlock conditions",
			output: &iotago.BasicOutput{
				Amount: 10_000_000,
				UnlockConditions: iotago.BasicOutputUnlockConditions{
					addressUnlockCond,
					timelockUnlockCond,
					timelockUnlockCond2,
				},
			},
			seriErr:   iotago.ErrArrayValidationViolatesUniqueness,
			deseriErr: iotago.ErrArrayValidationViolatesUniqueness,
		},
		{
			name: "fail - AnchorOutput contains duplicate unlock conditions",
			output: &iotago.AnchorOutput{
				Amount: 10_000_000,
				UnlockConditions: iotago.AnchorOutputUnlockConditions{
					stateCtrlUnlockCond, stateCtrlUnlockCond, govUnlockCond,
				},
				Features: iotago.AnchorOutputFeatures{},
			},
			// The errors don't match up here, but that's fine.
			seriErr:   iotago.ErrArrayValidationViolatesUniqueness,
			deseriErr: serializer.ErrArrayValidationMaxElementsExceeded,
		},
		{
			name: "fail - NFTOutput contains duplicate unlock conditions",
			output: &iotago.NFTOutput{
				Amount: 10_000_000,
				UnlockConditions: iotago.NFTOutputUnlockConditions{
					addressUnlockCond,
					timelockUnlockCond,
					timelockUnlockCond2,
				},
			},
			seriErr:   iotago.ErrArrayValidationViolatesUniqueness,
			deseriErr: iotago.ErrArrayValidationViolatesUniqueness,
		},
	}

	for _, test := range tests {
		t.Run(test.name, test.Run)
	}
}

// Tests that lexical order & uniqueness are checked for features across all relevant outputs.
func TestTransactionOutputFeatureLexicalOrderAndUniqueness(t *testing.T) {
	addressUnlockCond := &iotago.AddressUnlockCondition{
		Address: tpkg.RandEd25519Address(),
	}
	immutableAccountAddressUnlockCond := &iotago.ImmutableAccountUnlockCondition{
		Address: tpkg.RandAccountAddress(),
	}
	stateCtrlUnlockCond := &iotago.StateControllerAddressUnlockCondition{
		Address: tpkg.RandEd25519Address(),
	}
	govUnlockCond := &iotago.GovernorAddressUnlockCondition{
		Address: tpkg.RandEd25519Address(),
	}

	// Feature Type 1
	senderFeat := &iotago.SenderFeature{
		Address: tpkg.RandEd25519Address(),
	}
	senderFeat2 := &iotago.SenderFeature{
		Address: tpkg.RandEd25519Address(),
	}

	// Feature Type 2
	metadataFeat := &iotago.MetadataFeature{
		Entries: iotago.MetadataFeatureEntries{
			"key": []byte("val"),
		},
	}
	metadataFeat2 := &iotago.MetadataFeature{
		Entries: iotago.MetadataFeatureEntries{
			"entry": []byte("theval"),
		},
	}

	// Feature Type 3
	stateMetadataFeat := &iotago.StateMetadataFeature{
		Entries: iotago.StateMetadataFeatureEntries{
			"key": []byte("value"),
		},
	}

	// Feature Type 4
	tagFeat := &iotago.TagFeature{
		Tag: tpkg.RandBytes(3),
	}
	tagFeat2 := &iotago.TagFeature{
		Tag: tpkg.RandBytes(6),
	}

	// Feature Type 6
	nativeTokenFeat := tpkg.RandNativeTokenFeature()

	tests := []transactionSerializeTest{
		{
			name: "fail - BasicOutput contains lexically unordered features",
			output: &iotago.BasicOutput{
				Amount: 1337,
				Mana:   500,
				UnlockConditions: iotago.BasicOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: iotago.BasicOutputFeatures{
					tagFeat, senderFeat,
				},
			},
			seriErr:   iotago.ErrArrayValidationOrderViolatesLexicalOrder,
			deseriErr: iotago.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - AccountOutput contains lexically unordered features",
			output: &iotago.AccountOutput{
				Amount: 1_000_000,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					addressUnlockCond,
				},
				Features: iotago.AccountOutputFeatures{
					metadataFeat, senderFeat,
				},
			},
			seriErr:   iotago.ErrArrayValidationOrderViolatesLexicalOrder,
			deseriErr: iotago.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - AnchorOutput contains lexically unordered features",
			output: &iotago.AnchorOutput{
				Amount: 1_000_000,
				UnlockConditions: iotago.AnchorOutputUnlockConditions{
					stateCtrlUnlockCond,
					govUnlockCond,
				},
				Features: iotago.AnchorOutputFeatures{
					stateMetadataFeat, metadataFeat,
				},
			},
			seriErr:   iotago.ErrArrayValidationOrderViolatesLexicalOrder,
			deseriErr: iotago.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - FoundryOutput contains lexically unordered features",
			output: &iotago.FoundryOutput{
				Amount:      1_000_000,
				TokenScheme: tpkg.RandTokenScheme(),
				UnlockConditions: iotago.FoundryOutputUnlockConditions{
					immutableAccountAddressUnlockCond,
				},
				Features: iotago.FoundryOutputFeatures{
					nativeTokenFeat, metadataFeat,
				},
			},
			seriErr:   iotago.ErrArrayValidationOrderViolatesLexicalOrder,
			deseriErr: iotago.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - NFTOutput contains lexically unordered features",
			output: &iotago.NFTOutput{
				Amount: 1337,
				UnlockConditions: iotago.NFTOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: iotago.NFTOutputFeatures{
					tagFeat, senderFeat,
				},
			},
			seriErr:   iotago.ErrArrayValidationOrderViolatesLexicalOrder,
			deseriErr: iotago.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - BasicOutput contains duplicate features",
			output: &iotago.BasicOutput{
				Amount: 1337,
				UnlockConditions: iotago.BasicOutputUnlockConditions{
					addressUnlockCond,
				},
				Features: iotago.BasicOutputFeatures{
					tagFeat,
					tagFeat2,
				},
			},
			seriErr:   iotago.ErrArrayValidationViolatesUniqueness,
			deseriErr: iotago.ErrArrayValidationViolatesUniqueness,
		},
		{
			name: "fail - AccountOutput contains duplicate features",
			output: &iotago.AccountOutput{
				Amount: 1337,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					addressUnlockCond,
				},
				Features: iotago.AccountOutputFeatures{
					senderFeat,
					senderFeat2,
				},
			},
			seriErr:   iotago.ErrArrayValidationViolatesUniqueness,
			deseriErr: iotago.ErrArrayValidationViolatesUniqueness,
		},
		{
			name: "fail - AnchorOutput contains duplicate features",
			output: &iotago.AnchorOutput{
				Amount: 1337,
				UnlockConditions: iotago.AnchorOutputUnlockConditions{
					stateCtrlUnlockCond,
					govUnlockCond,
				},
				Features: iotago.AnchorOutputFeatures{
					senderFeat,
					senderFeat2,
				},
			},
			seriErr:   iotago.ErrArrayValidationViolatesUniqueness,
			deseriErr: iotago.ErrArrayValidationViolatesUniqueness,
		},
		{
			name: "fail - FoundryOutput contains duplicate features",
			output: &iotago.FoundryOutput{
				Amount:      1_000_000,
				TokenScheme: tpkg.RandTokenScheme(),
				UnlockConditions: iotago.FoundryOutputUnlockConditions{
					immutableAccountAddressUnlockCond,
				},
				Features: iotago.FoundryOutputFeatures{
					metadataFeat, metadataFeat2,
				},
			},
			seriErr:   iotago.ErrArrayValidationViolatesUniqueness,
			deseriErr: iotago.ErrArrayValidationViolatesUniqueness,
		},
		{
			name: "fail - NFTOutput contains duplicate features",
			output: &iotago.NFTOutput{
				Amount: 1337,
				UnlockConditions: iotago.NFTOutputUnlockConditions{
					addressUnlockCond,
				},
				Features: iotago.NFTOutputFeatures{
					tagFeat,
					tagFeat2,
				},
			},
			seriErr:   iotago.ErrArrayValidationViolatesUniqueness,
			deseriErr: iotago.ErrArrayValidationViolatesUniqueness,
		},
	}

	for _, test := range tests {
		t.Run(test.name, test.Run)
	}
}

// Tests that lexical order & uniqueness are checked for immutable features across all relevant outputs.
func TestTransactionOutputImmutableFeatureLexicalOrderAndUniqueness(t *testing.T) {
	addressUnlockCond := &iotago.AddressUnlockCondition{
		Address: tpkg.RandEd25519Address(),
	}
	stateCtrlUnlockCond := &iotago.StateControllerAddressUnlockCondition{
		Address: tpkg.RandEd25519Address(),
	}
	govUnlockCond := &iotago.GovernorAddressUnlockCondition{
		Address: tpkg.RandEd25519Address(),
	}

	// Feature Type 1
	issuerFeat := &iotago.IssuerFeature{
		Address: tpkg.RandEd25519Address(),
	}
	// Create a second issuer feature to ensure uniqueness is checked based on the type of the feature.
	issuerFeat2 := &iotago.IssuerFeature{
		Address: tpkg.RandEd25519Address(),
	}

	// Feature Type 2
	metadataFeat := &iotago.MetadataFeature{
		Entries: iotago.MetadataFeatureEntries{
			"key": []byte("val"),
		},
	}

	tests := []transactionSerializeTest{
		{
			name: "fail - AccountOutput contains lexically unordered immutable features",
			output: &iotago.AccountOutput{
				Amount: 1_000_000,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					addressUnlockCond,
				},
				ImmutableFeatures: iotago.AccountOutputImmFeatures{
					metadataFeat, issuerFeat,
				},
			},
			seriErr:   iotago.ErrArrayValidationOrderViolatesLexicalOrder,
			deseriErr: iotago.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - NFTOutput contains lexically unordered immutable features",
			output: &iotago.NFTOutput{
				Amount: 1_000_000,
				UnlockConditions: iotago.NFTOutputUnlockConditions{
					addressUnlockCond,
				},
				ImmutableFeatures: iotago.NFTOutputImmFeatures{
					metadataFeat, issuerFeat,
				},
			},
			seriErr:   iotago.ErrArrayValidationOrderViolatesLexicalOrder,
			deseriErr: iotago.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - AnchorOutput contains lexically unordered immutable features",
			output: &iotago.AnchorOutput{
				Amount: 1_000_000,
				UnlockConditions: iotago.AnchorOutputUnlockConditions{
					stateCtrlUnlockCond,
					govUnlockCond,
				},
				ImmutableFeatures: iotago.AnchorOutputImmFeatures{
					metadataFeat, issuerFeat,
				},
			},
			seriErr:   iotago.ErrArrayValidationOrderViolatesLexicalOrder,
			deseriErr: iotago.ErrArrayValidationOrderViolatesLexicalOrder,
		},
		{
			name: "fail - AccountOutput contains duplicate immutable features",
			output: &iotago.AccountOutput{
				Amount: 1_000_000,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					addressUnlockCond,
				},
				ImmutableFeatures: iotago.AccountOutputImmFeatures{
					issuerFeat, issuerFeat2,
				},
			},
			seriErr:   iotago.ErrArrayValidationViolatesUniqueness,
			deseriErr: iotago.ErrArrayValidationViolatesUniqueness,
		},
		{
			name: "fail - NFTOutput contains duplicate immutable features",
			output: &iotago.NFTOutput{
				Amount: 1_000_000,
				UnlockConditions: iotago.NFTOutputUnlockConditions{
					addressUnlockCond,
				},
				ImmutableFeatures: iotago.NFTOutputImmFeatures{
					issuerFeat, issuerFeat2,
				},
			},
			seriErr:   iotago.ErrArrayValidationViolatesUniqueness,
			deseriErr: iotago.ErrArrayValidationViolatesUniqueness,
		},
		{
			name: "fail - AnchorOutput contains duplicate immutable features",
			output: &iotago.AnchorOutput{
				Amount: 1_000_000,
				UnlockConditions: iotago.AnchorOutputUnlockConditions{
					stateCtrlUnlockCond,
					govUnlockCond,
				},
				ImmutableFeatures: iotago.AnchorOutputImmFeatures{
					issuerFeat, issuerFeat2,
				},
			},
			seriErr:   iotago.ErrArrayValidationViolatesUniqueness,
			deseriErr: iotago.ErrArrayValidationViolatesUniqueness,
		},
	}

	for _, test := range tests {
		t.Run(test.name, test.Run)
	}
}
