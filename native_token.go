package iotago

import (
	"errors"
	"math/big"

	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	// MinNativeTokenCountPerOutput min number of different native tokens that can reside in one output.
	MinNativeTokenCountPerOutput = 0
	// MaxNativeTokenCountPerOutput max number of different native tokens that can reside in one output.
	MaxNativeTokenCountPerOutput = 64

	// MaxNativeTokensCount is the max number of native tokens which can occur in a transaction (sum input/output side).
	MaxNativeTokensCount = 64

	// Uint256ByteSize defines the size of an uint256.
	Uint256ByteSize = 32

	// NativeTokenIDLength is the byte length of a NativeTokenID which is the same thing as a FoundryID.
	NativeTokenIDLength = FoundryIDLength

	// NativeTokenVByteCost defines the static virtual byte cost of a NativeToken.
	NativeTokenVByteCost = NativeTokenIDLength + Uint256ByteSize
)

var (
	// ErrNativeTokenAmountLessThanEqualZero gets returned when a NativeToken.Amount is not bigger than 0.
	ErrNativeTokenAmountLessThanEqualZero = errors.New("native token must be a value bigger than zero")
	// ErrNativeTokenSumExceedsUint256 gets returned when a NativeToken.Amount addition results in a value bigger than the max value of a uint256.
	ErrNativeTokenSumExceedsUint256 = errors.New("native token sum exceeds max value of a uint256")
	// ErrNonUniqueNativeTokens gets returned when multiple NativeToken(s) with the same NativeTokenID exist within sets.
	ErrNonUniqueNativeTokens = errors.New("non unique native tokens")
	// ErrNativeTokenSumUnbalanced gets returned when two NativeTokenSum(s) are unbalanced.
	ErrNativeTokenSumUnbalanced = errors.New("native token sums are unbalanced")

	nativeTokensArrayRules = &serializer.ArrayRules{
		Min: MinNativeTokenCountPerOutput,
		Max: MaxNativeTokenCountPerOutput,
		// uniqueness must be checked only by examining the actual NativeTokenID bytes
		UniquenessSliceFunc: func(next []byte) []byte { return next[:NativeTokenIDLength] },
		ValidationMode:      serializer.ArrayValidationModeNoDuplicates | serializer.ArrayValidationModeLexicalOrdering,
	}
)

// NativeTokenArrayRules returns array rules defining the constraints on a slice of NativeTokens.
func NativeTokenArrayRules() serializer.ArrayRules {
	return *nativeTokensArrayRules
}

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

// NativeTokenSumFunc gets called with a NativeTokenID and the sums of input/output side.
type NativeTokenSumFunc func(id NativeTokenID, aSum *big.Int, bSum *big.Int) error

// NativeTokensSet is a set of NativeToken(s).
type NativeTokensSet map[NativeTokenID]*NativeToken

// NativeTokens is a set of NativeToken.
type NativeTokens []*NativeToken

// Set converts the slice into a NativeTokenSet.
// Returns an error if a NativeTokenID occurs multiple times.
func (n NativeTokens) Set() (NativeTokensSet, error) {
	set := make(NativeTokensSet)
	for _, token := range n {
		if _, has := set[token.ID]; has {
			return nil, ErrNonUniqueNativeTokens
		}
		set[token.ID] = token
	}
	return set, nil
}

// MustSet works like Set but panics if an error occurs.
// This function is therefore only safe to be called when it is given,
// that a NativeTokens slice does not contain the same NativeTokenID multiple times.
func (n NativeTokens) MustSet() NativeTokensSet {
	set, err := n.Set()
	if err != nil {
		panic(err)
	}
	return set
}

// Clone clones this slice of NativeToken(s).
func (n NativeTokens) Clone() NativeTokens {
	cpy := make(NativeTokens, len(n))
	for i, ele := range n {
		cpy[i] = ele.Clone()
	}
	return cpy
}

func (n NativeTokens) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	// length prefix + (native token count * static native token cost)
	return rentStruct.VBFactorData.Multiply(uint64(serializer.OneByte + len(n)*NativeTokenVByteCost))
}

func (n NativeTokens) Size() int {
	sum := serializer.OneByte // 1 byte length prefix
	for _, token := range n {
		sum += token.Size()
	}
	return sum
}

// Equal checks whether other is equal to this slice.
func (n NativeTokens) Equal(other NativeTokens) bool {
	if len(n) != len(other) {
		return false
	}
	for i, a := range n {
		if !a.Equal(other[i]) {
			return false
		}
	}
	return true
}

// NativeToken represents a token which resides natively on the ledger.
type NativeToken struct {
	ID     NativeTokenID `serix:"0,mapKey=id"`
	Amount *big.Int      `serix:"1,mapKey=amount"`
}

// Clone clones the NativeToken.
func (n *NativeToken) Clone() *NativeToken {
	cpy := &NativeToken{}
	copy(cpy.ID[:], n.ID[:])
	cpy.Amount = new(big.Int).Set(n.Amount)
	return cpy
}

func (n *NativeToken) VBytes(_ *RentStructure, _ VBytesFunc) uint64 {
	return NativeTokenVByteCost
}

// Equal checks whether other is equal to this NativeToken.
func (n *NativeToken) Equal(other *NativeToken) bool {
	if n.ID != other.ID {
		return false
	}
	return n.Amount.Cmp(other.Amount) == 0
}

func (n *NativeToken) Size() int {
	// amount = 32 bytes(uint256)
	return NativeTokenIDLength + serializer.UInt256ByteSize
}
