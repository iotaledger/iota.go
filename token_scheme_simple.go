package iotago

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

var (
	// ErrSimpleTokenSchemeTransition gets returned when a SimpleTokenScheme transition is invalid.
	ErrSimpleTokenSchemeTransition = errors.New("simple token scheme transition invalid")
	// ErrSimpleTokenSchemeInvalidMaximumSupply gets returned when a SimpleTokenScheme's max supply is invalid.
	ErrSimpleTokenSchemeInvalidMaximumSupply = errors.New("simple token scheme's maximum supply is invalid")
	// ErrSimpleTokenSchemeInvalidMintedMeltedTokens gets returned when a SimpleTokenScheme's minted supply is invalid.
	ErrSimpleTokenSchemeInvalidMintedMeltedTokens = errors.New("simple token scheme's minted/melted tokens counters are invalid")
)

// SimpleTokenScheme is a TokenScheme which works with minted/melted/maximum supply counters.
type SimpleTokenScheme struct {
	// The amount of tokens which has been minted.
	MintedTokens *big.Int `serix:"0,mapKey=mintedTokens"`
	// The amount of tokens which has been melted.
	MeltedTokens *big.Int `serix:"1,mapKey=meltedTokens"`
	// The maximum supply of tokens controlled.
	MaximumSupply *big.Int `serix:"2,mapKey=maximumSupply"`
}

func (s *SimpleTokenScheme) Clone() TokenScheme {
	return &SimpleTokenScheme{
		MintedTokens:  new(big.Int).Set(s.MintedTokens),
		MeltedTokens:  new(big.Int).Set(s.MeltedTokens),
		MaximumSupply: new(big.Int).Set(s.MaximumSupply),
	}
}

func (s *SimpleTokenScheme) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	return rentStruct.VBFactorData.Multiply(serializer.OneByte) +
		// minted/melted supply, max. supply
		rentStruct.VBFactorData.Multiply(Uint256ByteSize+Uint256ByteSize+Uint256ByteSize)
}

func (s *SimpleTokenScheme) Type() TokenSchemeType {
	return TokenSchemeSimple
}

func (s *SimpleTokenScheme) SyntacticalValidation() error {
	if r := s.MaximumSupply.Cmp(common.Big0); r != 1 {
		return fmt.Errorf("%w: less than equal zero", ErrSimpleTokenSchemeInvalidMaximumSupply)
	}

	// minted - melted > 0: can never have melted more than minted
	mintedMeltedDelta := big.NewInt(0).Sub(s.MintedTokens, s.MeltedTokens)
	if r := mintedMeltedDelta.Cmp(common.Big0); r == -1 {
		return fmt.Errorf("%w: minted/melted delta less than zero: %s", ErrSimpleTokenSchemeInvalidMintedMeltedTokens, mintedMeltedDelta)
	}

	// minted - melted <= max supply: can never have minted more than max supply
	if r := mintedMeltedDelta.Cmp(s.MaximumSupply); r == 1 {
		return fmt.Errorf("%w: minted/melted delta more than maximum supply: %s (delta) vs. %s (max supply)", ErrSimpleTokenSchemeInvalidMintedMeltedTokens, mintedMeltedDelta, s.MaximumSupply)
	}

	return nil
}

func (s *SimpleTokenScheme) StateTransition(transType ChainTransitionType, nextState TokenScheme, in *big.Int, out *big.Int) error {
	switch transType {
	case ChainTransitionTypeGenesis:
		return s.genesisValid(out)
	case ChainTransitionTypeStateChange:
		return s.stateChangeValid(nextState, in, out)
	case ChainTransitionTypeDestroy:
		return s.destructionValid(out, in)
	default:
		panic(fmt.Sprintf("invalid transition type in SimpleTokenScheme %d", transType))
	}
}

// checks that the melted tokens are zero on genesis and that the minted token count
// equals the amount of tokens on the output side of the transaction.
func (s *SimpleTokenScheme) genesisValid(outSum *big.Int) error {
	switch {
	case s.MeltedTokens.Cmp(common.Big0) != 0:
		return fmt.Errorf("%w: melted supply must be zero", ErrSimpleTokenSchemeTransition)
	case outSum.Cmp(s.MintedTokens) != 0:
		return fmt.Errorf("%w: genesis requires that output tokens amount equal minted count: minted %s vs. output tokens %s", ErrSimpleTokenSchemeTransition, s.MintedTokens, outSum)
	}
	return nil
}

// SimpleTokenScheme enforces that all tokens that have been minted are melted when the foundry gets destroyed.
func (s *SimpleTokenScheme) destructionValid(out *big.Int, in *big.Int) error {
	tokenDiff := big.NewInt(0).Sub(out, in)
	if big.NewInt(0).Add(s.MintedTokens, tokenDiff).Cmp(s.MeltedTokens) != 0 {
		return fmt.Errorf("%w: all minted tokens must have been melted up on destruction: minted (%s) + token diff (%d) != melted tokens (%s)", ErrNativeTokenSumUnbalanced, s.MintedTokens, tokenDiff, s.MeltedTokens)
	}
	return nil
}

// checks the balance between the in/out tokens and the invariants concerning supply counter changes.
func (s *SimpleTokenScheme) stateChangeValid(nextState TokenScheme, in *big.Int, out *big.Int) error {
	next, is := nextState.(*SimpleTokenScheme)
	if !is {
		return fmt.Errorf("%w: can only transition to same type but got %T instead", ErrSimpleTokenSchemeTransition, nextState)
	}

	switch {
	case s.MaximumSupply.Cmp(next.MaximumSupply) != 0:
		return fmt.Errorf("%w: maximum supply mismatch wanted %s but got %s", ErrSimpleTokenSchemeTransition, s.MaximumSupply, next.MaximumSupply)
	case s.MintedTokens.Cmp(next.MintedTokens) == 1:
		return fmt.Errorf("%w: current minted supply (%s) bigger than next minted supply (%s)", ErrSimpleTokenSchemeTransition, s.MintedTokens, next.MintedTokens)
	case s.MeltedTokens.Cmp(next.MeltedTokens) == 1:
		return fmt.Errorf("%w: current melted supply (%s) bigger than next melted supply (%s)", ErrSimpleTokenSchemeTransition, s.MeltedTokens, next.MeltedTokens)
	}

	var (
		tokenDiff         = big.NewInt(0).Sub(out, in)
		tokenDiffType     = tokenDiff.Cmp(common.Big0)
		mintedSupplyDelta = big.NewInt(0).Sub(next.MintedTokens, s.MintedTokens)
		meltedSupplyDelta = big.NewInt(0).Sub(next.MeltedTokens, s.MeltedTokens)
	)

	switch {
	case tokenDiffType == 1:
		switch {
		case mintedSupplyDelta.Cmp(tokenDiff) != 0:
			// positive token diff requires the minted supply delta to equal the token diff
			return fmt.Errorf("%w: positive token diff not balanced by minted supply change: next minted supply %s - current minted supply %s = %s != token delta %s", ErrNativeTokenSumUnbalanced, next.MintedTokens, s.MintedTokens, mintedSupplyDelta, tokenDiff)
		case next.MeltedTokens.Cmp(s.MeltedTokens) != 0:
			// must not change melted supply while minting
			return fmt.Errorf("%w: positive token diff requires equal melted supply between current/next state: current (melted=%s), next (melted=%s)", ErrNativeTokenSumUnbalanced, s.MeltedTokens, next.MeltedTokens)
		}

	case tokenDiffType == -1:
		switch {
		case meltedSupplyDelta.Cmp(big.NewInt(0).Abs(tokenDiff)) == 1:
			// negative token diff requires the melted supply delta to be equal less than the token diff.
			// can be less than because we support burning and melting at the same time
			return fmt.Errorf("%w: negative token diff not balanced by melted supply change: next melted supply %s - current melted supply %s = %s which is > abs. delta %s", ErrNativeTokenSumUnbalanced, next.MintedTokens, s.MintedTokens, meltedSupplyDelta, tokenDiff)
		case next.MintedTokens.Cmp(s.MintedTokens) != 0:
			// must not change minting supply while melting
			return fmt.Errorf("%w: negative token diff requires equal minted supply between current/next state: current (minted=%s), next (minted=%s)", ErrNativeTokenSumUnbalanced, s.MintedTokens, next.MintedTokens)
		}

	case tokenDiffType == 0:
		switch {
		case s.MintedTokens.Cmp(next.MintedTokens) != 0 || s.MeltedTokens.Cmp(next.MeltedTokens) != 0:
			// no mutations to minted/melted fields while balance is kept
			return fmt.Errorf("%w: zero token diff requires equal minted/melted supply between current/next state: current (minted/melted=%s/%s), next (minted/melted=%s/%s)", ErrNativeTokenSumUnbalanced, s.MintedTokens, s.MeltedTokens, next.MintedTokens, next.MeltedTokens)
		}
	}

	return nil
}

func (s *SimpleTokenScheme) Size() int {
	return util.NumByteLen(byte(TokenSchemeSimple)) +
		util.NumByteLen(s.MintedTokens) +
		util.NumByteLen(s.MeltedTokens) +
		util.NumByteLen(s.MaximumSupply)
}
