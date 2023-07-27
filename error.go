package iotago

import (
	"github.com/iotaledger/hive.go/ierrors"
)

// Errors used by failure codes

var (
	// ErrUTXOInputInvalid gets returned when the input could not be retrieved.
	ErrUTXOInputInvalid = ierrors.New("failed to retrieve input references")
	// ErrBICInputInvalid gets returned when the BIC input could not be retrieved.
	ErrBICInputInvalid = ierrors.New("could not retrieve BIC input")
	// ErrRewardInputInvalid gets returned when the reward input could not be retrieved.
	ErrRewardInputInvalid = ierrors.New("could not retrieve reward input")
	// ErrCommitmentInputInvalid gets returned when the commitment could not be retrieved.
	ErrCommitmentInputInvalid = ierrors.New("could not retrieve commitment")
	// ErrTxTypeInvalid gets returned for unknown transaction types.
	ErrTxTypeInvalid = ierrors.New("unknown transaction type")
	// ErrUnlockBlockSignatureInvalid gets returned when the block signature unlock block signature is invalid.
	ErrUnlockBlockSignatureInvalid = ierrors.New("block signature unlock block signature invalid")
	// ErrNativeTokenSetInvalid gets returned when the provided native tokens are invalid.
	ErrNativeTokenSetInvalid = ierrors.New("provided native tokens are invalid")
	// ErrChainTransitionInvalid gets returned when the chain transition state failed.
	ErrChainTransitionInvalid = ierrors.New("chain transition state failed")
	// ErrManaAmountInvalid gets returned when the mana amount is invalid.
	ErrManaAmountInvalid = ierrors.New("invalid mana amount, calculation error, or overflow")

	// ErrUnknownInputType gets returned for unsupported input types.
	ErrUnknownInputType = ierrors.New("unsupported input type")
	// ErrUnknownOutputType gets returned for unknown output types.
	ErrUnknownOutputType = ierrors.New("unknown output type")
	// ErrCommitmentInputMissing gets returned when the commitment has not been provided when needed.
	ErrCommitmentInputMissing = ierrors.New("commitment input required with reward or BIC input")
	// ErrNoStakingFeature gets returned when the validator reward could not be claimed.
	ErrNoStakingFeature = ierrors.New("validator reward claim failed due to no staking feature provided")

	// ErrFailedToClaimValidatorReward gets returned when the validator reward could not be claimed.
	ErrFailedToClaimValidatorReward = ierrors.New("validator staking claim failed")
	// ErrFailedToClaimDelegatorReward gets returned when the delegator reward could not be claimed.
	ErrFailedToClaimDelegatorReward = ierrors.New("delegator staking claim failed")
)
