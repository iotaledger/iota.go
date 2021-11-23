package iotago_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/hive.go/serializer"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/iota.go/v3"
	"github.com/stretchr/testify/assert"
)

func TestUnlockBlockSelector(t *testing.T) {
	_, err := iotago.UnlockBlockSelector(100)
	assert.True(t, errors.Is(err, iotago.ErrUnknownUnlockBlockType))
}

func TestSignatureUnlockBlock_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target serializer.Serializable
		err    error
	}
	tests := []test{
		func() test {
			edSigBlock, edSigBlockData := tpkg.RandEd25519SignatureUnlockBlock()
			return test{"ok", edSigBlockData, edSigBlock, nil}
		}(),
		func() test {
			edSigBlock, edSigBlockData := tpkg.RandEd25519SignatureUnlockBlock()
			return test{"not enough data", edSigBlockData[:5], edSigBlock, serializer.ErrDeserializationNotEnoughData}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edSig := &iotago.SignatureUnlockBlock{}
			bytesRead, err := edSig.Deserialize(tt.source, serializer.DeSeriModePerformValidation, DefZeroRentParas)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, len(tt.source), bytesRead)
			assert.EqualValues(t, tt.target, edSig)
		})
	}
}

func TestUnlockBlockSignature_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iotago.SignatureUnlockBlock
		target []byte
	}
	tests := []test{
		func() test {
			edSigBlock, edSigBlockData := tpkg.RandEd25519SignatureUnlockBlock()
			return test{"ok", edSigBlock, edSigBlockData}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edData, err := tt.source.Serialize(serializer.DeSeriModePerformValidation, DefZeroRentParas)
			assert.NoError(t, err)
			assert.Equal(t, tt.target, edData)
		})
	}
}

func TestReferenceUnlockBlock_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target serializer.Serializable
		err    error
	}
	tests := []test{
		func() test {
			refBlock, refBlockData := tpkg.RandReferenceUnlockBlock()
			return test{"ok", refBlockData, refBlock, nil}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edSig := &iotago.ReferenceUnlockBlock{}
			bytesRead, err := edSig.Deserialize(tt.source, serializer.DeSeriModePerformValidation, DefZeroRentParas)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, len(tt.source), bytesRead)
			assert.EqualValues(t, tt.target, edSig)
		})
	}
}

func TestUnlockBlockReference_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iotago.ReferenceUnlockBlock
		target []byte
	}
	tests := []test{
		func() test {
			refBlock, refBlockData := tpkg.RandReferenceUnlockBlock()
			return test{"ok", refBlock, refBlockData}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edData, err := tt.source.Serialize(serializer.DeSeriModePerformValidation, DefZeroRentParas)
			assert.NoError(t, err)
			assert.Equal(t, tt.target, edData)
		})
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
				func() iotago.UnlockBlock {
					block, _ := tpkg.RandEd25519SignatureUnlockBlock()
					return block
				}(),
				func() iotago.UnlockBlock {
					block, _ := tpkg.RandEd25519SignatureUnlockBlock()
					return block
				}(),
				func() iotago.UnlockBlock {
					return &iotago.ReferenceUnlockBlock{Reference: 0}
				}(),
			},
			wantErr: nil,
		},
		{
			name: "ok - chainable referential unlock block",
			unlockBlocks: iotago.UnlockBlocks{
				func() iotago.UnlockBlock {
					block, _ := tpkg.RandEd25519SignatureUnlockBlock()
					return block
				}(),
				func() iotago.UnlockBlock {
					return &iotago.AliasUnlockBlock{Reference: 0}
				}(),
				func() iotago.UnlockBlock {
					return &iotago.AliasUnlockBlock{Reference: 1}
				}(),
				func() iotago.UnlockBlock {
					return &iotago.NFTUnlockBlock{Reference: 2}
				}(),
			},
			wantErr: nil,
		},
		{
			name: "fail - duplicate ed25519 sig block",
			unlockBlocks: iotago.UnlockBlocks{
				func() iotago.UnlockBlock {
					return &iotago.SignatureUnlockBlock{Signature: &iotago.Ed25519Signature{
						PublicKey: [32]byte{},
						Signature: [64]byte{},
					}}
				}(),
				func() iotago.UnlockBlock {
					return &iotago.SignatureUnlockBlock{Signature: &iotago.Ed25519Signature{
						PublicKey: [32]byte{},
						Signature: [64]byte{},
					}}
				}(),
			},
			wantErr: iotago.ErrSigUnlockBlocksNotUnique,
		},
		{
			name: "fail - reference unlock block invalid ref",
			unlockBlocks: iotago.UnlockBlocks{
				func() iotago.UnlockBlock {
					block, _ := tpkg.RandEd25519SignatureUnlockBlock()
					return block
				}(),
				func() iotago.UnlockBlock {
					block, _ := tpkg.RandEd25519SignatureUnlockBlock()
					return block
				}(),
				func() iotago.UnlockBlock {
					return &iotago.ReferenceUnlockBlock{Reference: 1337}
				}(),
			},
			wantErr: iotago.ErrReferentialUnlockBlockInvalid,
		},
		{
			name: "fail - reference unlock block refs non sig unlock block",
			unlockBlocks: iotago.UnlockBlocks{
				func() iotago.UnlockBlock {
					block, _ := tpkg.RandEd25519SignatureUnlockBlock()
					return block
				}(),
				func() iotago.UnlockBlock {
					return &iotago.ReferenceUnlockBlock{Reference: 0}
				}(),
				func() iotago.UnlockBlock {
					return &iotago.ReferenceUnlockBlock{Reference: 1}
				}(),
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
