package iotago_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/hive.go/core/serix"

	"github.com/iotaledger/iota.go/v3/tpkg"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
)

func TestReceiptMilestoneOpt_DeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok - receipt milestone option",
			source: tpkg.RandReceipt(),
			target: &iotago.ReceiptMilestoneOpt{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
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
			m := &iotago.ReceiptMilestoneOpt{}
			_, err := v2API.Decode(tt.in, m, serix.WithValidation())
			if err != nil {
				return
			}

			seriData, err := v2API.Encode(m, serix.WithValidation())
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
		source    *iotago.ReceiptMilestoneOpt
		prevInput *iotago.TreasuryOutput
		err       error
	}
	currentTreasury := &iotago.TreasuryOutput{Amount: 10_000_000}
	inputID := tpkg.Rand32ByteArray()
	sampleTreasuryTx := &iotago.TreasuryTransaction{Output: &iotago.TreasuryOutput{Amount: 3_000_000}}
	treasuryInput := &iotago.TreasuryInput{}
	copy(treasuryInput[:], inputID[:])
	sampleTreasuryTx.Input = treasuryInput

	tests := []test{
		func() test {
			receipt, _ := iotago.NewReceiptBuilder(100).AddEntry(&iotago.MigratedFundsEntry{
				TailTransactionHash: tpkg.Rand49ByteArray(),
				Address:             tpkg.RandEd25519Address(),
				Deposit:             7_000_000,
			}).AddTreasuryTransaction(sampleTreasuryTx).Build()
			return test{"ok", receipt, currentTreasury, nil}
		}(),
		func() test {
			receipt, _ := iotago.NewReceiptBuilder(100).AddEntry(&iotago.MigratedFundsEntry{
				TailTransactionHash: tpkg.Rand49ByteArray(),
				Address:             tpkg.RandEd25519Address(),
				Deposit:             1000,
			}).AddTreasuryTransaction(sampleTreasuryTx).Build()
			return test{"err - migrated less tha minimum", receipt, currentTreasury, iotago.ErrInvalidReceiptMilestoneOpt}
		}(),
		func() test {
			receipt, _ := iotago.NewReceiptBuilder(100).AddEntry(&iotago.MigratedFundsEntry{
				TailTransactionHash: tpkg.Rand49ByteArray(),
				Address:             tpkg.RandEd25519Address(),
				Deposit:             tpkg.TestTokenSupply + 1,
			}).AddTreasuryTransaction(sampleTreasuryTx).Build()
			return test{"err - total supply overflow", receipt, currentTreasury, iotago.ErrInvalidReceiptMilestoneOpt}
		}(),
		func() test {
			receipt, _ := iotago.NewReceiptBuilder(100).AddEntry(&iotago.MigratedFundsEntry{
				TailTransactionHash: tpkg.Rand49ByteArray(),
				Address:             tpkg.RandEd25519Address(),
				Deposit:             6_000_000,
			}).AddTreasuryTransaction(sampleTreasuryTx).Build()
			return test{"err - invalid new treasury amount", receipt, currentTreasury, iotago.ErrInvalidReceiptMilestoneOpt}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := iotago.ValidateReceipt(tt.source, tt.prevInput, tpkg.TestTokenSupply)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
		})
	}
}
