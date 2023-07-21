package iotago

import (
	"context"
	"encoding/binary"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/hexutil"
	"github.com/iotaledger/iota.go/v4/util"
)

const (
	// FoundryIDLength is the byte length of a FoundryID consisting out of the account address, serial number and token scheme.
	FoundryIDLength = AccountAddressSerializedBytesSize + serializer.UInt32ByteSize + serializer.OneByte
)

var (
	// ErrNonUniqueFoundryOutputs gets returned when multiple FoundryOutput(s) with the same FoundryID exist within an OutputsByType.
	ErrNonUniqueFoundryOutputs = ierrors.New("non unique foundries within outputs")
	// ErrInvalidFoundryStateTransition gets returned when a foundry is doing an invalid state transition.
	ErrInvalidFoundryStateTransition = ierrors.New("invalid foundry state transition")

	emptyFoundryID = [FoundryIDLength]byte{}
)

// FoundryID defines the identifier for a foundry consisting out of the address, serial number and TokenScheme.
type FoundryID [FoundryIDLength]byte

func (fID FoundryID) ToHex() string {
	return hexutil.EncodeHex(fID[:])
}

func (fID FoundryID) Addressable() bool {
	return false
}

// FoundrySerialNumber returns the serial number of the foundry.
func (fID FoundryID) FoundrySerialNumber() uint32 {
	return binary.LittleEndian.Uint32(fID[AccountAddressSerializedBytesSize : AccountAddressSerializedBytesSize+serializer.UInt32ByteSize])
}

func (fID FoundryID) Matches(other ChainID) bool {
	otherFID, is := other.(FoundryID)
	if !is {
		return false
	}
	return fID == otherFID
}

func (fID FoundryID) ToAddress() ChainAddress {
	panic("foundry ID is not addressable")
}

func (fID FoundryID) Empty() bool {
	return fID == emptyFoundryID
}

func (fID FoundryID) Key() interface{} {
	return fID.String()
}

func (fID FoundryID) String() string {
	return hexutil.EncodeHex(fID[:])
}

// FoundryOutputs is a slice of FoundryOutput(s).
type FoundryOutputs []*FoundryOutput

// FoundryOutputsSet is a set of FoundryOutput(s).
type FoundryOutputsSet map[FoundryID]*FoundryOutput

type (
	foundryOutputUnlockCondition  interface{ UnlockCondition }
	foundryOutputFeature          interface{ Feature }
	foundryOutputImmFeature       interface{ Feature }
	FoundryOutputUnlockConditions = UnlockConditions[foundryOutputUnlockCondition]
	FoundryOutputFeatures         = Features[foundryOutputFeature]
	FoundryOutputImmFeatures      = Features[foundryOutputImmFeature]
)

// FoundryOutput is an output type which controls the supply of user defined native tokens.
type FoundryOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount BaseToken `serix:"0,mapKey=amount"`
	// The native tokens held by the output.
	NativeTokens NativeTokens `serix:"1,mapKey=nativeTokens,omitempty"`
	// The serial number of the foundry.
	SerialNumber uint32 `serix:"2,mapKey=serialNumber"`
	// The token scheme this foundry uses.
	TokenScheme TokenScheme `serix:"3,mapKey=tokenScheme"`
	// The unlock conditions on this output.
	Conditions FoundryOutputUnlockConditions `serix:"4,mapKey=unlockConditions,omitempty"`
	// The feature on the output.
	Features FoundryOutputFeatures `serix:"5,mapKey=features,omitempty"`
	// The immutable feature on the output.
	ImmutableFeatures FoundryOutputImmFeatures `serix:"6,mapKey=immutableFeatures,omitempty"`
}

func (f *FoundryOutput) Clone() Output {
	return &FoundryOutput{
		Amount:            f.Amount,
		NativeTokens:      f.NativeTokens.Clone(),
		SerialNumber:      f.SerialNumber,
		TokenScheme:       f.TokenScheme.Clone(),
		Conditions:        f.Conditions.Clone(),
		Features:          f.Features.Clone(),
		ImmutableFeatures: f.ImmutableFeatures.Clone(),
	}
}

func (f *FoundryOutput) Ident() Address {
	return f.UnlockConditionSet().ImmutableAccount().Address
}

func (f *FoundryOutput) UnlockableBy(ident Address, txCreationTime SlotIndex) bool {
	ok, _ := outputUnlockable(f, nil, ident, txCreationTime)
	return ok
}

func (f *FoundryOutput) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return outputOffsetVByteCost(rentStruct) +
		// prefix + amount + stored mana
		rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+serializer.UInt64ByteSize+serializer.UInt64ByteSize) +
		f.NativeTokens.VBytes(rentStruct, nil) +
		// serial number
		rentStruct.VBFactorData.Multiply(serializer.UInt32ByteSize) +
		f.TokenScheme.VBytes(rentStruct, nil) +
		f.Conditions.VBytes(rentStruct, nil) +
		f.Features.VBytes(rentStruct, nil) +
		f.ImmutableFeatures.VBytes(rentStruct, nil)
}

func (f *FoundryOutput) WorkScore(workScoreStructure *WorkScoreStructure) WorkScore {
	return f.ImmutableFeatures.WorkScore(workScoreStructure) +
		// prefix + amount + stored mana
		workScoreStructure.Factors.Data.Multiply(serializer.SmallTypeDenotationByteSize+serializer.UInt64ByteSize+serializer.UInt64ByteSize) +
		f.NativeTokens.WorkScore(workScoreStructure) +
		// serial number
		workScoreStructure.Factors.Data.Multiply(serializer.UInt32ByteSize) +
		f.Conditions.WorkScore(workScoreStructure) +
		f.Features.WorkScore(workScoreStructure) +
		f.ImmutableFeatures.WorkScore(workScoreStructure)
}

func (f *FoundryOutput) Chain() ChainID {
	foundryID, err := f.ID()
	if err != nil {
		panic(err)
	}
	return foundryID
}

// ID returns the FoundryID of this FoundryOutput.
func (f *FoundryOutput) ID() (FoundryID, error) {
	serixAPI := commonSerixAPI()
	var foundryID FoundryID
	addrBytes, err := serixAPI.Encode(context.Background(), f.Ident())
	if err != nil {
		return foundryID, err
	}
	copy(foundryID[:], addrBytes)
	binary.LittleEndian.PutUint32(foundryID[len(addrBytes):], f.SerialNumber)
	foundryID[len(foundryID)-1] = byte(f.TokenScheme.Type())
	return foundryID, nil
}

// MustID works like ID but panics if an error occurs.
func (f *FoundryOutput) MustID() FoundryID {
	id, err := f.ID()
	if err != nil {
		panic(err)
	}
	return id
}

// MustNativeTokenID works like NativeTokenID but panics if there is an error.
func (f *FoundryOutput) MustNativeTokenID() NativeTokenID {
	nativeTokenID, err := f.NativeTokenID()
	if err != nil {
		panic(err)
	}
	return nativeTokenID
}

// NativeTokenID returns the NativeTokenID this FoundryOutput operates on.
func (f *FoundryOutput) NativeTokenID() (NativeTokenID, error) {
	return f.ID()
}

func (f *FoundryOutput) NativeTokenList() NativeTokens {
	return f.NativeTokens
}

func (f *FoundryOutput) FeatureSet() FeatureSet {
	return f.Features.MustSet()
}

func (f *FoundryOutput) UnlockConditionSet() UnlockConditionSet {
	return f.Conditions.MustSet()
}

func (f *FoundryOutput) ImmutableFeatureSet() FeatureSet {
	return f.ImmutableFeatures.MustSet()
}

func (f *FoundryOutput) Deposit() BaseToken {
	return f.Amount
}

func (f *FoundryOutput) StoredMana() Mana {
	return 0
}

func (f *FoundryOutput) Type() OutputType {
	return OutputFoundry
}

func (f *FoundryOutput) Size() int {
	return util.NumByteLen(byte(OutputFoundry)) +
		BaseTokenSize +
		f.NativeTokens.Size() +
		util.NumByteLen(f.SerialNumber) +
		f.TokenScheme.Size() +
		f.Conditions.Size() +
		f.Features.Size() +
		f.ImmutableFeatures.Size()
}
