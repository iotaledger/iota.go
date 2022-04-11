package iotago

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
)

// PayloadType denotes a type of payload.
type PayloadType uint32

const (
	// Deprecated payload types
	// PayloadTransactionTIP7 = 0
	// PayloadMilestoneTIP8 = 1
	// PayloadIndexationTIP6 = 2
	// PayloadReceiptTIP17TIP8 = 3

	// PayloadTreasuryTransaction denotes a TreasuryTransaction.
	PayloadTreasuryTransaction PayloadType = 4
	// PayloadTaggedData denotes a TaggedData payload.
	PayloadTaggedData PayloadType = 5
	// PayloadTransaction denotes a Transaction.
	PayloadTransaction PayloadType = 6
	// PayloadMilestone denotes a Milestone.
	PayloadMilestone PayloadType = 7
)

func (payloadType PayloadType) String() string {
	if int(payloadType) >= len(payloadNames) {
		return fmt.Sprintf("unknown payload type: %d", payloadType)
	}
	return payloadNames[payloadType]
}

var (
	payloadNames = [PayloadMilestone + 1]string{
		"Deprecated-TransactionTIP7",
		"Deprecated-MilestoneTIP8",
		"Deprecated-IndexationTIP6",
		"Deprecated-ReceiptTIP17TIP8",
		"TreasuryTransaction",
		"TaggedData",
		"Transaction",
		"Milestone",
	}
)

var (
	// ErrTypeIsNotSupportedPayload gets returned when a serializable was found to not be a supported Payload.
	ErrTypeIsNotSupportedPayload = errors.New("serializable is not a supported payload")
)

// Payload is an object which can be embedded into other objects.
type Payload interface {
	serializer.SerializableWithSize

	// PayloadType returns the type of the payload.
	PayloadType() PayloadType
}

// PayloadSelector implements SerializableSelectorFunc for payload types.
func PayloadSelector(payloadType uint32) (serializer.Serializable, error) {
	var seri serializer.Serializable
	switch PayloadType(payloadType) {
	case PayloadTransaction:
		seri = &Transaction{}
	case PayloadMilestone:
		seri = &Milestone{}
	case PayloadTaggedData:
		seri = &TaggedData{}
	case PayloadTreasuryTransaction:
		seri = &TreasuryTransaction{}
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownPayloadType, payloadType)
	}
	return seri, nil
}

func payloadFromJSONRawMsg(jPayload *json.RawMessage) (Payload, error) {
	jsonPayload, err := DeserializeObjectFromJSON(jPayload, jsonPayloadSelector)
	if err != nil {
		return nil, err
	}

	payload, err := jsonPayload.ToSerializable()
	if err != nil {
		return nil, err
	}
	return payload.(Payload), nil
}

// selects the json object for the given type.
func jsonPayloadSelector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch PayloadType(ty) {
	case PayloadTransaction:
		obj = &jsonTransaction{}
	case PayloadMilestone:
		obj = &jsonMilestone{}
	case PayloadTaggedData:
		obj = &jsonTaggedData{}
	default:
		return nil, fmt.Errorf("unable to decode payload type from JSON: %w", ErrUnknownPayloadType)
	}
	return obj, nil
}
