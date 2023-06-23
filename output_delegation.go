package iotago

import (
	"errors"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/hexutil"
	"github.com/iotaledger/iota.go/v4/util"
	"golang.org/x/crypto/blake2b"
)

const (
	// 	DelegationIDLength is the byte length of a DelegationID.
	DelegationIDLength = blake2b.Size256
)

var (
	// ErrInvalidDelegationTransition gets returned when a Delegation Output's initial state machine is invalid.
	ErrInvalidDelegationGenesis = errors.New("invalid delegation output genesis")
	// ErrInvalidDelegationTransition gets returned when a Delegation Output is doing an invalid state transition.
	ErrInvalidDelegationTransition = errors.New("invalid delegation output transition")
	// ErrInvalidDelegationRewardsClaiming gets returned when a Delegation Output is doing an invalid state transition.
	ErrInvalidDelegationRewardsClaiming = errors.New("invalid delegation mana rewards claiming")
	emptyDelegationID                   = [DelegationIDLength]byte{}
)

func EmptyDelegationId() DelegationID {
	return emptyDelegationID
}

// DelegationID is the identifier for a Delegation Output.
// It is computed as the Blake2b-256 hash of the OutputID of the output which created the Delegation Output.
type DelegationID [DelegationIDLength]byte

// DelegationIDs are DelegationID(s).
type DelegationIDs []DelegationID

func (delegationId DelegationID) Addressable() bool {
	return false
}

func (delegationId DelegationID) Key() interface{} {
	return delegationId.String()
}

func (delegationId DelegationID) Empty() bool {
	return delegationId == emptyDelegationID
}

func (delegationId DelegationID) ToAddress() ChainAddress {
	panic("Delegation ID is not addressable")
}

func (delegationId DelegationID) Matches(other ChainID) bool {
	otherDelegationId, isDelegationId := other.(DelegationID)
	if !isDelegationId {
		return false
	}
	return delegationId == otherDelegationId
}

func (delegationId DelegationID) String() string {
	return hexutil.EncodeHex(delegationId[:])
}

func (delegationId DelegationID) ToHex() string {
	return hexutil.EncodeHex(delegationId[:])
}

func (id DelegationID) FromOutputID(outid OutputID) ChainID {
	return DelegationIDFromOutputID(outid)
}

// DelegationIDFromOutputID returns the DelegationID computed from a given OutputID.
func DelegationIDFromOutputID(outputID OutputID) DelegationID {
	return blake2b.Sum256(outputID[:])
}

type (
	delegationOutputUnlockCondition  interface{ UnlockCondition }
	delegationOutputImmFeature       interface{ Feature }
	DelegationOutputUnlockConditions = UnlockConditions[delegationOutputUnlockCondition]
	DelegationOutputImmFeatures      = Features[delegationOutputImmFeature]
)

// DelegationOutput is an output type used to implement delegation.
type DelegationOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount uint64 `serix:"0,mapKey=amount"`
	// The amount of IOTA tokens that were delegated when the output was created.
	DelegatedAmount uint64 `serix:"1,mapKey=delegatedAmount"`
	// The identifier for this output.
	DelegationID DelegationID `serix:"2,mapKey=delegationId"`
	// The Account ID of the validator to which this output is delegating.
	ValidatorID AccountID `serix:"3,mapKey=validatorId"`
	// The index of the first epoch for which this output delegates.
	StartEpoch uint64 `serix:"4,mapKey=startEpoch"`
	// The index of the last epoch for which this output delegates.
	EndEpoch uint64 `serix:"5,mapKey=endEpoch"`
	// The unlock conditions on this output.
	Conditions DelegationOutputUnlockConditions `serix:"6,mapKey=unlockConditions,omitempty"`
	// The immutable feature on the output.
	ImmutableFeatures DelegationOutputImmFeatures `serix:"7,mapKey=immutableFeatures,omitempty"`
}

func (d *DelegationOutput) Clone() Output {
	return &DelegationOutput{
		Amount:            d.Amount,
		DelegatedAmount:   d.DelegatedAmount,
		DelegationID:      d.DelegationID,
		ValidatorID:       d.ValidatorID,
		StartEpoch:        d.StartEpoch,
		EndEpoch:          d.EndEpoch,
		Conditions:        d.Conditions.Clone(),
		ImmutableFeatures: d.ImmutableFeatures.Clone(),
	}
}

func (d *DelegationOutput) Ident() Address {
	return d.Conditions.MustSet().Address().Address
}

func (d *DelegationOutput) UnlockableBy(ident Address, txCreationTime SlotIndex) bool {
	ok, _ := outputUnlockable(d, nil, ident, txCreationTime)
	return ok
}

func (d *DelegationOutput) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return outputOffsetVByteCost(rentStruct) +
		// type prefix + amount + delegated amount + start epoch + end epoch
		rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+serializer.UInt64ByteSize*4) +
		rentStruct.VBFactorData.Multiply(DelegationIDLength) +
		rentStruct.VBFactorData.Multiply(AccountIDLength) +
		d.Conditions.VBytes(rentStruct, nil) +
		d.ImmutableFeatures.VBytes(rentStruct, nil)
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

func (d *DelegationOutput) ImmutableFeatureSet() FeatureSet {
	return d.ImmutableFeatures.MustSet()
}

func (d *DelegationOutput) Deposit() uint64 {
	return d.Amount
}

func (d *DelegationOutput) StoredMana() uint64 {
	return 0
}

func (d *DelegationOutput) Type() OutputType {
	return OutputDelegation
}

func (d *DelegationOutput) Size() int {
	return util.NumByteLen(byte(OutputDelegation)) +
		util.NumByteLen(d.Amount) +
		util.NumByteLen(d.DelegatedAmount) +
		DelegationIDLength +
		AccountIDLength +
		util.NumByteLen(d.StartEpoch) +
		util.NumByteLen(d.EndEpoch) +
		d.Conditions.Size() +
		d.ImmutableFeatures.Size()
}
