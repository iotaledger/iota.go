//nolint:forcetypeassert,dupl,nlreturn,scopelint
package nova_test

import (
	"bytes"
	"crypto/ed25519"
	"math/big"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/builder"
	"github.com/iotaledger/iota.go/v4/tpkg"
	"github.com/iotaledger/iota.go/v4/vm"
	"github.com/iotaledger/iota.go/v4/vm/nova"
)

const (
	OneIOTA iotago.BaseToken = 1_000_000

	betaPerYear                  float64 = 1 / 3.0
	slotsPerEpochExponent                = 13
	slotDurationSeconds                  = 10
	bitsCount                            = 63
	generationRate                       = 1
	generationRateExponent               = 27
	decayFactorsExponent                 = 32
	decayFactorEpochsSumExponent         = 20
)

var (
	novaVM = nova.NewVirtualMachine()

	schedulerRate   iotago.WorkScore = 100000
	testProtoParams                  = iotago.NewV3ProtocolParameters(
		iotago.WithNetworkOptions("test", "test"),
		iotago.WithSupplyOptions(tpkg.TestTokenSupply, 100, 1, 10, 100, 100, 100),
		iotago.WithWorkScoreOptions(1, 100, 20, 20, 20, 20, 100, 100, 100, 200),
		iotago.WithTimeProviderOptions(0, 100, slotDurationSeconds, slotsPerEpochExponent),
		iotago.WithManaOptions(bitsCount,
			generationRate,
			generationRateExponent,
			tpkg.ManaDecayFactors(betaPerYear, 1<<slotsPerEpochExponent, slotDurationSeconds, decayFactorsExponent),
			decayFactorsExponent,
			tpkg.ManaDecayFactorEpochsSum(betaPerYear, 1<<slotsPerEpochExponent, slotDurationSeconds, decayFactorEpochsSumExponent),
			decayFactorEpochsSumExponent,
		),
		iotago.WithStakingOptions(10, 10, 10),
		iotago.WithLivenessOptions(15, 30, 10, 20, 24),
		iotago.WithCongestionControlOptions(500, 500, 500, 8*schedulerRate, 5*schedulerRate, schedulerRate, 1000, 100),
	)

	testAPI = iotago.V3API(testProtoParams)
)

func TestNFTTransition(t *testing.T) {
	_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()

	inputIDs := tpkg.RandOutputIDs(1)
	inputs := vm.InputSet{
		inputIDs[0]: &iotago.NFTOutput{
			Amount: OneIOTA,
			NFTID:  iotago.NFTID{},
			UnlockConditions: iotago.NFTOutputUnlockConditions{
				&iotago.AddressUnlockCondition{Address: ident1},
			},
			Features: nil,
		},
	}

	nftAddr := iotago.NFTAddressFromOutputID(inputIDs[0])
	nftID := nftAddr.NFTID()

	transaction := &iotago.Transaction{
		API: testAPI,
		TransactionEssence: &iotago.TransactionEssence{
			Inputs:       inputIDs.UTXOInputs(),
			Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
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
		},
	}

	sigs, err := transaction.Sign(ident1AddrKeys)
	require.NoError(t, err)

	tx := &iotago.SignedTransaction{
		API:         testAPI,
		Transaction: transaction,
		Unlocks: iotago.Unlocks{
			&iotago.SignatureUnlock{Signature: sigs[0]},
		},
	}

	require.NoError(t, validateAndExecuteSignedTransaction(tx, vm.ResolvedInputs{InputSet: inputs}))
}

func TestCirculatingSupplyMelting(t *testing.T) {
	_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
	accountIdent1 := tpkg.RandAccountAddress()

	inputIDs := tpkg.RandOutputIDs(3)
	inputs := vm.InputSet{
		inputIDs[0]: &iotago.BasicOutput{
			Amount: OneIOTA,
			UnlockConditions: iotago.BasicOutputUnlockConditions{
				&iotago.AddressUnlockCondition{Address: ident1},
			},
		},
		inputIDs[1]: &iotago.AccountOutput{
			Amount:         OneIOTA,
			AccountID:      accountIdent1.AccountID(),
			FoundryCounter: 1,
			UnlockConditions: iotago.AccountOutputUnlockConditions{
				&iotago.AddressUnlockCondition{Address: ident1},
			},
			Features: nil,
		},
		inputIDs[2]: &iotago.FoundryOutput{
			Amount:       OneIOTA,
			SerialNumber: 1,
			TokenScheme: &iotago.SimpleTokenScheme{
				MintedTokens:  big.NewInt(50),
				MeltedTokens:  big.NewInt(0),
				MaximumSupply: big.NewInt(50),
			},
			UnlockConditions: iotago.FoundryOutputUnlockConditions{
				&iotago.ImmutableAccountUnlockCondition{Address: accountIdent1},
			},
			Features: nil,
		},
	}

	// set input BasicOutput NativeToken to 50 which get melted
	foundryNativeTokenID := inputs[inputIDs[2]].(*iotago.FoundryOutput).MustNativeTokenID()
	inputs[inputIDs[0]].(*iotago.BasicOutput).Features.Upsert(&iotago.NativeTokenFeature{
		ID:     foundryNativeTokenID,
		Amount: new(big.Int).SetInt64(50),
	})

	transaction := &iotago.Transaction{
		API: testAPI,
		TransactionEssence: &iotago.TransactionEssence{
			Inputs:       inputIDs.UTXOInputs(),
			Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
		},
		Outputs: iotago.TxEssenceOutputs{
			&iotago.AccountOutput{
				Amount:         OneIOTA,
				AccountID:      accountIdent1.AccountID(),
				FoundryCounter: 1,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: ident1},
				},
				Features: nil,
			},
			&iotago.FoundryOutput{
				Amount:       2 * OneIOTA,
				SerialNumber: 1,
				TokenScheme: &iotago.SimpleTokenScheme{
					MintedTokens:  big.NewInt(50),
					MeltedTokens:  big.NewInt(50),
					MaximumSupply: big.NewInt(50),
				},
				UnlockConditions: iotago.FoundryOutputUnlockConditions{
					&iotago.ImmutableAccountUnlockCondition{Address: accountIdent1},
				},
				Features: nil,
			},
		},
	}

	sigs, err := transaction.Sign(ident1AddrKeys)
	require.NoError(t, err)

	tx := &iotago.SignedTransaction{
		API:         testAPI,
		Transaction: transaction,
		Unlocks: iotago.Unlocks{
			&iotago.SignatureUnlock{Signature: sigs[0]},
			&iotago.ReferenceUnlock{Reference: 0},
			&iotago.AccountUnlock{Reference: 1},
		},
	}

	require.NoError(t, validateAndExecuteSignedTransaction(tx, vm.ResolvedInputs{InputSet: inputs}))
}

func TestNovaTransactionExecution(t *testing.T) {
	type test struct {
		name           string
		vmParams       *vm.Params
		resolvedInputs vm.ResolvedInputs
		tx             *iotago.SignedTransaction
		wantErr        error
	}
	tests := []*test{
		// ok
		func() *test {
			var (
				_, ident1, ident1AddrKeys = tpkg.RandEd25519Identity()
				_, ident2, ident2AddrKeys = tpkg.RandEd25519Identity()
				_, ident3, ident3AddrKeys = tpkg.RandEd25519Identity()
				_, ident4, ident4AddrKeys = tpkg.RandEd25519Identity()
				_, ident5, _              = tpkg.RandEd25519Identity()
			)

			var (
				defaultAmount        = OneIOTA
				storageDepositReturn = OneIOTA / 2
				nativeTokenTransfer1 = tpkg.RandNativeTokenFeature()
				nativeTokenTransfer2 = tpkg.RandNativeTokenFeature()
			)

			var (
				nft1ID = tpkg.Rand32ByteArray()
				nft2ID = tpkg.Rand32ByteArray()
			)

			inputIDs := tpkg.RandOutputIDs(18)

			account1InputID := inputIDs[6]

			account1AccountID := iotago.AccountIDFromOutputID(account1InputID)
			account1AccountAddress := account1AccountID.ToAddress().(*iotago.AccountAddress)

			anchor1InputID := inputIDs[8]
			anchor2InputID := inputIDs[9]

			anchor1AnchorID := iotago.AnchorIDFromOutputID(anchor1InputID)
			anchor2AnchorID := iotago.AnchorIDFromOutputID(anchor2InputID)

			foundry1InputID := inputIDs[11]
			foundry2InputID := inputIDs[12]
			foundry3InputID := inputIDs[13]
			foundry4InputID := inputIDs[14]

			nft1InputID := inputIDs[15]

			inputs := vm.InputSet{
				// basic output with no features [defaultAmount] (owned by ident1)
				// => output 0: change ownership to ident5
				inputIDs[0]: &iotago.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},

				// basic output with native token feature - nativeTokenTransfer1 [defaultAmount] (owned by ident2)
				// => output 1: change ownership to ident3
				inputIDs[1]: &iotago.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident2},
					},
					Features: iotago.BasicOutputFeatures{
						nativeTokenTransfer1,
					},
				},

				// basic output with native token feature - nativeTokenTransfer2 [defaultAmount] (owned by ident2)
				// => output 2: change ownership to ident4
				inputIDs[2]: &iotago.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident2},
					},
					Features: iotago.BasicOutputFeatures{
						nativeTokenTransfer2,
					},
				},

				// basic output with expiration unlock condition - slot: 500, return: ident1 [defaultAmount] (originally owned by ident2 => creation slot 750 => owned by ident1)
				// => output 3: remove expiration unlock condition
				inputIDs[3]: &iotago.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident2},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: ident1,
							Slot:          500,
						},
					},
				},

				// basic output with timelock unlock condition - slot: 500 [defaultAmount] (owned by ident2 => creation slot 750 => can be unlocked)
				// => output 4: remove timelock unlock condition
				inputIDs[4]: &iotago.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident2},
						&iotago.TimelockUnlockCondition{
							Slot: 500,
						},
					},
				},

				// basic output [defaultAmount + storageDepositReturn] (owned by ident2 => creation slot 750 => can be unlocked, owned by ident2)
				//					 storage deposit return unlock condition - return: ident1
				// 			       	 timelock unlock condition 				 - slot 500
				// 			       	 expiration unlock condition 			 - slot: 900, return: ident1
				// => output 5: storageDepositReturn to ident1
				// => output 14: defaultAmount
				inputIDs[5]: &iotago.BasicOutput{
					Amount: defaultAmount + storageDepositReturn,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident2},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: ident1,
							Amount:        storageDepositReturn,
						},
						&iotago.TimelockUnlockCondition{
							Slot: 500,
						},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: ident1,
							Slot:          900,
						},
					},
				},

				// account output with no features - foundry counter 5 [defaultAmount] (owned by ident3) => going to be transitioned
				// => output 6: output transition (foundry counter 5 => 6, added metadata)
				account1InputID: &iotago.AccountOutput{
					Amount:         defaultAmount,
					AccountID:      account1AccountID,
					FoundryCounter: 5,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident3},
					},
					Features: nil,
				},

				// account output with no features [defaultAmount] (owned by ident3) => going to be destroyed
				// => output 7: destroyed and new account output created
				inputIDs[7]: &iotago.AccountOutput{
					Amount:         defaultAmount,
					AccountID:      iotago.AccountID{},
					FoundryCounter: 0,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident3},
					},
					Features: nil,
				},

				// anchor output with no features - state index 0 [defaultAmount] (owned by - state: ident3, gov: ident4) => going to be governance transitioned
				// => output 8: governance transition (added metadata)
				anchor1InputID: &iotago.AnchorOutput{
					Amount:     defaultAmount,
					AnchorID:   anchor1AnchorID,
					StateIndex: 0,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident3},
						&iotago.GovernorAddressUnlockCondition{Address: ident4},
					},
					Features: iotago.AnchorOutputFeatures{
						&iotago.StateMetadataFeature{Entries: iotago.StateMetadataFeatureEntries{"data": []byte("gov transitioning")}},
					},
				},

				// anchor output with no features - state index 5 [defaultAmount] (owned by - state: ident3, gov: ident4) => going to be state transitioned
				// => output 9: state transition (state index 5 => 6, changed state metadata)
				anchor2InputID: &iotago.AnchorOutput{
					Amount:     defaultAmount,
					AnchorID:   anchor2AnchorID,
					StateIndex: 5,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident3},
						&iotago.GovernorAddressUnlockCondition{Address: ident4},
					},
					Features: iotago.AnchorOutputFeatures{
						&iotago.StateMetadataFeature{Entries: iotago.StateMetadataFeatureEntries{"data": []byte("state transitioning")}},
					},
				},

				// anchor output with no features - state index 0 [defaultAmount] (owned by - state: ident3, gov: ident3) => going to be destroyed
				// => output 10: destroyed and new anchor output created
				inputIDs[10]: &iotago.AnchorOutput{
					Amount:     defaultAmount,
					AnchorID:   iotago.AnchorID{},
					StateIndex: 0,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident3},
						&iotago.GovernorAddressUnlockCondition{Address: ident3},
					},
					Features: iotago.AnchorOutputFeatures{
						&iotago.StateMetadataFeature{Entries: iotago.StateMetadataFeatureEntries{"data": []byte("going to be destroyed")}},
					},
				},

				// foundry output - serialNumber: 1, minted: 100, melted: 0, max: 1000 [defaultAmount] (owned by account1AccountAddress)
				// => output 11: mint 100 new tokens
				foundry1InputID: &iotago.FoundryOutput{
					Amount:       defaultAmount,
					SerialNumber: 1,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(100),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
					UnlockConditions: iotago.FoundryOutputUnlockConditions{
						&iotago.ImmutableAccountUnlockCondition{Address: account1AccountAddress},
					},
					Features: nil,
				},

				// foundry output - serialNumber: 2, minted: 100, melted: 0, max: 1000 [defaultAmount] (owned by account1AccountAddress)
				//				  - native token balance later updated to 100 (still on input side)
				// => output 12: melt 50 tokens
				foundry2InputID: &iotago.FoundryOutput{
					Amount:       defaultAmount,
					SerialNumber: 2,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(100),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
					UnlockConditions: iotago.FoundryOutputUnlockConditions{
						&iotago.ImmutableAccountUnlockCondition{Address: account1AccountAddress},
					},
					Features: iotago.FoundryOutputFeatures{
						// native token feature added later
					},
				},

				// foundry output - serialNumber: 3, minted: 100, melted: 0, max: 1000 [defaultAmount] (owned by account1AccountAddress)
				// => output 13: add metadata
				foundry3InputID: &iotago.FoundryOutput{
					Amount:       defaultAmount,
					SerialNumber: 3,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(100),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
					UnlockConditions: iotago.FoundryOutputUnlockConditions{
						&iotago.ImmutableAccountUnlockCondition{Address: account1AccountAddress},
					},
					Features: nil,
				},

				// foundry output - serialNumber: 4, minted: 100, melted: 0, max: 1000 [defaultAmount] (owned by account1AccountAddress)
				//				  - native token balance later updated to 50 (still on input side)
				// => output 15: foundry destroyed
				foundry4InputID: &iotago.FoundryOutput{
					Amount:       defaultAmount,
					SerialNumber: 4,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(100),
						MeltedTokens:  big.NewInt(50),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
					UnlockConditions: iotago.FoundryOutputUnlockConditions{
						&iotago.ImmutableAccountUnlockCondition{Address: account1AccountAddress},
					},
					Features: nil,
				},

				// NFT output with issuer (ident3) and immutable metadata feature [defaultAmount] (owned by ident3) => going to be transferred to ident4
				// => output 16: transfer to ident4
				nft1InputID: &iotago.NFTOutput{
					Amount: defaultAmount,
					NFTID:  nft1ID,
					UnlockConditions: iotago.NFTOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident3},
					},
					Features: iotago.NFTOutputFeatures{
						&iotago.IssuerFeature{Address: ident3},
					},
					ImmutableFeatures: iotago.NFTOutputImmFeatures{
						&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("transfer to 4")}},
					},
				},

				// NFT output with immutable features [defaultAmount] (owned by ident4) => going to be destroyed
				// => output 17: destroyed and new NFT output created
				inputIDs[16]: &iotago.NFTOutput{
					Amount: defaultAmount,
					NFTID:  nft2ID,
					UnlockConditions: iotago.NFTOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident4},
					},
					Features: iotago.NFTOutputFeatures{
						&iotago.IssuerFeature{Address: ident3},
					},
					ImmutableFeatures: iotago.NFTOutputImmFeatures{
						&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("going to be destroyed")}},
					},
				},

				// basic output with no features [defaultAmount] (owned by nft1ID)
				// => output 18: change ownership to ident5
				inputIDs[17]: &iotago.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: iotago.NFTID(nft1ID).ToAddress()},
					},
				},
			}

			foundry1Ident3NativeTokenID := inputs[foundry1InputID].(*iotago.FoundryOutput).MustNativeTokenID()
			foundry2Ident3NativeTokenID := inputs[foundry2InputID].(*iotago.FoundryOutput).MustNativeTokenID()
			foundry4Ident3NativeTokenID := inputs[foundry4InputID].(*iotago.FoundryOutput).MustNativeTokenID()

			inputs[foundry2InputID].(*iotago.FoundryOutput).Features.Upsert(&iotago.NativeTokenFeature{
				ID:     foundry2Ident3NativeTokenID,
				Amount: big.NewInt(100),
			})

			inputs[foundry4InputID].(*iotago.FoundryOutput).Features.Upsert(&iotago.NativeTokenFeature{
				ID:     foundry4Ident3NativeTokenID,
				Amount: big.NewInt(50),
			})

			// new foundry output - serialNumber: 6, minted: 100, melted: 0, max: 1000 (owned by account1AccountAddress)
			//					  - native token balance 100
			newFoundryWithInitialSupply := &iotago.FoundryOutput{
				Amount:       defaultAmount,
				SerialNumber: 6,
				TokenScheme: &iotago.SimpleTokenScheme{
					MintedTokens:  big.NewInt(100),
					MeltedTokens:  big.NewInt(0),
					MaximumSupply: new(big.Int).SetInt64(1000),
				},
				UnlockConditions: iotago.FoundryOutputUnlockConditions{
					&iotago.ImmutableAccountUnlockCondition{Address: account1AccountAddress},
				},
				Features: nil,
			}
			newFoundryNativeTokenID := newFoundryWithInitialSupply.MustNativeTokenID()
			newFoundryWithInitialSupply.Features.Upsert(&iotago.NativeTokenFeature{
				ID:     newFoundryNativeTokenID,
				Amount: big.NewInt(100),
			})

			creationSlot := iotago.SlotIndex(750)
			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs:       inputIDs.UTXOInputs(),
					CreationSlot: creationSlot,
					Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
				},
				Outputs: iotago.TxEssenceOutputs{
					// basic output [defaultAmount] (owned by ident5)
					// => input 0
					&iotago.BasicOutput{
						Amount: defaultAmount,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident5},
						},
					},

					// basic output with native token feature - nativeTokenTransfer1 [defaultAmount] (owned by ident3)
					// => input 1
					&iotago.BasicOutput{
						Amount: defaultAmount,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident3},
						},
						Features: iotago.BasicOutputFeatures{
							nativeTokenTransfer1,
						},
					},

					// basic output with native token feature - nativeTokenTransfer2 [defaultAmount] (owned by ident4)
					// => input 2
					&iotago.BasicOutput{
						Amount: defaultAmount,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident4},
						},
						Features: iotago.BasicOutputFeatures{
							nativeTokenTransfer2,
						},
					},

					// basic output [defaultAmount] (owned by ident2)
					// => input 3
					&iotago.BasicOutput{
						Amount: defaultAmount,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
					},

					// basic output [defaultAmount] (owned by ident2)
					// => input 4
					&iotago.BasicOutput{
						Amount: defaultAmount,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
					},

					// basic output [storageDepositReturn] (owned by ident1)
					// => input 5
					&iotago.BasicOutput{
						Amount: storageDepositReturn,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},

					// transitioned account output [defaultAmount] (owned by ident3)
					// => input 6
					&iotago.AccountOutput{
						Amount:         defaultAmount,
						AccountID:      account1AccountID,
						FoundryCounter: 6,
						UnlockConditions: iotago.AccountOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident3},
						},
						Features: iotago.AccountOutputFeatures{
							&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("transitioned")}},
						},
					},

					// new account output [defaultAmount] (owned by ident3)
					// => input 7
					&iotago.AccountOutput{
						Amount:         defaultAmount,
						AccountID:      iotago.AccountID{},
						FoundryCounter: 0,
						UnlockConditions: iotago.AccountOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident3},
						},
						Features: iotago.AccountOutputFeatures{
							&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("new")}},
						},
					},

					// governance transitioned anchor output [defaultAmount] (owned by - state: ident3, gov: ident4)
					// => input 8
					&iotago.AnchorOutput{
						Amount:     defaultAmount,
						AnchorID:   anchor1AnchorID,
						StateIndex: 0,
						UnlockConditions: iotago.AnchorOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident3},
							&iotago.GovernorAddressUnlockCondition{Address: ident4},
						},
						Features: iotago.AnchorOutputFeatures{
							&iotago.StateMetadataFeature{Entries: iotago.StateMetadataFeatureEntries{"data": []byte("gov transitioning")}},
							&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("the gov mutation on this output")}},
						},
					},

					// state transitioned anchor output [defaultAmount] (owned by - state: ident3, gov: ident4)
					// => input 9
					&iotago.AnchorOutput{
						Amount:     defaultAmount,
						AnchorID:   anchor2AnchorID,
						StateIndex: 6,
						UnlockConditions: iotago.AnchorOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident3},
							&iotago.GovernorAddressUnlockCondition{Address: ident4},
						},
						Features: iotago.AnchorOutputFeatures{
							&iotago.StateMetadataFeature{Entries: iotago.StateMetadataFeatureEntries{
								"data":  []byte("state transitioning"),
								"added": []byte("next state"),
							}},
						},
					},

					// new anchor output [defaultAmount] (owned by - state: ident3, gov: ident4)
					// => input 10
					&iotago.AnchorOutput{
						Amount:     defaultAmount,
						AnchorID:   iotago.AnchorID{},
						StateIndex: 0,
						UnlockConditions: iotago.AnchorOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident3},
							&iotago.GovernorAddressUnlockCondition{Address: ident4},
						},
						Features: iotago.AnchorOutputFeatures{
							&iotago.StateMetadataFeature{Entries: iotago.StateMetadataFeatureEntries{"data": []byte("a new anchor output")}},
						},
					},

					// foundry output - serialNumber: 1, minted: 200, melted: 0, max: 1000 [defaultAmount] (owned by account1AccountAddress)
					//				  - native token balance 100 (freshly minted)
					// => input 11
					&iotago.FoundryOutput{
						Amount:       defaultAmount,
						SerialNumber: 1,
						TokenScheme: &iotago.SimpleTokenScheme{
							MintedTokens:  new(big.Int).SetInt64(200),
							MeltedTokens:  big.NewInt(0),
							MaximumSupply: new(big.Int).SetInt64(1000),
						},
						UnlockConditions: iotago.FoundryOutputUnlockConditions{
							&iotago.ImmutableAccountUnlockCondition{Address: account1AccountAddress},
						},
						Features: iotago.FoundryOutputFeatures{
							&iotago.NativeTokenFeature{
								ID:     foundry1Ident3NativeTokenID,
								Amount: new(big.Int).SetUint64(100), // freshly minted
							},
						},
					},

					// foundry output - serialNumber: 2, minted: 100, melted: 50, max: 1000 [defaultAmount] (owned by account1AccountAddress)
					//				  - native token balance 50 (melted 50)
					// => input 12
					&iotago.FoundryOutput{
						Amount:       defaultAmount,
						SerialNumber: 2,
						TokenScheme: &iotago.SimpleTokenScheme{
							MintedTokens:  new(big.Int).SetInt64(100),
							MeltedTokens:  big.NewInt(50),
							MaximumSupply: new(big.Int).SetInt64(1000),
						},
						UnlockConditions: iotago.FoundryOutputUnlockConditions{
							&iotago.ImmutableAccountUnlockCondition{Address: account1AccountAddress},
						},
						Features: iotago.FoundryOutputFeatures{
							&iotago.NativeTokenFeature{
								ID:     foundry2Ident3NativeTokenID,
								Amount: new(big.Int).SetUint64(50), // melted to 50
							},
						},
					},

					// foundry output - serialNumber: 3, minted: 100, melted: 0, max: 1000 [defaultAmount] (owned by account1AccountAddress)
					// => input 13
					&iotago.FoundryOutput{
						Amount:       defaultAmount,
						SerialNumber: 3,
						TokenScheme: &iotago.SimpleTokenScheme{
							MintedTokens:  new(big.Int).SetInt64(100),
							MeltedTokens:  big.NewInt(0),
							MaximumSupply: new(big.Int).SetInt64(1000),
						},
						UnlockConditions: iotago.FoundryOutputUnlockConditions{
							&iotago.ImmutableAccountUnlockCondition{Address: account1AccountAddress},
						},
						Features: iotago.FoundryOutputFeatures{
							&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("interesting metadata")}},
						},
					},

					// foundry output - serialNumber: 6, minted: 100, melted: 0, max: 1000 [defaultAmount] (owned by account1AccountAddress)
					//				  - native token balance 100
					// => input 5
					newFoundryWithInitialSupply,

					// basic output [defaultAmount] (owned by ident3)
					// => input 14 (foundry 4 destruction remainder)
					&iotago.BasicOutput{
						Amount: defaultAmount,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident3},
						},
					},

					// NFT output transitioned and changed ownership [defaultAmount] (owned by ident4)
					// => input 15
					&iotago.NFTOutput{
						Amount: defaultAmount,
						NFTID:  nft1ID,
						UnlockConditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident4},
						},
						Features: iotago.NFTOutputFeatures{
							&iotago.IssuerFeature{Address: ident3},
						},
						ImmutableFeatures: iotago.NFTOutputImmFeatures{
							&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("transfer to 4")}},
						},
					},

					// new NFT output [defaultAmount] (owned by ident4)
					// => input 16
					&iotago.NFTOutput{
						Amount: defaultAmount,
						NFTID:  iotago.NFTID{},
						UnlockConditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident4},
						},
						Features: nil,
						ImmutableFeatures: iotago.NFTOutputImmFeatures{
							&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("immutable metadata")}},
						},
					},

					// basic output [defaultAmount] (owned by ident5)
					// => input 17
					&iotago.BasicOutput{
						Amount: defaultAmount,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident5},
						},
					},
				},
			}

			sigs, err := transaction.Sign(ident1AddrKeys, ident2AddrKeys, ident3AddrKeys, ident4AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet:        inputs,
					CommitmentInput: &iotago.Commitment{Slot: creationSlot},
				},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						// basic
						&iotago.SignatureUnlock{Signature: sigs[0]}, // basic output (owned by ident1) => (ident1 == Reference 0)
						&iotago.SignatureUnlock{Signature: sigs[1]}, // basic output (owned by ident2) => (ident2 == Reference 1)
						&iotago.ReferenceUnlock{Reference: 1},       // basic output (owned by ident2)
						&iotago.ReferenceUnlock{Reference: 0},       // basic output (owned by ident1)
						&iotago.ReferenceUnlock{Reference: 1},       // basic output (owned by ident2)
						&iotago.ReferenceUnlock{Reference: 1},       // basic output (owned by ident2)
						// account
						&iotago.SignatureUnlock{Signature: sigs[2]}, // account output (owned by ident3) => (ident3 == Reference 6)
						&iotago.ReferenceUnlock{Reference: 6},       // account output (owned by ident3)
						// anchor
						&iotago.SignatureUnlock{Signature: sigs[3]}, // anchor output (owned by state: ident3, gov: ident4) => governance transitioned => (ident4 == Reference 8)
						&iotago.ReferenceUnlock{Reference: 6},       // anchor output (owned by state: ident3, gov: ident4) => state transitioned
						&iotago.ReferenceUnlock{Reference: 6},       // anchor output (owned by state: ident3, gov: ident3) => governance transitioned
						// foundries
						&iotago.AccountUnlock{Reference: 6}, // foundry output (owned by account1AccountAddress)
						&iotago.AccountUnlock{Reference: 6}, // foundry output (owned by account1AccountAddress)
						&iotago.AccountUnlock{Reference: 6}, // foundry output (owned by account1AccountAddress)
						&iotago.AccountUnlock{Reference: 6}, // foundry output (owned by account1AccountAddress)
						// nfts
						&iotago.ReferenceUnlock{Reference: 6}, // NFT output (owned by ident3)
						&iotago.ReferenceUnlock{Reference: 8}, // NFT output (owned by ident4)
						&iotago.NFTUnlock{Reference: 15},      // basic output (owned by nft1ID)
					},
				},
				wantErr: nil,
			}
		}(),

		// fail - changed immutable account address unlock
		func() *test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(2)
			inFoundry := &iotago.FoundryOutput{
				Amount:       100,
				SerialNumber: 5,
				TokenScheme: &iotago.SimpleTokenScheme{
					MintedTokens:  new(big.Int).SetInt64(1000),
					MeltedTokens:  big.NewInt(0),
					MaximumSupply: new(big.Int).SetInt64(10000),
				},
				UnlockConditions: iotago.FoundryOutputUnlockConditions{
					&iotago.ImmutableAccountUnlockCondition{Address: accountAddr1},
				},
			}
			outFoundry := inFoundry.Clone().(*iotago.FoundryOutput)
			// change the immutable account address unlock
			outFoundry.UnlockConditions = iotago.FoundryOutputUnlockConditions{
				&iotago.ImmutableAccountUnlockCondition{Address: tpkg.RandAccountAddress()},
			}

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AccountOutput{
					Amount:    100,
					AccountID: accountAddr1.AccountID(),
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: inFoundry,
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs:       inputIDs.UTXOInputs(),
					Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.AccountOutput{
						Amount:    100,
						AccountID: accountAddr1.AccountID(),
						UnlockConditions: iotago.AccountOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					outFoundry,
				},
			}

			sigs, err := transaction.Sign(ident1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - changed immutable account address unlock",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						// should be an AccountUnlock
						&iotago.AccountUnlock{Reference: 0},
					},
				},
				// Changing the immutable account address unlock changes foundryID, therefore the chain is broken.
				// Next state of the foundry is empty, meaning it is interpreted as a destroy operation, and native tokens
				// are not balanced.
				wantErr: iotago.ErrNativeTokenSumUnbalanced,
			}
		}(),

		// ok - modify block issuer account
		func() *test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AccountOutput{
					Amount:    100,
					AccountID: accountAddr1.AccountID(),
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: iotago.NewBlockIssuerKeys(),
							ExpirySlot:      100,
						},
					},
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					CreationSlot: 110,
					ContextInputs: iotago.TxEssenceContextInputs{
						&iotago.BlockIssuanceCreditInput{
							AccountID: accountAddr1.AccountID(),
						},
					},
					Inputs:       inputIDs.UTXOInputs(),
					Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.AccountOutput{
						Amount:    100,
						AccountID: accountAddr1.AccountID(),
						Features: iotago.AccountOutputFeatures{
							&iotago.BlockIssuerFeature{
								BlockIssuerKeys: iotago.NewBlockIssuerKeys(),
								ExpirySlot:      1000,
							},
						},
						UnlockConditions: iotago.AccountOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
			}

			bicInputs := vm.BlockIssuanceCreditInputSet{
				accountAddr1.AccountID(): 0,
			}

			commitmentInput := &iotago.Commitment{
				Slot: 110,
			}

			sigs, err := transaction.Sign(ident1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "ok - modify block issuer account",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs, BlockIssuanceCreditInputSet: bicInputs, CommitmentInput: commitmentInput},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - set block issuer expiry to max value
		func() *test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AccountOutput{
					Amount:    100,
					AccountID: accountAddr1.AccountID(),
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: iotago.NewBlockIssuerKeys(),
							ExpirySlot:      100,
						},
					},
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					CreationSlot: 110,
					ContextInputs: iotago.TxEssenceContextInputs{
						&iotago.BlockIssuanceCreditInput{
							AccountID: accountAddr1.AccountID(),
						},
					},
					Inputs:       inputIDs.UTXOInputs(),
					Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.AccountOutput{
						Amount:    100,
						AccountID: accountAddr1.AccountID(),
						Features: iotago.AccountOutputFeatures{
							&iotago.BlockIssuerFeature{
								BlockIssuerKeys: iotago.NewBlockIssuerKeys(),
								ExpirySlot:      iotago.MaxSlotIndex,
							},
						},
						UnlockConditions: iotago.AccountOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
			}

			bicInputs := vm.BlockIssuanceCreditInputSet{
				accountAddr1.AccountID(): 0,
			}

			commitmentInput := &iotago.Commitment{
				Slot: 110,
			}

			sigs, err := transaction.Sign(ident1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "ok - set block issuer expiry to max value",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs, BlockIssuanceCreditInputSet: bicInputs, CommitmentInput: commitmentInput},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - remove expired block issuer feature from new account
		func() *test {
			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()

			creationSlot := iotago.SlotIndex(110)
			inputIDs := tpkg.RandOutputIDs(1)
			accountID := iotago.AccountIDFromOutputID(inputIDs[0])

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AccountOutput{
					Amount:    100,
					AccountID: iotago.EmptyAccountID,
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: iotago.NewBlockIssuerKeys(),
							ExpirySlot:      100,
						},
					},
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					CreationSlot: creationSlot,
					ContextInputs: iotago.TxEssenceContextInputs{
						&iotago.BlockIssuanceCreditInput{
							AccountID: accountID,
						},
					},
					Inputs:       inputIDs.UTXOInputs(),
					Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.AccountOutput{
						Amount:    100,
						AccountID: accountID,
						Features:  iotago.AccountOutputFeatures{},
						UnlockConditions: iotago.AccountOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
			}

			bicInputs := vm.BlockIssuanceCreditInputSet{
				accountID: 0,
			}

			commitmentInput := &iotago.Commitment{
				Slot: creationSlot,
			}

			sigs, err := transaction.Sign(ident1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "ok - remove expired block issuer feature from new account",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet:                    inputs,
					BlockIssuanceCreditInputSet: bicInputs,
					CommitmentInput:             commitmentInput,
				},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),

		// fail - destroy block issuer account with expiry at slot with max value
		func() *test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AccountOutput{
					Amount:    100,
					AccountID: accountAddr1.AccountID(),
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: iotago.NewBlockIssuerKeys(),
							ExpirySlot:      iotago.MaxSlotIndex,
						},
					},
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					CreationSlot: 110,
					ContextInputs: iotago.TxEssenceContextInputs{
						&iotago.BlockIssuanceCreditInput{
							AccountID: accountAddr1.AccountID(),
						},
					},
					Inputs:       inputIDs.UTXOInputs(),
					Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 100,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
			}

			bicInputs := vm.BlockIssuanceCreditInputSet{
				accountAddr1.AccountID(): 0,
			}

			commitmentInput := &iotago.Commitment{
				Slot: 110,
			}

			sigs, err := transaction.Sign(ident1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - destroy block issuer account with expiry at slot with max value",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs, BlockIssuanceCreditInputSet: bicInputs, CommitmentInput: commitmentInput},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},

				wantErr: iotago.ErrInvalidBlockIssuerTransition,
			}
		}(),

		// ok - destroy block issuer account
		func() *test {
			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(1)
			// Simulate the scenario where the input account's ID is unset.
			accountID := iotago.AccountIDFromOutputID(inputIDs[0])

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AccountOutput{
					Amount:    100,
					AccountID: iotago.EmptyAccountID,
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: iotago.NewBlockIssuerKeys(),
							ExpirySlot:      100,
						},
					},
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					CreationSlot: 110,
					ContextInputs: iotago.TxEssenceContextInputs{
						&iotago.BlockIssuanceCreditInput{
							AccountID: accountID,
						},
					},
					Inputs:       inputIDs.UTXOInputs(),
					Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 100,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
			}

			bicInputs := vm.BlockIssuanceCreditInputSet{
				accountID: 0,
			}

			commitment := &iotago.Commitment{
				Slot: 110,
			}

			sigs, err := transaction.Sign(ident1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "ok - destroy block issuer account",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs, BlockIssuanceCreditInputSet: bicInputs, CommitmentInput: commitment},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),

		// fail - destroy block issuer account without supplying BIC
		func() *test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AccountOutput{
					Amount:    100,
					AccountID: accountAddr1.AccountID(),
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: iotago.NewBlockIssuerKeys(),
							ExpirySlot:      100,
						},
					},
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					CreationSlot: 110,
					Inputs:       inputIDs.UTXOInputs(),
					Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 100,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
			}

			commitment := &iotago.Commitment{
				Slot: 110,
			}

			sigs, err := transaction.Sign(ident1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - destroy block issuer account without supplying BIC",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs, CommitmentInput: commitment},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrBlockIssuanceCreditInputRequired,
			}
		}(),

		// fail - modify block issuer without supplying BIC
		func() *test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AccountOutput{
					Amount:    100,
					AccountID: accountAddr1.AccountID(),
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: iotago.NewBlockIssuerKeys(),
							ExpirySlot:      100,
						},
					},
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					CreationSlot: 110,
					Inputs:       inputIDs.UTXOInputs(),
					Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.AccountOutput{
						Amount:    100,
						AccountID: accountAddr1.AccountID(),
						Features: iotago.AccountOutputFeatures{
							&iotago.BlockIssuerFeature{
								BlockIssuerKeys: iotago.NewBlockIssuerKeys(),
								ExpirySlot:      1000,
							},
						},
						UnlockConditions: iotago.AccountOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
			}

			sigs, err := transaction.Sign(ident1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - modify block issuer without supplying BIC",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrBlockIssuanceCreditInputRequired,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAndExecuteSignedTransaction(tt.tx, tt.resolvedInputs)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

type txBuilder struct {
	// the amount of randomly created ed25519 addresses with private keys
	ed25519AddrCnt int
	// used to created own addresses for the test
	addressesFunc func(ed25519Addresses []iotago.Address) []iotago.Address
	// used to create inputs for the test
	inputsFunc func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output
	// used to create outputs for the test (optional)
	outputsFunc func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address, totalInputAmount iotago.BaseToken, totalInputMana iotago.Mana) iotago.TxEssenceOutputs
	// used to create unlocks for the test
	unlocksFunc func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks
}

type txExecTest struct {
	// the name of the testcase
	name string
	// the txBuilder that builds the transaction for the testcase
	txBuilder *txBuilder
	// hook that gets executed before the transaction is signed (optional)
	txPreSignHook func(t *iotago.Transaction)
	// expected error during execution of the transaction
	wantErr error
}

func runNovaTransactionExecutionTest(t *testing.T, test *txExecTest) {
	t.Helper()

	t.Run(test.name, func(t *testing.T) {
		// generate random ed25519 addresses
		ed25519Addresses, ed25519AddressesWithKeys := tpkg.RandEd25519IdentitiesSortedByAddress(test.txBuilder.ed25519AddrCnt)

		// pass the ed25519 testAddresses and get the complete list of testAddresses
		testAddresses := make([]iotago.Address, 0)
		if test.txBuilder.addressesFunc != nil {
			testAddresses = test.txBuilder.addressesFunc(ed25519Addresses)
		}

		inputs := test.txBuilder.inputsFunc(ed25519Addresses, testAddresses)
		if len(inputs) == 0 {
			require.FailNow(t, "no outputs given")
		}

		// create the input set
		inputIDs := tpkg.RandOutputIDsWithCreationSlot(0, uint16(len(inputs)))
		inputSet := vm.InputSet{}
		var totalInputAmount iotago.BaseToken
		for idx, output := range inputs {
			inputSet[inputIDs[idx]] = output
			totalInputAmount += output.BaseTokenAmount()
		}

		// calculate the mana on input side
		// HINT: all outputs are created at slot 0 and the transaction is executed at slot 10000
		var txCreationSlot iotago.SlotIndex = 10000

		totalInputMana, err := vm.TotalManaIn(testAPI.ManaDecayProvider(), testAPI.StorageScoreStructure(), txCreationSlot, inputSet, vm.RewardsInputSet{})
		require.NoError(t, err)

		outputs := iotago.TxEssenceOutputs{
			// collect everything on a basic output with a random ed25519 address
			&iotago.BasicOutput{
				Amount: totalInputAmount,
				UnlockConditions: iotago.BasicOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
		}
		if test.txBuilder.outputsFunc != nil {
			outputs = test.txBuilder.outputsFunc(ed25519Addresses, testAddresses, totalInputAmount, totalInputMana)
		}

		// create the transaction
		tx := &iotago.Transaction{
			API: testAPI,
			TransactionEssence: &iotago.TransactionEssence{
				NetworkID:     testProtoParams.NetworkID(),
				CreationSlot:  txCreationSlot,
				ContextInputs: iotago.TxEssenceContextInputs{},
				Inputs:        inputIDs.UTXOInputs(),
				Allotments:    iotago.Allotments{},
				Capabilities:  iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
			},
			Outputs: outputs,
		}

		// execute the pre sign hook
		if test.txPreSignHook != nil {
			test.txPreSignHook(tx)
		}

		// sign the transaction essence
		sigs, err := tx.Sign(ed25519AddressesWithKeys...)
		require.NoError(t, err)

		// pass the signatures and get the unlock conditions
		unlocks := test.txBuilder.unlocksFunc(sigs, testAddresses)

		signedTx := &iotago.SignedTransaction{
			API:         testAPI,
			Transaction: tx,
			Unlocks:     unlocks,
		}

		txBytes, err := testAPI.Encode(signedTx, serix.WithValidation())
		require.NoError(t, err)

		// we deserialize to be sure that all serix rules are applied (like lexically ordering or multi addresses)
		signedTx = &iotago.SignedTransaction{}
		_, err = testAPI.Decode(txBytes, signedTx, serix.WithValidation())
		require.NoError(t, err)

		// execute the transaction
		err = validateAndExecuteSignedTransaction(signedTx, vm.ResolvedInputs{InputSet: inputSet})
		if test.wantErr != nil {
			require.ErrorIs(t, err, test.wantErr)
			return
		}
		require.NoError(t, err)
	})
}

func TestNovaTransactionExecution_RestrictedAddress(t *testing.T) {

	defaultAmount := OneIOTA

	tests := []*txExecTest{
		// ok - restricted ed25519 address unlock
		func() *txExecTest {
			return &txExecTest{
				name: "ok - restricted ed25519 address unlock",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 1,
					addressesFunc: func(ed25519Addresses []iotago.Address) []iotago.Address {
						return []iotago.Address{
							&iotago.RestrictedAddress{
								Address:             ed25519Addresses[0],
								AllowedCapabilities: iotago.AddressCapabilitiesBitMask{},
							},
						}
					},
					inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
						return []iotago.Output{
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
						}
					},
					outputsFunc: nil,
					unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
						return iotago.Unlocks{
							&iotago.SignatureUnlock{Signature: sigs[0]},
						}
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - restricted account address unlock
		func() *txExecTest {
			return &txExecTest{
				name: "ok - restricted account address unlock",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 1,
					addressesFunc: func(ed25519Addresses []iotago.Address) []iotago.Address {
						accountAddress := tpkg.RandAccountAddress()
						return []iotago.Address{
							accountAddress,
							&iotago.RestrictedAddress{
								Address:             accountAddress,
								AllowedCapabilities: iotago.AddressCapabilitiesBitMask{},
							},
						}
					},
					inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
						return []iotago.Output{
							// we add an output with a Ed25519 address to be able to check the AccountUnlock in the RestrictedAddress
							&iotago.AccountOutput{
								Amount:         defaultAmount,
								AccountID:      testAddresses[0].(*iotago.AccountAddress).AccountID(),
								FoundryCounter: 0,
								UnlockConditions: iotago.AccountOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
								Features: nil,
							},
							// owned by restricted account address
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[1]},
								},
							},
						}
					},
					outputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address, totalInputAmount iotago.BaseToken, totalInputMana iotago.Mana) iotago.TxEssenceOutputs {
						return iotago.TxEssenceOutputs{
							&iotago.AccountOutput{
								Amount:         defaultAmount,
								AccountID:      testAddresses[0].(*iotago.AccountAddress).AccountID(),
								FoundryCounter: 0,
								UnlockConditions: iotago.AccountOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
								Features: nil,
							},
							&iotago.BasicOutput{
								Amount: totalInputAmount - defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
							},
						}
					},
					unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
						return iotago.Unlocks{
							&iotago.SignatureUnlock{Signature: sigs[0]}, // account unlock
							&iotago.AccountUnlock{Reference: 0},
						}
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - restricted anchor address unlock
		func() *txExecTest {
			return &txExecTest{
				name: "ok - restricted anchor address unlock",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 2,
					addressesFunc: func(ed25519Addresses []iotago.Address) []iotago.Address {
						anchorAddress := tpkg.RandAnchorAddress()
						return []iotago.Address{
							anchorAddress,
							&iotago.RestrictedAddress{
								Address:             anchorAddress,
								AllowedCapabilities: iotago.AddressCapabilitiesBitMask{},
							},
						}
					},
					inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
						return []iotago.Output{
							// we add an output with a Ed25519 address to be able to check the AnchorUnlock in the RestrictedAddress
							&iotago.AnchorOutput{
								Amount:     defaultAmount,
								AnchorID:   testAddresses[0].(*iotago.AnchorAddress).AnchorID(),
								StateIndex: 1,
								UnlockConditions: iotago.AnchorOutputUnlockConditions{
									&iotago.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
									&iotago.GovernorAddressUnlockCondition{Address: ed25519Addresses[1]},
								},
								Features: iotago.AnchorOutputFeatures{
									&iotago.StateMetadataFeature{Entries: iotago.StateMetadataFeatureEntries{"data": []byte("current state")}},
								},
							},
							// owned by restricted anchor address
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[1]},
								},
							},
						}
					},
					outputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address, totalInputAmount iotago.BaseToken, totalInputMana iotago.Mana) iotago.TxEssenceOutputs {
						return iotago.TxEssenceOutputs{
							// the anchor unlock needs to be a state transition (governor doesn't work for anchor reference unlocks)
							&iotago.AnchorOutput{
								Amount:     defaultAmount,
								AnchorID:   testAddresses[0].(*iotago.AnchorAddress).AnchorID(),
								StateIndex: 2,
								UnlockConditions: iotago.AnchorOutputUnlockConditions{
									&iotago.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
									&iotago.GovernorAddressUnlockCondition{Address: ed25519Addresses[1]},
								},
								Features: iotago.AnchorOutputFeatures{
									&iotago.StateMetadataFeature{Entries: iotago.StateMetadataFeatureEntries{"data": []byte("next state")}},
								},
							},
							&iotago.BasicOutput{
								Amount: totalInputAmount - defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
							},
						}
					},
					unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
						return iotago.Unlocks{
							&iotago.SignatureUnlock{Signature: sigs[0]}, // anchor state controller unlock
							&iotago.AnchorUnlock{Reference: 0},
						}
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - restricted NFT unlock
		func() *txExecTest {
			return &txExecTest{
				name: "ok - restricted NFT unlock",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 2,
					addressesFunc: func(ed25519Addresses []iotago.Address) []iotago.Address {
						nftAddress := tpkg.RandNFTAddress()
						return []iotago.Address{
							nftAddress,
							&iotago.RestrictedAddress{
								Address:             nftAddress,
								AllowedCapabilities: iotago.AddressCapabilitiesBitMask{},
							},
						}
					},
					inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
						return []iotago.Output{
							// we add an output with a Ed25519 address to be able to check the NFT Unlock in the RestrictedAddress
							&iotago.NFTOutput{
								Amount: defaultAmount,
								NFTID:  testAddresses[0].(*iotago.NFTAddress).NFTID(),
								UnlockConditions: iotago.NFTOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
								Features: iotago.NFTOutputFeatures{
									&iotago.IssuerFeature{Address: ed25519Addresses[1]},
								},
								ImmutableFeatures: iotago.NFTOutputImmFeatures{
									&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("immutable")}},
								},
							},
							// owned by restricted NFT address
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[1]},
								},
							},
						}
					},
					outputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address, totalInputAmount iotago.BaseToken, totalInputMana iotago.Mana) iotago.TxEssenceOutputs {
						return iotago.TxEssenceOutputs{
							&iotago.NFTOutput{
								Amount: defaultAmount,
								NFTID:  testAddresses[0].(*iotago.NFTAddress).NFTID(),
								UnlockConditions: iotago.NFTOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
								Features: iotago.NFTOutputFeatures{
									&iotago.IssuerFeature{Address: ed25519Addresses[1]},
									&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("some new metadata")}},
								},
								ImmutableFeatures: iotago.NFTOutputImmFeatures{
									&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("immutable")}},
								},
							},
							&iotago.BasicOutput{
								Amount: totalInputAmount - defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
							},
						}
					},
					unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
						return iotago.Unlocks{
							&iotago.SignatureUnlock{Signature: sigs[0]}, // NFT unlock
							&iotago.NFTUnlock{Reference: 0},
						}
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - restricted multi address unlock
		func() *txExecTest {
			return &txExecTest{
				name: "ok - restricted multi address unlock",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 1,
					addressesFunc: func(ed25519Addresses []iotago.Address) []iotago.Address {
						nftAddress := tpkg.RandNFTAddress()
						return []iotago.Address{
							nftAddress,
							&iotago.RestrictedAddress{
								Address: &iotago.MultiAddress{
									Addresses: []*iotago.AddressWithWeight{
										{
											Address: ed25519Addresses[0],
											Weight:  1,
										},
										{
											Address: nftAddress,
											Weight:  1,
										},
									},
									Threshold: 2,
								},
								AllowedCapabilities: iotago.AddressCapabilitiesBitMask{},
							},
						}
					},
					inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
						return []iotago.Output{
							// we add an output with a Ed25519 address to be able to check the NFT Unlock in the RestrictedAddress multi address
							&iotago.NFTOutput{
								Amount: defaultAmount,
								NFTID:  testAddresses[0].(*iotago.NFTAddress).NFTID(),
								UnlockConditions: iotago.NFTOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
								Features: iotago.NFTOutputFeatures{
									&iotago.IssuerFeature{Address: ed25519Addresses[0]},
								},
								ImmutableFeatures: iotago.NFTOutputImmFeatures{
									&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("immutable")}},
								},
							},
							// owned by restricted multi address
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[1]},
								},
							},
						}
					},
					outputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address, totalInputAmount iotago.BaseToken, totalInputMana iotago.Mana) iotago.TxEssenceOutputs {
						return iotago.TxEssenceOutputs{
							&iotago.NFTOutput{
								Amount: defaultAmount,
								NFTID:  testAddresses[0].(*iotago.NFTAddress).NFTID(),
								UnlockConditions: iotago.NFTOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
								Features: iotago.NFTOutputFeatures{
									&iotago.IssuerFeature{Address: ed25519Addresses[0]},
									&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("some new metadata")}},
								},
								ImmutableFeatures: iotago.NFTOutputImmFeatures{
									&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("immutable")}},
								},
							},
							&iotago.BasicOutput{
								Amount: totalInputAmount - defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
							},
						}
					},
					unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
						return iotago.Unlocks{
							&iotago.SignatureUnlock{Signature: sigs[0]}, // NFT unlock
							&iotago.MultiUnlock{
								Unlocks: []iotago.Unlock{
									&iotago.ReferenceUnlock{Reference: 0},
									&iotago.NFTUnlock{Reference: 0},
								},
							},
						}
					},
				},
				wantErr: nil,
			}
		}(),
	}
	for _, tt := range tests {
		runNovaTransactionExecutionTest(t, tt)
	}
}

func TestNovaTransactionExecution_MultiAddress(t *testing.T) {

	defaultAmount := OneIOTA

	tests := []*txExecTest{
		// ok - threshold == cumulativeWeight (threshold reached)
		func() *txExecTest {
			return &txExecTest{
				name: "ok - threshold == cumulativeWeight (threshold reached)",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 2,
					addressesFunc: func(ed25519Addresses []iotago.Address) []iotago.Address {
						return []iotago.Address{
							// only 2 mandatory addresses
							&iotago.MultiAddress{
								Addresses: []*iotago.AddressWithWeight{
									{
										Address: ed25519Addresses[0],
										Weight:  1,
									},
									{
										Address: ed25519Addresses[1],
										Weight:  1,
									},
								},
								Threshold: 2,
							},
						}
					},
					inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
						return []iotago.Output{
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
						}
					},
					outputsFunc: nil,
					unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
						return iotago.Unlocks{
							&iotago.MultiUnlock{
								Unlocks: []iotago.Unlock{
									&iotago.SignatureUnlock{Signature: sigs[0]},
									&iotago.SignatureUnlock{Signature: sigs[1]},
								},
							},
						}
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - threshold < cumulativeWeight (threshold reached)
		func() *txExecTest {
			return &txExecTest{
				name: "ok - threshold < cumulativeWeight (threshold reached)",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 2,
					addressesFunc: func(ed25519Addresses []iotago.Address) []iotago.Address {
						return []iotago.Address{
							// only 2 mandatory addresses
							&iotago.MultiAddress{
								Addresses: []*iotago.AddressWithWeight{
									{
										Address: ed25519Addresses[0],
										Weight:  1,
									},
									{
										Address: ed25519Addresses[1],
										Weight:  1,
									},
								},
								Threshold: 1,
							},
						}
					},
					inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
						return []iotago.Output{
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
						}
					},
					outputsFunc: nil,
					unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
						return iotago.Unlocks{
							&iotago.MultiUnlock{
								Unlocks: []iotago.Unlock{
									&iotago.SignatureUnlock{Signature: sigs[0]},
									&iotago.SignatureUnlock{Signature: sigs[1]},
								},
							},
						}
					},
				},
				wantErr: nil,
			}
		}(),

		// fail - threshold == cumulativeWeight (threshold not reached)
		func() *txExecTest {
			return &txExecTest{
				name: "fail - threshold == cumulativeWeight (threshold not reached)",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 2,
					addressesFunc: func(ed25519Addresses []iotago.Address) []iotago.Address {
						return []iotago.Address{
							// only 2 mandatory addresses
							&iotago.MultiAddress{
								Addresses: []*iotago.AddressWithWeight{
									{
										Address: ed25519Addresses[0],
										Weight:  1,
									},
									{
										Address: ed25519Addresses[1],
										Weight:  1,
									},
								},
								Threshold: 2,
							},
						}
					},
					inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
						return []iotago.Output{
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
						}
					},
					outputsFunc: nil,
					unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
						return iotago.Unlocks{
							&iotago.MultiUnlock{
								Unlocks: []iotago.Unlock{
									// we only unlock one of the addresses
									&iotago.SignatureUnlock{Signature: sigs[0]},
									&iotago.EmptyUnlock{},
								},
							},
						}
					},
				},
				wantErr: iotago.ErrMultiAddressUnlockThresholdNotReached,
			}
		}(),

		// fail - threshold < cumulativeWeight (threshold not reached)
		func() *txExecTest {
			return &txExecTest{
				name: "fail - threshold < cumulativeWeight (threshold not reached)",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 2,
					addressesFunc: func(ed25519Addresses []iotago.Address) []iotago.Address {
						return []iotago.Address{
							// only 2 mandatory addresses
							&iotago.MultiAddress{
								Addresses: []*iotago.AddressWithWeight{
									{
										Address: ed25519Addresses[0],
										Weight:  2,
									},
									{
										Address: ed25519Addresses[1],
										Weight:  2,
									},
								},
								Threshold: 3,
							},
						}
					},
					inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
						return []iotago.Output{
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
						}
					},
					outputsFunc: nil,
					unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
						return iotago.Unlocks{
							&iotago.MultiUnlock{
								Unlocks: []iotago.Unlock{
									// we only unlock one of the addresses
									&iotago.SignatureUnlock{Signature: sigs[0]},
									&iotago.EmptyUnlock{},
								},
							},
						}
					},
				},
				wantErr: iotago.ErrMultiAddressUnlockThresholdNotReached,
			}
		}(),

		// fail - len(multiAddr) != len(multiUnlock)
		func() *txExecTest {
			return &txExecTest{
				name: "fail - len(multiAddr) != len(multiUnlock)",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 3,
					addressesFunc: func(ed25519Addresses []iotago.Address) []iotago.Address {
						return []iotago.Address{
							&iotago.MultiAddress{
								Addresses: []*iotago.AddressWithWeight{
									{
										Address: ed25519Addresses[0],
										Weight:  1,
									},
									{
										Address: ed25519Addresses[1],
										Weight:  1,
									},
									{
										Address: ed25519Addresses[2],
										Weight:  1,
									},
								},
								Threshold: 2,
							},
						}
					},
					inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
						return []iotago.Output{
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
						}
					},
					outputsFunc: nil,
					unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
						return iotago.Unlocks{
							&iotago.MultiUnlock{
								Unlocks: []iotago.Unlock{
									&iotago.SignatureUnlock{Signature: sigs[0]},
									&iotago.SignatureUnlock{Signature: sigs[1]},
									// Empty unlock missing here
								},
							},
						}
					},
				},
				wantErr: iotago.ErrMultiAddressAndUnlockLengthDoesNotMatch,
			}
		}(),

		// ok - Reference unlock
		func() *txExecTest {
			return &txExecTest{
				name: "ok - Reference unlock",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 2,
					addressesFunc: func(ed25519Addresses []iotago.Address) []iotago.Address {
						return []iotago.Address{
							// only 2 mandatory addresses
							&iotago.MultiAddress{
								Addresses: []*iotago.AddressWithWeight{
									{
										Address: ed25519Addresses[0],
										Weight:  1,
									},
									{
										Address: ed25519Addresses[1],
										Weight:  1,
									},
								},
								Threshold: 2,
							},
						}
					},
					inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
						return []iotago.Output{
							// we add a basic output with a Ed25519 address to be able to check the RefUnlock in the MultiAddress
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[1]},
								},
							},
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
						}
					},
					outputsFunc: nil,
					unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
						return iotago.Unlocks{
							&iotago.SignatureUnlock{Signature: sigs[1]},
							&iotago.MultiUnlock{
								Unlocks: []iotago.Unlock{
									&iotago.SignatureUnlock{Signature: sigs[0]},
									&iotago.ReferenceUnlock{Reference: 0},
								},
							},
						}
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - MultiAddress Reference unlock
		func() *txExecTest {
			return &txExecTest{
				name: "ok - MultiAddress Reference unlock",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 2,
					addressesFunc: func(ed25519Addresses []iotago.Address) []iotago.Address {
						return []iotago.Address{
							&iotago.MultiAddress{
								Addresses: []*iotago.AddressWithWeight{
									{
										Address: ed25519Addresses[0],
										Weight:  1,
									},
									{
										Address: ed25519Addresses[1],
										Weight:  1,
									},
								},
								Threshold: 2,
							},
						}
					},
					inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
						return []iotago.Output{
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
						}
					},
					outputsFunc: nil,
					unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
						return iotago.Unlocks{
							&iotago.MultiUnlock{
								Unlocks: []iotago.Unlock{
									&iotago.SignatureUnlock{Signature: sigs[0]},
									&iotago.SignatureUnlock{Signature: sigs[1]},
								},
							},
							&iotago.ReferenceUnlock{Reference: 0},
							&iotago.ReferenceUnlock{Reference: 0},
						}
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - Account unlock
		func() *txExecTest {
			return &txExecTest{
				name: "ok - Account unlock",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 2,
					addressesFunc: func(ed25519Addresses []iotago.Address) []iotago.Address {
						accountAddress := tpkg.RandAccountAddress()
						return []iotago.Address{
							accountAddress,
							// ed25519 address + account address
							&iotago.MultiAddress{
								Addresses: []*iotago.AddressWithWeight{
									{
										Address: ed25519Addresses[1],
										Weight:  1,
									},
									{
										Address: accountAddress,
										Weight:  1,
									},
								},
								Threshold: 2,
							},
						}
					},
					inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
						return []iotago.Output{
							// we add an output with a Ed25519 address to be able to check the AccountUnlock in the MultiAddress
							&iotago.AccountOutput{
								Amount:         defaultAmount,
								AccountID:      testAddresses[0].(*iotago.AccountAddress).AccountID(),
								FoundryCounter: 0,
								UnlockConditions: iotago.AccountOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
								Features: nil,
							},
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[1]},
								},
							},
							// owned by ed25519 address + account address
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[1]},
								},
							},
						}
					},
					outputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address, totalInputAmount iotago.BaseToken, totalInputMana iotago.Mana) iotago.TxEssenceOutputs {
						return iotago.TxEssenceOutputs{
							&iotago.AccountOutput{
								Amount:         defaultAmount,
								AccountID:      testAddresses[0].(*iotago.AccountAddress).AccountID(),
								FoundryCounter: 0,
								UnlockConditions: iotago.AccountOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
								Features: nil,
							},
							&iotago.BasicOutput{
								Amount: totalInputAmount - defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
							},
						}
					},
					unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
						// this is a bit complicated in the test, because the addresses are generated randomly,
						// but the MultiAddresses get sorted lexically, so we have to find out the correct order in the MultiUnlock.

						accountAddress := testAddresses[0]
						multiAddress := testAddresses[1].(*iotago.MultiAddress)

						// sort the addresses in the multi like the serializer will do
						slices.SortFunc(multiAddress.Addresses, func(a *iotago.AddressWithWeight, b *iotago.AddressWithWeight) int {
							return bytes.Compare(a.Address.ID(), b.Address.ID())
						})

						// search the index of the account address in the multi address
						foundAccountAddressIndex := -1
						for idx, address := range multiAddress.Addresses {
							if address.Address.Equal(accountAddress) {
								foundAccountAddressIndex = idx
								break
							}
						}

						var multiUnlock *iotago.MultiUnlock

						switch foundAccountAddressIndex {
						case -1:
							require.FailNow(t, "account address not found in multi address")

						case 0:
							multiUnlock = &iotago.MultiUnlock{
								Unlocks: []iotago.Unlock{
									&iotago.AccountUnlock{Reference: 0},
									&iotago.ReferenceUnlock{Reference: 1},
								},
							}

						case 1:
							multiUnlock = &iotago.MultiUnlock{
								Unlocks: []iotago.Unlock{
									&iotago.ReferenceUnlock{Reference: 1},
									&iotago.AccountUnlock{Reference: 0},
								},
							}

						default:
							require.FailNow(t, "unknown account address index found in multi address")
						}

						return iotago.Unlocks{
							&iotago.SignatureUnlock{Signature: sigs[0]}, // account unlock
							&iotago.SignatureUnlock{Signature: sigs[1]}, // basic output unlock
							multiUnlock,
						}
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - Anchor unlock (state transition)
		func() *txExecTest {
			return &txExecTest{
				name: "ok - Anchor unlock (state transition)",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 2,
					addressesFunc: func(ed25519Addresses []iotago.Address) []iotago.Address {
						anchorAddress := tpkg.RandAnchorAddress()
						return []iotago.Address{
							anchorAddress,
							// ed25519 address + anchor address
							&iotago.MultiAddress{
								Addresses: []*iotago.AddressWithWeight{
									{
										Address: ed25519Addresses[1],
										Weight:  1,
									},
									{
										Address: anchorAddress,
										Weight:  1,
									},
								},
								Threshold: 2,
							},
						}
					},
					inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
						return []iotago.Output{
							// we add an output with a Ed25519 address to be able to check the AnchorUnlock in the MultiAddress
							&iotago.AnchorOutput{
								Amount:     defaultAmount,
								AnchorID:   testAddresses[0].(*iotago.AnchorAddress).AnchorID(),
								StateIndex: 1,
								UnlockConditions: iotago.AnchorOutputUnlockConditions{
									&iotago.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
									&iotago.GovernorAddressUnlockCondition{Address: ed25519Addresses[1]},
								},
								Features: iotago.AnchorOutputFeatures{
									&iotago.StateMetadataFeature{Entries: iotago.StateMetadataFeatureEntries{"data": []byte("current state")}},
								},
							},
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[1]},
								},
							},
							// owned by ed25519 address + anchor address
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[1]},
								},
							},
						}
					},
					outputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address, totalInputAmount iotago.BaseToken, totalInputMana iotago.Mana) iotago.TxEssenceOutputs {
						return iotago.TxEssenceOutputs{
							// the anchor unlock needs to be a state transition (governor doesn't work for anchor reference unlocks)
							&iotago.AnchorOutput{
								Amount:     defaultAmount,
								AnchorID:   testAddresses[0].(*iotago.AnchorAddress).AnchorID(),
								StateIndex: 2,
								UnlockConditions: iotago.AnchorOutputUnlockConditions{
									&iotago.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
									&iotago.GovernorAddressUnlockCondition{Address: ed25519Addresses[1]},
								},
								Features: iotago.AnchorOutputFeatures{
									&iotago.StateMetadataFeature{Entries: iotago.StateMetadataFeatureEntries{"data": []byte("next state")}},
								},
							},
							&iotago.BasicOutput{
								Amount: totalInputAmount - defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
							},
						}
					},
					unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
						// this is a bit complicated in the test, because the addresses are generated randomly,
						// but the MultiAddresses get sorted lexically, so we have to find out the correct order in the MultiUnlock.

						anchorAddress := testAddresses[0]
						multiAddress := testAddresses[1].(*iotago.MultiAddress)

						// sort the addresses in the multi like the serializer will do
						slices.SortFunc(multiAddress.Addresses, func(a *iotago.AddressWithWeight, b *iotago.AddressWithWeight) int {
							return bytes.Compare(a.Address.ID(), b.Address.ID())
						})

						// search the index of the anchor address in the multi address
						foundAnchorAddressIndex := -1
						for idx, address := range multiAddress.Addresses {
							if address.Address.Equal(anchorAddress) {
								foundAnchorAddressIndex = idx
								break
							}
						}

						var multiUnlock *iotago.MultiUnlock

						switch foundAnchorAddressIndex {
						case -1:
							require.FailNow(t, "anchor address not found in multi address")

						case 0:
							multiUnlock = &iotago.MultiUnlock{
								Unlocks: []iotago.Unlock{
									&iotago.AnchorUnlock{Reference: 0},
									&iotago.ReferenceUnlock{Reference: 1},
								},
							}

						case 1:
							multiUnlock = &iotago.MultiUnlock{
								Unlocks: []iotago.Unlock{
									&iotago.ReferenceUnlock{Reference: 1},
									&iotago.AnchorUnlock{Reference: 0},
								},
							}

						default:
							require.FailNow(t, "unknown anchor address index found in multi address")
						}

						return iotago.Unlocks{
							&iotago.SignatureUnlock{Signature: sigs[0]}, // anchor state controller unlock
							&iotago.SignatureUnlock{Signature: sigs[1]}, // basic output unlock
							multiUnlock,
						}
					},
				},
				wantErr: nil,
			}
		}(),

		// fail - Anchor unlock (governance transition)
		func() *txExecTest {
			return &txExecTest{
				name: "fail - Anchor unlock (governance transition)",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 2,
					addressesFunc: func(ed25519Addresses []iotago.Address) []iotago.Address {
						anchorAddress := tpkg.RandAnchorAddress()
						return []iotago.Address{
							anchorAddress,
							// ed25519 address + anchor address
							&iotago.MultiAddress{
								Addresses: []*iotago.AddressWithWeight{
									{
										Address: ed25519Addresses[0],
										Weight:  1,
									},
									{
										Address: anchorAddress,
										Weight:  1,
									},
								},
								Threshold: 2,
							},
						}
					},
					inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
						return []iotago.Output{
							// we add an output with a Ed25519 address to be able to check the AnchorUnlock in the MultiAddress
							&iotago.AnchorOutput{
								Amount:     defaultAmount,
								AnchorID:   testAddresses[0].(*iotago.AnchorAddress).AnchorID(),
								StateIndex: 1,
								UnlockConditions: iotago.AnchorOutputUnlockConditions{
									&iotago.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
									&iotago.GovernorAddressUnlockCondition{Address: ed25519Addresses[1]},
								},
								Features: iotago.AnchorOutputFeatures{
									&iotago.StateMetadataFeature{Entries: iotago.StateMetadataFeatureEntries{"data": []byte("governance transition")}},
								},
							},
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
							},
							// owned by ed25519 address + anchor address
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[1]},
								},
							},
						}
					},
					outputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address, totalInputAmount iotago.BaseToken, totalInputMana iotago.Mana) iotago.TxEssenceOutputs {
						return iotago.TxEssenceOutputs{
							// the anchor unlock needs to be a state transition (governor doesn't work for anchor reference unlocks)
							&iotago.AnchorOutput{
								Amount:     defaultAmount,
								AnchorID:   testAddresses[0].(*iotago.AnchorAddress).AnchorID(),
								StateIndex: 1,
								UnlockConditions: iotago.AnchorOutputUnlockConditions{
									&iotago.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
									&iotago.GovernorAddressUnlockCondition{Address: ed25519Addresses[1]},
								},
								Features: iotago.AnchorOutputFeatures{
									&iotago.StateMetadataFeature{Entries: iotago.StateMetadataFeatureEntries{"data": []byte("governance transition")}},
								},
							},
							&iotago.BasicOutput{
								Amount: totalInputAmount - defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
							},
						}
					},
					unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
						// this is a bit complicated in the test, because the addresses are generated randomly,
						// but the MultiAddresses get sorted lexically, so we have to find out the correct order in the MultiUnlock.

						anchorAddress := testAddresses[0]
						multiAddress := testAddresses[1].(*iotago.MultiAddress)

						// sort the addresses in the multi like the serializer will do
						slices.SortFunc(multiAddress.Addresses, func(a *iotago.AddressWithWeight, b *iotago.AddressWithWeight) int {
							return bytes.Compare(a.Address.ID(), b.Address.ID())
						})

						// search the index of the anchor address in the multi address
						foundAnchorAddressIndex := -1
						for idx, address := range multiAddress.Addresses {
							if address.Address.Equal(anchorAddress) {
								foundAnchorAddressIndex = idx
								break
							}
						}

						var multiUnlock *iotago.MultiUnlock

						switch foundAnchorAddressIndex {
						case -1:
							require.FailNow(t, "anchor address not found in multi address")

						case 0:
							multiUnlock = &iotago.MultiUnlock{
								Unlocks: []iotago.Unlock{
									&iotago.AnchorUnlock{Reference: 0},
									&iotago.ReferenceUnlock{Reference: 1},
								},
							}

						case 1:
							multiUnlock = &iotago.MultiUnlock{
								Unlocks: []iotago.Unlock{
									&iotago.ReferenceUnlock{Reference: 1},
									&iotago.AnchorUnlock{Reference: 0},
								},
							}

						default:
							require.FailNow(t, "unknown anchor address index found in multi address")
						}

						return iotago.Unlocks{
							&iotago.SignatureUnlock{Signature: sigs[1]}, // anchor governor unlock
							&iotago.SignatureUnlock{Signature: sigs[0]}, // basic output unlock
							multiUnlock,
						}
					},
				},
				wantErr: iotago.ErrInvalidInputUnlock,
			}
		}(),

		// ok - NFT unlock
		func() *txExecTest {
			return &txExecTest{
				name: "ok - NFT unlock",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 2,
					addressesFunc: func(ed25519Addresses []iotago.Address) []iotago.Address {
						nftAddress := tpkg.RandNFTAddress()
						return []iotago.Address{
							nftAddress,
							// ed25519 address + NFT address
							&iotago.MultiAddress{
								Addresses: []*iotago.AddressWithWeight{
									{
										Address: ed25519Addresses[1],
										Weight:  1,
									},
									{
										Address: nftAddress,
										Weight:  1,
									},
								},
								Threshold: 2,
							},
						}
					},
					inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
						return []iotago.Output{
							// we add an output with a Ed25519 address to be able to check the NFT Unlock in the MultiAddress
							&iotago.NFTOutput{
								Amount: defaultAmount,
								NFTID:  testAddresses[0].(*iotago.NFTAddress).NFTID(),
								UnlockConditions: iotago.NFTOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
								Features: iotago.NFTOutputFeatures{
									&iotago.IssuerFeature{Address: ed25519Addresses[1]},
								},
								ImmutableFeatures: iotago.NFTOutputImmFeatures{
									&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("immutable")}},
								},
							},
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[1]},
								},
							},
							// owned by ed25519 address + NFT address
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[1]},
								},
							},
						}
					},
					outputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address, totalInputAmount iotago.BaseToken, totalInputMana iotago.Mana) iotago.TxEssenceOutputs {
						return iotago.TxEssenceOutputs{
							&iotago.NFTOutput{
								Amount: defaultAmount,
								NFTID:  testAddresses[0].(*iotago.NFTAddress).NFTID(),
								UnlockConditions: iotago.NFTOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
								Features: iotago.NFTOutputFeatures{
									&iotago.IssuerFeature{Address: ed25519Addresses[1]},
									&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("some new metadata")}},
								},
								ImmutableFeatures: iotago.NFTOutputImmFeatures{
									&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("immutable")}},
								},
							},
							&iotago.BasicOutput{
								Amount: totalInputAmount - defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
							},
						}
					},
					unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
						// this is a bit complicated in the test, because the addresses are generated randomly,
						// but the MultiAddresses get sorted lexically, so we have to find out the correct order in the MultiUnlock.

						nftAddress := testAddresses[0]
						multiAddress := testAddresses[1].(*iotago.MultiAddress)

						// sort the addresses in the multi like the serializer will do
						slices.SortFunc(multiAddress.Addresses, func(a *iotago.AddressWithWeight, b *iotago.AddressWithWeight) int {
							return bytes.Compare(a.Address.ID(), b.Address.ID())
						})

						// search the index of the NFT address in the multi address
						foundNFTAddressIndex := -1
						for idx, address := range multiAddress.Addresses {
							if address.Address.Equal(nftAddress) {
								foundNFTAddressIndex = idx
								break
							}
						}

						var multiUnlock *iotago.MultiUnlock

						switch foundNFTAddressIndex {
						case -1:
							require.FailNow(t, "NFT address not found in multi address")

						case 0:
							multiUnlock = &iotago.MultiUnlock{
								Unlocks: []iotago.Unlock{
									&iotago.NFTUnlock{Reference: 0},
									&iotago.ReferenceUnlock{Reference: 1},
								},
							}

						case 1:
							multiUnlock = &iotago.MultiUnlock{
								Unlocks: []iotago.Unlock{
									&iotago.ReferenceUnlock{Reference: 1},
									&iotago.NFTUnlock{Reference: 0},
								},
							}

						default:
							require.FailNow(t, "unknown NFT address index found in multi address")
						}

						return iotago.Unlocks{
							&iotago.SignatureUnlock{Signature: sigs[0]}, // NFT unlock
							&iotago.SignatureUnlock{Signature: sigs[1]}, // basic output unlock
							multiUnlock,
						}
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - multiple MultiAddresses in one TX - no signature reuse
		func() *txExecTest {
			return &txExecTest{
				name: "ok - multiple MultiAddresses in one TX - no signature reuse",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 4,
					addressesFunc: func(ed25519Addresses []iotago.Address) []iotago.Address {
						return []iotago.Address{
							&iotago.MultiAddress{
								Addresses: []*iotago.AddressWithWeight{
									{
										Address: ed25519Addresses[0],
										Weight:  1,
									},
									{
										Address: ed25519Addresses[1],
										Weight:  1,
									},
								},
								Threshold: 2,
							},
							&iotago.MultiAddress{
								Addresses: []*iotago.AddressWithWeight{
									{
										// optional
										Address: ed25519Addresses[0],
										Weight:  1,
									},
									{
										// optional
										Address: ed25519Addresses[2],
										Weight:  1,
									},
									{
										// mandatory
										Address: ed25519Addresses[3],
										Weight:  2,
									},
								},
								Threshold: 3,
							},
						}
					},
					inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
						return []iotago.Output{
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[1]},
								},
							},
						}
					},
					outputsFunc: nil,
					unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
						return iotago.Unlocks{
							&iotago.MultiUnlock{
								Unlocks: []iotago.Unlock{
									&iotago.SignatureUnlock{Signature: sigs[0]},
									&iotago.SignatureUnlock{Signature: sigs[1]},
								},
							},
							&iotago.MultiUnlock{
								Unlocks: []iotago.Unlock{
									&iotago.EmptyUnlock{},
									&iotago.SignatureUnlock{Signature: sigs[2]},
									&iotago.SignatureUnlock{Signature: sigs[3]},
								},
							},
						}
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - multiple MultiAddresses in one TX - signature reuse in different multi unlocks
		func() *txExecTest {
			return &txExecTest{
				name: "ok - multiple MultiAddresses in one TX - signature reuse in different multi unlocks",
				txBuilder: &txBuilder{
					ed25519AddrCnt: 4,
					addressesFunc: func(ed25519Addresses []iotago.Address) []iotago.Address {
						return []iotago.Address{
							&iotago.MultiAddress{
								Addresses: []*iotago.AddressWithWeight{
									{
										Address: ed25519Addresses[0],
										Weight:  1,
									},
									{
										Address: ed25519Addresses[1],
										Weight:  1,
									},
								},
								Threshold: 2,
							},
							&iotago.MultiAddress{
								Addresses: []*iotago.AddressWithWeight{
									{
										// optional
										Address: ed25519Addresses[0],
										Weight:  1,
									},
									{
										// optional
										Address: ed25519Addresses[2],
										Weight:  1,
									},
									{
										// mandatory
										Address: ed25519Addresses[3],
										Weight:  2,
									},
								},
								Threshold: 3,
							},
						}
					},
					inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
						return []iotago.Output{
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
							&iotago.BasicOutput{
								Amount: defaultAmount,
								UnlockConditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[1]},
								},
							},
						}
					},
					outputsFunc: nil,
					unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
						return iotago.Unlocks{
							&iotago.MultiUnlock{
								Unlocks: []iotago.Unlock{
									&iotago.SignatureUnlock{Signature: sigs[0]},
									&iotago.SignatureUnlock{Signature: sigs[1]},
								},
							},
							&iotago.MultiUnlock{
								Unlocks: []iotago.Unlock{
									&iotago.SignatureUnlock{Signature: sigs[0]},
									&iotago.EmptyUnlock{},
									&iotago.SignatureUnlock{Signature: sigs[3]},
								},
							},
						}
					},
				},
				wantErr: nil,
			}
		}(),
	}
	for _, tt := range tests {
		runNovaTransactionExecutionTest(t, tt)
	}
}

func TestNovaTransactionExecution_TxCapabilities(t *testing.T) {

	defaultAmount := OneIOTA

	// builds a transaction that burns native tokens
	burnNativeTokenTxBuilder := &txBuilder{
		ed25519AddrCnt: 1,
		inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
			return []iotago.Output{
				&iotago.BasicOutput{
					Amount: defaultAmount,
					// add native tokens
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
					Features: iotago.BasicOutputFeatures{
						tpkg.RandNativeTokenFeature(),
					},
				},
			}
		},
		outputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address, totalInputAmount iotago.BaseToken, totalInputMana iotago.Mana) iotago.TxEssenceOutputs {
			return iotago.TxEssenceOutputs{
				&iotago.BasicOutput{
					Amount: totalInputAmount,
					Mana:   totalInputMana,
					// burn the native tokens
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
				},
			}
		},
		unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
			return iotago.Unlocks{
				&iotago.SignatureUnlock{Signature: sigs[0]},
			}
		},
	}

	// builds a transaction that melts native tokens
	meltNativeTokenTxBuilder := &txBuilder{
		ed25519AddrCnt: 1,
		addressesFunc: func(ed25519Addresses []iotago.Address) []iotago.Address {
			accountAddress := tpkg.RandAccountAddress()
			return []iotago.Address{
				accountAddress,
			}
		},
		inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
			foundryID, err := iotago.FoundryIDFromAddressAndSerialNumberAndTokenScheme(testAddresses[0], 1, iotago.TokenSchemeSimple)
			require.NoError(t, err)

			return []iotago.Output{
				&iotago.AccountOutput{
					Amount:         defaultAmount,
					Mana:           0,
					AccountID:      testAddresses[0].(*iotago.AccountAddress).AccountID(),
					FoundryCounter: 1,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
					Features:          iotago.AccountOutputFeatures{},
					ImmutableFeatures: iotago.AccountOutputImmFeatures{},
				},
				&iotago.FoundryOutput{
					Amount:       defaultAmount,
					SerialNumber: 1,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  big.NewInt(100),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: big.NewInt(100),
					},
					UnlockConditions: iotago.FoundryOutputUnlockConditions{
						&iotago.ImmutableAccountUnlockCondition{
							Address: testAddresses[0].(*iotago.AccountAddress),
						},
					},
					Features:          iotago.FoundryOutputFeatures{},
					ImmutableFeatures: iotago.FoundryOutputImmFeatures{},
				},
				&iotago.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
					Features: iotago.BasicOutputFeatures{
						&iotago.NativeTokenFeature{
							ID:     foundryID,
							Amount: big.NewInt(100),
						},
					},
				},
			}
		},
		outputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address, totalInputAmount iotago.BaseToken, totalInputMana iotago.Mana) iotago.TxEssenceOutputs {
			foundryID, err := iotago.FoundryIDFromAddressAndSerialNumberAndTokenScheme(testAddresses[0], 1, iotago.TokenSchemeSimple)
			require.NoError(t, err)

			return iotago.TxEssenceOutputs{
				&iotago.AccountOutput{
					Amount:         totalInputAmount - defaultAmount,
					Mana:           totalInputMana,
					AccountID:      testAddresses[0].(*iotago.AccountAddress).AccountID(),
					FoundryCounter: 1,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
					Features:          iotago.AccountOutputFeatures{},
					ImmutableFeatures: iotago.AccountOutputImmFeatures{},
				},
				&iotago.FoundryOutput{
					Amount:       defaultAmount,
					SerialNumber: 1,
					TokenScheme: &iotago.SimpleTokenScheme{
						// melt the native tokens
						MintedTokens:  big.NewInt(100),
						MeltedTokens:  big.NewInt(50),
						MaximumSupply: big.NewInt(100),
					},
					UnlockConditions: iotago.FoundryOutputUnlockConditions{
						&iotago.ImmutableAccountUnlockCondition{
							Address: testAddresses[0].(*iotago.AccountAddress),
						},
					},
					Features: iotago.FoundryOutputFeatures{
						&iotago.NativeTokenFeature{
							ID:     foundryID,
							Amount: big.NewInt(50),
						},
					},
					ImmutableFeatures: iotago.FoundryOutputImmFeatures{},
				},
			}
		},
		unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
			return iotago.Unlocks{
				&iotago.SignatureUnlock{Signature: sigs[0]},
				&iotago.AccountUnlock{Reference: 0},
				&iotago.ReferenceUnlock{Reference: 0},
			}
		},
	}

	// builds a transaction that burns and melts native tokens
	burnAndMeltNativeTokenTxBuilder := &txBuilder{
		ed25519AddrCnt: 1,
		addressesFunc: func(ed25519Addresses []iotago.Address) []iotago.Address {
			accountAddress := tpkg.RandAccountAddress()
			return []iotago.Address{
				accountAddress,
			}
		},
		inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
			foundryID, err := iotago.FoundryIDFromAddressAndSerialNumberAndTokenScheme(testAddresses[0], 1, iotago.TokenSchemeSimple)
			require.NoError(t, err)

			return []iotago.Output{
				&iotago.AccountOutput{
					Amount:         defaultAmount,
					Mana:           0,
					AccountID:      testAddresses[0].(*iotago.AccountAddress).AccountID(),
					FoundryCounter: 1,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
					Features:          iotago.AccountOutputFeatures{},
					ImmutableFeatures: iotago.AccountOutputImmFeatures{},
				},
				&iotago.FoundryOutput{
					Amount:       defaultAmount,
					SerialNumber: 1,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  big.NewInt(100),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: big.NewInt(100),
					},
					UnlockConditions: iotago.FoundryOutputUnlockConditions{
						&iotago.ImmutableAccountUnlockCondition{
							Address: testAddresses[0].(*iotago.AccountAddress),
						},
					},
					Features:          iotago.FoundryOutputFeatures{},
					ImmutableFeatures: iotago.FoundryOutputImmFeatures{},
				},
				&iotago.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
					Features: iotago.BasicOutputFeatures{
						&iotago.NativeTokenFeature{
							ID:     foundryID,
							Amount: big.NewInt(100),
						},
					},
				},
			}
		},
		outputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address, totalInputAmount iotago.BaseToken, totalInputMana iotago.Mana) iotago.TxEssenceOutputs {
			return iotago.TxEssenceOutputs{
				&iotago.AccountOutput{
					Amount:         totalInputAmount - defaultAmount,
					Mana:           totalInputMana,
					AccountID:      testAddresses[0].(*iotago.AccountAddress).AccountID(),
					FoundryCounter: 1,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
					Features:          iotago.AccountOutputFeatures{},
					ImmutableFeatures: iotago.AccountOutputImmFeatures{},
				},
				&iotago.FoundryOutput{
					Amount:       defaultAmount,
					SerialNumber: 1,
					TokenScheme: &iotago.SimpleTokenScheme{
						// melt the native tokens
						MintedTokens:  big.NewInt(100),
						MeltedTokens:  big.NewInt(50),
						MaximumSupply: big.NewInt(100),
					},
					UnlockConditions: iotago.FoundryOutputUnlockConditions{
						&iotago.ImmutableAccountUnlockCondition{
							Address: testAddresses[0].(*iotago.AccountAddress),
						},
					},
					Features:          iotago.FoundryOutputFeatures{},
					ImmutableFeatures: iotago.FoundryOutputImmFeatures{},
				},
			}
		},
		unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
			return iotago.Unlocks{
				&iotago.SignatureUnlock{Signature: sigs[0]},
				&iotago.AccountUnlock{Reference: 0},
				&iotago.ReferenceUnlock{Reference: 0},
			}
		},
	}

	// builds a transaction that burns mana
	burnManaTxBuilder := &txBuilder{
		ed25519AddrCnt: 1,
		inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
			return []iotago.Output{
				&iotago.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
				},
			}
		},
		outputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address, totalInputAmount iotago.BaseToken, totalInputMana iotago.Mana) iotago.TxEssenceOutputs {
			return iotago.TxEssenceOutputs{
				&iotago.BasicOutput{
					Amount: totalInputAmount,
					// burn mana
					Mana: totalInputMana - 10,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
				},
			}
		},
		unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
			return iotago.Unlocks{
				&iotago.SignatureUnlock{Signature: sigs[0]},
			}
		},
	}

	// builds a transaction that destroys an account
	destroyAccountTxBuilder := &txBuilder{
		ed25519AddrCnt: 1,
		inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
			return []iotago.Output{
				&iotago.AccountOutput{
					Amount: defaultAmount,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
				},
			}
		},
		outputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address, totalInputAmount iotago.BaseToken, totalInputMana iotago.Mana) iotago.TxEssenceOutputs {
			return iotago.TxEssenceOutputs{
				// destroy the account output
				&iotago.BasicOutput{
					Amount: totalInputAmount,
					Mana:   totalInputMana,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
				},
			}
		},
		unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
			return iotago.Unlocks{
				&iotago.SignatureUnlock{Signature: sigs[0]},
			}
		},
	}

	// builds a transaction that destroys an anchor
	destroyAnchorTxBuilder := &txBuilder{
		ed25519AddrCnt: 1,
		inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
			return []iotago.Output{
				&iotago.AnchorOutput{
					Amount: defaultAmount,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.GovernorAddressUnlockCondition{Address: ed25519Addresses[0]},
						&iotago.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
					},
				},
			}
		},
		outputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address, totalInputAmount iotago.BaseToken, totalInputMana iotago.Mana) iotago.TxEssenceOutputs {
			return iotago.TxEssenceOutputs{
				// destroy the anchor output
				&iotago.BasicOutput{
					Amount: totalInputAmount,
					Mana:   totalInputMana,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
				},
			}
		},
		unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
			return iotago.Unlocks{
				&iotago.SignatureUnlock{Signature: sigs[0]},
			}
		},
	}

	// builds a transaction that destroys a foundry
	destroyFoundryTxBuilder := &txBuilder{
		ed25519AddrCnt: 1,
		addressesFunc: func(ed25519Addresses []iotago.Address) []iotago.Address {
			accountAddress := tpkg.RandAccountAddress()
			return []iotago.Address{
				accountAddress,
			}
		},
		inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
			return []iotago.Output{
				&iotago.AccountOutput{
					Amount:         defaultAmount,
					Mana:           0,
					AccountID:      testAddresses[0].(*iotago.AccountAddress).AccountID(),
					FoundryCounter: 1,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
					Features:          iotago.AccountOutputFeatures{},
					ImmutableFeatures: iotago.AccountOutputImmFeatures{},
				},
				&iotago.FoundryOutput{
					Amount:       defaultAmount,
					SerialNumber: 1,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  big.NewInt(100),
						MeltedTokens:  big.NewInt(100),
						MaximumSupply: big.NewInt(100),
					},
					UnlockConditions: iotago.FoundryOutputUnlockConditions{
						&iotago.ImmutableAccountUnlockCondition{
							Address: testAddresses[0].(*iotago.AccountAddress),
						},
					},
					Features:          iotago.FoundryOutputFeatures{},
					ImmutableFeatures: iotago.FoundryOutputImmFeatures{},
				},
			}
		},
		outputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address, totalInputAmount iotago.BaseToken, totalInputMana iotago.Mana) iotago.TxEssenceOutputs {
			return iotago.TxEssenceOutputs{
				// destroy the foundry output
				&iotago.AccountOutput{
					Amount:         totalInputAmount,
					Mana:           totalInputMana,
					AccountID:      testAddresses[0].(*iotago.AccountAddress).AccountID(),
					FoundryCounter: 1,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
					Features:          iotago.AccountOutputFeatures{},
					ImmutableFeatures: iotago.AccountOutputImmFeatures{},
				},
			}
		},
		unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
			return iotago.Unlocks{
				&iotago.SignatureUnlock{Signature: sigs[0]},
				&iotago.AccountUnlock{Reference: 0},
			}
		},
	}

	// builds a transaction that destroys a NFT
	destroyNFTTxBuilder := &txBuilder{
		ed25519AddrCnt: 1,
		inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
			return []iotago.Output{
				&iotago.NFTOutput{
					Amount: defaultAmount,
					UnlockConditions: iotago.NFTOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
				},
			}
		},
		outputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address, totalInputAmount iotago.BaseToken, totalInputMana iotago.Mana) iotago.TxEssenceOutputs {
			return iotago.TxEssenceOutputs{
				// destroy the NFT output
				&iotago.BasicOutput{
					Amount: totalInputAmount,
					Mana:   totalInputMana,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
					},
				},
			}
		},
		unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
			return iotago.Unlocks{
				&iotago.SignatureUnlock{Signature: sigs[0]},
			}
		},
	}

	tests := []*txExecTest{
		// ok - burn native tokens (burning enabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "ok - burn native tokens (burning enabled)",
				txBuilder: burnNativeTokenTxBuilder,
				txPreSignHook: func(t *iotago.Transaction) {
					t.Capabilities = iotago.TransactionCapabilitiesBitMaskWithCapabilities(
						iotago.WithTransactionCanBurnNativeTokens(true),
					)
				},
				wantErr: nil,
			}
		}(),

		// fail - burn native tokens (burning disabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "fail - burn native tokens (burning disabled)",
				txBuilder: burnNativeTokenTxBuilder,
				txPreSignHook: func(t *iotago.Transaction) {
					t.Capabilities = iotago.TransactionCapabilitiesBitMaskWithCapabilities(
						iotago.WithTransactionCanDoAnything(),
						iotago.WithTransactionCanBurnNativeTokens(false),
					)
				},
				wantErr: iotago.ErrTxCapabilitiesNativeTokenBurningNotAllowed,
			}
		}(),

		// ok - melt native tokens (burning enabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "ok - melt native tokens (burning enabled)",
				txBuilder: meltNativeTokenTxBuilder,
				txPreSignHook: func(t *iotago.Transaction) {
					t.Capabilities = iotago.TransactionCapabilitiesBitMaskWithCapabilities(
						iotago.WithTransactionCanBurnNativeTokens(true),
					)
				},
				wantErr: nil,
			}
		}(),

		// ok - melt native tokens (burning disabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "ok - melt native tokens (burning disabled)",
				txBuilder: meltNativeTokenTxBuilder,
				txPreSignHook: func(t *iotago.Transaction) {
					t.Capabilities = iotago.TransactionCapabilitiesBitMaskWithCapabilities(
						iotago.WithTransactionCanDoAnything(),
						iotago.WithTransactionCanBurnNativeTokens(false),
					)
				},
				wantErr: nil,
			}
		}(),

		// ok - burn and melt native tokens (burning enabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "fail - burn and melt native tokens (burning enabled)",
				txBuilder: burnAndMeltNativeTokenTxBuilder,
				txPreSignHook: func(t *iotago.Transaction) {
					t.Capabilities = iotago.TransactionCapabilitiesBitMaskWithCapabilities(
						iotago.WithTransactionCanBurnNativeTokens(true),
					)
				},
				wantErr: iotago.ErrNativeTokenSumUnbalanced,
			}
		}(),

		// fail - burn and melt native tokens (burning disabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "fail - burn and melt native tokens (burning disabled)",
				txBuilder: burnAndMeltNativeTokenTxBuilder,
				txPreSignHook: func(t *iotago.Transaction) {
					t.Capabilities = iotago.TransactionCapabilitiesBitMaskWithCapabilities(
						iotago.WithTransactionCanDoAnything(),
						iotago.WithTransactionCanBurnNativeTokens(false),
					)
				},
				wantErr: iotago.ErrNativeTokenSumUnbalanced,
			}
		}(),

		// ok - burn mana (burning enabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "ok - burn mana (burning enabled)",
				txBuilder: burnManaTxBuilder,
				txPreSignHook: func(t *iotago.Transaction) {
					t.Capabilities = iotago.TransactionCapabilitiesBitMaskWithCapabilities(
						iotago.WithTransactionCanBurnMana(true),
					)
				},
				wantErr: nil,
			}
		}(),

		// fail - burn mana (burning disabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "fail - burn mana (burning disabled)",
				txBuilder: burnManaTxBuilder,
				txPreSignHook: func(t *iotago.Transaction) {
					t.Capabilities = iotago.TransactionCapabilitiesBitMaskWithCapabilities(
						iotago.WithTransactionCanDoAnything(),
						iotago.WithTransactionCanBurnMana(false),
					)
				},
				wantErr: iotago.ErrTxCapabilitiesManaBurningNotAllowed,
			}
		}(),

		// ok - destroy account (destruction enabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "ok - destroy account (destruction enabled)",
				txBuilder: destroyAccountTxBuilder,
				txPreSignHook: func(t *iotago.Transaction) {
					t.Capabilities = iotago.TransactionCapabilitiesBitMaskWithCapabilities(
						iotago.WithTransactionCanDestroyAccountOutputs(true),
					)
				},
				wantErr: nil,
			}
		}(),

		// fail - destroy account (destruction disabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "fail - destroy account (destruction disabled)",
				txBuilder: destroyAccountTxBuilder,
				txPreSignHook: func(t *iotago.Transaction) {
					t.Capabilities = iotago.TransactionCapabilitiesBitMaskWithCapabilities(
						iotago.WithTransactionCanDoAnything(),
						iotago.WithTransactionCanDestroyAccountOutputs(false),
					)
				},
				wantErr: iotago.ErrTxCapabilitiesAccountDestructionNotAllowed,
			}
		}(),

		// ok - destroy anchor (destruction enabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "ok - destroy anchor (destruction enabled)",
				txBuilder: destroyAnchorTxBuilder,
				txPreSignHook: func(t *iotago.Transaction) {
					t.Capabilities = iotago.TransactionCapabilitiesBitMaskWithCapabilities(
						iotago.WithTransactionCanDestroyAnchorOutputs(true),
					)
				},
				wantErr: nil,
			}
		}(),

		// fail - destroy anchor (destruction disabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "fail - destroy anchor (destruction disabled)",
				txBuilder: destroyAnchorTxBuilder,
				txPreSignHook: func(t *iotago.Transaction) {
					t.Capabilities = iotago.TransactionCapabilitiesBitMaskWithCapabilities(
						iotago.WithTransactionCanDoAnything(),
						iotago.WithTransactionCanDestroyAnchorOutputs(false),
					)
				},
				wantErr: iotago.ErrTxCapabilitiesAnchorDestructionNotAllowed,
			}
		}(),

		// ok - destroy foundry (destruction enabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "ok - destroy foundry (destruction enabled)",
				txBuilder: destroyFoundryTxBuilder,
				txPreSignHook: func(t *iotago.Transaction) {
					t.Capabilities = iotago.TransactionCapabilitiesBitMaskWithCapabilities(
						iotago.WithTransactionCanDestroyFoundryOutputs(true),
					)
				},
				wantErr: nil,
			}
		}(),

		// fail - destroy foundry (destruction disabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "fail - destroy foundry (destruction disabled)",
				txBuilder: destroyFoundryTxBuilder,
				txPreSignHook: func(t *iotago.Transaction) {
					t.Capabilities = iotago.TransactionCapabilitiesBitMaskWithCapabilities(
						iotago.WithTransactionCanDoAnything(),
						iotago.WithTransactionCanDestroyFoundryOutputs(false),
					)
				},
				wantErr: iotago.ErrTxCapabilitiesFoundryDestructionNotAllowed,
			}
		}(),

		// ok - destroy NFT (destruction enabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "ok - destroy NFT (destruction enabled)",
				txBuilder: destroyNFTTxBuilder,
				txPreSignHook: func(t *iotago.Transaction) {
					t.Capabilities = iotago.TransactionCapabilitiesBitMaskWithCapabilities(
						iotago.WithTransactionCanDestroyNFTOutputs(true),
					)
				},
				wantErr: nil,
			}
		}(),

		// fail - destroy NFT (destruction disabled)
		func() *txExecTest {
			return &txExecTest{
				name:      "fail - destroy NFT (destruction disabled)",
				txBuilder: destroyNFTTxBuilder,
				txPreSignHook: func(t *iotago.Transaction) {
					t.Capabilities = iotago.TransactionCapabilitiesBitMaskWithCapabilities(
						iotago.WithTransactionCanDoAnything(),
						iotago.WithTransactionCanDestroyNFTOutputs(false),
					)
				},
				wantErr: iotago.ErrTxCapabilitiesNFTDestructionNotAllowed,
			}
		}(),
	}

	for _, tt := range tests {
		runNovaTransactionExecutionTest(t, tt)
	}
}

// TODO: add test case for transaction with context inputs.
func TestTxSemanticInputUnlocks(t *testing.T) {
	type test struct {
		name           string
		vmParams       *vm.Params
		resolvedInputs vm.ResolvedInputs
		tx             *iotago.SignedTransaction
		wantErr        error
	}
	tests := []*test{
		// ok
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, ident2AddrKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(12)

			accountInputID := inputIDs[3]
			accountIdent1 := iotago.AccountAddressFromOutputID(accountInputID)

			anchorInputID := inputIDs[8]
			anchorIdent1 := iotago.AnchorAddressFromOutputID(anchorInputID)

			nftIdent1 := tpkg.RandNFTAddress()
			nftIdent2 := tpkg.RandNFTAddress()

			defaultAmount := OneIOTA

			inputs := vm.InputSet{
				// basic output to create a signature unlock (owned by ident1)
				inputIDs[0]: &iotago.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				// basic output unlockable by sender as expired (owned by ident2)
				inputIDs[1]: &iotago.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: ident2,
							Slot:          5,
						},
					},
				},
				// basic output not unlockable by sender as not expired (owned by ident1)
				inputIDs[2]: &iotago.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: ident2,
							Slot:          30,
						},
					},
				},

				// account output that ownes the following outputs (owned by ident1)
				accountInputID: &iotago.AccountOutput{
					Amount:    defaultAmount,
					AccountID: iotago.AccountID{}, // empty on purpose as validation should resolve
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				// basic output (owned by accountIdent1)
				inputIDs[4]: &iotago.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: accountIdent1},
					},
				},
				// NFT output (owned by accountIdent1)
				inputIDs[5]: &iotago.NFTOutput{
					Amount: defaultAmount,
					NFTID:  nftIdent1.NFTID(),
					UnlockConditions: iotago.NFTOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: accountIdent1},
					},
				},
				// basic output (owned by nftIdent1)
				inputIDs[6]: &iotago.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: nftIdent1},
					},
				},
				// foundry output (owned by accountIdent1)
				inputIDs[7]: &iotago.FoundryOutput{
					Amount:       defaultAmount,
					SerialNumber: 0,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetInt64(100),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetInt64(1000),
					},
					UnlockConditions: iotago.FoundryOutputUnlockConditions{
						&iotago.ImmutableAccountUnlockCondition{Address: accountIdent1},
					},
				},

				// anchor output that ownes the following outputs (owned by ident1)
				inputIDs[8]: &iotago.AnchorOutput{
					Amount:   defaultAmount,
					AnchorID: iotago.AnchorID{}, // empty on purpose as validation should resolve
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident1},
						&iotago.GovernorAddressUnlockCondition{Address: ident1},
					},
				},
				// basic output (owned by anchorIdent1)
				inputIDs[9]: &iotago.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: anchorIdent1},
					},
				},
				// NFT output (owned by anchorIdent1)
				inputIDs[10]: &iotago.NFTOutput{
					Amount: defaultAmount,
					NFTID:  nftIdent2.NFTID(),
					UnlockConditions: iotago.NFTOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: anchorIdent1},
					},
				},
				// basic output (owned by nftIdent2)
				inputIDs[11]: &iotago.BasicOutput{
					Amount: defaultAmount,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: nftIdent2},
					},
				},
			}

			creationSlot := iotago.SlotIndex(10)
			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs:       inputIDs.UTXOInputs(),
					CreationSlot: creationSlot,
					Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.AccountOutput{
						Amount:    defaultAmount / 2,
						AccountID: accountIdent1.AccountID(),
						UnlockConditions: iotago.AccountOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					&iotago.AnchorOutput{
						Amount:     defaultAmount / 2,
						AnchorID:   anchorIdent1.AnchorID(),
						StateIndex: 1,
						UnlockConditions: iotago.AnchorOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident1},
						},
					},
				},
			}

			sigs, err := transaction.Sign(ident1AddrKeys, ident2AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet: inputs,
					CommitmentInput: &iotago.Commitment{
						Slot: iotago.SlotIndex(0),
					},
				},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]}, // basic output (owned by ident1)
						&iotago.SignatureUnlock{Signature: sigs[1]}, // basic output (owned by ident2)
						&iotago.ReferenceUnlock{Reference: 0},       // basic output (owned by ident1)
						&iotago.ReferenceUnlock{Reference: 0},       // account output (owned by ident1)
						&iotago.AccountUnlock{Reference: 3},         // basic output (owned by accountIdent1)
						&iotago.AccountUnlock{Reference: 3},         // NFT output (owned by accountIdent1)
						&iotago.NFTUnlock{Reference: 5},             // basic output (owned by nftIdent1)
						&iotago.AccountUnlock{Reference: 3},         // foundry output (owned by accountIdent1)
						&iotago.ReferenceUnlock{Reference: 0},       // anchor output (owned by ident1)
						&iotago.AnchorUnlock{Reference: 8},          // basic output (owned by anchorIdent1)
						&iotago.AnchorUnlock{Reference: 8},          // NFT output (owned by anchorIdent1)
						&iotago.NFTUnlock{Reference: 10},            // basic output (owned by nftIdent2
					},
				},
				wantErr: nil,
			}
		}(),

		// fail - invalid signature
		func() *test {
			ident1Sk, ident1, _ := tpkg.RandEd25519Identity()
			_, _, ident2AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			transaction := &iotago.Transaction{
				API:                testAPI,
				TransactionEssence: &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()},
			}

			sigs, err := transaction.Sign(ident2AddrKeys)
			require.NoError(t, err)

			copy(sigs[0].(*iotago.Ed25519Signature).PublicKey[:], ident1Sk.Public().(ed25519.PublicKey))

			return &test{
				name: "fail - invalid signature",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrEd25519SignatureInvalid,
			}
		}(),

		// fail - should contain reference unlock
		func() *test {
			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			transaction := &iotago.Transaction{API: testAPI, TransactionEssence: &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
			}}

			sigs, err := transaction.Sign(ident1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - should contain reference unlock",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrInvalidInputUnlock,
			}
		}(),

		// fail - should contain account unlock
		func() *test {
			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)

			accountIdent1 := iotago.AccountAddressFromOutputID(inputIDs[0])
			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AccountOutput{
					Amount:    100,
					AccountID: iotago.AccountID{},
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: accountIdent1},
					},
				},
			}

			transaction := &iotago.Transaction{API: testAPI, TransactionEssence: &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
			}}

			sigs, err := transaction.Sign(ident1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - should contain account unlock",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.ReferenceUnlock{Reference: 0},
					},
				},
				wantErr: iotago.ErrInvalidInputUnlock,
			}
		}(),

		// fail - should contain anchor unlock
		func() *test {
			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)

			anchorIdent1 := iotago.AnchorAddressFromOutputID(inputIDs[0])
			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AnchorOutput{
					Amount:   100,
					AnchorID: iotago.AnchorID{},
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident1},
						&iotago.GovernorAddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: anchorIdent1},
					},
				},
			}

			transaction := &iotago.Transaction{API: testAPI, TransactionEssence: &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
			}}

			sigs, err := transaction.Sign(ident1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - should contain anchor unlock",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.ReferenceUnlock{Reference: 0},
					},
				},
				wantErr: iotago.ErrInvalidInputUnlock,
			}
		}(),

		// fail - should contain NFT unlock
		func() *test {
			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)

			nftIdent1 := iotago.NFTAddressFromOutputID(inputIDs[0])
			inputs := vm.InputSet{
				inputIDs[0]: &iotago.NFTOutput{
					Amount: 100,
					NFTID:  iotago.NFTID{},
					UnlockConditions: iotago.NFTOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: nftIdent1},
					},
				},
			}

			transaction := &iotago.Transaction{API: testAPI, TransactionEssence: &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
			}}

			sigs, err := transaction.Sign(ident1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - should contain NFT unlock",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.ReferenceUnlock{Reference: 0},
					},
				},
				wantErr: iotago.ErrInvalidInputUnlock,
			}
		}(),

		// fail - circular NFT unlock
		func() *test {
			inputIDs := tpkg.RandOutputIDs(2)

			nftIdent1 := iotago.NFTAddressFromOutputID(inputIDs[0])
			nftIdent2 := iotago.NFTAddressFromOutputID(inputIDs[1])

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.NFTOutput{
					Amount: 100,
					NFTID:  nftIdent1.NFTID(),
					UnlockConditions: iotago.NFTOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: nftIdent2},
					},
				},
				inputIDs[1]: &iotago.NFTOutput{
					Amount: 100,
					NFTID:  nftIdent2.NFTID(),
					UnlockConditions: iotago.NFTOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: nftIdent2},
					},
				},
			}

			transaction := &iotago.Transaction{API: testAPI, TransactionEssence: &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
			}}
			_, err := transaction.Sign()
			require.NoError(t, err)
			return &test{
				name: "fail - circular NFT unlock",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.NFTUnlock{Reference: 1},
						&iotago.NFTUnlock{Reference: 0},
					},
				},
				wantErr: iotago.ErrInvalidInputUnlock,
			}
		}(),

		// fail - sender can not unlock yet
		func() *test {
			_, ident1, _ := tpkg.RandEd25519Identity()
			_, ident2, ident2AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: ident2,
							Slot:          20,
						},
					},
				},
			}

			creationSlot := iotago.SlotIndex(5)
			transaction := &iotago.Transaction{API: testAPI, TransactionEssence: &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs(), CreationSlot: creationSlot}}

			sigs, err := transaction.Sign(ident2AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - sender can not unlock yet",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet: inputs,
					CommitmentInput: &iotago.Commitment{
						Slot: iotago.SlotIndex(0),
					},
				},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrExpirationConditionUnlockFailed,
			}
		}(),

		// fail - receiver can not unlock anymore
		func() *test {
			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()
			_, ident2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: ident2,
							Slot:          10,
						},
					},
				},
			}

			creationSlot := iotago.SlotIndex(10)
			transaction := &iotago.Transaction{API: testAPI, TransactionEssence: &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs(), CreationSlot: creationSlot}}

			sigs, err := transaction.Sign(ident1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - receiver can not unlock anymore",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet: inputs,
					CommitmentInput: &iotago.Commitment{
						Slot: creationSlot,
					},
				},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrEd25519PubKeyAndAddrMismatch,
			}
		}(),

		// fail - referencing other account unlocked by source account
		func() *test {
			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(3)

			var (
				accountAddr1 = tpkg.RandAccountAddress()
				accountAddr2 = tpkg.RandAccountAddress()
				accountAddr3 = tpkg.RandAccountAddress()
			)

			inputs := vm.InputSet{
				// owned by ident1
				inputIDs[0]: &iotago.AccountOutput{
					Amount:    100,
					AccountID: accountAddr1.AccountID(),
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				// owned by account1
				inputIDs[1]: &iotago.AccountOutput{
					Amount:    100,
					AccountID: accountAddr2.AccountID(),
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: accountAddr1},
					},
				},
				// owned by account1
				inputIDs[2]: &iotago.AccountOutput{
					Amount:    100,
					AccountID: accountAddr3.AccountID(),
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: accountAddr1},
					},
				},
			}

			transaction := &iotago.Transaction{API: testAPI, TransactionEssence: &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
			}}

			sigs, err := transaction.Sign(ident1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - referencing other account unlocked by source account",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.AccountUnlock{Reference: 0},
						// error, should be 0, because account3 is unlocked by account1, not account2
						&iotago.AccountUnlock{Reference: 1},
					},
				},
				wantErr: iotago.ErrInvalidInputUnlock,
			}
		}(),

		// fail - referencing other anchor unlocked by source anchor
		func() *test {
			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(3)

			var (
				anchorAddr1 = tpkg.RandAnchorAddress()
				anchorAddr2 = tpkg.RandAnchorAddress()
				anchorAddr3 = tpkg.RandAnchorAddress()
			)

			inputs := vm.InputSet{
				// owned by ident1
				inputIDs[0]: &iotago.AnchorOutput{
					Amount:   100,
					AnchorID: anchorAddr1.AnchorID(),
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident1},
						&iotago.GovernorAddressUnlockCondition{Address: ident1},
					},
				},
				// owned by anchor1
				inputIDs[1]: &iotago.AnchorOutput{
					Amount:   100,
					AnchorID: anchorAddr2.AnchorID(),
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: anchorAddr1},
						&iotago.GovernorAddressUnlockCondition{Address: anchorAddr1},
					},
				},
				// owned by anchor1
				inputIDs[2]: &iotago.AnchorOutput{
					Amount:   100,
					AnchorID: anchorAddr3.AnchorID(),
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: anchorAddr1},
						&iotago.GovernorAddressUnlockCondition{Address: anchorAddr1},
					},
				},
			}

			transaction := &iotago.Transaction{API: testAPI, TransactionEssence: &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
			}}

			sigs, err := transaction.Sign(ident1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - referencing other anchor unlocked by source anchor",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.AnchorUnlock{Reference: 0},
						// error, should be 0, because anchor3 is unlocked by anchor1, not anchor2
						&iotago.AnchorUnlock{Reference: 1},
					},
				},
				wantErr: iotago.ErrInvalidInputUnlock,
			}
		}(),

		// fail - anchor output not state transitioning
		func() *test {
			_, ident1, _ := tpkg.RandEd25519Identity()
			_, ident2, ident2AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)

			anchorAddr1 := tpkg.RandAnchorAddress()

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AnchorOutput{
					Amount:   100,
					AnchorID: anchorAddr1.AnchorID(),
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident1},
						&iotago.GovernorAddressUnlockCondition{Address: ident2},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: anchorAddr1},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.AnchorOutput{
						Amount:   100,
						AnchorID: anchorAddr1.AnchorID(),
						UnlockConditions: iotago.AnchorOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident2},
						},
					},
				},
			}

			sigs, err := transaction.Sign(ident2AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - anchor output not state transitioning",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.AnchorUnlock{Reference: 0},
					},
				},
				wantErr: iotago.ErrInvalidInputUnlock,
			}
		}(),

		// fail - wrong unlock for foundry
		func() *test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(2)
			foundryOutput := &iotago.FoundryOutput{
				Amount:       100,
				SerialNumber: 5,
				TokenScheme: &iotago.SimpleTokenScheme{
					MintedTokens:  new(big.Int).SetInt64(1000),
					MeltedTokens:  big.NewInt(0),
					MaximumSupply: new(big.Int).SetInt64(10000),
				},
				UnlockConditions: iotago.FoundryOutputUnlockConditions{
					&iotago.ImmutableAccountUnlockCondition{Address: accountAddr1},
				},
			}

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AccountOutput{
					Amount:    100,
					AccountID: accountAddr1.AccountID(),
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: foundryOutput,
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.AccountOutput{
						Amount:    100,
						AccountID: accountAddr1.AccountID(),
						UnlockConditions: iotago.AccountOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					foundryOutput,
				},
			}

			sigs, err := transaction.Sign(ident1AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - wrong unlock for foundry",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						// should be an AccountUnlock
						&iotago.ReferenceUnlock{Reference: 0},
					},
				},
				wantErr: iotago.ErrInvalidInputUnlock,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := novaVM.ValidateUnlocks(tt.tx, tt.resolvedInputs)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

// TODO: add test case for block issuer with keys (differently priced).
func TestTxSemanticDeposit(t *testing.T) {
	type test struct {
		name           string
		vmParams       *vm.Params
		resolvedInputs vm.ResolvedInputs
		tx             *iotago.SignedTransaction
		wantErr        error
	}
	tests := []*test{
		// ok
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, ident2AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(3)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				// unlocked by ident1 as it is not expired
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 500,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: ident2,
							Amount:        420,
						},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: ident2,
							Slot:          30,
						},
					},
				},
				// unlocked by ident2 as it is expired
				inputIDs[2]: &iotago.BasicOutput{
					Amount: 500,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: ident2,
							Amount:        420,
						},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: ident2,
							Slot:          2,
						},
					},
				},
			}

			creationSlot := iotago.SlotIndex(5)
			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs:       inputIDs.UTXOInputs(),
					CreationSlot: creationSlot,
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 180,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
					&iotago.BasicOutput{
						// return via ident1 + reclaim
						Amount: 420 + 500,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
					},
				},
			}
			sigs, err := transaction.Sign(ident1AddrKeys, ident2AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet: inputs,
					CommitmentInput: &iotago.Commitment{
						Slot: creationSlot,
					},
				},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.ReferenceUnlock{Reference: 0},
						&iotago.SignatureUnlock{Signature: sigs[1]},
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - more storage deposit returned via more outputs
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 1000,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: ident2,
							Amount:        420,
						},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						// returns 200 to ident2
						Amount: 200,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
					},
					&iotago.BasicOutput{
						// returns 221 to ident2
						Amount: 221,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
					},
					&iotago.BasicOutput{
						// remainder to random address
						Amount: 579,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := transaction.Sign(ident1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok - more storage deposit returned via more outputs",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),

		// fail - unbalanced, more on output than input
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 50,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs:       inputIDs.UTXOInputs(),
					CreationSlot: 5,
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 100,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := transaction.Sign(ident1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - unbalanced, more on output than input",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrInputOutputSumMismatch,
			}
		}(),

		// fail - unbalanced, more on input than output
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs:       inputIDs.UTXOInputs(),
					CreationSlot: 5,
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 50,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := transaction.Sign(ident1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - unbalanced, more on input than output",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrInputOutputSumMismatch,
			}
		}(),

		// fail - return not fulfilled
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 500,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: ident2,
							Amount:        420,
						},
						// not yet expired, so ident1 needs to unlock
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: ident2,
							Slot:          30,
						},
					},
				},
			}

			creationSlot := iotago.SlotIndex(5)
			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs:       inputIDs.UTXOInputs(),
					CreationSlot: creationSlot,
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 500,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
			}
			sigs, err := transaction.Sign(ident1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - return not fulfilled",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet: inputs,
					CommitmentInput: &iotago.Commitment{
						Slot: creationSlot,
					},
				},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrReturnAmountNotFulFilled,
			}
		}(),

		// fail - storage deposit return not basic output
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 500,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: ident2,
							Amount:        420,
						},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 80,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					&iotago.NFTOutput{
						Amount: 420,
						UnlockConditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
					},
				},
			}
			sigs, err := transaction.Sign(ident1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - storage deposit return not basic output",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrReturnAmountNotFulFilled,
			}
		}(),

		// fail - storage deposit return has additional unlocks
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 500,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: ident2,
							Amount:        420,
						},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 80,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					&iotago.BasicOutput{
						Amount: 420,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident2},
							&iotago.ExpirationUnlockCondition{
								ReturnAddress: ident1,
								Slot:          10,
							},
						},
					},
				},
			}
			sigs, err := transaction.Sign(ident1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - storage deposit return has additional unlocks",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrReturnAmountNotFulFilled,
			}
		}(),

		// fail - storage deposit return has feature
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 500,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: ident2,
							Amount:        420,
						},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 80,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					&iotago.BasicOutput{
						Amount: 420,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
						Features: iotago.BasicOutputFeatures{
							&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("foo")}},
						},
					},
				},
			}
			sigs, err := transaction.Sign(ident1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - storage deposit return has feature",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrReturnAmountNotFulFilled,
			}
		}(),

		// fail - storage deposit return has native tokens
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)
			ntID := tpkg.Rand38ByteArray()

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 500,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: ident2,
							Amount:        420,
						},
					},
					Features: iotago.BasicOutputFeatures{
						&iotago.NativeTokenFeature{
							ID:     ntID,
							Amount: new(big.Int).SetUint64(1000),
						},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 80,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					&iotago.BasicOutput{
						Amount: 420,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
						Features: iotago.BasicOutputFeatures{
							&iotago.NativeTokenFeature{
								ID:     ntID,
								Amount: new(big.Int).SetUint64(1000),
							},
						},
					},
				},
			}
			sigs, err := transaction.Sign(ident1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - storage deposit return has native tokens",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrReturnAmountNotFulFilled,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAndExecuteSignedTransaction(tt.tx, tt.resolvedInputs, vm.ExecFuncBalancedBaseTokens())
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestTxSemanticNativeTokens(t *testing.T) {
	foundryAccountIdent := tpkg.RandAccountAddress()
	foundryMaxSupply := new(big.Int).SetInt64(1000)
	foundryMintedSupply := new(big.Int).SetInt64(500)

	inUnrelatedFoundryOutput := &iotago.FoundryOutput{
		Amount:       100,
		SerialNumber: 0,
		TokenScheme: &iotago.SimpleTokenScheme{
			MintedTokens:  foundryMintedSupply,
			MeltedTokens:  big.NewInt(0),
			MaximumSupply: foundryMaxSupply,
		},
		UnlockConditions: iotago.FoundryOutputUnlockConditions{
			&iotago.ImmutableAccountUnlockCondition{Address: foundryAccountIdent},
		},
	}

	outUnrelatedFoundryOutput := &iotago.FoundryOutput{
		Amount:       100,
		SerialNumber: 0,
		TokenScheme: &iotago.SimpleTokenScheme{
			MintedTokens:  foundryMintedSupply,
			MeltedTokens:  big.NewInt(0),
			MaximumSupply: foundryMaxSupply,
		},
		UnlockConditions: iotago.FoundryOutputUnlockConditions{
			&iotago.ImmutableAccountUnlockCondition{Address: foundryAccountIdent},
		},
	}

	type test struct {
		name           string
		vmParams       *vm.Params
		resolvedInputs vm.ResolvedInputs
		tx             *iotago.SignedTransaction
		wantErr        error
	}
	tests := []*test{
		// ok
		func() *test {
			inputIDs := tpkg.RandOutputIDs(2)

			nativeTokenFeature1 := tpkg.RandNativeTokenFeature()
			nativeTokenFeature2 := tpkg.RandNativeTokenFeature()

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
					Features: iotago.BasicOutputFeatures{
						nativeTokenFeature1,
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
					Features: iotago.BasicOutputFeatures{
						nativeTokenFeature2,
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 100,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: iotago.BasicOutputFeatures{
							nativeTokenFeature1,
						},
					},
					&iotago.BasicOutput{
						Amount: 100,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: iotago.BasicOutputFeatures{
							nativeTokenFeature2,
						},
					},
				},
			}

			return &test{
				name: "ok",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks:     iotago.Unlocks{},
				},
				wantErr: nil,
			}
		}(),

		// ok - consolidate native token (same type)
		func() *test {
			inputIDs := tpkg.RandOutputIDs(iotago.MaxInputsCount)
			nativeToken := tpkg.RandNativeTokenFeature()

			inputs := vm.InputSet{}
			for i := 0; i < iotago.MaxInputsCount; i++ {
				inputs[inputIDs[i]] = &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
					Features: iotago.BasicOutputFeatures{
						&iotago.NativeTokenFeature{
							ID:     nativeToken.ID,
							Amount: big.NewInt(1),
						},
					},
				}
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 100 * iotago.MaxInputsCount,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: iotago.BasicOutputFeatures{
							&iotago.NativeTokenFeature{
								ID:     nativeToken.ID,
								Amount: big.NewInt(iotago.MaxInputsCount),
							},
						},
					},
				},
			}

			return &test{
				name: "ok - consolidate native token (same type)",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks:     iotago.Unlocks{},
				},
				wantErr: nil,
			}
		}(),

		// ok - most possible tokens in a tx
		func() *test {
			inputIDs := tpkg.RandOutputIDs(iotago.MaxInputsCount)

			nativeTokenFeatures := make([]*iotago.NativeTokenFeature, iotago.MaxInputsCount)
			for i := 0; i < iotago.MaxInputsCount; i++ {
				nativeTokenFeatures[i] = tpkg.RandNativeTokenFeature()
			}

			inputs := vm.InputSet{}
			for i := 0; i < iotago.MaxInputsCount; i++ {
				inputs[inputIDs[i]] = &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
					Features: iotago.BasicOutputFeatures{
						nativeTokenFeatures[i],
					},
				}
			}

			outputs := make(iotago.TxEssenceOutputs, iotago.MaxOutputsCount)
			for i := range outputs {
				outputs[i] = &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
					Features: iotago.BasicOutputFeatures{
						nativeTokenFeatures[i],
					},
				}
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: outputs,
			}

			return &test{
				name: "ok - most possible tokens in a tx",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks:     iotago.Unlocks{},
				},
				wantErr: nil,
			}
		}(),

		// fail - unbalanced on output
		func() *test {
			inputIDs := tpkg.RandOutputIDs(1)

			nativeTokenFeature := tpkg.RandNativeTokenFeature()

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
					Features: iotago.BasicOutputFeatures{
						nativeTokenFeature,
					},
				},
			}

			// unbalance by making one token be excess on the output side
			cpyNativeTokenFeature := nativeTokenFeature.Clone()
			cpyNativeTokenFeature.(*iotago.NativeTokenFeature).Amount = big.NewInt(0).Add(nativeTokenFeature.Amount, big.NewInt(1))

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 100,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: iotago.BasicOutputFeatures{
							cpyNativeTokenFeature,
						},
					},
				},
			}

			return &test{
				name: "fail - unbalanced on output",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks:     iotago.Unlocks{},
				},
				wantErr: iotago.ErrNativeTokenSumUnbalanced,
			}
		}(),

		// fail - unbalanced with unrelated foundry in term of new output tokens
		func() *test {
			inputIDs := tpkg.RandOutputIDs(3)

			nativeTokenFeature1 := tpkg.RandNativeTokenFeature()
			nativeTokenFeature2 := nativeTokenFeature1.Clone()

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
					Features: iotago.BasicOutputFeatures{
						nativeTokenFeature1,
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
					Features: iotago.BasicOutputFeatures{
						nativeTokenFeature2,
					},
				},
				inputIDs[2]: inUnrelatedFoundryOutput,
			}

			// unbalance by making one token be excess on the output side
			cpyNativeTokenFeature := nativeTokenFeature1.Clone()
			cpyNativeTokenFeature.(*iotago.NativeTokenFeature).Amount = big.NewInt(0).Add(nativeTokenFeature1.Amount, big.NewInt(1))

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 100,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: iotago.BasicOutputFeatures{
							cpyNativeTokenFeature,
						},
					},
					&iotago.BasicOutput{
						Amount: 100,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: iotago.BasicOutputFeatures{
							nativeTokenFeature2,
						},
					},
					outUnrelatedFoundryOutput,
				},
			}

			return &test{
				name: "fail - unbalanced with unrelated foundry in term of new output tokens",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks:     iotago.Unlocks{},
				},
				wantErr: iotago.ErrNativeTokenSumUnbalanced,
			}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := novaVM.Execute(tt.tx.Transaction, tt.resolvedInputs, make(vm.UnlockedIdentities), vm.ExecFuncBalancedNativeTokens())
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestTxSemanticOutputsSender(t *testing.T) {
	type test struct {
		name           string
		vmParams       *vm.Params
		resolvedInputs vm.ResolvedInputs
		tx             *iotago.SignedTransaction
		wantErr        error
	}
	tests := []*test{
		// ok
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(4)
			accountAddr := tpkg.RandAccountAddress()
			anchorAddr := tpkg.RandAnchorAddress()
			nftAddr := tpkg.RandNFTAddress()

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: &iotago.AccountOutput{
					Amount:    100,
					AccountID: accountAddr.AccountID(),
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[2]: &iotago.AnchorOutput{
					Amount:     100,
					AnchorID:   anchorAddr.AnchorID(),
					StateIndex: 1,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident1},
						&iotago.GovernorAddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[3]: &iotago.NFTOutput{
					Amount: 100,
					NFTID:  nftAddr.NFTID(),
					UnlockConditions: iotago.NFTOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: func() iotago.TxEssenceOutputs {
					outputs := make(iotago.TxEssenceOutputs, 0)

					// we need to do a state transition to unlock the sender feature for the anchor output
					outputs = append(outputs, &iotago.AnchorOutput{
						Amount:     100,
						AnchorID:   anchorAddr.AnchorID(),
						StateIndex: 2,
						UnlockConditions: iotago.AnchorOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident1},
						},
					})

					for _, sender := range []iotago.Address{ident1, accountAddr, anchorAddr, nftAddr} {
						outputs = append(outputs, &iotago.BasicOutput{
							Amount: 1337,
							UnlockConditions: iotago.BasicOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
							},
							Features: iotago.BasicOutputFeatures{
								&iotago.SenderFeature{Address: sender},
							},
						})

						outputs = append(outputs, &iotago.AccountOutput{
							Amount: 1337,
							UnlockConditions: iotago.AccountOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: ident1},
							},
							Features: iotago.AccountOutputFeatures{
								&iotago.SenderFeature{Address: sender},
							},
						})

						outputs = append(outputs, &iotago.AnchorOutput{
							Amount: 1337,
							UnlockConditions: iotago.AnchorOutputUnlockConditions{
								&iotago.StateControllerAddressUnlockCondition{Address: ident1},
								&iotago.GovernorAddressUnlockCondition{Address: ident1},
							},
							Features: iotago.AnchorOutputFeatures{
								&iotago.SenderFeature{Address: sender},
							},
						})

						outputs = append(outputs, &iotago.NFTOutput{
							Amount: 1337,
							UnlockConditions: iotago.NFTOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
							},
							Features: iotago.NFTOutputFeatures{
								&iotago.SenderFeature{Address: sender},
							},
						})
					}
					return outputs
				}(),
			}
			sigs, err := transaction.Sign(ident1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.ReferenceUnlock{Reference: 0},
						&iotago.ReferenceUnlock{Reference: 0},
						&iotago.ReferenceUnlock{Reference: 0},
					},
				},
				wantErr: nil,
			}
		}(),

		// fail - sender not unlocked
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 1337,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: iotago.BasicOutputFeatures{
							&iotago.SenderFeature{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := transaction.Sign(ident1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - sender not unlocked",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrSenderFeatureNotUnlocked,
			}
		}(),

		// fail - sender not unlocked due to governance transition
		func() *test {
			_, stateController, _ := tpkg.RandEd25519Identity()
			_, governor, governorAddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)
			anchorAddr := tpkg.RandAnchorAddress()
			anchorID := anchorAddr.AnchorID()

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AnchorOutput{
					Amount:   100,
					AnchorID: anchorID,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: stateController},
						&iotago.GovernorAddressUnlockCondition{Address: governor},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.AnchorOutput{
						Amount:   100,
						AnchorID: anchorID,
						UnlockConditions: iotago.AnchorOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
						Features: iotago.AnchorOutputFeatures{
							&iotago.SenderFeature{Address: anchorAddr},
						},
					},
				},
			}
			sigs, err := transaction.Sign(governorAddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - sender not unlocked due to governance transition",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrSenderFeatureNotUnlocked,
			}
		}(),

		// ok - anchor addr unlocked with state transition
		func() *test {
			_, stateController, stateControllerAddrKeys := tpkg.RandEd25519Identity()
			_, governor, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)
			anchorAddr := tpkg.RandAnchorAddress()
			anchorID := anchorAddr.AnchorID()
			currentStateIndex := uint32(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AnchorOutput{
					Amount:     100,
					AnchorID:   anchorID,
					StateIndex: currentStateIndex,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: stateController},
						&iotago.GovernorAddressUnlockCondition{Address: governor},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.AnchorOutput{
						Amount:     100,
						AnchorID:   anchorID,
						StateIndex: currentStateIndex + 1,
						UnlockConditions: iotago.AnchorOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
						Features: iotago.AnchorOutputFeatures{
							&iotago.SenderFeature{Address: anchorAddr},
						},
					},
				},
			}
			sigs, err := transaction.Sign(stateControllerAddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok - anchor addr unlocked with state transition",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - sender is governor address
		func() *test {
			_, stateController, _ := tpkg.RandEd25519Identity()
			_, governor, governorAddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)
			anchorAddr := tpkg.RandAnchorAddress()
			anchorID := anchorAddr.AnchorID()

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AnchorOutput{
					Amount:   100,
					AnchorID: anchorID,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: stateController},
						&iotago.GovernorAddressUnlockCondition{Address: governor},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.AnchorOutput{
						Amount:   100,
						AnchorID: anchorID,
						UnlockConditions: iotago.AnchorOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
						Features: iotago.AnchorOutputFeatures{
							&iotago.SenderFeature{Address: governor},
						},
					},
				},
			}
			sigs, err := transaction.Sign(governorAddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok - sender is governor address",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - multi address in sender feature
		func() *test {
			_, ident1, ident1Keys := tpkg.RandEd25519Identity()
			_, ident2, ident2Keys := tpkg.RandEd25519Identity()

			multiAddr := iotago.MultiAddress{
				Addresses: iotago.AddressesWithWeight{
					{
						Address: ident1,
						Weight:  5,
					},
					{
						Address: ident2,
						Weight:  10,
					},
					{
						Address: tpkg.RandNFTAddress(),
						Weight:  1,
					},
				},
				Threshold: 12,
			}

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: &multiAddr},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 100,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: iotago.BasicOutputFeatures{
							&iotago.SenderFeature{Address: &multiAddr},
						},
					},
				},
			}

			sigs, err := transaction.Sign(ident1Keys, ident2Keys)
			require.NoError(t, err)

			return &test{
				name: "ok - multi address in sender feature",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.MultiUnlock{
							Unlocks: iotago.Unlocks{
								&iotago.SignatureUnlock{Signature: sigs[0]},
								&iotago.SignatureUnlock{Signature: sigs[1]},
								&iotago.EmptyUnlock{},
							},
						},
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - restricted multi address in sender and issuer feature
		func() *test {
			_, ident1, ident1Keys := tpkg.RandEd25519Identity()
			_, ident2, ident2Keys := tpkg.RandEd25519Identity()

			multiAddr := iotago.MultiAddress{
				Addresses: iotago.AddressesWithWeight{
					{
						Address: ident1,
						Weight:  5,
					},
					{
						Address: ident2,
						Weight:  10,
					},
					{
						Address: tpkg.RandAccountAddress(),
						Weight:  1,
					},
				},
				Threshold: 12,
			}

			restrictedAddr := iotago.RestrictedAddress{
				Address:             &multiAddr,
				AllowedCapabilities: iotago.AddressCapabilitiesBitMaskWithCapabilities(iotago.WithAddressCanReceiveMana(true)),
			}

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: &restrictedAddr},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.NFTOutput{
						Amount: 100,
						UnlockConditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: iotago.NFTOutputFeatures{
							// We can use the restricted address...
							&iotago.SenderFeature{Address: &restrictedAddr},
						},
						ImmutableFeatures: iotago.NFTOutputImmFeatures{
							// ...or the underlying address.
							&iotago.IssuerFeature{Address: &multiAddr},
						},
					},
				},
			}

			sigs, err := transaction.Sign(ident1Keys, ident2Keys)
			require.NoError(t, err)

			return &test{
				name: "ok - restricted multi address in sender and issuer feature",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.MultiUnlock{
							Unlocks: iotago.Unlocks{
								&iotago.SignatureUnlock{Signature: sigs[0]},
								&iotago.SignatureUnlock{Signature: sigs[1]},
								&iotago.EmptyUnlock{},
							},
						},
					},
				},
				wantErr: nil,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAndExecuteSignedTransaction(tt.tx, tt.resolvedInputs, vm.ExecFuncSenderUnlocked())
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestTxSemanticOutputsIssuer(t *testing.T) {
	type test struct {
		name           string
		vmParams       *vm.Params
		resolvedInputs vm.ResolvedInputs
		tx             *iotago.SignedTransaction
		wantErr        error
	}
	tests := []*test{
		// fail - issuer not unlocked due to governance transition
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, stateController, _ := tpkg.RandEd25519Identity()
			_, governor, governorAddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)
			anchorAddr := tpkg.RandAnchorAddress()
			anchorID := anchorAddr.AnchorID()

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AnchorOutput{
					Amount:   100,
					AnchorID: anchorID,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: stateController},
						&iotago.GovernorAddressUnlockCondition{Address: governor},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.AnchorOutput{
						Amount:   100,
						AnchorID: anchorID,
						UnlockConditions: iotago.AnchorOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
					},
					&iotago.NFTOutput{
						Amount: 100,
						UnlockConditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
						ImmutableFeatures: iotago.NFTOutputImmFeatures{
							&iotago.IssuerFeature{Address: anchorAddr},
						},
					},
				},
			}
			sigs, err := transaction.Sign(governorAddrKeys, ident1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - issuer not unlocked due to governance transition",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.SignatureUnlock{Signature: sigs[1]},
					},
				},
				wantErr: iotago.ErrIssuerFeatureNotUnlocked,
			}
		}(),

		// ok - issuer unlocked with state transition
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, stateController, stateControllerAddrKeys := tpkg.RandEd25519Identity()
			_, governor, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)
			anchorAddr := tpkg.RandAnchorAddress()
			anchorID := anchorAddr.AnchorID()
			currentStateIndex := uint32(1)

			nftAddr := tpkg.RandNFTAddress()

			inputs := vm.InputSet{
				// possible issuers: anchorAddr, stateController, nftAddr, ident1
				inputIDs[0]: &iotago.AnchorOutput{
					Amount:     100,
					AnchorID:   anchorID,
					StateIndex: currentStateIndex,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: stateController},
						&iotago.GovernorAddressUnlockCondition{Address: governor},
					},
				},
				inputIDs[1]: &iotago.NFTOutput{
					Amount: 900,
					NFTID:  nftAddr.NFTID(),
					UnlockConditions: iotago.NFTOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: iotago.TxEssenceOutputs{
					// transitioned anchor + nft
					&iotago.AnchorOutput{
						Amount:     100,
						AnchorID:   anchorID,
						StateIndex: currentStateIndex + 1,
						UnlockConditions: iotago.AnchorOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
					},
					&iotago.NFTOutput{
						Amount: 100,
						NFTID:  nftAddr.NFTID(),
						UnlockConditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					// issuer is anchorAddr
					&iotago.NFTOutput{
						Amount: 100,
						UnlockConditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
						ImmutableFeatures: iotago.NFTOutputImmFeatures{
							&iotago.IssuerFeature{Address: anchorAddr},
						},
					},
					&iotago.AnchorOutput{
						Amount: 100,
						UnlockConditions: iotago.AnchorOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
						ImmutableFeatures: iotago.AnchorOutputImmFeatures{
							&iotago.IssuerFeature{Address: anchorAddr},
						},
					},
					// issuer is stateController
					&iotago.NFTOutput{
						Amount: 100,
						UnlockConditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
						ImmutableFeatures: iotago.NFTOutputImmFeatures{
							&iotago.IssuerFeature{Address: stateController},
						},
					},
					&iotago.AnchorOutput{
						Amount: 100,
						UnlockConditions: iotago.AnchorOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
						ImmutableFeatures: iotago.AnchorOutputImmFeatures{
							&iotago.IssuerFeature{Address: stateController},
						},
					},
					// issuer is nftAddr
					&iotago.NFTOutput{
						Amount: 100,
						UnlockConditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
						ImmutableFeatures: iotago.NFTOutputImmFeatures{
							&iotago.IssuerFeature{Address: nftAddr},
						},
					},
					&iotago.AnchorOutput{
						Amount: 100,
						UnlockConditions: iotago.AnchorOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
						ImmutableFeatures: iotago.AnchorOutputImmFeatures{
							&iotago.IssuerFeature{Address: nftAddr},
						},
					},
					// issuer is ident1
					&iotago.NFTOutput{
						Amount: 100,
						UnlockConditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
						ImmutableFeatures: iotago.NFTOutputImmFeatures{
							&iotago.IssuerFeature{Address: ident1},
						},
					},
					&iotago.AnchorOutput{
						Amount: 100,
						UnlockConditions: iotago.AnchorOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
						ImmutableFeatures: iotago.AnchorOutputImmFeatures{
							&iotago.IssuerFeature{Address: ident1},
						},
					},
				},
			}
			sigs, err := transaction.Sign(stateControllerAddrKeys, ident1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok - issuer unlocked with state transition",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.SignatureUnlock{Signature: sigs[1]},
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - issuer is the governor
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, stateController, _ := tpkg.RandEd25519Identity()
			_, governor, governorAddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)
			anchorAddr := tpkg.RandAnchorAddress()
			anchorID := anchorAddr.AnchorID()

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AnchorOutput{
					Amount:   100,
					AnchorID: anchorID,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: stateController},
						&iotago.GovernorAddressUnlockCondition{Address: governor},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.AnchorOutput{
						Amount:   100,
						AnchorID: anchorID,
						UnlockConditions: iotago.AnchorOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
					},
					&iotago.NFTOutput{
						Amount: 100,
						UnlockConditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
						ImmutableFeatures: iotago.NFTOutputImmFeatures{
							&iotago.IssuerFeature{Address: governor},
						},
					},
				},
			}
			sigs, err := transaction.Sign(governorAddrKeys, ident1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok - issuer is the governor",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.SignatureUnlock{Signature: sigs[1]},
					},
				},
				wantErr: nil,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAndExecuteSignedTransaction(tt.tx, tt.resolvedInputs)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestTxSemanticTimelocks(t *testing.T) {
	type test struct {
		name           string
		vmParams       *vm.Params
		resolvedInputs vm.ResolvedInputs
		tx             *iotago.SignedTransaction
		wantErr        error
	}
	tests := []*test{
		// ok
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.TimelockUnlockCondition{
							Slot: 5,
						},
					},
				},
			}

			creationSlot := iotago.SlotIndex(10)
			transaction := &iotago.Transaction{API: testAPI, TransactionEssence: &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs(), CreationSlot: creationSlot}}
			sigs, err := transaction.Sign(ident1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet: inputs,
					CommitmentInput: &iotago.Commitment{
						Slot: creationSlot,
					},
				},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),

		// fail - timelock not expired
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.TimelockUnlockCondition{
							Slot: 25,
						},
					},
				},
			}

			creationSlot := iotago.SlotIndex(10)
			transaction := &iotago.Transaction{API: testAPI, TransactionEssence: &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs(), CreationSlot: creationSlot}}
			sigs, err := transaction.Sign(ident1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - timelock not expired",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet: inputs,
					CommitmentInput: &iotago.Commitment{
						Slot: creationSlot,
					},
				},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrTimelockNotExpired,
			}
		}(),

		// fail - no commitment input for timelock
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.TimelockUnlockCondition{
							Slot: 1000,
						},
					},
				},
			}

			creationSlot := iotago.SlotIndex(1005)
			transaction := &iotago.Transaction{API: testAPI, TransactionEssence: &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs(), CreationSlot: creationSlot}}
			sigs, err := transaction.Sign(ident1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - no commitment input for timelock",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet: inputs,
				},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrTimelockConditionCommitmentInputRequired,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := novaVM.Execute(tt.tx.Transaction, tt.resolvedInputs, make(vm.UnlockedIdentities), vm.ExecFuncTimelocks())
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

// TODO: add some more failing test cases.
func TestTxSemanticMana(t *testing.T) {
	type test struct {
		name           string
		vmParams       *vm.Params
		resolvedInputs vm.ResolvedInputs
		tx             *iotago.SignedTransaction
		wantErr        error
	}
	tests := []*test{
		// ok - stored Mana only without allotment"
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDsWithCreationSlot(10, 1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: OneIOTA,
					Mana:   iotago.MaxMana,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs:       inputIDs.UTXOInputs(),
					CreationSlot: 10 + 100*testProtoParams.ParamEpochDurationInSlots(),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: OneIOTA,
						Mana: func() iotago.Mana {
							var creationSlot iotago.SlotIndex = 10
							targetSlot := 10 + 100*testProtoParams.ParamEpochDurationInSlots()

							input := inputs[inputIDs[0]]
							storageScoreStructure := iotago.NewStorageScoreStructure(testProtoParams.StorageScoreParameters())
							potentialMana, err := iotago.PotentialMana(testAPI.ManaDecayProvider(), storageScoreStructure, input, creationSlot, targetSlot)
							require.NoError(t, err)

							storedMana, err := testAPI.ManaDecayProvider().ManaWithDecay(iotago.MaxMana, creationSlot, targetSlot)
							require.NoError(t, err)

							return potentialMana + storedMana
						}(),
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := transaction.Sign(ident1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok - stored Mana only without allotment",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),

		// ok - stored and allotted
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDsWithCreationSlot(10, 1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: OneIOTA,
					Mana:   iotago.MaxMana,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
					Allotments: iotago.Allotments{
						&iotago.Allotment{Mana: 50},
					},
					CreationSlot: 10 + 100*testProtoParams.ParamEpochDurationInSlots(),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: OneIOTA,
						Mana: func() iotago.Mana {
							var creationSlot iotago.SlotIndex = 10
							targetSlot := 10 + 100*testProtoParams.ParamEpochDurationInSlots()

							input := inputs[inputIDs[0]]
							storageScoreStructure := iotago.NewStorageScoreStructure(testProtoParams.StorageScoreParameters())
							potentialMana, err := iotago.PotentialMana(testAPI.ManaDecayProvider(), storageScoreStructure, input, creationSlot, targetSlot)
							require.NoError(t, err)

							storedMana, err := testAPI.ManaDecayProvider().ManaWithDecay(iotago.MaxMana, creationSlot, targetSlot)
							require.NoError(t, err)

							// generated mana + decay - allotment
							return potentialMana + storedMana - 50
						}(),
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := transaction.Sign(ident1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok - stored and allotted",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),

		// fail - input created after tx
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDsWithCreationSlot(20, 1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 5,
					Mana:   10,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs:       inputIDs.UTXOInputs(),
					CreationSlot: 15,
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 5,
						Mana:   35,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := transaction.Sign(ident1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - input created after tx",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrInputCreationAfterTxCreation,
			}
		}(),

		// ok - input created in same slot as tx
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDsWithCreationSlot(15, 1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 5,
					Mana:   10,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs:       inputIDs.UTXOInputs(),
					CreationSlot: 15,
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 5,
						Mana:   10,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := transaction.Sign(ident1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok - input created in same slot as tx",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),

		// fail - mana overflow on the input side sum
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDsWithCreationSlot(15, 2)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 5,
					Mana:   iotago.MaxMana,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 5,
					Mana:   10,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs:       inputIDs.UTXOInputs(),
					CreationSlot: 15,
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 5,
						Mana:   9,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := transaction.Sign(ident1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - mana overflow on the input side sum",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.ReferenceUnlock{Reference: 0},
					},
				},
				wantErr: iotago.ErrManaOverflow,
			}
		}(),

		// fail - mana overflow on the output side sum
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDsWithCreationSlot(15, 1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 5,
					Mana:   10,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs:       inputIDs.UTXOInputs(),
					CreationSlot: 15,
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 5,
						Mana:   1,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
					&iotago.BasicOutput{
						Amount: 5,
						Mana:   iotago.MaxMana,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := transaction.Sign(ident1AddrKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - mana overflow on the output side sum",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrManaOverflow,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAndExecuteSignedTransaction(tt.tx, tt.resolvedInputs, vm.ExecFuncBalancedMana())
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestManaRewardsClaimingStaking(t *testing.T) {
	_, ident, identAddrKeys := tpkg.RandEd25519Identity()
	accountIdent := tpkg.RandAccountAddress()
	accountID := accountIdent.AccountID()

	var manaRewardAmount iotago.Mana = 200
	currentEpoch := iotago.EpochIndex(20)
	currentSlot := testAPI.TimeProvider().EpochStart(currentEpoch)

	blockIssuerFeature := &iotago.BlockIssuerFeature{
		BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
		ExpirySlot:      currentSlot + 500,
	}

	var creationSlot iotago.SlotIndex = 100
	balance := OneIOTA * 10

	inputIDs := tpkg.RandOutputIDsWithCreationSlot(creationSlot, 1)
	inputs := vm.InputSet{
		inputIDs[0]: &iotago.AccountOutput{
			Amount:         balance,
			AccountID:      accountID,
			Mana:           0,
			FoundryCounter: 0,
			UnlockConditions: iotago.AccountOutputUnlockConditions{
				&iotago.AddressUnlockCondition{Address: ident},
			},
			Features: iotago.AccountOutputFeatures{
				blockIssuerFeature,
				&iotago.StakingFeature{
					StakedAmount: 100,
					FixedCost:    50,
					StartEpoch:   currentEpoch - 10,
					EndEpoch:     currentEpoch - 1,
				},
			},
		},
	}

	inputMinDeposit := lo.PanicOnErr(testAPI.StorageScoreStructure().MinDeposit(inputs[inputIDs[0]]))

	transaction := &iotago.Transaction{
		API: testAPI,
		TransactionEssence: &iotago.TransactionEssence{
			Inputs:       inputIDs.UTXOInputs(),
			CreationSlot: currentSlot,
		},
		Outputs: iotago.TxEssenceOutputs{
			&iotago.AccountOutput{
				Amount:         OneIOTA * 5,
				Mana:           lo.PanicOnErr(testAPI.ManaDecayProvider().ManaGenerationWithDecay(balance-inputMinDeposit, creationSlot, currentSlot)),
				AccountID:      accountID,
				FoundryCounter: 0,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: ident},
				},
				Features: iotago.AccountOutputFeatures{
					blockIssuerFeature,
				},
			},
			&iotago.BasicOutput{
				Amount: OneIOTA * 5,
				Mana:   manaRewardAmount,
				UnlockConditions: iotago.BasicOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: accountIdent},
				},
				Features: nil,
			},
		},
	}

	sigs, err := transaction.Sign(identAddrKeys)
	require.NoError(t, err)

	tx := &iotago.SignedTransaction{
		API:         testAPI,
		Transaction: transaction,
		Unlocks: iotago.Unlocks{
			&iotago.SignatureUnlock{Signature: sigs[0]},
		},
	}

	resolvedInputs := vm.ResolvedInputs{
		InputSet: inputs,
		RewardsInputSet: map[iotago.ChainID]iotago.Mana{
			accountID: manaRewardAmount,
		},
		CommitmentInput: &iotago.Commitment{
			Slot: currentSlot,
		},
		BlockIssuanceCreditInputSet: vm.BlockIssuanceCreditInputSet{
			accountID: 1000,
		},
	}
	require.NoError(t, validateAndExecuteSignedTransaction(tx, resolvedInputs))
}

func TestManaRewardsClaimingDelegation(t *testing.T) {
	_, ident, identAddrKeys := tpkg.RandEd25519Identity()

	const manaRewardAmount iotago.Mana = 200
	currentSlot := 20 * testProtoParams.ParamEpochDurationInSlots()
	currentEpoch := testAPI.TimeProvider().EpochFromSlot(currentSlot)

	inputIDs := tpkg.RandOutputIDs(1)
	inputs := vm.InputSet{
		inputIDs[0]: &iotago.DelegationOutput{
			Amount:           OneIOTA * 10,
			DelegatedAmount:  OneIOTA * 10,
			DelegationID:     iotago.EmptyDelegationID(),
			ValidatorAddress: &iotago.AccountAddress{},
			StartEpoch:       currentEpoch,
			EndEpoch:         currentEpoch + 5,
			UnlockConditions: iotago.DelegationOutputUnlockConditions{
				&iotago.AddressUnlockCondition{Address: ident},
			},
		},
	}
	delegationID := iotago.DelegationIDFromOutputID(inputIDs[0])

	transaction := &iotago.Transaction{
		API: testAPI,
		TransactionEssence: &iotago.TransactionEssence{
			Inputs:       inputIDs.UTXOInputs(),
			CreationSlot: currentSlot,
			Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
		},
		Outputs: iotago.TxEssenceOutputs{
			&iotago.BasicOutput{
				Amount: OneIOTA * 10,
				Mana:   manaRewardAmount,
				UnlockConditions: iotago.BasicOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: ident},
				},
				Features: nil,
			},
		},
	}

	sigs, err := transaction.Sign(identAddrKeys)
	require.NoError(t, err)

	tx := &iotago.SignedTransaction{
		API:         testAPI,
		Transaction: transaction,
		Unlocks: iotago.Unlocks{
			&iotago.SignatureUnlock{Signature: sigs[0]},
		},
	}

	resolvedInputs := vm.ResolvedInputs{
		InputSet: inputs,
		RewardsInputSet: map[iotago.ChainID]iotago.Mana{
			delegationID: manaRewardAmount,
		},
	}
	require.NoError(t, validateAndExecuteSignedTransaction(tx, resolvedInputs))
}

func TestTxSyntacticAddressRestrictions(t *testing.T) {
	type testParameters struct {
		name    string
		address iotago.Address
		wantErr error
	}
	type test struct {
		createTestOutput     func(address iotago.Address) iotago.Output
		createTestParameters []func() testParameters
	}

	_, ident, identAddrKeys := tpkg.RandEd25519Identity()
	addr := tpkg.RandEd25519Address()

	tests := []*test{
		{
			createTestOutput: func(address iotago.Address) iotago.Output {
				return &iotago.BasicOutput{
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: address},
					},
					Features: iotago.BasicOutputFeatures{
						tpkg.RandNativeTokenFeature(),
					},
				}
			},
			createTestParameters: []func() testParameters{
				func() testParameters {
					return testParameters{
						name:    "ok - Native Token Address in Output with Native Tokens",
						address: iotago.RestrictedAddressWithCapabilities(addr, iotago.WithAddressCanReceiveNativeTokens(true)),
						wantErr: nil,
					}
				},
				func() testParameters {
					return testParameters{
						name:    "fail - Non Native Token Address in Output with Native Tokens",
						address: iotago.RestrictedAddressWithCapabilities(addr),
						wantErr: iotago.ErrAddressCannotReceiveNativeTokens,
					}
				},
			},
		},
		{
			createTestOutput: func(address iotago.Address) iotago.Output {
				return &iotago.BasicOutput{
					Mana: iotago.Mana(4),
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: address},
					},
				}
			},
			createTestParameters: []func() testParameters{
				func() testParameters {
					return testParameters{
						name:    "ok - Mana Address in Output with Mana",
						address: iotago.RestrictedAddressWithCapabilities(addr, iotago.WithAddressCanReceiveMana(true)),
						wantErr: nil,
					}
				},
				func() testParameters {
					return testParameters{
						name:    "fail - Non Mana Address in Output with Mana",
						address: iotago.RestrictedAddressWithCapabilities(addr),
						wantErr: iotago.ErrAddressCannotReceiveMana,
					}
				},
			},
		},
		{
			createTestOutput: func(address iotago.Address) iotago.Output {
				return &iotago.BasicOutput{
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: address},
						&iotago.TimelockUnlockCondition{
							Slot: 500,
						},
					},
				}
			},
			createTestParameters: []func() testParameters{
				func() testParameters {
					return testParameters{
						name:    "ok - Timelock Unlock Condition Address in Output with Timelock Unlock Condition",
						address: iotago.RestrictedAddressWithCapabilities(addr, iotago.WithAddressCanReceiveOutputsWithTimelockUnlockCondition(true)),
						wantErr: nil,
					}
				},
				func() testParameters {
					return testParameters{
						name:    "fail - Non Timelock Unlock Condition Address in Output with Timelock Unlock Condition",
						address: iotago.RestrictedAddressWithCapabilities(addr),
						wantErr: iotago.ErrAddressCannotReceiveTimelockUnlockCondition,
					}
				},
			},
		},
		{
			createTestOutput: func(address iotago.Address) iotago.Output {
				return &iotago.BasicOutput{
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: address},
						&iotago.ExpirationUnlockCondition{
							Slot:          500,
							ReturnAddress: ident,
						},
					},
				}
			},
			createTestParameters: []func() testParameters{
				func() testParameters {
					return testParameters{
						name:    "ok - Expiration Unlock Condition Address in Output with Expiration Unlock Condition",
						address: iotago.RestrictedAddressWithCapabilities(addr, iotago.WithAddressCanReceiveOutputsWithExpirationUnlockCondition(true)),
						wantErr: nil,
					}
				},
				func() testParameters {
					return testParameters{
						name:    "fail - Non Expiration Unlock Condition Address in Output with Expiration Unlock Condition",
						address: iotago.RestrictedAddressWithCapabilities(addr),
						wantErr: iotago.ErrAddressCannotReceiveExpirationUnlockCondition,
					}
				},
			},
		},
		{
			createTestOutput: func(address iotago.Address) iotago.Output {
				return &iotago.BasicOutput{
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: address},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: ident,
						},
					},
				}
			},
			createTestParameters: []func() testParameters{
				func() testParameters {
					return testParameters{
						name:    "ok - Storage Deposit Return Unlock Condition Address in Output with Storage Deposit Return Unlock Condition",
						address: iotago.RestrictedAddressWithCapabilities(addr, iotago.WithAddressCanReceiveOutputsWithStorageDepositReturnUnlockCondition(true)),
						wantErr: nil,
					}
				},
				func() testParameters {
					return testParameters{
						name:    "fail - Non Storage Deposit Return Unlock Condition Address in Output with Storage Deposit Return Unlock Condition",
						address: iotago.RestrictedAddressWithCapabilities(addr),
						wantErr: iotago.ErrAddressCannotReceiveStorageDepositReturnUnlockCondition,
					}
				},
			},
		},
		{
			createTestOutput: func(address iotago.Address) iotago.Output {
				return &iotago.AccountOutput{
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: address},
					},
				}
			},
			createTestParameters: []func() testParameters{
				func() testParameters {
					return testParameters{
						name:    "ok - Account Output Address in Account Output",
						address: iotago.RestrictedAddressWithCapabilities(addr, iotago.WithAddressCanReceiveAccountOutputs(true)),
						wantErr: nil,
					}
				},
				func() testParameters {
					return testParameters{
						name:    "fail - Non Account Output Address in Account Output",
						address: iotago.RestrictedAddressWithCapabilities(addr),
						wantErr: iotago.ErrAddressCannotReceiveAccountOutput,
					}
				},
			},
		},
		{
			createTestOutput: func(address iotago.Address) iotago.Output {
				return &iotago.AnchorOutput{
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: address},
					},
				}
			},
			createTestParameters: []func() testParameters{
				func() testParameters {
					return testParameters{
						name:    "ok - Anchor Output Address in State Controller UC in Anchor Output",
						address: iotago.RestrictedAddressWithCapabilities(addr, iotago.WithAddressCanReceiveAnchorOutputs(true)),
						wantErr: nil,
					}
				},
				func() testParameters {
					return testParameters{
						name:    "fail - Non Anchor Output Address in State Controller UC in Anchor Output",
						address: iotago.RestrictedAddressWithCapabilities(addr),
						wantErr: iotago.ErrAddressCannotReceiveAnchorOutput,
					}
				},
			},
		},
		{
			createTestOutput: func(address iotago.Address) iotago.Output {
				return &iotago.AnchorOutput{
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.GovernorAddressUnlockCondition{Address: address},
					},
				}
			},
			createTestParameters: []func() testParameters{
				func() testParameters {
					return testParameters{
						name:    "ok - Anchor Output Address in Governor UC in Anchor Output",
						address: iotago.RestrictedAddressWithCapabilities(addr, iotago.WithAddressCanReceiveAnchorOutputs(true)),
						wantErr: nil,
					}
				},
				func() testParameters {
					return testParameters{
						name:    "fail - Non Anchor Output Address in Governor UC in Anchor Output",
						address: iotago.RestrictedAddressWithCapabilities(addr),
						wantErr: iotago.ErrAddressCannotReceiveAnchorOutput,
					}
				},
			},
		},
		{
			createTestOutput: func(address iotago.Address) iotago.Output {
				return &iotago.NFTOutput{
					UnlockConditions: iotago.NFTOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: address},
					},
				}
			},
			createTestParameters: []func() testParameters{
				func() testParameters {
					return testParameters{
						name:    "ok - NFT Output Address in NFT Output",
						address: iotago.RestrictedAddressWithCapabilities(addr, iotago.WithAddressCanReceiveNFTOutputs(true)),
						wantErr: nil,
					}
				},
				func() testParameters {
					return testParameters{
						name:    "fail - Non NFT Output Address in NFT Output",
						address: iotago.RestrictedAddressWithCapabilities(addr),
						wantErr: iotago.ErrAddressCannotReceiveNFTOutput,
					}
				},
			},
		},
		{
			createTestOutput: func(address iotago.Address) iotago.Output {
				return &iotago.DelegationOutput{
					UnlockConditions: iotago.DelegationOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: address},
					},
					ValidatorAddress: &iotago.AccountAddress{},
				}
			},
			createTestParameters: []func() testParameters{
				func() testParameters {
					return testParameters{
						name:    "ok - Delegation Output Address in Delegation Output",
						address: iotago.RestrictedAddressWithCapabilities(addr, iotago.WithAddressCanReceiveDelegationOutputs(true)),
						wantErr: nil,
					}
				},
				func() testParameters {
					return testParameters{
						name:    "fail - Non Delegation Output Address in Delegation Output",
						address: iotago.RestrictedAddressWithCapabilities(addr),
						wantErr: iotago.ErrAddressCannotReceiveDelegationOutput,
					}
				},
			},
		},
		{
			createTestOutput: func(address iotago.Address) iotago.Output {
				return &iotago.BasicOutput{
					Mana: 42,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						&iotago.ExpirationUnlockCondition{
							// only the return address is restricted here
							ReturnAddress: address,
						},
					},
				}
			},
			createTestParameters: []func() testParameters{
				func() testParameters {
					return testParameters{
						name:    "ok - Mana Return Address in Output with Mana",
						address: iotago.RestrictedAddressWithCapabilities(addr, iotago.WithAddressCanReceiveMana(true), iotago.WithAddressCanReceiveOutputsWithExpirationUnlockCondition(true)),
						wantErr: nil,
					}
				},
				func() testParameters {
					return testParameters{
						name:    "fail - Non Mana Return Address in Output with Mana",
						address: iotago.RestrictedAddressWithCapabilities(addr),
						wantErr: iotago.ErrAddressCannotReceiveMana,
					}
				},
			},
		},
	}

	makeTransaction := func(output iotago.Output) (vm.InputSet, iotago.Signature, *iotago.Transaction) {
		inputIDs := tpkg.RandOutputIDsWithCreationSlot(10, 1)

		inputs := vm.InputSet{
			inputIDs[0]: &iotago.BasicOutput{
				Amount: iotago.BaseToken(1_000_000),
				UnlockConditions: iotago.BasicOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: ident},
				},
			},
		}

		transaction := &iotago.Transaction{
			API: testAPI,
			TransactionEssence: &iotago.TransactionEssence{
				NetworkID:    testAPI.ProtocolParameters().NetworkID(),
				Inputs:       inputIDs.UTXOInputs(),
				CreationSlot: 10,
			},
			Outputs: iotago.TxEssenceOutputs{
				output,
			},
		}
		sigs, err := transaction.Sign(identAddrKeys)
		require.NoError(t, err)

		return inputs, sigs[0], transaction
	}

	for _, tt := range tests {
		for _, makeTestInput := range tt.createTestParameters {
			testInput := makeTestInput()
			t.Run(testInput.name, func(t *testing.T) {
				testOutput := tt.createTestOutput(testInput.address)

				_, sig, transaction := makeTransaction(testOutput)

				tx := &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sig},
					},
				}

				addressRestrictionFunc := iotago.OutputsSyntacticalAddressRestrictions()

				for index, output := range tx.Transaction.Outputs {
					err := addressRestrictionFunc(index, output)

					if testInput.wantErr != nil {
						require.ErrorIs(t, err, testInput.wantErr)
						return
					}

					require.NoError(t, err)
				}
			})
		}
	}
}

func TestTxSemanticImplicitAccountCreationAndTransition(t *testing.T) {
	type TestInput struct {
		inputID      iotago.OutputID
		input        iotago.Output
		unlockTarget iotago.Address
	}

	type test struct {
		name                    string
		inputs                  []TestInput
		keys                    []iotago.AddressKeys
		resolvedCommitmentInput iotago.Commitment
		resolvedBICInputSet     vm.BlockIssuanceCreditInputSet
		outputs                 []iotago.Output
		wantErr                 error
	}

	_, edIdent, edIdentAddrKeys := tpkg.RandEd25519Identity()
	_, implicitAccountIdent, implicitAccountIdentAddrKeys := tpkg.RandImplicitAccountIdentity()
	exampleAmount := iotago.BaseToken(1_000_000)
	exampleMana := iotago.Mana(10_000_000)
	exampleNativeTokenFeature := tpkg.RandNativeTokenFeature()
	outputID1 := tpkg.RandOutputID(0)
	outputID2 := tpkg.RandOutputID(1)
	accountID1 := iotago.AccountIDFromOutputID(outputID1)
	accountID2 := iotago.AccountIDFromOutputID(outputID2)
	currentSlot := iotago.SlotIndex(10)
	commitmentSlot := currentSlot - testAPI.ProtocolParameters().MaxCommittableAge()

	dummyImplicitAccount := &iotago.BasicOutput{
		Amount: 0,
		UnlockConditions: iotago.BasicOutputUnlockConditions{
			&iotago.AddressUnlockCondition{Address: implicitAccountIdent},
		},
	}
	exampleMetadataFeature := &iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": tpkg.RandBytes(40)}}
	exampleMetadataFeatureStorageDeposit := iotago.BaseToken(exampleMetadataFeature.Size()*int(testAPI.StorageScoreStructure().FactorData())) * testAPI.StorageScoreStructure().StorageCost()

	storageScore := dummyImplicitAccount.StorageScore(testAPI.StorageScoreStructure(), nil)
	minAmountImplicitAccount := testAPI.StorageScoreStructure().StorageCost() * iotago.BaseToken(storageScore)

	exampleInputs := []TestInput{
		{
			inputID: outputID1,
			input: &iotago.BasicOutput{
				Amount: exampleAmount,
				UnlockConditions: iotago.BasicOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: edIdent},
				},
				Features: iotago.BasicOutputFeatures{
					exampleNativeTokenFeature,
				},
			},
			unlockTarget: edIdent,
		},
	}

	tests := []*test{
		{
			name:   "ok - implicit account creation",
			inputs: exampleInputs,
			outputs: []iotago.Output{
				&iotago.BasicOutput{
					Amount: exampleAmount,
					Mana:   0,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: implicitAccountIdent},
					},
				},
			},
			keys:    []iotago.AddressKeys{edIdentAddrKeys},
			wantErr: nil,
		},
		{
			name:   "fail - implicit account contains timelock unlock conditions",
			inputs: exampleInputs,
			outputs: []iotago.Output{
				&iotago.BasicOutput{
					Amount: exampleAmount,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: implicitAccountIdent},
						&iotago.TimelockUnlockCondition{Slot: 500},
					},
				},
			},
			keys:    []iotago.AddressKeys{edIdentAddrKeys},
			wantErr: iotago.ErrAddressCannotReceiveTimelockUnlockCondition,
		},
		{
			name:   "fail - implicit account contains expiration unlock conditions",
			inputs: exampleInputs,
			outputs: []iotago.Output{
				&iotago.BasicOutput{
					Amount: exampleAmount,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: implicitAccountIdent},
						&iotago.ExpirationUnlockCondition{
							// The implicit account creation address should disallow this expiration UC.
							ReturnAddress: tpkg.RandEd25519Address(),
							Slot:          500,
						},
					},
				},
			},
			keys:    []iotago.AddressKeys{edIdentAddrKeys},
			wantErr: iotago.ErrAddressCannotReceiveExpirationUnlockCondition,
		},
		{
			name:   "fail - implicit account contains storage deposit return unlock conditions",
			inputs: exampleInputs,
			outputs: []iotago.Output{
				&iotago.BasicOutput{
					Amount: exampleAmount,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: implicitAccountIdent},
						&iotago.StorageDepositReturnUnlockCondition{
							// The implicit account creation address should disallow this SDRUC.
							ReturnAddress: tpkg.RandEd25519Address(),
							Amount:        20_000,
						},
					},
				},
			},
			keys:    []iotago.AddressKeys{edIdentAddrKeys},
			wantErr: iotago.ErrAddressCannotReceiveStorageDepositReturnUnlockCondition,
		},
		{
			name:   "ok - implicit account contains features",
			inputs: exampleInputs,
			outputs: []iotago.Output{
				&iotago.BasicOutput{
					Amount: exampleAmount,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: implicitAccountIdent},
					},
					Features: iotago.BasicOutputFeatures{
						&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": tpkg.RandBytes(40)}},
						&iotago.TagFeature{
							Tag: tpkg.RandBytes(12),
						},
						&iotago.SenderFeature{
							Address: edIdentAddrKeys.Address,
						},
						&iotago.NativeTokenFeature{
							ID:     exampleNativeTokenFeature.ID,
							Amount: exampleNativeTokenFeature.Amount,
						},
					},
				},
			},
			keys:    []iotago.AddressKeys{edIdentAddrKeys},
			wantErr: nil,
		},
		{
			name: "ok - implicit account transitioned to account with block issuer feature",
			inputs: []TestInput{
				{
					inputID: outputID1,
					input: &iotago.BasicOutput{
						Amount: exampleAmount,
						Mana:   exampleMana,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: implicitAccountIdent},
						},
					},
					unlockTarget: implicitAccountIdent,
				},
			},
			resolvedBICInputSet: vm.BlockIssuanceCreditInputSet{
				accountID1: iotago.BlockIssuanceCredits(0),
			},
			resolvedCommitmentInput: iotago.Commitment{
				Slot: commitmentSlot,
			},
			outputs: []iotago.Output{
				&iotago.AccountOutput{
					Amount:    exampleAmount,
					Mana:      exampleMana,
					AccountID: accountID1,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{
							Address: edIdent,
						},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							ExpirySlot: iotago.MaxSlotIndex,
							BlockIssuerKeys: iotago.NewBlockIssuerKeys(
								iotago.Ed25519PublicKeyBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
								iotago.Ed25519PublicKeyBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
								iotago.Ed25519PublicKeyBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
							),
						},
					},
				},
			},
			keys:    []iotago.AddressKeys{implicitAccountIdentAddrKeys},
			wantErr: nil,
		},
		{
			name: "ok - implicit account with native tokens transitioned to account with block issuer feature",
			inputs: []TestInput{
				{
					inputID: outputID1,
					input: &iotago.BasicOutput{
						Amount: exampleAmount,
						Mana:   exampleMana,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: implicitAccountIdent},
						},
						Features: iotago.BasicOutputFeatures{
							exampleNativeTokenFeature,
						},
					},
					unlockTarget: implicitAccountIdent,
				},
			},
			resolvedBICInputSet: vm.BlockIssuanceCreditInputSet{
				accountID1: iotago.BlockIssuanceCredits(0),
			},
			resolvedCommitmentInput: iotago.Commitment{
				Slot: commitmentSlot,
			},
			outputs: []iotago.Output{
				&iotago.AccountOutput{
					Amount:    exampleAmount,
					Mana:      exampleMana,
					AccountID: accountID1,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{
							Address: edIdent,
						},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							ExpirySlot: iotago.MaxSlotIndex,
							BlockIssuerKeys: iotago.NewBlockIssuerKeys(
								iotago.Ed25519PublicKeyBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
								iotago.Ed25519PublicKeyBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
								iotago.Ed25519PublicKeyBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
							),
						},
						exampleNativeTokenFeature,
					},
				},
			},
			keys:    []iotago.AddressKeys{implicitAccountIdentAddrKeys},
			wantErr: nil,
		},
		{
			name: "fail - implicit account transitioned to account without block issuer feature",
			inputs: []TestInput{
				{
					inputID: outputID1,
					input: &iotago.BasicOutput{
						Amount: exampleAmount,
						Mana:   exampleMana,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: implicitAccountIdent},
						},
					},
					unlockTarget: implicitAccountIdent,
				},
			},
			resolvedBICInputSet: vm.BlockIssuanceCreditInputSet{
				accountID1: iotago.BlockIssuanceCredits(0),
			},
			resolvedCommitmentInput: iotago.Commitment{
				Slot: commitmentSlot,
			},
			outputs: []iotago.Output{
				&iotago.AccountOutput{
					Amount:    exampleAmount,
					Mana:      exampleMana,
					AccountID: accountID1,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{
							Address: edIdent,
						},
					},
					Features: iotago.AccountOutputFeatures{},
				},
			},
			keys:    []iotago.AddressKeys{implicitAccountIdentAddrKeys},
			wantErr: iotago.ErrInvalidBlockIssuerTransition,
		},
		{
			name: "fail - attempt to destroy implicit account",
			inputs: []TestInput{
				{
					inputID: outputID1,
					input: &iotago.BasicOutput{
						Amount: exampleAmount,
						Mana:   exampleMana,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: implicitAccountIdent},
						},
					},
					unlockTarget: implicitAccountIdent,
				},
			},
			resolvedBICInputSet: vm.BlockIssuanceCreditInputSet{
				accountID1: iotago.BlockIssuanceCredits(0),
			},
			resolvedCommitmentInput: iotago.Commitment{
				Slot: commitmentSlot,
			},
			outputs: []iotago.Output{
				&iotago.BasicOutput{
					Amount: exampleAmount,
					Mana:   exampleMana,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: edIdent},
					},
				},
			},
			keys:    []iotago.AddressKeys{implicitAccountIdentAddrKeys},
			wantErr: iotago.ErrImplicitAccountDestructionDisallowed,
		},
		{
			name: "ok - implicit account with OffsetImplicitAccountCreationAddress can be transitioned",
			inputs: []TestInput{
				{
					inputID: outputID1,
					input: &iotago.BasicOutput{
						Amount: minAmountImplicitAccount,
						Mana:   0,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: implicitAccountIdent},
						},
					},
					unlockTarget: implicitAccountIdent,
				},
			},
			resolvedBICInputSet: vm.BlockIssuanceCreditInputSet{
				accountID1: iotago.BlockIssuanceCredits(0),
			},
			resolvedCommitmentInput: iotago.Commitment{
				Slot: commitmentSlot,
			},
			outputs: []iotago.Output{
				&iotago.AccountOutput{
					Amount:    minAmountImplicitAccount,
					Mana:      0,
					AccountID: accountID1,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{
							Address: edIdent,
						},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							ExpirySlot: iotago.MaxSlotIndex,
							BlockIssuerKeys: iotago.NewBlockIssuerKeys(
								iotago.Ed25519PublicKeyBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
							),
						},
					},
				},
			},
			keys:    []iotago.AddressKeys{implicitAccountIdentAddrKeys},
			wantErr: nil,
		},
		{
			name: "ok - implicit account with minimal amount and metadata feat can be transitioned",
			inputs: []TestInput{
				{
					inputID: outputID1,
					input: &iotago.BasicOutput{
						Amount: minAmountImplicitAccount + exampleMetadataFeatureStorageDeposit,
						Mana:   0,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: implicitAccountIdent},
						},
						Features: iotago.BasicOutputFeatures{
							exampleMetadataFeature,
						},
					},
					unlockTarget: implicitAccountIdent,
				},
			},
			resolvedBICInputSet: vm.BlockIssuanceCreditInputSet{
				accountID1: iotago.BlockIssuanceCredits(0),
			},
			resolvedCommitmentInput: iotago.Commitment{
				Slot: commitmentSlot,
			},
			outputs: []iotago.Output{
				&iotago.AccountOutput{
					Amount:    minAmountImplicitAccount + exampleMetadataFeatureStorageDeposit,
					Mana:      0,
					AccountID: accountID1,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{
							Address: edIdent,
						},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							ExpirySlot: iotago.MaxSlotIndex,
							BlockIssuerKeys: iotago.NewBlockIssuerKeys(
								iotago.Ed25519PublicKeyBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
							),
						},
						exampleMetadataFeature,
					},
				},
			},
			keys:    []iotago.AddressKeys{implicitAccountIdentAddrKeys},
			wantErr: nil,
		},
		{
			name: "ok - implicit account conversion transaction can contain other non-implicit-account outputs",
			inputs: []TestInput{
				{
					inputID: outputID1,
					input: &iotago.BasicOutput{
						Amount: minAmountImplicitAccount,
						Mana:   0,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: implicitAccountIdent},
						},
					},
					unlockTarget: implicitAccountIdent,
				},
				{
					inputID: tpkg.RandOutputID(1),
					input: &iotago.BasicOutput{
						Amount: exampleAmount,
						Mana:   0,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: edIdent},
						},
					},
					unlockTarget: edIdent,
				},
			},
			resolvedBICInputSet: vm.BlockIssuanceCreditInputSet{
				accountID1: iotago.BlockIssuanceCredits(0),
			},
			resolvedCommitmentInput: iotago.Commitment{
				Slot: commitmentSlot,
			},
			outputs: []iotago.Output{
				&iotago.AccountOutput{
					// Fund new account with additional base tokens from another output.
					Amount:    minAmountImplicitAccount + exampleAmount,
					Mana:      0,
					AccountID: accountID1,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{
							Address: edIdent,
						},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							ExpirySlot: iotago.MaxSlotIndex,
							BlockIssuerKeys: iotago.NewBlockIssuerKeys(
								iotago.Ed25519PublicKeyBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
							),
						},
					},
				},
			},
			keys:    []iotago.AddressKeys{implicitAccountIdentAddrKeys, edIdentAddrKeys},
			wantErr: nil,
		},
		{
			name: "fail - transaction contains more than one implicit account on the input side",
			inputs: []TestInput{
				{
					inputID: outputID1,
					input: &iotago.BasicOutput{
						Amount: minAmountImplicitAccount,
						Mana:   0,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: implicitAccountIdent},
						},
					},
					unlockTarget: implicitAccountIdent,
				},
				{
					inputID: outputID2,
					input: &iotago.BasicOutput{
						Amount: exampleAmount,
						Mana:   0,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: implicitAccountIdent},
						},
					},
					unlockTarget: implicitAccountIdent,
				},
			},
			resolvedBICInputSet: vm.BlockIssuanceCreditInputSet{
				accountID1: iotago.BlockIssuanceCredits(0),
				accountID2: iotago.BlockIssuanceCredits(0),
			},
			resolvedCommitmentInput: iotago.Commitment{
				Slot: commitmentSlot,
			},
			outputs: []iotago.Output{
				&iotago.AccountOutput{
					Amount:    minAmountImplicitAccount,
					Mana:      0,
					AccountID: accountID1,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{
							Address: edIdent,
						},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							ExpirySlot: iotago.MaxSlotIndex,
							BlockIssuerKeys: iotago.NewBlockIssuerKeys(
								iotago.Ed25519PublicKeyBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
							),
						},
					},
				},
				&iotago.AccountOutput{
					Amount:    exampleAmount,
					Mana:      0,
					AccountID: accountID2,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{
							Address: edIdent,
						},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							ExpirySlot: iotago.MaxSlotIndex,
							BlockIssuerKeys: iotago.NewBlockIssuerKeys(
								iotago.Ed25519PublicKeyBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
							),
						},
					},
				},
			},
			keys:    []iotago.AddressKeys{implicitAccountIdentAddrKeys, edIdentAddrKeys},
			wantErr: iotago.ErrMultipleImplicitAccountCreationAddresses,
		},
		{
			name: "fail - transaction moves mana off an implicit account",
			inputs: []TestInput{
				{
					inputID: outputID1,
					input: &iotago.BasicOutput{
						Amount: exampleAmount,
						Mana:   exampleMana,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: implicitAccountIdent},
						},
					},
					unlockTarget: implicitAccountIdent,
				},
			},
			resolvedBICInputSet: vm.BlockIssuanceCreditInputSet{
				accountID1: iotago.BlockIssuanceCredits(0),
			},
			resolvedCommitmentInput: iotago.Commitment{
				Slot: commitmentSlot,
			},
			outputs: []iotago.Output{
				&iotago.AccountOutput{
					Amount:    minAmountImplicitAccount,
					Mana:      exampleMana / 2,
					AccountID: accountID1,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{
							Address: edIdent,
						},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							ExpirySlot: iotago.MaxSlotIndex,
							BlockIssuerKeys: iotago.NewBlockIssuerKeys(
								iotago.Ed25519PublicKeyBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray()),
							),
						},
					},
				},
				&iotago.BasicOutput{
					Amount: exampleAmount - minAmountImplicitAccount,
					Mana:   exampleMana / 2,
					UnlockConditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{
							Address: edIdent,
						},
					},
				},
			},
			keys:    []iotago.AddressKeys{implicitAccountIdentAddrKeys},
			wantErr: iotago.ErrInvalidBlockIssuerTransition,
		},
	}

	for idx, tt := range tests {
		resolvedInputs := vm.ResolvedInputs{
			InputSet: vm.InputSet{},
		}

		txBuilder := builder.NewTransactionBuilder(testAPI)
		txBuilder.WithTransactionCapabilities(
			iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanBurnNativeTokens(true)),
		)

		for _, input := range tests[idx].inputs {
			txBuilder.AddInput(&builder.TxInput{
				UnlockTarget: input.unlockTarget,
				InputID:      input.inputID,
				Input:        input.input,
			},
			)

			resolvedInputs.InputSet[input.inputID] = input.input
		}

		for _, output := range tests[idx].outputs {
			txBuilder.AddOutput(output)
		}
		tx := lo.PanicOnErr(txBuilder.Build(iotago.NewInMemoryAddressSigner(tt.keys...)))

		resolvedInputs.BlockIssuanceCreditInputSet = tests[idx].resolvedBICInputSet
		resolvedInputs.CommitmentInput = &tests[idx].resolvedCommitmentInput

		t.Run(tt.name, func(t *testing.T) {
			var err error
			// Some constraints are implicitly tested as part of the address restrictions, which are syntactic checks.
			err = tx.Transaction.SyntacticallyValidate(tx.API)
			if err == nil {
				err = validateAndExecuteSignedTransaction(tx, resolvedInputs)
			}

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

// Ensure that the storage score offset for implicit accounts is the
// minimum required for a full block issuer account.
func TestTxSyntacticImplicitAccountMinDeposit(t *testing.T) {
	_, implicitAccountIdent, _ := tpkg.RandImplicitAccountIdentity()

	implicitAccount := &iotago.BasicOutput{
		Amount: 0,
		UnlockConditions: iotago.BasicOutputUnlockConditions{
			&iotago.AddressUnlockCondition{Address: implicitAccountIdent},
		},
	}
	storageScore := implicitAccount.StorageScore(testAPI.StorageScoreStructure(), nil)
	minAmount := testAPI.StorageScoreStructure().StorageCost() * iotago.BaseToken(storageScore)
	implicitAccount.Amount = minAmount
	depositValidationFunc := iotago.OutputsSyntacticalDepositAmount(testAPI.ProtocolParameters(), testAPI.StorageScoreStructure())
	require.NoError(t, depositValidationFunc(0, implicitAccount))

	convertedAccount := &iotago.AccountOutput{
		Amount: implicitAccount.Amount,
		UnlockConditions: iotago.AccountOutputUnlockConditions{
			&iotago.AddressUnlockCondition{
				Address: &iotago.Ed25519Address{},
			},
		},
		Features: iotago.AccountOutputFeatures{
			&iotago.BlockIssuerFeature{
				BlockIssuerKeys: iotago.BlockIssuerKeys{
					&iotago.Ed25519PublicKeyBlockIssuerKey{},
				},
			},
		},
	}

	require.NoError(t, depositValidationFunc(0, convertedAccount))
}

func validateAndExecuteSignedTransaction(tx *iotago.SignedTransaction, resolvedInputs vm.ResolvedInputs, execFunctions ...vm.ExecFunc) (err error) {
	unlockedIdentities, err := novaVM.ValidateUnlocks(tx, resolvedInputs)
	if err != nil {
		return err
	}

	return lo.Return2(novaVM.Execute(tx.Transaction, resolvedInputs, unlockedIdentities, execFunctions...))
}
