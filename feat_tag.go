package iotago

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

const (
	// MinTagLength defines the min. length of a tag feature tag.
	MinTagLength = 1
	// MaxTagLength defines the max. length of a tag feature tag.
	MaxTagLength = 64
)

// TagFeature is a feature which allows to additionally tag an output by a user defined value.
type TagFeature struct {
	Tag []byte
}

func (s *TagFeature) Clone() Feature {
	return &TagFeature{Tag: append([]byte(nil), s.Tag...)}
}

func (s *TagFeature) VBytes(rentStruct *RentStructure, f VBytesFunc) VBytes {
	if f != nil {
		return f(rentStruct)
	}
	return rentStruct.VBFactorData.Multiply(VBytes(serializer.SmallTypeDenotationByteSize + serializer.OneByte + len(s.Tag)))
}

func (s *TagFeature) Equal(other Feature) bool {
	otherFeat, is := other.(*TagFeature)
	if !is {
		return false
	}

	return bytes.Equal(s.Tag, otherFeat.Tag)
}

func (s *TagFeature) Type() FeatureType {
	return FeatureTag
}

func (s *TagFeature) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(FeatureTag), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize tag feature: %w", err)
		}).
		ReadVariableByteSlice(&s.Tag, serializer.SeriLengthPrefixTypeAsByte, func(err error) error {
			return fmt.Errorf("unable to deserialize tag for tag feature: %w", err)
		}, MinTagLength, MaxTagLength).
		Done()
}

func (s *TagFeature) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(byte(FeatureTag), func(err error) error {
			return fmt.Errorf("unable to serialize tag feature type ID: %w", err)
		}).
		WriteVariableByteSlice(s.Tag, serializer.SeriLengthPrefixTypeAsByte, func(err error) error {
			return fmt.Errorf("unable to serialize tag feature tag: %w", err)
		}, MinTagLength, MaxTagLength).
		Serialize()
}

func (s *TagFeature) Size() int {
	// tag length prefix = 1 byte
	return util.NumByteLen(byte(FeatureSender)) + serializer.OneByte + len(s.Tag)
}

func (s *TagFeature) MarshalJSON() ([]byte, error) {
	jTagFeat := &jsonTagFeature{}
	jTagFeat.Type = int(FeatureTag)
	jTagFeat.Tag = EncodeHex(s.Tag)
	return json.Marshal(jTagFeat)
}

func (s *TagFeature) UnmarshalJSON(bytes []byte) error {
	jTagFeat := &jsonTagFeature{}
	if err := json.Unmarshal(bytes, jTagFeat); err != nil {
		return err
	}
	seri, err := jTagFeat.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*TagFeature)
	return nil
}

// jsonTagFeature defines the json representation of an TagFeature.
type jsonTagFeature struct {
	Type int    `json:"type"`
	Tag  string `json:"tag"`
}

func (j *jsonTagFeature) ToSerializable() (serializer.Serializable, error) {
	dataBytes, err := DecodeHex(j.Tag)
	if err != nil {
		return nil, fmt.Errorf("unable to decode tag from JSON for tag feature: %w", err)
	}
	return &TagFeature{Tag: dataBytes}, nil
}
