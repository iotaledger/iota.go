//nolint:scopelint
package iotago_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
	"github.com/iotaledger/iota.go/v4/tpkg/frameworks"
)

func TestEd25519Signature_DeSerialize(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name:   "ok",
			Source: tpkg.RandEd25519Signature(),
			Target: &iotago.Ed25519Signature{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

func TestEd25519Signature_Valid(t *testing.T) {
	type test struct {
		Address   tpkg.HexBytes `json:"address"`
		Message   tpkg.HexBytes `json:"message"`
		PublicKey tpkg.HexBytes `json:"pubKey"`
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
			_, err = tpkg.ZeroCostTestAPI.Decode(tt.Address, addr)
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
