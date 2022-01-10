package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
)

// SimpleTokenScheme is a token scheme which checks that the TokenTag within a foundry matches
// the token ID of native tokens held by the foundry.
type SimpleTokenScheme struct{}

func (s *SimpleTokenScheme) Clone() TokenScheme {
	return &SimpleTokenScheme{}
}

func (s *SimpleTokenScheme) VByteCost(costStruct *RentStructure, override VByteCostFunc) uint64 {
	return costStruct.VBFactorKey.With(costStruct.VBFactorData).Multiply(serializer.OneByte)
}

func (s *SimpleTokenScheme) Type() TokenSchemeType {
	return TokenSchemeSimple
}

func (s *SimpleTokenScheme) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(TokenSchemeSimple), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize simple token scheme: %w", err)
		}).
		Done()
}

func (s *SimpleTokenScheme) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(TokenSchemeSimple, func(err error) error {
			return fmt.Errorf("unable to serialize simple token scheme type ID: %w", err)
		}).
		Serialize()
}

func (s *SimpleTokenScheme) MarshalJSON() ([]byte, error) {
	jSimpleTokenScheme := &jsonSimpleTokenScheme{
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
