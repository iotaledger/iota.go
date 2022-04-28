package iotago

import (
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
	msOptNames = [MilestoneOptProtocolParams + 1]string{"MilestoneOptReceipt", "MilestoneOptProtocolParams"}
)

func (msOptType MilestoneOptType) String() string {
	if int(msOptType) >= len(msOptNames) {
		return fmt.Sprintf("unknown milestone option type: %d", msOptType)
	}
	return msOptNames[msOptType]
}

// MilestoneOpt is an object carried within a Milestone.
type MilestoneOpt interface {
	Sizer

	// Type returns the type of the MilestoneOpt.
	Type() MilestoneOptType

	// Clone clones the MilestoneOpt.
	Clone() MilestoneOpt
}

// MilestoneOpts is a slice of MilestoneOpt(s).
type MilestoneOpts []MilestoneOpt

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

// MilestoneOptSet is a set of MilestoneOpt(s).
type MilestoneOptSet map[MilestoneOptType]MilestoneOpt

// Clone clones the FeatureSet.
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
