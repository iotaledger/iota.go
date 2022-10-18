package iotago_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

func TestEd25519Signature_DeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok",
			source: tpkg.RandEd25519Signature(),
			target: &iotago.Ed25519Signature{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
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
	b, err := os.ReadFile(filepath.Join("testdata", t.Name()+".json"))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(b, &tests))

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			// deserialize the address from the test
			addr := &iotago.Ed25519Address{}
			_, err = v2API.Decode(tt.Address, addr)
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
