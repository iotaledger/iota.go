package iotago

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	// ErrNonUniqueFeatures gets returned when multiple Feature(s) with the same FeatureType exist within sets.
	ErrNonUniqueFeatures = errors.New("non unique features within outputs")
	// ErrInvalidFeatureTransition gets returned when a Feature's transition within a ChainConstrainedOutput is invalid.
	ErrInvalidFeatureTransition = errors.New("invalid feature transition")
)

// FeatureType defines the type of features.
type FeatureType byte

const (
	// FeatureSender denotes a SenderFeature.
	FeatureSender FeatureType = iota
	// FeatureIssuer denotes an IssuerFeature.
	FeatureIssuer
	// FeatureMetadata denotes a MetadataFeature.
	FeatureMetadata
	// FeatureTag denotes a TagFeature.
	FeatureTag
)

func (featType FeatureType) String() string {
	if int(featType) >= len(featNames) {
		return fmt.Sprintf("unknown feature type: %d", featType)
	}
	return featNames[featType]
}

var (
	featNames = [FeatureTag + 1]string{
		"SenderFeature", "IssuerFeature", "MetadataFeature", "TagFeature",
	}
)

// Features is a slice of Feature(s).
type Features []Feature

// Clone clones the Features.
func (f Features) Clone() Features {
	cpy := make(Features, len(f))
	for i, v := range f {
		cpy[i] = v.Clone()
	}
	return cpy
}

func (f Features) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	var sumCost uint64
	for _, feat := range f {
		sumCost += feat.VBytes(rentStruct, nil)
	}

	// length prefix + sum cost of features
	return rentStruct.VBFactorData.Multiply(serializer.OneByte) + sumCost
}

func (f Features) ByteSizeKey() uint64 {
	var size uint64 = 0
	for _, feat := range f {
		size += feat.ByteSizeKey()
	}
	return size
}

func (f Features) ByteSizeData() uint64 {
	// length prefix
	var size uint64 = serializer.OneByte
	for _, feat := range f {
		size += feat.ByteSizeData()
	}
	return size
}

func (f Features) ToSerializables() serializer.Serializables {
	seris := make(serializer.Serializables, len(f))
	for i, x := range f {
		seris[i] = x.(serializer.Serializable)
	}
	return seris
}

func (f *Features) FromSerializables(seris serializer.Serializables) {
	*f = make(Features, len(seris))
	for i, seri := range seris {
		(*f)[i] = seri.(Feature)
	}
}

func (f Features) Size() int {
	sum := serializer.OneByte // 1 byte length prefix
	for _, feat := range f {
		sum += feat.Size()
	}
	return sum
}

// Set converts the slice into a FeatureSet.
// Returns an error if a FeatureType occurs multiple times.
func (f Features) Set() (FeatureSet, error) {
	set := make(FeatureSet)
	for _, feat := range f {
		if _, has := set[feat.Type()]; has {
			return nil, ErrNonUniqueFeatures
		}
		set[feat.Type()] = feat
	}
	return set, nil
}

// MustSet works like Set but panics if an error occurs.
// This function is therefore only safe to be called when it is given,
// that a Features slice does not contain the same FeatureType multiple times.
func (f Features) MustSet() FeatureSet {
	set, err := f.Set()
	if err != nil {
		panic(err)
	}
	return set
}

// Equal checks whether this slice is equal to other.
func (f Features) Equal(other Features) bool {
	if len(f) != len(other) {
		return false
	}
	for i, feat := range f {
		if !feat.Equal(other[i]) {
			return false
		}
	}
	return true
}

// FeatureSet is a set of Feature(s).
type FeatureSet map[FeatureType]Feature

// Clone clones the FeatureSet.
func (f FeatureSet) Clone() FeatureSet {
	cpy := make(FeatureSet, len(f))
	for k, v := range f {
		cpy[k] = v.Clone()
	}
	return cpy
}

// SenderFeature returns the SenderFeature in the set or nil.
func (f FeatureSet) SenderFeature() *SenderFeature {
	b, has := f[FeatureSender]
	if !has {
		return nil
	}
	return b.(*SenderFeature)
}

// IssuerFeature returns the IssuerFeature in the set or nil.
func (f FeatureSet) IssuerFeature() *IssuerFeature {
	b, has := f[FeatureIssuer]
	if !has {
		return nil
	}
	return b.(*IssuerFeature)
}

// MetadataFeature returns the MetadataFeature in the set or nil.
func (f FeatureSet) MetadataFeature() *MetadataFeature {
	b, has := f[FeatureMetadata]
	if !has {
		return nil
	}
	return b.(*MetadataFeature)
}

// TagFeature returns the TagFeature in the set or nil.
func (f FeatureSet) TagFeature() *TagFeature {
	b, has := f[FeatureTag]
	if !has {
		return nil
	}
	return b.(*TagFeature)
}

// EveryTuple runs f for every key which exists in both this set and other.
// Returns a bool indicating whether all element of this set existed on the other set.
func (f FeatureSet) EveryTuple(other FeatureSet, fun func(a Feature, b Feature) error) (bool, error) {
	hadAll := true
	for ty, featA := range f {
		featB, has := other[ty]
		if !has {
			hadAll = false
			continue
		}
		if err := fun(featA, featB); err != nil {
			return false, err
		}
	}
	return hadAll, nil
}

// Feature is an abstract building block extending the features of an Output.
type Feature interface {
	serializer.SerializableWithSize
	NonEphemeralObject

	// Type returns the type of the Feature.
	Type() FeatureType

	// Equal tells whether this Feature is equal to other.
	Equal(other Feature) bool

	// Clone clones the Feature.
	Clone() Feature
}

// FeatureSelector implements SerializableSelectorFunc for features.
func FeatureSelector(featType uint32) (Feature, error) {
	var seri Feature
	switch FeatureType(featType) {
	case FeatureSender:
		seri = &SenderFeature{}
	case FeatureIssuer:
		seri = &IssuerFeature{}
	case FeatureMetadata:
		seri = &MetadataFeature{}
	case FeatureTag:
		seri = &TagFeature{}
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownFeatureType, featType)
	}
	return seri, nil
}

// selects the json object for the given type.
func jsonFeatureSelector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch FeatureType(ty) {
	case FeatureSender:
		obj = &jsonSenderFeature{}
	case FeatureIssuer:
		obj = &jsonIssuerFeature{}
	case FeatureMetadata:
		obj = &jsonMetadataFeature{}
	case FeatureTag:
		obj = &jsonTagFeature{}
	default:
		return nil, fmt.Errorf("unable to decode feature type from JSON: %w", ErrUnknownFeatureType)
	}
	return obj, nil
}

func featuresFromJSONRawMsg(jFeatures []*json.RawMessage) (Features, error) {
	feats, err := jsonRawMsgsToSerializables(jFeatures, jsonFeatureSelector)
	if err != nil {
		return nil, err
	}
	var features Features
	features.FromSerializables(feats)
	return features, nil
}
