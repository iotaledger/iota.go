package tpkg

import (
	iotago "github.com/iotaledger/iota.go/v4"
)

var TestAPI = iotago.V3API(
	iotago.NewV3ProtocolParameters(
		iotago.WithNetworkOptions("TestJungle", "tgl"),
		iotago.WithSupplyOptions(TestTokenSupply, 0, 0, 0, 0, 0),
		iotago.WithWorkScoreOptions(0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0), // all zero except block offset gives all blocks workscore = 1
	),
)

// TestNetworkID is a test network ID.
var TestNetworkID = TestAPI.ProtocolParameters().NetworkID()

const (
	// TestTokenSupply is a test token supply constant.
	// Do not use this constant outside of unit tests, instead, query it via a node.
	TestTokenSupply = 2_779_530_283_277_761
)
