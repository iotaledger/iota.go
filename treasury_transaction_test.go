package iotago_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/tpkg"

	"github.com/iotaledger/iota.go/v3"
	"github.com/stretchr/testify/assert"
)

func TestTreasuryTransaction_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target *iotago.TreasuryTransaction
		err    error
	}
	tests := []test{
		func() test {
			tx, txData := tpkg.RandTreasuryTransaction()
			return test{"ok- w/o tx", txData, tx, nil}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &iotago.TreasuryTransaction{}
			bytesRead, err := tx.Deserialize(tt.source, serializer.DeSeriModePerformValidation, DefZeroRentParas)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, len(tt.source), bytesRead)
			assert.EqualValues(t, tt.target, tx)
		})
	}
}

func TestTreasuryTransaction_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iotago.TreasuryTransaction
		target []byte
	}
	tests := []test{
		func() test {
			tx, txData := tpkg.RandTreasuryTransaction()
			return test{"ok- w/o tx", tx, txData}
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
