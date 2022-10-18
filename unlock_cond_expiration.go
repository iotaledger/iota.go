package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

// ExpirationUnlockCondition is an unlock condition which puts a time constraint on whether the receiver or return identity
// can consume an output depending on the latest confirmed milestone's timestamp T:
//   - only the receiver identity can consume the output, if T is before than the one defined in the condition.
//   - only the return identity can consume the output, if T is at the same time or after the one defined in the condition.
type ExpirationUnlockCondition struct {
	// The identity who is allowed to use the output after the expiration has happened.
	ReturnAddress Address `serix:"0,mapKey=returnAddress"`
	// The unix time in second resolution at which the expiration happens.
	UnixTime uint32 `serix:"1,mapKey=unixTime,omitempty"`
}

func (s *ExpirationUnlockCondition) Clone() UnlockCondition {
	return &ExpirationUnlockCondition{
		ReturnAddress: s.ReturnAddress.Clone(),
		UnixTime:      s.UnixTime,
	}
}

func (s *ExpirationUnlockCondition) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	return rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+serializer.UInt32ByteSize) +
		s.ReturnAddress.VBytes(rentStruct, nil)
}

func (s *ExpirationUnlockCondition) Equal(other UnlockCondition) bool {
	otherCond, is := other.(*ExpirationUnlockCondition)
	if !is {
		return false
	}

	switch {
	case !s.ReturnAddress.Equal(otherCond.ReturnAddress):
		return false
	case s.UnixTime != otherCond.UnixTime:
		return false
	}

	return true
}

func (s *ExpirationUnlockCondition) Type() UnlockConditionType {
	return UnlockConditionExpiration
}

func (s *ExpirationUnlockCondition) Size() int {
	return util.NumByteLen(byte(UnlockConditionExpiration)) + s.ReturnAddress.Size() +
		+util.NumByteLen(s.UnixTime)
}
