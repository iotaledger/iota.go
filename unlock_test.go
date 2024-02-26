//nolint:scopelint
package iotago_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
	"github.com/iotaledger/iota.go/v4/tpkg/frameworks"
)

func TestUnlock_DeSerialize(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name:   "ok - signature",
			Source: tpkg.RandEd25519SignatureUnlock(),
			Target: &iotago.SignatureUnlock{},
		},
		{
			Name:   "ok - reference",
			Source: tpkg.RandReferenceUnlock(),
			Target: &iotago.ReferenceUnlock{},
		},
		{
			Name:   "ok - account",
			Source: tpkg.RandAccountUnlock(),
			Target: &iotago.AccountUnlock{},
		},
		{
			Name:   "ok - anchor",
			Source: tpkg.RandAnchorUnlock(),
			Target: &iotago.AnchorUnlock{},
		},
		{
			Name:   "ok - NFT",
			Source: tpkg.RandNFTUnlock(),
			Target: &iotago.NFTUnlock{},
		},
		{
			Name:   "ok - Multi",
			Source: tpkg.RandMultiUnlock(),
			Target: &iotago.MultiUnlock{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

func TestSignatureUniqueAndReferenceUnlocksValidator(t *testing.T) {
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
				&iotago.AnchorUnlock{Reference: 2},
				&iotago.NFTUnlock{Reference: 3},
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
				&iotago.AnchorUnlock{Reference: 2},
				&iotago.MultiUnlock{
					Unlocks: []iotago.Unlock{
						&iotago.NFTUnlock{Reference: 3},
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
			wantErr: iotago.ErrSignatureUnlockNotUnique,
		},
		{
			name: "fail - signature reuse outside and inside the multi unlocks - 1",
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
			wantErr: iotago.ErrSignatureUnlockNotUnique,
		},
		{
			name: "fail - signature reuse outside and inside the multi unlocks - 2",
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
				&iotago.SignatureUnlock{Signature: &iotago.Ed25519Signature{
					PublicKey: [32]byte{},
					Signature: [64]byte{},
				}},
			},
			wantErr: iotago.ErrSignatureUnlockNotUnique,
		},
		{
			name: "ok - duplicate ed25519 sig block in different multi unlocks",
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
			wantErr: nil,
		},
		{
			name: "fail - duplicate multi unlock",
			unlocks: iotago.Unlocks{
				&iotago.MultiUnlock{
					Unlocks: []iotago.Unlock{
						&iotago.SignatureUnlock{Signature: &iotago.Ed25519Signature{
							PublicKey: [32]byte{},
							Signature: [64]byte{},
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
			wantErr: iotago.ErrMultiUnlockNotUnique,
		},
		{
			name: "fail - reference unlock invalid reference",
			unlocks: iotago.Unlocks{
				tpkg.RandEd25519SignatureUnlock(),
				tpkg.RandEd25519SignatureUnlock(),
				&iotago.ReferenceUnlock{Reference: 1337},
			},
			wantErr: iotago.ErrReferentialUnlockInvalid,
		},
		{
			name: "fail - reference unlock invalid reference in multi unlock",
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
			name: "fail - reference unlock references non sig unlock",
			unlocks: iotago.Unlocks{
				tpkg.RandEd25519SignatureUnlock(),
				&iotago.ReferenceUnlock{Reference: 0},
				&iotago.ReferenceUnlock{Reference: 1},
			},
			wantErr: iotago.ErrReferentialUnlockInvalid,
		},
		{
			name: "fail - reference unlock references non sig unlock in multi unlock",
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
		{
			name: "fail - referenced a multi unlock in in itself",
			unlocks: iotago.Unlocks{
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
			valFunc := iotago.SignatureUniqueAndReferenceUnlocksValidator(tpkg.ZeroCostTestAPI)
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
