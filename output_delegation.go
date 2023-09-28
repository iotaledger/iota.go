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
	// ErrInvalidDelegationTransition gets returned when a Delegation Output is doing an invalid state transition.
	ErrInvalidDelegationTransition = ierrors.New("invalid delegation output transition")
	// ErrInvalidDelegationRewardsClaiming gets returned when it is invalid to claim rewards from a delegation output.
	ErrInvalidDelegationRewardsClaiming = ierrors.New("invalid delegation mana rewards claiming")
	// ErrInvalidDelegationNonZeroedID gets returned when a delegation output's delegation ID is not zeroed initially.
	ErrInvalidDelegationNonZeroedID = ierrors.New("delegation ID must be zeroed initially")
	// ErrInvalidDelegationModified gets returned when a delegation output's immutable fields are modified.
	ErrInvalidDelegationModified = ierrors.New("delegated amount, validator ID and start epoch cannot be modified")
	// ErrInvalidDelegationStartEpoch gets returned when a delegation output's start epoch is not set correctly
	// relative to the slot of the current epoch in which the voting power is calculated.
	ErrInvalidDelegationStartEpoch = ierrors.New("invalid start epoch")
	// ErrInvalidDelegationAmount gets returned when a delegation output's delegated amount is not equal to the amount.
	ErrInvalidDelegationAmount = ierrors.New("delegated amount equal to the amount")
	// ErrInvalidDelegationNonZeroEndEpoch gets returned when a delegation output's end epoch is not zero at genesis.
	ErrInvalidDelegationNonZeroEndEpoch = ierrors.New("end epoch must be set to zero at output genesis")
	// ErrInvalidDelegationEndEpoch gets returned when a delegation output's end epoch is not set correctly
	// relative to the slot of the current epoch in which the voting power is calculated.
	ErrInvalidDelegationEndEpoch = ierrors.New("invalid end epoch")
	// ErrDelegationCommitmentInputRequired gets returned when no commitment input was passed in a TX containing a Delegation Output.
	ErrDelegationCommitmentInputRequired = ierrors.New("delegation output validation requires a commitment input")
	emptyDelegationID                    = [DelegationIDLength]byte{}
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
	delegationOutputUnlockCondition  interface{ UnlockCondition }
	DelegationOutputUnlockConditions = UnlockConditions[delegationOutputUnlockCondition]
)

// DelegationOutput is an output type used to implement delegation.
type DelegationOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount BaseToken `serix:"0,mapKey=amount"`
	// The amount of IOTA tokens that were delegated when the output was created.
	DelegatedAmount BaseToken `serix:"1,mapKey=delegatedAmount"`
	// The identifier for this output.
	DelegationID DelegationID `serix:"2,mapKey=delegationId"`
	// The Account ID of the validator to which this output is delegating.
	ValidatorAddress *AccountAddress `serix:"3,mapKey=validatorAddress"`
	// The index of the first epoch for which this output delegates.
	StartEpoch EpochIndex `serix:"4,mapKey=startEpoch"`
	// The index of the last epoch for which this output delegates.
	EndEpoch EpochIndex `serix:"5,mapKey=endEpoch"`
	// The unlock conditions on this output.
	Conditions DelegationOutputUnlockConditions `serix:"6,mapKey=unlockConditions,omitempty"`
}

func (d *DelegationOutput) Clone() Output {
	return &DelegationOutput{
		Amount:           d.Amount,
		DelegatedAmount:  d.DelegatedAmount,
		DelegationID:     d.DelegationID,
		ValidatorAddress: d.ValidatorAddress,
		StartEpoch:       d.StartEpoch,
		EndEpoch:         d.EndEpoch,
		Conditions:       d.Conditions.Clone(),
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

	if !d.Conditions.Equal(otherOutput.Conditions) {
		return false
	}

	return true
}

func (d *DelegationOutput) Ident() Address {
	return d.Conditions.MustSet().Address().Address
}

func (d *DelegationOutput) UnlockableBy(ident Address, pastBoundedSlot SlotIndex, futureBoundedSlot SlotIndex) bool {
	ok, _ := outputUnlockableBy(d, nil, ident, pastBoundedSlot, futureBoundedSlot)
	return ok
}

func (d *DelegationOutput) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return outputOffsetVByteCost(rentStruct) +
		// TODO: Align vbyte factor weight of each field with TIP.
		// type prefix + amount + delegated amount + start epoch + end epoch
		rentStruct.VBFactorDelegation().Multiply(serializer.SmallTypeDenotationByteSize+BaseTokenSize+BaseTokenSize+EpochIndexLength+EpochIndexLength) +
		rentStruct.VBFactorDelegation().Multiply(DelegationIDLength) +
		rentStruct.VBFactorDelegation().Multiply(AccountAddressSerializedBytesSize) +
		d.Conditions.VBytes(rentStruct, nil)
}

func (d *DelegationOutput) syntacticallyValidate() error {
	// Address should never be nil.
	address := d.Conditions.MustSet().Address().Address

	if address.Type() == AddressImplicitAccountCreation {
		return ErrImplicitAccountCreationAddressInInvalidOutput
	}

	return nil
}

func (d *DelegationOutput) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
	return d.Conditions.WorkScore(workScoreStructure)
}

func (d *DelegationOutput) Chain() ChainID {
	return d.DelegationID
}

func (d *DelegationOutput) NativeTokenList() NativeTokens {
	return make(NativeTokens, 0)
}

func (d *DelegationOutput) FeatureSet() FeatureSet {
	return make(FeatureSet, 0)
}

func (d *DelegationOutput) UnlockConditionSet() UnlockConditionSet {
	return d.Conditions.MustSet()
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
		d.Conditions.Size()
}
