package models

import (
	iotago "github.com/iotaledger/iota.go/v4"
)

type (
	// AddPeerRequest defines the request for a POST peer REST API call.
	AddPeerRequest struct {
		// The libp2p multi address of the peer.
		MultiAddress string `json:"multiAddress"`
		// The alias of to iditify the peer.
		Alias *string `json:"alias,omitempty"`
	}

	// PeerResponse defines the response of a GET peer REST API call.
	PeerResponse struct {
		// The libp2p identifier of the peer.
		ID string `json:"id"`
		// The libp2p multi addresses of the peer.
		MultiAddresses []string `json:"multiAddresses"`
		// The alias to identify the peer.
		Alias *string `json:"alias,omitempty"`
		// The relation (static, autopeered) of the peer.
		Relation string `json:"relation"`
		// Whether the peer is connected.
		Connected bool `json:"connected"`
		// The gossip related information about this peer.
		Gossip *GossipInfo `json:"gossip,omitempty"`
	}

	// GossipInfo represents information about an ongoing gossip protocol.
	GossipInfo struct {
		// The last received heartbeat by the given node.
		Heartbeat *GossipHeartbeat `json:"heartbeat"`
		// The metrics about sent and received protocol messages.
		Metrics PeerGossipMetrics `json:"metrics"`
	}

	// GossipHeartbeat represents a gossip heartbeat message.
	// Peers send each other this gossip protocol message when their
	// state is updated, such as when:
	//	- a new slot was received
	//	- the solid slot changed
	//	- the node performed pruning of data
	GossipHeartbeat struct {
		// The solid slot of the node.
		SolidSlotIndex iotago.SlotIndex `json:"solidSlotIndex"`
		// The oldest known slot index by the node.
		PrunedSlotIndex iotago.SlotIndex `json:"prunedSlotIndex"`
		// The latest known slot index by the node.
		LatestSlotIndex iotago.SlotIndex `json:"latestSlotIndex"`
		// The amount of currently connected peers.
		ConnectedPeers int `json:"connectedPeers"`
		// The amount of currently connected peers who also
		// are synchronized with the network.
		SyncedPeers int `json:"syncedPeers"`
	}

	// PeerGossipMetrics defines the peer gossip metrics.
	PeerGossipMetrics struct {
		// The total amount of received new blocks.
		NewBlocks uint32 `json:"newBlocks"`
		// The total amount of received known blocks.
		KnownBlocks uint32 `json:"knownBlocks"`
		// The total amount of received blocks.
		ReceivedBlocks uint32 `json:"receivedBlocks"`
		// The total amount of received block requests.
		ReceivedBlockRequests uint32 `json:"receivedBlockRequests"`
		// The total amount of received slot requests.
		ReceivedSlotRequests uint32 `json:"receivedSlotRequests"`
		// The total amount of received heartbeats.
		ReceivedHeartbeats uint32 `json:"receivedHeartbeats"`
		// The total amount of sent blocks.
		SentBlocks uint32 `json:"sentBlocks"`
		// The total amount of sent block request.
		SentBlockRequests uint32 `json:"sentBlockRequests"`
		// The total amount of sent slot request.
		SentSlotRequests uint32 `json:"sentSlotRequests"`
		// The total amount of sent heartbeats.
		SentHeartbeats uint32 `json:"sentHeartbeats"`
		// The total amount of packets which couldn't be sent.
		DroppedPackets uint32 `json:"droppedPackets"`
	}

	// PruneDatabaseRequest defines the request of a prune database REST API call.
	PruneDatabaseRequest struct {
		// The pruning target slot index.
		Index *iotago.SlotIndex `json:"index,omitempty"`
		// The pruning depth.
		Depth *iotago.SlotIndex `json:"depth,omitempty"`
		// The target size of the database.
		TargetDatabaseSize *string `json:"targetDatabaseSize,omitempty"`
	}

	// PruneDatabaseResponse defines the response of a prune database REST API call.
	PruneDatabaseResponse struct {
		// The index of the current oldest slot in the database.
		Index iotago.SlotIndex `json:"index"`
	}

	// CreateSnapshotsRequest defines the request of a create snapshots REST API call.
	CreateSnapshotsRequest struct {
		// The index of the snapshot.
		Index iotago.SlotIndex `json:"index"`
	}

	// CreateSnapshotsResponse defines the response of a create snapshots REST API call.
	CreateSnapshotsResponse struct {
		// The index of the snapshot.
		Index iotago.SlotIndex `json:"index"`
		// The file path of the snapshot file.
		FilePath string `json:"filePath"`
	}
)
