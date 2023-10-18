//nolint:forcetypeassert,dupl,nlreturn,scopelint
package nova_test

import (
	"bytes"
	"crypto/ed25519"
	"math/big"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/safemath"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/builder"
	"github.com/iotaledger/iota.go/v4/tpkg"
	"github.com/iotaledger/iota.go/v4/vm"
	"github.com/iotaledger/iota.go/v4/vm/nova"
)

const (
	OneMi = 1_000_000

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
		iotago.WithTimeProviderOptions(100, slotDurationSeconds, slotsPerEpochExponent),
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
			Amount: OneMi,
			NFTID:  iotago.NFTID{},
			Conditions: iotago.NFTOutputUnlockConditions{
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
				Amount: OneMi,
				NFTID:  nftID,
				Conditions: iotago.NFTOutputUnlockConditions{
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
			Amount: OneMi,
			Conditions: iotago.BasicOutputUnlockConditions{
				&iotago.AddressUnlockCondition{Address: ident1},
			},
		},
		inputIDs[1]: &iotago.AccountOutput{
			Amount:         OneMi,
			AccountID:      accountIdent1.AccountID(),
			StateIndex:     1,
			StateMetadata:  nil,
			FoundryCounter: 1,
			Conditions: iotago.AccountOutputUnlockConditions{
				&iotago.StateControllerAddressUnlockCondition{Address: ident1},
				&iotago.GovernorAddressUnlockCondition{Address: ident1},
			},
			Features: nil,
		},
		inputIDs[2]: &iotago.FoundryOutput{
			Amount:       OneMi,
			SerialNumber: 1,
			TokenScheme: &iotago.SimpleTokenScheme{
				MintedTokens:  big.NewInt(50),
				MeltedTokens:  big.NewInt(0),
				MaximumSupply: big.NewInt(50),
			},
			Conditions: iotago.FoundryOutputUnlockConditions{
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
				Amount:         OneMi,
				AccountID:      accountIdent1.AccountID(),
				StateIndex:     2,
				StateMetadata:  nil,
				FoundryCounter: 1,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: ident1},
					&iotago.GovernorAddressUnlockCondition{Address: ident1},
				},
				Features: nil,
			},
			&iotago.FoundryOutput{
				Amount:       2 * OneMi,
				SerialNumber: 1,
				TokenScheme: &iotago.SimpleTokenScheme{
					MintedTokens:  big.NewInt(50),
					MeltedTokens:  big.NewInt(50),
					MaximumSupply: big.NewInt(50),
				},
				Conditions: iotago.FoundryOutputUnlockConditions{
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
		func() *test {
			var (
				_, ident1, ident1AddrKeys = tpkg.RandEd25519Identity()
				_, ident2, ident2AddrKeys = tpkg.RandEd25519Identity()
				_, ident3, ident3AddrKeys = tpkg.RandEd25519Identity()
				_, ident4, ident4AddrKeys = tpkg.RandEd25519Identity()
				_, ident5, _              = tpkg.RandEd25519Identity()
			)

			var (
				defaultAmount        iotago.BaseToken = OneMi
				storageDepositReturn iotago.BaseToken = OneMi / 2
				nativeTokenTransfer1                  = tpkg.RandNativeTokenFeature()
				nativeTokenTransfer2                  = tpkg.RandNativeTokenFeature()
			)

			var (
				nft1ID = tpkg.Rand32ByteArray()
				nft2ID = tpkg.Rand32ByteArray()
			)

			inputIDs := tpkg.RandOutputIDs(16)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: defaultAmount,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: defaultAmount,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident2},
					},
					Features: iotago.BasicOutputFeatures{
						nativeTokenTransfer1,
					},
				},
				inputIDs[2]: &iotago.BasicOutput{
					Amount: defaultAmount,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident2},
					},
					Features: iotago.BasicOutputFeatures{
						nativeTokenTransfer2,
					},
				},
				inputIDs[3]: &iotago.BasicOutput{
					Amount: defaultAmount,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident2},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: ident1,
							Slot:          500,
						},
					},
				},
				inputIDs[4]: &iotago.BasicOutput{
					Amount: defaultAmount,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident2},
						&iotago.TimelockUnlockCondition{
							Slot: 500,
						},
					},
				},
				inputIDs[5]: &iotago.BasicOutput{
					Amount: defaultAmount + storageDepositReturn,
					Conditions: iotago.BasicOutputUnlockConditions{
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
				inputIDs[6]: &iotago.AccountOutput{
					Amount:         defaultAmount,
					AccountID:      iotago.AccountID{},
					StateIndex:     0,
					StateMetadata:  []byte("gov transitioning"),
					FoundryCounter: 0,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident3},
						&iotago.GovernorAddressUnlockCondition{Address: ident4},
					},
					Features: nil,
				},
				inputIDs[7]: &iotago.AccountOutput{
					Amount:         defaultAmount + defaultAmount, // to fund also the new account output
					AccountID:      iotago.AccountID{},
					StateIndex:     5,
					StateMetadata:  []byte("current state"),
					FoundryCounter: 5,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident3},
						&iotago.GovernorAddressUnlockCondition{Address: ident4},
					},
					Features: nil,
				},
				inputIDs[8]: &iotago.AccountOutput{
					Amount:         defaultAmount,
					AccountID:      iotago.AccountID{},
					StateIndex:     0,
					StateMetadata:  []byte("going to be destroyed"),
					FoundryCounter: 0,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident3},
						&iotago.GovernorAddressUnlockCondition{Address: ident3},
					},
					Features: nil,
				},
				inputIDs[9]: &iotago.FoundryOutput{
					Amount:       defaultAmount,
					SerialNumber: 1,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(100),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
					Conditions: iotago.FoundryOutputUnlockConditions{
						&iotago.ImmutableAccountUnlockCondition{Address: iotago.AccountIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AccountAddress)},
					},
					Features: nil,
				},
				inputIDs[10]: &iotago.FoundryOutput{
					Amount:       defaultAmount,
					SerialNumber: 2,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(100),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
					Conditions: iotago.FoundryOutputUnlockConditions{
						&iotago.ImmutableAccountUnlockCondition{Address: iotago.AccountIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AccountAddress)},
					},
					Features: iotago.FoundryOutputFeatures{
						// native token feature added later
					},
				},
				inputIDs[11]: &iotago.FoundryOutput{
					Amount:       defaultAmount,
					SerialNumber: 3,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(100),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
					Conditions: iotago.FoundryOutputUnlockConditions{
						&iotago.ImmutableAccountUnlockCondition{Address: iotago.AccountIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AccountAddress)},
					},
					Features: nil,
				},
				inputIDs[12]: &iotago.FoundryOutput{
					Amount:       defaultAmount,
					SerialNumber: 4,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(100),
						MeltedTokens:  big.NewInt(50),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
					Conditions: iotago.FoundryOutputUnlockConditions{
						&iotago.ImmutableAccountUnlockCondition{Address: iotago.AccountIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AccountAddress)},
					},
					Features: nil,
				},
				inputIDs[13]: &iotago.NFTOutput{
					Amount: defaultAmount,
					NFTID:  nft1ID,
					Conditions: iotago.NFTOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident3},
					},
					Features: iotago.NFTOutputFeatures{
						&iotago.IssuerFeature{Address: ident3},
					},
					ImmutableFeatures: iotago.NFTOutputImmFeatures{
						&iotago.MetadataFeature{Data: []byte("transfer to 4")},
					},
				},
				inputIDs[14]: &iotago.NFTOutput{
					Amount: defaultAmount,
					NFTID:  nft2ID,
					Conditions: iotago.NFTOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident4},
					},
					Features: iotago.NFTOutputFeatures{
						&iotago.IssuerFeature{Address: ident3},
					},
					ImmutableFeatures: iotago.NFTOutputImmFeatures{
						&iotago.MetadataFeature{Data: []byte("going to be destroyed")},
					},
				},
				inputIDs[15]: &iotago.BasicOutput{
					Amount: defaultAmount,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: iotago.NFTID(nft1ID).ToAddress()},
					},
				},
			}

			foundry1Ident3NativeTokenID := inputs[inputIDs[9]].(*iotago.FoundryOutput).MustNativeTokenID()
			foundry2Ident3NativeTokenID := inputs[inputIDs[10]].(*iotago.FoundryOutput).MustNativeTokenID()
			foundry4Ident3NativeTokenID := inputs[inputIDs[12]].(*iotago.FoundryOutput).MustNativeTokenID()

			newFoundryWithInitialSupply := &iotago.FoundryOutput{
				Amount:       defaultAmount,
				SerialNumber: 6,
				TokenScheme: &iotago.SimpleTokenScheme{
					MintedTokens:  big.NewInt(100),
					MeltedTokens:  big.NewInt(0),
					MaximumSupply: new(big.Int).SetInt64(1000),
				},
				Conditions: iotago.FoundryOutputUnlockConditions{
					&iotago.ImmutableAccountUnlockCondition{Address: iotago.AccountIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AccountAddress)},
				},
				Features: nil,
			}
			newFoundryNativeTokenID := newFoundryWithInitialSupply.MustNativeTokenID()
			newFoundryWithInitialSupply.Features.Upsert(&iotago.NativeTokenFeature{
				ID:     newFoundryNativeTokenID,
				Amount: big.NewInt(100),
			})

			inputs[inputIDs[10]].(*iotago.FoundryOutput).Features.Upsert(&iotago.NativeTokenFeature{
				ID:     foundry2Ident3NativeTokenID,
				Amount: big.NewInt(100),
			})

			inputs[inputIDs[12]].(*iotago.FoundryOutput).Features.Upsert(&iotago.NativeTokenFeature{
				ID:     foundry4Ident3NativeTokenID,
				Amount: big.NewInt(50),
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
					&iotago.BasicOutput{
						Amount: defaultAmount,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident5},
						},
					},
					&iotago.BasicOutput{
						Amount: defaultAmount,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident3},
						},
						Features: iotago.BasicOutputFeatures{
							nativeTokenTransfer1,
						},
					},
					&iotago.BasicOutput{
						Amount: defaultAmount,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident4},
						},
						Features: iotago.BasicOutputFeatures{
							nativeTokenTransfer2,
						},
					},
					&iotago.BasicOutput{
						Amount: defaultAmount,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
					},
					&iotago.BasicOutput{
						Amount: defaultAmount,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
					},
					&iotago.BasicOutput{
						Amount: storageDepositReturn,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					&iotago.AccountOutput{
						Amount:         defaultAmount,
						AccountID:      iotago.AccountID{},
						StateIndex:     0,
						StateMetadata:  []byte("a new account output"),
						FoundryCounter: 0,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident3},
							&iotago.GovernorAddressUnlockCondition{Address: ident4},
						},
						Features: nil,
					},
					&iotago.AccountOutput{
						Amount:         defaultAmount,
						AccountID:      iotago.AccountIDFromOutputID(inputIDs[6]),
						StateIndex:     0,
						StateMetadata:  []byte("gov transitioning"),
						FoundryCounter: 0,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident3},
							&iotago.GovernorAddressUnlockCondition{Address: ident4},
						},
						Features: iotago.AccountOutputFeatures{
							&iotago.MetadataFeature{Data: []byte("the gov mutation on this output")},
						},
					},
					&iotago.AccountOutput{
						Amount:         defaultAmount,
						AccountID:      iotago.AccountIDFromOutputID(inputIDs[7]),
						StateIndex:     6,
						StateMetadata:  []byte("next state"),
						FoundryCounter: 6,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident3},
							&iotago.GovernorAddressUnlockCondition{Address: ident4},
						},
						Features: nil,
					},
					// new foundry
					newFoundryWithInitialSupply,
					&iotago.FoundryOutput{
						Amount:       defaultAmount,
						SerialNumber: 1,
						TokenScheme: &iotago.SimpleTokenScheme{
							MintedTokens:  new(big.Int).SetInt64(200),
							MeltedTokens:  big.NewInt(0),
							MaximumSupply: new(big.Int).SetInt64(1000),
						},
						Conditions: iotago.FoundryOutputUnlockConditions{
							&iotago.ImmutableAccountUnlockCondition{Address: iotago.AccountIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AccountAddress)},
						},
						Features: iotago.FoundryOutputFeatures{
							&iotago.NativeTokenFeature{
								ID:     foundry1Ident3NativeTokenID,
								Amount: new(big.Int).SetUint64(100), // freshly minted
							},
						},
					},
					&iotago.FoundryOutput{
						Amount:       defaultAmount,
						SerialNumber: 2,
						TokenScheme: &iotago.SimpleTokenScheme{
							MintedTokens:  new(big.Int).SetInt64(100),
							MeltedTokens:  big.NewInt(50),
							MaximumSupply: new(big.Int).SetInt64(1000),
						},
						Conditions: iotago.FoundryOutputUnlockConditions{
							&iotago.ImmutableAccountUnlockCondition{Address: iotago.AccountIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AccountAddress)},
						},
						Features: iotago.FoundryOutputFeatures{
							&iotago.NativeTokenFeature{
								ID:     foundry2Ident3NativeTokenID,
								Amount: new(big.Int).SetUint64(50), // melted to 50
							},
						},
					},
					&iotago.FoundryOutput{
						Amount:       defaultAmount,
						SerialNumber: 3,
						TokenScheme: &iotago.SimpleTokenScheme{
							MintedTokens:  new(big.Int).SetInt64(100),
							MeltedTokens:  big.NewInt(0),
							MaximumSupply: new(big.Int).SetInt64(1000),
						},
						Conditions: iotago.FoundryOutputUnlockConditions{
							&iotago.ImmutableAccountUnlockCondition{Address: iotago.AccountIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AccountAddress)},
						},
						Features: iotago.FoundryOutputFeatures{
							&iotago.MetadataFeature{Data: []byte("interesting metadata")},
						},
					},
					// from foundry 4 ident 3 destruction remainder
					&iotago.BasicOutput{
						Amount: defaultAmount,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident3},
						},
					},
					&iotago.NFTOutput{
						Amount: defaultAmount,
						NFTID:  iotago.NFTID{},
						Conditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident4},
						},
						Features: nil,
						ImmutableFeatures: iotago.NFTOutputImmFeatures{
							&iotago.MetadataFeature{Data: []byte("immutable metadata")},
						},
					},
					&iotago.NFTOutput{
						Amount: defaultAmount,
						NFTID:  nft1ID,
						Conditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident4},
						},
						Features: iotago.NFTOutputFeatures{
							&iotago.IssuerFeature{Address: ident3},
						},
						ImmutableFeatures: iotago.NFTOutputImmFeatures{
							&iotago.MetadataFeature{Data: []byte("transfer to 4")},
						},
					},
					// from NFT ident 4 destruction remainder
					&iotago.BasicOutput{
						Amount: defaultAmount,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident3},
						},
					},
					// from NFT 1 to ident 5
					&iotago.BasicOutput{
						Amount: defaultAmount,
						Conditions: iotago.BasicOutputUnlockConditions{
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
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.SignatureUnlock{Signature: sigs[1]},
						&iotago.ReferenceUnlock{Reference: 1},
						&iotago.ReferenceUnlock{Reference: 0},
						&iotago.ReferenceUnlock{Reference: 1},
						&iotago.ReferenceUnlock{Reference: 1},
						// account
						&iotago.SignatureUnlock{Signature: sigs[3]},
						&iotago.SignatureUnlock{Signature: sigs[2]},
						&iotago.ReferenceUnlock{Reference: 7},
						// foundries
						&iotago.AccountUnlock{Reference: 7},
						&iotago.AccountUnlock{Reference: 7},
						&iotago.AccountUnlock{Reference: 7},
						&iotago.AccountUnlock{Reference: 7},
						// nfts
						&iotago.ReferenceUnlock{Reference: 7},
						&iotago.ReferenceUnlock{Reference: 6},
						&iotago.NFTUnlock{Reference: 13},
					},
				},
				wantErr: nil,
			}
		}(),
		func() *test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()
			_, ident2, _ := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(2)
			inFoundry := &iotago.FoundryOutput{
				Amount:       100,
				SerialNumber: 5,
				TokenScheme: &iotago.SimpleTokenScheme{
					MintedTokens:  new(big.Int).SetInt64(1000),
					MeltedTokens:  big.NewInt(0),
					MaximumSupply: new(big.Int).SetInt64(10000),
				},
				Conditions: iotago.FoundryOutputUnlockConditions{
					&iotago.ImmutableAccountUnlockCondition{Address: accountAddr1},
				},
			}
			outFoundry := inFoundry.Clone().(*iotago.FoundryOutput)
			// change the immutable account address unlock
			outFoundry.Conditions = iotago.FoundryOutputUnlockConditions{
				&iotago.ImmutableAccountUnlockCondition{Address: tpkg.RandAccountAddress()},
			}

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AccountOutput{
					Amount:     100,
					StateIndex: 0,
					AccountID:  accountAddr1.AccountID(),
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident1},
						&iotago.GovernorAddressUnlockCondition{Address: ident2},
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
						Amount:     100,
						StateIndex: 1,
						AccountID:  accountAddr1.AccountID(),
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident2},
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
		func() *test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, ident1, _ := tpkg.RandEd25519Identity()
			_, ident2, ident2AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AccountOutput{
					Amount:     100,
					StateIndex: 0,
					AccountID:  accountAddr1.AccountID(),
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: iotago.NewBlockIssuerKeys(),
							ExpirySlot:      100,
						},
					},
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident1},
						&iotago.GovernorAddressUnlockCondition{Address: ident2},
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
						Amount:     100,
						StateIndex: 0,
						AccountID:  accountAddr1.AccountID(),
						Features: iotago.AccountOutputFeatures{
							&iotago.BlockIssuerFeature{
								BlockIssuerKeys: iotago.NewBlockIssuerKeys(),
								ExpirySlot:      1000,
							},
						},
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident2},
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

			sigs, err := transaction.Sign(ident2AddressKeys)
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
		func() *test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, ident1, _ := tpkg.RandEd25519Identity()
			_, ident2, ident2AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AccountOutput{
					Amount:     100,
					StateIndex: 0,
					AccountID:  accountAddr1.AccountID(),
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: iotago.NewBlockIssuerKeys(),
							ExpirySlot:      100,
						},
					},
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident1},
						&iotago.GovernorAddressUnlockCondition{Address: ident2},
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
						Amount:     100,
						StateIndex: 0,
						AccountID:  accountAddr1.AccountID(),
						Features: iotago.AccountOutputFeatures{
							&iotago.BlockIssuerFeature{
								BlockIssuerKeys: iotago.NewBlockIssuerKeys(),
								ExpirySlot:      iotago.MaxSlotIndex,
							},
						},
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident2},
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

			sigs, err := transaction.Sign(ident2AddressKeys)
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
		func() *test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AccountOutput{
					Amount:     100,
					StateIndex: 0,
					AccountID:  accountAddr1.AccountID(),
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: iotago.NewBlockIssuerKeys(),
							ExpirySlot:      iotago.MaxSlotIndex,
						},
					},
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident1},
						&iotago.GovernorAddressUnlockCondition{Address: ident1},
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
						Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AccountOutput{
					Amount:     100,
					StateIndex: 0,
					AccountID:  accountAddr1.AccountID(),
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: iotago.NewBlockIssuerKeys(),
							ExpirySlot:      100,
						},
					},
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident1},
						&iotago.GovernorAddressUnlockCondition{Address: ident1},
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
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
			}

			bicInputs := vm.BlockIssuanceCreditInputSet{
				accountAddr1.AccountID(): 0,
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
		func() *test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AccountOutput{
					Amount:     100,
					StateIndex: 0,
					AccountID:  accountAddr1.AccountID(),
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: iotago.NewBlockIssuerKeys(),
							ExpirySlot:      100,
						},
					},
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident1},
						&iotago.GovernorAddressUnlockCondition{Address: ident1},
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
						Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AccountOutput{
					Amount:     100,
					StateIndex: 0,
					AccountID:  accountAddr1.AccountID(),
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: iotago.NewBlockIssuerKeys(),
							ExpirySlot:      100,
						},
					},
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident1},
						&iotago.GovernorAddressUnlockCondition{Address: ident1},
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
						Amount:     100,
						StateIndex: 0,
						AccountID:  accountAddr1.AccountID(),
						Features: iotago.AccountOutputFeatures{
							&iotago.BlockIssuerFeature{
								BlockIssuerKeys: iotago.NewBlockIssuerKeys(),
								ExpirySlot:      1000,
							},
						},
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident1},
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

		totalInputMana, err := vm.TotalManaIn(testAPI.ManaDecayProvider(), testAPI.RentStructure(), txCreationSlot, inputSet)
		require.NoError(t, err)

		outputs := iotago.TxEssenceOutputs{
			// collect everything on a basic output with a random ed25519 address
			&iotago.BasicOutput{
				Amount: totalInputAmount,
				Conditions: iotago.BasicOutputUnlockConditions{
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

	var defaultAmount iotago.BaseToken = OneMi

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
								Conditions: iotago.BasicOutputUnlockConditions{
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
					ed25519AddrCnt: 2,
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
								StateIndex:     1,
								StateMetadata:  []byte("current state"),
								FoundryCounter: 0,
								Conditions: iotago.AccountOutputUnlockConditions{
									&iotago.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
									&iotago.GovernorAddressUnlockCondition{Address: ed25519Addresses[1]},
								},
								Features: nil,
							},
							// owned by restricted account address
							&iotago.BasicOutput{
								Amount: defaultAmount,
								Conditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[1]},
								},
							},
						}
					},
					outputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address, totalInputAmount iotago.BaseToken, totalInputMana iotago.Mana) iotago.TxEssenceOutputs {
						return iotago.TxEssenceOutputs{
							// the account unlock needs to be a state transition (governor doesn't work for account reference unlocks)
							&iotago.AccountOutput{
								Amount:         defaultAmount,
								AccountID:      testAddresses[0].(*iotago.AccountAddress).AccountID(),
								StateIndex:     2,
								StateMetadata:  []byte("next state"),
								FoundryCounter: 0,
								Conditions: iotago.AccountOutputUnlockConditions{
									&iotago.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
									&iotago.GovernorAddressUnlockCondition{Address: ed25519Addresses[1]},
								},
								Features: nil,
							},
							&iotago.BasicOutput{
								Amount: totalInputAmount - defaultAmount,
								Conditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
							},
						}
					},
					unlocksFunc: func(sigs []iotago.Signature, testAddresses []iotago.Address) iotago.Unlocks {
						return iotago.Unlocks{
							&iotago.SignatureUnlock{Signature: sigs[0]}, // account state controller unlock
							&iotago.AccountUnlock{Reference: 0},
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
								Conditions: iotago.NFTOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
								Features: iotago.NFTOutputFeatures{
									&iotago.IssuerFeature{Address: ed25519Addresses[1]},
								},
								ImmutableFeatures: iotago.NFTOutputImmFeatures{
									&iotago.MetadataFeature{Data: []byte("immutable")},
								},
							},
							// owned by restricted NFT address
							&iotago.BasicOutput{
								Amount: defaultAmount,
								Conditions: iotago.BasicOutputUnlockConditions{
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
								Conditions: iotago.NFTOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
								Features: iotago.NFTOutputFeatures{
									&iotago.IssuerFeature{Address: ed25519Addresses[1]},
									&iotago.MetadataFeature{Data: []byte("some new metadata")},
								},
								ImmutableFeatures: iotago.NFTOutputImmFeatures{
									&iotago.MetadataFeature{Data: []byte("immutable")},
								},
							},
							&iotago.BasicOutput{
								Amount: totalInputAmount - defaultAmount,
								Conditions: iotago.BasicOutputUnlockConditions{
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
	}
	for _, tt := range tests {
		runNovaTransactionExecutionTest(t, tt)
	}
}

func TestNovaTransactionExecution_MultiAddress(t *testing.T) {

	var defaultAmount iotago.BaseToken = OneMi

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
								Conditions: iotago.BasicOutputUnlockConditions{
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
								Conditions: iotago.BasicOutputUnlockConditions{
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
								Conditions: iotago.BasicOutputUnlockConditions{
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
								Conditions: iotago.BasicOutputUnlockConditions{
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
								Conditions: iotago.BasicOutputUnlockConditions{
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
								Conditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[1]},
								},
							},
							&iotago.BasicOutput{
								Amount: defaultAmount,
								Conditions: iotago.BasicOutputUnlockConditions{
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
								Conditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
							&iotago.BasicOutput{
								Amount: defaultAmount,
								Conditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
							&iotago.BasicOutput{
								Amount: defaultAmount,
								Conditions: iotago.BasicOutputUnlockConditions{
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

		// ok - Account unlock (state transition)
		func() *txExecTest {
			return &txExecTest{
				name: "ok - Account unlock (state transition)",
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
								StateIndex:     1,
								StateMetadata:  []byte("current state"),
								FoundryCounter: 0,
								Conditions: iotago.AccountOutputUnlockConditions{
									&iotago.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
									&iotago.GovernorAddressUnlockCondition{Address: ed25519Addresses[1]},
								},
								Features: nil,
							},
							&iotago.BasicOutput{
								Amount: defaultAmount,
								Conditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[1]},
								},
							},
							// owned by ed25519 address + account address
							&iotago.BasicOutput{
								Amount: defaultAmount,
								Conditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[1]},
								},
							},
						}
					},
					outputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address, totalInputAmount iotago.BaseToken, totalInputMana iotago.Mana) iotago.TxEssenceOutputs {
						return iotago.TxEssenceOutputs{
							// the account unlock needs to be a state transition (governor doesn't work for account reference unlocks)
							&iotago.AccountOutput{
								Amount:         defaultAmount,
								AccountID:      testAddresses[0].(*iotago.AccountAddress).AccountID(),
								StateIndex:     2,
								StateMetadata:  []byte("next state"),
								FoundryCounter: 0,
								Conditions: iotago.AccountOutputUnlockConditions{
									&iotago.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
									&iotago.GovernorAddressUnlockCondition{Address: ed25519Addresses[1]},
								},
								Features: nil,
							},
							&iotago.BasicOutput{
								Amount: totalInputAmount - defaultAmount,
								Conditions: iotago.BasicOutputUnlockConditions{
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
							&iotago.SignatureUnlock{Signature: sigs[0]}, // account state controller unlock
							&iotago.SignatureUnlock{Signature: sigs[1]}, // basic output unlock
							multiUnlock,
						}
					},
				},
				wantErr: nil,
			}
		}(),

		// fail - Account unlock (governance transition)
		func() *txExecTest {
			return &txExecTest{
				name: "fail - Account unlock (governance transition)",
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
										Address: ed25519Addresses[0],
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
								StateIndex:     1,
								StateMetadata:  []byte("governance transition"),
								FoundryCounter: 0,
								Conditions: iotago.AccountOutputUnlockConditions{
									&iotago.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
									&iotago.GovernorAddressUnlockCondition{Address: ed25519Addresses[1]},
								},
								Features: nil,
							},
							&iotago.BasicOutput{
								Amount: defaultAmount,
								Conditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
							},
							// owned by ed25519 address + account address
							&iotago.BasicOutput{
								Amount: defaultAmount,
								Conditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[1]},
								},
							},
						}
					},
					outputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address, totalInputAmount iotago.BaseToken, totalInputMana iotago.Mana) iotago.TxEssenceOutputs {
						return iotago.TxEssenceOutputs{
							// the account unlock needs to be a state transition (governor doesn't work for account reference unlocks)
							&iotago.AccountOutput{
								Amount:         defaultAmount,
								AccountID:      testAddresses[0].(*iotago.AccountAddress).AccountID(),
								StateIndex:     1,
								StateMetadata:  []byte("governance transition"),
								FoundryCounter: 0,
								Conditions: iotago.AccountOutputUnlockConditions{
									&iotago.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
									&iotago.GovernorAddressUnlockCondition{Address: ed25519Addresses[1]},
								},
								Features: nil,
							},
							&iotago.BasicOutput{
								Amount: totalInputAmount - defaultAmount,
								Conditions: iotago.BasicOutputUnlockConditions{
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
							&iotago.SignatureUnlock{Signature: sigs[1]}, // account governor unlock
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
								Conditions: iotago.NFTOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
								Features: iotago.NFTOutputFeatures{
									&iotago.IssuerFeature{Address: ed25519Addresses[1]},
								},
								ImmutableFeatures: iotago.NFTOutputImmFeatures{
									&iotago.MetadataFeature{Data: []byte("immutable")},
								},
							},
							&iotago.BasicOutput{
								Amount: defaultAmount,
								Conditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[1]},
								},
							},
							// owned by ed25519 address + NFT address
							&iotago.BasicOutput{
								Amount: defaultAmount,
								Conditions: iotago.BasicOutputUnlockConditions{
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
								Conditions: iotago.NFTOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: ed25519Addresses[0]},
								},
								Features: iotago.NFTOutputFeatures{
									&iotago.IssuerFeature{Address: ed25519Addresses[1]},
									&iotago.MetadataFeature{Data: []byte("some new metadata")},
								},
								ImmutableFeatures: iotago.NFTOutputImmFeatures{
									&iotago.MetadataFeature{Data: []byte("immutable")},
								},
							},
							&iotago.BasicOutput{
								Amount: totalInputAmount - defaultAmount,
								Conditions: iotago.BasicOutputUnlockConditions{
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
								Conditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
							&iotago.BasicOutput{
								Amount: defaultAmount,
								Conditions: iotago.BasicOutputUnlockConditions{
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
								Conditions: iotago.BasicOutputUnlockConditions{
									&iotago.AddressUnlockCondition{Address: testAddresses[0]},
								},
							},
							&iotago.BasicOutput{
								Amount: defaultAmount,
								Conditions: iotago.BasicOutputUnlockConditions{
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

	var defaultAmount iotago.BaseToken = OneMi

	// builds a transaction that burns native tokens
	burnNativeTokenTxBuilder := &txBuilder{
		ed25519AddrCnt: 1,
		inputsFunc: func(ed25519Addresses []iotago.Address, testAddresses []iotago.Address) []iotago.Output {
			return []iotago.Output{
				&iotago.BasicOutput{
					Amount: defaultAmount,
					// add native tokens
					Conditions: iotago.BasicOutputUnlockConditions{
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
					Conditions: iotago.BasicOutputUnlockConditions{
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
					StateIndex:     0,
					StateMetadata:  []byte{},
					FoundryCounter: 1,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
						&iotago.GovernorAddressUnlockCondition{Address: ed25519Addresses[0]},
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
					Conditions: iotago.FoundryOutputUnlockConditions{
						&iotago.ImmutableAccountUnlockCondition{
							Address: testAddresses[0].(*iotago.AccountAddress),
						},
					},
					Features:          iotago.FoundryOutputFeatures{},
					ImmutableFeatures: iotago.FoundryOutputImmFeatures{},
				},
				&iotago.BasicOutput{
					Amount: defaultAmount,
					Conditions: iotago.BasicOutputUnlockConditions{
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
					StateIndex:     1,
					StateMetadata:  []byte{},
					FoundryCounter: 1,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
						&iotago.GovernorAddressUnlockCondition{Address: ed25519Addresses[0]},
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
					Conditions: iotago.FoundryOutputUnlockConditions{
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
					StateIndex:     0,
					StateMetadata:  []byte{},
					FoundryCounter: 1,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
						&iotago.GovernorAddressUnlockCondition{Address: ed25519Addresses[0]},
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
					Conditions: iotago.FoundryOutputUnlockConditions{
						&iotago.ImmutableAccountUnlockCondition{
							Address: testAddresses[0].(*iotago.AccountAddress),
						},
					},
					Features:          iotago.FoundryOutputFeatures{},
					ImmutableFeatures: iotago.FoundryOutputImmFeatures{},
				},
				&iotago.BasicOutput{
					Amount: defaultAmount,
					Conditions: iotago.BasicOutputUnlockConditions{
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
					StateIndex:     1,
					StateMetadata:  []byte{},
					FoundryCounter: 1,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
						&iotago.GovernorAddressUnlockCondition{Address: ed25519Addresses[0]},
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
					Conditions: iotago.FoundryOutputUnlockConditions{
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
					Conditions: iotago.BasicOutputUnlockConditions{
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
					Conditions: iotago.BasicOutputUnlockConditions{
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
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.GovernorAddressUnlockCondition{Address: ed25519Addresses[0]},
						&iotago.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
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
					Conditions: iotago.BasicOutputUnlockConditions{
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
					StateIndex:     0,
					StateMetadata:  []byte{},
					FoundryCounter: 1,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
						&iotago.GovernorAddressUnlockCondition{Address: ed25519Addresses[0]},
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
					Conditions: iotago.FoundryOutputUnlockConditions{
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
					StateIndex:     1,
					StateMetadata:  []byte{},
					FoundryCounter: 1,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ed25519Addresses[0]},
						&iotago.GovernorAddressUnlockCondition{Address: ed25519Addresses[0]},
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
					Conditions: iotago.NFTOutputUnlockConditions{
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
					Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, ident2AddrKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(8)
			accountIdent1 := iotago.AccountAddressFromOutputID(inputIDs[1])
			nftIdent1 := tpkg.RandNFTAddress()

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: &iotago.AccountOutput{
					Amount:    100,
					AccountID: iotago.AccountID{}, // empty on purpose as validation should resolve
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident1},
						&iotago.GovernorAddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[2]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: accountIdent1},
					},
				},
				inputIDs[3]: &iotago.NFTOutput{
					Amount: 100,
					NFTID:  nftIdent1.NFTID(),
					Conditions: iotago.NFTOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: accountIdent1},
					},
				},
				inputIDs[4]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: nftIdent1},
					},
				},
				// unlockable by sender as expired
				inputIDs[5]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: ident2,
							Slot:          5,
						},
					},
				},
				// not unlockable by sender as not expired
				inputIDs[6]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: ident2,
							Slot:          30,
						},
					},
				},
				inputIDs[7]: &iotago.FoundryOutput{
					Amount:       100,
					SerialNumber: 0,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetInt64(100),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetInt64(1000),
					},
					Conditions: iotago.FoundryOutputUnlockConditions{
						&iotago.ImmutableAccountUnlockCondition{Address: accountIdent1},
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
						Amount:     100,
						AccountID:  accountIdent1.AccountID(),
						StateIndex: 1,
						Conditions: iotago.AccountOutputUnlockConditions{
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
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.ReferenceUnlock{Reference: 0},
						&iotago.AccountUnlock{Reference: 1},
						&iotago.AccountUnlock{Reference: 1},
						&iotago.NFTUnlock{Reference: 3},
						&iotago.SignatureUnlock{Signature: sigs[1]},
						&iotago.ReferenceUnlock{Reference: 0},
						&iotago.AccountUnlock{Reference: 1},
					},
				},
				wantErr: nil,
			}
		}(),
		func() *test {
			ident1Sk, ident1, _ := tpkg.RandEd25519Identity()
			_, _, ident2AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)

			accountIdent1 := iotago.AccountAddressFromOutputID(inputIDs[0])
			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AccountOutput{
					Amount:    100,
					AccountID: iotago.AccountID{},
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident1},
						&iotago.GovernorAddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)

			nftIdent1 := iotago.NFTAddressFromOutputID(inputIDs[0])
			inputs := vm.InputSet{
				inputIDs[0]: &iotago.NFTOutput{
					Amount: 100,
					NFTID:  iotago.NFTID{},
					Conditions: iotago.NFTOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			inputIDs := tpkg.RandOutputIDs(2)

			nftIdent1 := iotago.NFTAddressFromOutputID(inputIDs[0])
			nftIdent2 := iotago.NFTAddressFromOutputID(inputIDs[1])

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.NFTOutput{
					Amount: 100,
					NFTID:  nftIdent1.NFTID(),
					Conditions: iotago.NFTOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: nftIdent2},
					},
				},
				inputIDs[1]: &iotago.NFTOutput{
					Amount: 100,
					NFTID:  nftIdent2.NFTID(),
					Conditions: iotago.NFTOutputUnlockConditions{
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
		func() *test {
			_, ident1, _ := tpkg.RandEd25519Identity()
			_, ident2, ident2AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()
			_, ident2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
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
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident1},
						&iotago.GovernorAddressUnlockCondition{Address: ident1},
					},
				},
				// owned by account1
				inputIDs[1]: &iotago.AccountOutput{
					Amount:    100,
					AccountID: accountAddr2.AccountID(),
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: accountAddr1},
						&iotago.GovernorAddressUnlockCondition{Address: accountAddr1},
					},
				},
				// owned by account1
				inputIDs[2]: &iotago.AccountOutput{
					Amount:    100,
					AccountID: accountAddr3.AccountID(),
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: accountAddr1},
						&iotago.GovernorAddressUnlockCondition{Address: accountAddr1},
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
		func() *test {
			_, ident1, _ := tpkg.RandEd25519Identity()
			_, ident2, ident2AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)

			accountAddr1 := tpkg.RandAccountAddress()

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AccountOutput{
					Amount:    100,
					AccountID: accountAddr1.AccountID(),
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident1},
						&iotago.GovernorAddressUnlockCondition{Address: ident2},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: accountAddr1},
					},
				},
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
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident2},
						},
					},
				},
			}

			sigs, err := transaction.Sign(ident2AddressKeys)
			require.NoError(t, err)

			return &test{
				name: "fail - account output not state transitioning",
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
					},
				},
				wantErr: iotago.ErrInvalidInputUnlock,
			}
		}(),
		func() *test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()
			_, ident2, _ := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(2)
			foundryOutput := &iotago.FoundryOutput{
				Amount:       100,
				SerialNumber: 5,
				TokenScheme: &iotago.SimpleTokenScheme{
					MintedTokens:  new(big.Int).SetInt64(1000),
					MeltedTokens:  big.NewInt(0),
					MaximumSupply: new(big.Int).SetInt64(10000),
				},
				Conditions: iotago.FoundryOutputUnlockConditions{
					&iotago.ImmutableAccountUnlockCondition{Address: accountAddr1},
				},
			}

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AccountOutput{
					Amount:     100,
					StateIndex: 0,
					AccountID:  accountAddr1.AccountID(),
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident1},
						&iotago.GovernorAddressUnlockCondition{Address: ident2},
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
						Amount:     100,
						StateIndex: 1,
						AccountID:  accountAddr1.AccountID(),
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident2},
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
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, ident2AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(3)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				// unlocked by ident1 as it is not expired
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 500,
					Conditions: iotago.BasicOutputUnlockConditions{
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
					Conditions: iotago.BasicOutputUnlockConditions{
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
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
					&iotago.BasicOutput{
						// return via ident1 + reclaim
						Amount: 420 + 500,
						Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 1000,
					Conditions: iotago.BasicOutputUnlockConditions{
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
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
					},
					&iotago.BasicOutput{
						// returns 221 to ident2
						Amount: 221,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
					},
					&iotago.BasicOutput{
						// remainder to random address
						Amount: 579,
						Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 50,
					Conditions: iotago.BasicOutputUnlockConditions{
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
						Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
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
						Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 500,
					Conditions: iotago.BasicOutputUnlockConditions{
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
						Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 500,
					Conditions: iotago.BasicOutputUnlockConditions{
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
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					&iotago.NFTOutput{
						Amount: 420,
						Conditions: iotago.NFTOutputUnlockConditions{
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
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 500,
					Conditions: iotago.BasicOutputUnlockConditions{
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
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					&iotago.BasicOutput{
						Amount: 420,
						Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 500,
					Conditions: iotago.BasicOutputUnlockConditions{
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
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					&iotago.BasicOutput{
						Amount: 420,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
						Features: iotago.BasicOutputFeatures{
							&iotago.MetadataFeature{Data: []byte("foo")},
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
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)
			ntID := tpkg.Rand38ByteArray()

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 500,
					Conditions: iotago.BasicOutputUnlockConditions{
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
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					&iotago.BasicOutput{
						Amount: 420,
						Conditions: iotago.BasicOutputUnlockConditions{
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
		Conditions: iotago.FoundryOutputUnlockConditions{
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
		Conditions: iotago.FoundryOutputUnlockConditions{
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
		func() *test {
			inputIDs := tpkg.RandOutputIDs(2)

			nativeTokenFeature1 := tpkg.RandNativeTokenFeature()
			nativeTokenFeature2 := tpkg.RandNativeTokenFeature()

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
					Features: iotago.BasicOutputFeatures{
						nativeTokenFeature1,
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
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
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: iotago.BasicOutputFeatures{
							nativeTokenFeature1,
						},
					},
					&iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			inputIDs := tpkg.RandOutputIDs(iotago.MaxInputsCount)
			nativeToken := tpkg.RandNativeTokenFeature()

			inputs := vm.InputSet{}
			for i := 0; i < iotago.MaxInputsCount; i++ {
				inputs[inputIDs[i]] = &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
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
						Conditions: iotago.BasicOutputUnlockConditions{
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
					Conditions: iotago.BasicOutputUnlockConditions{
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
					Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			inputIDs := tpkg.RandOutputIDs(1)

			nativeTokenFeature := tpkg.RandNativeTokenFeature()

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
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
						Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			inputIDs := tpkg.RandOutputIDs(3)

			nativeTokenFeature1 := tpkg.RandNativeTokenFeature()
			nativeTokenFeature2 := nativeTokenFeature1.Clone()

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
					Features: iotago.BasicOutputFeatures{
						nativeTokenFeature1,
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
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
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: iotago.BasicOutputFeatures{
							cpyNativeTokenFeature,
						},
					},
					&iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)
			nftAddr := tpkg.RandNFTAddress()

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: &iotago.NFTOutput{
					Amount: 100,
					NFTID:  nftAddr.NFTID(),
					Conditions: iotago.NFTOutputUnlockConditions{
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
					// sender is an Ed25519 address
					&iotago.BasicOutput{
						Amount: 1337,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: iotago.BasicOutputFeatures{
							&iotago.SenderFeature{Address: ident1},
						},
					},
					&iotago.AccountOutput{
						Amount: 1337,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident1},
						},
						Features: iotago.AccountOutputFeatures{
							&iotago.SenderFeature{Address: ident1},
						},
					},
					&iotago.NFTOutput{
						Amount: 1337,
						Conditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: iotago.NFTOutputFeatures{
							&iotago.SenderFeature{Address: ident1},
						},
					},
					// sender is an NFT address
					&iotago.BasicOutput{
						Amount: 1337,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: iotago.BasicOutputFeatures{
							&iotago.SenderFeature{Address: nftAddr},
						},
					},
					&iotago.AccountOutput{
						Amount: 1337,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident1},
						},
						Features: iotago.AccountOutputFeatures{
							&iotago.SenderFeature{Address: nftAddr},
						},
					},
					&iotago.NFTOutput{
						Amount: 1337,
						Conditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: iotago.NFTOutputFeatures{
							&iotago.SenderFeature{Address: nftAddr},
						},
					},
				},
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
					},
				},
				wantErr: nil,
			}
		}(),
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
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
						Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			_, stateController, _ := tpkg.RandEd25519Identity()
			_, governor, governorAddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)
			accountAddr := tpkg.RandAccountAddress()
			accountID := accountAddr.AccountID()

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AccountOutput{
					Amount:    100,
					AccountID: accountID,
					Conditions: iotago.AccountOutputUnlockConditions{
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
					&iotago.AccountOutput{
						Amount:    100,
						AccountID: accountID,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
						Features: iotago.AccountOutputFeatures{
							&iotago.SenderFeature{Address: accountAddr},
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
		func() *test {
			_, stateController, stateControllerAddrKeys := tpkg.RandEd25519Identity()
			_, governor, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)
			accountAddr := tpkg.RandAccountAddress()
			accountID := accountAddr.AccountID()
			currentStateIndex := uint32(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AccountOutput{
					Amount:     100,
					AccountID:  accountID,
					StateIndex: currentStateIndex,
					Conditions: iotago.AccountOutputUnlockConditions{
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
					&iotago.AccountOutput{
						Amount:     100,
						AccountID:  accountID,
						StateIndex: currentStateIndex + 1,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
						Features: iotago.AccountOutputFeatures{
							&iotago.SenderFeature{Address: accountAddr},
						},
					},
				},
			}
			sigs, err := transaction.Sign(stateControllerAddrKeys)
			require.NoError(t, err)

			return &test{
				name: "ok - account addr unlocked with state transition",
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
		func() *test {
			_, stateController, _ := tpkg.RandEd25519Identity()
			_, governor, governorAddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)
			accountAddr := tpkg.RandAccountAddress()
			accountID := accountAddr.AccountID()

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AccountOutput{
					Amount:    100,
					AccountID: accountID,
					Conditions: iotago.AccountOutputUnlockConditions{
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
					&iotago.AccountOutput{
						Amount:    100,
						AccountID: accountID,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
						Features: iotago.AccountOutputFeatures{
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
					Conditions: iotago.BasicOutputUnlockConditions{
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
						Conditions: iotago.BasicOutputUnlockConditions{
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
					Conditions: iotago.BasicOutputUnlockConditions{
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
						Conditions: iotago.NFTOutputUnlockConditions{
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
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, stateController, _ := tpkg.RandEd25519Identity()
			_, governor, governorAddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)
			accountAddr := tpkg.RandAccountAddress()
			accountID := accountAddr.AccountID()

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AccountOutput{
					Amount:    100,
					AccountID: accountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: stateController},
						&iotago.GovernorAddressUnlockCondition{Address: governor},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
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
					&iotago.AccountOutput{
						Amount:    100,
						AccountID: accountID,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
					},
					&iotago.NFTOutput{
						Amount: 100,
						Conditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
						ImmutableFeatures: iotago.NFTOutputImmFeatures{
							&iotago.IssuerFeature{Address: accountAddr},
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
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, stateController, stateControllerAddrKeys := tpkg.RandEd25519Identity()
			_, governor, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)
			accountAddr := tpkg.RandAccountAddress()
			accountID := accountAddr.AccountID()
			currentStateIndex := uint32(1)

			nftAddr := tpkg.RandNFTAddress()

			inputs := vm.InputSet{
				// possible issuers: accountAddr, stateController, nftAddr, ident1
				inputIDs[0]: &iotago.AccountOutput{
					Amount:     100,
					AccountID:  accountID,
					StateIndex: currentStateIndex,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: stateController},
						&iotago.GovernorAddressUnlockCondition{Address: governor},
					},
				},
				inputIDs[1]: &iotago.NFTOutput{
					Amount: 900,
					NFTID:  nftAddr.NFTID(),
					Conditions: iotago.NFTOutputUnlockConditions{
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
					// transitioned account + nft
					&iotago.AccountOutput{
						Amount:     100,
						AccountID:  accountID,
						StateIndex: currentStateIndex + 1,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
					},
					&iotago.NFTOutput{
						Amount: 100,
						NFTID:  nftAddr.NFTID(),
						Conditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					// issuer is accountAddr
					&iotago.NFTOutput{
						Amount: 100,
						Conditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
						ImmutableFeatures: iotago.NFTOutputImmFeatures{
							&iotago.IssuerFeature{Address: accountAddr},
						},
					},
					&iotago.AccountOutput{
						Amount: 100,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
						ImmutableFeatures: iotago.AccountOutputImmFeatures{
							&iotago.IssuerFeature{Address: accountAddr},
						},
					},
					// issuer is stateController
					&iotago.NFTOutput{
						Amount: 100,
						Conditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
						ImmutableFeatures: iotago.NFTOutputImmFeatures{
							&iotago.IssuerFeature{Address: stateController},
						},
					},
					&iotago.AccountOutput{
						Amount: 100,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
						ImmutableFeatures: iotago.AccountOutputImmFeatures{
							&iotago.IssuerFeature{Address: stateController},
						},
					},
					// issuer is nftAddr
					&iotago.NFTOutput{
						Amount: 100,
						Conditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
						ImmutableFeatures: iotago.NFTOutputImmFeatures{
							&iotago.IssuerFeature{Address: nftAddr},
						},
					},
					&iotago.AccountOutput{
						Amount: 100,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
						ImmutableFeatures: iotago.AccountOutputImmFeatures{
							&iotago.IssuerFeature{Address: nftAddr},
						},
					},
					// issuer is ident1
					&iotago.NFTOutput{
						Amount: 100,
						Conditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
						ImmutableFeatures: iotago.NFTOutputImmFeatures{
							&iotago.IssuerFeature{Address: ident1},
						},
					},
					&iotago.AccountOutput{
						Amount: 100,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
						ImmutableFeatures: iotago.AccountOutputImmFeatures{
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
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, stateController, _ := tpkg.RandEd25519Identity()
			_, governor, governorAddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)
			accountAddr := tpkg.RandAccountAddress()
			accountID := accountAddr.AccountID()

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.AccountOutput{
					Amount:    100,
					AccountID: accountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: stateController},
						&iotago.GovernorAddressUnlockCondition{Address: governor},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
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
					&iotago.AccountOutput{
						Amount:    100,
						AccountID: accountID,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
					},
					&iotago.NFTOutput{
						Amount: 100,
						Conditions: iotago.NFTOutputUnlockConditions{
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
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.TimelockUnlockCondition{
							Slot: 25,
						},
					},
				},
			}

			creationSlot := iotago.SlotIndex(10)
			transaction := &iotago.Transaction{API: testAPI, TransactionEssence: &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs(), CreationSlot: 10}}
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
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.TimelockUnlockCondition{
							Slot: 1337,
						},
					},
				},
			}

			creationSlot := iotago.SlotIndex(666)
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
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDsWithCreationSlot(10, 1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: OneMi,
					Mana:   iotago.MaxMana,
					Conditions: iotago.BasicOutputUnlockConditions{
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
						Amount: OneMi,
						Mana: func() iotago.Mana {
							var creationSlot iotago.SlotIndex = 10
							targetSlot := 10 + 100*testProtoParams.ParamEpochDurationInSlots()

							input := inputs[inputIDs[0]]
							storageScoreStructure := iotago.NewStorageScoreStructure(testProtoParams.StorageScoreParameters())
							minDeposit, err := storageScoreStructure.MinDeposit(input)
							require.NoError(t, err)
							excessBaseTokens, err := safemath.SafeSub(input.BaseTokenAmount(), minDeposit)
							require.NoError(t, err)
							potentialMana, err := testProtoParams.ManaDecayProvider().ManaGenerationWithDecay(excessBaseTokens, creationSlot, targetSlot)
							require.NoError(t, err)

							storedMana, err := testProtoParams.ManaDecayProvider().ManaWithDecay(iotago.MaxMana, creationSlot, targetSlot)
							require.NoError(t, err)

							return potentialMana + storedMana
						}(),
						Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDsWithCreationSlot(10, 1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: OneMi,
					Mana:   iotago.MaxMana,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			transaction := &iotago.Transaction{
				API: testAPI,
				TransactionEssence: &iotago.TransactionEssence{
					Inputs: inputIDs.UTXOInputs(),
					Allotments: iotago.Allotments{
						&iotago.Allotment{Value: 50},
					},
					CreationSlot: 10 + 100*testProtoParams.ParamEpochDurationInSlots(),
				},
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: OneMi,
						Mana: func() iotago.Mana {
							var createdSlot iotago.SlotIndex = 10
							targetSlot := 10 + 100*testProtoParams.ParamEpochDurationInSlots()

							input := inputs[inputIDs[0]]
							storageScoreStructure := iotago.NewStorageScoreStructure(testProtoParams.StorageScoreParameters())
							minDeposit, err := storageScoreStructure.MinDeposit(input)
							require.NoError(t, err)
							excessBaseTokens, err := safemath.SafeSub(input.BaseTokenAmount(), minDeposit)
							require.NoError(t, err)
							potentialMana, err := testProtoParams.ManaDecayProvider().ManaGenerationWithDecay(excessBaseTokens, createdSlot, targetSlot)
							require.NoError(t, err)

							storedMana, err := testProtoParams.ManaDecayProvider().ManaWithDecay(iotago.MaxMana, createdSlot, targetSlot)
							require.NoError(t, err)

							// generated mana + decay - allotment
							return potentialMana + storedMana - 50
						}(),
						Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDsWithCreationSlot(20, 1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 5,
					Mana:   10,
					Conditions: iotago.BasicOutputUnlockConditions{
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
						Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDsWithCreationSlot(15, 1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 5,
					Mana:   10,
					Conditions: iotago.BasicOutputUnlockConditions{
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
						Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDsWithCreationSlot(15, 2)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 5,
					Mana:   iotago.MaxMana,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 5,
					Mana:   10,
					Conditions: iotago.BasicOutputUnlockConditions{
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
						Conditions: iotago.BasicOutputUnlockConditions{
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
		func() *test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDsWithCreationSlot(15, 1)

			inputs := vm.InputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 5,
					Mana:   10,
					Conditions: iotago.BasicOutputUnlockConditions{
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
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
					&iotago.BasicOutput{
						Amount: 5,
						Mana:   iotago.MaxMana,
						Conditions: iotago.BasicOutputUnlockConditions{
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

	var manaRewardAmount iotago.Mana = 200
	currentEpoch := iotago.EpochIndex(20)
	currentSlot := testProtoParams.TimeProvider().EpochStart(currentEpoch)

	inputIDs := tpkg.RandOutputIDs(1)
	inputs := vm.InputSet{
		inputIDs[0]: &iotago.AccountOutput{
			Amount:         OneMi * 10,
			AccountID:      accountIdent.AccountID(),
			StateIndex:     1,
			StateMetadata:  nil,
			Mana:           0,
			FoundryCounter: 0,
			Conditions: iotago.AccountOutputUnlockConditions{
				&iotago.StateControllerAddressUnlockCondition{Address: ident},
				&iotago.GovernorAddressUnlockCondition{Address: ident},
			},
			Features: iotago.AccountOutputFeatures{
				&iotago.StakingFeature{
					StakedAmount: 100,
					FixedCost:    50,
					StartEpoch:   currentEpoch - 10,
					EndEpoch:     currentEpoch - 1,
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
			&iotago.AccountOutput{
				Amount:         OneMi * 5,
				AccountID:      accountIdent.AccountID(),
				StateIndex:     2,
				StateMetadata:  nil,
				FoundryCounter: 0,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: ident},
					&iotago.GovernorAddressUnlockCondition{Address: ident},
				},
				Features: nil,
			},
			&iotago.BasicOutput{
				Amount: OneMi * 5,
				Mana:   manaRewardAmount,
				Conditions: iotago.BasicOutputUnlockConditions{
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
			accountIdent.AccountID(): manaRewardAmount,
		},
		CommitmentInput: &iotago.Commitment{
			Slot: currentSlot,
		},
	}
	require.NoError(t, validateAndExecuteSignedTransaction(tx, resolvedInputs))
}

func TestManaRewardsClaimingDelegation(t *testing.T) {
	_, ident, identAddrKeys := tpkg.RandEd25519Identity()

	const manaRewardAmount iotago.Mana = 200
	currentSlot := 20 * testProtoParams.ParamEpochDurationInSlots()
	currentEpoch := testProtoParams.TimeProvider().EpochFromSlot(currentSlot)

	inputIDs := tpkg.RandOutputIDs(1)
	inputs := vm.InputSet{
		inputIDs[0]: &iotago.DelegationOutput{
			Amount:           OneMi * 10,
			DelegatedAmount:  OneMi * 10,
			DelegationID:     iotago.EmptyDelegationID(),
			ValidatorAddress: &iotago.AccountAddress{},
			StartEpoch:       currentEpoch,
			EndEpoch:         currentEpoch + 5,
			Conditions: iotago.DelegationOutputUnlockConditions{
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
				Amount: OneMi * 10,
				Mana:   manaRewardAmount,
				Conditions: iotago.BasicOutputUnlockConditions{
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

func TestTxSemanticAddressRestrictions(t *testing.T) {
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

	iotago.RestrictedAddressWithCapabilities(addr)

	tests := []*test{
		{
			createTestOutput: func(address iotago.Address) iotago.Output {
				return &iotago.BasicOutput{
					Conditions: iotago.BasicOutputUnlockConditions{
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
					Conditions: iotago.BasicOutputUnlockConditions{
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
					Conditions: iotago.BasicOutputUnlockConditions{
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
					Conditions: iotago.BasicOutputUnlockConditions{
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
					Conditions: iotago.BasicOutputUnlockConditions{
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
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: address},
					},
				}
			},
			createTestParameters: []func() testParameters{
				func() testParameters {
					return testParameters{
						name:    "ok - Account Output Address in State Controller UC in Account Output",
						address: iotago.RestrictedAddressWithCapabilities(addr, iotago.WithAddressCanReceiveAccountOutputs(true)),
						wantErr: nil,
					}
				},
				func() testParameters {
					return testParameters{
						name:    "fail - Non Account Output Address in State Controller UC in Account Output",
						address: iotago.RestrictedAddressWithCapabilities(addr),
						wantErr: iotago.ErrAddressCannotReceiveAccountOutput,
					}
				},
			},
		},
		{
			createTestOutput: func(address iotago.Address) iotago.Output {
				return &iotago.AccountOutput{
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.GovernorAddressUnlockCondition{Address: address},
					},
				}
			},
			createTestParameters: []func() testParameters{
				func() testParameters {
					return testParameters{
						name:    "ok - Account Output Address in Governor UC in Account Output",
						address: iotago.RestrictedAddressWithCapabilities(addr, iotago.WithAddressCanReceiveAccountOutputs(true)),
						wantErr: nil,
					}
				},
				func() testParameters {
					return testParameters{
						name:    "fail - Non Account Output Address in Governor UC in Account Output",
						address: iotago.RestrictedAddressWithCapabilities(addr),
						wantErr: iotago.ErrAddressCannotReceiveAccountOutput,
					}
				},
			},
		},
		{
			createTestOutput: func(address iotago.Address) iotago.Output {
				return &iotago.NFTOutput{
					Conditions: iotago.NFTOutputUnlockConditions{
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
					Conditions: iotago.DelegationOutputUnlockConditions{
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
					Conditions: iotago.BasicOutputUnlockConditions{
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
				Conditions: iotago.BasicOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: ident},
				},
			},
		}

		transaction := &iotago.Transaction{
			API: testAPI,
			TransactionEssence: &iotago.TransactionEssence{
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

				inputs, sig, transaction := makeTransaction(testOutput)

				resolvedInputs := vm.ResolvedInputs{InputSet: inputs}
				tx := &iotago.SignedTransaction{
					API:         testAPI,
					Transaction: transaction,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sig},
					},
				}

				_, err := novaVM.Execute(tx.Transaction, resolvedInputs, make(vm.UnlockedIdentities), vm.ExecFuncAddressRestrictions())
				if testInput.wantErr != nil {
					require.ErrorIs(t, err, testInput.wantErr)
					return
				}

				require.NoError(t, err)
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
		Conditions: iotago.BasicOutputUnlockConditions{
			&iotago.AddressUnlockCondition{Address: implicitAccountIdent},
		},
	}
	storageScore := dummyImplicitAccount.StorageScore(testAPI.RentStructure(), nil)
	minAmountImplicitAccount := testAPI.RentStructure().StorageCost() * iotago.BaseToken(storageScore)

	exampleInputs := []TestInput{
		{
			inputID: outputID1,
			input: &iotago.BasicOutput{
				Amount: exampleAmount,
				Conditions: iotago.BasicOutputUnlockConditions{
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
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: implicitAccountIdent},
					},
				},
			},
			keys:    []iotago.AddressKeys{edIdentAddrKeys},
			wantErr: nil,
		},
		{
			name:   "fail - implicit account contains native tokens",
			inputs: exampleInputs,
			outputs: []iotago.Output{
				&iotago.BasicOutput{
					Amount: exampleAmount,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: implicitAccountIdent},
					},
					Features: iotago.BasicOutputFeatures{
						exampleNativeTokenFeature,
					},
				},
			},
			keys:    []iotago.AddressKeys{edIdentAddrKeys},
			wantErr: iotago.ErrAddressCannotReceiveNativeTokens,
		},
		{
			name:   "fail - implicit account contains timelock unlock conditions",
			inputs: exampleInputs,
			outputs: []iotago.Output{
				&iotago.BasicOutput{
					Amount: exampleAmount,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: implicitAccountIdent},
						&iotago.TimelockUnlockCondition{Slot: 500},
					},
				},
			},
			keys:    []iotago.AddressKeys{edIdentAddrKeys},
			wantErr: iotago.ErrAddressCannotReceiveTimelockUnlockCondition,
		},
		{
			name: "ok - implicit account transitioned to account with block issuer feature",
			inputs: []TestInput{
				{
					inputID: outputID1,
					input: &iotago.BasicOutput{
						Amount: exampleAmount,
						Mana:   exampleMana,
						Conditions: iotago.BasicOutputUnlockConditions{
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
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{
							Address: edIdent,
						},
						&iotago.GovernorAddressUnlockCondition{
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
			name: "fail - implicit account transitioned to account without block issuer feature",
			inputs: []TestInput{
				{
					inputID: outputID1,
					input: &iotago.BasicOutput{
						Amount: exampleAmount,
						Mana:   exampleMana,
						Conditions: iotago.BasicOutputUnlockConditions{
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
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{
							Address: edIdent,
						},
						&iotago.GovernorAddressUnlockCondition{
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
						Conditions: iotago.BasicOutputUnlockConditions{
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
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: edIdent},
					},
				},
			},
			keys:    []iotago.AddressKeys{implicitAccountIdentAddrKeys},
			wantErr: iotago.ErrImplicitAccountDestructionDisallowed,
		},
		{
			name: "ok - implicit account with StorageScoreOffsetImplicitAccountCreationAddress can be transitioned",
			inputs: []TestInput{
				{
					inputID: outputID1,
					input: &iotago.BasicOutput{
						Amount: minAmountImplicitAccount,
						Mana:   0,
						Conditions: iotago.BasicOutputUnlockConditions{
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
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{
							Address: edIdent,
						},
						&iotago.GovernorAddressUnlockCondition{
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
			name: "ok - implicit account conversion transaction can contain other non-implicit-account outputs",
			inputs: []TestInput{
				{
					inputID: outputID1,
					input: &iotago.BasicOutput{
						Amount: minAmountImplicitAccount,
						Mana:   0,
						Conditions: iotago.BasicOutputUnlockConditions{
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
						Conditions: iotago.BasicOutputUnlockConditions{
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
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{
							Address: edIdent,
						},
						&iotago.GovernorAddressUnlockCondition{
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
						Conditions: iotago.BasicOutputUnlockConditions{
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
						Conditions: iotago.BasicOutputUnlockConditions{
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
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{
							Address: edIdent,
						},
						&iotago.GovernorAddressUnlockCondition{
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
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{
							Address: edIdent,
						},
						&iotago.GovernorAddressUnlockCondition{
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
						Conditions: iotago.BasicOutputUnlockConditions{
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
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{
							Address: edIdent,
						},
						&iotago.GovernorAddressUnlockCondition{
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
					Conditions: iotago.BasicOutputUnlockConditions{
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
			err := validateAndExecuteSignedTransaction(tx, resolvedInputs)
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
		Conditions: iotago.BasicOutputUnlockConditions{
			&iotago.AddressUnlockCondition{Address: implicitAccountIdent},
		},
	}
	storageScore := implicitAccount.StorageScore(testAPI.RentStructure(), nil)
	minAmount := testAPI.RentStructure().StorageCost() * iotago.BaseToken(storageScore)
	implicitAccount.Amount = minAmount
	depositValidationFunc := iotago.OutputsSyntacticalDepositAmount(testAPI.ProtocolParameters(), testAPI.RentStructure())
	require.NoError(t, depositValidationFunc(0, implicitAccount))

	convertedAccount := &iotago.AccountOutput{
		Amount: implicitAccount.Amount,
		Conditions: iotago.AccountOutputUnlockConditions{
			&iotago.GovernorAddressUnlockCondition{
				Address: &iotago.Ed25519Address{},
			},
			&iotago.StateControllerAddressUnlockCondition{
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
