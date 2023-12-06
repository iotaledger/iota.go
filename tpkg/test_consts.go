package tpkg

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v4"
)

// IOTAMainnetV3TestProtocolParameters reflect the planned protocol parameters to be used for IOTA mainnet.
// TODO: provide a link to the IOTA mainnet protocol parameters TIP.
var IOTAMainnetV3TestProtocolParameters = iotago.NewV3SnapshotProtocolParameters()

// ShimmerMainnetV3TestProtocolParameters reflect the planned protocol parameters to be used for Shimmer mainnet.
// TODO: provide a link to the Shimmer mainnet protocol parameters TIP.
var ShimmerMainnetV3TestProtocolParameters = iotago.NewV3SnapshotProtocolParameters(
	iotago.WithStorageOptions(100, 1, 10, 100, 100, 100),
	iotago.WithWorkScoreOptions(0, 1, 0, 0, 0, 0, 0, 0, 0, 0),
	iotago.WithTimeProviderOptions(0, time.Now().Unix(), 10, 13),
	iotago.WithLivenessOptions(15, 30, 10, 20, 60),
	iotago.WithSupplyOptions(1813620509061365, 63, 1, 17, 32, 21, 70),
	iotago.WithCongestionControlOptions(1, 0, 0, 800_000, 500_000, 100_000, 1000, 100),
	iotago.WithStakingOptions(10, 10, 10),
	iotago.WithVersionSignalingOptions(7, 5, 7),
	iotago.WithRewardsOptions(8, 8, 31, 2, 1),
	iotago.WithTargetCommitteeSize(32),
)

// FixedGenesisV3TestProtocolParameters are protocol parameters with a fixed genesis value for testing purposes.
var FixedGenesisV3TestProtocolParameters = iotago.NewV3SnapshotProtocolParameters(
	iotago.WithTimeProviderOptions(65898, time.Unix(1690879505, 0).UTC().Unix(), 10, 13),
	iotago.WithLivenessOptions(15, 30, 10, 20, 60),
)

// ZeroCostV3TestProtocolParameters are protocol parameters that give zero storage costs and block workscore =1 for all blocks.
// This is useful for testing purposes.
var ZeroCostV3TestProtocolParameters = iotago.NewV3SnapshotProtocolParameters(
	iotago.WithStorageOptions(0, 0, 0, 0, 0, 0),               // zero storage score
	iotago.WithWorkScoreOptions(0, 1, 0, 0, 0, 0, 0, 0, 0, 0), // all blocks workscore = 1
)

var ZeroCostTestAPI = iotago.V3API(ZeroCostV3TestProtocolParameters)

// TestNetworkID is a test network ID.
var TestNetworkID = IOTAMainnetV3TestProtocolParameters.NetworkID()
