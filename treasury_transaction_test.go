package iotago_test

import (
	"errors"
	"github.com/iotaledger/hive.go/serializer"
	"github.com/iotaledger/iota.go/v2/tpkg"
	"testing"

	"github.com/iotaledger/iota.go/v2"
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
			bytesRead, err := tx.Deserialize(tt.source, serializer.DeSeriModePerformValidation)
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
			edData, err := tt.source.Serialize(serializer.DeSeriModePerformValidation)
			assert.NoError(t, err)
			assert.Equal(t, tt.target, edData)
		})
	}
}
