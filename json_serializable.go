package iotago

import (
	"encoding/json"
	"errors"
	"github.com/iotaledger/hive.go/serializer"
)

var (
	// ErrInvalidJSON gets returned when invalid JSON is tried to get parsed.
	ErrInvalidJSON = errors.New("invalid json")
)

// JSONSerializable is an object which can return a Serializable.
type JSONSerializable interface {
	// ToSerializable returns the Serializable form of the JSONSerializable.
	ToSerializable() (serializer.Serializable, error)
}

// JSONObjectEnvelope defines the envelope for looking-ahead an object's type
// before deserializing it to its actual object.
type JSONObjectEnvelope struct {
	Type int `json:"type"`
}

// JSONSerializableSelectorFunc is a function that given a type int, returns an empty instance of the given underlying type.
// If the type doesn't resolve, an error is returned.
type JSONSerializableSelectorFunc func(ty int) (JSONSerializable, error)

// DeserializeObjectFromJSON reads out the type of the given raw json message,
// then selects the appropriate object type and deserializes the given *json.RawMessage into it.
func DeserializeObjectFromJSON(raw *json.RawMessage, selector JSONSerializableSelectorFunc) (JSONSerializable, error) {
	j, err := raw.MarshalJSON()
	if err != nil {
		return nil, err
	}

	envelope := &JSONObjectEnvelope{}
	if err := json.Unmarshal(j, envelope); err != nil {
		return nil, err
	}

	obj, err := selector(envelope.Type)
	if err != nil {
		return nil, err
	}

	rawJSON, err := raw.MarshalJSON()
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(rawJSON, obj); err != nil {
		return nil, err
	}

	return obj, nil
}
