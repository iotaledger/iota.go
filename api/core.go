package api

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

func (b BlockState) EncodeJSON() (any, error) {
	if b > BlockStateFailed {
		return nil, ierrors.Errorf("invalid block state: %d", b)
	}

	return b.String(), nil
}

func (b *BlockState) DecodeJSON(state any) error {
	if state == nil {
		return ierrors.New("given block state is nil")
	}

	blockState, ok := state.(string)
	if !ok {
		return ierrors.Errorf("invalid type: %T", state)
	}

	switch blockState {
	case "unknown":
		*b = BlockStateUnknown
	case "pending":
		*b = BlockStatePending
	case "accepted":
		*b = BlockStateAccepted
	case "confirmed":
		*b = BlockStateConfirmed
	case "finalized":
		*b = BlockStateFinalized
	case "rejected":
		*b = BlockStateRejected
	case "failed":
		*b = BlockStateFailed
	default:
		return ierrors.Errorf("invalid block state: %s", blockState)
	}

	return nil
}

const (
	BlockFailureNone                      BlockFailureReason = 0
	BlockFailureIsTooOld                  BlockFailureReason = 1
	BlockFailureParentIsTooOld            BlockFailureReason = 2
	BlockFailureParentNotFound            BlockFailureReason = 3
	BlockFailureIssuerAccountNotFound     BlockFailureReason = 4
	BlockFailureManaCostCalculationFailed BlockFailureReason = 5
	BlockFailureBurnedInsufficientMana    BlockFailureReason = 6
	BlockFailureAccountLocked             BlockFailureReason = 7
	BlockFailureAccountExpired            BlockFailureReason = 8
	BlockFailureSignatureInvalid          BlockFailureReason = 9
	BlockFailureDroppedDueToCongestion    BlockFailureReason = 10
	BlockFailurePayloadInvalid            BlockFailureReason = 11
	BlockFailureInvalid                   BlockFailureReason = 255
)

var blocksErrorsFailureReasonMap = map[error]BlockFailureReason{
	iotago.ErrIssuerAccountNotFound:     BlockFailureIssuerAccountNotFound,
	iotago.ErrBurnedInsufficientMana:    BlockFailureBurnedInsufficientMana,
	iotago.ErrFailedToCalculateManaCost: BlockFailureManaCostCalculationFailed,
	iotago.ErrAccountLocked:             BlockFailureAccountLocked,
	iotago.ErrAccountExpired:            BlockFailureAccountExpired,
	iotago.ErrInvalidSignature:          BlockFailureSignatureInvalid,
}

func (t BlockFailureReason) Bytes() ([]byte, error) {
	return []byte{byte(t)}, nil
}

func BlockFailureReasonFromBytes(b []byte) (BlockFailureReason, int, error) {
	if len(b) < BlockFailureReasonLength {
		return 0, 0, ierrors.New("invalid block failure reason size")
	}

	return BlockFailureReason(b[0]), BlockFailureReasonLength, nil
}

func DetermineBlockFailureReason(err error) BlockFailureReason {
	errorList := make([]error, 0)
	errorList = unwrapErrors(err, errorList)

	// Map the error to the block failure reason.
	// The strategy is to map the first failure reason that exists in order of most-detailed to least-detailed error.
	for _, err := range errorList {
		if blockFailureReason, matches := blocksErrorsFailureReasonMap[err]; matches {
			return blockFailureReason
		}
	}

	// Use most general failure reason if no other error matches.
	return BlockFailureInvalid
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

func (t TransactionState) EncodeJSON() (any, error) {
	if t > TransactionStateFailed {
		return nil, ierrors.Errorf("invalid transaction state: %d", t)
	}

	return t.String(), nil
}

func (t *TransactionState) DecodeJSON(state any) error {
	if state == nil {
		return ierrors.New("given transaction state is nil")
	}

	transactionState, ok := state.(string)
	if !ok {
		return ierrors.Errorf("invalid type: %T", state)
	}

	switch transactionState {
	case "noTransaction":
		*t = TransactionStateNoTransaction
	case "pending":
		*t = TransactionStatePending
	case "accepted":
		*t = TransactionStateAccepted
	case "confirmed":
		*t = TransactionStateConfirmed
	case "finalized":
		*t = TransactionStateFinalized
	case "failed":
		*t = TransactionStateFailed
	default:
		return ierrors.Errorf("invalid transaction state: %s", transactionState)
	}

	return nil
}

const (
	TxFailureNone TransactionFailureReason = 0

	TxFailureConflictRejected TransactionFailureReason = 1

	TxFailureInputAlreadySpent            TransactionFailureReason = 2
	TxFailureInputCreationAfterTxCreation TransactionFailureReason = 3
	TxFailureUnlockSignatureInvalid       TransactionFailureReason = 4

	TxFailureChainAddressUnlockInvalid            TransactionFailureReason = 5
	TxFailureDirectUnlockableAddressUnlockInvalid TransactionFailureReason = 6
	TxFailureMultiAddressUnlockInvalid            TransactionFailureReason = 7

	TxFailureCommitmentInputReferenceInvalid TransactionFailureReason = 8
	TxFailureBICInputReferenceInvalid        TransactionFailureReason = 9
	TxFailureRewardInputReferenceInvalid     TransactionFailureReason = 10

	TxFailureStakingRewardCalculationFailure    TransactionFailureReason = 11
	TxFailureDelegationRewardCalculationFailure TransactionFailureReason = 12

	TxFailureInputOutputBaseTokenMismatch TransactionFailureReason = 13

	TxFailureManaOverflow                             TransactionFailureReason = 14
	TxFailureInputOutputManaMismatch                  TransactionFailureReason = 15
	TxFailureManaDecayCreationIndexExceedsTargetIndex TransactionFailureReason = 16

	TxFailureNativeTokenSumUnbalanced TransactionFailureReason = 17

	TxFailureSimpleTokenSchemeMintedMeltedTokenDecrease TransactionFailureReason = 18
	TxFailureSimpleTokenSchemeMintingInvalid            TransactionFailureReason = 19
	TxFailureSimpleTokenSchemeMeltingInvalid            TransactionFailureReason = 20
	TxFailureSimpleTokenSchemeMaximumSupplyChanged      TransactionFailureReason = 21
	TxFailureSimpleTokenSchemeGenesisInvalid            TransactionFailureReason = 22

	TxFailureMultiAddressLengthUnlockLengthMismatch TransactionFailureReason = 23
	TxFailureMultiAddressUnlockThresholdNotReached  TransactionFailureReason = 24

	TxFailureSenderFeatureNotUnlocked TransactionFailureReason = 25

	TxFailureIssuerFeatureNotUnlocked TransactionFailureReason = 26

	TxFailureStakingRewardInputMissing             TransactionFailureReason = 27
	TxFailureStakingBlockIssuerFeatureMissing      TransactionFailureReason = 28
	TxFailureStakingCommitmentInputMissing         TransactionFailureReason = 29
	TxFailureStakingRewardClaimingInvalid          TransactionFailureReason = 30
	TxFailureStakingFeatureRemovedBeforeUnbonding  TransactionFailureReason = 31
	TxFailureStakingFeatureModifiedBeforeUnbonding TransactionFailureReason = 32
	TxFailureStakingStartEpochInvalid              TransactionFailureReason = 33
	TxFailureStakingEndEpochTooEarly               TransactionFailureReason = 34

	TxFailureBlockIssuerCommitmentInputMissing TransactionFailureReason = 35
	TxFailureBlockIssuanceCreditInputMissing   TransactionFailureReason = 36
	TxFailureBlockIssuerNotExpired             TransactionFailureReason = 37
	TxFailureBlockIssuerExpiryTooEarly         TransactionFailureReason = 38
	TxFailureManaMovedOffBlockIssuerAccount    TransactionFailureReason = 39
	TxFailureAccountLocked                     TransactionFailureReason = 40

	TxFailureTimelockCommitmentInputMissing TransactionFailureReason = 41
	TxFailureTimelockNotExpired             TransactionFailureReason = 42

	TxFailureExpirationCommitmentInputMissing TransactionFailureReason = 43
	TxFailureExpirationNotUnlockable          TransactionFailureReason = 44

	TxFailureReturnAmountNotFulFilled TransactionFailureReason = 45

	TxFailureNewChainOutputHasNonZeroedID        TransactionFailureReason = 46
	TxFailureChainOutputImmutableFeaturesChanged TransactionFailureReason = 47

	TxFailureImplicitAccountDestructionDisallowed     TransactionFailureReason = 48
	TxFailureMultipleImplicitAccountCreationAddresses TransactionFailureReason = 49

	TxFailureAccountInvalidFoundryCounter TransactionFailureReason = 50

	TxFailureAnchorInvalidStateTransition      TransactionFailureReason = 51
	TxFailureAnchorInvalidGovernanceTransition TransactionFailureReason = 52

	TxFailureFoundryTransitionWithoutAccount TransactionFailureReason = 53
	TxFailureFoundrySerialInvalid            TransactionFailureReason = 54

	TxFailureDelegationCommitmentInputMissing  TransactionFailureReason = 55
	TxFailureDelegationRewardInputMissing      TransactionFailureReason = 56
	TxFailureDelegationRewardsClaimingInvalid  TransactionFailureReason = 57
	TxFailureDelegationOutputTransitionedTwice TransactionFailureReason = 58
	TxFailureDelegationModified                TransactionFailureReason = 59
	TxFailureDelegationStartEpochInvalid       TransactionFailureReason = 60
	TxFailureDelegationAmountMismatch          TransactionFailureReason = 61
	TxFailureDelegationEndEpochNotZero         TransactionFailureReason = 62
	TxFailureDelegationEndEpochInvalid         TransactionFailureReason = 63

	TxFailureCapabilitiesNativeTokenBurningNotAllowed TransactionFailureReason = 64
	TxFailureCapabilitiesManaBurningNotAllowed        TransactionFailureReason = 65
	TxFailureCapabilitiesAccountDestructionNotAllowed TransactionFailureReason = 66
	TxFailureCapabilitiesAnchorDestructionNotAllowed  TransactionFailureReason = 67
	TxFailureCapabilitiesFoundryDestructionNotAllowed TransactionFailureReason = 68
	TxFailureCapabilitiesNFTDestructionNotAllowed     TransactionFailureReason = 69

	TxFailureSemanticValidationFailed TransactionFailureReason = 255
)

var txErrorsFailureReasonMap = map[error]TransactionFailureReason{
	// ================================
	// Pre-Transaction Execution Errors
	// ================================

	// tx level errors
	iotago.ErrTxConflictRejected: TxFailureConflictRejected,

	// input
	iotago.ErrInputAlreadySpent:            TxFailureInputAlreadySpent,
	iotago.ErrInputCreationAfterTxCreation: TxFailureInputCreationAfterTxCreation,
	iotago.ErrUnlockSignatureInvalid:       TxFailureUnlockSignatureInvalid,

	// unlock
	iotago.ErrChainAddressUnlockInvalid:            TxFailureChainAddressUnlockInvalid,
	iotago.ErrDirectUnlockableAddressUnlockInvalid: TxFailureDirectUnlockableAddressUnlockInvalid,
	iotago.ErrMultiAddressUnlockInvalid:            TxFailureMultiAddressUnlockInvalid,

	// context inputs
	iotago.ErrCommitmentInputReferenceInvalid: TxFailureCommitmentInputReferenceInvalid,
	iotago.ErrBICInputReferenceInvalid:        TxFailureBICInputReferenceInvalid,
	iotago.ErrRewardInputReferenceInvalid:     TxFailureRewardInputReferenceInvalid,

	// reward calculation
	iotago.ErrStakingRewardCalculationFailure:    TxFailureStakingRewardCalculationFailure,
	iotago.ErrDelegationRewardCalculationFailure: TxFailureDelegationRewardCalculationFailure,

	// ============================
	// Transaction Execution Errors
	// ============================

	// amount
	iotago.ErrInputOutputBaseTokenMismatch: TxFailureInputOutputBaseTokenMismatch,

	// mana
	iotago.ErrManaOverflow:                             TxFailureManaOverflow,
	iotago.ErrInputOutputManaMismatch:                  TxFailureInputOutputManaMismatch,
	iotago.ErrManaDecayCreationIndexExceedsTargetIndex: TxFailureManaDecayCreationIndexExceedsTargetIndex,

	// native token
	iotago.ErrNativeTokenSumUnbalanced: TxFailureNativeTokenSumUnbalanced,

	// simple token scheme
	iotago.ErrSimpleTokenSchemeMintedMeltedTokenDecrease: TxFailureSimpleTokenSchemeMintedMeltedTokenDecrease,
	iotago.ErrSimpleTokenSchemeMintingInvalid:            TxFailureSimpleTokenSchemeMintingInvalid,
	iotago.ErrSimpleTokenSchemeMeltingInvalid:            TxFailureSimpleTokenSchemeMeltingInvalid,
	iotago.ErrSimpleTokenSchemeMaximumSupplyChanged:      TxFailureSimpleTokenSchemeMaximumSupplyChanged,
	iotago.ErrSimpleTokenSchemeGenesisInvalid:            TxFailureSimpleTokenSchemeGenesisInvalid,

	// multi address
	iotago.ErrMultiAddressLengthUnlockLengthMismatch: TxFailureMultiAddressLengthUnlockLengthMismatch,
	iotago.ErrMultiAddressUnlockThresholdNotReached:  TxFailureMultiAddressUnlockThresholdNotReached,

	// sender feature
	iotago.ErrSenderFeatureNotUnlocked: TxFailureSenderFeatureNotUnlocked,

	// issuer feature
	iotago.ErrIssuerFeatureNotUnlocked: TxFailureIssuerFeatureNotUnlocked,

	// staking feature
	iotago.ErrStakingRewardInputMissing:             TxFailureStakingRewardInputMissing,
	iotago.ErrStakingBlockIssuerFeatureMissing:      TxFailureStakingBlockIssuerFeatureMissing,
	iotago.ErrStakingCommitmentInputMissing:         TxFailureStakingCommitmentInputMissing,
	iotago.ErrStakingRewardClaimingInvalid:          TxFailureStakingRewardClaimingInvalid,
	iotago.ErrStakingFeatureRemovedBeforeUnbonding:  TxFailureStakingFeatureRemovedBeforeUnbonding,
	iotago.ErrStakingFeatureModifiedBeforeUnbonding: TxFailureStakingFeatureModifiedBeforeUnbonding,
	iotago.ErrStakingStartEpochInvalid:              TxFailureStakingStartEpochInvalid,
	iotago.ErrStakingEndEpochTooEarly:               TxFailureStakingEndEpochTooEarly,

	// block issuer feature
	iotago.ErrBlockIssuerCommitmentInputMissing: TxFailureBlockIssuerCommitmentInputMissing,
	iotago.ErrBlockIssuanceCreditInputMissing:   TxFailureBlockIssuanceCreditInputMissing,
	iotago.ErrBlockIssuerNotExpired:             TxFailureBlockIssuerNotExpired,
	iotago.ErrBlockIssuerExpiryTooEarly:         TxFailureBlockIssuerExpiryTooEarly,
	iotago.ErrManaMovedOffBlockIssuerAccount:    TxFailureManaMovedOffBlockIssuerAccount,
	iotago.ErrAccountLocked:                     TxFailureAccountLocked,

	// timelock unlock condition
	iotago.ErrTimelockCommitmentInputMissing: TxFailureTimelockCommitmentInputMissing,
	iotago.ErrTimelockNotExpired:             TxFailureTimelockNotExpired,

	// expiration unlock condition
	iotago.ErrExpirationCommitmentInputMissing: TxFailureExpirationCommitmentInputMissing,
	iotago.ErrExpirationNotUnlockable:          TxFailureExpirationNotUnlockable,

	// storage deposit return unlock condition
	iotago.ErrReturnAmountNotFulFilled: TxFailureReturnAmountNotFulFilled,

	// generic chain output errors
	iotago.ErrNewChainOutputHasNonZeroedID:        TxFailureNewChainOutputHasNonZeroedID,
	iotago.ErrChainOutputImmutableFeaturesChanged: TxFailureChainOutputImmutableFeaturesChanged,

	// implicit account
	iotago.ErrImplicitAccountDestructionDisallowed:     TxFailureImplicitAccountDestructionDisallowed,
	iotago.ErrMultipleImplicitAccountCreationAddresses: TxFailureMultipleImplicitAccountCreationAddresses,

	// account
	iotago.ErrAccountInvalidFoundryCounter: TxFailureAccountInvalidFoundryCounter,

	iotago.ErrAnchorInvalidStateTransition:      TxFailureAnchorInvalidStateTransition,
	iotago.ErrAnchorInvalidGovernanceTransition: TxFailureAnchorInvalidGovernanceTransition,

	// foundry
	iotago.ErrFoundryTransitionWithoutAccount: TxFailureFoundryTransitionWithoutAccount,
	iotago.ErrFoundrySerialInvalid:            TxFailureFoundrySerialInvalid,

	// delegation
	iotago.ErrDelegationCommitmentInputMissing:  TxFailureDelegationCommitmentInputMissing,
	iotago.ErrDelegationRewardInputMissing:      TxFailureDelegationRewardInputMissing,
	iotago.ErrDelegationRewardsClaimingInvalid:  TxFailureDelegationRewardsClaimingInvalid,
	iotago.ErrDelegationOutputTransitionedTwice: TxFailureDelegationOutputTransitionedTwice,
	iotago.ErrDelegationModified:                TxFailureDelegationModified,
	iotago.ErrDelegationStartEpochInvalid:       TxFailureDelegationStartEpochInvalid,
	iotago.ErrDelegationAmountMismatch:          TxFailureDelegationAmountMismatch,
	iotago.ErrDelegationEndEpochNotZero:         TxFailureDelegationEndEpochNotZero,
	iotago.ErrDelegationEndEpochInvalid:         TxFailureDelegationEndEpochInvalid,

	// tx capabilities
	iotago.ErrTxCapabilitiesNativeTokenBurningNotAllowed: TxFailureCapabilitiesNativeTokenBurningNotAllowed,
	iotago.ErrTxCapabilitiesManaBurningNotAllowed:        TxFailureCapabilitiesManaBurningNotAllowed,
	iotago.ErrTxCapabilitiesAccountDestructionNotAllowed: TxFailureCapabilitiesAccountDestructionNotAllowed,
	iotago.ErrTxCapabilitiesAnchorDestructionNotAllowed:  TxFailureCapabilitiesAnchorDestructionNotAllowed,
	iotago.ErrTxCapabilitiesFoundryDestructionNotAllowed: TxFailureCapabilitiesFoundryDestructionNotAllowed,
	iotago.ErrTxCapabilitiesNFTDestructionNotAllowed:     TxFailureCapabilitiesNFTDestructionNotAllowed,
}

func (t TransactionFailureReason) Bytes() ([]byte, error) {
	return []byte{byte(t)}, nil
}

func TransactionFailureReasonFromBytes(b []byte) (TransactionFailureReason, int, error) {
	if len(b) < TransactionFailureReasonLength {
		return 0, 0, ierrors.New("invalid transaction failure reason size")
	}

	return TransactionFailureReason(b[0]), TransactionFailureReasonLength, nil
}

// Unwraps the given err into the given errList by recursively unwrapping it.
//
// In case of joined errors, the right-most error is unwrapped first, which corresponds
// to a post-order depth-traversal of err's tree.
//
// This means errList will contain the most-detailed errors first (those lower in the error tree).
func unwrapErrors(err error, errList []error) []error {
	//nolint:errorlint // false positive: we're not switching on a specific error type.
	switch x := err.(type) {
	case interface{ Unwrap() []error }:
		errors := x.Unwrap()
		// Iterate the errors in reverse, so we walk the tree in post-order.
		for i := len(errors) - 1; i >= 0; i-- {
			err := errors[i]
			if err != nil {
				errList = unwrapErrors(err, errList)
				errList = append(errList, err)
			}
		}
	case interface{ Unwrap() error }:
		err = x.Unwrap()
		if err != nil {
			errList = unwrapErrors(err, errList)
			errList = append(errList, err)
		}
	}

	return errList
}

func DetermineTransactionFailureReason(err error) TransactionFailureReason {
	errorList := make([]error, 0)
	errorList = unwrapErrors(err, errorList)

	// Map the error to the transaction failure reason.
	// The strategy is to map the first failure reason that exists in order of most-detailed to least-detailed error.
	for _, err := range errorList {
		if txFailureReason, matches := txErrorsFailureReasonMap[err]; matches {
			return txFailureReason
		}
	}

	// Use most general failure reason if no other error matches.
	return TxFailureSemanticValidationFailed
}

type (
	// InfoResponse defines the response of a GET info REST API call.
	InfoResponse struct {
		// The name of the node software.
		Name string `serix:",lenPrefix=uint8"`
		// The semver version of the node software.
		Version string `serix:",lenPrefix=uint8"`
		// The current status of this node.
		Status *InfoResNodeStatus `serix:""`
		// The metrics of this node.
		Metrics *InfoResNodeMetrics `serix:""`
		// The protocol parameters used by this node.
		ProtocolParameters []*InfoResProtocolParameters `serix:",lenPrefix=uint8"`
		// The base token of the network.
		BaseToken *InfoResBaseToken `serix:""`
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
		Name string `serix:",lenPrefix=uint8"`
		// The base token ticker symbol.
		TickerSymbol string `serix:",lenPrefix=uint8"`
		// The base token unit.
		Unit string `serix:",lenPrefix=uint8"`
		// The base token subunit.
		Subunit string `serix:",lenPrefix=uint8,omitempty"`
		// The base token amount of decimals.
		Decimals uint32 `serix:""`
	}

	// IssuanceBlockHeaderResponse defines the response of a GET block issuance REST API call.
	IssuanceBlockHeaderResponse struct {
		// StrongParents are the strong parents of the block.
		StrongParents iotago.BlockIDs `serix:",lenPrefix=uint8"`
		// WeakParents are the weak parents of the block.
		WeakParents iotago.BlockIDs `serix:",lenPrefix=uint8,omitempty"`
		// ShallowLikeParents are the shallow like parents of the block.
		ShallowLikeParents iotago.BlockIDs `serix:",lenPrefix=uint8,omitempty"`
		// LatestParentBlockIssuingTime is the latest issuing time of the returned parents.
		LatestParentBlockIssuingTime time.Time `serix:""`
		// LatestFinalizedSlot is the latest finalized slot.
		LatestFinalizedSlot iotago.SlotIndex `serix:""`
		// LatestCommitment is the latest commitment of the node.
		LatestCommitment *iotago.Commitment `serix:""`
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
		BlockState BlockState `serix:""`
		// BlockFailureReason if applicable indicates the error that occurred during the block processing.
		BlockFailureReason BlockFailureReason `serix:",omitempty"`
		// TransactionMetadata is the metadata of the transaction that is contained in the block.
		TransactionMetadata *TransactionMetadataResponse `serix:",optional,omitempty"`
	}

	// BlockWithMetadataResponse defines the response of a GET full block REST API call.
	BlockWithMetadataResponse struct {
		Block    *iotago.Block          `serix:""`
		Metadata *BlockMetadataResponse `serix:""`
	}

	// TransactionMetadataResponse defines the response of a GET transaction metadata REST API call.
	TransactionMetadataResponse struct {
		// TransactionID is the hex encoded transaction ID of the transaction.
		TransactionID iotago.TransactionID `serix:""`
		// TransactionState might be pending, conflicting, confirmed, finalized, rejected.
		TransactionState TransactionState `serix:""`
		// TransactionFailureReason if applicable indicates the error that occurred during the transaction processing.
		TransactionFailureReason TransactionFailureReason `serix:",omitempty"`
	}

	// OutputResponse defines the response of a GET outputs REST API call.
	OutputResponse struct {
		Output        iotago.TxEssenceOutput `serix:""`
		OutputIDProof *iotago.OutputIDProof  `serix:""`
	}

	// OutputWithID returns an output with its corresponding ID.
	OutputWithID struct {
		OutputID iotago.OutputID        `serix:""`
		Output   iotago.TxEssenceOutput `serix:""`
	}

	OutputInclusionMetadata struct {
		// Slot is the slot in which the output was included.
		Slot iotago.SlotIndex `serix:""`
		// TransactionID is the transaction ID that created the output.
		TransactionID iotago.TransactionID `serix:""`
		// CommitmentID is the commitment ID that includes the creation of the output.
		CommitmentID iotago.CommitmentID `serix:",omitempty"`
	}

	OutputConsumptionMetadata struct {
		// Slot is the slot in which the output was spent.
		Slot iotago.SlotIndex `serix:""`
		// TransactionID is the transaction ID that spent the output.
		TransactionID iotago.TransactionID `serix:""`
		// CommitmentID is the commitment ID that includes the spending of the output.
		CommitmentID iotago.CommitmentID `serix:",omitempty"`
	}

	// OutputMetadata defines the response of a GET outputs metadata REST API call.
	OutputMetadata struct {
		// OutputID is the hex encoded output ID.
		OutputID iotago.OutputID `serix:""`
		// BlockID is the block ID that contains the output.
		BlockID iotago.BlockID `serix:""`
		// Included is the metadata of the output if it is included in the ledger.
		Included *OutputInclusionMetadata `serix:""`
		// Spent is the metadata of the output if it is marked as spent in the ledger.
		Spent *OutputConsumptionMetadata `serix:",optional,omitempty"`
		// LatestCommitmentID is the latest commitment ID of a node.
		LatestCommitmentID iotago.CommitmentID `serix:""`
	}

	// OutputWithMetadataResponse defines the response of a GET full outputs REST API call.
	OutputWithMetadataResponse struct {
		Output        iotago.TxEssenceOutput `serix:""`
		OutputIDProof *iotago.OutputIDProof  `serix:""`
		Metadata      *OutputMetadata        `serix:""`
	}

	// UTXOChangesResponse defines the response for the UTXO changes per slot REST API call.
	UTXOChangesResponse struct {
		// CommitmentID is the commitment ID of the requested slot that contains the changes.
		CommitmentID iotago.CommitmentID `serix:""`
		// The outputs that are created in this slot.
		CreatedOutputs iotago.OutputIDs `serix:",lenPrefix=uint32"`
		// The outputs that are consumed in this slot.
		ConsumedOutputs iotago.OutputIDs `serix:",lenPrefix=uint32"`
	}

	// UTXOChangesFullResponse defines the response for the UTXO changes per slot REST API call.
	// It returns the full information about the outputs with their corresponding ID.
	UTXOChangesFullResponse struct {
		// CommitmentID is the commitment ID of the requested slot that contains the changes.
		CommitmentID iotago.CommitmentID `serix:""`
		// The outputs that are created in this slot.
		CreatedOutputs []*OutputWithID `serix:",lenPrefix=uint32"`
		// The outputs that are consumed in this slot.
		ConsumedOutputs []*OutputWithID `serix:",lenPrefix=uint32"`
	}

	// CongestionResponse defines the response for the congestion REST API call.
	CongestionResponse struct {
		// Slot is the slot for which the estimate is provided.
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
		// StakingEndEpoch is the epoch until which the validator registered to stake.
		StakingEndEpoch iotago.EpochIndex `serix:""`
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
		Validators []*ValidatorResponse `serix:"stakers,lenPrefix=uint16"`
		PageSize   uint32               `serix:""`
		Cursor     string               `serix:",lenPrefix=uint8,omitempty"`
	}

	// ManaRewardsResponse defines the response for the mana rewards REST API call.
	ManaRewardsResponse struct {
		// StartEpoch is the first epoch for which rewards can be claimed.
		// This value is useful for checking if rewards have expired (by comparing against the staking or delegation start)
		// or would expire soon (by checking its relation to the rewards retention period).
		StartEpoch iotago.EpochIndex `serix:""`
		// EndEpoch is the last epoch for which rewards can be claimed.
		EndEpoch iotago.EpochIndex `serix:""`
		// The amount of totally available decayed rewards the requested output may claim.
		Rewards iotago.Mana `serix:""`
		// The rewards of the latest committed epoch of the staking pool to which this validator or delegator belongs.
		// The ratio of this value and the maximally possible rewards for the latest committed epoch can be used to determine
		// how well the validator of this staking pool performed in that epoch.
		// Note that if the pool was not part of the committee in the latest committed epoch, this value is 0.
		LatestCommittedEpochPoolRewards iotago.Mana `serix:""`
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
		Committee           []*CommitteeMemberResponse `serix:",lenPrefix=uint8"`
		TotalStake          iotago.BaseToken           `serix:""`
		TotalValidatorStake iotago.BaseToken           `serix:""`
		Epoch               iotago.EpochIndex          `serix:""`
	}
)
