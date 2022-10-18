package iotago_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
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
			name:   "ok - alias",
			source: tpkg.RandAliasUnlock(),
			target: &iotago.AliasUnlock{},
		},
		{
			name:   "ok - NFT",
			source: tpkg.RandNFTUnlock(),
			target: &iotago.NFTUnlock{},
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
				&iotago.AliasUnlock{Reference: 0},
				&iotago.AliasUnlock{Reference: 1},
				&iotago.NFTUnlock{Reference: 2},
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
			name: "fail - reference unlock invalid ref",
			unlocks: iotago.Unlocks{
				tpkg.RandEd25519SignatureUnlock(),
				tpkg.RandEd25519SignatureUnlock(),
				&iotago.ReferenceUnlock{Reference: 1337},
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valFunc := iotago.UnlocksSigUniqueAndRefValidator()
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
