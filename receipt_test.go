package iotago_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/tpkg"

	"github.com/iotaledger/iota.go/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReceipt_DeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "",
			source: tpkg.RandReceipt(),
			target: &iotago.Receipt{},
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
			m := &iotago.Receipt{}
			_, err := m.Deserialize(tt.in, serializer.DeSeriModePerformValidation, nil)
			if err != nil {
				return
			}

			seriData, err := m.Serialize(serializer.DeSeriModePerformValidation, nil)
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
			}).AddTreasuryTransaction(sampleTreasuryTx).Build(iotago.ZeroRentParas)
			return test{"ok", receipt, currentTreasury, nil}
		}(),
		func() test {
			receipt, _ := iotago.NewReceiptBuilder(100).AddEntry(&iotago.MigratedFundsEntry{
				TailTransactionHash: tpkg.Rand49ByteArray(),
				Address:             tpkg.RandEd25519Address(),
				Deposit:             1000,
			}).AddTreasuryTransaction(sampleTreasuryTx).Build(iotago.ZeroRentParas)
			return test{"err - migrated less tha minimum", receipt, currentTreasury, iotago.ErrInvalidReceipt}
		}(),
		func() test {
			receipt, _ := iotago.NewReceiptBuilder(100).AddEntry(&iotago.MigratedFundsEntry{
				TailTransactionHash: tpkg.Rand49ByteArray(),
				Address:             tpkg.RandEd25519Address(),
				Deposit:             iotago.TokenSupply + 1,
			}).AddTreasuryTransaction(sampleTreasuryTx).Build(iotago.ZeroRentParas)
			return test{"err - total supply overflow", receipt, currentTreasury, iotago.ErrInvalidReceipt}
		}(),
		func() test {
			receipt, _ := iotago.NewReceiptBuilder(100).AddEntry(&iotago.MigratedFundsEntry{
				TailTransactionHash: tpkg.Rand49ByteArray(),
				Address:             tpkg.RandEd25519Address(),
				Deposit:             6_000_000,
			}).AddTreasuryTransaction(sampleTreasuryTx).Build(iotago.ZeroRentParas)
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
