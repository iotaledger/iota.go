package iotago_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/iota.go/v2"
	"github.com/stretchr/testify/assert"
)

func TestTransactionEssenceSelector(t *testing.T) {
	_, err := iotago.TransactionEssenceSelector(100)
	assert.True(t, errors.Is(err, iotago.ErrUnknownTransactionEssenceType))
}

func TestTransactionEssence_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target iotago.Serializable
		err    error
	}
	tests := []test{
		func() test {
			unTx, unTxData := randTransactionEssence()
			return test{"ok", unTxData, unTx, nil}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &iotago.TransactionEssence{}
			bytesRead, err := tx.Deserialize(tt.source, iotago.DeSeriModePerformValidation)
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

func TestTransactionEssence_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iotago.TransactionEssence
		target []byte
	}
	tests := []test{
		func() test {
			unTx, unTxData := randTransactionEssence()
			return test{"ok", unTx, unTxData}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edData, err := tt.source.Serialize(iotago.DeSeriModePerformValidation)
			assert.NoError(t, err)
			assert.Equal(t, tt.target, edData)
		})
	}
}
