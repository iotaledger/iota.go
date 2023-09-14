//nolint:scopelint
package iotago_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestUnlock_DeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok - signature",
			source: tpkg.RandEd25519SignatureUnlock(),
			target: &iotago.SignatureUnlock{},
		},
		{
			name:   "ok - reference",
			source: tpkg.RandReferenceUnlock(),
			target: &iotago.ReferenceUnlock{},
		},
		{
			name:   "ok - account",
			source: tpkg.RandAccountUnlock(),
			target: &iotago.AccountUnlock{},
		},
		{
			name:   "ok - NFT",
			source: tpkg.RandNFTUnlock(),
			target: &iotago.NFTUnlock{},
		},
		{
			name:   "ok - Multi",
			source: tpkg.RandMultiUnlock(),
			target: &iotago.MultiUnlock{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestUnlocksSigUniqueAndRefValidator(t *testing.T) {
	tests := []struct {
		name    string
		unlocks iotago.Unlocks
		wantErr error
	}{
		{
			name: "ok",
			unlocks: iotago.Unlocks{
				tpkg.RandEd25519SignatureUnlock(),
				tpkg.RandEd25519SignatureUnlock(),
				&iotago.ReferenceUnlock{Reference: 0},
			},
			wantErr: nil,
		},
		{
			name: "ok - chainable referential unlock",
			unlocks: iotago.Unlocks{
				tpkg.RandEd25519SignatureUnlock(),
				&iotago.AccountUnlock{Reference: 0},
				&iotago.AccountUnlock{Reference: 1},
				&iotago.NFTUnlock{Reference: 2},
			},
			wantErr: nil,
		},
		{
			name: "ok - multi unlock",
			unlocks: iotago.Unlocks{
				tpkg.RandEd25519SignatureUnlock(),
				tpkg.RandEd25519SignatureUnlock(),
				&iotago.MultiUnlock{
					Unlocks: []iotago.Unlock{
						&iotago.ReferenceUnlock{Reference: 0},
						&iotago.ReferenceUnlock{Reference: 1},
						&iotago.EmptyUnlock{},
						tpkg.RandEd25519SignatureUnlock(),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - chainable referential unlock in multi unlock",
			unlocks: iotago.Unlocks{
				tpkg.RandEd25519SignatureUnlock(),
				&iotago.AccountUnlock{Reference: 0},
				&iotago.AccountUnlock{Reference: 1},
				&iotago.MultiUnlock{
					Unlocks: []iotago.Unlock{
						&iotago.NFTUnlock{Reference: 2},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - duplicate ed25519 sig block",
			unlocks: iotago.Unlocks{
				&iotago.SignatureUnlock{Signature: &iotago.Ed25519Signature{
					PublicKey: [32]byte{},
					Signature: [64]byte{},
				}},
				&iotago.SignatureUnlock{Signature: &iotago.Ed25519Signature{
					PublicKey: [32]byte{},
					Signature: [64]byte{},
				}},
			},
			wantErr: iotago.ErrSigUnlockNotUnique,
		},
		{
			name: "fail - duplicate ed25519 sig block in multi unlock",
			unlocks: iotago.Unlocks{
				&iotago.SignatureUnlock{Signature: &iotago.Ed25519Signature{
					PublicKey: [32]byte{},
					Signature: [64]byte{},
				}},
				&iotago.MultiUnlock{
					Unlocks: []iotago.Unlock{
						&iotago.SignatureUnlock{Signature: &iotago.Ed25519Signature{
							PublicKey: [32]byte{},
							Signature: [64]byte{},
						}},
					},
				},
			},
			wantErr: iotago.ErrSigUnlockNotUnique,
		},
		{
			name: "fail - duplicate ed25519 sig block in multi unlock - 2",
			unlocks: iotago.Unlocks{
				&iotago.MultiUnlock{
					Unlocks: []iotago.Unlock{
						&iotago.SignatureUnlock{Signature: &iotago.Ed25519Signature{
							PublicKey: [32]byte{},
							Signature: [64]byte{},
						}},
						&iotago.SignatureUnlock{Signature: &iotago.Ed25519Signature{
							PublicKey: [32]byte{0x01},
							Signature: [64]byte{0x01},
						}},
					},
				},
				&iotago.MultiUnlock{
					Unlocks: []iotago.Unlock{
						&iotago.SignatureUnlock{Signature: &iotago.Ed25519Signature{
							PublicKey: [32]byte{},
							Signature: [64]byte{},
						}},
					},
				},
			},
			wantErr: iotago.ErrSigUnlockNotUnique,
		},
		{
			name: "fail - duplicate multi unlock",
			unlocks: iotago.Unlocks{
				&iotago.MultiUnlock{
					Unlocks: []iotago.Unlock{
						&iotago.EmptyUnlock{},
					},
				},
				&iotago.MultiUnlock{
					Unlocks: []iotago.Unlock{
						&iotago.EmptyUnlock{},
					},
				},
			},
			wantErr: iotago.ErrMultiUnlockNotUnique,
		},
		{
			name: "fail - reference unlock invalid ref",
			unlocks: iotago.Unlocks{
				tpkg.RandEd25519SignatureUnlock(),
				tpkg.RandEd25519SignatureUnlock(),
				&iotago.ReferenceUnlock{Reference: 1337},
			},
			wantErr: iotago.ErrReferentialUnlockInvalid,
		},
		{
			name: "fail - reference unlock invalid ref in multi unlock",
			unlocks: iotago.Unlocks{
				tpkg.RandEd25519SignatureUnlock(),
				tpkg.RandEd25519SignatureUnlock(),
				&iotago.MultiUnlock{
					Unlocks: []iotago.Unlock{
						&iotago.ReferenceUnlock{Reference: 1337},
					},
				},
			},
			wantErr: iotago.ErrReferentialUnlockInvalid,
		},
		{
			name: "fail - reference unlock refs non sig unlock",
			unlocks: iotago.Unlocks{
				tpkg.RandEd25519SignatureUnlock(),
				&iotago.ReferenceUnlock{Reference: 0},
				&iotago.ReferenceUnlock{Reference: 1},
			},
			wantErr: iotago.ErrReferentialUnlockInvalid,
		},
		{
			name: "fail - reference unlock refs non sig unlock in multi unlock",
			unlocks: iotago.Unlocks{
				tpkg.RandEd25519SignatureUnlock(),
				&iotago.ReferenceUnlock{Reference: 0},
				&iotago.MultiUnlock{
					Unlocks: []iotago.Unlock{
						&iotago.ReferenceUnlock{Reference: 1},
					},
				},
			},
			wantErr: iotago.ErrReferentialUnlockInvalid,
		},
		{
			name: "fail - empty unlock outside multi unlock",
			unlocks: iotago.Unlocks{
				tpkg.RandEd25519SignatureUnlock(),
				&iotago.EmptyUnlock{},
			},
			wantErr: iotago.ErrEmptyUnlockOutsideMultiUnlock,
		},
		{
			name: "fail - nested multi unlock",
			unlocks: iotago.Unlocks{
				&iotago.MultiUnlock{
					Unlocks: []iotago.Unlock{
						&iotago.MultiUnlock{
							Unlocks: []iotago.Unlock{
								tpkg.RandEd25519SignatureUnlock(),
							},
						},
					},
				},
			},
			wantErr: iotago.ErrNestedMultiUnlock,
		},
		{
			name: "ok - referenced a multi unlock",
			unlocks: iotago.Unlocks{
				&iotago.MultiUnlock{
					Unlocks: []iotago.Unlock{
						tpkg.RandEd25519SignatureUnlock(),
					},
				},
				&iotago.ReferenceUnlock{Reference: 0},
			},
			wantErr: nil,
		},
		{
			name: "fail - referenced a multi unlock in a multi unlock",
			unlocks: iotago.Unlocks{
				&iotago.MultiUnlock{
					Unlocks: []iotago.Unlock{
						tpkg.RandEd25519SignatureUnlock(),
					},
				},
				&iotago.MultiUnlock{
					Unlocks: []iotago.Unlock{
						&iotago.ReferenceUnlock{Reference: 0},
					},
				},
			},
			wantErr: iotago.ErrReferentialUnlockInvalid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valFunc := iotago.UnlocksSigUniqueAndRefValidator(tpkg.TestAPI)
			var runErr error
			for index, unlock := range tt.unlocks {
				if err := valFunc(index, unlock); err != nil {
					runErr = err
				}
			}
			require.ErrorIs(t, runErr, tt.wantErr)
		})
	}
}
