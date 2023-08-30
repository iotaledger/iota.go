//nolint:forcetypeassert,dupl,nlreturn,scopelint
package stardust_test

import (
	"crypto/ed25519"
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
	"github.com/iotaledger/iota.go/v4/vm"
	"github.com/iotaledger/iota.go/v4/vm/stardust"
)

const (
	OneMi = 1_000_000

	betaPerYear                  float64 = 1 / 3.0
	slotsPerEpochExponent                = 13
	slotDurationSeconds                  = 10
	generationRate                       = 1
	generationRateExponent               = 27
	decayFactorsExponent                 = 32
	decayFactorEpochsSumExponent         = 20
)

var (
	stardustVM = stardust.NewVirtualMachine()

	schedulerRate   iotago.WorkScore = 100000
	testProtoParams                  = iotago.NewV3ProtocolParameters(
		iotago.WithNetworkOptions("test", "test"),
		iotago.WithSupplyOptions(tpkg.TestTokenSupply, 100, 1, 10, 100, 100),
		iotago.WithWorkScoreOptions(1, 100, 500, 20, 20, 20, 20, 100, 100, 100, 200, 4),
		iotago.WithTimeProviderOptions(100, slotDurationSeconds, slotsPerEpochExponent),
		iotago.WithManaOptions(generationRate,
			generationRateExponent,
			tpkg.ManaDecayFactors(betaPerYear, 1<<slotsPerEpochExponent, slotDurationSeconds, decayFactorsExponent),
			decayFactorsExponent,
			tpkg.ManaDecayFactorEpochsSum(betaPerYear, 1<<slotsPerEpochExponent, slotDurationSeconds, decayFactorEpochsSumExponent),
			decayFactorEpochsSumExponent,
		),
		iotago.WithStakingOptions(10),
		iotago.WithLivenessOptions(3, 10, 20, 24),
		iotago.WithCongestionControlOptions(500, 500, 500, 8*schedulerRate, 5*schedulerRate, schedulerRate, 1, 100*iotago.MaxBlockSize),
	)

	testAPI = iotago.V3API(testProtoParams)
)

func TestNFTTransition(t *testing.T) {
	_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()

	inputIDs := tpkg.RandOutputIDs(1)
	inputs := vm.InputSet{
		inputIDs[0]: vm.OutputWithCreationSlot{
			Output: &iotago.NFTOutput{
				Amount: OneMi,
				NFTID:  iotago.NFTID{},
				Conditions: iotago.NFTOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: ident1},
				},
				Features: nil,
			},
			CreationSlot: iotago.SlotIndex(0),
		},
	}

	nftAddr := iotago.NFTAddressFromOutputID(inputIDs[0])
	nftID := nftAddr.NFTID()

	essence := &iotago.TransactionEssence{
		Inputs: inputIDs.UTXOInputs(),
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

	sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddrKeys)
	require.NoError(t, err)

	tx := &iotago.Transaction{
		Essence: essence,
		Unlocks: iotago.Unlocks{
			&iotago.SignatureUnlock{Signature: sigs[0]},
		},
	}
	resolvedInputs := vm.ResolvedInputs{InputSet: inputs}
	require.NoError(t, stardustVM.Execute(tx, &vm.Params{
		API: testAPI,
	}, resolvedInputs))
}

func TestCirculatingSupplyMelting(t *testing.T) {
	_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
	accountIdent1 := tpkg.RandAccountAddress()

	inputIDs := tpkg.RandOutputIDs(3)
	inputs := vm.InputSet{
		inputIDs[0]: vm.OutputWithCreationSlot{
			Output: &iotago.BasicOutput{
				Amount: OneMi,
				Conditions: iotago.BasicOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: ident1},
				},
			},
		},
		inputIDs[1]: vm.OutputWithCreationSlot{
			Output: &iotago.AccountOutput{
				Amount:         OneMi,
				NativeTokens:   nil,
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
		},
		inputIDs[2]: vm.OutputWithCreationSlot{
			Output: &iotago.FoundryOutput{
				Amount:       OneMi,
				NativeTokens: nil,
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
		},
	}

	// set input BasicOutput NativeToken to 50 which get melted
	foundryNativeTokenID := inputs[inputIDs[2]].Output.(*iotago.FoundryOutput).MustNativeTokenID()
	inputs[inputIDs[0]].Output.(*iotago.BasicOutput).NativeTokens = iotago.NativeTokens{
		{
			ID:     foundryNativeTokenID,
			Amount: new(big.Int).SetInt64(50),
		},
	}

	essence := &iotago.TransactionEssence{
		Inputs: inputIDs.UTXOInputs(),
		Outputs: iotago.TxEssenceOutputs{
			&iotago.AccountOutput{
				Amount:         OneMi,
				NativeTokens:   nil,
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
				NativeTokens: nil,
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

	sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddrKeys)
	require.NoError(t, err)

	tx := &iotago.Transaction{
		Essence: essence,
		Unlocks: iotago.Unlocks{
			&iotago.SignatureUnlock{Signature: sigs[0]},
			&iotago.ReferenceUnlock{Reference: 0},
			&iotago.AccountUnlock{Reference: 1},
		},
	}

	resolvedInputs := vm.ResolvedInputs{InputSet: inputs}
	require.NoError(t, stardustVM.Execute(tx, &vm.Params{
		API: testAPI,
	}, resolvedInputs))
}

func TestStardustTransactionExecution(t *testing.T) {
	type test struct {
		name           string
		vmParams       *vm.Params
		resolvedInputs vm.ResolvedInputs
		tx             *iotago.Transaction
		wantErr        error
	}
	tests := []test{
		func() test {
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
				nativeTokenTransfer1                  = tpkg.RandSortNativeTokens(10)
				nativeTokenTransfer2                  = tpkg.RandSortNativeTokens(10)
			)

			var (
				nft1ID = tpkg.Rand32ByteArray()
				nft2ID = tpkg.Rand32ByteArray()
			)

			inputIDs := tpkg.RandOutputIDs(16)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: defaultAmount,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
				inputIDs[1]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount:       defaultAmount,
						NativeTokens: nativeTokenTransfer1,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
					},
				},
				inputIDs[2]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount:       defaultAmount,
						NativeTokens: nativeTokenTransfer2,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
					},
				},
				inputIDs[3]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: defaultAmount,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident2},
							&iotago.ExpirationUnlockCondition{
								ReturnAddress: ident1,
								SlotIndex:     500,
							},
						},
					},
				},
				inputIDs[4]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: defaultAmount,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident2},
							&iotago.TimelockUnlockCondition{
								SlotIndex: 500,
							},
						},
					},
				},
				inputIDs[5]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: defaultAmount + storageDepositReturn,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident2},
							&iotago.StorageDepositReturnUnlockCondition{
								ReturnAddress: ident1,
								Amount:        storageDepositReturn,
							},
							&iotago.TimelockUnlockCondition{
								SlotIndex: 500,
							},
							&iotago.ExpirationUnlockCondition{
								ReturnAddress: ident1,
								SlotIndex:     900,
							},
						},
					},
				},
				inputIDs[6]: vm.OutputWithCreationSlot{
					Output: &iotago.AccountOutput{
						Amount:         defaultAmount,
						NativeTokens:   nil,
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
				},
				inputIDs[7]: vm.OutputWithCreationSlot{
					Output: &iotago.AccountOutput{
						Amount:         defaultAmount + defaultAmount, // to fund also the new account output
						NativeTokens:   nil,
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
				},
				inputIDs[8]: vm.OutputWithCreationSlot{
					Output: &iotago.AccountOutput{
						Amount:         defaultAmount,
						NativeTokens:   nil,
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
				},
				inputIDs[9]: vm.OutputWithCreationSlot{
					Output: &iotago.FoundryOutput{
						Amount:       defaultAmount,
						NativeTokens: nil,
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
				},
				inputIDs[10]: vm.OutputWithCreationSlot{
					Output: &iotago.FoundryOutput{
						Amount:       defaultAmount,
						NativeTokens: nil, // filled out later
						SerialNumber: 2,
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
				},
				inputIDs[11]: vm.OutputWithCreationSlot{
					Output: &iotago.FoundryOutput{
						Amount:       defaultAmount,
						NativeTokens: nil,
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
				},
				inputIDs[12]: vm.OutputWithCreationSlot{
					Output: &iotago.FoundryOutput{
						Amount:       defaultAmount,
						NativeTokens: nil,
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
				},
				inputIDs[13]: vm.OutputWithCreationSlot{
					Output: &iotago.NFTOutput{
						Amount:       defaultAmount,
						NativeTokens: nil,
						NFTID:        nft1ID,
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
				},
				inputIDs[14]: vm.OutputWithCreationSlot{
					Output: &iotago.NFTOutput{
						Amount:       defaultAmount,
						NativeTokens: nil,
						NFTID:        nft2ID,
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
				},
				inputIDs[15]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: defaultAmount,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: iotago.NFTID(nft1ID).ToAddress()},
						},
					},
				},
			}

			foundry1Ident3NativeTokenID := inputs[inputIDs[9]].Output.(*iotago.FoundryOutput).MustNativeTokenID()
			foundry2Ident3NativeTokenID := inputs[inputIDs[10]].Output.(*iotago.FoundryOutput).MustNativeTokenID()
			foundry4Ident3NativeTokenID := inputs[inputIDs[12]].Output.(*iotago.FoundryOutput).MustNativeTokenID()

			newFoundryWithInitialSupply := &iotago.FoundryOutput{
				Amount:       defaultAmount,
				NativeTokens: nil,
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
			newFoundryWithInitialSupply.NativeTokens = iotago.NativeTokens{
				{
					ID:     newFoundryNativeTokenID,
					Amount: big.NewInt(100),
				},
			}

			inputs[inputIDs[10]].Output.(*iotago.FoundryOutput).NativeTokens = iotago.NativeTokens{
				{
					ID:     foundry2Ident3NativeTokenID,
					Amount: big.NewInt(100),
				},
			}

			inputs[inputIDs[12]].Output.(*iotago.FoundryOutput).NativeTokens = iotago.NativeTokens{
				{
					ID:     foundry4Ident3NativeTokenID,
					Amount: big.NewInt(50),
				},
			}

			creationSlot := iotago.SlotIndex(750)
			essence := &iotago.TransactionEssence{
				Inputs:       inputIDs.UTXOInputs(),
				CreationSlot: creationSlot,
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: defaultAmount,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident5},
						},
					},
					&iotago.BasicOutput{
						Amount:       defaultAmount,
						NativeTokens: nativeTokenTransfer1,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident3},
						},
					},
					&iotago.BasicOutput{
						Amount:       defaultAmount,
						NativeTokens: nativeTokenTransfer2,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident4},
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
						NativeTokens:   nil,
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
						NativeTokens:   nil,
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
						NativeTokens:   nil,
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
						Amount: defaultAmount,
						NativeTokens: iotago.NativeTokens{
							{
								ID:     foundry1Ident3NativeTokenID,
								Amount: new(big.Int).SetUint64(100), // freshly minted
							},
						},
						SerialNumber: 1,
						TokenScheme: &iotago.SimpleTokenScheme{
							MintedTokens:  new(big.Int).SetInt64(200),
							MeltedTokens:  big.NewInt(0),
							MaximumSupply: new(big.Int).SetInt64(1000),
						},
						Conditions: iotago.FoundryOutputUnlockConditions{
							&iotago.ImmutableAccountUnlockCondition{Address: iotago.AccountIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AccountAddress)},
						},
						Features: nil,
					},
					&iotago.FoundryOutput{
						Amount: defaultAmount,
						NativeTokens: iotago.NativeTokens{
							{
								ID:     foundry2Ident3NativeTokenID,
								Amount: new(big.Int).SetUint64(50), // melted to 50
							},
						},
						SerialNumber: 2,
						TokenScheme: &iotago.SimpleTokenScheme{
							MintedTokens:  new(big.Int).SetInt64(100),
							MeltedTokens:  big.NewInt(50),
							MaximumSupply: new(big.Int).SetInt64(1000),
						},
						Conditions: iotago.FoundryOutputUnlockConditions{
							&iotago.ImmutableAccountUnlockCondition{Address: iotago.AccountIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AccountAddress)},
						},
						Features: nil,
					},
					&iotago.FoundryOutput{
						Amount:       defaultAmount,
						NativeTokens: nil,
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
						Amount:       defaultAmount,
						NativeTokens: nil,
						NFTID:        iotago.NFTID{},
						Conditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident4},
						},
						Features: nil,
						ImmutableFeatures: iotago.NFTOutputImmFeatures{
							&iotago.MetadataFeature{Data: []byte("immutable metadata")},
						},
					},
					&iotago.NFTOutput{
						Amount:       defaultAmount,
						NativeTokens: nil,
						NFTID:        nft1ID,
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

			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddrKeys, ident2AddrKeys, ident3AddrKeys, ident4AddrKeys)
			require.NoError(t, err)

			return test{
				name: "ok",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet:        inputs,
					CommitmentInput: &iotago.Commitment{Index: creationSlot},
				},
				tx: &iotago.Transaction{
					Essence: essence,
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
		func() test {
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
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.AccountOutput{
						Amount:     100,
						StateIndex: 0,
						AccountID:  accountAddr1.AccountID(),
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident2},
						},
					},
				},
				inputIDs[1]: vm.OutputWithCreationSlot{
					Output: inFoundry,
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
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

			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddressKeys)
			require.NoError(t, err)

			return test{
				name: "fail - changed immutable account address unlock",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
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
		func() test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, ident1, _ := tpkg.RandEd25519Identity()
			_, ident2, ident2AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.AccountOutput{
						Amount:     100,
						StateIndex: 0,
						AccountID:  accountAddr1.AccountID(),
						Features: iotago.AccountOutputFeatures{
							&iotago.BlockIssuerFeature{
								BlockIssuerKeys: iotago.BlockIssuerKeys{},
								ExpirySlot:      100,
							},
						},
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident2},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				CreationSlot: 110,
				ContextInputs: iotago.TxEssenceContextInputs{
					&iotago.BlockIssuanceCreditInput{
						AccountID: accountAddr1.AccountID(),
					},
				},
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.TxEssenceOutputs{
					&iotago.AccountOutput{
						Amount:     100,
						StateIndex: 0,
						AccountID:  accountAddr1.AccountID(),
						Features: iotago.AccountOutputFeatures{
							&iotago.BlockIssuerFeature{
								BlockIssuerKeys: iotago.BlockIssuerKeys{},
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
				Index: 110,
			}

			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident2AddressKeys)
			require.NoError(t, err)

			return test{
				name: "ok - modify block issuer account",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs, BlockIssuanceCreditInputSet: bicInputs, CommitmentInput: commitmentInput},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),

		func() test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, ident1, _ := tpkg.RandEd25519Identity()
			_, ident2, ident2AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.AccountOutput{
						Amount:     100,
						StateIndex: 0,
						AccountID:  accountAddr1.AccountID(),
						Features: iotago.AccountOutputFeatures{
							&iotago.BlockIssuerFeature{
								BlockIssuerKeys: iotago.BlockIssuerKeys{},
								ExpirySlot:      100,
							},
						},
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident2},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				CreationSlot: 110,
				ContextInputs: iotago.TxEssenceContextInputs{
					&iotago.BlockIssuanceCreditInput{
						AccountID: accountAddr1.AccountID(),
					},
				},
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.TxEssenceOutputs{
					&iotago.AccountOutput{
						Amount:     100,
						StateIndex: 0,
						AccountID:  accountAddr1.AccountID(),
						Features: iotago.AccountOutputFeatures{
							&iotago.BlockIssuerFeature{
								BlockIssuerKeys: iotago.BlockIssuerKeys{},
								ExpirySlot:      iotago.SlotIndex(math.MaxUint64),
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
				Index: 110,
			}

			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident2AddressKeys)
			require.NoError(t, err)

			return test{
				name: "ok - set block issuer expiry to max value",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs, BlockIssuanceCreditInputSet: bicInputs, CommitmentInput: commitmentInput},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),

		func() test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.AccountOutput{
						Amount:     100,
						StateIndex: 0,
						AccountID:  accountAddr1.AccountID(),
						Features: iotago.AccountOutputFeatures{
							&iotago.BlockIssuerFeature{
								BlockIssuerKeys: iotago.BlockIssuerKeys{},
								ExpirySlot:      iotago.SlotIndex(math.MaxUint64),
							},
						},
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident1},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				CreationSlot: 110,
				ContextInputs: iotago.TxEssenceContextInputs{
					&iotago.BlockIssuanceCreditInput{
						AccountID: accountAddr1.AccountID(),
					},
				},
				Inputs: inputIDs.UTXOInputs(),
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
				Index: 110,
			}

			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddressKeys)
			require.NoError(t, err)

			return test{
				name: "fail - destroy block issuer account with expiry at slot with max value",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs, BlockIssuanceCreditInputSet: bicInputs, CommitmentInput: commitmentInput},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},

				wantErr: iotago.ErrInvalidBlockIssuerTransition,
			}
		}(),

		func() test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.AccountOutput{
						Amount:     100,
						StateIndex: 0,
						AccountID:  accountAddr1.AccountID(),
						Features: iotago.AccountOutputFeatures{
							&iotago.BlockIssuerFeature{
								BlockIssuerKeys: iotago.BlockIssuerKeys{},
								ExpirySlot:      100,
							},
						},
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident1},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				CreationSlot: 110,
				ContextInputs: iotago.TxEssenceContextInputs{
					&iotago.BlockIssuanceCreditInput{
						AccountID: accountAddr1.AccountID(),
					},
				},
				Inputs: inputIDs.UTXOInputs(),
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
				Index: 110,
			}

			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddressKeys)
			require.NoError(t, err)

			return test{
				name: "ok - destroy block issuer account",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs, BlockIssuanceCreditInputSet: bicInputs, CommitmentInput: commitment},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),

		func() test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.AccountOutput{
						Amount:     100,
						StateIndex: 0,
						AccountID:  accountAddr1.AccountID(),
						Features: iotago.AccountOutputFeatures{
							&iotago.BlockIssuerFeature{
								BlockIssuerKeys: iotago.BlockIssuerKeys{},
								ExpirySlot:      100,
							},
						},
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident1},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				CreationSlot: 110,
				Inputs:       inputIDs.UTXOInputs(),
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
				Index: 110,
			}

			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddressKeys)
			require.NoError(t, err)

			return test{
				name: "fail - destroy block issuer account without supplying BIC",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs, CommitmentInput: commitment},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrBlockIssuanceCreditInputRequired,
			}
		}(),
		func() test {
			accountAddr1 := tpkg.RandAccountAddress()

			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.AccountOutput{
						Amount:     100,
						StateIndex: 0,
						AccountID:  accountAddr1.AccountID(),
						Features: iotago.AccountOutputFeatures{
							&iotago.BlockIssuerFeature{
								BlockIssuerKeys: iotago.BlockIssuerKeys{},
								ExpirySlot:      100,
							},
						},
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident1},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				CreationSlot: 110,
				Inputs:       inputIDs.UTXOInputs(),
				Outputs: iotago.TxEssenceOutputs{
					&iotago.AccountOutput{
						Amount:     100,
						StateIndex: 0,
						AccountID:  accountAddr1.AccountID(),
						Features: iotago.AccountOutputFeatures{
							&iotago.BlockIssuerFeature{
								BlockIssuerKeys: iotago.BlockIssuerKeys{},
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

			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddressKeys)
			require.NoError(t, err)

			return test{
				name: "fail - modify block issuer without supplying BIC",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
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
			err := stardustVM.Execute(tt.tx, tt.vmParams, tt.resolvedInputs)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

// TODO: add test case for transaction with context inputs.
func TestTxSemanticInputUnlocks(t *testing.T) {
	type test struct {
		name           string
		vmParams       *vm.Params
		resolvedInputs vm.ResolvedInputs
		tx             *iotago.Transaction
		wantErr        error
	}
	tests := []test{
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, ident2AddrKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(8)
			accountIdent1 := iotago.AccountAddressFromOutputID(inputIDs[1])
			nftIdent1 := tpkg.RandNFTAddress()

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
				inputIDs[1]: vm.OutputWithCreationSlot{
					Output: &iotago.AccountOutput{
						Amount:    100,
						AccountID: iotago.AccountID{}, // empty on purpose as validation should resolve
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident1},
						},
					},
				},
				inputIDs[2]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: &accountIdent1},
						},
					},
				},
				inputIDs[3]: vm.OutputWithCreationSlot{
					Output: &iotago.NFTOutput{
						Amount: 100,
						NFTID:  nftIdent1.NFTID(),
						Conditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: &accountIdent1},
						},
					},
				},
				inputIDs[4]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: nftIdent1},
						},
					},
				},
				// unlockable by sender as expired
				inputIDs[5]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
							&iotago.ExpirationUnlockCondition{
								ReturnAddress: ident2,
								SlotIndex:     5,
							},
						},
					},
				},
				// not unlockable by sender as not expired
				inputIDs[6]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
							&iotago.ExpirationUnlockCondition{
								ReturnAddress: ident2,
								SlotIndex:     30,
							},
						},
					},
				},
				inputIDs[7]: vm.OutputWithCreationSlot{
					Output: &iotago.FoundryOutput{
						Amount:       100,
						SerialNumber: 0,
						TokenScheme: &iotago.SimpleTokenScheme{
							MintedTokens:  new(big.Int).SetInt64(100),
							MeltedTokens:  big.NewInt(0),
							MaximumSupply: new(big.Int).SetInt64(1000),
						},
						Conditions: iotago.FoundryOutputUnlockConditions{
							&iotago.ImmutableAccountUnlockCondition{Address: &accountIdent1},
						},
					},
				},
			}

			creationSlot := iotago.SlotIndex(10)
			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
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
				CreationSlot: creationSlot,
			}

			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddrKeys, ident2AddrKeys)
			require.NoError(t, err)

			return test{
				name: "ok",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet: inputs,
					CommitmentInput: &iotago.Commitment{
						Index: iotago.SlotIndex(0),
					},
				},
				tx: &iotago.Transaction{
					Essence: essence,
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
		func() test {
			ident1Sk, ident1, _ := tpkg.RandEd25519Identity()
			_, _, ident2AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident2AddrKeys)
			require.NoError(t, err)

			copy(sigs[0].(*iotago.Ed25519Signature).PublicKey[:], ident1Sk.Public().(ed25519.PublicKey))

			return test{
				name: "fail - invalid signature",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrEd25519SignatureInvalid,
			}
		}(),
		func() test {
			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
				inputIDs[1]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddressKeys)
			require.NoError(t, err)

			return test{
				name: "fail - should contain reference unlock",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrInvalidInputUnlock,
			}
		}(),
		func() test {
			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)

			accountIdent1 := iotago.AccountAddressFromOutputID(inputIDs[0])
			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.AccountOutput{
						Amount:    100,
						AccountID: iotago.AccountID{},
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident1},
						},
					},
				},
				inputIDs[1]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: &accountIdent1},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddressKeys)
			require.NoError(t, err)

			return test{
				name: "fail - should contain account unlock",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.ReferenceUnlock{Reference: 0},
					},
				},
				wantErr: iotago.ErrInvalidInputUnlock,
			}
		}(),
		func() test {
			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)

			nftIdent1 := iotago.NFTAddressFromOutputID(inputIDs[0])
			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.NFTOutput{
						Amount: 100,
						NFTID:  iotago.NFTID{},
						Conditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
				inputIDs[1]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: &nftIdent1},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddressKeys)
			require.NoError(t, err)

			return test{
				name: "fail - should contain NFT unlock",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.ReferenceUnlock{Reference: 0},
					},
				},
				wantErr: iotago.ErrInvalidInputUnlock,
			}
		}(),
		func() test {
			inputIDs := tpkg.RandOutputIDs(2)

			nftIdent1 := iotago.NFTAddressFromOutputID(inputIDs[0])
			nftIdent2 := iotago.NFTAddressFromOutputID(inputIDs[1])

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.NFTOutput{
						Amount: 100,
						NFTID:  nftIdent1.NFTID(),
						Conditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: &nftIdent2},
						},
					},
				},
				inputIDs[1]: vm.OutputWithCreationSlot{
					Output: &iotago.NFTOutput{
						Amount: 100,
						NFTID:  nftIdent2.NFTID(),
						Conditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: &nftIdent2},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}
			_, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI))
			require.NoError(t, err)
			return test{
				name: "fail - circular NFT unlock",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.NFTUnlock{Reference: 1},
						&iotago.NFTUnlock{Reference: 0},
					},
				},
				wantErr: iotago.ErrInvalidInputUnlock,
			}
		}(),
		func() test {
			_, ident1, _ := tpkg.RandEd25519Identity()
			_, ident2, ident2AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
							&iotago.ExpirationUnlockCondition{
								ReturnAddress: ident2,
								SlotIndex:     20,
							},
						},
					},
				},
			}

			creationSlot := iotago.SlotIndex(5)
			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs(), CreationSlot: creationSlot}

			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident2AddressKeys)
			require.NoError(t, err)

			return test{
				name: "fail - sender can not unlock yet",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet: inputs,
					CommitmentInput: &iotago.Commitment{
						Index: iotago.SlotIndex(0),
					},
				},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrExpirationConditionUnlockFailed,
			}
		}(),
		func() test {
			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()
			_, ident2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
							&iotago.ExpirationUnlockCondition{
								ReturnAddress: ident2,
								SlotIndex:     10,
							},
						},
					},
				},
			}

			creationSlot := iotago.SlotIndex(10)
			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs(), CreationSlot: creationSlot}

			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddressKeys)
			require.NoError(t, err)

			return test{
				name: "fail - receiver can not unlock anymore",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet: inputs,
					CommitmentInput: &iotago.Commitment{
						Index: creationSlot,
					},
				},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrEd25519PubKeyAndAddrMismatch,
			}
		}(),
		func() test {
			_, ident1, ident1AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(3)

			var (
				accountAddr1 = tpkg.RandAccountAddress()
				accountAddr2 = tpkg.RandAccountAddress()
				accountAddr3 = tpkg.RandAccountAddress()
			)

			inputs := vm.InputSet{
				// owned by ident1
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.AccountOutput{
						Amount:    100,
						AccountID: accountAddr1.AccountID(),
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident1},
						},
					},
				},
				// owned by account1
				inputIDs[1]: vm.OutputWithCreationSlot{
					Output: &iotago.AccountOutput{
						Amount:    100,
						AccountID: accountAddr2.AccountID(),
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: accountAddr1},
							&iotago.GovernorAddressUnlockCondition{Address: accountAddr1},
						},
					},
				},
				// owned by account1
				inputIDs[2]: vm.OutputWithCreationSlot{
					Output: &iotago.AccountOutput{
						Amount:    100,
						AccountID: accountAddr3.AccountID(),
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: accountAddr1},
							&iotago.GovernorAddressUnlockCondition{Address: accountAddr1},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddressKeys)
			require.NoError(t, err)

			return test{
				name: "fail - referencing other account unlocked by source account",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
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
		func() test {
			_, ident1, _ := tpkg.RandEd25519Identity()
			_, ident2, ident2AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)

			accountAddr1 := tpkg.RandAccountAddress()

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.AccountOutput{
						Amount:    100,
						AccountID: accountAddr1.AccountID(),
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident2},
						},
					},
				},
				inputIDs[1]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: accountAddr1},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
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

			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident2AddressKeys)
			require.NoError(t, err)

			return test{
				name: "fail - account output not state transitioning",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.AccountUnlock{Reference: 0},
					},
				},
				wantErr: iotago.ErrInvalidInputUnlock,
			}
		}(),
		func() test {
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
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.AccountOutput{
						Amount:     100,
						StateIndex: 0,
						AccountID:  accountAddr1.AccountID(),
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident2},
						},
					},
				},
				inputIDs[1]: vm.OutputWithCreationSlot{
					Output: foundryOutput,
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
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

			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddressKeys)
			require.NoError(t, err)

			return test{
				name: "fail - wrong unlock for foundry",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
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
			err := stardustVM.Execute(tt.tx, tt.vmParams, tt.resolvedInputs, vm.ExecFuncInputUnlocks())
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
		tx             *iotago.Transaction
		wantErr        error
	}
	tests := []test{
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, ident2AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(3)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
				// unlocked by ident1 as it is not expired
				inputIDs[1]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 500,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
							&iotago.StorageDepositReturnUnlockCondition{
								ReturnAddress: ident2,
								Amount:        420,
							},
							&iotago.ExpirationUnlockCondition{
								ReturnAddress: ident2,
								SlotIndex:     30,
							},
						},
					},
				},
				// unlocked by ident2 as it is expired
				inputIDs[2]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 500,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
							&iotago.StorageDepositReturnUnlockCondition{
								ReturnAddress: ident2,
								Amount:        420,
							},
							&iotago.ExpirationUnlockCondition{
								ReturnAddress: ident2,
								SlotIndex:     2,
							},
						},
					},
				},
			}

			creationSlot := iotago.SlotIndex(5)
			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
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
				CreationSlot: creationSlot,
			}
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddrKeys, ident2AddrKeys)
			require.NoError(t, err)

			return test{
				name: "ok",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet: inputs,
					CommitmentInput: &iotago.Commitment{
						Index: creationSlot,
					},
				},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.ReferenceUnlock{Reference: 0},
						&iotago.SignatureUnlock{Signature: sigs[1]},
					},
				},
				wantErr: nil,
			}
		}(),
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 1000,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
							&iotago.StorageDepositReturnUnlockCondition{
								ReturnAddress: ident2,
								Amount:        420,
							},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
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
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "ok - more storage deposit returned via more outputs",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 50,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
				CreationSlot: 5,
			}
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "fail - unbalanced, more on output than input",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrInputOutputSumMismatch,
			}
		}(),
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 50,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
				CreationSlot: 5,
			}
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "fail - unbalanced, more on input than output",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrInputOutputSumMismatch,
			}
		}(),
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
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
								SlotIndex:     30,
							},
						},
					},
				},
			}

			creationSlot := iotago.SlotIndex(5)
			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 500,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
				CreationSlot: creationSlot,
			}
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "fail - return not fulfilled",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet: inputs,
					CommitmentInput: &iotago.Commitment{
						Index: creationSlot,
					},
				},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrReturnAmountNotFulFilled,
			}
		}(),
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 500,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
							&iotago.StorageDepositReturnUnlockCondition{
								ReturnAddress: ident2,
								Amount:        420,
							},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
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
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "fail - storage deposit return not basic output",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrReturnAmountNotFulFilled,
			}
		}(),
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 500,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
							&iotago.StorageDepositReturnUnlockCondition{
								ReturnAddress: ident2,
								Amount:        420,
							},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
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
								SlotIndex:     10,
							},
						},
					},
				},
			}
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "fail - storage deposit return has additional unlocks",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrReturnAmountNotFulFilled,
			}
		}(),
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 500,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
							&iotago.StorageDepositReturnUnlockCondition{
								ReturnAddress: ident2,
								Amount:        420,
							},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
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
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "fail - storage deposit return has feature",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrReturnAmountNotFulFilled,
			}
		}(),
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)
			ntID := tpkg.Rand38ByteArray()

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 500,
						NativeTokens: iotago.NativeTokens{
							&iotago.NativeToken{
								ID:     ntID,
								Amount: new(big.Int).SetUint64(1000),
							},
						},
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
							&iotago.StorageDepositReturnUnlockCondition{
								ReturnAddress: ident2,
								Amount:        420,
							},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 80,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					&iotago.BasicOutput{
						Amount: 420,
						NativeTokens: iotago.NativeTokens{
							&iotago.NativeToken{
								ID:     ntID,
								Amount: new(big.Int).SetUint64(1000),
							},
						},
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
					},
				},
			}
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "fail - storage deposit return has native tokens",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
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
			err := stardustVM.Execute(tt.tx, tt.vmParams, tt.resolvedInputs, vm.ExecFuncInputUnlocks(), vm.ExecFuncBalancedBaseTokens())
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
		tx             *iotago.Transaction
		wantErr        error
	}
	tests := []test{
		func() test {
			inputIDs := tpkg.RandOutputIDs(2)

			ntCount := 10
			nativeTokens := tpkg.RandSortNativeTokens(ntCount)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount:       100,
						NativeTokens: nativeTokens[:ntCount/2],
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
				inputIDs[1]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount:       100,
						NativeTokens: nativeTokens[ntCount/2:],
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount:       200,
						NativeTokens: nativeTokens,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}

			return test{
				name: "ok",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{},
				},
				wantErr: nil,
			}
		}(),
		func() test {
			inputIDs := tpkg.RandOutputIDs(iotago.MaxNativeTokensCount)
			nativeToken := tpkg.RandNativeToken()

			inputs := vm.InputSet{}
			for i := 0; i < iotago.MaxNativeTokensCount; i++ {
				inputs[inputIDs[i]] = vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 100,
						NativeTokens: []*iotago.NativeToken{
							{
								ID:     nativeToken.ID,
								Amount: big.NewInt(1),
							},
						},
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				}
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 200,
						NativeTokens: []*iotago.NativeToken{
							{
								ID:     nativeToken.ID,
								Amount: big.NewInt(iotago.MaxNativeTokensCount),
							},
						},
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}

			return test{
				name: "ok - exceeds limit (in+out) but same native token",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{},
				},
				wantErr: nil,
			}
		}(),
		func() test {
			inputIDs := tpkg.RandOutputIDs(1)

			inCount := 20
			outCount := 250

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount:       100,
						NativeTokens: tpkg.RandSortNativeTokens(inCount),
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount:       200,
						NativeTokens: tpkg.RandSortNativeTokens(outCount),
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}

			return test{
				name: "fail - exceeds limit (in+out)",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{},
				},
				wantErr: iotago.ErrMaxNativeTokensCountExceeded,
			}
		}(),
		func() test {
			numDistinctNTs := iotago.MaxNativeTokensCount + 1
			inputIDs := tpkg.RandOutputIDs(uint16(numDistinctNTs))

			inputs := vm.InputSet{}
			for i := 0; i < numDistinctNTs; i++ {
				inputs[inputIDs[i]] = vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount:       100,
						NativeTokens: tpkg.RandSortNativeTokens(1),
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				}
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 100 * iotago.BaseToken(numDistinctNTs),
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}

			return test{
				name: "fail - too many on input side already",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{},
				},
				wantErr: iotago.ErrMaxNativeTokensCountExceeded,
			}
		}(),
		func() test {
			numDistinctNTs := iotago.MaxNativeTokensCount + 1
			tokens := tpkg.RandSortNativeTokens(numDistinctNTs)
			inputIDs := tpkg.RandOutputIDs(uint16(numDistinctNTs))

			inputs := vm.InputSet{}
			for i := 0; i < numDistinctNTs; i++ {
				inputs[inputIDs[i]] = vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				}
			}

			outs := make(iotago.TxEssenceOutputs, numDistinctNTs)
			for i := range outs {
				outs[i] = &iotago.BasicOutput{
					Amount:       100,
					NativeTokens: iotago.NativeTokens{tokens[i]},
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				}
			}

			essence := &iotago.TransactionEssence{
				Inputs:  inputIDs.UTXOInputs(),
				Outputs: outs,
			}

			return test{
				name: "fail - too many on output side already",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{},
				},
				wantErr: iotago.ErrMaxNativeTokensCountExceeded,
			}
		}(),
		func() test {
			numDistinctNTs := iotago.MaxNativeTokensCount
			tokens := tpkg.RandSortNativeTokens(numDistinctNTs)
			inputIDs := tpkg.RandOutputIDs(iotago.MaxInputsCount)

			inputs := vm.InputSet{}
			for i := 0; i < iotago.MaxInputsCount; i++ {
				inputs[inputIDs[i]] = vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount:       100,
						NativeTokens: tokens,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				}
			}

			outputs := make(iotago.TxEssenceOutputs, iotago.MaxOutputsCount)
			for i := range outputs {
				outputs[i] = &iotago.BasicOutput{
					Amount:       100,
					NativeTokens: tokens,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				}
			}

			essence := &iotago.TransactionEssence{
				Inputs:  inputIDs.UTXOInputs(),
				Outputs: outputs,
			}

			return test{
				name: "ok - most possible tokens in a tx",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{},
				},
				wantErr: nil,
			}
		}(),
		func() test {
			numDistinctNTs := iotago.MaxNativeTokensCount
			tokens := tpkg.RandSortNativeTokens(numDistinctNTs)
			inputIDs := tpkg.RandOutputIDs(iotago.MaxInputsCount)

			inputs := vm.InputSet{}
			for i := 0; i < iotago.MaxInputsCount; i++ {
				inputs[inputIDs[i]] = vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount:       100,
						NativeTokens: tokens,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				}
			}

			outputs := make(iotago.TxEssenceOutputs, iotago.MaxOutputsCount)
			for i := range outputs {
				outputs[i] = &iotago.BasicOutput{
					Amount:       100,
					NativeTokens: tokens,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				}
			}

			// add one more distinct native token to the last output
			oneMore := tokens.Clone()
			oneMore[len(oneMore)-1] = tpkg.RandNativeToken()

			outputs[iotago.MaxOutputsCount-1] = &iotago.BasicOutput{
				Amount:       100,
				NativeTokens: oneMore,
				Conditions: iotago.BasicOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs:  inputIDs.UTXOInputs(),
				Outputs: outputs,
			}

			return test{
				name: "fail - max nt count just exceeded",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{},
				},
				wantErr: iotago.ErrMaxNativeTokensCountExceeded,
			}
		}(),
		func() test {
			inputIDs := tpkg.RandOutputIDs(1)

			ntCount := 10
			nativeTokens := tpkg.RandSortNativeTokens(ntCount)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount:       100,
						NativeTokens: nativeTokens,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}

			// unbalance by making one token be excess on the output side
			cpyNativeTokens := nativeTokens.Clone()
			amountToModify := cpyNativeTokens[ntCount/2].Amount
			amountToModify.Add(amountToModify, big.NewInt(1))

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount:       100,
						NativeTokens: cpyNativeTokens,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}

			return test{
				name: "fail - unbalanced on output",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{},
				},
				wantErr: iotago.ErrNativeTokenSumUnbalanced,
			}
		}(),
		func() test {
			inputIDs := tpkg.RandOutputIDs(3)

			ntCount := 20
			nativeTokens := tpkg.RandSortNativeTokens(ntCount)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount:       100,
						NativeTokens: nativeTokens[:ntCount/2],
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
				inputIDs[1]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount:       100,
						NativeTokens: nativeTokens[ntCount/2:],
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
				inputIDs[2]: vm.OutputWithCreationSlot{
					Output: inUnrelatedFoundryOutput,
				},
			}

			// add a new token to the output side
			cpyNativeTokens := nativeTokens.Clone()
			cpyNativeTokens = append(cpyNativeTokens, tpkg.RandNativeToken())

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount:       100,
						NativeTokens: cpyNativeTokens[:ntCount/2],
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
					&iotago.BasicOutput{
						Amount:       100,
						NativeTokens: cpyNativeTokens[ntCount/2:],
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
					outUnrelatedFoundryOutput,
				},
			}

			return test{
				name: "fail - unbalanced with unrelated foundry in term of new output tokens",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{},
				},
				wantErr: iotago.ErrNativeTokenSumUnbalanced,
			}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := stardustVM.Execute(tt.tx, tt.vmParams, tt.resolvedInputs, vm.ExecFuncBalancedNativeTokens())
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
		tx             *iotago.Transaction
		wantErr        error
	}
	tests := []test{
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)
			nftAddr := tpkg.RandNFTAddress()

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
				inputIDs[1]: vm.OutputWithCreationSlot{
					Output: &iotago.NFTOutput{
						Amount: 100,
						NFTID:  nftAddr.NFTID(),
						Conditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
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
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "ok",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.ReferenceUnlock{Reference: 0},
					},
				},
				wantErr: nil,
			}
		}(),
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
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
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "fail - sender not unlocked",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrSenderFeatureNotUnlocked,
			}
		}(),
		func() test {
			_, stateController, _ := tpkg.RandEd25519Identity()
			_, governor, governorAddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)
			accountAddr := tpkg.RandAccountAddress()
			accountID := accountAddr.AccountID()

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.AccountOutput{
						Amount:    100,
						AccountID: accountID,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
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
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), governorAddrKeys)
			require.NoError(t, err)

			return test{
				name: "fail - sender not unlocked due to governance transition",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrSenderFeatureNotUnlocked,
			}
		}(),
		func() test {
			_, stateController, stateControllerAddrKeys := tpkg.RandEd25519Identity()
			_, governor, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)
			accountAddr := tpkg.RandAccountAddress()
			accountID := accountAddr.AccountID()
			currentStateIndex := uint32(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.AccountOutput{
						Amount:     100,
						AccountID:  accountID,
						StateIndex: currentStateIndex,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
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
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), stateControllerAddrKeys)
			require.NoError(t, err)

			return test{
				name: "ok - account addr unlocked with state transition",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),
		func() test {
			_, stateController, _ := tpkg.RandEd25519Identity()
			_, governor, governorAddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)
			accountAddr := tpkg.RandAccountAddress()
			accountID := accountAddr.AccountID()

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.AccountOutput{
						Amount:    100,
						AccountID: accountID,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
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
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), governorAddrKeys)
			require.NoError(t, err)

			return test{
				name: "ok - sender is governor address",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := stardustVM.Execute(tt.tx, tt.vmParams, tt.resolvedInputs, vm.ExecFuncInputUnlocks(), vm.ExecFuncSenderUnlocked())
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
		tx             *iotago.Transaction
		wantErr        error
	}
	tests := []test{
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, stateController, _ := tpkg.RandEd25519Identity()
			_, governor, governorAddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)
			accountAddr := tpkg.RandAccountAddress()
			accountID := accountAddr.AccountID()

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.AccountOutput{
						Amount:    100,
						AccountID: accountID,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
					},
				},
				inputIDs[1]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
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
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), governorAddrKeys, ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "fail - issuer not unlocked due to governance transition",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.SignatureUnlock{Signature: sigs[1]},
					},
				},
				wantErr: iotago.ErrIssuerFeatureNotUnlocked,
			}
		}(),
		func() test {
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
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.AccountOutput{
						Amount:     100,
						AccountID:  accountID,
						StateIndex: currentStateIndex,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
					},
				},
				inputIDs[1]: vm.OutputWithCreationSlot{
					Output: &iotago.NFTOutput{
						Amount: 900,
						NFTID:  nftAddr.NFTID(),
						Conditions: iotago.NFTOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
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
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), stateControllerAddrKeys, ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "ok - issuer unlocked with state transition",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.SignatureUnlock{Signature: sigs[1]},
					},
				},
				wantErr: nil,
			}
		}(),
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, stateController, _ := tpkg.RandEd25519Identity()
			_, governor, governorAddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)
			accountAddr := tpkg.RandAccountAddress()
			accountID := accountAddr.AccountID()

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.AccountOutput{
						Amount:    100,
						AccountID: accountID,
						Conditions: iotago.AccountOutputUnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
					},
				},
				inputIDs[1]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
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
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), governorAddrKeys, ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "ok - issuer is the governor",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
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
			err := stardustVM.Execute(tt.tx, tt.vmParams, tt.resolvedInputs)
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
		tx             *iotago.Transaction
		wantErr        error
	}
	tests := []test{
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
							&iotago.TimelockUnlockCondition{
								SlotIndex: 5,
							},
						},
					},
				},
			}

			creationSlot := iotago.SlotIndex(10)
			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs(), CreationSlot: creationSlot}
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "ok",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet: inputs,
					CommitmentInput: &iotago.Commitment{
						Index: creationSlot,
					},
				},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
							&iotago.TimelockUnlockCondition{
								SlotIndex: 25,
							},
						},
					},
				},
			}

			creationSlot := iotago.SlotIndex(10)
			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs(), CreationSlot: 10}
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "fail - timelock not expired",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet: inputs,
					CommitmentInput: &iotago.Commitment{
						Index: creationSlot,
					},
				},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrTimelockNotExpired,
			}
		}(),
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
							&iotago.TimelockUnlockCondition{
								SlotIndex: 1337,
							},
						},
					},
				},
			}

			creationSlot := iotago.SlotIndex(666)
			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs(), CreationSlot: creationSlot}
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "fail - timelock not expired",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet: inputs,
					CommitmentInput: &iotago.Commitment{
						Index: creationSlot,
					},
				},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrTimelockNotExpired,
			}
		}(),
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
							&iotago.TimelockUnlockCondition{
								SlotIndex: 1000,
							},
						},
					},
				},
			}

			creationSlot := iotago.SlotIndex(1005)
			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs(), CreationSlot: creationSlot}
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "fail - no commitment input for timelock",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{
					InputSet: inputs,
				},
				tx: &iotago.Transaction{
					Essence: essence,
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
			err := stardustVM.Execute(tt.tx, tt.vmParams, tt.resolvedInputs, vm.ExecFuncTimelocks())
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
		tx             *iotago.Transaction
		wantErr        error
	}
	tests := []test{
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: OneMi,
						Mana:   math.MaxUint64,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					CreationSlot: 10,
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: OneMi,
						Mana: func() iotago.Mana {
							var slotIndexCreated iotago.SlotIndex = 10
							slotIndexTarget := 10 + 100*testProtoParams.ParamEpochDurationInSlots()

							input := inputs[inputIDs[0]]
							excessBaseTokens := input.Output.BaseTokenAmount() - testProtoParams.RentStructure().MinDeposit(input.Output)
							potentialMana, err := testProtoParams.ManaDecayProvider().ManaGenerationWithDecay(excessBaseTokens, slotIndexCreated, slotIndexTarget)
							require.NoError(t, err)

							storedMana, err := testProtoParams.ManaDecayProvider().ManaWithDecay(math.MaxUint64, slotIndexCreated, slotIndexTarget)
							require.NoError(t, err)

							return potentialMana + storedMana
						}(),
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
				CreationSlot: 10 + 100*testProtoParams.ParamEpochDurationInSlots(),
			}
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "ok - stored Mana only without allotment",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: OneMi,
						Mana:   math.MaxUint64,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					CreationSlot: 10,
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: OneMi,
						Mana: func() iotago.Mana {
							var slotIndexCreated iotago.SlotIndex = 10
							slotIndexTarget := 10 + 100*testProtoParams.ParamEpochDurationInSlots()

							input := inputs[inputIDs[0]]
							excessBaseTokens := input.Output.BaseTokenAmount() - testProtoParams.RentStructure().MinDeposit(input.Output)
							potentialMana, err := testProtoParams.ManaDecayProvider().ManaGenerationWithDecay(excessBaseTokens, slotIndexCreated, slotIndexTarget)
							require.NoError(t, err)

							storedMana, err := testProtoParams.ManaDecayProvider().ManaWithDecay(math.MaxUint64, slotIndexCreated, slotIndexTarget)
							require.NoError(t, err)

							// generated mana + decay - allotment
							return potentialMana + storedMana - 50
						}(),
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
				Allotments: iotago.Allotments{
					&iotago.Allotment{Value: 50},
				},
				CreationSlot: 10 + 100*testProtoParams.ParamEpochDurationInSlots(),
			}
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "ok - stored and allotted",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 5,
						Mana:   10,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					CreationSlot: 20,
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 5,
						Mana:   35,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
				CreationSlot: 15,
			}
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "fail - input created after tx",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrInputCreationAfterTxCreation,
			}
		}(),
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 5,
						Mana:   10,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					CreationSlot: 15,
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 5,
						Mana:   10,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
				CreationSlot: 15,
			}
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "ok - input created in same slot as tx",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: nil,
			}
		}(),
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 5,
						Mana:   math.MaxUint64,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					CreationSlot: 15,
				},
				inputIDs[1]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 5,
						Mana:   10,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					CreationSlot: 15,
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.TxEssenceOutputs{
					&iotago.BasicOutput{
						Amount: 5,
						Mana:   9,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
				CreationSlot: 15,
			}
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "fail - mana overflow on the input side sum",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.ReferenceUnlock{Reference: 0},
					},
				},
				wantErr: iotago.ErrManaOverflow,
			}
		}(),
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := vm.InputSet{
				inputIDs[0]: vm.OutputWithCreationSlot{
					Output: &iotago.BasicOutput{
						Amount: 5,
						Mana:   10,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					CreationSlot: 15,
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
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
						Mana:   math.MaxUint64,
						Conditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
				CreationSlot: 15,
			}
			sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "fail - mana overflow on the output side sum",
				vmParams: &vm.Params{
					API: testAPI,
				},
				resolvedInputs: vm.ResolvedInputs{InputSet: inputs},
				tx: &iotago.Transaction{
					Essence: essence,
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
			err := stardustVM.Execute(tt.tx, tt.vmParams, tt.resolvedInputs, vm.ExecFuncInputUnlocks(), vm.ExecFuncBalancedMana())
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
		inputIDs[0]: vm.OutputWithCreationSlot{
			Output: &iotago.AccountOutput{
				Amount:         OneMi * 10,
				NativeTokens:   nil,
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
		},
	}

	essence := &iotago.TransactionEssence{
		Inputs: inputIDs.UTXOInputs(),
		Outputs: iotago.TxEssenceOutputs{
			&iotago.AccountOutput{
				Amount:         OneMi * 5,
				NativeTokens:   nil,
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
				Amount:       OneMi * 5,
				NativeTokens: nil,
				Mana:         manaRewardAmount,
				Conditions: iotago.BasicOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: accountIdent},
				},
				Features: nil,
			},
		},
	}

	sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), identAddrKeys)
	require.NoError(t, err)

	tx := &iotago.Transaction{
		Essence: essence,
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
			Index: currentSlot,
		},
	}
	require.NoError(t, stardustVM.Execute(tx, &vm.Params{
		API: testAPI,
	}, resolvedInputs))
}

func TestManaRewardsClaimingDelegation(t *testing.T) {
	_, ident, identAddrKeys := tpkg.RandEd25519Identity()
	emptyAccountAddress := iotago.AccountAddress(iotago.EmptyAccountID())

	const manaRewardAmount iotago.Mana = 200
	currentSlot := 20 * testProtoParams.ParamEpochDurationInSlots()
	currentEpoch := testProtoParams.TimeProvider().EpochFromSlot(currentSlot)

	inputIDs := tpkg.RandOutputIDs(1)
	inputs := vm.InputSet{
		inputIDs[0]: vm.OutputWithCreationSlot{
			Output: &iotago.DelegationOutput{
				Amount:           OneMi * 10,
				DelegatedAmount:  OneMi * 10,
				DelegationID:     iotago.EmptyDelegationID(),
				ValidatorAddress: &emptyAccountAddress,
				StartEpoch:       currentEpoch,
				EndEpoch:         currentEpoch + 5,
				Conditions: iotago.DelegationOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: ident},
				},
			},
		},
	}
	delegationID := iotago.DelegationIDFromOutputID(inputIDs[0])

	essence := &iotago.TransactionEssence{
		Inputs: inputIDs.UTXOInputs(),
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
		CreationSlot: currentSlot,
	}

	sigs, err := essence.Sign(testAPI, inputIDs.OrderedSet(inputs.OutputSet()).MustCommitment(testAPI), identAddrKeys)
	require.NoError(t, err)

	tx := &iotago.Transaction{
		Essence: essence,
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
	require.NoError(t, stardustVM.Execute(tx, &vm.Params{
		API: testAPI,
	}, resolvedInputs))
}
