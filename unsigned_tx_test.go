package iota_test

import (
	"errors"
	"testing"

	"github.com/luca-moser/iota"
	"github.com/stretchr/testify/assert"
)

func TestTransactionSelector(t *testing.T) {
	_, err := iota.TransactionSelector(100)
	assert.True(t, errors.Is(err, iota.ErrUnknownTransactionType))
}

func TestUnsignedTransaction_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target iota.Serializable
		err    error
	}
	tests := []test{
		func() test {
			unTx, unTxData := randUnsignedTransaction()
			return test{"ok", unTxData, unTx, nil}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &iota.UnsignedTransaction{}
			bytesRead, err := tx.Deserialize(tt.source, iota.DeSeriModePerformValidation)
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

func TestUnsignedTransaction_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iota.UnsignedTransaction
		target []byte
	}
	tests := []test{
		func() test {
			unTx, unTxData := randUnsignedTransaction()
			return test{"ok", unTx, unTxData}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edData, err := tt.source.Serialize(iota.DeSeriModePerformValidation)
			assert.NoError(t, err)
			assert.Equal(t, tt.target, edData)
		})
	}
}
