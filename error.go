package iotago

import (
	"errors"
)

var (
	// ErrUnsupportedPayloadType gets returned for unsupported payload types.
	ErrUnsupportedPayloadType = errors.New("unsupported payload type")
	// ErrUnsupportedObjectType gets returned for unsupported object types.
	ErrUnsupportedObjectType = errors.New("unsupported object type")
	// ErrUnsupportedInputType gets returned for unsupported input types.
	ErrUnsupportedInputType = errors.New("unsupported input type")
	// ErrUnsupportedFeatureType gets returned when an unsupported feature exists in a set.
	ErrUnsupportedFeatureType = errors.New("unsupported feature type")
	// ErrUnsupportedUnlockConditionType gets returned when an unsupported unlock condition exists in a set.
	ErrUnsupportedUnlockConditionType = errors.New("unsupported unlock condition type")
	// ErrUnsupportedMilestoneOptType gets returned when an unsupported milestone option exists in a set.
	ErrUnsupportedMilestoneOptType = errors.New("unsupported milestone option type")
	// ErrUnknownPayloadType gets returned for unknown payload types.
	ErrUnknownPayloadType = errors.New("unknown payload type")
	// ErrUnknownAddrType gets returned for unknown address types.
	ErrUnknownAddrType = errors.New("unknown address type")
	// ErrUnknownFeatureType gets returned for unknown feature types.
	ErrUnknownFeatureType = errors.New("unknown feature type")
	// ErrUnknownMilestoneOptType gets returned for unknown milestone options types.
	ErrUnknownMilestoneOptType = errors.New("unknown milestone option type")
	// ErrUnknownUnlockConditionType gets returned for unknown unlock condition types.
	ErrUnknownUnlockConditionType = errors.New("unknown unlock condition type")
	// ErrUnknownInputType gets returned for unknown input types.
	ErrUnknownInputType = errors.New("unknown input type")
	// ErrUnknownOutputType gets returned for unknown output types.
	ErrUnknownOutputType = errors.New("unknown output type")
	// ErrUnknownTokenSchemeType gets returned for unknown token scheme types.
	ErrUnknownTokenSchemeType = errors.New("unknown token scheme type")
	// ErrUnknownTransactionEssenceType gets returned for unknown transaction essence types.
	ErrUnknownTransactionEssenceType = errors.New("unknown transaction essence type")
	// ErrUnknownUnlockType gets returned for unknown unlock.
	ErrUnknownUnlockType = errors.New("unknown unlock type")
	// ErrUnknownSignatureType gets returned for unknown signature types.
	ErrUnknownSignatureType = errors.New("unknown signature type")
	// ErrDecodeJSONUint256Str gets returned when an uint256 string could not be decoded to a big.int.
	ErrDecodeJSONUint256Str = errors.New("could not deserialize JSON uint256 string to big.Int")
)
