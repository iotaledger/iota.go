package iotago

import (
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

const (
	// 	DelegationIDLength is the byte length of a DelegationID.
	DelegationIDLength = blake2b.Size256
)

var (
	// ErrDelegationTransitionInvalid gets returned when a Delegation Output is doing an invalid state transition.
	ErrDelegationTransitionInvalid = ierrors.New("invalid delegation output transition")
	// ErrDelegationCommitmentInputMissing gets returned when no commitment input was passed in a TX containing a Delegation Output.
	ErrDelegationCommitmentInputMissing = ierrors.New("delegation output validation requires a commitment input")
	// ErrDelegationRewardInputMissing gets returned when no reward input was passed in a TX destroying a Delegation Output.
	ErrDelegationRewardInputMissing = ierrors.New("delegation output cannot be destroyed without a reward input")
	// ErrDelegationRewardsClaimingInvalid gets returned when it is invalid to claim rewards from a delegation output.
	ErrDelegationRewardsClaimingInvalid = ierrors.New("invalid delegation mana rewards claiming")
	// ErrDelegationOutputTransitionedTwice gets returned when a delegation output is attempted to be transitioned twice.
	ErrDelegationOutputTransitionedTwice = ierrors.New("delegation output attempted to be transitioned twice")
	// ErrDelegationModified gets returned when a delegation output's immutable fields are modified.
	ErrDelegationModified = ierrors.New("delegated amount, validator ID and start epoch cannot be modified")
	// ErrDelegationStartEpochInvalid gets returned when a delegation output's start epoch is not set correctly
	// relative to the slot of the current epoch in which the voting power is calculated.
	ErrDelegationStartEpochInvalid = ierrors.New("invalid start epoch")
	// ErrDelegationAmountMismatch gets returned when a delegation output's delegated amount is not equal to the amount.
	ErrDelegationAmountMismatch = ierrors.New("delegated amount does not match amount")
	// ErrDelegationEndEpochNotZero gets returned when a delegation output's end epoch is not zero at genesis.
	ErrDelegationEndEpochNotZero = ierrors.New("end epoch must be set to zero at output genesis")
	// ErrDelegationEndEpochInvalid gets returned when a delegation output's end epoch is not set correctly
	// relative to the slot of the current epoch in which the voting power is calculated.
	ErrDelegationEndEpochInvalid = ierrors.New("delegation end epoch does not match current epoch")
	// ErrDelegationValidatorAddressEmpty gets returned when the Validator Address in a Delegation Output is empty.
	ErrDelegationValidatorAddressEmpty = ierrors.New("delegation output's validator address is empty")
	emptyDelegationID                  = [DelegationIDLength]byte{}
)

func EmptyDelegationID() DelegationID {
	return emptyDelegationID
}

// DelegationID is the identifier for a Delegation Output.
// It is computed as the Blake2b-256 hash of the OutputID of the output which created the Delegation Output.
type DelegationID [DelegationIDLength]byte

// DelegationIDs are DelegationID(s).
type DelegationIDs []DelegationID

func (delegationID DelegationID) Addressable() bool {
	return false
}

func (delegationID DelegationID) Key() interface{} {
	return delegationID.String()
}

func (delegationID DelegationID) Empty() bool {
	return delegationID == emptyDelegationID
}

func (delegationID DelegationID) ToAddress() ChainAddress {
	panic("Delegation ID is not addressable")
}

func (delegationID DelegationID) Matches(other ChainID) bool {
	otherDelegationID, isDelegationID := other.(DelegationID)
	if !isDelegationID {
		return false
	}

	return delegationID == otherDelegationID
}

func (delegationID DelegationID) String() string {
	return hexutil.EncodeHex(delegationID[:])
}

func (delegationID DelegationID) ToHex() string {
	return hexutil.EncodeHex(delegationID[:])
}

func (delegationID DelegationID) FromOutputID(outid OutputID) ChainID {
	return DelegationIDFromOutputID(outid)
}

// DelegationIDFromOutputID returns the DelegationID computed from a given OutputID.
func DelegationIDFromOutputID(outputID OutputID) DelegationID {
	return blake2b.Sum256(outputID[:])
}

type (
	DelegationOutputUnlockCondition  interface{ UnlockCondition }
	DelegationOutputUnlockConditions = UnlockConditions[DelegationOutputUnlockCondition]
)

// DelegationOutput is an output type used to implement delegation.
type DelegationOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount BaseToken `serix:""`
	// The amount of IOTA tokens that were delegated when the output was created.
	DelegatedAmount BaseToken `serix:""`
	// The identifier for this output.
	DelegationID DelegationID `serix:""`
	// The Account ID of the validator to which this output is delegating.
	ValidatorAddress *AccountAddress `serix:""`
	// The index of the first epoch for which this output delegates.
	StartEpoch EpochIndex `serix:""`
	// The index of the last epoch for which this output delegates.
	EndEpoch EpochIndex `serix:""`
	// The unlock conditions on this output.
	UnlockConditions DelegationOutputUnlockConditions `serix:",omitempty"`
}

func (d *DelegationOutput) Clone() Output {
	return &DelegationOutput{
		Amount:           d.Amount,
		DelegatedAmount:  d.DelegatedAmount,
		DelegationID:     d.DelegationID,
		ValidatorAddress: d.ValidatorAddress,
		StartEpoch:       d.StartEpoch,
		EndEpoch:         d.EndEpoch,
		UnlockConditions: d.UnlockConditions.Clone(),
	}
}

func (d *DelegationOutput) Equal(other Output) bool {
	otherOutput, isSameType := other.(*DelegationOutput)
	if !isSameType {
		return false
	}

	if d.Amount != otherOutput.Amount {
		return false
	}

	if d.DelegatedAmount != otherOutput.DelegatedAmount {
		return false
	}

	if d.DelegationID != otherOutput.DelegationID {
		return false
	}

	if !d.ValidatorAddress.Equal(otherOutput.ValidatorAddress) {
		return false
	}

	if d.StartEpoch != otherOutput.StartEpoch {
		return false
	}

	if d.EndEpoch != otherOutput.EndEpoch {
		return false
	}

	if !d.UnlockConditions.Equal(otherOutput.UnlockConditions) {
		return false
	}

	return true
}

func (d *DelegationOutput) Ident() Address {
	return d.UnlockConditions.MustSet().Address().Address
}

func (d *DelegationOutput) UnlockableBy(ident Address, pastBoundedSlot SlotIndex, futureBoundedSlot SlotIndex) bool {
	ok, _ := outputUnlockableBy(d, nil, ident, pastBoundedSlot, futureBoundedSlot)
	return ok
}

func (d *DelegationOutput) StorageScore(storageScoreStruct *StorageScoreStructure, _ StorageScoreFunc) StorageScore {
	return storageScoreStruct.OffsetOutput +
		storageScoreStruct.FactorData().Multiply(StorageScore(d.Size())) +
		storageScoreStruct.OffsetDelegation() +
		d.UnlockConditions.StorageScore(storageScoreStruct, nil)
}

func (d *DelegationOutput) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	workScoreConditions, err := d.UnlockConditions.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	return workScoreParameters.Output.Add(workScoreConditions)
}

func (d *DelegationOutput) ChainID() ChainID {
	return d.DelegationID
}

func (d *DelegationOutput) FeatureSet() FeatureSet {
	return make(FeatureSet, 0)
}

func (d *DelegationOutput) UnlockConditionSet() UnlockConditionSet {
	return d.UnlockConditions.MustSet()
}

func (d *DelegationOutput) BaseTokenAmount() BaseToken {
	return d.Amount
}

func (d *DelegationOutput) StoredMana() Mana {
	return 0
}

func (d *DelegationOutput) Type() OutputType {
	return OutputDelegation
}

func (d *DelegationOutput) Size() int {
	// OutputType
	return serializer.OneByte +
		BaseTokenSize +
		BaseTokenSize +
		DelegationIDLength +
		// Account Address Type Byte
		serializer.SmallTypeDenotationByteSize +
		AccountAddressBytesLength +
		EpochIndexLength +
		EpochIndexLength +
		d.UnlockConditions.Size()
}
