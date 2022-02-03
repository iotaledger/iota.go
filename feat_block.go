package iotago

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	// ErrNonUniqueFeatureBlocks gets returned when multiple FeatureBlock(s) with the same FeatureBlockType exist within sets.
	ErrNonUniqueFeatureBlocks = errors.New("non unique feature blocks within outputs")
	// ErrInvalidFeatureBlockTransition gets returned when a FeatureBlock's transition within a ChainConstrainedOutput is invalid.
	ErrInvalidFeatureBlockTransition = errors.New("invalid feature block transition")
	// ErrImmutableFeatureBlocksMutated gets returned when an immutable FeatureBlocks is mutated between chain transitions.
	ErrImmutableFeatureBlocksMutated = errors.New("invalid feature block transition")
)

// FeatureBlockType defines the type of feature blocks.
type FeatureBlockType byte

const (
	// FeatureBlockSender denotes a SenderFeatureBlock.
	FeatureBlockSender FeatureBlockType = iota
	// FeatureBlockIssuer denotes an IssuerFeatureBlock.
	FeatureBlockIssuer
	// FeatureBlockMetadata denotes a MetadataFeatureBlock.
	FeatureBlockMetadata
	// FeatureBlockTag denotes a TagFeatureBlock.
	FeatureBlockTag
)

func (featBlockType FeatureBlockType) String() string {
	if int(featBlockType) >= len(featBlockNames) {
		return fmt.Sprintf("unknown feature block type: %d", featBlockType)
	}
	return featBlockNames[featBlockType]
}

var (
	featBlockNames = [FeatureBlockTag + 1]string{
		"SenderFeatureBlock", "IssuerFeatureBlock", "MetadataFeatureBlock", "TagFeatureBlock",
	}
)

// FeatureBlocks is a slice of FeatureBlock(s).
type FeatureBlocks []FeatureBlock

// Clone clones the FeatureBlocks.
func (f FeatureBlocks) Clone() FeatureBlocks {
	cpy := make(FeatureBlocks, len(f))
	for i, v := range f {
		cpy[i] = v.Clone()
	}
	return cpy
}

func (f FeatureBlocks) VByteCost(costStruct *RentStructure, _ VByteCostFunc) uint64 {
	var sumCost uint64
	for _, featBlock := range f {
		sumCost += featBlock.VByteCost(costStruct, nil)
	}

	// length prefix + sum cost of blocks
	return costStruct.VBFactorData.Multiply(serializer.OneByte) + sumCost
}

func (f FeatureBlocks) ToSerializables() serializer.Serializables {
	seris := make(serializer.Serializables, len(f))
	for i, x := range f {
		seris[i] = x.(serializer.Serializable)
	}
	return seris
}

func (f *FeatureBlocks) FromSerializables(seris serializer.Serializables) {
	*f = make(FeatureBlocks, len(seris))
	for i, seri := range seris {
		(*f)[i] = seri.(FeatureBlock)
	}
}

// Set converts the slice into a FeatureBlocksSet.
// Returns an error if a FeatureBlockType occurs multiple times.
func (f FeatureBlocks) Set() (FeatureBlocksSet, error) {
	set := make(FeatureBlocksSet)
	for _, block := range f {
		if _, has := set[block.Type()]; has {
			return nil, ErrNonUniqueFeatureBlocks
		}
		set[block.Type()] = block
	}
	return set, nil
}

// MustSet works like Set but panics if an error occurs.
// This function is therefore only safe to be called when it is given,
// that a FeatureBlocks slice does not contain the same FeatureBlockType multiple times.
func (f FeatureBlocks) MustSet() FeatureBlocksSet {
	set, err := f.Set()
	if err != nil {
		panic(err)
	}
	return set
}

// Equal checks whether this slice is equal to other.
func (f FeatureBlocks) Equal(other FeatureBlocks) bool {
	if len(f) != len(other) {
		return false
	}
	for i, aBlock := range f {
		if !aBlock.Equal(other[i]) {
			return false
		}
	}
	return true
}

// FeatureBlocksSet is a set of FeatureBlock(s).
type FeatureBlocksSet map[FeatureBlockType]FeatureBlock

// Clone clones the FeatureBlockSet.
func (f FeatureBlocksSet) Clone() FeatureBlocksSet {
	cpy := make(FeatureBlocksSet, len(f))
	for k, v := range f {
		cpy[k] = v.Clone()
	}
	return cpy
}

// SenderFeatureBlock returns the SenderFeatureBlock in the set or nil.
func (f FeatureBlocksSet) SenderFeatureBlock() *SenderFeatureBlock {
	b, has := f[FeatureBlockSender]
	if !has {
		return nil
	}
	return b.(*SenderFeatureBlock)
}

// IssuerFeatureBlock returns the IssuerFeatureBlock in the set or nil.
func (f FeatureBlocksSet) IssuerFeatureBlock() *IssuerFeatureBlock {
	b, has := f[FeatureBlockIssuer]
	if !has {
		return nil
	}
	return b.(*IssuerFeatureBlock)
}

// MetadataFeatureBlock returns the MetadataFeatureBlock in the set or nil.
func (f FeatureBlocksSet) MetadataFeatureBlock() *MetadataFeatureBlock {
	b, has := f[FeatureBlockMetadata]
	if !has {
		return nil
	}
	return b.(*MetadataFeatureBlock)
}

// TagFeatureBlock returns the TagFeatureBlock in the set or nil.
func (f FeatureBlocksSet) TagFeatureBlock() *TagFeatureBlock {
	b, has := f[FeatureBlockTag]
	if !has {
		return nil
	}
	return b.(*TagFeatureBlock)
}

// EveryTuple runs f for every key which exists in both this set and other.
// Returns a bool indicating whether all element of this set existed on the other set.
func (f FeatureBlocksSet) EveryTuple(other FeatureBlocksSet, fun func(a FeatureBlock, b FeatureBlock) error) (bool, error) {
	hadAll := true
	for ty, blockA := range f {
		blockB, has := other[ty]
		if !has {
			hadAll = false
			continue
		}
		if err := fun(blockA, blockB); err != nil {
			return false, err
		}
	}
	return hadAll, nil
}

// FeatureBlock is an abstract building block extending the features of an Output.
type FeatureBlock interface {
	serializer.Serializable
	NonEphemeralObject

	// Type returns the type of the FeatureBlock.
	Type() FeatureBlockType

	// Equal tells whether this FeatureBlock is equal to other.
	Equal(other FeatureBlock) bool

	// Clone clones the FeatureBlock.
	Clone() FeatureBlock
}

// FeatureBlockSelector implements SerializableSelectorFunc for feature blocks.
func FeatureBlockSelector(featBlockType uint32) (FeatureBlock, error) {
	var seri FeatureBlock
	switch FeatureBlockType(featBlockType) {
	case FeatureBlockSender:
		seri = &SenderFeatureBlock{}
	case FeatureBlockIssuer:
		seri = &IssuerFeatureBlock{}
	case FeatureBlockMetadata:
		seri = &MetadataFeatureBlock{}
	case FeatureBlockTag:
		seri = &TagFeatureBlock{}
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownFeatureBlockType, featBlockType)
	}
	return seri, nil
}

// selects the json object for the given type.
func jsonFeatureBlockSelector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch FeatureBlockType(ty) {
	case FeatureBlockSender:
		obj = &jsonSenderFeatureBlock{}
	case FeatureBlockIssuer:
		obj = &jsonIssuerFeatureBlock{}
	case FeatureBlockMetadata:
		obj = &jsonMetadataFeatureBlock{}
	case FeatureBlockTag:
		obj = &jsonTagFeatureBlock{}
	default:
		return nil, fmt.Errorf("unable to decode feature block type from JSON: %w", ErrUnknownFeatureBlockType)
	}
	return obj, nil
}

func featureBlocksFromJSONRawMsg(jFeatureBlocks []*json.RawMessage) (FeatureBlocks, error) {
	blocks, err := jsonRawMsgsToSerializables(jFeatureBlocks, jsonFeatureBlockSelector)
	if err != nil {
		return nil, err
	}
	var featureBlocks FeatureBlocks
	featureBlocks.FromSerializables(blocks)
	return featureBlocks, nil
}
