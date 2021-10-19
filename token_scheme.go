package iotago

import (
	"fmt"

	"github.com/iotaledger/hive.go/serializer"
)

// TokenSchemeType defines the type of token schemes.
type TokenSchemeType = byte

const (
	// TokenSchemeSimple denotes a type of output which is locked by a signature and deposits onto a single address.
	TokenSchemeSimple TokenSchemeType = iota
)

// TokenSchemeSelector implements SerializableSelectorFunc for token scheme types.
func TokenSchemeSelector(tokenSchemeType uint32) (serializer.Serializable, error) {
	var seri serializer.Serializable
	switch byte(tokenSchemeType) {
	case TokenSchemeSimple:
		seri = &SimpleTokenScheme{}
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownTokenSchemeType, tokenSchemeType)
	}
	return seri, nil
}

// jsonTokenSchemeSelector selects the json token scheme implementation for the given type.
func jsonTokenSchemeSelector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch byte(ty) {
	case TokenSchemeSimple:
		obj = &jsonSimpleTokenScheme{}
	default:
		return nil, fmt.Errorf("unable to decode token scheme type from JSON: %w", ErrUnknownTokenSchemeType)
	}
	return obj, nil
}
