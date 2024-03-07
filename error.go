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
	// ErrRMCNotFound gets returned when the RMC could not be found from the slot commitment.
	ErrRMCNotFound = ierrors.New("could not retrieve RMC for slot commitment")
	// ErrFailedToCalculateManaCost gets returned when the Mana cost could not be calculated.
	ErrFailedToCalculateManaCost = ierrors.New("could not calculate Mana cost for block")
	// ErrAccountExpired gets returned when the account is expired.
	ErrAccountExpired = ierrors.New("account expired")
	// ErrInvalidSignature gets returned when the signature is invalid.
	ErrInvalidSignature = ierrors.New("invalid signature")
)

// Errors that can occur before the transaction is executed.
var (
	// ErrUTXOInputInvalid gets returned when the UTXO input is invalid.
	ErrUTXOInputInvalid = ierrors.New("UTXO input is invalid")
	// ErrUnknownInputType gets returned for unknown input types.
	ErrUnknownInputType = ierrors.New("unknown input type")
	// ErrUnknownOutputType gets returned for unknown output types.
	ErrUnknownOutputType = ierrors.New("unknown output type")
	// ErrCommitmentInputMissing gets returned when the commitment has not been provided when needed.
	ErrCommitmentInputMissing = ierrors.New("commitment input required with reward or BIC input")
	// ErrCommitmentInputReferenceInvalid gets returned when the commitment input references an invalid commitment.
	ErrCommitmentInputReferenceInvalid = ierrors.New("commitment input references an invalid or non-existent commitment")
	// ErrBICInputReferenceInvalid gets returned when the BIC input is invalid.
	ErrBICInputReferenceInvalid = ierrors.New("BIC input reference cannot be loaded")
	// ErrRewardInputReferenceInvalid gets returned when the reward input does not reference a staking account or a delegation output.
	ErrRewardInputReferenceInvalid = ierrors.New("reward input does not reference a staking account or a delegation output")
	// ErrStakingRewardCalculationFailure gets returned when the validator reward could not be calculated due to storage issues or overflow.
	ErrStakingRewardCalculationFailure = ierrors.New("staking rewards could not be calculated due to storage issues or overflow")
	// ErrDelegationRewardCalculationFailure gets returned when the delegation reward could not be calculated due to storage issues or overflow.
	ErrDelegationRewardCalculationFailure = ierrors.New("delegation rewards could not be calculated due to storage issues or overflow")
	// ErrTxConflictRejected gets returned when the transaction was conflicting and the transaction was rejected.
	ErrTxConflictRejected = ierrors.New("transaction was conflicting and was rejected")
	// ErrTxOrphaned gets returned when the transaction was orphaned.
	ErrTxOrphaned = ierrors.New("transaction was orphaned")
	// ErrInputAlreadySpent gets returned when the input is already spent.
	ErrInputAlreadySpent = ierrors.New("input already spent")
)
