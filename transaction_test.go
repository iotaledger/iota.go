package iotago_test

import (
	"crypto/ed25519"
	"math/big"
	"testing"

	"github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/stretchr/testify/require"
)

func TestTransactionDeSerialize(t *testing.T) {
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

func TestNFTTransition(t *testing.T) {
	_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()

	inputIDs := tpkg.RandOutputIDs(1)
	inputs := iotago.OutputSet{
		inputIDs[0]: &iotago.NFTOutput{
			Amount: OneMi,
			NFTID:  iotago.NFTID{},
			Conditions: iotago.UnlockConditions{
				&iotago.AddressUnlockCondition{Address: ident1},
			},
			Blocks: nil,
		},
	}

	nftAddr := iotago.NFTAddressFromOutputID(inputIDs[0])
	nftID := nftAddr.NFTID()

	essence := &iotago.TransactionEssence{
		Inputs: inputIDs.UTXOInputs(),
		Outputs: iotago.Outputs{
			&iotago.NFTOutput{
				Amount: OneMi,
				NFTID:  nftID,
				Conditions: iotago.UnlockConditions{
					&iotago.AddressUnlockCondition{Address: ident1},
				},
				Blocks: nil,
			},
		},
	}

	sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys)
	require.NoError(t, err)

	tx := &iotago.Transaction{
		Essence: essence,
		UnlockBlocks: iotago.UnlockBlocks{
			&iotago.SignatureUnlockBlock{Signature: sigs[0]},
		},
	}

	require.NoError(t, tx.SemanticallyValidate(&iotago.SemanticValidationContext{
		ExtParas:   nil,
		WorkingSet: nil,
	}, inputs))
}

func TestCirculatingSupplyMelting(t *testing.T) {
	_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
	aliasIdent1 := tpkg.RandAliasAddress()

	tokenTag := tpkg.Rand12ByteArray()

	inputIDs := tpkg.RandOutputIDs(3)
	inputs := iotago.OutputSet{
		inputIDs[0]: &iotago.BasicOutput{
			Amount: OneMi,
			Conditions: iotago.UnlockConditions{
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
			Conditions: iotago.UnlockConditions{
				&iotago.StateControllerAddressUnlockCondition{Address: ident1},
				&iotago.GovernorAddressUnlockCondition{Address: ident1},
			},
			Blocks: nil,
		},
		inputIDs[2]: &iotago.FoundryOutput{
			Amount:       OneMi,
			NativeTokens: nil,
			SerialNumber: 1,
			TokenTag:     tokenTag,
			TokenScheme: &iotago.SimpleTokenScheme{
				MintedTokens:  big.NewInt(50),
				MeltedTokens:  big.NewInt(0),
				MaximumSupply: big.NewInt(50),
			},
			Conditions: iotago.UnlockConditions{
				&iotago.ImmutableAliasUnlockCondition{Address: aliasIdent1},
			},
			Blocks: nil,
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
		Outputs: iotago.Outputs{
			&iotago.AliasOutput{
				Amount:         OneMi,
				NativeTokens:   nil,
				AliasID:        aliasIdent1.AliasID(),
				StateIndex:     1,
				StateMetadata:  nil,
				FoundryCounter: 1,
				Conditions: iotago.UnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: ident1},
					&iotago.GovernorAddressUnlockCondition{Address: ident1},
				},
				Blocks: nil,
			},
			&iotago.FoundryOutput{
				Amount:       2 * OneMi,
				NativeTokens: nil,
				SerialNumber: 1,
				TokenTag:     tokenTag,
				TokenScheme: &iotago.SimpleTokenScheme{
					MintedTokens:  big.NewInt(50),
					MeltedTokens:  big.NewInt(50),
					MaximumSupply: big.NewInt(50),
				},
				Conditions: iotago.UnlockConditions{
					&iotago.ImmutableAliasUnlockCondition{Address: aliasIdent1},
				},
				Blocks: nil,
			},
		},
	}

	sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys)
	require.NoError(t, err)

	tx := &iotago.Transaction{
		Essence: essence,
		UnlockBlocks: iotago.UnlockBlocks{
			&iotago.SignatureUnlockBlock{Signature: sigs[0]},
			&iotago.ReferenceUnlockBlock{Reference: 0},
			&iotago.AliasUnlockBlock{Reference: 1},
		},
	}

	require.NoError(t, tx.SemanticallyValidate(&iotago.SemanticValidationContext{
		ExtParas:   nil,
		WorkingSet: nil,
	}, inputs))
}

func TestTransactionSemanticValidation(t *testing.T) {
	type test struct {
		name    string
		svCtx   *iotago.SemanticValidationContext
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
				defaultAmount            uint64 = OneMi
				confirmingMilestoneIndex uint32 = 750
				storageDepositReturn     uint64 = OneMi / 2
				nativeTokenTransfer1            = tpkg.RandSortNativeTokens(10)
				nativeTokenTransfer2            = tpkg.RandSortNativeTokens(10)
			)

			var (
				foundry1Ident3TokenTag = tpkg.Rand12ByteArray()
				foundry2Ident3TokenTag = tpkg.Rand12ByteArray()
				foundry3Ident3TokenTag = tpkg.Rand12ByteArray()
				foundry4Ident3TokenTag = tpkg.Rand12ByteArray()
			)

			var (
				nft1ID = tpkg.Rand20ByteArray()
				nft2ID = tpkg.Rand20ByteArray()
			)

			inputIDs := tpkg.RandOutputIDs(16)

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: defaultAmount,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount:       defaultAmount,
					NativeTokens: nativeTokenTransfer1,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident2},
					},
				},
				inputIDs[2]: &iotago.BasicOutput{
					Amount:       defaultAmount,
					NativeTokens: nativeTokenTransfer2,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident2},
					},
				},
				inputIDs[3]: &iotago.BasicOutput{
					Amount: defaultAmount,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident2},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress:  ident1,
							MilestoneIndex: 500,
						},
					},
				},
				inputIDs[4]: &iotago.BasicOutput{
					Amount: defaultAmount,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident2},
						&iotago.TimelockUnlockCondition{
							MilestoneIndex: 500,
						},
					},
				},
				inputIDs[5]: &iotago.BasicOutput{
					Amount: defaultAmount + storageDepositReturn,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident2},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: ident1,
							Amount:        storageDepositReturn,
						},
						&iotago.TimelockUnlockCondition{
							MilestoneIndex: 500,
						},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress:  ident1,
							MilestoneIndex: 900,
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
					Conditions: iotago.UnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident3},
						&iotago.GovernorAddressUnlockCondition{Address: ident4},
					},
					Blocks: nil,
				},
				inputIDs[7]: &iotago.AliasOutput{
					Amount:         defaultAmount + defaultAmount, // to fund also the new alias output
					NativeTokens:   nil,
					AliasID:        iotago.AliasID{},
					StateIndex:     5,
					StateMetadata:  []byte("current state"),
					FoundryCounter: 5,
					Conditions: iotago.UnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident3},
						&iotago.GovernorAddressUnlockCondition{Address: ident4},
					},
					Blocks: nil,
				},
				inputIDs[8]: &iotago.AliasOutput{
					Amount:         defaultAmount,
					NativeTokens:   nil,
					AliasID:        iotago.AliasID{},
					StateIndex:     0,
					StateMetadata:  []byte("going to be destroyed"),
					FoundryCounter: 0,
					Conditions: iotago.UnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident3},
						&iotago.GovernorAddressUnlockCondition{Address: ident3},
					},
					Blocks: nil,
				},
				inputIDs[9]: &iotago.FoundryOutput{
					Amount:       defaultAmount,
					NativeTokens: nil,
					SerialNumber: 1,
					TokenTag:     foundry1Ident3TokenTag,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(100),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
					Conditions: iotago.UnlockConditions{
						&iotago.ImmutableAliasUnlockCondition{Address: iotago.AliasIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AliasAddress)},
					},
					Blocks: nil,
				},
				inputIDs[10]: &iotago.FoundryOutput{
					Amount:       defaultAmount,
					NativeTokens: nil, // filled out later
					SerialNumber: 2,
					TokenTag:     foundry2Ident3TokenTag,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(100),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
					Conditions: iotago.UnlockConditions{
						&iotago.ImmutableAliasUnlockCondition{Address: iotago.AliasIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AliasAddress)},
					},
					Blocks: nil,
				},
				inputIDs[11]: &iotago.FoundryOutput{
					Amount:       defaultAmount,
					NativeTokens: nil,
					SerialNumber: 3,
					TokenTag:     foundry3Ident3TokenTag,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(100),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
					Conditions: iotago.UnlockConditions{
						&iotago.ImmutableAliasUnlockCondition{Address: iotago.AliasIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AliasAddress)},
					},
					Blocks: nil,
				},
				inputIDs[12]: &iotago.FoundryOutput{
					Amount:       defaultAmount,
					NativeTokens: nil,
					SerialNumber: 4,
					TokenTag:     foundry4Ident3TokenTag,
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetUint64(100),
						MeltedTokens:  big.NewInt(50),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
					Conditions: iotago.UnlockConditions{
						&iotago.ImmutableAliasUnlockCondition{Address: iotago.AliasIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AliasAddress)},
					},
					Blocks: nil,
				},
				inputIDs[13]: &iotago.NFTOutput{
					Amount:       defaultAmount,
					NativeTokens: nil,
					NFTID:        nft1ID,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident3},
					},
					Blocks: iotago.FeatureBlocks{
						&iotago.IssuerFeatureBlock{Address: ident3},
					},
					ImmutableBlocks: iotago.FeatureBlocks{
						&iotago.MetadataFeatureBlock{Data: []byte("transfer to 4")},
					},
				},
				inputIDs[14]: &iotago.NFTOutput{
					Amount:       defaultAmount,
					NativeTokens: nil,
					NFTID:        nft2ID,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident4},
					},
					Blocks: iotago.FeatureBlocks{
						&iotago.IssuerFeatureBlock{Address: ident3},
					},
					ImmutableBlocks: iotago.FeatureBlocks{
						&iotago.MetadataFeatureBlock{Data: []byte("going to be destroyed")},
					},
				},
				inputIDs[15]: &iotago.BasicOutput{
					Amount: defaultAmount,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: iotago.NFTID(nft1ID).ToAddress()},
					},
				},
			}

			foundry1Ident3NativeTokenID := inputs[inputIDs[9]].(*iotago.FoundryOutput).MustNativeTokenID()
			foundry2Ident3NativeTokenID := inputs[inputIDs[10]].(*iotago.FoundryOutput).MustNativeTokenID()
			foundry4Ident3NativeTokenID := inputs[inputIDs[12]].(*iotago.FoundryOutput).MustNativeTokenID()

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
				Outputs: iotago.Outputs{
					&iotago.BasicOutput{
						Amount: defaultAmount,
						Conditions: iotago.UnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident5},
						},
					},
					&iotago.BasicOutput{
						Amount:       defaultAmount,
						NativeTokens: nativeTokenTransfer1,
						Conditions: iotago.UnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident3},
						},
					},
					&iotago.BasicOutput{
						Amount:       defaultAmount,
						NativeTokens: nativeTokenTransfer2,
						Conditions: iotago.UnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident4},
						},
					},
					&iotago.BasicOutput{
						Amount: defaultAmount,
						Conditions: iotago.UnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
					},
					&iotago.BasicOutput{
						Amount: defaultAmount,
						Conditions: iotago.UnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
					},
					&iotago.BasicOutput{
						Amount: storageDepositReturn,
						Conditions: iotago.UnlockConditions{
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
						Conditions: iotago.UnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident3},
							&iotago.GovernorAddressUnlockCondition{Address: ident4},
						},
						Blocks: nil,
					},
					&iotago.AliasOutput{
						Amount:         defaultAmount,
						NativeTokens:   nil,
						AliasID:        iotago.AliasIDFromOutputID(inputIDs[6]),
						StateIndex:     0,
						StateMetadata:  []byte("gov transitioning"),
						FoundryCounter: 0,
						Conditions: iotago.UnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident3},
							&iotago.GovernorAddressUnlockCondition{Address: ident4},
						},
						Blocks: iotago.FeatureBlocks{
							&iotago.MetadataFeatureBlock{Data: []byte("the gov mutation on this output")},
						},
					},
					&iotago.AliasOutput{
						Amount:         defaultAmount,
						NativeTokens:   nil,
						AliasID:        iotago.AliasIDFromOutputID(inputIDs[7]),
						StateIndex:     6,
						StateMetadata:  []byte("next state"),
						FoundryCounter: 6,
						Conditions: iotago.UnlockConditions{
							&iotago.StateControllerAddressUnlockCondition{Address: ident3},
							&iotago.GovernorAddressUnlockCondition{Address: ident4},
						},
						Blocks: nil,
					},
					// new foundry
					&iotago.FoundryOutput{
						Amount:       defaultAmount,
						NativeTokens: nil,
						SerialNumber: 6,
						TokenTag:     tpkg.Rand12ByteArray(),
						TokenScheme: &iotago.SimpleTokenScheme{
							MintedTokens:  big.NewInt(0),
							MeltedTokens:  big.NewInt(0),
							MaximumSupply: new(big.Int).SetInt64(1000),
						},
						Conditions: iotago.UnlockConditions{
							&iotago.ImmutableAliasUnlockCondition{Address: iotago.AliasIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AliasAddress)},
						},
						Blocks: nil,
					},
					&iotago.FoundryOutput{
						Amount: defaultAmount,
						NativeTokens: iotago.NativeTokens{
							{
								ID:     foundry1Ident3NativeTokenID,
								Amount: new(big.Int).SetUint64(100), // freshly minted
							},
						},
						SerialNumber: 1,
						TokenTag:     foundry1Ident3TokenTag,
						TokenScheme: &iotago.SimpleTokenScheme{
							MintedTokens:  new(big.Int).SetInt64(200),
							MeltedTokens:  big.NewInt(0),
							MaximumSupply: new(big.Int).SetInt64(1000),
						},
						Conditions: iotago.UnlockConditions{
							&iotago.ImmutableAliasUnlockCondition{Address: iotago.AliasIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AliasAddress)},
						},
						Blocks: nil,
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
						TokenTag:     foundry2Ident3TokenTag,
						TokenScheme: &iotago.SimpleTokenScheme{
							MintedTokens:  new(big.Int).SetInt64(100),
							MeltedTokens:  big.NewInt(50),
							MaximumSupply: new(big.Int).SetInt64(1000),
						},
						Conditions: iotago.UnlockConditions{
							&iotago.ImmutableAliasUnlockCondition{Address: iotago.AliasIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AliasAddress)},
						},
						Blocks: nil,
					},
					&iotago.FoundryOutput{
						Amount:       defaultAmount,
						NativeTokens: nil,
						SerialNumber: 3,
						TokenTag:     foundry3Ident3TokenTag,
						TokenScheme: &iotago.SimpleTokenScheme{
							MintedTokens:  new(big.Int).SetInt64(100),
							MeltedTokens:  big.NewInt(0),
							MaximumSupply: new(big.Int).SetInt64(1000),
						},
						Conditions: iotago.UnlockConditions{
							&iotago.ImmutableAliasUnlockCondition{Address: iotago.AliasIDFromOutputID(inputIDs[7]).ToAddress().(*iotago.AliasAddress)},
						},
						Blocks: iotago.FeatureBlocks{
							&iotago.MetadataFeatureBlock{Data: []byte("interesting metadata")},
						},
					},
					// from foundry 4 ident 3 destruction remainder
					&iotago.BasicOutput{
						Amount: defaultAmount,
						Conditions: iotago.UnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident3},
						},
					},
					&iotago.NFTOutput{
						Amount:       defaultAmount,
						NativeTokens: nil,
						NFTID:        iotago.NFTID{},
						Conditions: iotago.UnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident4},
						},
						Blocks: nil,
						ImmutableBlocks: iotago.FeatureBlocks{
							&iotago.MetadataFeatureBlock{Data: []byte("immutable metadata")},
						},
					},
					&iotago.NFTOutput{
						Amount:       defaultAmount,
						NativeTokens: nil,
						NFTID:        nft1ID,
						Conditions: iotago.UnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident4},
						},
						Blocks: iotago.FeatureBlocks{
							&iotago.IssuerFeatureBlock{Address: ident3},
						},
						ImmutableBlocks: iotago.FeatureBlocks{
							&iotago.MetadataFeatureBlock{Data: []byte("transfer to 4")},
						},
					},
					// from NFT ident 4 destruction remainder
					&iotago.BasicOutput{
						Amount: defaultAmount,
						Conditions: iotago.UnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident3},
						},
					},
					// from NFT 1 to ident 5
					&iotago.BasicOutput{
						Amount: defaultAmount,
						Conditions: iotago.UnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident5},
						},
					},
				},
			}

			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys, ident2AddrKeys, ident3AddrKeys, ident4AddrKeys)
			require.NoError(t, err)

			return test{
				name: "ok",
				svCtx: &iotago.SemanticValidationContext{
					ExtParas: &iotago.ExternalUnlockParameters{ConfMsIndex: confirmingMilestoneIndex},
				},
				inputs: inputs,
				tx: &iotago.Transaction{
					Essence: essence,
					UnlockBlocks: iotago.UnlockBlocks{
						// basic
						&iotago.SignatureUnlockBlock{Signature: sigs[0]},
						&iotago.SignatureUnlockBlock{Signature: sigs[1]},
						&iotago.ReferenceUnlockBlock{Reference: 1},
						&iotago.ReferenceUnlockBlock{Reference: 0},
						&iotago.ReferenceUnlockBlock{Reference: 1},
						&iotago.ReferenceUnlockBlock{Reference: 1},
						// alias
						&iotago.SignatureUnlockBlock{Signature: sigs[3]},
						&iotago.SignatureUnlockBlock{Signature: sigs[2]},
						&iotago.ReferenceUnlockBlock{Reference: 7},
						// foundries
						&iotago.AliasUnlockBlock{Reference: 7},
						&iotago.AliasUnlockBlock{Reference: 7},
						&iotago.AliasUnlockBlock{Reference: 7},
						&iotago.AliasUnlockBlock{Reference: 7},
						// nfts
						&iotago.ReferenceUnlockBlock{Reference: 7},
						&iotago.ReferenceUnlockBlock{Reference: 6},
						&iotago.NFTUnlockBlock{Reference: 13},
					},
				},
				wantErr: nil,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tx.SemanticallyValidate(tt.svCtx, tt.inputs)
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
		svCtx   *iotago.SemanticValidationContext
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
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: &iotago.AliasOutput{
					Amount:  100,
					AliasID: iotago.AliasID{}, // empty on purpose as validation should resolve
					Conditions: iotago.UnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident1},
						&iotago.GovernorAddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[2]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: &aliasIdent1},
					},
				},
				inputIDs[3]: &iotago.NFTOutput{
					Amount: 100,
					NFTID:  nftIdent1.NFTID(),
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: &aliasIdent1},
					},
				},
				inputIDs[4]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: nftIdent1},
					},
				},
				// unlockable by sender as expired
				inputIDs[5]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress:  ident2,
							MilestoneIndex: 5,
						},
					},
				},
				// not unlockable by sender as not expired
				inputIDs[6]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress:  ident2,
							MilestoneIndex: 20,
						},
					},
				},
				inputIDs[7]: &iotago.FoundryOutput{
					Amount:       100,
					SerialNumber: 0,
					TokenTag:     tpkg.Rand12ByteArray(),
					TokenScheme: &iotago.SimpleTokenScheme{
						MintedTokens:  new(big.Int).SetInt64(100),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetInt64(1000),
					},
					Conditions: iotago.UnlockConditions{
						&iotago.ImmutableAliasUnlockCondition{Address: &aliasIdent1},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys, ident2AddrKeys)
			require.NoError(t, err)

			return test{
				name: "ok",
				svCtx: &iotago.SemanticValidationContext{
					ExtParas: &iotago.ExternalUnlockParameters{
						ConfMsIndex: 10,
						ConfUnix:    1337,
					},
				},
				inputs: inputs,
				tx: &iotago.Transaction{
					Essence: essence,
					UnlockBlocks: iotago.UnlockBlocks{
						&iotago.SignatureUnlockBlock{Signature: sigs[0]},
						&iotago.ReferenceUnlockBlock{Reference: 0},
						&iotago.AliasUnlockBlock{Reference: 1},
						&iotago.AliasUnlockBlock{Reference: 1},
						&iotago.NFTUnlockBlock{Reference: 3},
						&iotago.SignatureUnlockBlock{Signature: sigs[1]},
						&iotago.ReferenceUnlockBlock{Reference: 0},
						&iotago.AliasUnlockBlock{Reference: 1},
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
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident2AddrKeys)
			require.NoError(t, err)

			copy(sigs[0].(*iotago.Ed25519Signature).PublicKey[:], ident1Sk.Public().(ed25519.PublicKey))

			return test{
				name:   "fail - invalid signature",
				svCtx:  &iotago.SemanticValidationContext{ExtParas: &iotago.ExternalUnlockParameters{}},
				inputs: inputs,
				tx: &iotago.Transaction{
					Essence: essence,
					UnlockBlocks: iotago.UnlockBlocks{
						&iotago.SignatureUnlockBlock{Signature: sigs[0]},
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
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddressKeys)
			require.NoError(t, err)

			return test{
				name:   "fail - should contain reference unlock block",
				svCtx:  &iotago.SemanticValidationContext{ExtParas: &iotago.ExternalUnlockParameters{}},
				inputs: inputs,
				tx: &iotago.Transaction{
					Essence: essence,
					UnlockBlocks: iotago.UnlockBlocks{
						&iotago.SignatureUnlockBlock{Signature: sigs[0]},
						&iotago.SignatureUnlockBlock{Signature: sigs[0]},
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
					Conditions: iotago.UnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: ident1},
						&iotago.GovernorAddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: &aliasIdent1},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddressKeys)
			require.NoError(t, err)

			return test{
				name:   "fail - should contain alias unlock block",
				svCtx:  &iotago.SemanticValidationContext{ExtParas: &iotago.ExternalUnlockParameters{}},
				inputs: inputs,
				tx: &iotago.Transaction{
					Essence: essence,
					UnlockBlocks: iotago.UnlockBlocks{
						&iotago.SignatureUnlockBlock{Signature: sigs[0]},
						&iotago.ReferenceUnlockBlock{Reference: 0},
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
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: &nftIdent1},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddressKeys)
			require.NoError(t, err)

			return test{
				name:   "fail - should contain nft unlock block",
				svCtx:  &iotago.SemanticValidationContext{ExtParas: &iotago.ExternalUnlockParameters{}},
				inputs: inputs,
				tx: &iotago.Transaction{
					Essence: essence,
					UnlockBlocks: iotago.UnlockBlocks{
						&iotago.SignatureUnlockBlock{Signature: sigs[0]},
						&iotago.ReferenceUnlockBlock{Reference: 0},
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
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress:  ident2,
							MilestoneIndex: 10,
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident2AddressKeys)
			require.NoError(t, err)

			return test{
				name: "fail - sender can not unlock yet",
				svCtx: &iotago.SemanticValidationContext{ExtParas: &iotago.ExternalUnlockParameters{
					ConfMsIndex: 5,
				}},
				inputs: inputs,
				tx: &iotago.Transaction{
					Essence: essence,
					UnlockBlocks: iotago.UnlockBlocks{
						&iotago.SignatureUnlockBlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrEd25519PubKeyAndAddrMismatch,
			}
		}(),
		func() test {
			_, ident1, _ := tpkg.RandEd25519Identity()
			_, ident2, ident2AddressKeys := tpkg.RandEd25519Identity()
			inputIDs := tpkg.RandOutputIDs(1)

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.BasicOutput{
					Amount: 100,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress:  ident2,
							MilestoneIndex: 10,
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident2AddressKeys)
			require.NoError(t, err)

			return test{
				name: "fail - receiver can not unlock anymore",
				svCtx: &iotago.SemanticValidationContext{ExtParas: &iotago.ExternalUnlockParameters{
					ConfMsIndex: 5,
				}},
				inputs: inputs,
				tx: &iotago.Transaction{
					Essence: essence,
					UnlockBlocks: iotago.UnlockBlocks{
						&iotago.SignatureUnlockBlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrEd25519PubKeyAndAddrMismatch,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valFunc := iotago.TxSemanticInputUnlocks()

			err := tt.tx.SemanticallyValidate(tt.svCtx, tt.inputs, valFunc)
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
		svCtx   *iotago.SemanticValidationContext
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
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
				// unlocked by ident1 as it is not expired
				inputIDs[1]: &iotago.BasicOutput{
					Amount: 500,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: ident2,
							Amount:        420,
						},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress:  ident2,
							MilestoneIndex: 10,
						},
					},
				},
				// unlocked by ident2 as it is expired
				inputIDs[2]: &iotago.BasicOutput{
					Amount: 500,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: ident2,
							Amount:        420,
						},
						&iotago.ExpirationUnlockCondition{
							ReturnAddress:  ident2,
							MilestoneIndex: 2,
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs{
					&iotago.BasicOutput{
						Amount: 180,
						Conditions: iotago.UnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
					&iotago.BasicOutput{
						// return via ident1 + reclaim
						Amount: 420 + 500,
						Conditions: iotago.UnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident2},
						},
					},
				},
			}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys, ident2AddrKeys)
			require.NoError(t, err)

			return test{
				name: "ok",
				svCtx: &iotago.SemanticValidationContext{ExtParas: &iotago.ExternalUnlockParameters{
					ConfMsIndex: 5,
				}},
				inputs: inputs,
				tx: &iotago.Transaction{
					Essence: essence,
					UnlockBlocks: iotago.UnlockBlocks{
						&iotago.SignatureUnlockBlock{Signature: sigs[0]},
						&iotago.ReferenceUnlockBlock{Reference: 0},
						&iotago.SignatureUnlockBlock{Signature: sigs[1]},
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
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs{
					&iotago.BasicOutput{
						Amount: 100,
						Conditions: iotago.UnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "fail - unbalanced, more on output than input",
				svCtx: &iotago.SemanticValidationContext{ExtParas: &iotago.ExternalUnlockParameters{
					ConfMsIndex: 5,
				}},
				inputs: inputs,
				tx: &iotago.Transaction{
					Essence: essence,
					UnlockBlocks: iotago.UnlockBlocks{
						&iotago.SignatureUnlockBlock{Signature: sigs[0]},
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
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs{
					&iotago.BasicOutput{
						Amount: 50,
						Conditions: iotago.UnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "fail - unbalanced, more on input than output",
				svCtx: &iotago.SemanticValidationContext{ExtParas: &iotago.ExternalUnlockParameters{
					ConfMsIndex: 5,
				}},
				inputs: inputs,
				tx: &iotago.Transaction{
					Essence: essence,
					UnlockBlocks: iotago.UnlockBlocks{
						&iotago.SignatureUnlockBlock{Signature: sigs[0]},
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
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.StorageDepositReturnUnlockCondition{
							ReturnAddress: ident2,
							Amount:        420,
						},
						// not yet expired, so ident1 needs to unlock
						&iotago.ExpirationUnlockCondition{
							ReturnAddress:  ident2,
							MilestoneIndex: 10,
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs{
					&iotago.BasicOutput{
						Amount: 500,
						Conditions: iotago.UnlockConditions{
							&iotago.AddressUnlockCondition{Address: ident1},
						},
					},
				},
			}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "fail - return not fulfilled",
				svCtx: &iotago.SemanticValidationContext{ExtParas: &iotago.ExternalUnlockParameters{
					ConfMsIndex: 5,
				}},
				inputs: inputs,
				tx: &iotago.Transaction{
					Essence: essence,
					UnlockBlocks: iotago.UnlockBlocks{
						&iotago.SignatureUnlockBlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrReturnAmountNotFulFilled,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := tt.tx.SemanticallyValidate(tt.svCtx, tt.inputs,
				iotago.TxSemanticInputUnlocks(),
				iotago.TxSemanticDeposit(),
			)
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
	foundryTokenTag := tpkg.Rand12ByteArray()
	foundryMaxSupply := new(big.Int).SetInt64(1000)
	foundryMintedSupply := new(big.Int).SetInt64(500)

	inUnrelatedFoundryOutput := &iotago.FoundryOutput{
		Amount:       100,
		SerialNumber: 0,
		TokenTag:     foundryTokenTag,
		TokenScheme: &iotago.SimpleTokenScheme{
			MintedTokens:  foundryMintedSupply,
			MeltedTokens:  big.NewInt(0),
			MaximumSupply: foundryMaxSupply,
		},
		Conditions: iotago.UnlockConditions{
			&iotago.ImmutableAliasUnlockCondition{Address: foundryAliasIdent},
		},
	}

	outUnrelatedFoundryOutput := &iotago.FoundryOutput{
		Amount:       100,
		SerialNumber: 0,
		TokenTag:     foundryTokenTag,
		TokenScheme: &iotago.SimpleTokenScheme{
			MintedTokens:  foundryMintedSupply,
			MeltedTokens:  big.NewInt(0),
			MaximumSupply: foundryMaxSupply,
		},
		Conditions: iotago.UnlockConditions{
			&iotago.ImmutableAliasUnlockCondition{Address: foundryAliasIdent},
		},
	}

	type test struct {
		name    string
		svCtx   *iotago.SemanticValidationContext
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
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount:       100,
					NativeTokens: nativeTokens[ntCount/2:],
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs{
					&iotago.BasicOutput{
						Amount:       200,
						NativeTokens: nativeTokens,
						Conditions: iotago.UnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}

			return test{
				name:   "ok",
				svCtx:  &iotago.SemanticValidationContext{ExtParas: &iotago.ExternalUnlockParameters{}},
				inputs: inputs,
				tx: &iotago.Transaction{
					Essence:      essence,
					UnlockBlocks: iotago.UnlockBlocks{},
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
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs{
					&iotago.BasicOutput{
						Amount:       200,
						NativeTokens: tpkg.RandSortNativeTokens(outCount),
						Conditions: iotago.UnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}

			return test{
				name:   "fail - exceeds limit (in+out)",
				svCtx:  &iotago.SemanticValidationContext{ExtParas: &iotago.ExternalUnlockParameters{}},
				inputs: inputs,
				tx: &iotago.Transaction{
					Essence:      essence,
					UnlockBlocks: iotago.UnlockBlocks{},
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
					Conditions: iotago.UnlockConditions{
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
				Outputs: iotago.Outputs{
					&iotago.BasicOutput{
						Amount:       100,
						NativeTokens: cpyNativeTokens,
						Conditions: iotago.UnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}

			return test{
				name:   "fail - unbalanced on output",
				svCtx:  &iotago.SemanticValidationContext{ExtParas: &iotago.ExternalUnlockParameters{}},
				inputs: inputs,
				tx: &iotago.Transaction{
					Essence:      essence,
					UnlockBlocks: iotago.UnlockBlocks{},
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
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				inputIDs[1]: &iotago.BasicOutput{
					Amount:       100,
					NativeTokens: nativeTokens[ntCount/2:],
					Conditions: iotago.UnlockConditions{
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
				Outputs: iotago.Outputs{
					&iotago.BasicOutput{
						Amount:       100,
						NativeTokens: cpyNativeTokens[:ntCount/2],
						Conditions: iotago.UnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
					&iotago.BasicOutput{
						Amount:       100,
						NativeTokens: cpyNativeTokens[ntCount/2:],
						Conditions: iotago.UnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
					},
					outUnrelatedFoundryOutput,
				},
			}

			return test{
				name:   "fail - unbalanced with unrelated foundry in term of new output tokens",
				svCtx:  &iotago.SemanticValidationContext{ExtParas: &iotago.ExternalUnlockParameters{}},
				inputs: inputs,
				tx: &iotago.Transaction{
					Essence:      essence,
					UnlockBlocks: iotago.UnlockBlocks{},
				},
				wantErr: iotago.ErrNativeTokenSumUnbalanced,
			}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tx.SemanticallyValidate(tt.svCtx, tt.inputs, iotago.TxSemanticNativeTokens())
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
		svCtx   *iotago.SemanticValidationContext
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
					Amount: 100,
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs{
					&iotago.BasicOutput{
						Amount: 1337,
						Conditions: iotago.UnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Blocks: iotago.FeatureBlocks{
							&iotago.SenderFeatureBlock{Address: ident1},
						},
					},
				},
			}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name:   "ok",
				svCtx:  &iotago.SemanticValidationContext{ExtParas: &iotago.ExternalUnlockParameters{}},
				inputs: inputs,
				tx: &iotago.Transaction{
					Essence: essence,
					UnlockBlocks: iotago.UnlockBlocks{
						&iotago.SignatureUnlockBlock{Signature: sigs[0]},
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
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs{
					&iotago.BasicOutput{
						Amount: 1337,
						Conditions: iotago.UnlockConditions{
							&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						},
						Blocks: iotago.FeatureBlocks{
							&iotago.SenderFeatureBlock{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name:   "fail - sender not unlocked",
				svCtx:  &iotago.SemanticValidationContext{ExtParas: &iotago.ExternalUnlockParameters{}},
				inputs: inputs,
				tx: &iotago.Transaction{
					Essence: essence,
					UnlockBlocks: iotago.UnlockBlocks{
						&iotago.SignatureUnlockBlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrSenderFeatureBlockNotUnlocked,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := tt.tx.SemanticallyValidate(tt.svCtx, tt.inputs,
				iotago.TxSemanticInputUnlocks(),
				iotago.TxSemanticOutputsSender(),
			)
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
		svCtx   *iotago.SemanticValidationContext
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
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.TimelockUnlockCondition{
							MilestoneIndex: 5,
							UnixTime:       1337,
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "ok",
				svCtx: &iotago.SemanticValidationContext{ExtParas: &iotago.ExternalUnlockParameters{
					ConfMsIndex: 10, ConfUnix: 6666,
				}},
				inputs: inputs,
				tx: &iotago.Transaction{
					Essence: essence,
					UnlockBlocks: iotago.UnlockBlocks{
						&iotago.SignatureUnlockBlock{Signature: sigs[0]},
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
					Conditions: iotago.UnlockConditions{
						&iotago.AddressUnlockCondition{Address: ident1},
						&iotago.TimelockUnlockCondition{
							MilestoneIndex: 15,
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}
			sigs, err := essence.Sign(inputIDs.OrderedSet(inputs).MustCommitment(), ident1AddrKeys)
			require.NoError(t, err)

			return test{
				name: "fail - ms index timelock not expired",
				svCtx: &iotago.SemanticValidationContext{ExtParas: &iotago.ExternalUnlockParameters{
					ConfMsIndex: 10,
				}},
				inputs: inputs,
				tx: &iotago.Transaction{
					Essence: essence,
					UnlockBlocks: iotago.UnlockBlocks{
						&iotago.SignatureUnlockBlock{Signature: sigs[0]},
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
					Conditions: iotago.UnlockConditions{
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
				svCtx: &iotago.SemanticValidationContext{ExtParas: &iotago.ExternalUnlockParameters{
					ConfUnix: 666,
				}},
				inputs: inputs,
				tx: &iotago.Transaction{
					Essence: essence,
					UnlockBlocks: iotago.UnlockBlocks{
						&iotago.SignatureUnlockBlock{Signature: sigs[0]},
					},
				},
				wantErr: iotago.ErrTimelockNotExpired,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valFunc := iotago.TxSemanticTimelock()

			err := tt.tx.SemanticallyValidate(tt.svCtx, tt.inputs, valFunc)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}
