package apimodels

import (
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v4"
)

type (
	Versions []uint32
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

type TransactionState byte
type TransactionFailureReason byte

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
	BlockFailureBookingFailure         BlockFailureReason = 3
	BlockFailureDroppedDueToCongestion BlockFailureReason = 4
	BlockFailurePayloadInvalid         BlockFailureReason = 5

	// TODO: see if needed after congestion PR is done.
	BlockFailureOrphanedDueNegativeCreditsBalance BlockFailureReason = 6
)

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
		// The protocol versions this node supports.
		SupportedProtocolVersions Versions `serix:"4,mapKey=supportedProtocolVersions"`
		// The protocol parameters used by this node.
		ProtocolParameters iotago.ProtocolParameters `serix:"5,mapKey=protocol"`
		// The base token of the network.
		BaseToken *InfoResBaseToken `serix:"6,mapKey=baseToken"`
		// The features this node exposes.
		Features []string `serix:"7,mapKey=features"`
	}

	// InfoResNodeStatus defines the status of the node in info response.
	InfoResNodeStatus struct {
		// Whether the node is healthy.
		IsHealthy bool `serix:"0,mapKey=isHealthy"`
		// The accepted tangle time.
		AcceptedTangleTime uint64 `serix:"1,mapKey=acceptedTangleTime"`
		// The relative accepted tangle time.
		RelativeAcceptedTangleTime uint64 `serix:"2,mapKey=relativeAcceptedTangleTime"`
		// The confirmed tangle time.
		ConfirmedTangleTime uint64 `serix:"3,mapKey=confirmedTangleTime"`
		// The relative confirmed tangle time.
		RelativeConfirmedTangleTime uint64 `serix:"4,mapKey=relativeConfirmedTangleTime"`
		// The id of the latest known commitment.
		LatestCommitmentID iotago.CommitmentID `serix:"5,mapKey=latestCommitmentId"`
		// The index of the latest finalized slot.
		LatestFinalizedSlot iotago.SlotIndex `serix:"6,mapKey=latestFinalizedSlot"`
		// The slot index of the latest accepted block.
		LatestAcceptedBlockSlot iotago.SlotIndex `serix:"7,mapKey=latestAcceptedBlockSlot"`
		// The slot index of the latest confirmed block.
		LatestConfirmedBlockSlot iotago.SlotIndex `serix:"8,mapKey=latestConfirmedBlockSlot"`
		// The index of the slot at which the tangle data was pruned.
		PruningSlot iotago.SlotIndex `serix:"9,mapKey=pruningSlot"`
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
		StrongParents iotago.BlockIDs `serix:"0,mapKey=strongParents"`
		// WeakParents are the weak parents of the block.
		WeakParents iotago.BlockIDs `serix:"1,mapKey=weakParents"`
		// ShallowLikeParents are the shallow like parents of the block.
		ShallowLikeParents iotago.BlockIDs `serix:"2,mapKey=shallowLikeParents"`
		// LatestFinalizedSlot is the index of the latest finalized slot.
		LatestFinalizedSlot iotago.SlotIndex `serix:"3,mapKey=latestFinalizedSlot"`
		// Commitment is the commitment of the block.
		Commitment iotago.Commitment `serix:"4,mapKey=commitment"`
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
		CreatedOutputs iotago.HexOutputIDs `serix:"1,mapKey=createdOutputs"`
		// The outputs that are consumed in this slot.
		ConsumedOutputs iotago.HexOutputIDs `serix:"2,mapKey=consumedOutputs"`
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

	// BlockIssuanceCreditsResponse defines the response for the block issuance credits REST API call.
	BlockIssuanceCreditsResponse struct {
		// SlotIndex is the index of the slot corresponding to the block issuance credits value returned.
		SlotIndex iotago.SlotIndex `serix:"0,mapKey=slotIndex"`
		// BlockIssuanceCredits is the block issuance credits value for the slot. Node is able to provide values only of already committed slots.
		BlockIssuanceCredits iotago.BlockIssuanceCredits `serix:"1,mapKey=blockIssuanceCredits"`
	}

	// ValidatorResponse defines the response used in stakers response REST API calls.
	ValidatorResponse struct {
		// AccountID is the hex encoded account ID of the validator.
		AccountID iotago.AccountID `serix:"0,mapKey=accountId"`
		// StakingEpochEnd is the epoch until which the validator registered to stake.
		StakingEpochEnd iotago.EpochIndex `serix:"1,mapKey=stakingEpochEnd"`
		// PoolStake is the sum of tokens delegated to the pooland the validator stake.
		PoolStake iotago.BaseToken `serix:"2,mapKey=poolStake"`
		// ValidatorStake is the stake of the validator.
		ValidatorStake iotago.BaseToken `serix:"3,mapKey=validatorStake"`
		// FixedCost is the fixed cost that the validator reciews from the total pool reward.
		FixedCost iotago.Mana `serix:"4,mapKey=fixedCost"`
		// LatestSuccessfulReward is the latest successful reward of the validator.
		LatestSupportedProtocolVersion iotago.Version `serix:"5,mapKey=latestSupportedProtocolVersion"`
	}

	// AccountStakingListResponse defines the response for the staking REST API call.
	AccountStakingListResponse struct {
		Stakers []ValidatorResponse `serix:"0,lengthPrefixType=uint8,mapKey=stakers"`
	}

	// ManaRewardsResponse defines the response for the mana rewards REST API call.
	ManaRewardsResponse struct {
		// EpochIndex is the epoch index for which the mana rewards are returned.
		EpochIndex iotago.EpochIndex `serix:"0,mapKey=epochIndex"`
		// The amount of totally available rewards the requested output may claim.
		Rewards iotago.Mana `serix:"1,mapKey=rewards"`
	}

	// CommitteeMemberResponse defines the response used in committee and staking response REST API calls.
	CommitteeMemberResponse struct {
		// AccountID is the hex encoded account ID of the validator.
		AccountID iotago.AccountID `serix:"0,mapKey=accountId"`
		// PoolStake is the sum of tokens delegated to the pooland the validator stake.
		PoolStake iotago.BaseToken `serix:"1,mapKey=poolStake"`
		// ValidatorStake is the stake of the validator.
		ValidatorStake iotago.BaseToken `serix:"2,mapKey=validatorStake"`
		// FixedCost is the fixed cost that the validator reciews from the total pool reward.
		FixedCost iotago.Mana `serix:"3,mapKey=fixedCost"`
	}

	// CommitteeResponse defines the response for the staking REST API call.
	CommitteeResponse struct {
		Committee           []CommitteeMemberResponse `serix:"0,lengthPrefixType=uint8,mapKey=committee"`
		TotalStake          iotago.BaseToken          `serix:"1,mapKey=totalStake"`
		TotalValidatorStake iotago.BaseToken          `serix:"2,mapKey=totalValidatorStake"`
		EpochIndex          iotago.EpochIndex         `serix:"3,mapKey=epochIndex"`
	}
)

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
