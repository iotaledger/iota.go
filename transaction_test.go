package iotago_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target serializer.Serializable
		err    error
	}
	tests := []test{
		func() test {
			txPayload, txPayloadData := tpkg.RandTransaction()
			return test{"ok", txPayloadData, txPayload, nil}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &iotago.Transaction{}
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

func TestTransaction_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iotago.Transaction
		target []byte
	}
	tests := []test{
		func() test {
			txPayload, txPayloadData := tpkg.RandTransaction()
			return test{"ok", txPayload, txPayloadData}
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

func TestTxSemanticInputUnlocks(t *testing.T) {
	tests := []struct {
		name    string
		inputs  iotago.OutputSet
		tx      *iotago.Transaction
		wantErr error
	}{
		{
			name:   "ok",
			inputs: iotago.OutputSet{},
			tx:     &iotago.Transaction{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			/*
				valFunc := iotago.TxSemanticInputUnlocks()
				var runErr error
				for index, inputs := range tt.inputs {
					if err := valFunc(); err != nil {
						runErr = err
					}
				}
				require.ErrorIs(t, runErr, tt.wantErr)
			*/
		})
	}
}
