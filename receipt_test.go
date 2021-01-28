package iota_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/iota.go"
	"github.com/stretchr/testify/assert"
)

func TestReceipt_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target *iota.Receipt
		err    error
	}
	tests := []test{
		func() test {
			receipt, receiptData := randReceipt(false)
			return test{"ok- w/o tx", receiptData, receipt, nil}
		}(),
		func() test {
			receipt, receiptData := randReceipt(true)
			return test{"ok - w/ tx", receiptData, receipt, nil}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receipt := &iota.Receipt{}
			bytesRead, err := receipt.Deserialize(tt.source, iota.DeSeriModePerformValidation)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, len(tt.source), bytesRead)
			assert.EqualValues(t, tt.target, receipt)
		})
	}
}

func TestReceipt_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iota.Receipt
		target []byte
	}
	tests := []test{
		func() test {
			receipt, receiptData := randReceipt(false)
			return test{"ok- w/o tx", receipt, receiptData}
		}(),
		func() test {
			receipt, receiptData := randReceipt(true)
			return test{"ok - w/ tx", receipt, receiptData}
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
