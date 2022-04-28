package iotago

import (
	"errors"
)

var (
	// ErrUnsupportedInputType gets returned for unsupported input types.
	ErrUnsupportedInputType = errors.New("unsupported input type")
	// ErrUnknownAddrType gets returned for unknown address types.
	ErrUnknownAddrType = errors.New("unknown address type")
	// ErrUnknownInputType gets returned for unknown input types.
	ErrUnknownInputType = errors.New("unknown input type")
	// ErrUnknownOutputType gets returned for unknown output types.
	ErrUnknownOutputType = errors.New("unknown output type")
	// ErrUnknownTransactionEssenceType gets returned for unknown transaction essence types.
	ErrUnknownTransactionEssenceType = errors.New("unknown transaction essence type")
	// ErrUnknownUnlockType gets returned for unknown unlock.
	ErrUnknownUnlockType = errors.New("unknown unlock type")
)
