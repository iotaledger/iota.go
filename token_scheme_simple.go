package iotago

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"

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
	MintedTokens *big.Int
	// The amount of tokens which has been melted.
	MeltedTokens *big.Int
	// The maximum supply of tokens controlled.
	MaximumSupply *big.Int
}

func (s *SimpleTokenScheme) Clone() TokenScheme {
	return &SimpleTokenScheme{
		MintedTokens:  new(big.Int).Set(s.MintedTokens),
		MeltedTokens:  new(big.Int).Set(s.MeltedTokens),
		MaximumSupply: new(big.Int).Set(s.MaximumSupply),
	}
}

func (s *SimpleTokenScheme) VByteCost(costStruct *RentStructure, _ VByteCostFunc) uint64 {
	return costStruct.VBFactorData.Multiply(serializer.OneByte) +
		// minted/melted supply, max. supply
		costStruct.VBFactorData.Multiply(Uint256ByteSize+Uint256ByteSize+Uint256ByteSize)
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

func (s *SimpleTokenScheme) Deserialize(data []byte, _ serializer.DeSerializationMode, _ interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(TokenSchemeSimple), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize simple token scheme: %w", err)
		}).
		ReadUint256(&s.MintedTokens, func(err error) error {
			return fmt.Errorf("unable to deserialize minted tokens for simple token scheme: %w", err)
		}).
		ReadUint256(&s.MeltedTokens, func(err error) error {
			return fmt.Errorf("unable to deserialize melted tokens for simple token scheme: %w", err)
		}).
		ReadUint256(&s.MaximumSupply, func(err error) error {
			return fmt.Errorf("unable to deserialize maximum supply for simple token scheme: %w", err)
		}).
		Done()
}

func (s *SimpleTokenScheme) Serialize(_ serializer.DeSerializationMode, _ interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(byte(TokenSchemeSimple), func(err error) error {
			return fmt.Errorf("unable to serialize simple token scheme type ID: %w", err)
		}).
		WriteUint256(s.MintedTokens, func(err error) error {
			return fmt.Errorf("unable to serialize simple token scheme minted tokens: %w", err)
		}).
		WriteUint256(s.MeltedTokens, func(err error) error {
			return fmt.Errorf("unable to serialize simple token scheme melted tokens: %w", err)
		}).
		WriteUint256(s.MaximumSupply, func(err error) error {
			return fmt.Errorf("unable to serialize simple token scheme maximum supply: %w", err)
		}).
		Serialize()
}

func (s *SimpleTokenScheme) Size() int {
	return util.NumByteLen(byte(TokenSchemeSimple)) +
		util.NumByteLen(s.MintedTokens) +
		util.NumByteLen(s.MeltedTokens) +
		util.NumByteLen(s.MaximumSupply)
}

func (s *SimpleTokenScheme) MarshalJSON() ([]byte, error) {
	jSimpleTokenScheme := &jsonSimpleTokenScheme{
		Type: int(TokenSchemeSimple),
	}
	jSimpleTokenScheme.MintedSupply = EncodeUint256(s.MintedTokens)
	jSimpleTokenScheme.MeltedTokens = EncodeUint256(s.MeltedTokens)
	jSimpleTokenScheme.MaximumSupply = EncodeUint256(s.MaximumSupply)
	return json.Marshal(jSimpleTokenScheme)
}

func (s *SimpleTokenScheme) UnmarshalJSON(bytes []byte) error {
	jSimpleTokenScheme := &jsonSimpleTokenScheme{}
	if err := json.Unmarshal(bytes, jSimpleTokenScheme); err != nil {
		return err
	}
	seri, err := jSimpleTokenScheme.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*SimpleTokenScheme)
	return nil
}

// jsonSimpleTokenScheme defines the json representation of a SimpleTokenScheme.
type jsonSimpleTokenScheme struct {
	Type          int    `json:"type"`
	MintedSupply  string `json:"mintedTokens"`
	MeltedTokens  string `json:"meltedTokens"`
	MaximumSupply string `json:"maximumSupply"`
}

func (j *jsonSimpleTokenScheme) ToSerializable() (serializer.Serializable, error) {
	var err error
	s := &SimpleTokenScheme{}
	s.MintedTokens, err = DecodeUint256(j.MintedSupply)
	if err != nil {
		return nil, fmt.Errorf("%w: minted tokens field of simple token scheme '%s'", ErrDecodeJSONUint256Str, j.MintedSupply)
	}

	s.MeltedTokens, err = DecodeUint256(j.MeltedTokens)
	if err != nil {
		return nil, fmt.Errorf("%w: melted tokens field of simple token scheme '%s'", ErrDecodeJSONUint256Str, j.MintedSupply)
	}

	s.MaximumSupply, err = DecodeUint256(j.MaximumSupply)
	if err != nil {
		return nil, fmt.Errorf("%w: maximum supply field of simple token scheme '%s', inner err %s", ErrDecodeJSONUint256Str, j.MaximumSupply, err)
	}

	return s, nil
}
