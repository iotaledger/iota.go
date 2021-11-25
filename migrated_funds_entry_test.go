package iotago_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/tpkg"

	"github.com/iotaledger/iota.go/v3"
	"github.com/stretchr/testify/assert"
)

func TestMigratedFundsEntry_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target *iotago.MigratedFundsEntry
		err    error
	}
	tests := []test{
		func() test {
			migFundsEntry, migFundsEntryData := tpkg.RandMigratedFundsEntry()
			return test{"ok- w/o migFundsEntry", migFundsEntryData, migFundsEntry, nil}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migFundsEntry := &iotago.MigratedFundsEntry{}
			bytesRead, err := migFundsEntry.Deserialize(tt.source, serializer.DeSeriModePerformValidation, DefZeroRentParas)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, len(tt.source), bytesRead)
			assert.EqualValues(t, tt.target, migFundsEntry)
		})
	}
}

func TestMigratedFundsEntry_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iotago.MigratedFundsEntry
		target []byte
	}
	tests := []test{
		func() test {
			migFundsEntry, migFundsEntryData := tpkg.RandMigratedFundsEntry()
			return test{"ok- w/o migFundsEntry", migFundsEntry, migFundsEntryData}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edData, err := tt.source.Serialize(serializer.DeSeriModePerformValidation, DefZeroRentParas)
			assert.NoError(t, err)
			assert.Equal(t, tt.target, edData)
		})
	}
}
