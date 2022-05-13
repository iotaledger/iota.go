package iotago

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	// ErrNonUniqueMilestoneOpts gets returned when multiple MilestoneOpt(s) with the same MilestoneOptType exist within sets.
	ErrNonUniqueMilestoneOpts = errors.New("non unique milestone options")
)

// MilestoneOptType defines the type of milestone options.
type MilestoneOptType byte

const (
	// MilestoneOptReceipt denotes a ReceiptMilestoneOpt milestone option.
	MilestoneOptReceipt MilestoneOptType = 0
	// MilestoneOptProtocolParams denotes a ProtocolParams milestone option.
	MilestoneOptProtocolParams MilestoneOptType = 1
)

var (
	// ErrTypeIsNotSupportedMilestoneOpt gets returned when a serializable was found to not be a supported MilestoneOpt.
	ErrTypeIsNotSupportedMilestoneOpt = errors.New("serializable is not a supported milestone option")
	msOptNames                        = [MilestoneOptProtocolParams + 1]string{"MilestoneOptReceipt", "MilestoneOptProtocolParams"}
)

func (msOptType MilestoneOptType) String() string {
	if int(msOptType) >= len(msOptNames) {
		return fmt.Sprintf("unknown milestone option type: %d", msOptType)
	}
	return msOptNames[msOptType]
}

// MilestoneOpt is an object carried within a Milestone.
type MilestoneOpt interface {
	serializer.SerializableWithSize

	// Type returns the type of the MilestoneOpt.
	Type() MilestoneOptType

	// Clone clones the MilestoneOpt.
	Clone() MilestoneOpt
}

// MilestoneOptSelector implements SerializableSelectorFunc for milestone options.
func MilestoneOptSelector(msOptType uint32) (MilestoneOpt, error) {
	var seri MilestoneOpt
	switch MilestoneOptType(msOptType) {
	case MilestoneOptReceipt:
		seri = &ReceiptMilestoneOpt{}
	case MilestoneOptProtocolParams:
		seri = &ProtocolParamsMilestoneOpt{}
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownMilestoneOptType, msOptType)
	}
	return seri, nil
}

// selects the json object for the given type.
func jsonMilestoneOptSelector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch MilestoneOptType(ty) {
	case MilestoneOptReceipt:
		obj = &jsonReceiptMilestoneOpt{}
	case MilestoneOptProtocolParams:
		obj = &jsonProtocolParamsMilestoneOpt{}
	default:
		return nil, fmt.Errorf("unable to decode milestone option type from JSON: %w", ErrUnknownMilestoneOptType)
	}
	return obj, nil
}

func milestoneOptsFromJSONRawMsg(jMilestoneOpts []*json.RawMessage) (MilestoneOpts, error) {
	opts, err := jsonRawMsgsToSerializables(jMilestoneOpts, jsonMilestoneOptSelector)
	if err != nil {
		return nil, err
	}
	var msOpts MilestoneOpts
	msOpts.FromSerializables(opts)
	return msOpts, nil
}

// MilestoneOpts is a slice of MilestoneOpt(s).
type MilestoneOpts []MilestoneOpt

func (m MilestoneOpts) ToSerializables() serializer.Serializables {
	seris := make(serializer.Serializables, len(m))
	for i, x := range m {
		seris[i] = x.(serializer.Serializable)
	}
	return seris
}

func (m *MilestoneOpts) FromSerializables(seris serializer.Serializables) {
	*m = make(MilestoneOpts, len(seris))
	for i, seri := range seris {
		(*m)[i] = seri.(MilestoneOpt)
	}
}

func (m MilestoneOpts) Size() int {
	sum := serializer.OneByte // 1 byte length prefix
	for _, opt := range m {
		sum += opt.Size()
	}
	return sum
}

// Set converts the slice into a MilestoneOptSet.
// Returns an error if a MilestoneOpt occurs multiple times.
func (m MilestoneOpts) Set() (MilestoneOptSet, error) {
	set := make(MilestoneOptSet)
	for _, opt := range m {
		if _, has := set[opt.Type()]; has {
			return nil, ErrNonUniqueMilestoneOpts
		}
		set[opt.Type()] = opt
	}
	return set, nil
}

// MustSet works like Set but panics if an error occurs.
// This function is therefore only safe to be called when it is given,
// that a MilestoneOpts slice does not contain the same MilestoneOptType multiple times.
func (m MilestoneOpts) MustSet() MilestoneOptSet {
	set, err := m.Set()
	if err != nil {
		panic(err)
	}
	return set
}

// MilestoneOptTypeSet is a set of MilestoneOptType.
type MilestoneOptTypeSet map[MilestoneOptType]struct{}

// checks whether the given Serializable is a MilestoneOpt and also supported MilestoneOptType.
func msOptWriteGuard(supportedMsOpts MilestoneOptTypeSet) serializer.SerializableWriteGuardFunc {
	return func(seri serializer.Serializable) error {
		if seri == nil {
			return fmt.Errorf("%w: because nil", ErrTypeIsNotSupportedMilestoneOpt)
		}

		obj, is := seri.(MilestoneOpt)
		if !is {
			return fmt.Errorf("%w: because not milestone option", ErrTypeIsNotSupportedMilestoneOpt)
		}

		if _, supported := supportedMsOpts[obj.Type()]; !supported {
			return fmt.Errorf("%w: because not in set %v", ErrTypeIsNotSupportedMilestoneOpt, supported)
		}

		return nil
	}
}

func msOptReadGuard(supportedMsOpts MilestoneOptTypeSet) serializer.SerializableReadGuardFunc {
	return func(ty uint32) (serializer.Serializable, error) {
		if _, supported := supportedMsOpts[MilestoneOptType(ty)]; !supported {
			return nil, fmt.Errorf("%w: because not in set %v (%d)", ErrTypeIsNotSupportedMilestoneOpt, supportedMsOpts, ty)
		}
		return MilestoneOptSelector(ty)
	}
}

// MilestoneOptSet is a set of MilestoneOpt(s).
type MilestoneOptSet map[MilestoneOptType]MilestoneOpt

// Clone clones the FeaturesSet.
func (set MilestoneOptSet) Clone() MilestoneOptSet {
	cpy := make(MilestoneOptSet, len(set))
	for k, v := range set {
		cpy[k] = v.Clone()
	}
	return cpy
}

// Receipt returns the ReceiptMilestoneOpt in the set or nil.
func (set MilestoneOptSet) Receipt() *ReceiptMilestoneOpt {
	b, has := set[MilestoneOptReceipt]
	if !has {
		return nil
	}
	return b.(*ReceiptMilestoneOpt)
}

// ProtocolParams returns the ProtocolParamsMilestoneOpt in the set or nil.
func (set MilestoneOptSet) ProtocolParams() *ProtocolParamsMilestoneOpt {
	b, has := set[MilestoneOptProtocolParams]
	if !has {
		return nil
	}
	return b.(*ProtocolParamsMilestoneOpt)
}
