package iotago_test

import (
	"errors"
	"github.com/iotaledger/hive.go/serializer"
	"github.com/iotaledger/iota.go/v2/tpkg"
	"testing"

	"github.com/iotaledger/iota.go/v2"
	"github.com/stretchr/testify/assert"
)

func TestTreasuryInput_Deserialize(t *testing.T) {
	randTreasuryInput, randTreasuryInputData := tpkg.RandTreasuryInput()
	tests := []struct {
		name   string
		data   []byte
		target *iotago.TreasuryInput
		err    error
	}{
		{"ok", randTreasuryInputData, randTreasuryInput, nil},
		{"not enough data", randTreasuryInputData[:iotago.TreasuryInputSerializedBytesSize-1], randTreasuryInput, serializer.ErrDeserializationNotEnoughData},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &iotago.TreasuryInput{}
			bytesRead, err := u.Deserialize(tt.data, serializer.DeSeriModePerformValidation)
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
	randTreasuryInput, randTreasuryInputData := tpkg.RandTreasuryInput()
	tests := []struct {
		name   string
		source *iotago.TreasuryInput
		target []byte
		err    error
	}{
		{"ok", randTreasuryInput, randTreasuryInputData, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.source.Serialize(serializer.DeSeriModePerformValidation)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.EqualValues(t, tt.target, data)
		})
	}
}
