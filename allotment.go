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
	AccountID AccountID `serix:"0"`
	Value     Mana      `serix:"1"`
}

func (a *Allotment) Clone() *Allotment {
	return &Allotment{
		AccountID: a.AccountID,
		Value:     a.Value,
	}
}

// Allotments is a slice of Allotment.
type Allotments []*Allotment

func (a Allotments) Clone() Allotments {
	return lo.CloneSlice(a)
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

func (a Allotments) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
	// Allotments requires invocation of account managers, so requires extra work.
	workScoreAllotments, err := workScoreStructure.Allotment.Multiply(len(a))
	if err != nil {
		return 0, err
	}

	return workScoreAllotments, nil
}

func (a Allotments) Get(id AccountID) Mana {
	for _, allotment := range a {
		if allotment.AccountID == id {
			return allotment.Value
		}
	}

	return 0
}
