//nolint:golint // false positives
package merklehasher

import (
	"bytes"
	"encoding/json"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

type hashable[V Value] interface {
	Hash(hasher *Hasher[V]) []byte
}

type leafValue[V Value] struct {
	Value []byte
}

type hashValue[V Value] struct {
	Value []byte
}

type Proof[V Value] struct {
	Left  hashable[V]
	Right hashable[V]
}

// ComputeProof computes the audit path given the values and the value we want to create the inclusion proof for.
func (t *Hasher[V]) ComputeProof(values []V, valueToProof V) (*Proof[V], error) {
	var found bool
	var index int
	valueToProofBytes, err := valueToProof.Bytes()
	if err != nil {
		return nil, err
	}

	for i := range values {
		valueBytes, err := values[i].Bytes()
		if err != nil {
			return nil, err
		}
		if bytes.Equal(valueToProofBytes, valueBytes) {
			index = i
			found = true

			break
		}
	}
	if !found {
		return nil, ierrors.Errorf("value %s is not contained in the given list", hexutil.EncodeHex(valueToProofBytes))
	}

	return t.ComputeProofForIndex(values, index)
}

// ComputeProofForIndex computes the audit path given the values and the index of the value we want to create the inclusion proof for.
func (t *Hasher[V]) ComputeProofForIndex(values []V, index int) (*Proof[V], error) {
	if len(values) < 2 {
		return nil, ierrors.New("you need at lest 2 items to create an inclusion proof")
	}
	if index >= len(values) {
		return nil, ierrors.Errorf("index %d out of bounds len=%d", index, len(values))
	}

	data := make([][]byte, len(values))
	for i := range values {
		valueBytes, err := values[i].Bytes()
		if err != nil {
			return nil, err
		}
		data[i] = valueBytes
	}

	p, err := t.computeProof(data, index)
	if err != nil {
		return nil, err
	}

	//nolint:forcetypeassert
	return p.(*Proof[V]), nil
}

func (t *Hasher[V]) computeProof(data [][]byte, index int) (hashable[V], error) {
	if len(data) < 2 {
		l := data[0]

		return &leafValue[V]{l}, nil
	}

	if len(data) == 2 {
		left := data[0]
		right := data[1]
		if index == 0 {
			return &Proof[V]{
				Left:  &leafValue[V]{left},
				Right: &hashValue[V]{t.hashLeaf(right)},
			}, nil
		}

		return &Proof[V]{
			Left:  &hashValue[V]{t.hashLeaf(left)},
			Right: &leafValue[V]{right},
		}, nil

	}

	k := largestPowerOfTwo(len(data))
	if index < k {
		// Inside left half
		left, err := t.computeProof(data[:k], index)
		if err != nil {
			return nil, err
		}
		right := t.Hash(data[k:])

		return &Proof[V]{
			Left:  left,
			Right: &hashValue[V]{right},
		}, nil
	}

	// Inside right half
	left := t.Hash(data[:k])
	right, err := t.computeProof(data[k:], index-k)
	if err != nil {
		return nil, err
	}

	return &Proof[V]{
		Left:  &hashValue[V]{left},
		Right: right,
	}, nil
}

func (l *leafValue[V]) Hash(hasher *Hasher[V]) []byte {
	return hasher.hashLeaf(l.Value)
}

type jsonValue struct {
	Value string `json:"value"`
}

func (l *leafValue[V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsonValue{
		Value: hexutil.EncodeHex(l.Value),
	})
}

func (l *leafValue[V]) UnmarshalJSON(bytes []byte) error {
	j := &jsonValue{}
	if err := json.Unmarshal(bytes, j); err != nil {
		return err
	}
	if len(j.Value) == 0 {
		return ierrors.New("missing value")
	}
	value, err := hexutil.DecodeHex(j.Value)
	if err != nil {
		return err
	}
	l.Value = value

	return nil
}

func (h *hashValue[V]) Hash(_ *Hasher[V]) []byte {
	return h.Value
}

type jsonHash struct {
	Hash string `json:"h"`
}

func (h *hashValue[V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsonHash{
		Hash: hexutil.EncodeHex(h.Value),
	})
}

func (h *hashValue[V]) UnmarshalJSON(bytes []byte) error {
	j := &jsonHash{}
	if err := json.Unmarshal(bytes, j); err != nil {
		return err
	}
	if len(j.Hash) == 0 {
		return ierrors.New("missing hash")
	}
	value, err := hexutil.DecodeHex(j.Hash)
	if err != nil {
		return err
	}
	h.Value = value

	return nil
}

func (p *Proof[V]) Hash(hasher *Hasher[V]) []byte {
	return hasher.hashNode(p.Left.Hash(hasher), p.Right.Hash(hasher))
}

type jsonPath struct {
	Left  *json.RawMessage `json:"l"`
	Right *json.RawMessage `json:"r"`
}

func containsLeafValue[V Value](hasheable hashable[V], value []byte) bool {
	switch t := hasheable.(type) {
	case *hashValue[V]:
		return false
	case *leafValue[V]:
		return bytes.Equal(value, t.Value)
	case *Proof[V]:
		return containsLeafValue[V](t.Right, value) || containsLeafValue[V](t.Left, value)
	}

	return false
}

func (p *Proof[V]) ContainsValue(value V) (bool, error) {
	valueBytes, err := value.Bytes()
	if err != nil {
		return false, err
	}

	return containsLeafValue[V](p, valueBytes), nil
}

func (p *Proof[V]) MarshalJSON() ([]byte, error) {
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

func unmarshalHashable[V Value](raw *json.RawMessage, hasheable *hashable[V]) error {
	h := new(hashValue[V])
	if err := json.Unmarshal(*raw, h); err == nil {
		*hasheable = h

		return nil
	}
	l := new(leafValue[V])
	if err := json.Unmarshal(*raw, l); err == nil {
		*hasheable = l

		return nil
	}

	p := new(Proof[V])
	if err := json.Unmarshal(*raw, p); err != nil {
		return err
	}
	*hasheable = p

	return nil
}

func (p *Proof[V]) UnmarshalJSON(bytes []byte) error {
	j := &jsonPath{}
	if err := json.Unmarshal(bytes, j); err != nil {
		return err
	}
	var left hashable[V]
	if err := unmarshalHashable(j.Left, &left); err != nil {
		return err
	}
	var right hashable[V]
	if err := unmarshalHashable(j.Right, &right); err != nil {
		return err
	}
	p.Left = left
	p.Right = right

	return nil
}
