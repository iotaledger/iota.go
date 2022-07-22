package iotago

import (
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	// ErrNonUniqueFeatures gets returned when multiple Feature(s) with the same FeatureType exist within sets.
	ErrNonUniqueFeatures = errors.New("non unique features within outputs")
	// ErrInvalidFeatureTransition gets returned when a Feature's transition within a ChainOutput is invalid.
	ErrInvalidFeatureTransition = errors.New("invalid feature transition")
)

// Feature is an abstract building block extending the features of an Output.
type Feature interface {
	Sizer
	NonEphemeralObject

	// Type returns the type of the Feature.
	Type() FeatureType

	// Equal tells whether this Feature is equal to other.
	Equal(other Feature) bool

	// Clone clones the Feature.
	Clone() Feature
}

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
type Features[T Feature] []Feature

// Clone clones the Features.
func (f Features[T]) Clone() Features[T] {
	cpy := make(Features[T], len(f))
	for i, v := range f {
		cpy[i] = v.Clone()
	}
	return cpy
}

func (f Features[T]) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	var sumCost uint64
	for _, feat := range f {
		sumCost += feat.VBytes(rentStruct, nil)
	}

	// length prefix + sum cost of features
	return rentStruct.VBFactorData.Multiply(serializer.OneByte) + sumCost
}

func (f Features[T]) Size() int {
	sum := serializer.OneByte // 1 byte length prefix
	for _, feat := range f {
		sum += feat.Size()
	}
	return sum
}

// Set converts the slice into a FeatureSet.
// Returns an error if a FeatureType occurs multiple times.
func (f Features[T]) Set() (FeatureSet, error) {
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
func (f Features[T]) MustSet() FeatureSet {
	set, err := f.Set()
	if err != nil {
		panic(err)
	}
	return set
}

// Equal checks whether this slice is equal to other.
func (f Features[T]) Equal(other Features[T]) bool {
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

// FeatureUnchanged checks whether the specified Feature type is unchanged between in and out.
// Unchanged also means that the block's existence is unchanged between both sets.
func FeatureUnchanged(featType FeatureType, inFeatSet FeatureSet, outFeatSet FeatureSet) error {
	in, inHas := inFeatSet[featType]
	out, outHas := outFeatSet[featType]

	switch {
	case outHas && !inHas:
		return fmt.Errorf("%w: %s in next state but not in previous", ErrInvalidFeatureTransition, featType)
	case !outHas && inHas:
		return fmt.Errorf("%w: %s in current state but not in next", ErrInvalidFeatureTransition, featType)
	}

	// not in both sets
	if in == nil {
		return nil
	}

	if !in.Equal(out) {
		return fmt.Errorf("%w: %s changed, in %v / out %v", ErrInvalidFeatureTransition, featType, in, out)
	}

	return nil
}
