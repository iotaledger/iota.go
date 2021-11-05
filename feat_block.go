package iotago

import (
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
type FeatureBlockType = byte

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
func FeatureBlockTypeToString(ty uint32) string {
	switch byte(ty) {
	case FeatureBlockSender:
		return "FeatureBlockSender"
	case FeatureBlockIssuer:
		return "FeatureBlockIssuer"
	case FeatureBlockReturn:
		return "FeatureBlockReturn"
	case FeatureBlockTimelockMilestoneIndex:
		return "FeatureBlockTimelockMilestoneIndex"
	case FeatureBlockTimelockUnix:
		return "FeatureBlockTimelockUnix"
	case FeatureBlockExpirationMilestoneIndex:
		return "FeatureBlockExpirationMilestoneIndex"
	case FeatureBlockExpirationUnix:
		return "FeatureBlockExpirationUnix"
	case FeatureBlockMetadata:
		return "FeatureBlockMetadata"
	}
	return ""
}

// FeatureBlock is an abstract building block extending the features of an Output.
type FeatureBlock interface {
	serializer.Serializable
	// Type returns the type of the FeatureBlock.
	Type() FeatureBlockType
}

func featureBlockSupported(seris serializer.Serializables, f func(ty uint32) bool) error {
	for i, seri := range seris {
		featBlock, isFeatureBlock := seri.(FeatureBlock)
		if !isFeatureBlock {
			return fmt.Errorf("%w: element at %d is not a feature block", ErrUnsupportedObjectType, i)
		}
		if !f(uint32(featBlock.Type())) {
			return fmt.Errorf("%w: element at %d with type %T", ErrUnsupportedFeatureBlockType, i, seri)
		}
	}
	return nil
}

// FeatureBlockSelector implements SerializableSelectorFunc for feature blocks.
func FeatureBlockSelector(featBlockType uint32) (serializer.Serializable, error) {
	var seri serializer.Serializable
	switch byte(featBlockType) {
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
	switch byte(ty) {
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
