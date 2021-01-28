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

func TestValidateReceipts(t *testing.T) {
	type test struct {
		name      string
		source    []*iota.Receipt
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
			addr1, _ := randEd25519Addr()
			addr2, _ := randEd25519Addr()
			receipt1, _ := iota.NewReceiptBuilder(100).AddEntry(&iota.MigratedFundsEntry{
				TailTransactionHash: rand49ByteHash(),
				Address:             addr1,
				Deposit:             1_000_000,
			}).Build()
			receipt2, _ := iota.NewReceiptBuilder(100).AddEntry(&iota.MigratedFundsEntry{
				TailTransactionHash: rand49ByteHash(),
				Address:             addr2,
				Deposit:             6_000_000,
			}).AddTreasuryTransaction(sampleTreasuryTx).Build()
			return test{
				"ok",
				[]*iota.Receipt{receipt1, receipt2},
				currentTreasury,
				nil,
			}
		}(),
		func() test {
			addr1, _ := randEd25519Addr()
			addr2, _ := randEd25519Addr()
			receipt1, _ := iota.NewReceiptBuilder(87234).AddEntry(&iota.MigratedFundsEntry{
				TailTransactionHash: rand49ByteHash(),
				Address:             addr1,
				Deposit:             1_000_000,
			}).Build()
			receipt2, _ := iota.NewReceiptBuilder(34875).AddEntry(&iota.MigratedFundsEntry{
				TailTransactionHash: rand49ByteHash(),
				Address:             addr2,
				Deposit:             6_000_000,
			}).AddTreasuryTransaction(sampleTreasuryTx).Build()
			return test{
				"err - multiple migrated at indices",
				[]*iota.Receipt{receipt1, receipt2},
				currentTreasury,
				iota.ErrInvalidReceiptsSet,
			}
		}(),
		func() test {
			addr1, _ := randEd25519Addr()
			receipt1, _ := iota.NewReceiptBuilder(100).AddEntry(&iota.MigratedFundsEntry{
				TailTransactionHash: rand49ByteHash(),
				Address:             addr1,
				Deposit:             1000,
			}).AddTreasuryTransaction(sampleTreasuryTx).Build()
			return test{
				"err - migrated less tha minimum",
				[]*iota.Receipt{receipt1},
				currentTreasury,
				iota.ErrInvalidReceiptsSet,
			}
		}(),
		func() test {
			addr1, _ := randEd25519Addr()
			receipt1, _ := iota.NewReceiptBuilder(100).AddEntry(&iota.MigratedFundsEntry{
				TailTransactionHash: rand49ByteHash(),
				Address:             addr1,
				Deposit:             iota.TokenSupply + 1,
			}).AddTreasuryTransaction(sampleTreasuryTx).Build()
			return test{
				"err - total supply overflow",
				[]*iota.Receipt{receipt1},
				currentTreasury,
				iota.ErrInvalidReceiptsSet,
			}
		}(),
		func() test {
			addr1, _ := randEd25519Addr()
			receipt1, _ := iota.NewReceiptBuilder(100).AddEntry(&iota.MigratedFundsEntry{
				TailTransactionHash: rand49ByteHash(),
				Address:             addr1,
				Deposit:             6_000_000,
			}).AddTreasuryTransaction(sampleTreasuryTx).Build()
			return test{
				"err - invalid new treasury amount",
				[]*iota.Receipt{receipt1},
				currentTreasury,
				iota.ErrInvalidReceiptsSet,
			}
		}(),
		func() test {
			addr1, _ := randEd25519Addr()
			addr2, _ := randEd25519Addr()
			// both receipts contain the same tail tx hash
			tailHash := rand49ByteHash()
			receipt1, _ := iota.NewReceiptBuilder(100).AddEntry(&iota.MigratedFundsEntry{
				TailTransactionHash: tailHash,
				Address:             addr1,
				Deposit:             1_000_000,
			}).Build()
			receipt2, _ := iota.NewReceiptBuilder(100).AddEntry(&iota.MigratedFundsEntry{
				TailTransactionHash: tailHash,
				Address:             addr2,
				Deposit:             6_000_000,
			}).AddTreasuryTransaction(sampleTreasuryTx).Build()
			return test{
				"err - duplicate tail transactions",
				[]*iota.Receipt{receipt1, receipt2},
				currentTreasury,
				iota.ErrInvalidReceiptsSet,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := iota.ValidateReceipts(tt.source, tt.prevInput)
			fmt.Println(err)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
		})
	}
}
