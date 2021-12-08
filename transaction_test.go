package iotago_test

import (
	"math/big"
	"testing"

	"github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/ed25519"
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
				dustDepositReturn        uint64 = OneMi / 2
				nativeTokenTransfer1            = tpkg.RandSortNativeTokens(50)
				nativeTokenTransfer2            = tpkg.RandSortNativeTokens(50)
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
				inputIDs[0]: &iotago.ExtendedOutput{Address: ident1, Amount: defaultAmount},
				inputIDs[1]: &iotago.ExtendedOutput{Address: ident2, Amount: defaultAmount, NativeTokens: nativeTokenTransfer1},
				inputIDs[2]: &iotago.ExtendedOutput{Address: ident2, Amount: defaultAmount, NativeTokens: nativeTokenTransfer2},
				inputIDs[3]: &iotago.ExtendedOutput{Address: ident2, Amount: defaultAmount, Blocks: iotago.FeatureBlocks{
					&iotago.ExpirationMilestoneIndexFeatureBlock{MilestoneIndex: 500},
					&iotago.SenderFeatureBlock{Address: ident1},
				}},
				inputIDs[4]: &iotago.ExtendedOutput{Address: ident2, Amount: defaultAmount, Blocks: iotago.FeatureBlocks{
					&iotago.TimelockMilestoneIndexFeatureBlock{MilestoneIndex: 500},
				}},
				inputIDs[5]: &iotago.ExtendedOutput{Address: ident2, Amount: defaultAmount + dustDepositReturn, Blocks: iotago.FeatureBlocks{
					&iotago.TimelockMilestoneIndexFeatureBlock{MilestoneIndex: 500},
					&iotago.ExpirationMilestoneIndexFeatureBlock{MilestoneIndex: 900},
					&iotago.SenderFeatureBlock{Address: ident1},
					&iotago.DustDepositReturnFeatureBlock{Amount: dustDepositReturn},
				}},
				inputIDs[6]: &iotago.AliasOutput{
					Amount:               defaultAmount,
					NativeTokens:         nil,
					AliasID:              iotago.AliasID{},
					StateController:      ident3,
					GovernanceController: ident4,
					StateIndex:           0,
					StateMetadata:        []byte("gov transitioning"),
					FoundryCounter:       0,
					Blocks:               nil,
				},
				inputIDs[7]: &iotago.AliasOutput{
					Amount:               defaultAmount + defaultAmount, // to fund also the new alias output
					NativeTokens:         nil,
					AliasID:              iotago.AliasID{},
					StateController:      ident3,
					GovernanceController: ident4,
					StateIndex:           5,
					StateMetadata:        []byte("current state"),
					FoundryCounter:       5,
					Blocks:               nil,
				},
				inputIDs[8]: &iotago.AliasOutput{
					Amount:               defaultAmount,
					NativeTokens:         nil,
					AliasID:              iotago.AliasID{},
					StateController:      ident3,
					GovernanceController: ident3,
					StateIndex:           0,
					StateMetadata:        []byte("going to be destroyed"),
					FoundryCounter:       0,
					Blocks:               nil,
				},
				inputIDs[9]: &iotago.FoundryOutput{
					Address:           iotago.AliasIDFromOutputID(inputIDs[7]).ToAddress(),
					Amount:            defaultAmount,
					NativeTokens:      nil,
					SerialNumber:      1,
					TokenTag:          foundry1Ident3TokenTag,
					CirculatingSupply: new(big.Int).SetUint64(100),
					MaximumSupply:     new(big.Int).SetUint64(1000),
					TokenScheme:       &iotago.SimpleTokenScheme{},
					Blocks:            nil,
				},
				inputIDs[10]: &iotago.FoundryOutput{
					Address:           iotago.AliasIDFromOutputID(inputIDs[7]).ToAddress(),
					Amount:            defaultAmount,
					NativeTokens:      nil, // filled out later
					SerialNumber:      2,
					TokenTag:          foundry2Ident3TokenTag,
					CirculatingSupply: new(big.Int).SetUint64(100),
					MaximumSupply:     new(big.Int).SetUint64(1000),
					TokenScheme:       &iotago.SimpleTokenScheme{},
					Blocks:            nil,
				},
				inputIDs[11]: &iotago.FoundryOutput{
					Address:           iotago.AliasIDFromOutputID(inputIDs[7]).ToAddress(),
					Amount:            defaultAmount,
					NativeTokens:      nil,
					SerialNumber:      3,
					TokenTag:          foundry3Ident3TokenTag,
					CirculatingSupply: new(big.Int).SetUint64(100),
					MaximumSupply:     new(big.Int).SetUint64(1000),
					TokenScheme:       &iotago.SimpleTokenScheme{},
					Blocks:            nil,
				},
				inputIDs[12]: &iotago.FoundryOutput{
					Address:           iotago.AliasIDFromOutputID(inputIDs[7]).ToAddress(),
					Amount:            defaultAmount,
					NativeTokens:      nil,
					SerialNumber:      4,
					TokenTag:          foundry4Ident3TokenTag,
					CirculatingSupply: new(big.Int).SetUint64(100),
					MaximumSupply:     new(big.Int).SetUint64(1000),
					TokenScheme:       &iotago.SimpleTokenScheme{},
					Blocks:            nil,
				},
				inputIDs[13]: &iotago.NFTOutput{
					Address:           ident3,
					Amount:            defaultAmount,
					NativeTokens:      nil,
					NFTID:             nft1ID,
					ImmutableMetadata: []byte("transfer to 4"),
					Blocks: iotago.FeatureBlocks{
						&iotago.IssuerFeatureBlock{Address: ident3},
					},
				},
				inputIDs[14]: &iotago.NFTOutput{
					Address:           ident4,
					Amount:            defaultAmount,
					NativeTokens:      nil,
					NFTID:             nft2ID,
					ImmutableMetadata: []byte("going to be destroyed"),
					Blocks: iotago.FeatureBlocks{
						&iotago.IssuerFeatureBlock{Address: ident3},
					},
				},
				inputIDs[15]: &iotago.ExtendedOutput{Address: iotago.NFTID(nft1ID).ToAddress(), Amount: defaultAmount},
			}

			foundry1Ident3NativeTokenID := inputs[inputIDs[9]].(*iotago.FoundryOutput).MustNativeTokenID()
			foundry2Ident3NativeTokenID := inputs[inputIDs[10]].(*iotago.FoundryOutput).MustNativeTokenID()

			inputs[inputIDs[10]].(*iotago.FoundryOutput).NativeTokens = iotago.NativeTokens{
				{
					ID:     foundry2Ident3NativeTokenID,
					Amount: big.NewInt(100),
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs{
					&iotago.ExtendedOutput{Address: ident5, Amount: defaultAmount},
					&iotago.ExtendedOutput{Address: ident3, Amount: defaultAmount, NativeTokens: nativeTokenTransfer1},
					&iotago.ExtendedOutput{Address: ident4, Amount: defaultAmount, NativeTokens: nativeTokenTransfer2},
					&iotago.ExtendedOutput{Address: ident2, Amount: defaultAmount},
					&iotago.ExtendedOutput{Address: ident2, Amount: defaultAmount},
					&iotago.ExtendedOutput{Address: ident1, Amount: dustDepositReturn},
					&iotago.AliasOutput{
						Amount:               defaultAmount,
						NativeTokens:         nil,
						AliasID:              iotago.AliasID{},
						StateController:      ident3,
						GovernanceController: ident4,
						StateIndex:           0,
						StateMetadata:        []byte("a new alias output"),
						FoundryCounter:       0,
						Blocks:               nil,
					},
					&iotago.AliasOutput{
						Amount:               defaultAmount,
						NativeTokens:         nil,
						AliasID:              iotago.AliasIDFromOutputID(inputIDs[6]),
						StateController:      ident3,
						GovernanceController: ident4,
						StateIndex:           0,
						StateMetadata:        []byte("gov transitioning"),
						FoundryCounter:       0,
						Blocks: iotago.FeatureBlocks{
							&iotago.MetadataFeatureBlock{Data: []byte("the gov mutation on this output")},
						},
					},
					&iotago.AliasOutput{
						Amount:               defaultAmount,
						NativeTokens:         nil,
						AliasID:              iotago.AliasIDFromOutputID(inputIDs[7]),
						StateController:      ident3,
						GovernanceController: ident4,
						StateIndex:           6,
						StateMetadata:        []byte("next state"),
						FoundryCounter:       6,
						Blocks:               nil,
					},
					// new foundry
					&iotago.FoundryOutput{
						Address:           iotago.AliasIDFromOutputID(inputIDs[7]).ToAddress(),
						Amount:            defaultAmount,
						NativeTokens:      nil,
						SerialNumber:      6,
						TokenTag:          tpkg.Rand12ByteArray(),
						CirculatingSupply: new(big.Int).SetInt64(0),
						MaximumSupply:     new(big.Int).SetInt64(1000),
						TokenScheme:       &iotago.SimpleTokenScheme{},
						Blocks:            nil,
					},
					&iotago.FoundryOutput{
						Address: iotago.AliasIDFromOutputID(inputIDs[7]).ToAddress(),
						Amount:  defaultAmount,
						NativeTokens: iotago.NativeTokens{
							{
								ID:     foundry1Ident3NativeTokenID,
								Amount: new(big.Int).SetUint64(100), // freshly minted
							},
						},
						SerialNumber:      1,
						TokenTag:          foundry1Ident3TokenTag,
						CirculatingSupply: new(big.Int).SetInt64(200),
						MaximumSupply:     new(big.Int).SetInt64(1000),
						TokenScheme:       &iotago.SimpleTokenScheme{},
						Blocks:            nil,
					},
					&iotago.FoundryOutput{
						Address: iotago.AliasIDFromOutputID(inputIDs[7]).ToAddress(),
						Amount:  defaultAmount,
						NativeTokens: iotago.NativeTokens{
							{
								ID:     foundry2Ident3NativeTokenID,
								Amount: new(big.Int).SetUint64(50), // burned to 50
							},
						},
						SerialNumber:      2,
						TokenTag:          foundry2Ident3TokenTag,
						CirculatingSupply: new(big.Int).SetInt64(50),
						MaximumSupply:     new(big.Int).SetInt64(1000),
						TokenScheme:       &iotago.SimpleTokenScheme{},
						Blocks:            nil,
					},
					&iotago.FoundryOutput{
						Address:           iotago.AliasIDFromOutputID(inputIDs[7]).ToAddress(),
						Amount:            defaultAmount,
						NativeTokens:      nil,
						SerialNumber:      3,
						TokenTag:          foundry3Ident3TokenTag,
						CirculatingSupply: new(big.Int).SetInt64(100),
						MaximumSupply:     new(big.Int).SetInt64(1000),
						TokenScheme:       &iotago.SimpleTokenScheme{},
						Blocks: iotago.FeatureBlocks{
							&iotago.MetadataFeatureBlock{Data: []byte("interesting metadata")},
						},
					},
					// from foundry 4 ident 3 destruction remainder
					&iotago.ExtendedOutput{Address: ident3, Amount: defaultAmount},
					&iotago.NFTOutput{
						Address:           ident4,
						Amount:            defaultAmount,
						NativeTokens:      nil,
						NFTID:             iotago.NFTID{},
						ImmutableMetadata: []byte("immutable metadata"),
						Blocks:            nil,
					},
					&iotago.NFTOutput{
						Address:           ident4,
						Amount:            defaultAmount,
						NativeTokens:      nil,
						NFTID:             nft1ID,
						ImmutableMetadata: []byte("transfer to 4"),
						Blocks: iotago.FeatureBlocks{
							&iotago.IssuerFeatureBlock{Address: ident3},
						},
					},
					// from NFT ident 4 destruction remainder
					&iotago.ExtendedOutput{Address: ident3, Amount: defaultAmount},
					// from NFT 1 to ident 5
					&iotago.ExtendedOutput{Address: ident5, Amount: defaultAmount},
				},
			}

			sigs, err := essence.Sign(ident1AddrKeys, ident2AddrKeys, ident3AddrKeys, ident4AddrKeys)
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
						// extended
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
				inputIDs[0]: &iotago.ExtendedOutput{Address: ident1, Amount: 100},
				inputIDs[1]: &iotago.AliasOutput{
					Amount:               100,
					AliasID:              iotago.AliasID{}, // empty on purpose as validation should resolve
					StateController:      ident1,
					GovernanceController: ident1,
				},
				inputIDs[2]: &iotago.ExtendedOutput{Address: &aliasIdent1, Amount: 100},
				inputIDs[3]: &iotago.NFTOutput{
					Address: &aliasIdent1,
					Amount:  100,
					NFTID:   nftIdent1.NFTID(),
				},
				inputIDs[4]: &iotago.ExtendedOutput{Address: nftIdent1, Amount: 100},
				// unlockable by sender as expired
				inputIDs[5]: &iotago.ExtendedOutput{
					Address: ident1,
					Amount:  100,
					Blocks: iotago.FeatureBlocks{
						&iotago.SenderFeatureBlock{Address: ident2},
						&iotago.ExpirationMilestoneIndexFeatureBlock{MilestoneIndex: 5},
					},
				},
				// not unlockable by sender as not expired
				inputIDs[6]: &iotago.ExtendedOutput{
					Address: ident1,
					Amount:  100,
					Blocks: iotago.FeatureBlocks{
						&iotago.SenderFeatureBlock{Address: ident2},
						&iotago.ExpirationMilestoneIndexFeatureBlock{MilestoneIndex: 20},
					},
				},
				inputIDs[7]: &iotago.FoundryOutput{
					Address:      &aliasIdent1,
					Amount:       100,
					SerialNumber: 0,
					TokenTag:     tpkg.Rand12ByteArray(),
					TokenScheme:  &iotago.SimpleTokenScheme{},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(ident1AddrKeys, ident2AddrKeys)
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
			inputIDs := tpkg.RandOutputIDs(8)

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.ExtendedOutput{Address: ident1, Amount: 100},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(ident2AddrKeys)
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
				inputIDs[0]: &iotago.ExtendedOutput{Address: ident1, Amount: 100},
				inputIDs[1]: &iotago.ExtendedOutput{Address: ident1, Amount: 100},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(ident1AddressKeys)
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
					Amount:               100,
					AliasID:              iotago.AliasID{},
					StateController:      ident1,
					GovernanceController: ident1,
				},
				inputIDs[1]: &iotago.ExtendedOutput{Address: &aliasIdent1, Amount: 100},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(ident1AddressKeys)
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
					Address: ident1,
					Amount:  100,
					NFTID:   iotago.NFTID{},
				},
				inputIDs[1]: &iotago.ExtendedOutput{Address: &nftIdent1, Amount: 100},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(ident1AddressKeys)
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
				inputIDs[0]: &iotago.ExtendedOutput{Address: ident1, Amount: 100, Blocks: iotago.FeatureBlocks{
					&iotago.SenderFeatureBlock{Address: ident2},
					&iotago.ExpirationMilestoneIndexFeatureBlock{MilestoneIndex: 10},
				}},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(ident2AddressKeys)
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
				inputIDs[0]: &iotago.ExtendedOutput{Address: ident1, Amount: 100, Blocks: iotago.FeatureBlocks{
					&iotago.SenderFeatureBlock{Address: ident2},
					&iotago.ExpirationMilestoneIndexFeatureBlock{MilestoneIndex: 10},
				}},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}

			sigs, err := essence.Sign(ident2AddressKeys)
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
				inputIDs[0]: &iotago.ExtendedOutput{Address: ident1, Amount: 100},
				// unlocked by ident1 as it is not expired
				inputIDs[1]: &iotago.ExtendedOutput{
					Address: ident1,
					Amount:  500,
					Blocks: iotago.FeatureBlocks{
						&iotago.DustDepositReturnFeatureBlock{Amount: 420},
						&iotago.SenderFeatureBlock{Address: ident2},
						&iotago.ExpirationMilestoneIndexFeatureBlock{MilestoneIndex: 10},
					},
				},
				// unlocked by ident2 as it is expired
				inputIDs[2]: &iotago.ExtendedOutput{
					Address: ident1,
					Amount:  500,
					Blocks: iotago.FeatureBlocks{
						&iotago.DustDepositReturnFeatureBlock{Amount: 420},
						&iotago.SenderFeatureBlock{Address: ident2},
						&iotago.ExpirationMilestoneIndexFeatureBlock{MilestoneIndex: 2},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs{
					&iotago.ExtendedOutput{
						Address: tpkg.RandEd25519Address(),
						Amount:  180,
					},
					&iotago.ExtendedOutput{
						Address: ident2,
						// return via ident1 + reclaim
						Amount: 420 + 500,
					},
				},
			}
			sigs, err := essence.Sign(ident1AddrKeys, ident2AddrKeys)
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
				inputIDs[0]: &iotago.ExtendedOutput{Address: ident1, Amount: 50},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs{
					&iotago.ExtendedOutput{
						Address: tpkg.RandEd25519Address(),
						Amount:  100,
					},
				},
			}
			sigs, err := essence.Sign(ident1AddrKeys)
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
				inputIDs[0]: &iotago.ExtendedOutput{Address: ident1, Amount: 100},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs{
					&iotago.ExtendedOutput{
						Address: tpkg.RandEd25519Address(),
						Amount:  50,
					},
				},
			}
			sigs, err := essence.Sign(ident1AddrKeys)
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
				inputIDs[0]: &iotago.ExtendedOutput{
					Address: ident1,
					Amount:  500,
					Blocks: iotago.FeatureBlocks{
						&iotago.DustDepositReturnFeatureBlock{Amount: 420},
						&iotago.SenderFeatureBlock{Address: ident2},
						// not yet expired, so ident1 needs to unlock
						&iotago.ExpirationMilestoneIndexFeatureBlock{MilestoneIndex: 10},
					},
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs{
					&iotago.ExtendedOutput{
						Address: ident1,
						Amount:  500,
					},
				},
			}
			sigs, err := essence.Sign(ident1AddrKeys)
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
	foundryTokenScheme := &iotago.SimpleTokenScheme{}
	foundryMaxSupply := new(big.Int).SetInt64(1000)
	foundryCircSupply := new(big.Int).SetInt64(500)

	inUnrelatedFoundryOutput := &iotago.FoundryOutput{
		Address:           foundryAliasIdent,
		Amount:            100,
		SerialNumber:      0,
		TokenTag:          foundryTokenTag,
		CirculatingSupply: foundryCircSupply,
		MaximumSupply:     foundryMaxSupply,
		TokenScheme:       foundryTokenScheme,
	}

	outUnrelatedFoundryOutput := &iotago.FoundryOutput{
		Address:           foundryAliasIdent,
		Amount:            100,
		SerialNumber:      0,
		TokenTag:          foundryTokenTag,
		CirculatingSupply: foundryCircSupply,
		MaximumSupply:     foundryMaxSupply,
		TokenScheme:       foundryTokenScheme,
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
			inputIDs := tpkg.RandOutputIDs(4)

			ntCount := 100
			nativeTokens := tpkg.RandSortNativeTokens(ntCount)

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.ExtendedOutput{
					Address: tpkg.RandEd25519Address(), Amount: 100,
					NativeTokens: nativeTokens[:ntCount/2],
				},
				inputIDs[1]: &iotago.ExtendedOutput{
					Address: tpkg.RandEd25519Address(), Amount: 100,
					NativeTokens: nativeTokens[ntCount/2:],
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs{
					&iotago.ExtendedOutput{
						Address:      tpkg.RandEd25519Address(),
						Amount:       200,
						NativeTokens: nativeTokens,
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
			inputIDs := tpkg.RandOutputIDs(2)

			ntCount := iotago.MaxNativeTokensCount + 1
			nativeTokens := tpkg.RandSortNativeTokens(ntCount)

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.ExtendedOutput{
					Address: tpkg.RandEd25519Address(), Amount: 100,
					NativeTokens: nativeTokens[:ntCount/2],
				},
				inputIDs[1]: &iotago.ExtendedOutput{
					Address: tpkg.RandEd25519Address(), Amount: 100,
					NativeTokens: nativeTokens[ntCount/2:],
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs{
					&iotago.ExtendedOutput{
						Address:      tpkg.RandEd25519Address(),
						Amount:       200,
						NativeTokens: nativeTokens,
					},
				},
			}

			return test{
				name:   "fail - exceeds limit (just in)",
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
			inputIDs := tpkg.RandOutputIDs(2)

			inCount := 20
			outCount := 250

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.ExtendedOutput{
					Address: tpkg.RandEd25519Address(), Amount: 100,
					NativeTokens: tpkg.RandSortNativeTokens(inCount),
				},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs{
					&iotago.ExtendedOutput{
						Address:      tpkg.RandEd25519Address(),
						Amount:       200,
						NativeTokens: tpkg.RandSortNativeTokens(outCount),
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
			inputIDs := tpkg.RandOutputIDs(2)

			ntCount := 20
			nativeTokens := tpkg.RandSortNativeTokens(ntCount)

			inputs := iotago.OutputSet{
				inputIDs[0]: &iotago.ExtendedOutput{
					Address: tpkg.RandEd25519Address(), Amount: 100,
					NativeTokens: nativeTokens[:ntCount/2],
				},
				inputIDs[1]: &iotago.ExtendedOutput{
					Address: tpkg.RandEd25519Address(), Amount: 100,
					NativeTokens: nativeTokens[ntCount/2:],
				},
			}

			// unbalance
			cpyNativeTokens := nativeTokens.Clone()
			cpyNativeTokens[ntCount/2].Amount = tpkg.RandUint256()

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs{
					&iotago.ExtendedOutput{
						Address:      tpkg.RandEd25519Address(),
						Amount:       100,
						NativeTokens: cpyNativeTokens[:ntCount/2],
					},
					&iotago.ExtendedOutput{
						Address:      tpkg.RandEd25519Address(),
						Amount:       100,
						NativeTokens: cpyNativeTokens[ntCount/2:],
					},
				},
			}

			return test{
				name:   "fail - unbalanced",
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
				inputIDs[0]: &iotago.ExtendedOutput{
					Address: tpkg.RandEd25519Address(), Amount: 100,
					NativeTokens: nativeTokens[:ntCount/2],
				},
				inputIDs[1]: &iotago.ExtendedOutput{
					Address: tpkg.RandEd25519Address(), Amount: 100,
					NativeTokens: nativeTokens[ntCount/2:],
				},
				inputIDs[2]: inUnrelatedFoundryOutput,
			}

			// unbalance
			cpyNativeTokens := nativeTokens.Clone()
			cpyNativeTokens[ntCount/2].Amount = tpkg.RandUint256()

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs{
					&iotago.ExtendedOutput{
						Address:      tpkg.RandEd25519Address(),
						Amount:       100,
						NativeTokens: cpyNativeTokens[:ntCount/2],
					},
					&iotago.ExtendedOutput{
						Address:      tpkg.RandEd25519Address(),
						Amount:       100,
						NativeTokens: cpyNativeTokens[ntCount/2:],
					},
					outUnrelatedFoundryOutput,
				},
			}

			// this test circumvents the short path of the validation func when no foundry exists
			return test{
				name:   "fail - unbalanced with unrelated foundry",
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
				inputIDs[0]: &iotago.ExtendedOutput{
					Address: tpkg.RandEd25519Address(), Amount: 100,
					NativeTokens: nativeTokens[:ntCount/2],
				},
				inputIDs[1]: &iotago.ExtendedOutput{
					Address: tpkg.RandEd25519Address(), Amount: 100,
					NativeTokens: nativeTokens[ntCount/2:],
				},
				inputIDs[2]: inUnrelatedFoundryOutput,
			}

			// add a new token to the output side
			cpyNativeTokens := nativeTokens.Clone()
			cpyNativeTokens = append(cpyNativeTokens, tpkg.RandNativeToken())

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs{
					&iotago.ExtendedOutput{
						Address:      tpkg.RandEd25519Address(),
						Amount:       100,
						NativeTokens: cpyNativeTokens[:ntCount/2],
					},
					&iotago.ExtendedOutput{
						Address:      tpkg.RandEd25519Address(),
						Amount:       100,
						NativeTokens: cpyNativeTokens[ntCount/2:],
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

			err := tt.tx.SemanticallyValidate(tt.svCtx, tt.inputs,
				iotago.TxSemanticNativeTokens(),
			)
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
				inputIDs[0]: &iotago.ExtendedOutput{Address: ident1, Amount: 100},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs{
					&iotago.ExtendedOutput{
						Address: tpkg.RandEd25519Address(),
						Amount:  1337,
						Blocks: iotago.FeatureBlocks{
							&iotago.SenderFeatureBlock{Address: ident1},
						},
					},
				},
			}
			sigs, err := essence.Sign(ident1AddrKeys)
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
				inputIDs[0]: &iotago.ExtendedOutput{Address: ident1, Amount: 100},
			}

			essence := &iotago.TransactionEssence{
				Inputs: inputIDs.UTXOInputs(),
				Outputs: iotago.Outputs{
					&iotago.ExtendedOutput{
						Address: tpkg.RandEd25519Address(),
						Amount:  1337,
						Blocks: iotago.FeatureBlocks{
							&iotago.SenderFeatureBlock{Address: tpkg.RandEd25519Address()},
						},
					},
				},
			}
			sigs, err := essence.Sign(ident1AddrKeys)
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
				inputIDs[0]: &iotago.ExtendedOutput{
					Address: ident1, Amount: 100,
					Blocks: iotago.FeatureBlocks{
						&iotago.TimelockMilestoneIndexFeatureBlock{
							MilestoneIndex: 5,
						},
						&iotago.TimelockUnixFeatureBlock{
							UnixTime: 1337,
						},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}
			sigs, err := essence.Sign(ident1AddrKeys)
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
				inputIDs[0]: &iotago.ExtendedOutput{
					Address: ident1, Amount: 100,
					Blocks: iotago.FeatureBlocks{
						&iotago.TimelockMilestoneIndexFeatureBlock{MilestoneIndex: 15},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}
			sigs, err := essence.Sign(ident1AddrKeys)
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
				inputIDs[0]: &iotago.ExtendedOutput{
					Address: ident1, Amount: 100,
					Blocks: iotago.FeatureBlocks{
						&iotago.TimelockUnixFeatureBlock{UnixTime: 1337},
					},
				},
			}

			essence := &iotago.TransactionEssence{Inputs: inputIDs.UTXOInputs()}
			sigs, err := essence.Sign(ident1AddrKeys)
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
