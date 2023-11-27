package iotago

import (
	"github.com/iotaledger/hive.go/ierrors"
)

// Errors used by failure codes

// Errors used for block failures.
var (
	// ErrBlockParentNotFound gets returned when the block parent could not be found.
	ErrBlockParentNotFound = ierrors.New("block parent not found")
	// ErrBlockIssuingTimeNonMonotonic gets returned when the block issuing time is not monotonically increasing compared to the block's parents.
	ErrBlockIssuingTimeNonMonotonic = ierrors.New("block issuing time is not monotonically increasing compared to parents")
	// ErrIssuerAccountNotFound gets returned when the issuer account could not be found.
	ErrIssuerAccountNotFound = ierrors.New("could not retrieve account information for block issuer")
	// ErrBurnedInsufficientMana gets returned when the issuer account burned insufficient Mana for a block.
	ErrBurnedInsufficientMana = ierrors.New("block issuer account burned insufficient Mana")
	// ErrBlockVersionInvalid gets returned when the block version is invalid to retrieve API.
	ErrBlockVersionInvalid = ierrors.New("could not retrieve API for block version")
	// ErrRMCNotFound gets returned when the RMC could not be found from the slot commitment.
	ErrRMCNotFound = ierrors.New("could not retrieve RMC for slot commitment")
	// ErrFailedToCalculateManaCost gets returned when the Mana cost could not be calculated.
	ErrFailedToCalculateManaCost = ierrors.New("could not calculate Mana cost for block")
	// ErrNegativeBIC gets returned when the BIC of the issuer account is negative.
	ErrNegativeBIC = ierrors.New("negative BIC")
	// ErrAccountExpired gets returned when the account is expired.
	ErrAccountExpired = ierrors.New("account expired")
	// ErrInvalidSignature gets returned when the signature is invalid.
	ErrInvalidSignature = ierrors.New("invalid signature")
)

// Errors used for transaction failures.
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
	ErrFailedToClaimStakingReward = ierrors.New("staking rewards claim failed")
	// ErrFailedToClaimDelegationReward gets returned when the delegation reward could not be claimed.
	ErrFailedToClaimDelegationReward = ierrors.New("delegation rewards claim failed")

	// ErrTxConflicting gets returned when the transaction is conflicting.
	ErrTxConflicting = ierrors.New("transaction is conflicting")
	// ErrInputAlreadySpent gets returned when the input is already spent.
	ErrInputAlreadySpent = ierrors.New("input already spent")
)
