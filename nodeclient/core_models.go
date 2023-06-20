package nodeclient

import (
	"encoding/json"
	"fmt"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

// TODO: use the API instance from Client instead.
var _internalAPI = iotago.V3API(&iotago.ProtocolParameters{})

type (
	httpOutput interface{ iotago.Output }
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func init() {
	api := _internalAPI.Underlying()
	must(api.RegisterInterfaceObjects((*httpOutput)(nil), (*iotago.BasicOutput)(nil)))
	must(api.RegisterInterfaceObjects((*httpOutput)(nil), (*iotago.AccountOutput)(nil)))
	must(api.RegisterInterfaceObjects((*httpOutput)(nil), (*iotago.FoundryOutput)(nil)))
	must(api.RegisterInterfaceObjects((*httpOutput)(nil), (*iotago.NFTOutput)(nil)))
}

type (
	Versions []uint32

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
		// The ID of the node
		IssuerID string `json:"issuerId"`
		// The current status of this node.
		Status *InfoResNodeStatus `json:"status"`
		// The metrics of this node.
		Metrics *InfoResNodeMetrics `json:"metrics"`
		// The protocol versions this node supports.
		SupportedProtocolVersions Versions `json:"supportedProtocolVersions"`
		// The protocol parameters used by this node.
		ProtocolParameters *json.RawMessage `json:"protocol"`
		// The base token of the network.
		BaseToken *InfoResBaseToken `json:"baseToken"`
		// The features this node exposes.
		Features []string `json:"features"`
	}

	// InfoResNodeStatus defines the status of the node in info response.
	InfoResNodeStatus struct {
		// Whether the node is healthy.
		IsHealthy bool `json:"isHealthy"`
		// The blockID of last accepted block.
		LastAcceptedBlockID string `json:"lastAcceptedBlockId"`
		// The blockID of the last confirmed block
		LastConfirmedBlockID string `json:"lastConfirmedBlockId"`
		// The latest finalized slot.
		FinalizedSlot iotago.SlotIndex `json:"latestFinalizedSlot"`
		// The Accepted Tangle Time
		ATT uint64 `json:"ATT"`
		// The Relative Accepted Tangle Time
		RATT uint64 `json:"RATT"`
		// The Confirmed Tangle Time
		CTT uint64 `json:"CTT"`
		// The Relative Confirmed Tangle Time
		RCTT uint64 `json:"RCTT"`
		// The latest known committed slot info.
		LatestCommittedSlot iotago.SlotIndex `json:"latestCommittedSlot"`
		// The slot index at which the last pruning commenced.
		PruningSlot iotago.SlotIndex `json:"pruningSlot"`
	}

	// InfoResNodeMetrics defines the metrics of a node in info response.
	InfoResNodeMetrics struct {
		// The current rate of new blocks per second, it's updated when a commitment is committed.
		BlocksPerSecond float64 `json:"blocksPerSecond"`
		// The current rate of confirmed blocks per second, it's updated when a commitment is committed.
		ConfirmedBlocksPerSecond float64 `json:"confirmedBlocksPerSecond"`
		// The ratio of confirmed blocks in relation to new blocks up until the latest commitment is committed.
		ConfirmedRate float64 `json:"confirmedRate"`
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

	// BlockIssuanceResponse defines the response of a GET block issuance REST API call.
	BlockIssuanceResponse struct {
		// StrongParents are the strong parents of the block.
		StrongParents []string `json:"strongParents"`
		// WeakParents are the weak parents of the block.
		WeakParents []string `json:"weakParents"`
		// ShallowLikeParents are the shallow like parents of the block.
		ShallowLikeParents []string `json:"shallowLikeParents"`
		// LatestFinalizedSlot is the lastest finalized slot index.
		LatestFinalizedSlot iotago.SlotIndex `json:"latestFinalizedSlot"`
		// Commitment is the commitment of the block.
		Commitment *json.RawMessage `json:"commitment"`
	}

	// BlockMetadataResponse defines the response of a GET block metadata REST API call.
	BlockMetadataResponse struct {
		// BlockID The hex encoded block ID of the block.
		BlockID string `json:"blockId"`
		// StrongParents are the strong parents of the block.
		StrongParents []string `json:"strongParents"`
		// WeakParents are the weak parents of the block.
		WeakParents []string `json:"weakParents"`
		// ShallowLikeParents are the shallow like parents of the block.
		ShallowLikeParents []string `json:"shallowLikeParents"`
		// BlockState might be pending, confirmed, finalized.
		BlockState string `json:"blockState"`
		// TxState might be pending, conflicting, confirmed, finalized, rejected.
		TxState string `json:"txState,omitempty"`
		// BlockStateReason if applicable indicates the error that occurred during the block processing.
		BlockStateReason string `json:"blockStateReason,omitempty"`
		// TxStateReason if applicable indicates the error that occurred during the transaction processing.
		TxStateReason string `json:"txStateReason,omitempty"`
		// ReissuePayload whether the block should be reissued.
		ReissuePayload *bool `json:"reissuePayload,omitempty"`
	}

	// OutputMetadataResponse defines the response of a GET outputs metadata REST API call.
	OutputMetadataResponse struct {
		// BlockID is the block ID that contains the output.
		BlockID string `json:"blockId"`
		// TransactionID is the transaction ID that creates the output.
		TransactionID string `json:"transactionId"`
		// OutputIndex is the index of the output.
		OutputIndex uint16 `json:"outputIndex"`
		// IsSpent indicates whether the output is spent or not.
		IsSpent bool `json:"isSpent"`
		// CommitmentIDSpent is the commitment ID that includes the spent output.
		CommitmentIDSpent string `json:"commitmentIdSpent,omitempty"`
		// TransactionIDSpent is the transaction ID that spends the output.
		TransactionIDSpent string `json:"transactionIdSpent,omitempty"`
		// IncludedCommitmentID is the commitment ID that includes the output.
		IncludedCommitmentID string `json:"includedCommitmentId,omitempty"`
		// LatestCommitmentID is the latest commitment ID of a node.
		LatestCommitmentID string `json:"latestCommitmentId"`
	}

	// CommitmentDetailsResponse defines the response of a GET milestone UTXO changes REST API call.
	CommitmentDetailsResponse struct {
		// The index of the requested commitment.
		Index iotago.SlotIndex `json:"index"`
		// The commitment ID of previous commitment.
		PrevID string `json:"prevId"`
		// The roots ID of merkle trees within the requested commitment.
		RootsID string `json:"rootsId"`
		// The cumulative weight of the requested slot.
		CumulativeWeight uint64 `json:"cumulativeWeight"`
		// TODO: decide what else to add here.
	}

	// UTXOChangesResponse defines the response for UTXO slot REST API call.
	UTXOChangesResponse struct {
		// The index of the requested commitment.
		Index iotago.SlotIndex `json:"index"`
		// The outputs that are created in this slot.
		CreatedOutputs []string `json:"createdOutputs"`
		// The outputs that are consumed in this slot.
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
		SolidMilestoneIndex iotago.SlotIndex `json:"solidMilestoneIndex"`
		// The milestone index at which the node pruned its data.
		PrunedMilestoneIndex iotago.SlotIndex `json:"prunedMilestoneIndex"`
		// The latest known milestone index by the node.
		LatestMilestoneIndex iotago.SlotIndex `json:"latestMilestoneIndex"`
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
	txIDBytes, err := hexutil.DecodeHex(nor.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("unable to decode raw transaction ID from JSON to transaction ID: %w", err)
	}
	var txID iotago.TransactionID
	copy(txID[:], txIDBytes)
	return &txID, nil
}

// ProtocolParameters returns the protocol parameters within the info response.
func (info *InfoResponse) DecodeProtocolParameters() (*iotago.ProtocolParameters, error) {
	protoJson, err := json.Marshal(info.ProtocolParameters)
	if err != nil {
		return nil, err
	}

	o := &iotago.ProtocolParameters{}
	if err := _internalAPI.JSONDecode(protoJson, o); err != nil {
		return nil, err
	}

	return o, nil
}

// DecodeCommitment returns the commitment within the block issuance response.
func (b *BlockIssuanceResponse) DecodeCommitment() (*iotago.Commitment, error) {
	commitmentJson, err := json.Marshal(b.Commitment)
	if err != nil {
		return nil, err
	}

	o := &iotago.Commitment{}
	if err := _internalAPI.JSONDecode(commitmentJson, o); err != nil {
		return nil, err
	}

	return o, nil
}

// Highest returns the highest version.
func (v Versions) Highest() byte {
	return byte(v[len(v)-1])
}

// Lowest returns the lowest version.
func (v Versions) Lowest() byte {
	return byte(v[0])
}

// Supports tells whether the given version is supported.
func (v Versions) Supports(ver byte) bool {
	for _, value := range v {
		if value == uint32(ver) {
			return true
		}
	}

	return false
}
