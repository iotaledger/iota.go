package iotago

import (
	"errors"
)

var (
	// ErrUnsupportedPayloadType gets returned for unsupported payload types.
	ErrUnsupportedPayloadType = errors.New("unsupported payload type")
	// ErrUnsupportedObjectType gets returned for unsupported object types.
	ErrUnsupportedObjectType = errors.New("unsupported object type")
	// ErrUnknownPayloadType gets returned for unknown payload types.
	ErrUnknownPayloadType = errors.New("unknown payload type")
	// ErrUnknownAddrType gets returned for unknown address types.
	ErrUnknownAddrType = errors.New("unknown address type")
	// ErrUnknownInputType gets returned for unknown input types.
	ErrUnknownInputType = errors.New("unknown input type")
	// ErrUnknownOutputType gets returned for unknown output types.
	ErrUnknownOutputType = errors.New("unknown output type")
	// ErrUnknownTransactionEssenceType gets returned for unknown transaction essence types.
	ErrUnknownTransactionEssenceType = errors.New("unknown transaction essence type")
	// ErrUnknownUnlockBlockType gets returned for unknown unlock blocks.
	ErrUnknownUnlockBlockType = errors.New("unknown unlock block type")
	// ErrUnknownSignatureType gets returned for unknown signature types.
	ErrUnknownSignatureType = errors.New("unknown signature type")
)
