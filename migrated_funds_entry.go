package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	// MinMigratedFundsEntryDeposit defines the minimum amount a MigratedFundsEntry must deposit.
	MinMigratedFundsEntryDeposit = 1_000_000
	// LegacyTailTransactionHashLength denotes the length of a legacy transaction.
	LegacyTailTransactionHashLength = 49
	// MigratedFundsEntrySerializedBytesSize is the serialized size of a MigratedFundsEntry.
	MigratedFundsEntrySerializedBytesSize = LegacyTailTransactionHashLength + Ed25519AddressSerializedBytesSize + serializer.UInt64ByteSize
)

// LegacyTailTransactionHash represents the bytes of a T5B1 encoded legacy tail transaction hash.
type LegacyTailTransactionHash = [49]byte

// MigratedFundsEntries is a slice of MigratedFundsEntry.
type MigratedFundsEntries []*MigratedFundsEntry

func (o MigratedFundsEntries) Clone() MigratedFundsEntries {
	cpy := make(MigratedFundsEntries, len(o))
	for i, or := range o {
		cpy[i] = or.Clone()
	}
	return cpy
}

func (o MigratedFundsEntries) Size() int {
	return serializer.UInt16ByteSize + (len(o) * MigratedFundsEntrySerializedBytesSize)
}

func (o *MigratedFundsEntries) FromAny(slice []any) {
	*o = make(MigratedFundsEntries, len(slice))
	for i, ele := range slice {
		(*o)[i] = ele.(*MigratedFundsEntry)
	}
}

// MigratedFundsEntry are funds which were migrated from a legacy network.
type MigratedFundsEntry struct {
	// The tail transaction hash of the migration bundle.
	TailTransactionHash LegacyTailTransactionHash `serix:"0,mapKey=tailTransactionHash"`
	// The target address of the migrated funds.
	Address Address `serix:"1,mapKey=address"`
	// The amount of the deposit.
	Deposit uint64 `serix:"2,mapKey=deposit"`
}

func (m *MigratedFundsEntry) Clone() *MigratedFundsEntry {
	cpy := &MigratedFundsEntry{
		Address: m.Address.Clone(),
		Deposit: m.Deposit,
	}
	copy(cpy.TailTransactionHash[:], m.TailTransactionHash[:])
	return cpy
}
