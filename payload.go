package iotago

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
)

// PayloadType denotes a type of a payload.
type PayloadType uint32

const (
	// PayloadTransaction denotes a Transaction.
	PayloadTransaction PayloadType = iota
	// PayloadMilestone denotes a Milestone.
	PayloadMilestone
	// PayloadReceipt denotes a Receipt.
	PayloadReceipt
	// PayloadTreasuryTransaction denotes a TreasuryTransaction.
	PayloadTreasuryTransaction
	// PayloadTaggedData denotes a TaggedData payload.
	PayloadTaggedData PayloadType = 5
)

func (payloadType PayloadType) String() string {
	if int(payloadType) >= len(payloadNames) {
		return fmt.Sprintf("unknown payload type: %d", payloadType)
	}
	return payloadNames[payloadType]
}

var (
	payloadNames = [PayloadTaggedData + 1]string{
		"Transaction",
		"Milestone",
		"Receipt",
		"TreasuryTransaction",
		"Deprecated-Indexation",
		"TaggedData",
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
	case PayloadReceipt:
		seri = &Receipt{}
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
