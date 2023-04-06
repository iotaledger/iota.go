package tpkg

import iotago "github.com/iotaledger/iota.go/v4"

// TestProtoParams is an instance of iotago.ProtocolParameters for testing purposes. It contains a zero vbyte rent cost.
// Only use this var in testing. Do not modify or use outside unit tests.
var TestProtoParams = &iotago.ProtocolParameters{
	Version:     3,
	NetworkName: "TestJungle",
	Bech32HRP:   "tgl",
	MinPoWScore: 0,
	RentStructure: iotago.RentStructure{
		VByteCost:    0,
		VBFactorData: 0,
		VBFactorKey:  0,
	},
	TokenSupply: TestTokenSupply,
}

// TestNetworkID is a test network ID.
var TestNetworkID = TestProtoParams.NetworkID()

const (
	// TestTokenSupply is a test token supply constant.
	// Do not use this constant outside of unit tests, instead, query it via a node.
	TestTokenSupply = 2_779_530_283_277_761

	// TestProtocolVersion is a dummy protocol version.
	// Do not use this constant outside of unit tests, instead, query it via a node.
	TestProtocolVersion = 3
)
