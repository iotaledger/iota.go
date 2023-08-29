package apimodels

import (
	iotago "github.com/iotaledger/iota.go/v4"
)

type (
	// AddPeerRequest defines the request for a POST peer REST API call.
	AddPeerRequest struct {
		// The libp2p multi address of the peer.
		MultiAddress string `serix:"0,mapKey=multiAddress"`
		// The alias of to identify the peer.
		Alias string `serix:"1,mapKey=alias,omitempty"`
	}

	// PeerInfo defines the response of a GET peer REST API call.
	PeerInfo struct {
		// The libp2p identifier of the peer.
		ID string `serix:"0,mapKey=id"`
		// The libp2p multi addresses of the peer.
		MultiAddresses []string `serix:"1,mapKey=multiAddresses,lengthPrefixType=uint8"`
		// The alias to identify the peer.
		Alias string `serix:"2,mapKey=alias,omitempty"`
		// The relation (static, autopeered) of the peer.
		Relation string `serix:"3,mapKey=relation"`
		// Whether the peer is connected.
		Connected bool `serix:"4,mapKey=connected"`
		// The gossip related information about this peer.
		Gossip *GossipInfo `serix:"5,mapKey=gossip,omitempty"`
	}

	PeersResponse struct {
		Peers []*PeerInfo `serix:"0,mapKey=peers,lengthPrefixType=uint8"`
	}

	// GossipInfo represents information about an ongoing gossip protocol.
	GossipInfo struct {
		// The last received heartbeat by the given node.
		Heartbeat *GossipHeartbeat `serix:"0,mapKey=heartbeat,omitempty"`
		// The metrics about sent and received protocol messages.
		Metrics *PeerGossipMetrics `serix:"1,mapKey=metrics,omitempty"`
	}

	// GossipHeartbeat represents a gossip heartbeat message.
	// Peers send each other this gossip protocol message when their
	// state is updated, such as when:
	//	- a new slot was received
	//	- the solid slot changed
	//	- the node performed pruning of data
	GossipHeartbeat struct {
		// The solid slot of the node.
		SolidSlotIndex iotago.SlotIndex `serix:"0,mapKey=solidSlotIndex"`
		// The oldest known slot index by the node.
		PrunedSlotIndex iotago.SlotIndex `serix:"1,mapKey=prunedSlotIndex"`
		// The latest known slot index by the node.
		LatestSlotIndex iotago.SlotIndex `serix:"2,mapKey=latestSlotIndex"`
		// The amount of currently connected peers.
		ConnectedPeers uint32 `serix:"3,mapKey=connectedPeers"`
		// The amount of currently connected peers who also
		// are synchronized with the network.
		SyncedPeers uint32 `serix:"4,mapKey=syncedPeers"`
	}

	// PeerGossipMetrics defines the peer gossip metrics.
	PeerGossipMetrics struct {
		// The total amount of received new blocks.
		NewBlocks uint32 `serix:"0,mapKey=newBlocks"`
		// The total amount of received known blocks.
		KnownBlocks uint32 `serix:"1,mapKey=knownBlocks"`
		// The total amount of received blocks.
		ReceivedBlocks uint32 `serix:"2,mapKey=receivedBlocks"`
		// The total amount of received block requests.
		ReceivedBlockRequests uint32 `serix:"3,mapKey=receivedBlockRequests"`
		// The total amount of received slot requests.
		ReceivedSlotRequests uint32 `serix:"4,mapKey=receivedSlotRequests"`
		// The total amount of received heartbeats.
		ReceivedHeartbeats uint32 `serix:"5,mapKey=receivedHeartbeats"`
		// The total amount of sent blocks.
		SentBlocks uint32 `serix:"6,mapKey=sentBlocks"`
		// The total amount of sent block request.
		SentBlockRequests uint32 `serix:"7,mapKey=sentBlockRequests"`
		// The total amount of sent slot request.
		SentSlotRequests uint32 `serix:"8,mapKey=sentSlotRequests"`
		// The total amount of sent heartbeats.
		SentHeartbeats uint32 `serix:"9,mapKey=sentHeartbeats"`
		// The total amount of packets which couldn't be sent.
		DroppedPackets uint32 `serix:"10,mapKey=droppedPackets"`
	}

	// PruneDatabaseRequest defines the request of a prune database REST API call.
	PruneDatabaseRequest struct {
		// The pruning target epoch index.
		Index iotago.EpochIndex `serix:"0,mapKey=index,omitempty"`
		// The pruning depth.
		Depth iotago.EpochIndex `serix:"1,mapKey=depth,omitempty"`
		// The target size of the database.
		TargetDatabaseSize string `serix:"2,mapKey=targetDatabaseSize,omitempty"`
	}

	// PruneDatabaseResponse defines the response of a prune database REST API call.
	PruneDatabaseResponse struct {
		// The index of the current oldest epoch in the database.
		Index iotago.EpochIndex `serix:"0,mapKey=index"`
	}

	// CreateSnapshotsRequest defines the request of a create snapshots REST API call.
	CreateSnapshotsRequest struct {
		// The index of the snapshot.
		Index iotago.SlotIndex `serix:"0,mapKey=index"`
	}

	// CreateSnapshotsResponse defines the response of a create snapshots REST API call.
	CreateSnapshotsResponse struct {
		// The index of the snapshot.
		Index iotago.SlotIndex `serix:"0,mapKey=index"`
		// The file path of the snapshot file.
		FilePath string `serix:"1,mapKey=filePath"`
	}
)
