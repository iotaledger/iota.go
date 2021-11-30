package iotago_test

import (
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

func TestTxSemanticTimelocks(t *testing.T) {

}
