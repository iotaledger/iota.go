package iotago

import (
	"bytes"
	"math"
	"sort"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
)

// BlockIssuanceCredits defines the type of block issuance credits.
type BlockIssuanceCredits int64

const MaxBlockIssuanceCredits = BlockIssuanceCredits(math.MaxInt64)

// Allotment is a struct that represents a list of account IDs and an allotted value.
type Allotment struct {
	AccountID AccountID `serix:""`
	Mana      Mana      `serix:""`
}

func (a *Allotment) Clone() *Allotment {
	return &Allotment{
		AccountID: a.AccountID,
		Mana:      a.Mana,
	}
}

func (a *Allotment) Equal(other *Allotment) bool {
	return a.AccountID == other.AccountID && a.Mana == other.Mana
}

func (a *Allotment) Compare(other *Allotment) int {
	return bytes.Compare(a.AccountID[:], other.AccountID[:])
}

// Allotments is a slice of Allotment.
type Allotments []*Allotment

func (a Allotments) Clone() Allotments {
	return lo.CloneSlice(a)
}

func (a Allotments) Equal(other Allotments) bool {
	if len(a) != len(other) {
		return false
	}

	for idx, allotment := range a {
		if !allotment.Equal(other[idx]) {
			return false
		}
	}

	return true
}

// Sort sorts the allotments in lexical order.
func (a Allotments) Sort() {
	sort.Slice(a, func(i, j int) bool {
		return bytes.Compare(a[i].AccountID[:], a[j].AccountID[:]) < 0
	})
}

func (a Allotments) Size() int {
	// LengthPrefixType
	return serializer.UInt16ByteSize + len(a)*(AccountIDLength+ManaSize)
}

func (a Allotments) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	// Allotments requires invocation of account managers, so requires extra work.
	workScoreAllotments, err := workScoreParameters.Allotment.Multiply(len(a))
	if err != nil {
		return 0, err
	}

	return workScoreAllotments, nil
}

func (a Allotments) Get(id AccountID) Mana {
	for _, allotment := range a {
		if allotment.AccountID == id {
			return allotment.Mana
		}
	}

	return 0
}
