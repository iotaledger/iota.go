package iotago_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/iotaledger/hive.go/serializer"
	"github.com/iotaledger/iota.go/v2"
	"github.com/iotaledger/iota.go/v2/tpkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignatureSelector(t *testing.T) {
	_, err := iotago.SignatureSelector(100)
	assert.True(t, errors.Is(err, iotago.ErrUnknownSignatureType))
}

func TestEd25519Signature_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target serializer.Serializable
		err    error
	}
	tests := []test{
		func() test {
			edSig, edSigData := tpkg.RandEd25519Signature()
			return test{"ok", edSigData, edSig, nil}
		}(),
		func() test {
			edSig, edSigData := tpkg.RandEd25519Signature()
			return test{"not enough data", edSigData[:5], edSig, serializer.ErrDeserializationNotEnoughData}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edSig := &iotago.Ed25519Signature{}
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

func TestEd25519Signature_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iotago.Ed25519Signature
		target []byte
	}
	tests := []test{
		func() test {
			edSig, edSigData := tpkg.RandEd25519Signature()
			return test{"ok", edSig, edSigData}
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

func TestEd25519Signature_Valid(t *testing.T) {
	type test struct {
		Address   tpkg.HexBytes `json:"address"`
		Message   tpkg.HexBytes `json:"message"`
		PublicKey tpkg.HexBytes `json:"pub_key"`
		Signature tpkg.HexBytes `json:"signature"`
		Valid     bool          `json:"valid"`
	}
	var tests []test
	// load the tests from file
	b, err := ioutil.ReadFile(filepath.Join("testdata", t.Name()+".json"))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(b, &tests))

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			// deserialize the address from the test
			addr := &iotago.Ed25519Address{}
			_, err = addr.Deserialize(tt.Address, serializer.DeSeriModePerformValidation)
			require.NoError(t, err)
			// create the signature type
			sig := &iotago.Ed25519Signature{}
			copy(sig.PublicKey[:], tt.PublicKey)
			copy(sig.Signature[:], tt.Signature)

			sigError := sig.Valid(tt.Message, addr)
			switch tt.Valid {
			case true:
				assert.NoError(t, sigError)
			case false:
				assert.Error(t, sigError)
			}
		})
	}
}
