package iota_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/iota.go/v2"
	"github.com/stretchr/testify/assert"
)

func TestTreasuryInput_Deserialize(t *testing.T) {
	randTreasuryInput, randTreasuryInputData := randTreasuryInput()
	tests := []struct {
		name   string
		data   []byte
		target *iota.TreasuryInput
		err    error
	}{
		{"ok", randTreasuryInputData, randTreasuryInput, nil},
		{"not enough data", randTreasuryInputData[:iota.TreasuryInputSerializedBytesSize-1], randTreasuryInput, iota.ErrDeserializationNotEnoughData},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &iota.TreasuryInput{}
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

func TestTreasuryInput_Serialize(t *testing.T) {
	randTreasuryInput, randTreasuryInputData := randTreasuryInput()
	tests := []struct {
		name   string
		source *iota.TreasuryInput
		target []byte
		err    error
	}{
		{"ok", randTreasuryInput, randTreasuryInputData, nil},
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
