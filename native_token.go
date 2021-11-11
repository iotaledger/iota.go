package iotago

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"

	"github.com/iotaledger/hive.go/serializer"
)

const (
	// MinNativeTokenCountPerOutput min number of different native tokens that can reside in one output.
	MinNativeTokenCountPerOutput = 0
	// MaxNativeTokenCountPerOutput max number of different native tokens that can reside in one output.
	MaxNativeTokenCountPerOutput = 256

	// MaxNativeTokensCount is the max number of native tokens which can occur in a transaction (sum input/output side).
	MaxNativeTokensCount = 256

	TokenTagLength = 12

	// NativeTokenIDLength is the byte length of a NativeTokenID consisting out of the FoundryID plus TokenTag.
	NativeTokenIDLength = FoundryIDLength + TokenTagLength
)

var (
	// ErrNativeTokenAmountLessThanEqualZero gets returned when a NativeToken.Amount is not bigger than 0.
	ErrNativeTokenAmountLessThanEqualZero = errors.New("native token must be a value bigger than zero")
	// ErrNativeTokenSumExceedsUint256 gets returned when a NativeToken.Amount addition results in a value bigger than the max value of a uint256.
	ErrNativeTokenSumExceedsUint256 = errors.New("native token sum exceeds max value of a uint256")
	// ErrNativeTokenSumUnbalanced gets returned when two NativeTokenSum(s) are unbalanced.
	ErrNativeTokenSumUnbalanced = errors.New("native token sums are unbalanced")
	nativeTokensArrayRules      = &serializer.ArrayRules{
		Min:            MinNativeTokenCountPerOutput,
		Max:            MaxNativeTokenCountPerOutput,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates | serializer.ArrayValidationModeLexicalOrdering,
	}
)

// NativeTokenID is an identifier which uniquely identifies a NativeToken.
type NativeTokenID [NativeTokenIDLength]byte

func (ntID NativeTokenID) String() string {
	return hex.EncodeToString(ntID[:])
}

// FoundryID returns the FoundryID to which this NativeTokenID belongs to.
func (ntID NativeTokenID) FoundryID() FoundryID {
	var foundryID FoundryID
	copy(foundryID[:], ntID[:])
	return foundryID
}

// NativeTokenSum is a mapping of NativeTokenID to a sum value.
type NativeTokenSum map[NativeTokenID]*big.Int

// Balanced checks whether the set of NativeTokens are balanced between the two NativeTokenSum.
// This function is only appropriate for checking NativeToken balances if there are no underlying foundry state transitions.
func (nts NativeTokenSum) Balanced(other NativeTokenSum) error {
	if len(nts) != len(other) {
		return fmt.Errorf("%w: length mismatch, source %d, other %d", ErrNativeTokenSumUnbalanced, len(nts), len(other))
	}
	for id, sum := range nts {
		otherSum := other[id]
		if otherSum == nil {
			return fmt.Errorf("%w: native token %s missing in other", ErrNativeTokenSumUnbalanced, id)
		}
		if sum.Cmp(otherSum) != 0 {
			return fmt.Errorf("%w: sum mismatch, source %d, other %d", ErrNativeTokenSumUnbalanced, sum, other)
		}
	}
	return nil
}

// BalancedWithDiffs works like Balanced but incorporates the foundry state transitions:
//	- There must not be a specific NativeToken set on the input side if on the output side its foundry is created.
//	- If a specific NativeToken set is balanced, then either no foundry diff is present,
//	  the foundry is destroyed or the foundry transitions with a zero diff.
//	- If a specific NativeToken set is unbalanced then a foundry diff with transition FoundryTransitionStateChange balancing that set must exist.
//	- For all new NativeToken sets on the outputs side, foundry diffs with FoundryTransitionNew balancing those sets must exist.
func (nts NativeTokenSum) BalancedWithDiffs(other NativeTokenSum, diffs FoundryStateDiffs) error {
	for id, srcSum := range nts {
		otherSum := other[id]
		if otherSum == nil {
			return fmt.Errorf("%w: native token %s missing in other", ErrNativeTokenSumUnbalanced, id)
		}

		foundryDiff := diffs[id.FoundryID()]

		// impossible invariant as the tokens already exist on the input side
		if foundryDiff != nil && foundryDiff.Transition == FoundryTransitionNew {
			return fmt.Errorf("%w: foundry %s can not be newly created if its tokens already exist on the inputs", ErrInvalidFoundryTransition, id.FoundryID())
		}

		switch srcSum.Cmp(otherSum) {
		// equal
		case 0:
			// we either have a foundry diff which state transitions with 0 delta or gets destroyed,
			// or we don't have a diff at all
			if foundryDiff == nil || foundryDiff.Transition == FoundryTransitionDestroyed {
				continue
			}
			// here we have a state transition which must be zero
			if foundryDiff.SupplyDiff.Cmp(common.Big0) != 0 {
				return fmt.Errorf("%w: native tokens are balanced on inputs/outputs but foundry %s's circulating supply diff is %s instead of zero", ErrInvalidFoundryTransition, id.FoundryID(), foundryDiff.SupplyDiff)
			}
		default:
			// must have a foundry diff with the appropriate diff
			if foundryDiff == nil {
				return fmt.Errorf("%w: missing foundry %s for native token %s", ErrMissingFoundryTransition, id.FoundryID(), id)
			}

			// impossible invariant as a destroyed foundry can't adjust the circulating supply
			if foundryDiff.Transition == FoundryTransitionDestroyed {
				return fmt.Errorf("%w: foundry %s can not be destroyed if tokens are unbalanced between inputs/outputs", ErrInvalidFoundryTransition, id.FoundryID())
			}

			actualDiff := new(big.Int).Sub(otherSum, srcSum)
			if actualDiff.Cmp(foundryDiff.SupplyDiff) != 0 {
				return fmt.Errorf("%w: foundry %s's circulating supply changes by %s but native token diff between inputs/outputs is %s", ErrNativeTokenSumUnbalanced, id.FoundryID(), foundryDiff.SupplyDiff, actualDiff)
			}
		}
	}

	// check that for newly created foundries the corresponding native tokens exist on the output side
	for foundryID, foundryDiff := range diffs {
		if foundryDiff.Transition != FoundryTransitionNew {
			continue
		}

		otherSum := other[foundryDiff.NativeTokenID]
		if otherSum == nil {
			return fmt.Errorf("%w: new foundry %s's circulating supply is %s but none of the new native tokens reside on the output side", ErrNativeTokenSumUnbalanced, foundryID, foundryDiff.SupplyDiff)
		}

		// since this is a new foundry, the foundry diff equals the actual wanted circulating supply
		if otherSum.Cmp(foundryDiff.SupplyDiff) != 0 {
			return fmt.Errorf("%w: new foundry %s's circulating supply is %s but the native tokens sum on output side is %s", ErrNativeTokenSumUnbalanced, foundryID, foundryDiff.SupplyDiff, otherSum)
		}
	}

	return nil
}

// NativeTokens is a set of NativeToken.
type NativeTokens []*NativeToken

func (n NativeTokens) ToSerializables() serializer.Serializables {
	seris := make(serializer.Serializables, len(n))
	for i, x := range n {
		seris[i] = x
	}
	return seris
}

func (n *NativeTokens) FromSerializables(seris serializer.Serializables) {
	*n = make(NativeTokens, len(seris))
	for i, seri := range seris {
		(*n)[i] = seri.(*NativeToken)
	}
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
	ID     NativeTokenID
	Amount *big.Int
}

// Equal checks whether other is equal to this NativeToken.
func (n *NativeToken) Equal(other *NativeToken) bool {
	if n.ID != other.ID {
		return false
	}
	return n.Amount.Cmp(other.Amount) == 0
}

func (n *NativeToken) Deserialize(data []byte, _ serializer.DeSerializationMode) (int, error) {
	return serializer.NewDeserializer(data).
		ReadBytesInPlace(n.ID[:], func(err error) error {
			return fmt.Errorf("unable to deserialize ID for native token: %w", err)
		}).
		ReadUint256(n.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for native token: %w", err)
		}).
		Done()
}

func (n *NativeToken) Serialize(_ serializer.DeSerializationMode) ([]byte, error) {
	return serializer.NewSerializer().
		WriteBytes(n.ID[:], func(err error) error {
			return fmt.Errorf("unable to serialize native token ID: %w", err)
		}).
		WriteUint256(n.Amount, func(err error) error {
			return fmt.Errorf("unable to serialize native token amount: %w", err)
		}).
		Serialize()
}

func (n *NativeToken) MarshalJSON() ([]byte, error) {
	jNativeToken := &jsonNativeToken{}
	jNativeToken.ID = hex.EncodeToString(n.ID[:])
	jNativeToken.Amount = n.Amount.String()
	return json.Marshal(jNativeToken)
}

func (n *NativeToken) UnmarshalJSON(bytes []byte) error {
	jNativeToken := &jsonNativeToken{}
	if err := json.Unmarshal(bytes, jNativeToken); err != nil {
		return err
	}
	seri, err := jNativeToken.ToSerializable()
	if err != nil {
		return err
	}
	*n = *seri.(*NativeToken)
	return nil
}

func nativeTokensFromJSONRawMsg(jNativeTokens []*json.RawMessage) (NativeTokens, error) {
	tokens, err := jsonRawMsgsToSerializables(jNativeTokens, func(ty int) (JSONSerializable, error) {
		return &jsonNativeToken{}, nil
	})
	if err != nil {
		return nil, err
	}
	var nativeTokens NativeTokens
	nativeTokens.FromSerializables(tokens)
	return nativeTokens, nil
}

// jsonNativeToken defines the json representation of a NativeToken.
type jsonNativeToken struct {
	ID     string `json:"id"`
	Amount string `json:"amount"`
}

func (j *jsonNativeToken) ToSerializable() (serializer.Serializable, error) {
	n := &NativeToken{}

	nftIDBytes, err := hex.DecodeString(j.ID)
	if err != nil {
		return nil, err
	}
	copy(n.ID[:], nftIDBytes)

	var ok bool
	n.Amount, ok = new(big.Int).SetString(j.Amount, 10)
	if !ok {
		return nil, fmt.Errorf("%w: amount field of native token '%s'", ErrDecodeJSONUint256Str, j.Amount)
	}

	return n, nil
}
