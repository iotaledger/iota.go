package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer"
)

// PayloadType denotes a type of a payload.
type PayloadType uint32

const (
	// PayloadTransaction denotes a Transaction.
	PayloadTransaction PayloadType = iota
	// PayloadMilestone denotes a Milestone.
	PayloadMilestone
	// PayloadIndexation denotes an Indexation.
	PayloadIndexation
	// PayloadReceipt denotes a Receipt.
	PayloadReceipt
	// PayloadTreasuryTransaction denotes a TreasuryTransaction.
	PayloadTreasuryTransaction
)

// Payload is an object which can be embedded into other objects.
type Payload interface {
	serializer.Serializable

	// PayloadType returns the type of the payload.
	PayloadType() PayloadType
}

// PayloadTypeToString returns the name of a Payload given the type.
func PayloadTypeToString(ty PayloadType) string {
	switch ty {
	case PayloadTransaction:
		return "Transaction"
	case PayloadMilestone:
		return "Milestone"
	case PayloadIndexation:
		return "Indexation"
	case PayloadReceipt:
		return "Receipt"
	case PayloadTreasuryTransaction:
		return "TreasuryTransaction"
	default:
		return "unknown payload"
	}
}

// PayloadSelector implements SerializableSelectorFunc for payload types.
func PayloadSelector(payloadType uint32) (serializer.Serializable, error) {
	var seri serializer.Serializable
	switch PayloadType(payloadType) {
	case PayloadTransaction:
		seri = &Transaction{}
	case PayloadMilestone:
		seri = &Milestone{}
	case PayloadIndexation:
		seri = &Indexation{}
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
	case PayloadIndexation:
		obj = &jsonIndexation{}
	default:
		return nil, fmt.Errorf("unable to decode payload type from JSON: %w", ErrUnknownPayloadType)
	}
	return obj, nil
}
