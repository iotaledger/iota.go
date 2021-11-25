package iotago_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/hive.go/v2/serializer"
	"github.com/iotaledger/iota.go/v3/tpkg"

	"github.com/iotaledger/iota.go/v3"
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
			bytesRead, err := u.Deserialize(tt.data, serializer.DeSeriModePerformValidation, DefZeroRentParas)
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
			data, err := tt.source.Serialize(serializer.DeSeriModePerformValidation, DefZeroRentParas)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.EqualValues(t, tt.target, data)
		})
	}
}
