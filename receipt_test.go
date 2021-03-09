package iotago_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/iota.go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReceipt_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target *iotago.Receipt
		err    error
	}
	tests := []test{
		func() test {
			receipt, receiptData := randReceipt()
			return test{"ok", receiptData, receipt, nil}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receipt := &iotago.Receipt{}
			bytesRead, err := receipt.Deserialize(tt.source, iotago.DeSeriModePerformValidation)
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
		source *iotago.Receipt
		target []byte
	}
	tests := []test{
		func() test {
			receipt, receiptData := randReceipt()
			return test{"ok", receipt, receiptData}
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

func TestReceiptFuzzingCrashers(t *testing.T) {
	type test struct {
		in []byte
	}
	tests := []test{
		{
			in: []byte("000"),
		},
		{
			in: []byte("00"),
		},
		{
			in: []byte("0"),
		},
		{
			in: []byte(""),
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.in), func(t *testing.T) {
			m := &iotago.Receipt{}
			_, err := m.Deserialize(tt.in, iotago.DeSeriModePerformValidation)
			if err != nil {
				return
			}

			seriData, err := m.Serialize(iotago.DeSeriModePerformValidation)
			if err != nil {
				return
			}

			require.EqualValues(t, tt.in[:len(seriData)], seriData)
		})
	}
}

func TestValidateReceipts(t *testing.T) {
	type test struct {
		name      string
		source    *iotago.Receipt
		prevInput *iotago.TreasuryOutput
		err       error
	}
	currentTreasury := &iotago.TreasuryOutput{Amount: 10_000_000}
	inputID := rand32ByteHash()
	sampleTreasuryTx := &iotago.TreasuryTransaction{Output: &iotago.TreasuryOutput{Amount: 3_000_000}}
	treasuryInput := &iotago.TreasuryInput{}
	copy(treasuryInput[:], inputID[:])
	sampleTreasuryTx.Input = treasuryInput

	tests := []test{
		func() test {
			addr, _ := randEd25519Addr()
			receipt, _ := iotago.NewReceiptBuilder(100).AddEntry(&iotago.MigratedFundsEntry{
				TailTransactionHash: rand49ByteHash(),
				Address:             addr,
				Deposit:             7_000_000,
			}).AddTreasuryTransaction(sampleTreasuryTx).Build()
			return test{"ok", receipt, currentTreasury, nil}
		}(),
		func() test {
			addr, _ := randEd25519Addr()
			receipt, _ := iotago.NewReceiptBuilder(100).AddEntry(&iotago.MigratedFundsEntry{
				TailTransactionHash: rand49ByteHash(),
				Address:             addr,
				Deposit:             1000,
			}).AddTreasuryTransaction(sampleTreasuryTx).Build()
			return test{"err - migrated less tha minimum", receipt, currentTreasury, iotago.ErrInvalidReceipt}
		}(),
		func() test {
			addr, _ := randEd25519Addr()
			receipt, _ := iotago.NewReceiptBuilder(100).AddEntry(&iotago.MigratedFundsEntry{
				TailTransactionHash: rand49ByteHash(),
				Address:             addr,
				Deposit:             iotago.TokenSupply + 1,
			}).AddTreasuryTransaction(sampleTreasuryTx).Build()
			return test{"err - total supply overflow", receipt, currentTreasury, iotago.ErrInvalidReceipt}
		}(),
		func() test {
			addr, _ := randEd25519Addr()
			receipt, _ := iotago.NewReceiptBuilder(100).AddEntry(&iotago.MigratedFundsEntry{
				TailTransactionHash: rand49ByteHash(),
				Address:             addr,
				Deposit:             6_000_000,
			}).AddTreasuryTransaction(sampleTreasuryTx).Build()
			return test{"err - invalid new treasury amount", receipt, currentTreasury, iotago.ErrInvalidReceipt}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := iotago.ValidateReceipt(tt.source, tt.prevInput)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
		})
	}
}
