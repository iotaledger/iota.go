package iotago_test

import (
	"errors"
	"github.com/iotaledger/hive.go/serializer"
	"github.com/iotaledger/iota.go/v2/tpkg"
	"testing"

	"github.com/iotaledger/iota.go/v2"
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
			bytesRead, err := migFundsEntry.Deserialize(tt.source, serializer.DeSeriModePerformValidation)
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
			edData, err := tt.source.Serialize(serializer.DeSeriModePerformValidation)
			assert.NoError(t, err)
			assert.Equal(t, tt.target, edData)
		})
	}
}
