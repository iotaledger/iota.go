package iotago_test

import (
	"errors"
	"github.com/iotaledger/iota.go/v2/test"
	"testing"

	"github.com/iotaledger/iota.go/v2"
	"github.com/stretchr/testify/assert"
)

func TestUTXOInput_Deserialize(t *testing.T) {
	randUTXOInput, randSerializedUTXOInput := test.RandUTXOInput()
	tests := []struct {
		name   string
		data   []byte
		target *iotago.UTXOInput
		err    error
	}{
		{"ok", randSerializedUTXOInput, randUTXOInput, nil},
		{"not enough data", randSerializedUTXOInput[:iotago.UTXOInputSize-1], randUTXOInput, iotago.ErrDeserializationNotEnoughData},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &iotago.UTXOInput{}
			bytesRead, err := u.Deserialize(tt.data, iotago.DeSeriModePerformValidation)
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
	randUTXOInput, randSerializedUTXOInput := test.RandUTXOInput()
	tests := []struct {
		name   string
		source *iotago.UTXOInput
		target []byte
		err    error
	}{
		{"ok", randUTXOInput, randSerializedUTXOInput, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.source.Serialize(iotago.DeSeriModePerformValidation)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.EqualValues(t, tt.target, data)
		})
	}
}
