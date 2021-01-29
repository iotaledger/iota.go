package iota_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/iota.go/v2"
	"github.com/stretchr/testify/assert"
)

func TestUTXOInput_Deserialize(t *testing.T) {
	randUTXOInput, randSerializedUTXOInput := randUTXOInput()
	tests := []struct {
		name   string
		data   []byte
		target *iota.UTXOInput
		err    error
	}{
		{"ok", randSerializedUTXOInput, randUTXOInput, nil},
		{"not enough data", randSerializedUTXOInput[:iota.UTXOInputSize-1], randUTXOInput, iota.ErrDeserializationNotEnoughData},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &iota.UTXOInput{}
			bytesRead, err := u.Deserialize(tt.data, iota.DeSeriModePerformValidation)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.Equal(t, len(tt.data), bytesRead)
			assert.EqualValues(t, tt.target, u)
		})
	}
}

func TestUTXOInput_Serialize(t *testing.T) {
	randUTXOInput, randSerializedUTXOInput := randUTXOInput()
	tests := []struct {
		name   string
		source *iota.UTXOInput
		target []byte
		err    error
	}{
		{"ok", randUTXOInput, randSerializedUTXOInput, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.source.Serialize(iota.DeSeriModePerformValidation)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.EqualValues(t, tt.target, data)
		})
	}
}
