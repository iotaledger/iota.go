package iotago

import (
	"cmp"
	"math/big"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	// Uint256ByteSize defines the size of an uint256.
	Uint256ByteSize = 32

	// NativeTokenIDLength is the byte length of a NativeTokenID which is the same thing as a FoundryID.
	NativeTokenIDLength = FoundryIDLength
)

var (
	// ErrNativeTokenAmountLessThanEqualZero gets returned when a NativeToken.Amount is not bigger than 0.
	ErrNativeTokenAmountLessThanEqualZero = ierrors.New("native token amount must be greater than zero")
	// ErrNativeTokenSumExceedsUint256 gets returned when a NativeToken.Amount addition results in a value bigger than the max value of a uint256.
	ErrNativeTokenSumExceedsUint256 = ierrors.New("native token sum exceeds max value of a uint256")
	// ErrNativeTokenSumUnbalanced gets returned when two NativeTokenSum(s) are unbalanced.
	ErrNativeTokenSumUnbalanced = ierrors.New("native token sum is unbalanced")
	// ErrFoundryIDNativeTokenIDMismatch gets returned when a native token features exists in a foundry output but the IDs mismatch.
	ErrFoundryIDNativeTokenIDMismatch = ierrors.New("native token ID in foundry output must match the foundry ID")
	// ErrNativeTokenSetInvalid gets returned when the provided native tokens are invalid.
	ErrNativeTokenSetInvalid = ierrors.New("provided native tokens are invalid")
)

// NativeTokenID is an identifier which uniquely identifies a NativeToken.
type NativeTokenID = FoundryID

// NativeTokenSum is a mapping of NativeTokenID to a sum value.
type NativeTokenSum map[NativeTokenID]*big.Int

// ValueOrBigInt0 returns the value for the given native token or a 0 big int.
func (sum NativeTokenSum) ValueOrBigInt0(id NativeTokenID) *big.Int {
	v, has := sum[id]
	if !has {
		return big.NewInt(0)
	}

	return v
}

// NativeTokenFeature is a feature that holds a native token which represents a token that resides natively on the ledger.
type NativeTokenFeature struct {
	ID     NativeTokenID `serix:""`
	Amount *big.Int      `serix:""`
}

// Clone clones the NativeTokenFeature.
func (n *NativeTokenFeature) Clone() Feature {
	cpy := &NativeTokenFeature{}
	copy(cpy.ID[:], n.ID[:])
	cpy.Amount = new(big.Int).Set(n.Amount)

	return cpy
}

func (n *NativeTokenFeature) Type() FeatureType {
	return FeatureNativeToken
}

func (n *NativeTokenFeature) StorageScore(_ *StorageScoreStructure, _ StorageScoreFunc) StorageScore {
	return 0
}

func (n *NativeTokenFeature) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	return workScoreParameters.NativeToken, nil
}

func (n *NativeTokenFeature) Compare(other Feature) int {
	return cmp.Compare(n.Type(), other.Type())
}

// Equal checks whether other is equal to this NativeToken.
func (n *NativeTokenFeature) Equal(other Feature) bool {
	otherFeat, is := other.(*NativeTokenFeature)
	if !is {
		return false
	}

	if n.ID != otherFeat.ID {
		return false
	}

	return n.Amount.Cmp(otherFeat.Amount) == 0
}

func (n *NativeTokenFeature) Size() int {
	// FeatureType + NativeTokenID + Amount (uint256)
	return serializer.SmallTypeDenotationByteSize + NativeTokenIDLength + serializer.UInt256ByteSize
}
