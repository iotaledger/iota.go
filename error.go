package iotago

import (
	"github.com/iotaledger/hive.go/ierrors"
)

// Errors used by failure codes

var (
	// ErrUTXOInputInvalid gets returned when the UTXO input is invalid.
	ErrUTXOInputInvalid = ierrors.New("UTXO input is invalid")
	// ErrBICInputInvalid gets returned when the BIC input is invalid.
	ErrBICInputInvalid = ierrors.New("BIC input is invalid")
	// ErrRewardInputInvalid gets returned when the reward input is invalid.
	ErrRewardInputInvalid = ierrors.New("reward input is invalid")
	// ErrCommitmentInputInvalid gets returned when the commitment input is invalid.
	ErrCommitmentInputInvalid = ierrors.New("commitment input is invalid")
	// ErrTxTypeInvalid gets returned for invalid transaction type.
	ErrTxTypeInvalid = ierrors.New("transaction type is invalid")
	// ErrUnlockBlockSignatureInvalid gets returned when the unlock block signature is invalid.
	ErrUnlockBlockSignatureInvalid = ierrors.New("unlock block signature is invalid")
	// ErrNativeTokenSetInvalid gets returned when the provided native tokens are invalid.
	ErrNativeTokenSetInvalid = ierrors.New("provided native tokens are invalid")
	// ErrChainTransitionInvalid gets returned when the chain transition was invalid.
	ErrChainTransitionInvalid = ierrors.New("chain transition is invalid")
	// ErrManaAmountInvalid gets returned when the mana amount is invalid.
	ErrManaAmountInvalid = ierrors.New("invalid mana amount, calculation error, or overflow")

	// ErrUnknownInputType gets returned for unknown input types.
	ErrUnknownInputType = ierrors.New("unknown input type")
	// ErrUnknownOutputType gets returned for unknown output types.
	ErrUnknownOutputType = ierrors.New("unknown output type")
	// ErrCommitmentInputMissing gets returned when the commitment has not been provided when needed.
	ErrCommitmentInputMissing = ierrors.New("commitment input required with reward or BIC input")
	// ErrNoStakingFeature gets returned when the validator reward could not be claimed.
	ErrNoStakingFeature = ierrors.New("staking reward claim failed due to no staking feature provided")

	// ErrFailedToClaimStakingReward gets returned when the validator reward could not be claimed.
	ErrFailedToClaimStakingReward = ierrors.New("staking claim failed")
	// ErrFailedToClaimDelegationReward gets returned when the delegation reward could not be claimed.
	ErrFailedToClaimDelegationReward = ierrors.New("delegation claim failed")

	// ErrTxConflicting gets returned when the transaction is conflicting.
	ErrTxConflicting = ierrors.New("transaction is conflicting")
	// ErrInputAlreadySpent gets returned when the input is already spent.
	ErrInputAlreadySpent = ierrors.New("input already spent")
)
