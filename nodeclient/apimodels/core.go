package apimodels

import (
	"time"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v4"
)

type BlockState byte
type BlockFailureReason byte

const (
	BlockStateLength         = serializer.OneByte
	BlockFailureReasonLength = serializer.OneByte
)

const (
	BlockStateUnknown BlockState = iota
	BlockStatePending
	BlockStateAccepted
	BlockStateConfirmed
	BlockStateFinalized
	BlockStateRejected
	BlockStateFailed
)

func (b BlockState) String() string {
	return []string{
		"unknown",
		"pending",
		"accepted",
		"confirmed",
		"finalized",
		"rejected",
		"failed",
	}[b]
}

func (b BlockState) Bytes() ([]byte, error) {
	return []byte{byte(b)}, nil
}

func BlockStateFromBytes(b []byte) (BlockState, int, error) {
	if len(b) < BlockStateLength {
		return 0, 0, ierrors.New("invalid block state size")
	}

	return BlockState(b[0]), BlockStateLength, nil
}

const (
	BlockFailureNone                   BlockFailureReason = 0
	BlockFailureIsTooOld               BlockFailureReason = 1
	BlockFailureParentIsTooOld         BlockFailureReason = 2
	BlockFailureParentNotFound         BlockFailureReason = 3
	BlockFailureParentInvalid          BlockFailureReason = 4
	BlockFailureDroppedDueToCongestion BlockFailureReason = 5
	BlockFailurePayloadInvalid         BlockFailureReason = 6

	// TODO: see if needed after congestion PR is done.
	BlockFailureOrphanedDueNegativeCreditsBalance BlockFailureReason = 6
)

func (t BlockFailureReason) Bytes() ([]byte, error) {
	return []byte{byte(t)}, nil
}

func BlockFailureReasonFromBytes(b []byte) (BlockFailureReason, int, error) {
	if len(b) < BlockFailureReasonLength {
		return 0, 0, ierrors.New("invalid block failure reason size")
	}

	return BlockFailureReason(b[0]), BlockFailureReasonLength, nil
}

type TransactionState byte
type TransactionFailureReason byte

const (
	TransactionStateLength         = serializer.OneByte
	TransactionFailureReasonLength = serializer.OneByte
)

const (
	TransactionStateNoTransaction TransactionState = iota
	TransactionStatePending
	TransactionStateAccepted
	TransactionStateConfirmed
	TransactionStateFinalized
	TransactionStateFailed
)

func (t TransactionState) String() string {
	return []string{
		"noTransaction",
		"pending",
		"accepted",
		"confirmed",
		"finalized",
		"failed",
	}[t]
}

func (t TransactionState) Bytes() ([]byte, error) {
	return []byte{byte(t)}, nil
}

func TransactionStateFromBytes(b []byte) (TransactionState, int, error) {
	if len(b) < TransactionStateLength {
		return 0, 0, ierrors.New("invalid transaction state size")
	}

	return TransactionState(b[0]), TransactionStateLength, nil
}

const (
	TxFailureNone                                  TransactionFailureReason = 0
	TxFailureUTXOInputAlreadySpent                 TransactionFailureReason = 1
	TxFailureConflicting                           TransactionFailureReason = 2
	TxFailureUTXOInputInvalid                      TransactionFailureReason = 3
	TxFailureTxTypeInvalid                         TransactionFailureReason = 4
	TxFailureSumOfInputAndOutputValuesDoesNotMatch TransactionFailureReason = 5
	TxFailureUnlockBlockSignatureInvalid           TransactionFailureReason = 6
	TxFailureConfiguredTimelockNotYetExpired       TransactionFailureReason = 7
	TxFailureGivenNativeTokensInvalid              TransactionFailureReason = 8
	TxFailureReturnAmountNotFulfilled              TransactionFailureReason = 9
	TxFailureInputUnlockInvalid                    TransactionFailureReason = 10
	TxFailureInputsCommitmentInvalid               TransactionFailureReason = 11
	TxFailureSenderNotUnlocked                     TransactionFailureReason = 12
	TxFailureChainStateTransitionInvalid           TransactionFailureReason = 13
	TxFailureInputCreationAfterTxCreation          TransactionFailureReason = 14
	TxFailureManaAmountInvalid                     TransactionFailureReason = 15
	TxFailureBICInputInvalid                       TransactionFailureReason = 16
	TxFailureRewardInputInvalid                    TransactionFailureReason = 17
	TxFailureCommitmentInputInvalid                TransactionFailureReason = 18
	TxFailureNoStakingFeature                      TransactionFailureReason = 19
	TxFailureFailedToClaimStakingReward            TransactionFailureReason = 20
	TxFailureFailedToClaimDelegationReward         TransactionFailureReason = 21
	TxFailureSemanticValidationFailed              TransactionFailureReason = 255
)

func (t TransactionFailureReason) Bytes() ([]byte, error) {
	return []byte{byte(t)}, nil
}

func TransactionFailureReasonFromBytes(b []byte) (TransactionFailureReason, int, error) {
	if len(b) < TransactionFailureReasonLength {
		return 0, 0, ierrors.New("invalid transaction failure reason size")
	}

	return TransactionFailureReason(b[0]), TransactionFailureReasonLength, nil
}

type (
	// InfoResponse defines the response of a GET info REST API call.
	InfoResponse struct {
		// The name of the node software.
		Name string `serix:"0,mapKey=name"`
		// The semver version of the node software.
		Version string `serix:"1,mapKey=version"`
		// The current status of this node.
		Status *InfoResNodeStatus `serix:"2,mapKey=status"`
		// The metrics of this node.
		Metrics *InfoResNodeMetrics `serix:"3,mapKey=metrics"`
		// The protocol parameters used by this node.
		ProtocolParameters []*InfoResProtocolParameters `serix:"4,mapKey=protocolParameters"`
		// The base token of the network.
		BaseToken *InfoResBaseToken `serix:"5,mapKey=baseToken"`
		// The features this node exposes.
		Features []string `serix:"6,mapKey=features"`
	}

	InfoResProtocolParameters struct {
		StartEpoch iotago.EpochIndex         `serix:"0,mapKey=startEpoch"`
		Parameters iotago.ProtocolParameters `serix:"1,mapKey=parameters"`
	}

	// InfoResNodeStatus defines the status of the node in info response.
	InfoResNodeStatus struct {
		// Whether the node is healthy.
		IsHealthy bool `serix:"0,mapKey=isHealthy"`
		// The accepted tangle time.
		AcceptedTangleTime time.Time `serix:"1,mapKey=acceptedTangleTime"`
		// The relative accepted tangle time.
		RelativeAcceptedTangleTime time.Time `serix:"2,mapKey=relativeAcceptedTangleTime"`
		// The confirmed tangle time.
		ConfirmedTangleTime time.Time `serix:"3,mapKey=confirmedTangleTime"`
		// The relative confirmed tangle time.
		RelativeConfirmedTangleTime time.Time `serix:"4,mapKey=relativeConfirmedTangleTime"`
		// The id of the latest known commitment.
		LatestCommitmentID iotago.CommitmentID `serix:"5,mapKey=latestCommitmentId"`
		// The index of the latest finalized slot.
		LatestFinalizedSlot iotago.SlotIndex `serix:"6,mapKey=latestFinalizedSlot"`
		// The slot index of the latest accepted block.
		LatestAcceptedBlockSlot iotago.SlotIndex `serix:"7,mapKey=latestAcceptedBlockSlot"`
		// The slot index of the latest confirmed block.
		LatestConfirmedBlockSlot iotago.SlotIndex `serix:"8,mapKey=latestConfirmedBlockSlot"`
		// The index of the epoch at which the tangle data was pruned.
		PruningEpoch iotago.EpochIndex `serix:"9,mapKey=pruningEpoch"`
	}

	// InfoResNodeMetrics defines the metrics of a node in info response.
	InfoResNodeMetrics struct {
		// The current rate of new blocks per second, it's updated when a commitment is committed.
		BlocksPerSecond float64 `serix:"0,mapKey=blocksPerSecond"`
		// The current rate of confirmed blocks per second, it's updated when a commitment is committed.
		ConfirmedBlocksPerSecond float64 `serix:"1,mapKey=confirmedBlocksPerSecond"`
		// The ratio of confirmed blocks in relation to new blocks up until the latest commitment is committed.
		ConfirmationRate float64 `serix:"2,mapKey=confirmationRate"`
	}

	// InfoResBaseToken defines the info res base token information.
	InfoResBaseToken struct {
		// The base token name.
		Name string `serix:"0,mapKey=name"`
		// The base token ticker symbol.
		TickerSymbol string `serix:"1,mapKey=tickerSymbol"`
		// The base token unit.
		Unit string `serix:"2,mapKey=unit"`
		// The base token subunit.
		Subunit string `serix:"3,mapKey=subunit,omitempty"`
		// The base token amount of decimals.
		Decimals uint32 `serix:"4,mapKey=decimals"`
		// The base token uses the metric prefix.
		UseMetricPrefix bool `serix:"5,mapKey=useMetricPrefix"`
	}

	// IssuanceBlockHeaderResponse defines the response of a GET block issuance REST API call.
	IssuanceBlockHeaderResponse struct {
		// StrongParents are the strong parents of the block.
		StrongParents iotago.BlockIDs `serix:"0,lengthPrefixType=uint8,mapKey=strongParents"`
		// WeakParents are the weak parents of the block.
		WeakParents iotago.BlockIDs `serix:"1,lengthPrefixType=uint8,mapKey=weakParents"`
		// ShallowLikeParents are the shallow like parents of the block.
		ShallowLikeParents iotago.BlockIDs `serix:"2,lengthPrefixType=uint8,mapKey=shallowLikeParents"`
		// LatestFinalizedSlot is the index of the latest finalized slot.
		LatestFinalizedSlot iotago.SlotIndex `serix:"3,mapKey=latestFinalizedSlot"`
		// Commitment is the commitment of the block.
		Commitment *iotago.Commitment `serix:"4,mapKey=commitment"`
	}

	// BlockCreatedResponse defines the response of a POST blocks REST API call.
	BlockCreatedResponse struct {
		// The hex encoded block ID of the block.
		BlockID iotago.BlockID `serix:"0,mapKey=blockId"`
	}

	// BlockMetadataResponse defines the response of a GET block metadata REST API call.
	BlockMetadataResponse struct {
		// BlockID The hex encoded block ID of the block.
		BlockID iotago.BlockID `serix:"0,mapKey=blockId"`
		// BlockState might be pending, rejected, failed, confirmed, finalized.
		BlockState string `serix:"1,mapKey=blockState"`
		// BlockFailureReason if applicable indicates the error that occurred during the block processing.
		BlockFailureReason BlockFailureReason `serix:"2,mapKey=blockFailureReason,omitempty"`
		// TxState might be pending, conflicting, confirmed, finalized, rejected.
		TxState string `serix:"3,mapKey=txState,omitempty"`
		// TxFailureReason if applicable indicates the error that occurred during the transaction processing.
		TxFailureReason TransactionFailureReason `serix:"4,mapKey=txFailureReason,omitempty"`
	}

	// OutputMetadataResponse defines the response of a GET outputs metadata REST API call.
	OutputMetadataResponse struct {
		// BlockID is the block ID that contains the output.
		BlockID iotago.BlockID `serix:"0,mapKey=blockId"`
		// TransactionID is the transaction ID that creates the output.
		TransactionID iotago.TransactionID `serix:"1,mapKey=transactionId"`
		// OutputIndex is the index of the output.
		OutputIndex uint16 `serix:"2,mapKey=outputIndex"`
		// IsSpent indicates whether the output is spent or not.
		IsSpent bool `serix:"3,mapKey=isSpent"`
		// CommitmentIDSpent is the commitment ID that includes the spent output.
		CommitmentIDSpent iotago.CommitmentID `serix:"4,mapKey=commitmentIdSpent,omitempty"`
		// TransactionIDSpent is the transaction ID that spends the output.
		TransactionIDSpent iotago.TransactionID `serix:"5,mapKey=transactionIdSpent,omitempty"`
		// IncludedCommitmentID is the commitment ID that includes the output.
		IncludedCommitmentID iotago.CommitmentID `serix:"6,mapKey=includedCommitmentId,omitempty"`
		// LatestCommitmentID is the latest commitment ID of a node.
		LatestCommitmentID iotago.CommitmentID `serix:"7,mapKey=latestCommitmentId"`
	}

	// UTXOChangesResponse defines the response for UTXO slot REST API call.
	UTXOChangesResponse struct {
		// The index of the requested commitment.
		Index iotago.SlotIndex `serix:"0,mapKey=index"`
		// The outputs that are created in this slot.
		CreatedOutputs iotago.OutputIDs `serix:"1,mapKey=createdOutputs"`
		// The outputs that are consumed in this slot.
		ConsumedOutputs iotago.OutputIDs `serix:"2,mapKey=consumedOutputs"`
	}

	// CongestionResponse defines the response for the congestion REST API call.
	CongestionResponse struct {
		// SlotIndex is the index of the slot for which the estimate is provided
		SlotIndex iotago.SlotIndex `serix:"0,mapKey=slotIndex"`
		// Ready indicates if a node is ready to issue a block in a current congestion or should wait.
		Ready bool `serix:"1,mapKey=ready"`
		// ReferenceManaCost (RMC) is the mana cost a user needs to burn to issue a block in SlotIndex slot.
		ReferenceManaCost iotago.Mana `serix:"2,mapKey=referenceManaCost"`
		// BlockIssuanceCredits (BIC) is the mana a user has on its BIC account exactly slotIndex - MaxCommittableASge in the past.
		// This balance needs to be > 0 zero, otherwise account is locked
		BlockIssuanceCredits iotago.BlockIssuanceCredits `serix:"3,mapKey=blockIssuanceCredits"`
	}

	// ValidatorResponse defines the response used in stakers response REST API calls.
	ValidatorResponse struct {
		// AccountID is the hex encoded account ID of the validator.
		AccountID iotago.AccountID `serix:"0,mapKey=accountId"`
		// StakingEpochEnd is the epoch until which the validator registered to stake.
		StakingEpochEnd iotago.EpochIndex `serix:"1,mapKey=stakingEpochEnd"`
		// PoolStake is the sum of tokens delegated to the pool and the validator stake.
		PoolStake iotago.BaseToken `serix:"2,mapKey=poolStake"`
		// ValidatorStake is the stake of the validator.
		ValidatorStake iotago.BaseToken `serix:"3,mapKey=validatorStake"`
		// FixedCost is the fixed cost that the validator receives from the total pool reward.
		FixedCost iotago.Mana `serix:"4,mapKey=fixedCost"`
		// Active indicates whether the validator was active recently, and would be considered during committee selection.
		Active bool `serix:"5,mapKey=active"`
		// LatestSupportedProtocolVersion is the latest supported protocol version of the validator.
		LatestSupportedProtocolVersion iotago.Version    `serix:"6,mapKey=latestSupportedProtocolVersion"`
		LatestSupportedProtocolHash    iotago.Identifier `serix:"7,mapKey=latestSupportedProtocolHash"`
	}

	// ValidatorsResponse defines the response for the staking REST API call.
	ValidatorsResponse struct {
		Validators []*ValidatorResponse `serix:"0,mapKey=stakers"`
		PageSize   uint32               `serix:"1,mapKey=pageSize"`
		Cursor     string               `serix:"2,mapKey=cursor,omitempty"`
	}

	// ManaRewardsResponse defines the response for the mana rewards REST API call.
	ManaRewardsResponse struct {
		// EpochStart is the starting epoch for the range for which the mana rewards are returned.
		EpochStart iotago.EpochIndex `serix:"0,mapKey=epochIndexStart"`
		// EpochEnd is the ending epoch for the range for which the mana rewards are returned, also the decay is only applied up to this point.
		EpochEnd iotago.EpochIndex `serix:"1,mapKey=epochIndexEnd"`
		// The amount of totally available rewards the requested output may claim, decayed up to EpochEnd (including).
		Rewards iotago.Mana `serix:"2,mapKey=rewards"`
	}

	// CommitteeMemberResponse defines the response used in committee and staking response REST API calls.
	CommitteeMemberResponse struct {
		// AccountID is the hex encoded account ID of the validator.
		AccountID iotago.AccountID `serix:"0,mapKey=accountId"`
		// PoolStake is the sum of tokens delegated to the pool and the validator stake.
		PoolStake iotago.BaseToken `serix:"1,mapKey=poolStake"`
		// ValidatorStake is the stake of the validator.
		ValidatorStake iotago.BaseToken `serix:"2,mapKey=validatorStake"`
		// FixedCost is the fixed cost that the validator received from the total pool reward.
		FixedCost iotago.Mana `serix:"3,mapKey=fixedCost"`
	}

	// CommitteeResponse defines the response for the staking REST API call.
	CommitteeResponse struct {
		Committee           []*CommitteeMemberResponse `serix:"0,mapKey=committee"`
		TotalStake          iotago.BaseToken           `serix:"1,mapKey=totalStake"`
		TotalValidatorStake iotago.BaseToken           `serix:"2,mapKey=totalValidatorStake"`
		EpochIndex          iotago.EpochIndex          `serix:"3,mapKey=epochIndex"`
	}
)
