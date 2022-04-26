package iotago_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/iota.go/v3"
	"github.com/stretchr/testify/assert"
)

func TestUnlockBlockSelector(t *testing.T) {
	_, err := iotago.UnlockBlockSelector(100)
	assert.True(t, errors.Is(err, iotago.ErrUnknownUnlockBlockType))
}

func TestUnlockBlock_DeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok - signature",
			source: tpkg.RandEd25519SignatureUnlockBlock(),
			target: &iotago.SignatureUnlockBlock{},
		},
		{
			name:   "ok - reference",
			source: tpkg.RandReferenceUnlockBlock(),
			target: &iotago.ReferenceUnlockBlock{},
		},
		{
			name:   "ok - alias",
			source: tpkg.RandAliasUnlockBlock(),
			target: &iotago.AliasUnlockBlock{},
		},
		{
			name:   "ok - NFT",
			source: tpkg.RandNFTUnlockBlock(),
			target: &iotago.NFTUnlockBlock{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestUnlockBlocksSigUniqueAndRefValidator(t *testing.T) {
	tests := []struct {
		name         string
		unlockBlocks iotago.UnlockBlocks
		wantErr      error
	}{
		{
			name: "ok",
			unlockBlocks: iotago.UnlockBlocks{
				tpkg.RandEd25519SignatureUnlockBlock(),
				tpkg.RandEd25519SignatureUnlockBlock(),
				&iotago.ReferenceUnlockBlock{Reference: 0},
			},
			wantErr: nil,
		},
		{
			name: "ok - chainable referential unlock block",
			unlockBlocks: iotago.UnlockBlocks{
				tpkg.RandEd25519SignatureUnlockBlock(),
				&iotago.AliasUnlockBlock{Reference: 0},
				&iotago.AliasUnlockBlock{Reference: 1},
				&iotago.NFTUnlockBlock{Reference: 2},
			},
			wantErr: nil,
		},
		{
			name: "fail - duplicate ed25519 sig block",
			unlockBlocks: iotago.UnlockBlocks{
				&iotago.SignatureUnlockBlock{Signature: &iotago.Ed25519Signature{
					PublicKey: [32]byte{},
					Signature: [64]byte{},
				}},
				&iotago.SignatureUnlockBlock{Signature: &iotago.Ed25519Signature{
					PublicKey: [32]byte{},
					Signature: [64]byte{},
				}},
			},
			wantErr: iotago.ErrSigUnlockBlocksNotUnique,
		},
		{
			name: "fail - reference unlock block invalid ref",
			unlockBlocks: iotago.UnlockBlocks{
				tpkg.RandEd25519SignatureUnlockBlock(),
				tpkg.RandEd25519SignatureUnlockBlock(),
				&iotago.ReferenceUnlockBlock{Reference: 1337},
			},
			wantErr: iotago.ErrReferentialUnlockBlockInvalid,
		},
		{
			name: "fail - reference unlock block refs non sig unlock block",
			unlockBlocks: iotago.UnlockBlocks{
				tpkg.RandEd25519SignatureUnlockBlock(),
				&iotago.ReferenceUnlockBlock{Reference: 0},
				&iotago.ReferenceUnlockBlock{Reference: 1},
			},
			wantErr: iotago.ErrReferentialUnlockBlockInvalid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valFunc := iotago.UnlockBlocksSigUniqueAndRefValidator()
			var runErr error
			for index, unlockBlock := range tt.unlockBlocks {
				if err := valFunc(index, unlockBlock); err != nil {
					runErr = err
				}
			}
			require.ErrorIs(t, runErr, tt.wantErr)
		})
	}
}
