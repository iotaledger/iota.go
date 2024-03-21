package api

import (
	iotago "github.com/iotaledger/iota.go/v4"
)

type (
	// AddPeerRequest defines the request for a POST peer REST API call.
	AddPeerRequest struct {
		// The libp2p multi address of the peer.
		MultiAddress string `serix:",lenPrefix=uint8"`
		// The alias to identify the peer.
		Alias string `serix:",lenPrefix=uint8,omitempty"`
	}

	// PeerInfo defines the response of a GET peer REST API call.
	PeerInfo struct {
		// The libp2p identifier of the peer.
		ID string `serix:",lenPrefix=uint8"`
		// The libp2p multi addresses of the peer.
		MultiAddresses []iotago.PrefixedStringUint8 `serix:",lenPrefix=uint8"`
		// The alias to identify the peer.
		Alias string `serix:",lenPrefix=uint8,omitempty"`
		// The relation (manual, autopeered) of the peer.
		Relation string `serix:",lenPrefix=uint8"`
		// Whether the peer is connected.
		Connected bool `serix:""`
		// The gossip metrics for this peer.
		GossipMetrics *PeerGossipMetrics `serix:""`
	}

	// PeerGossipMetrics defines the peer gossip metrics.
	PeerGossipMetrics struct {
		// The total amount of received packets.
		PacketsReceived uint32 `serix:""`
		// The total amount of sent packets.
		PacketsSent uint32 `serix:""`
	}

	PeersResponse struct {
		Peers []*PeerInfo `serix:",lenPrefix=uint8"`
	}

	// PruneDatabaseRequest defines the request of a prune database REST API call.
	PruneDatabaseRequest struct {
		// The pruning target epoch.
		Epoch iotago.EpochIndex `serix:",omitempty"`
		// The pruning depth.
		Depth iotago.EpochIndex `serix:",omitempty"`
		// The target size of the database.
		TargetDatabaseSize string `serix:",lenPrefix=uint8,omitempty"`
	}

	// PruneDatabaseResponse defines the response of a prune database REST API call.
	PruneDatabaseResponse struct {
		// The current oldest epoch in the database.
		Epoch iotago.EpochIndex `serix:""`
	}
)
