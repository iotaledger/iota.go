package iotago

import (
	"github.com/iotaledger/hive.go/ierrors"
)

var (
	// ErrUnsupportedInputType gets returned for unsupported input types.
	ErrUnsupportedInputType = ierrors.New("unsupported input type")
	// ErrUnknownAddrType gets returned for unknown address types.
	ErrUnknownAddrType = ierrors.New("unknown address type")
	// ErrUnknownInputType gets returned for unknown input types.
	ErrUnknownInputType = ierrors.New("unknown input type")
	// ErrUnknownOutputType gets returned for unknown output types.
	ErrUnknownOutputType = ierrors.New("unknown output type")
	// ErrUnknownTransactionEssenceType gets returned for unknown transaction essence types.
	ErrUnknownTransactionEssenceType = ierrors.New("unknown transaction essence type")
	// ErrUnknownUnlockType gets returned for unknown unlock.
	ErrUnknownUnlockType = ierrors.New("unknown unlock type")
	//ErrUnexpectedUnderlyingType gets returned for unknown input type of transaction.
	ErrUnexpectedUnderlyingType = ierrors.New("unexpected underlying type provided by the interface")
)
