package merklehasher

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/iotaledger/iota.go/v3"
)

type hashable interface {
	Hash(hasher *Hasher) []byte
}

type leafValue struct {
	Value []byte
}

type hashValue struct {
	Value []byte
}

type InclusionProof struct {
	Left  hashable
	Right hashable
}

// ComputeInclusionProof computes the audit path given the blockIDs and the blockID we want to create the inclusion proof for
func (t *Hasher) ComputeInclusionProof(blockIDs iotago.BlockIDs, blockID iotago.BlockID) (*InclusionProof, error) {
	var found bool
	var index int
	for i := range blockIDs {
		if blockID == blockIDs[i] {
			index = i
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("blockID %s is not contained in the given list", blockID.ToHex())
	}
	return t.ComputeInclusionProofForIndex(blockIDs, index)
}

// ComputeInclusionProofForIndex computes the audit path given the blockIDs and the index of the blockID we want to create the inclusion proof for
func (t *Hasher) ComputeInclusionProofForIndex(blockIDs iotago.BlockIDs, index int) (*InclusionProof, error) {
	if len(blockIDs) < 2 {
		return nil, errors.New("you need at lest 2 items to create an inclusion proof")
	}
	if index >= len(blockIDs) {
		return nil, fmt.Errorf("index %d out of bounds len=%d", index, len(blockIDs))
	}

	data := make([][]byte, len(blockIDs))
	for i := range blockIDs {
		data[i] = blockIDs[i][:]
	}

	p, err := t.computeProof(data, index)
	if err != nil {
		return nil, err
	}
	return p.(*InclusionProof), nil
}

func (t *Hasher) computeProof(data [][]byte, index int) (hashable, error) {
	if len(data) < 2 {
		l := data[0]
		return &leafValue{l}, nil
	}

	if len(data) == 2 {
		left := data[0]
		right := data[1]
		if index == 0 {
			return &InclusionProof{
				Left:  &leafValue{left},
				Right: &hashValue{t.hashLeaf(right)},
			}, nil
		} else {
			return &InclusionProof{
				Left:  &hashValue{t.hashLeaf(left)},
				Right: &leafValue{right},
			}, nil
		}
	}

	k := largestPowerOfTwo(len(data))
	if index < k {
		// Inside left half
		left, err := t.computeProof(data[:k], index)
		if err != nil {
			return nil, err
		}
		right := t.Hash(data[k:])
		return &InclusionProof{
			Left:  left,
			Right: &hashValue{right},
		}, nil
	} else {
		// Inside right half
		left := t.Hash(data[:k])
		right, err := t.computeProof(data[k:], index-k)
		if err != nil {
			return nil, err
		}
		return &InclusionProof{
			Left:  &hashValue{left},
			Right: right,
		}, nil
	}
}

func (l *leafValue) Hash(hasher *Hasher) []byte {
	return hasher.hashLeaf(l.Value)
}

type jsonValue struct {
	Value string `json:"value"`
}

func (l *leafValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsonValue{
		Value: iotago.EncodeHex(l.Value),
	})
}

func (l *leafValue) UnmarshalJSON(bytes []byte) error {
	j := &jsonValue{}
	if err := json.Unmarshal(bytes, j); err != nil {
		return err
	}
	if len(j.Value) == 0 {
		return errors.New("missing value")
	}
	value, err := iotago.DecodeHex(j.Value)
	if err != nil {
		return err
	}
	l.Value = value
	return nil
}

func (h *hashValue) Hash(_ *Hasher) []byte {
	return h.Value
}

type jsonHash struct {
	Hash string `json:"h"`
}

func (h *hashValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsonHash{
		Hash: iotago.EncodeHex(h.Value),
	})
}

func (h *hashValue) UnmarshalJSON(bytes []byte) error {
	j := &jsonHash{}
	if err := json.Unmarshal(bytes, j); err != nil {
		return err
	}
	if len(j.Hash) == 0 {
		return errors.New("missing hash")
	}
	value, err := iotago.DecodeHex(j.Hash)
	if err != nil {
		return err
	}
	h.Value = value
	return nil
}

func (p *InclusionProof) Hash(hasher *Hasher) []byte {
	return hasher.hashNode(p.Left.Hash(hasher), p.Right.Hash(hasher))
}

type jsonPath struct {
	Left  *json.RawMessage `json:"l"`
	Right *json.RawMessage `json:"r"`
}

func containsLeafValue(hasheable hashable, value []byte) bool {
	switch t := hasheable.(type) {
	case *hashValue:
		return false
	case *leafValue:
		return bytes.Equal(value, t.Value)
	case *InclusionProof:
		return containsLeafValue(t.Right, value) || containsLeafValue(t.Left, value)
	}
	return false
}

func (p *InclusionProof) ContainsValue(value iotago.BlockID) (bool, error) {
	return containsLeafValue(p, value[:]), nil
}

func (p *InclusionProof) MarshalJSON() ([]byte, error) {
	jsonLeft, err := json.Marshal(p.Left)
	if err != nil {
		return nil, err
	}
	jsonRight, err := json.Marshal(p.Right)
	if err != nil {
		return nil, err
	}
	rawLeft := json.RawMessage(jsonLeft)
	rawRight := json.RawMessage(jsonRight)
	return json.Marshal(&jsonPath{
		Left:  &rawLeft,
		Right: &rawRight,
	})
}

func unmarshalHashable(raw *json.RawMessage, hasheable *hashable) error {
	h := &hashValue{}
	if err := json.Unmarshal(*raw, h); err == nil {
		*hasheable = h
		return nil
	}
	l := &leafValue{}
	if err := json.Unmarshal(*raw, l); err == nil {
		*hasheable = l
		return nil
	}

	p := &InclusionProof{}
	err := json.Unmarshal(*raw, p)
	if err != nil {
		return err
	}
	*hasheable = p
	return nil
}

func (p *InclusionProof) UnmarshalJSON(bytes []byte) error {
	j := &jsonPath{}
	if err := json.Unmarshal(bytes, j); err != nil {
		return err
	}
	var left hashable
	if err := unmarshalHashable(j.Left, &left); err != nil {
		return err
	}
	var right hashable
	if err := unmarshalHashable(j.Right, &right); err != nil {
		return err
	}
	p.Left = left
	p.Right = right
	return nil
}
