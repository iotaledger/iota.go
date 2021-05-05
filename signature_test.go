package iotago_test

import (
	"errors"
	test2 "github.com/iotaledger/iota.go/v2/test"
	"testing"

	"github.com/iotaledger/iota.go/v2"
	"github.com/stretchr/testify/assert"
)

func TestSignatureSelector(t *testing.T) {
	_, err := iotago.SignatureSelector(100)
	assert.True(t, errors.Is(err, iotago.ErrUnknownSignatureType))
}

func TestEd25519Signature_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target iotago.Serializable
		err    error
	}
	tests := []test{
		func() test {
			edSig, edSigData := test2.RandEd25519Signature()
			return test{"ok", edSigData, edSig, nil}
		}(),
		func() test {
			edSig, edSigData := test2.RandEd25519Signature()
			return test{"not enough data", edSigData[:5], edSig, iotago.ErrDeserializationNotEnoughData}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edSig := &iotago.Ed25519Signature{}
			bytesRead, err := edSig.Deserialize(tt.source, iotago.DeSeriModePerformValidation)
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

func TestEd25519Signature_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iotago.Ed25519Signature
		target []byte
	}
	tests := []test{
		func() test {
			edSig, edSigData := test2.RandEd25519Signature()
			return test{"ok", edSig, edSigData}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edData, err := tt.source.Serialize(iotago.DeSeriModePerformValidation)
			assert.NoError(t, err)
			assert.Equal(t, tt.target, edData)
		})
	}
}
