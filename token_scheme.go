package iotago

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
)

// TokenSchemeType defines the type of token schemes.
type TokenSchemeType byte

const (
	// TokenSchemeSimple denotes a type of output which is locked by a signature and deposits onto a single address.
	TokenSchemeSimple TokenSchemeType = iota
)

var (
	// ErrTypeIsNotSupportedTokenScheme gets returned when a serializable was found to not be a supported TokenScheme.
	ErrTypeIsNotSupportedTokenScheme = errors.New("serializable is not an address")
)

// TokenScheme defines a scheme for to be used for an OutputFoundry.
type TokenScheme interface {
	serializer.Serializable
	NonEphemeralObject

	// Type returns the type of the TokenScheme.
	Type() TokenSchemeType

	// Clone clones the TokenScheme.
	Clone() TokenScheme
}

// TokenSchemeTypeToString returns a name for the given TokenScheme type.
func TokenSchemeTypeToString(ty TokenSchemeType) string {
	switch ty {
	case TokenSchemeSimple:
		return "SimpleTokenScheme"
	default:
		return "unknown token scheme"
	}
}

// checks whether the given serializable is a TokenScheme.
func tokenSchemeWriteGuard(seri serializer.Serializable) error {
	if seri == nil {
		return fmt.Errorf("%w: because nil", ErrTypeIsNotSupportedTokenScheme)
	}
	switch seri.(type) {
	case *SimpleTokenScheme:
	default:
		return ErrTypeIsNotSupportedTokenScheme
	}
	return nil
}

func wrappedTokenSchemeSelector(tokenSchemeType uint32) (serializer.Serializable, error) {
	return TokenSchemeSelector(tokenSchemeType)
}

// TokenSchemeSelector implements SerializableSelectorFunc for token scheme types.
func TokenSchemeSelector(tokenSchemeType uint32) (TokenScheme, error) {
	var seri TokenScheme
	switch TokenSchemeType(tokenSchemeType) {
	case TokenSchemeSimple:
		seri = &SimpleTokenScheme{}
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownTokenSchemeType, tokenSchemeType)
	}
	return seri, nil
}

func tokenSchemeFromJSONRawMsg(jTokenScheme *json.RawMessage) (TokenScheme, error) {
	tokenScheme, err := DeserializeObjectFromJSON(jTokenScheme, jsonTokenSchemeSelector)
	if err != nil {
		return nil, fmt.Errorf("unable to decode token scheme from JSON: %w", err)
	}
	return tokenScheme.(TokenScheme), nil
}

// jsonTokenSchemeSelector selects the json token scheme implementation for the given type.
func jsonTokenSchemeSelector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch TokenSchemeType(ty) {
	case TokenSchemeSimple:
		obj = &jsonSimpleTokenScheme{}
	default:
		return nil, fmt.Errorf("unable to decode token scheme type from JSON: %w", ErrUnknownTokenSchemeType)
	}
	return obj, nil
}
