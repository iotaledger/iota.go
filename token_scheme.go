package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer"
)

// TokenSchemeType defines the type of token schemes.
type TokenSchemeType byte

const (
	// TokenSchemeSimple denotes a type of output which is locked by a signature and deposits onto a single address.
	TokenSchemeSimple TokenSchemeType = iota
)

// TokenScheme defines a scheme for to be used for an OutputFoundry.
type TokenScheme interface {
	serializer.Serializable

	// Type returns the type of the TokenScheme.
	Type() TokenSchemeType
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

// TokenSchemeSelector implements SerializableSelectorFunc for token scheme types.
func TokenSchemeSelector(tokenSchemeType uint32) (serializer.Serializable, error) {
	var seri serializer.Serializable
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
