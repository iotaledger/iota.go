package nodeclient

import (
	"encoding/json"
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
)

// TODO: use the API instance from Client instead.
var _internalAPI = iotago.V2API(&iotago.ProtocolParameters{})

type (
	httpOutput   interface{ iotago.Output }
	outputResAux struct {
		Output httpOutput `serix:"0,mapKey=output"`
	}
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func init() {
	api := _internalAPI.Underlying()
	must(api.RegisterInterfaceObjects((*httpOutput)(nil), (*iotago.BasicOutput)(nil)))
	must(api.RegisterInterfaceObjects((*httpOutput)(nil), (*iotago.AliasOutput)(nil)))
	must(api.RegisterInterfaceObjects((*httpOutput)(nil), (*iotago.FoundryOutput)(nil)))
	must(api.RegisterInterfaceObjects((*httpOutput)(nil), (*iotago.NFTOutput)(nil)))
}

type (
	// RoutesResponse defines the response of a GET routes REST API call.
	RoutesResponse struct {
		Routes []string `json:"routes"`
	}

	// InfoResponse defines the response of a GET info REST API call.
	InfoResponse struct {
		// The name of the node software.
		Name string `json:"name"`
		// The semver version of the node software.
		Version string `json:"version"`
		// Status information.
		Status InfoResStatus `json:"status"`
		// Protocol information.
		Protocol *json.RawMessage `json:"protocol"`
		// BaseToken information.
		BaseToken *InfoResBaseToken `json:"baseToken"`
		// Metrics information.
		Metrics InfoResMetrics `json:"metrics"`
		// The features this node exposes.
		Features []string `json:"features"`
	}

	// InfoResStatus defines info res status information.
	InfoResStatus struct {
		// Whether the node is healthy.
		IsHealthy bool `json:"isHealthy"`
		// The latest known milestone index.
		LatestMilestone InfoResMilestone `json:"latestMilestone"`
		// The current confirmed milestone's index.
		ConfirmedMilestone InfoResMilestone `json:"confirmedMilestone"`
		// The milestone index at which the last pruning commenced.
		PruningIndex iotago.MilestoneIndex `json:"pruningIndex"`
	}

	// InfoResMilestone defines the info res milestone information.
	InfoResMilestone struct {
		// The index of the milestone.
		Index iotago.MilestoneIndex `json:"index"`
		// The unix time of the milestone payload.
		Timestamp uint32 `json:"timestamp,omitempty"`
		// The IO of the milestone.
		MilestoneID string `json:"milestoneId,omitempty"`
	}

	// InfoResBaseToken defines the info res base token information.
	InfoResBaseToken struct {
		// The base token name.
		Name string `json:"name"`
		// The base token ticker symbol.
		TickerSymbol string `json:"tickerSymbol"`
		// The base token unit.
		Unit string `json:"unit"`
		// The base token subunit.
		Subunit string `json:"subunit,omitempty"`
		// The base token amount of decimals.
		Decimals uint32 `json:"decimals"`
		// The base token uses the metric prefix.
		UseMetricPrefix bool `json:"useMetricPrefix"`
	}

	// InfoResMetrics defines info res metrics information.
	InfoResMetrics struct {
		// The current rate of new blocks per second.
		BlocksPerSecond float64 `json:"blocksPerSecond"`
		// The current rate of referenced blocks per second.
		ReferencedBlocksPerSecond float64 `json:"referencedBlocksPerSecond"`
		// The ratio of referenced blocks in relation to new blocks of the last confirmed milestone.
		ReferencedRate float64 `json:"referencedRate"`
	}

	// TipsResponse defines the response of a GET tips REST API call.
	TipsResponse struct {
		// The hex encoded block IDs of the tips.
		TipsHex []string `json:"tips"`
	}

	// BlockMetadataResponse defines the response of a GET block metadata REST API call.
	BlockMetadataResponse struct {
		// The hex encoded ID of the block.
		BlockID string `json:"blockId"`
		// The hex encoded IDs of the parent block references.
		Parents []string `json:"parents"`
		// Whether the block is solid.
		Solid bool `json:"isSolid"`
		// The milestone index that references this block.
		ReferencedByMilestoneIndex iotago.MilestoneIndex `json:"referencedByMilestoneIndex,omitempty"`
		// If this block represents a milestone this is the milestone index
		MilestoneIndex iotago.MilestoneIndex `json:"milestoneIndex,omitempty"`
		// The ledger inclusion state of the transaction payload.
		LedgerInclusionState string `json:"ledgerInclusionState,omitempty"`
		// Whether the block should be promoted.
		ShouldPromote *bool `json:"shouldPromote,omitempty"`
		// Whether the block should be reattached.
		ShouldReattach *bool `json:"shouldReattach,omitempty"`
		// The reason why this block is marked as conflicting.
		ConflictReason uint8 `json:"conflictReason,omitempty"`
		// If this block is referenced by a milestone this returns the index of that block inside the milestone by whiteflag ordering.
		WhiteFlagIndex *uint32 `json:"whiteFlagIndex,omitempty"`
	}

	// ChildrenResponse defines the response of a GET children REST API call.
	ChildrenResponse struct {
		// The hex encoded block ID of the block.
		BlockID string `json:"blockId"`
		// The maximum count of results that are returned by the node.
		MaxResults uint32 `json:"maxResults"`
		// The actual count of results that are returned.
		Count uint32 `json:"count"`
		// The hex encoded IDs of the children of this block.
		Children []string `json:"children"`
	}

	// OutputMetadataResponse defines the response of a GET outputs metadata REST API call.
	OutputMetadataResponse struct {
		// The hex encoded ID of the block.
		BlockID string `json:"blockId"`
		// The hex encoded transaction id from which this output originated.
		TransactionID string `json:"transactionId"`
		// The index of the output.
		OutputIndex uint16 `json:"outputIndex"`
		// Whether this output is spent.
		Spent bool `json:"isSpent"`
		// The milestone index at which this output was spent.
		MilestoneIndexSpent iotago.MilestoneIndex `json:"milestoneIndexSpent,omitempty"`
		// The milestone timestamp this output was spent.
		MilestoneTimestampSpent uint32 `json:"milestoneTimestampSpent,omitempty"`
		// The transaction this output was spent with.
		TransactionIDSpent string `json:"transactionIdSpent,omitempty"`
		// The milestone index at which this output was booked into the ledger.
		MilestoneIndexBooked iotago.MilestoneIndex `json:"milestoneIndexBooked"`
		// The milestone timestamp this output was booked in the ledger.
		MilestoneTimestampBooked uint32 `json:"milestoneTimestampBooked"`
		// The ledger index at which this output was available at.
		LedgerIndex iotago.MilestoneIndex `json:"ledgerIndex"`
	}

	// OutputResponse defines the response of a GET outputs REST API call.
	OutputResponse struct {
		Metadata *OutputMetadataResponse `json:"metadata"`
		// The output in its serialized form.
		RawOutput *json.RawMessage `json:"output"`
	}

	// TreasuryResponse defines the response of a GET treasury REST API call.
	TreasuryResponse struct {
		MilestoneID string `json:"milestoneId"`
		Amount      string `json:"amount"`
	}

	// ReceiptsResponse defines the response for receipts GET related REST API calls.
	ReceiptsResponse struct {
		Receipts []*ReceiptTuple `json:"receipts"`
	}

	// ReceiptTuple represents a receipt and the milestone index in which it was contained.
	ReceiptTuple struct {
		Receipt        *json.RawMessage      `json:"receipt"`
		MilestoneIndex iotago.MilestoneIndex `json:"milestoneIndex"`
	}

	// MilestoneUTXOChangesResponse defines the response of a GET milestone UTXO changes REST API call.
	MilestoneUTXOChangesResponse struct {
		// The index of the milestone.
		Index iotago.MilestoneIndex `json:"index"`
		// The output IDs (transaction hash + output index) of the newly created outputs.
		CreatedOutputs []string `json:"createdOutputs"`
		// The output IDs (transaction hash + output index) of the consumed (spent) outputs.
		ConsumedOutputs []string `json:"consumedOutputs"`
	}

	// ComputeWhiteFlagMutationsRequest defines the request for a POST ComputeWhiteFlagMutations REST API call.
	ComputeWhiteFlagMutationsRequest struct {
		// The index of the milestone.
		Index iotago.MilestoneIndex `json:"index"`
		// The timestamp of the milestone.
		Timestamp uint32 `json:"timestamp"`
		// The hex encoded IDs of the parent blocks the milestone references.
		Parents []string `json:"parents"`
		// The hex encoded milestone ID of the previous milestone.
		PreviousMilestoneID string `json:"previousMilestoneId"`
	}

	// ComputeWhiteFlagMutationsResponseInternal defines the internal response for a POST ComputeWhiteFlagMutations REST API call.
	ComputeWhiteFlagMutationsResponseInternal struct {
		// The hex encoded inclusion merkle tree root as a result of the white flag computation.
		InclusionMerkleRoot string `json:"inclusionMerkleRoot"`
		// The hex encoded applied merkle tree root as a result of the white flag computation.
		AppliedMerkleRoot string `json:"appliedMerkleRoot"`
	}

	// ComputeWhiteFlagMutationsResponse defines the response for a POST ComputeWhiteFlagMutations REST API call.
	ComputeWhiteFlagMutationsResponse struct {
		// The inclusion merkle tree root as a result of the white flag computation.
		InclusionMerkleRoot iotago.MilestoneMerkleProof
		// The applied merkle tree root as a result of the white flag computation.
		AppliedMerkleRoot iotago.MilestoneMerkleProof
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
		SolidMilestoneIndex iotago.MilestoneIndex `json:"solidMilestoneIndex"`
		// The milestone index at which the node pruned its data.
		PrunedMilestoneIndex iotago.MilestoneIndex `json:"prunedMilestoneIndex"`
		// The latest known milestone index by the node.
		LatestMilestoneIndex iotago.MilestoneIndex `json:"latestMilestoneIndex"`
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
		// The total amount of received milestone requests.
		ReceivedMilestoneRequests uint32 `json:"receivedMilestoneRequests"`
		// The total amount of received heartbeats.
		ReceivedHeartbeats uint32 `json:"receivedHeartbeats"`
		// The total amount of sent blocks.
		SentBlocks uint32 `json:"sentBlocks"`
		// The total amount of sent block request.
		SentBlockRequests uint32 `json:"sentBlockRequests"`
		// The total amount of sent milestone request.
		SentMilestoneRequests uint32 `json:"sentMilestoneRequests"`
		// The total amount of sent heartbeats.
		SentHeartbeats uint32 `json:"sentHeartbeats"`
		// The total amount of packets which couldn't be sent.
		DroppedPackets uint32 `json:"droppedPackets"`
	}
)

// TxID returns the TransactionID.
func (nor *OutputMetadataResponse) TxID() (*iotago.TransactionID, error) {
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
	outResJson, err := json.Marshal(nor)
	if err != nil {
		return nil, err
	}

	o := &outputResAux{}
	if err := _internalAPI.JSONDecode(outResJson, o); err != nil {
		return nil, err
	}

	return o.Output, nil
}

// ProtocolParameters returns the protocol parameters within the info response.
func (info *InfoResponse) ProtocolParameters() (*iotago.ProtocolParameters, error) {
	protoJson, err := json.Marshal(info.Protocol)
	if err != nil {
		return nil, err
	}

	o := &iotago.ProtocolParameters{}
	if err := _internalAPI.JSONDecode(protoJson, o); err != nil {
		return nil, err
	}

	return o, nil
}
