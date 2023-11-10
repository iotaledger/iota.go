package apimodels

import (
	iotago "github.com/iotaledger/iota.go/v4"
)

type (
	// AddPeerRequest defines the request for a POST peer REST API call.
	AddPeerRequest struct {
		// The libp2p multi address of the peer.
		MultiAddress string `serix:""`
		// The alias of to identify the peer.
		Alias string `serix:",omitempty"`
	}

	// PeerInfo defines the response of a GET peer REST API call.
	PeerInfo struct {
		// The libp2p identifier of the peer.
		ID string `serix:""`
		// The libp2p multi addresses of the peer.
		MultiAddresses []string `serix:",lenPrefix=uint8"`
		// The alias to identify the peer.
		Alias string `serix:",omitempty"`
		// The relation (static, autopeered) of the peer.
		Relation string `serix:""`
		// Whether the peer is connected.
		Connected bool `serix:""`
		// The gossip related information about this peer.
		Gossip *GossipInfo `serix:",omitempty"`
	}

	PeersResponse struct {
		Peers []*PeerInfo `serix:",lenPrefix=uint8"`
	}

	// GossipInfo represents information about an ongoing gossip protocol.
	GossipInfo struct {
		// The last received heartbeat by the given node.
		Heartbeat *GossipHeartbeat `serix:",omitempty"`
		// The metrics about sent and received protocol messages.
		Metrics *PeerGossipMetrics `serix:",omitempty"`
	}

	// GossipHeartbeat represents a gossip heartbeat message.
	// Peers send each other this gossip protocol message when their
	// state is updated, such as when:
	//	- a new slot was received
	//	- the solid slot changed
	//	- the node performed pruning of data
	GossipHeartbeat struct {
		// The solid slot of the node.
		SolidSlot iotago.SlotIndex `serix:""`
		// The oldest known slot by the node.
		PrunedSlot iotago.SlotIndex `serix:""`
		// The latest known slot by the node.
		LatestSlot iotago.SlotIndex `serix:""`
		// The amount of currently connected peers.
		ConnectedPeers uint32 `serix:""`
		// The amount of currently connected peers who also
		// are synchronized with the network.
		SyncedPeers uint32 `serix:""`
	}

	// PeerGossipMetrics defines the peer gossip metrics.
	PeerGossipMetrics struct {
		// The total amount of received new blocks.
		NewBlocks uint32 `serix:""`
		// The total amount of received known blocks.
		KnownBlocks uint32 `serix:""`
		// The total amount of received blocks.
		ReceivedBlocks uint32 `serix:""`
		// The total amount of received block requests.
		ReceivedBlockRequests uint32 `serix:""`
		// The total amount of received slot requests.
		ReceivedSlotRequests uint32 `serix:""`
		// The total amount of received heartbeats.
		ReceivedHeartbeats uint32 `serix:""`
		// The total amount of sent blocks.
		SentBlocks uint32 `serix:""`
		// The total amount of sent block request.
		SentBlockRequests uint32 `serix:""`
		// The total amount of sent slot request.
		SentSlotRequests uint32 `serix:""`
		// The total amount of sent heartbeats.
		SentHeartbeats uint32 `serix:""`
		// The total amount of packets which couldn't be sent.
		DroppedPackets uint32 `serix:""`
	}

	// PruneDatabaseRequest defines the request of a prune database REST API call.
	PruneDatabaseRequest struct {
		// The pruning target epoch.
		Epoch iotago.EpochIndex `serix:",omitempty"`
		// The pruning depth.
		Depth iotago.EpochIndex `serix:",omitempty"`
		// The target size of the database.
		TargetDatabaseSize string `serix:",omitempty"`
	}

	// PruneDatabaseResponse defines the response of a prune database REST API call.
	PruneDatabaseResponse struct {
		// The current oldest epoch in the database.
		Epoch iotago.EpochIndex `serix:""`
	}

	// CreateSnapshotsRequest defines the request of a create snapshots REST API call.
	CreateSnapshotsRequest struct {
		// The slot of the snapshot.
		Slot iotago.SlotIndex `serix:""`
	}

	// CreateSnapshotsResponse defines the response of a create snapshots REST API call.
	CreateSnapshotsResponse struct {
		// The slot of the snapshot.
		Slot iotago.SlotIndex `serix:""`
		// The file path of the snapshot file.
		FilePath string `serix:""`
	}
)
