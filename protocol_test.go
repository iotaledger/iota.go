package iotago_test

import iotago "github.com/iotaledger/iota.go/v3"

const (
	OneMi = 1_000_000
)

var (
	// DefZeroRentParas are the default parameters for de/serialization using zero vbyte rent cost.
	DefZeroRentParas = &iotago.DeSerializationParameters{RentStructure: &iotago.RentStructure{
		VByteCost:    0,
		VBFactorData: 0,
		VBFactorKey:  0,
	}}
)
