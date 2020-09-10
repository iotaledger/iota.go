package iota_test

import (
	"errors"
	"testing"

	"github.com/luca-moser/iota"
	"github.com/stretchr/testify/assert"
)

func TestUnlockBlockSelector(t *testing.T) {
	_, err := iota.UnlockBlockSelector(100)
	assert.True(t, errors.Is(err, iota.ErrUnknownUnlockBlockType))
}

func TestSignatureUnlockBlock_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target iota.Serializable
		err    error
	}
	tests := []test{
		func() test {
			edSigBlock, edSigBlockData := randEd25519SignatureUnlockBlock()
			return test{"ok", edSigBlockData, edSigBlock, nil}
		}(),
		func() test {
			edSigBlock, edSigBlockData := randEd25519SignatureUnlockBlock()
			return test{"not enough data", edSigBlockData[:5], edSigBlock, iota.ErrDeserializationNotEnoughData}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edSig := &iota.SignatureUnlockBlock{}
			bytesRead, err := edSig.Deserialize(tt.source, iota.DeSeriModePerformValidation)
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
		source *iota.SignatureUnlockBlock
		target []byte
	}
	tests := []test{
		func() test {
			edSigBlock, edSigBlockData := randEd25519SignatureUnlockBlock()
			return test{"ok", edSigBlock, edSigBlockData}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edData, err := tt.source.Serialize(iota.DeSeriModePerformValidation)
			assert.NoError(t, err)
			assert.Equal(t, tt.target, edData)
		})
	}
}

func TestReferenceUnlockBlock_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target iota.Serializable
		err    error
	}
	tests := []test{
		func() test {
			refBlock, refBlockData := randReferenceUnlockBlock()
			return test{"ok", refBlockData, refBlock, nil}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edSig := &iota.ReferenceUnlockBlock{}
			bytesRead, err := edSig.Deserialize(tt.source, iota.DeSeriModePerformValidation)
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
		source *iota.ReferenceUnlockBlock
		target []byte
	}
	tests := []test{
		func() test {
			refBlock, refBlockData := randReferenceUnlockBlock()
			return test{"ok", refBlock, refBlockData}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edData, err := tt.source.Serialize(iota.DeSeriModePerformValidation)
			assert.NoError(t, err)
			assert.Equal(t, tt.target, edData)
		})
	}
}

func TestUnlockBlockValidatorFunc(t *testing.T) {
	type args struct {
		inputs []iota.Serializable
		funcs  []iota.UnlockBlockValidatorFunc
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"ok",
			args{inputs: []iota.Serializable{
				func() iota.Serializable {
					block, _ := randEd25519SignatureUnlockBlock()
					return block
				}(),
				func() iota.Serializable {
					block, _ := randEd25519SignatureUnlockBlock()
					return block
				}(),
				func() iota.Serializable {
					return &iota.ReferenceUnlockBlock{Reference: 0}
				}(),
			}, funcs: []iota.UnlockBlockValidatorFunc{iota.UnlockBlocksSigUniqueAndRefValidator()}}, false,
		},
		{
			"duplicate ed25519 sig block",
			args{inputs: []iota.Serializable{
				func() iota.Serializable {
					return &iota.SignatureUnlockBlock{Signature: &iota.Ed25519Signature{
						PublicKey: [32]byte{},
						Signature: [64]byte{},
					}}
				}(),
				func() iota.Serializable {
					return &iota.SignatureUnlockBlock{Signature: &iota.Ed25519Signature{
						PublicKey: [32]byte{},
						Signature: [64]byte{},
					}}
				}(),
			}, funcs: []iota.UnlockBlockValidatorFunc{iota.UnlockBlocksSigUniqueAndRefValidator()}}, true,
		},
		{
			"invalid ref",
			args{inputs: []iota.Serializable{
				func() iota.Serializable {
					block, _ := randEd25519SignatureUnlockBlock()
					return block
				}(),
				func() iota.Serializable {
					block, _ := randEd25519SignatureUnlockBlock()
					return block
				}(),
				func() iota.Serializable {
					return &iota.ReferenceUnlockBlock{Reference: 2}
				}(),
			}, funcs: []iota.UnlockBlockValidatorFunc{iota.UnlockBlocksSigUniqueAndRefValidator()}}, true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := iota.ValidateUnlockBlocks(tt.args.inputs, tt.args.funcs...); (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
