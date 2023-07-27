package iotago

import (
	"github.com/iotaledger/hive.go/ierrors"
)

var (
	ErrFailedToRetrieveInput = ierrors.New("failed to retrieve input references")
	// ErrUnsupportedInputType gets returned for unsupported input types.
	ErrUnsupportedInputType = ierrors.New("unsupported input type")
	// ErrUnsupportedContextInputType gets returned for unsupported context input types.
	ErrUnsupportedContextInputType = ierrors.New("unsupported context input type")
	// ErrUnknownAddrType gets returned for unknown address types.
	ErrUnknownAddrType = ierrors.New("unknown address type")
	// ErrUnknownOutputType gets returned for unknown output types.
	ErrUnknownOutputType = ierrors.New("unknown output type")
	// ErrUnknownTransactionEssenceType gets returned for unknown transaction essence types.
	ErrUnknownTransactionEssenceType = ierrors.New("unknown transaction essence type")
	// ErrUnknownTransactinType gets returned for unknown transaction types.
	ErrUnknownTransactinType = ierrors.New("unknown transaction type")
	// ErrUnknownUnlockType gets returned for unknown unlock.
	ErrUnknownUnlockType = ierrors.New("unknown unlock type")
	//ErrUnexpectedUnderlyingType gets returned for unknown input type of transaction.
	ErrUnexpectedUnderlyingType = ierrors.New("unexpected underlying type provided by the interface")
	// ErrCouldNotResolveBICInput gets returned when the BIC input could not be retrieved.
	ErrCouldNotResolveBICInput = ierrors.New("could not retrieve BIC input")
	// ErrCouldNotResolveRewardInput gets returned when the reward input could not be retrieved.
	ErrCouldNotResolveRewardInput = ierrors.New("could not retrieve reward input")
	// ErrCouldNorRetrieveCommittment gets returned when the commitment could not be retrieved.
	ErrCouldNorRetrieveCommittment = ierrors.New("could not retrieve committment")
	// ErrCommittmentInputMissing gets returned when the commitment has not been provided when needed.
	ErrCommittmentInputMissing = ierrors.New("commitment input required with reward or BIC input")
	// ErrNoStakingFeature gets returned when the validator reward could not be claimed.
	ErrNoStakingFeature = ierrors.New("validator reward claim failed due to no staking feature provided")
	// ErrFailedToClaimValidatorReward gets returned when the validator reward could not be claimed.
	ErrFailedToClaimValidatorReward = ierrors.New("validator staking claim failed")
	// ErrFailedToClaimDelegatorReward gets returned when the delegator reward could not be claimed.
	ErrFailedToClaimDelegatorReward = ierrors.New("delegator staking claim failed")
	// ErrUnlockBlockSignatureInvalid gets returned when the block signature unlock block signature is invalid.
	ErrUnlockBlockSignatureInvalid = ierrors.New("block signature unlock block signature invalid")
	// ErrInvalidNativeTokenSet gets returned when the provided native tokens are invalid.
	ErrInvalidNativeTokenSet = ierrors.New("provided native tokens are invalid")
	// ErrChainTransitionInvalid gets returned when the chain transition state failed.
	ErrChainTransitionInvalid = ierrors.New("chain transition state failed")
	// ErrInvalidManaAmount gets returned when the mana amount is invalid.
	ErrInvalidManaAmount = ierrors.New("invalid mana amount, calculation error, or overflow")
)
