package models

import (
	"encoding/json"

	"github.com/iotaledger/hive.go/ierrors"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

type (
	Versions []uint32

	BlockState       string
	TransactionState string
)

const (
	BlockStatePending   BlockState = "pending"
	BlockStateConfirmed BlockState = "confirmed"
	BlockStateFinalized BlockState = "finalized"

	ErrBlockInvalid                        = 1
	ErrBlockOrphanedDueToCongestionControl = 2
	ErrBlockOrphanedDueToNegativeCredits   = 3

	TransactionStatePending   TransactionState = "pending"
	TransactionStateConfirmed TransactionState = "confirmed"
	TransactionStateFinalized TransactionState = "finalized"

	ErrTxStateReferencedUTXOAlreadySpent            = 1
	ErrTxStateTxConflicting                         = 2
	ErrTxStateReferencedUTXONotFound                = 3
	ErrTxStateSumOfInputAndOutputValuesDoesNotMatch = 4
	ErrTxStateUnlockBlockSignatureInvalid           = 5
	ErrTxStateConfiguredTimelockNotYetExpired       = 6
	ErrTxStateGivenNativeTokensInvalid              = 7
	ErrTxStateReturnAmountNotFulfilled              = 8
	ErrTxStateInputUnlockInvalid                    = 9
	ErrTxStateInputsCommitmentInvalid               = 10
	ErrTxStateSenderNotUnlocked                     = 11
	ErrTxStateChainStateTransitionInvalid           = 12
	ErrTxStateSemanticValidationFailed              = 255
)

type (
	// InfoResponse defines the response of a GET info REST API call.
	InfoResponse struct {
		// The name of the node software.
		Name string `json:"name"`
		// The semver version of the node software.
		Version string `json:"version"`
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
		// The accepted tangle time.
		AcceptedTangleTime uint64 `json:"acceptedTangleTime"`
		// The relative accepted tangle time.
		RelativeAcceptedTangleTime uint64 `json:"relativeAcceptedTangleTime"`
		// The confirmed tangle time.
		ConfirmedTangleTime uint64 `json:"confirmedTangleTime"`
		// The relative confirmed tangle time.
		RelativeConfirmedTangleTime uint64 `json:"relativeConfirmedTangleTime"`
		// The index of the latest known committed slot.
		LatestCommittedSlot iotago.SlotIndex `json:"latestCommittedSlot"`
		// The index of the latest finalized slot.
		LatestFinalizedSlot iotago.SlotIndex `json:"latestFinalizedSlot"`
		// The index of the slot at which the tangle data was pruned.
		PruningSlot iotago.SlotIndex `json:"pruningSlot"`
		// The blockID of the latest accepted block.
		LatestAcceptedBlockID string `json:"latestAcceptedBlockId"`
		// The blockID of the latest confirmed block.
		LatestConfirmedBlockID string `json:"latestConfirmedBlockId"`
	}

	// InfoResNodeMetrics defines the metrics of a node in info response.
	InfoResNodeMetrics struct {
		// The current rate of new blocks per second, it's updated when a commitment is committed.
		BlocksPerSecond float64 `json:"blocksPerSecond"`
		// The current rate of confirmed blocks per second, it's updated when a commitment is committed.
		ConfirmedBlocksPerSecond float64 `json:"confirmedBlocksPerSecond"`
		// The ratio of confirmed blocks in relation to new blocks up until the latest commitment is committed.
		ConfirmationRate float64 `json:"confirmationRate"`
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

	// IssuanceBlockHeaderResponse defines the response of a GET block issuance REST API call.
	IssuanceBlockHeaderResponse struct {
		// StrongParents are the strong parents of the block.
		StrongParents []string `json:"strongParents"`
		// WeakParents are the weak parents of the block.
		WeakParents []string `json:"weakParents"`
		// ShallowLikeParents are the shallow like parents of the block.
		ShallowLikeParents []string `json:"shallowLikeParents"`
		// LatestFinalizedSlot is the index of the latest finalized slot.
		LatestFinalizedSlot iotago.SlotIndex `json:"latestFinalizedSlot"`
		// Commitment is the commitment of the block.
		Commitment iotago.Commitment `json:"commitment"`
	}

	// BlockCreatedResponse defines the response of a POST blocks REST API call.
	BlockCreatedResponse struct {
		// The hex encoded block ID of the block.
		BlockID string `json:"blockId"`
	}

	// BlockMetadataResponse defines the response of a GET block metadata REST API call.
	BlockMetadataResponse struct {
		// BlockID The hex encoded block ID of the block.
		BlockID string `json:"blockId"`
		// BlockState might be pending, confirmed, finalized.
		BlockState BlockState `json:"blockState"`
		// TxState might be pending, conflicting, confirmed, finalized, rejected.
		TxState TransactionState `json:"txState,omitempty"`
		// BlockStateReason if applicable indicates the error that occurred during the block processing.
		BlockStateReason int `json:"blockStateReason,omitempty"`
		// TxStateReason if applicable indicates the error that occurred during the transaction processing.
		TxStateReason int `json:"txStateReason,omitempty"`
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

	// UTXOChangesResponse defines the response for UTXO slot REST API call.
	UTXOChangesResponse struct {
		// The index of the requested commitment.
		Index iotago.SlotIndex `json:"index"`
		// The outputs that are created in this slot.
		CreatedOutputs iotago.HexOutputIDs `json:"createdOutputs"`
		// The outputs that are consumed in this slot.
		ConsumedOutputs iotago.HexOutputIDs `json:"consumedOutputs"`
	}

	//CongestionResponse defines the response for the congestion REST API call.
	CongestionResponse struct {
		// SlotIndex is the index of the slot for which the estimate is provided
		SlotIndex iotago.SlotIndex `json:"slotIndex"`
		// Ready indicates if a node is ready to issue a block in a current congestion or should wait.
		Ready bool `json:"ready"`
		// ReferenceManaCost (RMC) is the mana cost a user needs to burn to issue a block in SlotIndex slot.
		ReferenceManaCost uint64 `json:"referenceManaCost"`
		// BlockIssuanceCredits (BIC) is the mana a user has on its BIC account exactly slotIndex - MaxCommittableASge in the past.
		// This balance needs to be > 0 zero, otherwise account is locked
		BlockIssuanceCredits uint64 `json:"blockIssuanceCredits"`
	}

	// BlockIssuanceCreditsResponse defines the response for the block issuance credits REST API call.
	BlockIssuanceCreditsResponse struct {
		// SlotIndex is the index of the slot corresponding to the block issuance credits value returned.
		SlotIndex iotago.SlotIndex `json:"slotIndex"`
		// BlockIssuanceCredits is the block issuance credits value for the slot. Node is able to provide values only of already committed slots.
		BlockIssuanceCredits uint64 `json:"blockIssuanceCredits"`
	}

	// ValidatorResponse defines the response used in stakers response REST API calls.
	ValidatorResponse struct {
		// AccountID is the hex encoded account ID of the validator.
		AccountID string `json:"accountId"`
		// StakingEpochEnd is the epoch until which the validator registered to stake.
		StakingEpochEnd iotago.EpochIndex `json:"stakingEpochEnd"`
		// PoolStake is the sum of tokens delegated to the pooland the validator stake.
		PoolStake uint64 `json:"poolStake"`
		// ValidatorStake is the stake of the validator.
		ValidatorStake uint64 `json:"validatorStake"`
		// FixedCost is the fixed cost that the validator reciews from the total pool reward.
		FixedCost uint64 `json:"fixedCost"`
		// LatestSuccessfulReward is the latest successful reward of the validator.
		LatestSupportedProtocolVersion uint64 `json:"latestSupportedProtocolVersion"`
	}

	// AccountStakingListResponse defines the response for the staking REST API call.
	AccountStakingListResponse struct {
		Stakers []ValidatorResponse `json:"stakers"`
	}

	// ManaRewardsResponse defines the response for the mana rewards REST API call.
	ManaRewardsResponse struct {
		// EpochIndex is the epoch index for which the mana rewards are returned.
		EpochIndex iotago.EpochIndex `json:"epochIndex"`
		// The amount of totally available rewards the requested output may claim.
		TotalReward uint64 `json:"totalReward"`
	}

	// CommitteeMemberResponse defines the response used in committee and staking response REST API calls.
	CommitteeMemberResponse struct {
		// AccountID is the hex encoded account ID of the validator.
		AccountID string `json:"accountId"`
		// PoolStake is the sum of tokens delegated to the pooland the validator stake.
		PoolStake uint64 `json:"poolStake"`
		// ValidatorStake is the stake of the validator.
		ValidatorStake uint64 `json:"validatorStake"`
		// FixedCost is the fixed cost that the validator reciews from the total pool reward.
		FixedCost uint64 `json:"fixedCost"`
	}

	// CommitteeResponse defines the response for the staking REST API call.
	CommitteeResponse struct {
		Committee           []CommitteeMemberResponse `json:"stakers"`
		TotalStake          uint64                    `json:"totalStake"`
		TotalValidatorStake uint64                    `json:"totalValidatorStake"`
		EpochIndex          iotago.EpochIndex         `json:"epochIndex"`
	}
)

// TxID returns the TransactionID.
func (o *OutputMetadataResponse) TxID() (*iotago.TransactionID, error) {
	txIDBytes, err := hexutil.DecodeHex(o.TransactionID)
	if err != nil {
		return nil, ierrors.Errorf("unable to decode raw transaction ID from JSON to transaction ID: %w", err)
	}
	var txID iotago.TransactionID
	copy(txID[:], txIDBytes)
	return &txID, nil
}

// DecodeProtocolParameters returns the protocol parameters within the info response.
func (i *InfoResponse) DecodeProtocolParameters() (iotago.ProtocolParameters, error) {
	protoJson, err := json.Marshal(i.ProtocolParameters)
	if err != nil {
		return nil, err
	}

	var o iotago.ProtocolParameters
	if err := _internalAPI.JSONDecode(protoJson, &o); err != nil {
		return nil, err
	}

	return o, nil
}

// DecodeCommitment returns the commitment within the block issuance response.
func (i *IssuanceBlockHeaderResponse) DecodeCommitment() (*iotago.Commitment, error) {
	commitmentJson, err := json.Marshal(i.Commitment)
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
