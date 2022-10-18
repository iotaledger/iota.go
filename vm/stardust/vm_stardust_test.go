package stardust_test

import (
	"crypto/ed25519"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/iota.go/v3/vm"
	"github.com/iotaledger/iota.go/v3/vm/stardust"
)

const (
	OneMi = 1_000_000
)

var stardustVM = stardust.NewVirtualMachine()

func TestNFTTransition(t *testing.T) {
	_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()

	inputIDs := tpkg.RandOutputIDs(1)
	inputs := iotago.OutputSet{
		inputIDs[0]: &iotago.NFTOutput{
			Amount: OneMi,
			NFTID:  iotago.NFTID{},
			Conditions: iotago.UnlockConditions[iotago.NFTUnlockCondition]{
				&iotago.AddressUnlockCondition{Address: ident1},
			},
			Features: nil,
		},
	}

	nftAddr := iotago.NFTAddressFromOutputID(inputIDs[0])
	nftID := nftAddr.NFTID()

	essence := &iotago.TransactionEssence{
		Inputs: inputIDs.UTXOInputs(),
		Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
			&iotago.NFTOutput{
				Amount: OneMi,
				NFTID:  nftID,
				Conditions: iotago.UnlockConditions[iotago.NFTUnlockCondition]{
					&iotago.AddressUnlockCondition{Address: ident1},
				},
				Features: nil,
			},
		},
	}

	sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys)
	require.NoError(t, err)

	tx := &iotago.Transaction{
		Essence: essence,
		Unlocks: iotago.Unlocks{
			&iotago.SignatureUnlock{Signature: sigs[0]},
		},
	}

	require.NoError(t, stardustVM.Execute(tx, &vm.Paras{}, inputs))
}

func TestCirculatingSupplyMelting(t *testing.T) {
	_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
	aliasIdent1 := tpkg.RandAliasAddress()

	inputIDs := tpkg.RandOutputIDs(3)
	inputs := iotago.OutputSet{
		inputIDs[0]: &iotago.BasicOutput{
			Amount: OneMi,
			Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
				&iotago.AddressUnlockCondition{Address: ident1},
			},
		},
		inputIDs[1]: &iotago.AliasOutput{
			Amount:         OneMi,
			NativeTokens:   nil,
			AliasID:        aliasIdent1.AliasID(),
			StateIndex:     1,
			StateMetadata:  nil,
			FoundryCounter: 1,
			Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
				&iotago.StateControllerAddressUnlockCondition{Address: ident1},
				&iotago.GovernorAddressUnlockCondition{Address: ident1},
			},
			Features: nil,
		},
		inputIDs[2]: &iotago.FoundryOutput{
			Amount:       OneMi,
			NativeTokens: nil,
			SerialNumber: 1,
			TokenScheme: &iotago.SimpleTokenScheme{
				MintedTokens:  big.NewInt(50),
				MeltedTokens:  big.NewInt(0),
				MaximumSupply: big.NewInt(50),
			},
			Conditions: iotago.UnlockConditions[iotago.FoundryUnlockCondition]{
				&iotago.ImmutableAliasUnlockCondition{Address: aliasIdent1},
			},
			Features: nil,
		},
	}

	// set input BasicOutput NativeToken to 50 which get melted
	foundryNativeTokenID := inputs[inputIDs[2]].(*iotago.FoundryOutput).MustNativeTokenID()
	inputs[inputIDs[0]].(*iotago.BasicOutput).NativeTokens = iotago.NativeTokens{
		{
			ID:     foundryNativeTokenID,
			Amount: new(big.Int).SetInt64(50),
		},
	}

	essence := &iotago.TransactionEssence{
		Inputs: inputIDs.UTXOInputs(),
		Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
			&iotago.AliasOutput{
				Amount:         OneMi,
				NativeTokens:   nil,
				AliasID:        aliasIdent1.AliasID(),
				StateIndex:     2,
				StateMetadata:  nil,
				FoundryCounter: 1,
				Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
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
				Conditions: iotago.UnlockConditions[iotago.FoundryUnlockCondition]{
					&iotago.ImmutableAliasUnlockCondition{Address: aliasIdent1},
				},
				Features: nil,
			},
		},
	}

	sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys)
	require.NoError(t, err)

	tx := &iotago.Transaction{
		Essence: essence,
		Unlocks: iotago.Unlocks{
			&iotago.SignatureUnlock{Signature: sigs[0]},
			&iotago.ReferenceUnlock{Reference: 0},
			&iotago.AliasUnlock{Reference: 1},
		},
	}

	require.NoError(t, stardustVM.Execute(tx, &vm.Paras{}, inputs))
}

func TestStardustTransactionExecution(t *testing.T) {
	type test struct {
		name    string
		vmParas *vm.Paras
		inputs  iotago.OutputSet
		tx      *iotago.Transaction
		wantErr error
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
				defaultAmount        uint64 = OneMi
				confirmingUnixTime   uint32 = 750
				storageDepositReturn uint64 = OneMi / 2
				nativeTokenTransfer1        = tpkg.RandSortNativeTokens(10)
				nativeTokenTransfer2        = tpkg.RandSortNativeTokens(10)
			)

			var (
				nft1ID = tpkg.Rand32ByteArray()
				nft2ID = tpkg.Rand32ByteArray()
			)

			inputIDs := tpkg.RandOutputIDs(16)

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: defaultAmount,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount:       defaultAmount,
					NativeTokens: nativeTokenTransfer1,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident2},
					},
				},
				inputIDs[2]: &iotago.BasicOutput{
					Amount:       defaultAmount,
					NativeTokens: nativeTokenTransfer2,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident2},
					},
				},
				inputIDs[3]: &iotago.BasicOutput{
					Amount: defaultAmount,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident2},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: ident1,
							UnixTime:      500,
						},
					},
				},
				inputIDs[4]: &iotago.BasicOutput{
					Amount: defaultAmount,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident2},
						&iotago.TimelockUnlockCondition{
							UnixTime: 500,
						},
					},
				},
				inputIDs[5]: &iotago.BasicOutput{
					Amount: defaultAmount + storageDepositReturn,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident2},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: ident1,
							Amount:        storageDepositReturn,
						},
						&iotago.TimelockUnlockCondition{
							UnixTime: 500,
						},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: ident1,
							UnixTime:      900,
						},
					},
				},
				inputIDs[6]: &iotago.AliasOutput{
					Amount:         defaultAmount,
					NativeTokens:   nil,
					AliasID:        iotago.AliasID{},
					StateIndex:     0,
					StateMetadata:  []byte("gov transitioning"),
					FoundryCounter: 0,
					Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
						&iotago.StateControllerAddressUnlockCondition{Address: ident3},
						&iotago.GovernorAddressUnlockCondition{Address: ident4},
					},
					Features: nil,
				},
				inputIDs[7]: &iotago.AliasOutput{
					Amount:         defaultAmount + defaultAmount, // to fund also the new alias output
					NativeTokens:   nil,
					AliasID:        iotago.AliasID{},
					StateIndex:     5,
					StateMetadata:  []byte("current state"),
					FoundryCounter: 5,
					Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
						&iotago.StateControllerAddressUnlockCondition{Address: ident3},
						&iotago.GovernorAddressUnlockCondition{Address: ident4},
					},
					Features: nil,
				},
				inputIDs[8]: &iotago.AliasOutput{
					Amount:         defaultAmount,
					NativeTokens:   nil,
					AliasID:        iotago.AliasID{},
					StateIndex:     0,
					StateMetadata:  []byte("going to be destroyed"),
					FoundryCounter: 0,
					Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
						&iotago.StateControllerAddressUnlockCondition{Address: ident3},
						&iotago.GovernorAddressUnlockCondition{Address: ident3},
					},
					Features: nil,
				},
				inputIDs[9]: &iotago.FoundryOutput{
					Amount:       defaultAmount,
					NativeTokens: nil,
					SerialNumber: 1,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(100),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
					Conditions: iotago.UnlockConditions[iotago.FoundryUnlockCondition]{
						&iotago.ImmutableAliasUnlockCondition{Address: iotago.AliasIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AliasAddress)},
					},
					Features: nil,
				},
				inputIDs[10]: &iotago.FoundryOutput{
					Amount:       defaultAmount,
					NativeTokens: nil, // filled out later
					SerialNumber: 2,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(100),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
					Conditions: iotago.UnlockConditions[iotago.FoundryUnlockCondition]{
						&iotago.ImmutableAliasUnlockCondition{Address: iotago.AliasIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AliasAddress)},
					},
					Features: nil,
				},
				inputIDs[11]: &iotago.FoundryOutput{
					Amount:       defaultAmount,
					NativeTokens: nil,
					SerialNumber: 3,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(100),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
					Conditions: iotago.UnlockConditions[iotago.FoundryUnlockCondition]{
						&iotago.ImmutableAliasUnlockCondition{Address: iotago.AliasIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AliasAddress)},
					},
					Features: nil,
				},
				inputIDs[12]: &iotago.FoundryOutput{
					Amount:       defaultAmount,
					NativeTokens: nil,
					SerialNumber: 4,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(100),
						MeltedTokens:  big.NewInt(50),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
					Conditions: iotago.UnlockConditions[iotago.FoundryUnlockCondition]{
						&iotago.ImmutableAliasUnlockCondition{Address: iotago.AliasIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AliasAddress)},
					},
					Features: nil,
				},
				inputIDs[13]: &iotago.NFTOutput{
					Amount:       defaultAmount,
					NativeTokens: nil,
					NFTID:        nft1ID,
					Conditions: iotago.UnlockConditions[iotago.NFTUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident3},
					},
					Features: iotago.Features[iotago.NFTFeature]{
						&iotago.IssuerFeature{Address: ident3},
					},
					ImmutableFeatures: iotago.Features[iotago.NFTImmFeature]{
						&iotago.MetadataFeature{Data: []byte("transfer to 4")},
					},
				},
				inputIDs[14]: &iotago.NFTOutput{
					Amount:       defaultAmount,
					NativeTokens: nil,
					NFTID:        nft2ID,
					Conditions: iotago.UnlockConditions[iotago.NFTUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident4},
					},
					Features: iotago.Features[iotago.NFTFeature]{
						&iotago.IssuerFeature{Address: ident3},
					},
					ImmutableFeatures: iotago.Features[iotago.NFTImmFeature]{
						&iotago.MetadataFeature{Data: []byte("going to be destroyed")},
					},
				},
				inputIDs[15]: &iotago.BasicOutput{
					Amount: defaultAmount,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: iotago.NFTID(nft1ID).ToAddress()},
					},
				},
			}

			foundry1Ident3NativeTokenID := inputs[inputIDs[9]].(*iotago.FoundryOutput).MustNativeTokenID()
			foundry2Ident3NativeTokenID := inputs[inputIDs[10]].(*iotago.FoundryOutput).MustNativeTokenID()
			foundry4Ident3NativeTokenID := inputs[inputIDs[12]].(*iotago.FoundryOutput).MustNativeTokenID()

			newFoundryWithInitialSupply := &iotago.FoundryOutput{
				Amount:       defaultAmount,
				NativeTokens: nil,
				SerialNumber: 6,
				TokenScheme: &iotago.SimpleTokenScheme{
					MintedTokens:  big.NewInt(100),
					MeltedTokens:  big.NewInt(0),
					MaximumSupply: new(big.Int).SetInt64(1000),
				},
				Conditions: iotago.UnlockConditions[iotago.FoundryUnlockCondition]{
					&iotago.ImmutableAliasUnlockCondition{Address: iotago.AliasIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AliasAddress)},
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

			inputs[inputIDs[10]].(*iotago.FoundryOutput).NativeTokens = iotago.NativeTokens{
				{
					ID:     foundry2Ident3NativeTokenID,
					Amount: big.NewInt(100),
				},
			}

			inputs[inputIDs[12]].(*iotago.FoundryOutput).NativeTokens = iotago.NativeTokens{
				{
					ID:     foundry4Ident3NativeTokenID,
					Amount: big.NewInt(50),
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.BasicOutput{
						Amount: defaultAmount,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident5},
						},
					},
					&iotago.BasicOutput{
						Amount:       defaultAmount,
						NativeTokens: nativeTokenTransfer1,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident3},
						},
					},
					&iotago.BasicOutput{
						Amount:       defaultAmount,
						NativeTokens: nativeTokenTransfer2,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident4},
						},
					},
					&iotago.BasicOutput{
						Amount: defaultAmount,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
					},
					&iotago.BasicOutput{
						Amount: defaultAmount,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
					},
					&iotago.BasicOutput{
						Amount: storageDepositReturn,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					&iotago.AliasOutput{
						Amount:         defaultAmount,
						NativeTokens:   nil,
						AliasID:        iotago.AliasID{},
						StateIndex:     0,
						StateMetadata:  []byte("a new alias output"),
						FoundryCounter: 0,
						Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
							&iotago.StateControllerAddressUnlockCondition{Address: ident3},
							&iotago.GovernorAddressUnlockCondition{Address: ident4},
						},
						Features: nil,
					},
					&iotago.AliasOutput{
						Amount:         defaultAmount,
						NativeTokens:   nil,
						AliasID:        iotago.AliasIDFromOutputID(inputIDs[6]),
						StateIndex:     0,
						StateMetadata:  []byte("gov transitioning"),
						FoundryCounter: 0,
						Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
							&iotago.StateControllerAddressUnlockCondition{Address: ident3},
							&iotago.GovernorAddressUnlockCondition{Address: ident4},
						},
						Features: iotago.Features[iotago.AliasFeature]{
							&iotago.MetadataFeature{Data: []byte("the gov mutation on this output")},
						},
					},
					&iotago.AliasOutput{
						Amount:         defaultAmount,
						NativeTokens:   nil,
						AliasID:        iotago.AliasIDFromOutputID(inputIDs[7]),
						StateIndex:     6,
						StateMetadata:  []byte("next state"),
						FoundryCounter: 6,
						Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
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
						Conditions: iotago.UnlockConditions[iotago.FoundryUnlockCondition]{
							&iotago.ImmutableAliasUnlockCondition{Address: iotago.AliasIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AliasAddress)},
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
						Conditions: iotago.UnlockConditions[iotago.FoundryUnlockCondition]{
							&iotago.ImmutableAliasUnlockCondition{Address: iotago.AliasIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AliasAddress)},
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
						Conditions: iotago.UnlockConditions[iotago.FoundryUnlockCondition]{
							&iotago.ImmutableAliasUnlockCondition{Address: iotago.AliasIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AliasAddress)},
						},
						Features: iotago.Features[iotago.FoundryFeature]{
							&iotago.MetadataFeature{Data: []byte("interesting metadata")},
						},
					},
					// from foundry 4 ident 3 destruction remainder
					&iotago.BasicOutput{
						Amount: defaultAmount,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident3},
						},
					},
					&iotago.NFTOutput{
						Amount:       defaultAmount,
						NativeTokens: nil,
						NFTID:        iotago.NFTID{},
						Conditions: iotago.UnlockConditions[iotago.NFTUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident4},
						},
						Features: nil,
						ImmutableFeatures: iotago.Features[iotago.NFTImmFeature]{
							&iotago.MetadataFeature{Data: []byte("immutable metadata")},
						},
					},
					&iotago.NFTOutput{
						Amount:       defaultAmount,
						NativeTokens: nil,
						NFTID:        nft1ID,
						Conditions: iotago.UnlockConditions[iotago.NFTUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident4},
						},
						Features: iotago.Features[iotago.NFTFeature]{
							&iotago.IssuerFeature{Address: ident3},
						},
						ImmutableFeatures: iotago.Features[iotago.NFTImmFeature]{
							&iotago.MetadataFeature{Data: []byte("transfer to 4")},
						},
					},
					// from NFT ident 4 destruction remainder
					&iotago.BasicOutput{
						Amount: defaultAmount,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident3},
						},
					},
					// from NFT 1 to ident 5
					&iotago.BasicOutput{
						Amount: defaultAmount,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident5},
						},
					},
				},
			}

			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys, ident2AddrKeys, ident3AddrKeys, ident4AddrKeys)
			require.NoError(t, err)

			return test{
				name: "ok",
				vmParas: &vm.Paras{
					External: &iotago.ExternalUnlockParameters{ConfUnix: confirmingUnixTime},
				},
				inputs: inputs,
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
						// alias
						&iotago.SignatureUnlock{Signature: sigs[3]},
						&iotago.SignatureUnlock{Signature: sigs[2]},
						&iotago.ReferenceUnlock{Reference: 7},
						// foundries
						&iotago.AliasUnlock{Reference: 7},
						&iotago.AliasUnlock{Reference: 7},
						&iotago.AliasUnlock{Reference: 7},
						&iotago.AliasUnlock{Reference: 7},
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
			var (
				aliasAddr1 = tpkg.RandAliasAddress()
			)

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
				Conditions: iotago.UnlockConditions[iotago.FoundryUnlockCondition]{
					&iotago.ImmutableAliasUnlockCondition{Address: aliasAddr1},
				},
			}
			outFoundry := inFoundry.Clone().(*iotago.FoundryOutput)
			// change the immutable alias address unlock
			outFoundry.Conditions = iotago.UnlockConditions[iotago.FoundryUnlockCondition]{
				&iotago.ImmutableAliasUnlockCondition{Address: tpkg.RandAliasAddress()},
			}

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.AliasOutput{
					Amount:     100,
					StateIndex: 0,
					AliasID:    aliasAddr1.AliasID(),
					Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
						&iotago.StateControllerAddressUnlockCondition{Address: ident1},
						&iotago.GovernorAddressUnlockCondition{Address: ident2},
					},
				},
				inputIDs[1]: inFoundry,
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.AliasOutput{
						Amount:     100,
						StateIndex: 1,
						AliasID:    aliasAddr1.AliasID(),
						Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident2},
						},
					},
					outFoundry,
				},
			}

			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddressKeys)
			require.NoError(t, err)

			return test{
				name:    "fail - changed immutable alias address unlock",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						// should be an AliasUnlock
						&iotago.AliasUnlock{Reference: 0},
					},
				},
				// Changing the immutable alias address unlock changes foundryID, therefore the chain is broken.
				// Next state of the foundry is empty, meaning it is interpreted as a destroy operation, and native tokens
				// are not balanced.
				wantErr: iotago.ErrNativeTokenSumUnbalanced,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := stardustVM.Execute(tt.tx, tt.vmParas, tt.inputs)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestTxSemanticInputUnlocks(t *testing.T) {
	type test struct {
		name    string
		vmParas *vm.Paras
		inputs  iotago.OutputSet
		tx      *iotago.Transaction
		wantErr error
	}
	tests := []test{
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, ident2AddrKeys := tpkg.RandEd25519Identity()

			inputIDs := tpkg.RandOutputIDs(8)
			aliasIdent1 := iotago.AliasAddressFromOutputID(inputIDs[1])
			nftIdent1 := tpkg.RandNFTAddress()

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: &iotago.AliasOutput{
					Amount:  100,
					AliasID: iotago.AliasID{}, // empty on purpose as validation should resolve
					Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
						&iotago.StateControllerAddressUnlockCondition{Address: ident1},
						&iotago.GovernorAddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[2]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: &aliasIdent1},
					},
				},
				inputIDs[3]: &iotago.NFTOutput{
					Amount: 100,
					NFTID:  nftIdent1.NFTID(),
					Conditions: iotago.UnlockConditions[iotago.NFTUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: &aliasIdent1},
					},
				},
				inputIDs[4]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: nftIdent1},
					},
				},
				// unlockable by sender as expired
				inputIDs[5]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: ident2,
							UnixTime:      5,
						},
					},
				},
				// not unlockable by sender as not expired
				inputIDs[6]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: ident2,
							UnixTime:      20,
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
					Conditions: iotago.UnlockConditions[iotago.FoundryUnlockCondition]{
						&iotago.ImmutableAliasUnlockCondition{Address: &aliasIdent1},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.AliasOutput{
						Amount:     100,
						AliasID:    aliasIdent1.AliasID(),
						StateIndex: 1,
						Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident1},
						},
					},
				},
			}

			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys, ident2AddrKeys)
			require.NoError(t, err)

			return test{
				name: "ok",
				vmParas: &vm.Paras{
					External: &iotago.ExternalUnlockParameters{
						ConfUnix: 10,
					},
				},
				inputs: inputs,
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.ReferenceUnlock{Reference: 0},
						&iotago.AliasUnlock{Reference: 1},
						&iotago.AliasUnlock{Reference: 1},
						&iotago.NFTUnlock{Reference: 3},
						&iotago.SignatureUnlock{Signature: sigs[1]},
						&iotago.ReferenceUnlock{Reference: 0},
						&iotago.AliasUnlock{Reference: 1},
					},
				},
				wantErr: nil,
			}
		}(),
		func() test {
			ident1Sk, ident1, _ := tpkg.RandEd25519Identity()
			_, _, ident2AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident2AddrKeys)
			require.NoError(t, err)

			copy(sigs[0].(*iotago.Ed25519Signature).PublicKey[:], ident1Sk.Public().(ed25519.PublicKey))

			return test{
				name:    "fail - invalid signature",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddressKeys)
			require.NoError(t, err)

			return test{
				name:    "fail - should contain reference unlock",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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

			aliasIdent1 := iotago.AliasAddressFromOutputID(inputIDs[0])
			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.AliasOutput{
					Amount:  100,
					AliasID: iotago.AliasID{},
					Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
						&iotago.StateControllerAddressUnlockCondition{Address: ident1},
						&iotago.GovernorAddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: &aliasIdent1},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddressKeys)
			require.NoError(t, err)

			return test{
				name:    "fail - should contain alias unlock",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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
			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.NFTOutput{
					Amount: 100,
					NFTID:  iotago.NFTID{},
					Conditions: iotago.UnlockConditions[iotago.NFTUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: &nftIdent1},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddressKeys)
			require.NoError(t, err)

			return test{
				name:    "fail - should contain NFT unlock",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.NFTOutput{
					Amount: 100,
					NFTID:  nftIdent1.NFTID(),
					Conditions: iotago.UnlockConditions[iotago.NFTUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: &nftIdent2},
					},
				},
				inputIDs[1]: &iotago.NFTOutput{
					Amount: 100,
					NFTID:  nftIdent2.NFTID(),
					Conditions: iotago.UnlockConditions[iotago.NFTUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: &nftIdent2},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}
			_, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment())
			require.NoError(t, err)
			return test{
				name:    "fail - circular NFT unlock",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: ident2,
							UnixTime:      10,
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident2AddressKeys)
			require.NoError(t, err)

			return test{
				name: "fail - sender can not unlock yet",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{
					ConfUnix: 5,
				}},
				inputs: inputs,
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
			_, ident2, _ := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: ident2,
							UnixTime:      10,
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddressKeys)
			require.NoError(t, err)

			return test{
				name: "fail - receiver can not unlock anymore",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{
					ConfUnix: 10,
				}},
				inputs: inputs,
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
				aliasAddr1 = tpkg.RandAliasAddress()
				aliasAddr2 = tpkg.RandAliasAddress()
				aliasAddr3 = tpkg.RandAliasAddress()
			)

			inputs := iotago.OutputSet{
				// owned by ident1
				inputIDs[0]: &iotago.AliasOutput{
					Amount:  100,
					AliasID: aliasAddr1.AliasID(),
					Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
						&iotago.StateControllerAddressUnlockCondition{Address: ident1},
						&iotago.GovernorAddressUnlockCondition{Address: ident1},
					},
				},
				// owned by alias1
				inputIDs[1]: &iotago.AliasOutput{
					Amount:  100,
					AliasID: aliasAddr2.AliasID(),
					Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
						&iotago.StateControllerAddressUnlockCondition{Address: aliasAddr1},
						&iotago.GovernorAddressUnlockCondition{Address: aliasAddr1},
					},
				},
				// owned by alias1
				inputIDs[2]: &iotago.AliasOutput{
					Amount:  100,
					AliasID: aliasAddr3.AliasID(),
					Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
						&iotago.StateControllerAddressUnlockCondition{Address: aliasAddr1},
						&iotago.GovernorAddressUnlockCondition{Address: aliasAddr1},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddressKeys)
			require.NoError(t, err)

			return test{
				name:    "fail - referencing other alias unlocked by source alias",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.AliasUnlock{Reference: 0},
						// error, should be 0, because alias3 is unlocked by alias1, not alias2
						&iotago.AliasUnlock{Reference: 1},
					},
				},
				wantErr: iotago.ErrInvalidInputUnlock,
			}
		}(),
		func() test {
			_, ident1, _ := tpkg.RandEd25519Identity()
			_, ident2, ident2AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)

			var (
				aliasAddr1 = tpkg.RandAliasAddress()
			)

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.AliasOutput{
					Amount:  100,
					AliasID: aliasAddr1.AliasID(),
					Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
						&iotago.StateControllerAddressUnlockCondition{Address: ident1},
						&iotago.GovernorAddressUnlockCondition{Address: ident2},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: aliasAddr1},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.AliasOutput{
						Amount:  100,
						AliasID: aliasAddr1.AliasID(),
						Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident2},
						},
					},
				},
			}

			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident2AddressKeys)
			require.NoError(t, err)

			return test{
				name:    "fail - alias output not state transitioning",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						&iotago.AliasUnlock{Reference: 0},
					},
				},
				wantErr: iotago.ErrInvalidInputUnlock,
			}
		}(),
		func() test {
			var (
				aliasAddr1 = tpkg.RandAliasAddress()
			)

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
				Conditions: iotago.UnlockConditions[iotago.FoundryUnlockCondition]{
					&iotago.ImmutableAliasUnlockCondition{Address: aliasAddr1},
				},
			}

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.AliasOutput{
					Amount:     100,
					StateIndex: 0,
					AliasID:    aliasAddr1.AliasID(),
					Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
						&iotago.StateControllerAddressUnlockCondition{Address: ident1},
						&iotago.GovernorAddressUnlockCondition{Address: ident2},
					},
				},
				inputIDs[1]: foundryOutput,
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.AliasOutput{
						Amount:     100,
						StateIndex: 1,
						AliasID:    aliasAddr1.AliasID(),
						Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident2},
						},
					},
					foundryOutput,
				},
			}

			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddressKeys)
			require.NoError(t, err)

			return test{
				name:    "fail - wrong unlock for foundry",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
						// should be an AliasUnlock
						&iotago.ReferenceUnlock{Reference: 0},
					},
				},
				wantErr: iotago.ErrInvalidInputUnlock,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := stardustVM.Execute(tt.tx, tt.vmParas, tt.inputs, vm.ExecFuncInputUnlocks())
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestTxSemanticDeposit(t *testing.T) {
	type test struct {
		name    string
		vmParas *vm.Paras
		inputs  iotago.OutputSet
		tx      *iotago.Transaction
		wantErr error
	}
	tests := []test{
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, ident2, ident2AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(3)

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				// unlocked by ident1 as it is not expired
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 500,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: ident2,
							Amount:        420,
						},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: ident2,
							UnixTime:      10,
						},
					},
				},
				// unlocked by ident2 as it is expired
				inputIDs[2]: &iotago.BasicOutput{
					Amount: 500,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: ident2,
							Amount:        420,
						},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: ident2,
							UnixTime:      2,
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.BasicOutput{
						Amount: 180,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
					&iotago.BasicOutput{
						// return via ident1 + reclaim
						Amount: 420 + 500,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
					},
				},
			}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys, ident2AddrKeys)
			require.NoError(t, err)

			return test{
				name: "ok",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{
					ConfUnix: 5,
				}},
				inputs: inputs,
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

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 1000,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: ident2,
							Amount:        420,
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.BasicOutput{
						// returns 200 to ident2
						Amount: 200,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
					},
					&iotago.BasicOutput{
						// returns 221 to ident2
						Amount: 221,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
					},
					&iotago.BasicOutput{
						// remainder to random address
						Amount: 579,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name:    "ok - more storage deposit returned via more outputs",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 50,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "fail - unbalanced, more on output than input",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{
					ConfUnix: 5,
				}},
				inputs: inputs,
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

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.BasicOutput{
						Amount: 50,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "fail - unbalanced, more on input than output",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{
					ConfUnix: 5,
				}},
				inputs: inputs,
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

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 500,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: ident2,
							Amount:        420,
						},
						// not yet expired, so ident1 needs to unlock
						&iotago.ExpirationUnlockCondition{
							ReturnAddress: ident2,
							UnixTime:      10,
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.BasicOutput{
						Amount: 500,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
			}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "fail - return not fulfilled",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{
					ConfUnix: 5,
				}},
				inputs: inputs,
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

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 500,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: ident2,
							Amount:        420,
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.BasicOutput{
						Amount: 80,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					&iotago.NFTOutput{
						Amount: 420,
						Conditions: iotago.UnlockConditions[iotago.NFTUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
					},
				},
			}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name:    "fail - storage deposit return not basic output",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 500,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: ident2,
							Amount:        420,
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.BasicOutput{
						Amount: 80,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					&iotago.BasicOutput{
						Amount: 420,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident2},
							&iotago.ExpirationUnlockCondition{
								ReturnAddress: ident1,
								UnixTime:      10,
							},
						},
					},
				},
			}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name:    "fail - storage deposit return has additional unlocks",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 500,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: ident2,
							Amount:        420,
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.BasicOutput{
						Amount: 80,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					&iotago.BasicOutput{
						Amount: 420,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
						Features: iotago.Features[iotago.BasicOutputFeature]{
							&iotago.MetadataFeature{Data: []byte("foo")},
						},
					},
				},
			}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name:    "fail - storage deposit return has feature",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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
			ntId := tpkg.Rand38ByteArray()

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 500,
					NativeTokens: iotago.NativeTokens{
						&iotago.NativeToken{
							ID:     ntId,
							Amount: new(big.Int).SetUint64(1000),
						},
					},
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: ident2,
							Amount:        420,
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.BasicOutput{
						Amount: 80,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					&iotago.BasicOutput{
						Amount: 420,
						NativeTokens: iotago.NativeTokens{
							&iotago.NativeToken{
								ID:     ntId,
								Amount: new(big.Int).SetUint64(1000),
							},
						},
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
					},
				},
			}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name:    "fail - storage deposit return has native tokens",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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

			err := stardustVM.Execute(tt.tx, tt.vmParas, tt.inputs, vm.ExecFuncInputUnlocks(), vm.ExecFuncBalancedDeposit())
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestTxSemanticNativeTokens(t *testing.T) {

	foundryAliasIdent := tpkg.RandAliasAddress()
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
		Conditions: iotago.UnlockConditions[iotago.FoundryUnlockCondition]{
			&iotago.ImmutableAliasUnlockCondition{Address: foundryAliasIdent},
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
		Conditions: iotago.UnlockConditions[iotago.FoundryUnlockCondition]{
			&iotago.ImmutableAliasUnlockCondition{Address: foundryAliasIdent},
		},
	}

	type test struct {
		name    string
		vmParas *vm.Paras
		inputs  iotago.OutputSet
		tx      *iotago.Transaction
		wantErr error
	}
	tests := []test{
		func() test {
			inputIDs := tpkg.RandOutputIDs(2)

			ntCount := 10
			nativeTokens := tpkg.RandSortNativeTokens(ntCount)

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount:       100,
					NativeTokens: nativeTokens[:ntCount/2],
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount:       100,
					NativeTokens: nativeTokens[ntCount/2:],
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.BasicOutput{
						Amount:       200,
						NativeTokens: nativeTokens,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}

			return test{
				name:    "ok",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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

			inputs := iotago.OutputSet{}
			for i := 0; i < iotago.MaxNativeTokensCount; i++ {
				inputs[inputIDs[i]] = &iotago.BasicOutput{
					Amount: 100,
					NativeTokens: []*iotago.NativeToken{
						{
							ID:     nativeToken.ID,
							Amount: big.NewInt(1),
						},
					},
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				}
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.BasicOutput{
						Amount: 200,
						NativeTokens: []*iotago.NativeToken{
							{
								ID:     nativeToken.ID,
								Amount: big.NewInt(iotago.MaxNativeTokensCount),
							},
						},
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}

			return test{
				name:    "ok - exceeds limit (in+out) but same native token",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount:       100,
					NativeTokens: tpkg.RandSortNativeTokens(inCount),
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.BasicOutput{
						Amount:       200,
						NativeTokens: tpkg.RandSortNativeTokens(outCount),
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}

			return test{
				name:    "fail - exceeds limit (in+out)",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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

			inputs := iotago.OutputSet{}
			for i := 0; i < numDistinctNTs; i++ {
				inputs[inputIDs[i]] = &iotago.BasicOutput{
					Amount:       100,
					NativeTokens: tpkg.RandSortNativeTokens(1),
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				}
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.BasicOutput{
						Amount: 100 * uint64(numDistinctNTs),
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}

			return test{
				name:    "fail - too many on input side already",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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

			inputs := iotago.OutputSet{}
			for i := 0; i < numDistinctNTs; i++ {
				inputs[inputIDs[i]] = &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				}
			}

			outs := make(iotago.Outputs[iotago.TxEssenceOutput], numDistinctNTs)
			for i := range outs {
				outs[i] = &iotago.BasicOutput{
					Amount:       100,
					NativeTokens: iotago.NativeTokens{tokens[i]},
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				}
			}

			essence := &iotago.TransactionEssence{
				Inputs:  inputIDs.UTXOInputs(),
				Outputs: outs,
			}

			return test{
				name:    "fail - too many on output side already",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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

			inputs := iotago.OutputSet{}
			for i := 0; i < iotago.MaxInputsCount; i++ {
				inputs[inputIDs[i]] = &iotago.BasicOutput{
					Amount:       100,
					NativeTokens: tokens,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				}
			}

			outputs := make(iotago.Outputs[iotago.TxEssenceOutput], iotago.MaxOutputsCount)
			for i := range outputs {
				outputs[i] = &iotago.BasicOutput{
					Amount:       100,
					NativeTokens: tokens,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				}
			}

			essence := &iotago.TransactionEssence{
				Inputs:  inputIDs.UTXOInputs(),
				Outputs: outputs,
			}

			return test{
				name:    "ok - most possible tokens in a tx",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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

			inputs := iotago.OutputSet{}
			for i := 0; i < iotago.MaxInputsCount; i++ {
				inputs[inputIDs[i]] = &iotago.BasicOutput{
					Amount:       100,
					NativeTokens: tokens,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				}
			}

			outputs := make(iotago.Outputs[iotago.TxEssenceOutput], iotago.MaxOutputsCount)
			for i := range outputs {
				outputs[i] = &iotago.BasicOutput{
					Amount:       100,
					NativeTokens: tokens,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
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
				Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs:  inputIDs.UTXOInputs(),
				Outputs: outputs,
			}

			return test{
				name:    "fail - max nt count just exceeded",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount:       100,
					NativeTokens: nativeTokens,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			}

			// unbalance by making one token be excess on the output side
			cpyNativeTokens := nativeTokens.Clone()
			amountToModify := cpyNativeTokens[ntCount/2].Amount
			amountToModify.Add(amountToModify, big.NewInt(1))

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.BasicOutput{
						Amount:       100,
						NativeTokens: cpyNativeTokens,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}

			return test{
				name:    "fail - unbalanced on output",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount:       100,
					NativeTokens: nativeTokens[:ntCount/2],
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount:       100,
					NativeTokens: nativeTokens[ntCount/2:],
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				inputIDs[2]: inUnrelatedFoundryOutput,
			}

			// add a new token to the output side
			cpyNativeTokens := nativeTokens.Clone()
			cpyNativeTokens = append(cpyNativeTokens, tpkg.RandNativeToken())

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.BasicOutput{
						Amount:       100,
						NativeTokens: cpyNativeTokens[:ntCount/2],
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
					&iotago.BasicOutput{
						Amount:       100,
						NativeTokens: cpyNativeTokens[ntCount/2:],
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
					outUnrelatedFoundryOutput,
				},
			}

			return test{
				name:    "fail - unbalanced with unrelated foundry in term of new output tokens",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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
			err := stardustVM.Execute(tt.tx, tt.vmParas, tt.inputs, vm.ExecFuncBalancedNativeTokens())
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
		name    string
		vmParas *vm.Paras
		inputs  iotago.OutputSet
		tx      *iotago.Transaction
		wantErr error
	}
	tests := []test{
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)
			nftAddr := tpkg.RandNFTAddress()

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: &iotago.NFTOutput{
					Amount: 100,
					NFTID:  nftAddr.NFTID(),
					Conditions: iotago.UnlockConditions[iotago.NFTUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					// sender is an Ed25519 address
					&iotago.BasicOutput{
						Amount: 1337,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: iotago.Features[iotago.BasicOutputFeature]{
							&iotago.SenderFeature{Address: ident1},
						},
					},
					&iotago.AliasOutput{
						Amount: 1337,
						Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident1},
						},
						Features: iotago.Features[iotago.AliasFeature]{
							&iotago.SenderFeature{Address: ident1},
						},
					},
					&iotago.NFTOutput{
						Amount: 1337,
						Conditions: iotago.UnlockConditions[iotago.NFTUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: iotago.Features[iotago.NFTFeature]{
							&iotago.SenderFeature{Address: ident1},
						},
					},
					// sender is an NFT address
					&iotago.BasicOutput{
						Amount: 1337,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: iotago.Features[iotago.BasicOutputFeature]{
							&iotago.SenderFeature{Address: nftAddr},
						},
					},
					&iotago.AliasOutput{
						Amount: 1337,
						Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
							&iotago.StateControllerAddressUnlockCondition{Address: ident1},
							&iotago.GovernorAddressUnlockCondition{Address: ident1},
						},
						Features: iotago.Features[iotago.AliasFeature]{
							&iotago.SenderFeature{Address: nftAddr},
						},
					},
					&iotago.NFTOutput{
						Amount: 1337,
						Conditions: iotago.UnlockConditions[iotago.NFTUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: iotago.Features[iotago.NFTFeature]{
							&iotago.SenderFeature{Address: nftAddr},
						},
					},
				},
			}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name:    "ok",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.BasicOutput{
						Amount: 1337,
						Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Features: iotago.Features[iotago.BasicOutputFeature]{
							&iotago.SenderFeature{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name:    "fail - sender not unlocked",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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
			aliasAddr := tpkg.RandAliasAddress()
			aliasId := aliasAddr.AliasID()

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.AliasOutput{
					Amount:  100,
					AliasID: aliasId,
					Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
						&iotago.StateControllerAddressUnlockCondition{Address: stateController},
						&iotago.GovernorAddressUnlockCondition{Address: governor},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.AliasOutput{
						Amount:  100,
						AliasID: aliasId,
						Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
						Features: iotago.Features[iotago.AliasFeature]{
							&iotago.SenderFeature{Address: aliasAddr},
						},
					},
				},
			}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), governorAddrKeys)
			require.NoError(t, err)

			return test{
				name:    "fail - sender not unlocked due to governance transition",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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
			aliasAddr := tpkg.RandAliasAddress()
			aliasId := aliasAddr.AliasID()
			currentStateIndex := uint32(1)

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.AliasOutput{
					Amount:     100,
					AliasID:    aliasId,
					StateIndex: currentStateIndex,
					Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
						&iotago.StateControllerAddressUnlockCondition{Address: stateController},
						&iotago.GovernorAddressUnlockCondition{Address: governor},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.AliasOutput{
						Amount:     100,
						AliasID:    aliasId,
						StateIndex: currentStateIndex + 1,
						Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
						Features: iotago.Features[iotago.AliasFeature]{
							&iotago.SenderFeature{Address: aliasAddr},
						},
					},
				},
			}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), stateControllerAddrKeys)
			require.NoError(t, err)

			return test{
				name:    "ok - alias addr unlocked with state transition",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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
			aliasAddr := tpkg.RandAliasAddress()
			aliasId := aliasAddr.AliasID()

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.AliasOutput{
					Amount:  100,
					AliasID: aliasId,
					Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
						&iotago.StateControllerAddressUnlockCondition{Address: stateController},
						&iotago.GovernorAddressUnlockCondition{Address: governor},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.AliasOutput{
						Amount:  100,
						AliasID: aliasId,
						Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
						Features: iotago.Features[iotago.AliasFeature]{
							&iotago.SenderFeature{Address: governor},
						},
					},
				},
			}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), governorAddrKeys)
			require.NoError(t, err)

			return test{
				name:    "ok - sender is governor address",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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

			err := stardustVM.Execute(tt.tx, tt.vmParas, tt.inputs, vm.ExecFuncInputUnlocks(), vm.ExecFuncSenderUnlocked())
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
		name    string
		vmParas *vm.Paras
		inputs  iotago.OutputSet
		tx      *iotago.Transaction
		wantErr error
	}
	tests := []test{
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			_, stateController, _ := tpkg.RandEd25519Identity()
			_, governor, governorAddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(2)
			aliasAddr := tpkg.RandAliasAddress()
			aliasId := aliasAddr.AliasID()

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.AliasOutput{
					Amount:  100,
					AliasID: aliasId,
					Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
						&iotago.StateControllerAddressUnlockCondition{Address: stateController},
						&iotago.GovernorAddressUnlockCondition{Address: governor},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.AliasOutput{
						Amount:  100,
						AliasID: aliasId,
						Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
					},
					&iotago.NFTOutput{
						Amount: 100,
						Conditions: iotago.UnlockConditions[iotago.NFTUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
						ImmutableFeatures: iotago.Features[iotago.NFTImmFeature]{
							&iotago.IssuerFeature{Address: aliasAddr},
						},
					},
				},
			}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), governorAddrKeys, ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name:    "fail - issuer not unlocked due to governance transition",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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
			aliasAddr := tpkg.RandAliasAddress()
			aliasId := aliasAddr.AliasID()
			currentStateIndex := uint32(1)

			nftAddr := tpkg.RandNFTAddress()

			inputs := iotago.OutputSet{
				// possible issuers: aliasAddr, stateController, nftAddr, ident1
				inputIDs[0]: &iotago.AliasOutput{
					Amount:     100,
					AliasID:    aliasId,
					StateIndex: currentStateIndex,
					Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
						&iotago.StateControllerAddressUnlockCondition{Address: stateController},
						&iotago.GovernorAddressUnlockCondition{Address: governor},
					},
				},
				inputIDs[1]: &iotago.NFTOutput{
					Amount: 900,
					NFTID:  nftAddr.NFTID(),
					Conditions: iotago.UnlockConditions[iotago.NFTUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					// transitioned alias + nft
					&iotago.AliasOutput{
						Amount:     100,
						AliasID:    aliasId,
						StateIndex: currentStateIndex + 1,
						Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
					},
					&iotago.NFTOutput{
						Amount: 100,
						NFTID:  nftAddr.NFTID(),
						Conditions: iotago.UnlockConditions[iotago.NFTUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
					// issuer is aliasAddr
					&iotago.NFTOutput{
						Amount: 100,
						Conditions: iotago.UnlockConditions[iotago.NFTUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
						ImmutableFeatures: iotago.Features[iotago.NFTImmFeature]{
							&iotago.IssuerFeature{Address: aliasAddr},
						},
					},
					&iotago.AliasOutput{
						Amount: 100,
						Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
						ImmutableFeatures: iotago.Features[iotago.AliasImmFeature]{
							&iotago.IssuerFeature{Address: aliasAddr},
						},
					},
					// issuer is stateController
					&iotago.NFTOutput{
						Amount: 100,
						Conditions: iotago.UnlockConditions[iotago.NFTUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
						ImmutableFeatures: iotago.Features[iotago.NFTImmFeature]{
							&iotago.IssuerFeature{Address: stateController},
						},
					},
					&iotago.AliasOutput{
						Amount: 100,
						Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
						ImmutableFeatures: iotago.Features[iotago.AliasImmFeature]{
							&iotago.IssuerFeature{Address: stateController},
						},
					},
					// issuer is nftAddr
					&iotago.NFTOutput{
						Amount: 100,
						Conditions: iotago.UnlockConditions[iotago.NFTUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
						ImmutableFeatures: iotago.Features[iotago.NFTImmFeature]{
							&iotago.IssuerFeature{Address: nftAddr},
						},
					},
					&iotago.AliasOutput{
						Amount: 100,
						Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
						ImmutableFeatures: iotago.Features[iotago.AliasImmFeature]{
							&iotago.IssuerFeature{Address: nftAddr},
						},
					},
					// issuer is ident1
					&iotago.NFTOutput{
						Amount: 100,
						Conditions: iotago.UnlockConditions[iotago.NFTUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
						ImmutableFeatures: iotago.Features[iotago.NFTImmFeature]{
							&iotago.IssuerFeature{Address: ident1},
						},
					},
					&iotago.AliasOutput{
						Amount: 100,
						Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
						ImmutableFeatures: iotago.Features[iotago.AliasImmFeature]{
							&iotago.IssuerFeature{Address: ident1},
						},
					},
				},
			}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), stateControllerAddrKeys, ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name:    "ok - issuer unlocked with state transition",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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
			aliasAddr := tpkg.RandAliasAddress()
			aliasId := aliasAddr.AliasID()

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.AliasOutput{
					Amount:  100,
					AliasID: aliasId,
					Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
						&iotago.StateControllerAddressUnlockCondition{Address: stateController},
						&iotago.GovernorAddressUnlockCondition{Address: governor},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
					&iotago.AliasOutput{
						Amount:  100,
						AliasID: aliasId,
						Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
							&iotago.StateControllerAddressUnlockCondition{Address: stateController},
							&iotago.GovernorAddressUnlockCondition{Address: governor},
						},
					},
					&iotago.NFTOutput{
						Amount: 100,
						Conditions: iotago.UnlockConditions[iotago.NFTUnlockCondition]{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
						ImmutableFeatures: iotago.Features[iotago.NFTImmFeature]{
							&iotago.IssuerFeature{Address: governor},
						},
					},
				},
			}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), governorAddrKeys, ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name:    "ok - issuer is the governor",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{}},
				inputs:  inputs,
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

			err := stardustVM.Execute(tt.tx, tt.vmParas, tt.inputs)
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
		name    string
		vmParas *vm.Paras
		inputs  iotago.OutputSet
		tx      *iotago.Transaction
		wantErr error
	}
	tests := []test{
		func() test {
			_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.TimelockUnlockCondition{
							UnixTime: 5,
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "ok",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{
					ConfUnix: 10,
				}},
				inputs: inputs,
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

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.TimelockUnlockCondition{
							UnixTime: 15,
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "fail - timelock not expired",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{
					ConfUnix: 10,
				}},
				inputs: inputs,
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

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.TimelockUnlockCondition{
							UnixTime: 1337,
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "fail - unix timelock not expired",
				vmParas: &vm.Paras{External: &iotago.ExternalUnlockParameters{
					ConfUnix: 666,
				}},
				inputs: inputs,
				tx: &iotago.Transaction{
					Essence: essence,
					Unlocks: iotago.Unlocks{
						&iotago.SignatureUnlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrTimelockNotExpired,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := stardustVM.Execute(tt.tx, tt.vmParas, tt.inputs, vm.ExecFuncTimelocks())
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}
