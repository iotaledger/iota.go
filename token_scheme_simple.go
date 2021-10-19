package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer"
)

// SimpleTokenScheme is a token scheme which checks that the TokenTag within a foundry matches
// the token ID of native tokens held by the foundry.
type SimpleTokenScheme struct{}

func (s *SimpleTokenScheme) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	return serializer.NewDeserializer(data).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := serializer.CheckTypeByte(data, TokenSchemeSimple); err != nil {
					return fmt.Errorf("unable to deserialize simple token scheme: %w", err)
				}
			}
			return nil
		}).
		Skip(serializer.SmallTypeDenotationByteSize, func(err error) error {
			return fmt.Errorf("unable to skip token scheme type during deserialization: %w", err)
		}).Done()
}

func (s *SimpleTokenScheme) Serialize(deSeriMode serializer.DeSerializationMode) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(TokenSchemeSimple, func(err error) error {
			return fmt.Errorf("unable to serialize simple token scheme type ID: %w", err)
		}).Serialize()
}

func (s *SimpleTokenScheme) MarshalJSON() ([]byte, error) {
	jSimpleTokenScheme := &jsonFoundryOutput{
		Type: int(TokenSchemeSimple),
	}
	return json.Marshal(jSimpleTokenScheme)
}

func (s *SimpleTokenScheme) UnmarshalJSON(bytes []byte) error {
	jSimpleTokenScheme := &jsonSimpleTokenScheme{}
	if err := json.Unmarshal(bytes, jSimpleTokenScheme); err != nil {
		return err
	}
	seri, err := jSimpleTokenScheme.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*SimpleTokenScheme)
	return nil
}

// jsonSimpleTokenScheme defines the json representation of a SimpleTokenScheme.
type jsonSimpleTokenScheme struct {
	Type int `json:"type"`
}

func (j *jsonSimpleTokenScheme) ToSerializable() (serializer.Serializable, error) {
	return &SimpleTokenScheme{}, nil
}
