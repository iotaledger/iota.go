package iotago

import (
	"fmt"
	"math/big"
)

// TokenSchemeType defines the type of token schemes.
type TokenSchemeType byte

const (
	// TokenSchemeSimple denotes a type of output which is locked by a signature and deposits onto a single address.
	TokenSchemeSimple TokenSchemeType = iota
)

func (tokenSchemeType TokenSchemeType) String() string {
	if int(tokenSchemeType) >= len(tokenSchemeNames) {
		return fmt.Sprintf("unknown token scheme type: %d", tokenSchemeType)
	}
	return tokenSchemeNames[tokenSchemeType]
}

var (
	tokenSchemeNames = [TokenSchemeSimple + 1]string{
		"SimpleTokenScheme",
	}
)

// TokenScheme defines a scheme for to be used for an OutputFoundry.
type TokenScheme interface {
	Sizer
	NonEphemeralObject

	// Type returns the type of the TokenScheme.
	Type() TokenSchemeType

	// Clone clones the TokenScheme.
	Clone() TokenScheme

	// StateTransition validates the transition of the token scheme against its new state.
	StateTransition(transType ChainTransitionType, nextState TokenScheme, in *big.Int, out *big.Int) error

	// SyntacticalValidation validates the syntactical rules.
	SyntacticalValidation() error
}
