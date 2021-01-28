package iota_test

import (
	"errors"
	"fmt"
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
			receipt, receiptData := randReceipt()
			return test{"ok", receiptData, receipt, nil}
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
			receipt, receiptData := randReceipt()
			return test{"ok", receipt, receiptData}
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

func TestValidateReceipts(t *testing.T) {
	type test struct {
		name      string
		source    *iota.Receipt
		prevInput *iota.TreasuryOutput
		err       error
	}
	currentTreasury := &iota.TreasuryOutput{Amount: 10_000_000}
	inputID := rand32ByteHash()
	sampleTreasuryTx := &iota.TreasuryTransaction{Output: &iota.TreasuryOutput{Amount: 3_000_000}}
	treasuryInput := &iota.TreasuryInput{}
	copy(treasuryInput[:], inputID[:])
	sampleTreasuryTx.Input = treasuryInput

	tests := []test{
		func() test {
			addr, _ := randEd25519Addr()
			receipt, _ := iota.NewReceiptBuilder(100).AddEntry(&iota.MigratedFundsEntry{
				TailTransactionHash: rand49ByteHash(),
				Address:             addr,
				Deposit:             7_000_000,
			}).AddTreasuryTransaction(sampleTreasuryTx).Build()
			return test{"ok", receipt, currentTreasury, nil}
		}(),
		func() test {
			addr, _ := randEd25519Addr()
			receipt, _ := iota.NewReceiptBuilder(100).AddEntry(&iota.MigratedFundsEntry{
				TailTransactionHash: rand49ByteHash(),
				Address:             addr,
				Deposit:             1000,
			}).AddTreasuryTransaction(sampleTreasuryTx).Build()
			return test{"err - migrated less tha minimum", receipt, currentTreasury, iota.ErrInvalidReceiptsSet}
		}(),
		func() test {
			addr, _ := randEd25519Addr()
			receipt, _ := iota.NewReceiptBuilder(100).AddEntry(&iota.MigratedFundsEntry{
				TailTransactionHash: rand49ByteHash(),
				Address:             addr,
				Deposit:             iota.TokenSupply + 1,
			}).AddTreasuryTransaction(sampleTreasuryTx).Build()
			return test{"err - total supply overflow", receipt, currentTreasury, iota.ErrInvalidReceiptsSet}
		}(),
		func() test {
			addr, _ := randEd25519Addr()
			receipt, _ := iota.NewReceiptBuilder(100).AddEntry(&iota.MigratedFundsEntry{
				TailTransactionHash: rand49ByteHash(),
				Address:             addr,
				Deposit:             6_000_000,
			}).AddTreasuryTransaction(sampleTreasuryTx).Build()
			return test{"err - invalid new treasury amount", receipt, currentTreasury, iota.ErrInvalidReceiptsSet}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := iota.ValidateReceipt(tt.source, tt.prevInput)
			fmt.Println(err)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
		})
	}
}
