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
	BlockFailureNone                      BlockFailureReason = 0
	BlockFailureIsTooOld                  BlockFailureReason = 1
	BlockFailureParentIsTooOld            BlockFailureReason = 2
	BlockFailureParentNotFound            BlockFailureReason = 3
	BlockFailureParentInvalid             BlockFailureReason = 4
	BlockFailureIssuerAccountNotFound     BlockFailureReason = 5
	BlockFailureVersionInvalid            BlockFailureReason = 6
	BlockFailureManaCostCalculationFailed BlockFailureReason = 7
	BlockFailureBurnedInsufficientMana    BlockFailureReason = 8
	BlockFailureAccountInvalid            BlockFailureReason = 9
	BlockFailureSignatureInvalid          BlockFailureReason = 10
	BlockFailureDroppedDueToCongestion    BlockFailureReason = 11
	BlockFailurePayloadInvalid            BlockFailureReason = 12
	BlockFailureInvalid                   BlockFailureReason = 255

	// TODO: see if needed after congestion PR is done.
	BlockFailureOrphanedDueNegativeCreditsBalance BlockFailureReason = 13
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
	TxFailureNone                                     TransactionFailureReason = 0
	TxFailureUTXOInputAlreadySpent                    TransactionFailureReason = 1
	TxFailureConflicting                              TransactionFailureReason = 2
	TxFailureUTXOInputInvalid                         TransactionFailureReason = 3
	TxFailureTxTypeInvalid                            TransactionFailureReason = 4
	TxFailureSumOfInputAndOutputValuesDoesNotMatch    TransactionFailureReason = 5
	TxFailureUnlockBlockSignatureInvalid              TransactionFailureReason = 6
	TxFailureConfiguredTimelockNotYetExpired          TransactionFailureReason = 7
	TxFailureGivenNativeTokensInvalid                 TransactionFailureReason = 8
	TxFailureReturnAmountNotFulfilled                 TransactionFailureReason = 9
	TxFailureInputUnlockInvalid                       TransactionFailureReason = 10
	TxFailureSenderNotUnlocked                        TransactionFailureReason = 11
	TxFailureChainStateTransitionInvalid              TransactionFailureReason = 12
	TxFailureInputCreationAfterTxCreation             TransactionFailureReason = 13
	TxFailureManaAmountInvalid                        TransactionFailureReason = 14
	TxFailureBICInputInvalid                          TransactionFailureReason = 15
	TxFailureRewardInputInvalid                       TransactionFailureReason = 16
	TxFailureCommitmentInputInvalid                   TransactionFailureReason = 17
	TxFailureNoStakingFeature                         TransactionFailureReason = 18
	TxFailureFailedToClaimStakingReward               TransactionFailureReason = 19
	TxFailureFailedToClaimDelegationReward            TransactionFailureReason = 20
	TxFailureCapabilitiesNativeTokenBurningNotAllowed TransactionFailureReason = 21
	TxFailureCapabilitiesManaBurningNotAllowed        TransactionFailureReason = 22
	TxFailureCapabilitiesAccountDestructionNotAllowed TransactionFailureReason = 23
	TxFailureCapabilitiesAnchorDestructionNotAllowed  TransactionFailureReason = 24
	TxFailureCapabilitiesFoundryDestructionNotAllowed TransactionFailureReason = 25
	TxFailureCapabilitiesNFTDestructionNotAllowed     TransactionFailureReason = 26
	TxFailureSemanticValidationFailed                 TransactionFailureReason = 255
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
		Name string `serix:""`
		// The semver version of the node software.
		Version string `serix:""`
		// The current status of this node.
		Status *InfoResNodeStatus `serix:""`
		// The metrics of this node.
		Metrics *InfoResNodeMetrics `serix:""`
		// The protocol parameters used by this node.
		ProtocolParameters []*InfoResProtocolParameters `serix:""`
		// The base token of the network.
		BaseToken *InfoResBaseToken `serix:""`
		// The features this node exposes.
		Features []string `serix:""`
	}

	// InfoResProtocolParameters defines the protocol parameters of a node in the InfoResponse.
	InfoResProtocolParameters struct {
		StartEpoch iotago.EpochIndex         `serix:""`
		Parameters iotago.ProtocolParameters `serix:""`
	}

	// InfoResNodeStatus defines the status of the node in the InfoResponse.
	InfoResNodeStatus struct {
		// Whether the node is healthy.
		IsHealthy bool `serix:""`
		// The accepted tangle time.
		AcceptedTangleTime time.Time `serix:""`
		// The relative accepted tangle time.
		RelativeAcceptedTangleTime time.Time `serix:""`
		// The confirmed tangle time.
		ConfirmedTangleTime time.Time `serix:""`
		// The relative confirmed tangle time.
		RelativeConfirmedTangleTime time.Time `serix:""`
		// The id of the latest known commitment.
		LatestCommitmentID iotago.CommitmentID `serix:""`
		// The latest finalized slot.
		LatestFinalizedSlot iotago.SlotIndex `serix:""`
		// The slot of the latest accepted block.
		LatestAcceptedBlockSlot iotago.SlotIndex `serix:""`
		// The slot of the latest confirmed block.
		LatestConfirmedBlockSlot iotago.SlotIndex `serix:""`
		// The epoch at which the tangle data was pruned.
		PruningEpoch iotago.EpochIndex `serix:""`
	}

	// InfoResNodeMetrics defines the metrics of a node in the InfoResponse.
	InfoResNodeMetrics struct {
		// The current rate of new blocks per second, it's updated when a commitment is committed.
		BlocksPerSecond float64 `serix:""`
		// The current rate of confirmed blocks per second, it's updated when a commitment is committed.
		ConfirmedBlocksPerSecond float64 `serix:""`
		// The ratio of confirmed blocks in relation to new blocks up until the latest commitment is committed.
		ConfirmationRate float64 `serix:""`
	}

	// InfoResBaseToken defines the base token of the node in the InfoResponse.
	InfoResBaseToken struct {
		// The base token name.
		Name string `serix:""`
		// The base token ticker symbol.
		TickerSymbol string `serix:""`
		// The base token unit.
		Unit string `serix:""`
		// The base token subunit.
		Subunit string `serix:",omitempty"`
		// The base token amount of decimals.
		Decimals uint32 `serix:""`
	}

	// IssuanceBlockHeaderResponse defines the response of a GET block issuance REST API call.
	IssuanceBlockHeaderResponse struct {
		// StrongParents are the strong parents of the block.
		StrongParents iotago.BlockIDs `serix:",lenPrefix=uint8"`
		// WeakParents are the weak parents of the block.
		WeakParents iotago.BlockIDs `serix:",lenPrefix=uint8"`
		// ShallowLikeParents are the shallow like parents of the block.
		ShallowLikeParents iotago.BlockIDs `serix:",lenPrefix=uint8"`
		// LatestFinalizedSlot is the latest finalized slot.
		LatestFinalizedSlot iotago.SlotIndex `serix:""`
		// Commitment is the latest commitment of the node.
		Commitment *iotago.Commitment `serix:""`
	}

	// BlockCreatedResponse defines the response of a POST blocks REST API call.
	BlockCreatedResponse struct {
		// The hex encoded block ID of the block.
		BlockID iotago.BlockID `serix:""`
	}

	// BlockMetadataResponse defines the response of a GET block metadata REST API call.
	BlockMetadataResponse struct {
		// BlockID The hex encoded block ID of the block.
		BlockID iotago.BlockID `serix:""`
		// BlockState might be pending, rejected, failed, confirmed, finalized.
		BlockState string `serix:""`
		// BlockFailureReason if applicable indicates the error that occurred during the block processing.
		BlockFailureReason BlockFailureReason `serix:",omitempty"`
		// TransactionState might be pending, conflicting, confirmed, finalized, rejected.
		TransactionState string `serix:",omitempty"`
		// TransactionFailureReason if applicable indicates the error that occurred during the transaction processing.
		TransactionFailureReason TransactionFailureReason `serix:",omitempty"`
	}

	// BlockWithMetadataResponse defines the response of a GET full block REST API call.
	BlockWithMetadataResponse struct {
		Block    *iotago.Block          `serix:""`
		Metadata *BlockMetadataResponse `serix:""`
	}

	// OutputResponse defines the response of a GET outputs REST API call.
	OutputResponse struct {
		Output        iotago.TxEssenceOutput `serix:""`
		OutputIDProof *iotago.OutputIDProof  `serix:""`
	}

	// OutputMetadata defines the response of a GET outputs metadata REST API call.
	OutputMetadata struct {
		// BlockID is the block ID that contains the output.
		BlockID iotago.BlockID `serix:""`
		// TransactionID is the transaction ID that creates the output.
		TransactionID iotago.TransactionID `serix:""`
		// OutputIndex is the index of the output.
		OutputIndex uint16 `serix:""`
		// IncludedCommitmentID is the commitment ID that includes the output.
		IncludedCommitmentID iotago.CommitmentID `serix:",omitempty"`
		// IsSpent indicates whether the output is spent or not.
		IsSpent bool `serix:""`
		// CommitmentIDSpent is the commitment ID that includes the spent output.
		CommitmentIDSpent iotago.CommitmentID `serix:",omitempty"`
		// TransactionIDSpent is the transaction ID that spends the output.
		TransactionIDSpent iotago.TransactionID `serix:",omitempty"`
		// LatestCommitmentID is the latest commitment ID of a node.
		LatestCommitmentID iotago.CommitmentID `serix:""`
	}

	// OutputWithMetadataResponse defines the response of a GET full outputs REST API call.
	OutputWithMetadataResponse struct {
		Output        iotago.TxEssenceOutput `serix:""`
		OutputIDProof *iotago.OutputIDProof  `serix:""`
		Metadata      *OutputMetadata        `serix:""`
	}

	// UTXOChangesResponse defines the response for UTXO slot REST API call.
	UTXOChangesResponse struct {
		// CommitmentID is the commitment ID of the requested slot that contains the changes.
		CommitmentID iotago.CommitmentID `serix:""`
		// The outputs that are created in this slot.
		CreatedOutputs iotago.OutputIDs `serix:""`
		// The outputs that are consumed in this slot.
		ConsumedOutputs iotago.OutputIDs `serix:""`
	}

	// CongestionResponse defines the response for the congestion REST API call.
	CongestionResponse struct {
		// Slot is the slot for which the estimate is provided
		Slot iotago.SlotIndex `serix:""`
		// Ready indicates if a node is ready to issue a block in a current congestion or should wait.
		Ready bool `serix:""`
		// ReferenceManaCost (RMC) is the mana cost a user needs to burn to issue a block in Slot slot.
		ReferenceManaCost iotago.Mana `serix:""`
		// BlockIssuanceCredits (BIC) is the mana a user has on its BIC account exactly slot - MaxCommittableASge in the past.
		// This balance needs to be > 0 zero, otherwise account is locked
		BlockIssuanceCredits iotago.BlockIssuanceCredits `serix:""`
	}

	// ValidatorResponse defines the response used in stakers response REST API calls.
	ValidatorResponse struct {
		// AddressBech32 is the account address of the validator.
		AddressBech32 string `serix:"address,lenPrefix=uint8"`
		// StakingEpochEnd is the epoch until which the validator registered to stake.
		StakingEpochEnd iotago.EpochIndex `serix:""`
		// PoolStake is the sum of tokens delegated to the pool and the validator stake.
		PoolStake iotago.BaseToken `serix:""`
		// ValidatorStake is the stake of the validator.
		ValidatorStake iotago.BaseToken `serix:""`
		// FixedCost is the fixed cost that the validator receives from the total pool reward.
		FixedCost iotago.Mana `serix:""`
		// Active indicates whether the validator was active recently, and would be considered during committee selection.
		Active bool `serix:""`
		// LatestSupportedProtocolVersion is the latest supported protocol version of the validator.
		LatestSupportedProtocolVersion iotago.Version `serix:""`
		// LatestSupportedProtocolHash is the protocol hash of the latest supported protocol of the validator.
		LatestSupportedProtocolHash iotago.Identifier `serix:""`
	}

	// ValidatorsResponse defines the response for the staking REST API call.
	ValidatorsResponse struct {
		Validators []*ValidatorResponse `serix:"stakers"`
		PageSize   uint32               `serix:""`
		Cursor     string               `serix:",omitempty"`
	}

	// ManaRewardsResponse defines the response for the mana rewards REST API call.
	ManaRewardsResponse struct {
		// EpochStart is the starting epoch for the range for which the mana rewards are returned.
		EpochStart iotago.EpochIndex `serix:""`
		// EpochEnd is the ending epoch for the range for which the mana rewards are returned, also the decay is only applied up to this point.
		EpochEnd iotago.EpochIndex `serix:""`
		// The amount of totally available rewards the requested output may claim, decayed up to EpochEnd (including).
		Rewards iotago.Mana `serix:""`
	}

	// CommitteeMemberResponse defines the response used in committee and staking response REST API calls.
	CommitteeMemberResponse struct {
		// AddressBech32 is the account address of the validator.
		AddressBech32 string `serix:"address,lenPrefix=uint8"`
		// PoolStake is the sum of tokens delegated to the pool and the validator stake.
		PoolStake iotago.BaseToken `serix:""`
		// ValidatorStake is the stake of the validator.
		ValidatorStake iotago.BaseToken `serix:""`
		// FixedCost is the fixed cost that the validator received from the total pool reward.
		FixedCost iotago.Mana `serix:""`
	}

	// CommitteeResponse defines the response for the staking REST API call.
	CommitteeResponse struct {
		Committee           []*CommitteeMemberResponse `serix:""`
		TotalStake          iotago.BaseToken           `serix:""`
		TotalValidatorStake iotago.BaseToken           `serix:""`
		Epoch               iotago.EpochIndex          `serix:""`
	}
)
