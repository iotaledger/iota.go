package nodeclient

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/iota.go/v3"
)

type (
	// InfoResponse defines the response of a GET info REST API call.
	InfoResponse struct {
		// The name of the node software.
		Name string `json:"name"`
		// The semver version of the node software.
		Version string `json:"version"`
		// Status information.
		Status InfoResStatus `json:"status"`
		// Protocol information.
		Protocol InfoResProtocol `json:"protocol"`
		// Metrics information.
		Metrics InfoResMetrics `json:"metrics"`
		// The features this node exposes.
		Features []string `json:"features"`
		// The plugins this node exposes.
		Plugins []string `json:"plugins"`
	}

	// InfoResStatus defines info res status information.
	InfoResStatus struct {
		// Whether the node is healthy.
		IsHealthy bool `json:"isHealthy"`
		// The timestamp of the latest known milestone.
		LatestMilestoneTimestamp int64 `json:"latestMilestoneTimestamp"`
		// The latest known milestone index.
		LatestMilestoneIndex uint32 `json:"latestMilestoneIndex"`
		// The current confirmed milestone's index.
		ConfirmedMilestoneIndex uint32 `json:"confirmedMilestoneIndex"`
		// The milestone index at which the last pruning commenced.
		PruningIndex uint32 `json:"pruningIndex"`
	}

	// InfoResProtocol defines info res protocol information.
	InfoResProtocol struct {
		// The human friendly name of the network on which the node operates on.
		NetworkName string `json:"networkName"`
		// The HRP prefix used for Bech32 addresses in the node's network.
		Bech32HRP string `json:"bech32HRP"`
		// The minimum pow score of the network.
		MinPowScore float64 `json:"minPoWScore"`
		// The rent structure used by given node/network.
		RentStructure iotago.RentStructure `json:"rentStructure"`
	}

	// InfoResMetrics defines info res metrics information.
	InfoResMetrics struct {
		// The current rate of new messages per second.
		MessagesPerSecond float64 `json:"messagesPerSecond"`
		// The current rate of referenced messages per second.
		ReferencedMessagesPerSecond float64 `json:"referencedMessagesPerSecond"`
		// The ratio of referenced messages in relation to new messages of the last confirmed milestone.
		ReferencedRate float64 `json:"referencedRate"`
	}

	// MessageMetadataResponse defines the response of a GET message metadata REST API call.
	MessageMetadataResponse struct {
		// The hex encoded message ID of the message.
		MessageID string `json:"messageId"`
		// The hex encoded message IDs of the parents the message references.
		Parents []string `json:"parentMessageIds"`
		// Whether the message is solid.
		Solid bool `json:"isSolid"`
		// The milestone index that references this message.
		ReferencedByMilestoneIndex *uint32 `json:"referencedByMilestoneIndex,omitempty"`
		// If this message represents a milestone this is the milestone index
		MilestoneIndex *uint32 `json:"milestoneIndex,omitempty"`
		// The ledger inclusion state of the transaction payload.
		LedgerInclusionState *string `json:"ledgerInclusionState,omitempty"`
		// Whether the message should be promoted.
		ShouldPromote *bool `json:"shouldPromote,omitempty"`
		// Whether the message should be reattached.
		ShouldReattach *bool `json:"shouldReattach,omitempty"`
		// The reason why this message is marked as conflicting.
		ConflictReason uint8 `json:"conflictReason,omitempty"`
	}

	// ChildrenResponse defines the response of a GET children REST API call.
	ChildrenResponse struct {
		// The hex encoded message ID of the message.
		MessageID string `json:"messageId"`
		// The maximum count of results that are returned by the node.
		MaxResults uint32 `json:"maxResults"`
		// The actual count of results that are returned.
		Count uint32 `json:"count"`
		// The hex encoded message IDs of the children of this message.
		Children []string `json:"childrenMessageIds"`
	}

	// OutputResponse defines the response of a GET outputs REST API call.
	OutputResponse struct {
		// The hex encoded message ID of the message.
		MessageID string `json:"messageId"`
		// The hex encoded transaction id from which this output originated.
		TransactionID string `json:"transactionId"`
		// The index of the output.
		OutputIndex uint16 `json:"outputIndex"`
		// Whether this output is spent.
		Spent bool `json:"isSpent"`
		// The ledger index at which this output was available at.
		LedgerIndex uint64 `json:"ledgerIndex"`
		// The output in its serialized form.
		RawOutput *json.RawMessage `json:"output"`
	}

	// TreasuryResponse defines the response of a GET treasury REST API call.
	TreasuryResponse struct {
		MilestoneID string `json:"milestoneId"`
		Amount      uint64 `json:"amount"`
	}

	// ReceiptsResponse defines the response for receipts GET related REST API calls.
	ReceiptsResponse struct {
		Receipts []*ReceiptTuple `json:"receipts"`
	}

	// ReceiptTuple represents a receipt and the milestone index in which it was contained.
	ReceiptTuple struct {
		Receipt        *iotago.Receipt `json:"receipt"`
		MilestoneIndex uint32          `json:"milestoneIndex"`
	}

	// MilestoneResponse defines the response of a GET milestones REST API call.
	MilestoneResponse struct {
		// The index of the milestone.
		Index uint32 `json:"index"`
		// The hex encoded message ID of the message.
		MessageID string `json:"messageId"`
		// The unix time of the milestone payload.
		Time int64 `json:"timestamp"`
	}

	// MilestoneUTXOChangesResponse defines the response of a GET milestone UTXO changes REST API call.
	MilestoneUTXOChangesResponse struct {
		// The index of the milestone.
		Index uint32 `json:"index"`
		// The output IDs (transaction hash + output index) of the newly created outputs.
		CreatedOutputs []string `json:"createdOutputs"`
		// The output IDs (transaction hash + output index) of the consumed (spent) outputs.
		ConsumedOutputs []string `json:"consumedOutputs"`
	}

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
	//	- a new milestone was received
	//	- the solid milestone changed
	//	- the node performed pruning of data
	GossipHeartbeat struct {
		// The solid milestone of the node.
		SolidMilestoneIndex uint32 `json:"solidMilestoneIndex"`
		// The milestone index at which the node pruned its data.
		PrunedMilestoneIndex uint32 `json:"prunedMilestoneIndex"`
		// The latest known milestone index by the node.
		LatestMilestoneIndex uint32 `json:"latestMilestoneIndex"`
		// The amount of currently connected neighbors.
		ConnectedNeighbors int `json:"connectedNeighbors"`
		// The amount of currently connected neighbors who also
		// are synchronized with the network.
		SyncedNeighbors int `json:"syncedNeighbors"`
	}

	// PeerGossipMetrics defines the peer gossip metrics.
	PeerGossipMetrics struct {
		// The total amount of received new messages.
		NewMessages uint32 `json:"newMessages"`
		// The total amount of received known messages.
		KnownMessages uint32 `json:"knownMessages"`
		// The total amount of received messages.
		ReceivedMessages uint32 `json:"receivedMessages"`
		// The total amount of received message requests.
		ReceivedMessageRequests uint32 `json:"receivedMessageRequests"`
		// The total amount of received milestone requests.
		ReceivedMilestoneRequests uint32 `json:"receivedMilestoneRequests"`
		// The total amount of received heartbeats.
		ReceivedHeartbeats uint32 `json:"receivedHeartbeats"`
		// The total amount of sent messages.
		SentMessages uint32 `json:"sentMessages"`
		// The total amount of sent message request.
		SentMessageRequests uint32 `json:"sentMessageRequests"`
		// The total amount of sent milestone request.
		SentMilestoneRequests uint32 `json:"sentMilestoneRequests"`
		// The total amount of sent heartbeats.
		SentHeartbeats uint32 `json:"sentHeartbeats"`
		// The total amount of packets which couldn't be sent.
		DroppedPackets uint32 `json:"droppedPackets"`
	}
)

// TxID returns the TransactionID.
func (nor *OutputResponse) TxID() (*iotago.TransactionID, error) {
	txIDBytes, err := iotago.DecodeHex(nor.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("unable to decode raw transaction ID from JSON to transaction ID: %w", err)
	}
	var txID iotago.TransactionID
	copy(txID[:], txIDBytes)
	return &txID, nil
}

// Output deserializes the RawOutput to an Output.
func (nor *OutputResponse) Output() (iotago.Output, error) {
	jsonSeri, err := iotago.DeserializeObjectFromJSON(nor.RawOutput, iotago.JsonOutputSelector)
	if err != nil {
		return nil, err
	}
	seri, err := jsonSeri.ToSerializable()
	if err != nil {
		return nil, err
	}
	output, isOutput := seri.(iotago.Output)
	if !isOutput {
		return nil, iotago.ErrUnknownOutputType
	}
	return output, nil
}
