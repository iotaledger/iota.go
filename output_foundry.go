package iotago

import (
	"context"
	"encoding/binary"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

const (
	FoundrySerialNumberLength = serializer.UInt32ByteSize
	FoundryTokenSchemeLength  = serializer.OneByte
	// FoundryIDLength is the byte length of a FoundryID consisting out of the account address, serial number and token scheme.
	FoundryIDLength = AccountAddressSerializedBytesSize + FoundrySerialNumberLength + FoundryTokenSchemeLength
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

func FoundryIDFromAddressAndSerialNumberAndTokenScheme(addr Address, serialNumber uint32, tokenScheme TokenSchemeType) (NativeTokenID, error) {
	serixAPI := CommonSerixAPI()
	var foundryID FoundryID
	addrBytes, err := serixAPI.Encode(context.Background(), addr)
	if err != nil {
		return foundryID, err
	}
	copy(foundryID[:], addrBytes)
	binary.LittleEndian.PutUint32(foundryID[len(addrBytes):], serialNumber)
	foundryID[len(foundryID)-1] = byte(tokenScheme)

	return foundryID, nil
}

func (fID FoundryID) ToHex() string {
	return hexutil.EncodeHex(fID[:])
}

func (fID FoundryID) Addressable() bool {
	return false
}

// FoundrySerialNumber returns the serial number of the foundry.
func (fID FoundryID) FoundrySerialNumber() uint32 {
	return binary.LittleEndian.Uint32(fID[AccountAddressSerializedBytesSize : AccountAddressSerializedBytesSize+FoundrySerialNumberLength])
}

func (fID FoundryID) Matches(other ChainID) bool {
	otherFID, is := other.(FoundryID)
	if !is {
		return false
	}

	return fID == otherFID
}

func (fID FoundryID) AccountAddress() (*AccountAddress, error) {
	var addr Address
	if _, err := CommonSerixAPI().Decode(context.Background(), fID[:], &addr); err != nil {
		return nil, err
	}

	accountAddr, isAccountAddr := addr.(*AccountAddress)
	if !isAccountAddr {
		return nil, ierrors.New("address is not an account address")
	}

	return accountAddr, nil
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
	// The serial number of the foundry.
	SerialNumber uint32 `serix:"1,mapKey=serialNumber"`
	// The token scheme this foundry uses.
	TokenScheme TokenScheme `serix:"2,mapKey=tokenScheme"`
	// The unlock conditions on this output.
	Conditions FoundryOutputUnlockConditions `serix:"3,mapKey=unlockConditions,omitempty"`
	// The feature on the output.
	Features FoundryOutputFeatures `serix:"4,mapKey=features,omitempty"`
	// The immutable feature on the output.
	ImmutableFeatures FoundryOutputImmFeatures `serix:"5,mapKey=immutableFeatures,omitempty"`
}

func (f *FoundryOutput) Clone() Output {
	return &FoundryOutput{
		Amount:            f.Amount,
		SerialNumber:      f.SerialNumber,
		TokenScheme:       f.TokenScheme.Clone(),
		Conditions:        f.Conditions.Clone(),
		Features:          f.Features.Clone(),
		ImmutableFeatures: f.ImmutableFeatures.Clone(),
	}
}

func (f *FoundryOutput) Equal(other Output) bool {
	otherOutput, isSameType := other.(*FoundryOutput)
	if !isSameType {
		return false
	}

	if f.Amount != otherOutput.Amount {
		return false
	}

	if f.SerialNumber != otherOutput.SerialNumber {
		return false
	}

	if !f.TokenScheme.Equal(otherOutput.TokenScheme) {
		return false
	}

	if !f.Conditions.Equal(otherOutput.Conditions) {
		return false
	}

	if !f.Features.Equal(otherOutput.Features) {
		return false
	}

	if !f.ImmutableFeatures.Equal(otherOutput.ImmutableFeatures) {
		return false
	}

	return true
}

func (f *FoundryOutput) Ident() Address {
	return f.UnlockConditionSet().ImmutableAccount().Address
}

func (f *FoundryOutput) UnlockableBy(ident Address, pastBoundedSlotIndex SlotIndex, futureBoundedSlotIndex SlotIndex) bool {
	ok, _ := outputUnlockableBy(f, nil, ident, pastBoundedSlotIndex, futureBoundedSlotIndex)
	return ok
}

func (f *FoundryOutput) StorageScore(storageScoreStruct *StorageScoreStructure, _ StorageScoreFunc) StorageScore {
	return offsetOutput(storageScoreStruct) +
		storageScoreStruct.FactorData().Multiply(StorageScore(f.Size())) +
		f.TokenScheme.StorageScore(storageScoreStruct, nil) +
		f.Conditions.StorageScore(storageScoreStruct, nil) +
		f.Features.StorageScore(storageScoreStruct, nil) +
		f.ImmutableFeatures.StorageScore(storageScoreStruct, nil)
}

func (f *FoundryOutput) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	workScoreTokenScheme, err := f.TokenScheme.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	workScoreConditions, err := f.Conditions.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	workScoreFeatures, err := f.Features.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	workScoreImmutableFeatures, err := f.ImmutableFeatures.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	return workScoreTokenScheme.Add(workScoreConditions, workScoreFeatures, workScoreImmutableFeatures)
}

func (f *FoundryOutput) syntacticallyValidate() error {
	nativeTokenFeature := f.FeatureSet().NativeToken()
	if nativeTokenFeature == nil {
		return nil
	}

	foundryID, err := f.FoundryID()
	if err != nil {
		return err
	}

	// NativeTokenFeature ID should have the same ID as the foundry
	if !foundryID.Matches(nativeTokenFeature.ID) {
		return ierrors.Wrapf(ErrFoundryIDNativeTokenIDMismatch, "FoundryID: %s, NativeTokenID: %s", foundryID, nativeTokenFeature.ID)
	}

	return nil
}

func (f *FoundryOutput) ChainID() ChainID {
	foundryID, err := f.FoundryID()
	if err != nil {
		panic(err)
	}

	return foundryID
}

// FoundryID returns the FoundryID of this FoundryOutput.
func (f *FoundryOutput) FoundryID() (FoundryID, error) {
	return FoundryIDFromAddressAndSerialNumberAndTokenScheme(f.Ident(), f.SerialNumber, f.TokenScheme.Type())
}

// MustFoundryID works like FoundryID but panics if an error occurs.
func (f *FoundryOutput) MustFoundryID() FoundryID {
	id, err := f.FoundryID()
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
	return f.FoundryID()
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

func (f *FoundryOutput) BaseTokenAmount() BaseToken {
	return f.Amount
}

func (f *FoundryOutput) StoredMana() Mana {
	return 0
}

func (f *FoundryOutput) Type() OutputType {
	return OutputFoundry
}

func (f *FoundryOutput) Size() int {
	// OutputType
	return serializer.OneByte +
		BaseTokenSize +
		FoundrySerialNumberLength +
		f.TokenScheme.Size() +
		f.Conditions.Size() +
		f.Features.Size() +
		f.ImmutableFeatures.Size()
}
