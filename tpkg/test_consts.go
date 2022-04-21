package tpkg

import "github.com/iotaledger/iota.go/v3"

// TestProtoParas is an instance of iotago.ProtocolParameters for testing purposes. It contains a zero vbyte rent cost.
// Only use this var in testing. Do not modify or use outside unit tests.
var TestProtoParas = &iotago.ProtocolParameters{
	TokenSupply: TestTokenSupply,
	RentStructure: iotago.RentStructure{
		VByteCost:    0,
		VBFactorData: 0,
		VBFactorKey:  0,
	},
}

const (
	// TestTokenSupply is a test token supply constant.
	// Do not use this constant outside of unit tests, instead, query it via a node.
	TestTokenSupply = 2_779_530_283_277_761

	// TestProtocolVersion is a dummy protocol version.
	// Do not use this constant outside of unit tests, instead, query it via a node.
	TestProtocolVersion = 2
)
