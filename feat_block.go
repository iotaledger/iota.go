package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer"
)

const (
	// MinFeatBlockCount min number of feature blocks in an output.
	MinFeatBlockCount = 0
	// MaxFeatBlockCount max number of feature blocks in an output.
	MaxFeatBlockCount = 9
)

var (
	featBlockArrayRules = &serializer.ArrayRules{
		Min: MinFeatBlockCount,
		Max: MaxFeatBlockCount,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}
)

// FeatureBlockType defines the type of feature blocks.
type FeatureBlockType byte

const (
	// FeatureBlockSender denotes a SenderFeatureBlock.
	FeatureBlockSender FeatureBlockType = iota
	// FeatureBlockIssuer denotes an IssuerFeatureBlock.
	FeatureBlockIssuer
	// FeatureBlockReturn denotes a ReturnFeatureBlock.
	FeatureBlockReturn
	// FeatureBlockTimelockMilestoneIndex denotes a TimelockMilestoneIndexFeatureBlock.
	FeatureBlockTimelockMilestoneIndex
	// FeatureBlockTimelockUnix denotes a TimelockUnixFeatureBlock.
	FeatureBlockTimelockUnix
	// FeatureBlockExpirationMilestoneIndex denotes an ExpirationMilestoneIndexFeatureBlock.
	FeatureBlockExpirationMilestoneIndex
	// FeatureBlockExpirationUnix denotes an ExpirationUnixFeatureBlock.
	FeatureBlockExpirationUnix
	// FeatureBlockMetadata denotes a MetadataFeatureBlock.
	FeatureBlockMetadata
	// FeatureBlockIndexation denotes an IndexationFeatureBlock.
	FeatureBlockIndexation
)

// FeatureBlockTypeToString returns the name of a FeatureBlock given the type.
func FeatureBlockTypeToString(ty FeatureBlockType) string {
	switch ty {
	case FeatureBlockSender:
		return "SenderFeatureBlock"
	case FeatureBlockIssuer:
		return "IssuerFeatureBlock"
	case FeatureBlockReturn:
		return "ReturnFeatureBlock"
	case FeatureBlockTimelockMilestoneIndex:
		return "TimelockMilestoneIndexFeatureBlock"
	case FeatureBlockTimelockUnix:
		return "TimelockUnixFeatureBlock"
	case FeatureBlockExpirationMilestoneIndex:
		return "ExpirationMilestoneIndexFeatureBlock"
	case FeatureBlockExpirationUnix:
		return "ExpirationUnixFeatureBlock"
	case FeatureBlockMetadata:
		return "MetadataFeatureBlock"
	}
	return "unknown feature block"
}

type FeatureBlocks []FeatureBlock

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

// FeatureBlock is an abstract building block extending the features of an Output.
type FeatureBlock interface {
	serializer.Serializable

	// Type returns the type of the FeatureBlock.
	Type() FeatureBlockType
}

func featureBlockSupported(featBlocks FeatureBlocks, f func(ty uint32) bool) error {
	for i, featBlock := range featBlocks {
		if !f(uint32(featBlock.Type())) {
			return fmt.Errorf("%w: element at %d with type %T", ErrUnsupportedFeatureBlockType, i, featBlock)
		}
	}
	return nil
}

// FeatureBlockSelector implements SerializableSelectorFunc for feature blocks.
func FeatureBlockSelector(featBlockType uint32) (serializer.Serializable, error) {
	var seri serializer.Serializable
	switch FeatureBlockType(featBlockType) {
	case FeatureBlockSender:
		seri = &SenderFeatureBlock{}
	case FeatureBlockIssuer:
		seri = &IssuerFeatureBlock{}
	case FeatureBlockReturn:
		seri = &ReturnFeatureBlock{}
	case FeatureBlockTimelockMilestoneIndex:
		seri = &TimelockMilestoneIndexFeatureBlock{}
	case FeatureBlockTimelockUnix:
		seri = &TimelockUnixFeatureBlock{}
	case FeatureBlockExpirationMilestoneIndex:
		seri = &ExpirationMilestoneIndexFeatureBlock{}
	case FeatureBlockExpirationUnix:
		seri = &ExpirationUnixFeatureBlock{}
	case FeatureBlockMetadata:
		seri = &MetadataFeatureBlock{}
	case FeatureBlockIndexation:
		seri = &IndexationFeatureBlock{}
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
	case FeatureBlockReturn:
		obj = &jsonReturnFeatureBlock{}
	case FeatureBlockTimelockMilestoneIndex:
		obj = &jsonTimelockMilestoneIndexFeatureBlock{}
	case FeatureBlockTimelockUnix:
		obj = &jsonTimelockUnixFeatureBlock{}
	case FeatureBlockExpirationMilestoneIndex:
		obj = &jsonExpirationMilestoneIndexFeatureBlock{}
	case FeatureBlockExpirationUnix:
		obj = &jsonExpirationUnixFeatureBlock{}
	case FeatureBlockMetadata:
		obj = &jsonMetadataFeatureBlock{}
	case FeatureBlockIndexation:
		obj = &jsonIndexationFeatureBlock{}
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
