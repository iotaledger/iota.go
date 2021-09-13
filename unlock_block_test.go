package iotago_test

import (
	"errors"
	"github.com/iotaledger/hive.go/serializer"
	"github.com/iotaledger/iota.go/v2/tpkg"
	"testing"

	"github.com/iotaledger/iota.go/v2"
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
			bytesRead, err := edSig.Deserialize(tt.source, serializer.DeSeriModePerformValidation)
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
			edData, err := tt.source.Serialize(serializer.DeSeriModePerformValidation)
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
			bytesRead, err := edSig.Deserialize(tt.source, serializer.DeSeriModePerformValidation)
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
			edData, err := tt.source.Serialize(serializer.DeSeriModePerformValidation)
			assert.NoError(t, err)
			assert.Equal(t, tt.target, edData)
		})
	}
}

func TestUnlockBlockValidatorFunc(t *testing.T) {
	type args struct {
		inputs []serializer.Serializable
		funcs  []iotago.UnlockBlockValidatorFunc
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"ok",
			args{inputs: []serializer.Serializable{
				func() serializer.Serializable {
					block, _ := tpkg.RandEd25519SignatureUnlockBlock()
					return block
				}(),
				func() serializer.Serializable {
					block, _ := tpkg.RandEd25519SignatureUnlockBlock()
					return block
				}(),
				func() serializer.Serializable {
					return &iotago.ReferenceUnlockBlock{Reference: 0}
				}(),
			}, funcs: []iotago.UnlockBlockValidatorFunc{iotago.UnlockBlocksSigUniqueAndRefValidator()}}, false,
		},
		{
			"duplicate ed25519 sig block",
			args{inputs: []serializer.Serializable{
				func() serializer.Serializable {
					return &iotago.SignatureUnlockBlock{Signature: &iotago.Ed25519Signature{
						PublicKey: [32]byte{},
						Signature: [64]byte{},
					}}
				}(),
				func() serializer.Serializable {
					return &iotago.SignatureUnlockBlock{Signature: &iotago.Ed25519Signature{
						PublicKey: [32]byte{},
						Signature: [64]byte{},
					}}
				}(),
			}, funcs: []iotago.UnlockBlockValidatorFunc{iotago.UnlockBlocksSigUniqueAndRefValidator()}}, true,
		},
		{
			"invalid ref",
			args{inputs: []serializer.Serializable{
				func() serializer.Serializable {
					block, _ := tpkg.RandEd25519SignatureUnlockBlock()
					return block
				}(),
				func() serializer.Serializable {
					block, _ := tpkg.RandEd25519SignatureUnlockBlock()
					return block
				}(),
				func() serializer.Serializable {
					return &iotago.ReferenceUnlockBlock{Reference: 2}
				}(),
			}, funcs: []iotago.UnlockBlockValidatorFunc{iotago.UnlockBlocksSigUniqueAndRefValidator()}}, true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := iotago.ValidateUnlockBlocks(tt.args.inputs, tt.args.funcs...); (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
