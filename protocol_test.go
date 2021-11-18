package iotago_test

import iotago "github.com/iotaledger/iota.go/v3"

var (
	// DefDeSeriParas are the default parameters for de/serialization.
	DefDeSeriParas = &iotago.DeSerializationParameters{RentStructure: &iotago.RentStructure{
		VByteCost:    0,
		VBFactorData: 0,
		VBFactorKey:  0,
	}}
)
