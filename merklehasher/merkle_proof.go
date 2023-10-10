//nolint:golint // false positives
package merklehasher

import (
	"bytes"
	"context"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

type hashable[V Value] interface {
	hashWithHasher(hasher *Hasher[V]) []byte
}

// valueHash contains the hash of the value for which the proof is being computed.
type valueHash[V Value] struct {
	Hash []byte `serix:"0,mapKey=hash,lengthPrefixType=uint8"`
}

// leafHash contains the hash of a leaf in the tree.
type leafHash[V Value] struct {
	Hash []byte `serix:"0,mapKey=hash,lengthPrefixType=uint8"`
}

// Pair contains the hashes of the left and right children of a node in the tree.
type Pair[V Value] struct {
	Left  hashable[V] `serix:"0,mapKey=l"`
	Right hashable[V] `serix:"1,mapKey=r"`
}

// nolint: tagliatelle // Does not understand generics
type Proof[V Value] struct {
	*Pair[V] `serix:"0"`
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
	return &Proof[V]{Pair: p.(*Pair[V])}, nil
}

func (t *Hasher[V]) computeProof(data [][]byte, index int) (hashable[V], error) {
	if len(data) < 2 {
		leaf := data[0]

		return &valueHash[V]{t.hashLeaf(leaf)}, nil
	}

	if len(data) == 2 {
		left := data[0]
		right := data[1]
		if index == 0 {
			return &Pair[V]{
				Left:  &valueHash[V]{t.hashLeaf(left)},
				Right: &leafHash[V]{t.hashLeaf(right)},
			}, nil
		}

		return &Pair[V]{
			Left:  &leafHash[V]{t.hashLeaf(left)},
			Right: &valueHash[V]{t.hashLeaf(right)},
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

		return &Pair[V]{
			Left:  left,
			Right: &leafHash[V]{right},
		}, nil
	}

	// Inside right half
	left := t.Hash(data[:k])
	right, err := t.computeProof(data[k:], index-k)
	if err != nil {
		return nil, err
	}

	return &Pair[V]{
		Left:  &leafHash[V]{left},
		Right: right,
	}, nil
}

func (l *valueHash[V]) hashWithHasher(_ *Hasher[V]) []byte {
	return l.Hash
}

func (h *leafHash[V]) hashWithHasher(_ *Hasher[V]) []byte {
	return h.Hash
}

func (p *Pair[V]) hashWithHasher(hasher *Hasher[V]) []byte {
	return hasher.hashNode(p.Left.hashWithHasher(hasher), p.Right.hashWithHasher(hasher))
}

func (p *Proof[V]) Hash(hasher *Hasher[V]) []byte {
	return p.Pair.hashWithHasher(hasher)
}

func containsValueHash[V Value](hasheable hashable[V], hashedValue []byte) bool {
	switch t := hasheable.(type) {
	case *leafHash[V]:
		return false
	case *valueHash[V]:
		return bytes.Equal(hashedValue, t.Hash)
	case *Pair[V]:
		return containsValueHash[V](t.Right, hashedValue) || containsValueHash[V](t.Left, hashedValue)
	}

	return false
}

func (p *Proof[V]) ContainsValue(value V, hasher *Hasher[V]) (bool, error) {
	valueBytes, err := value.Bytes()
	if err != nil {
		return false, err
	}

	return containsValueHash[V](p.Pair, hasher.hashLeaf(valueBytes)), nil
}

func serixAPI[V Value]() *serix.API {
	must := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	api := serix.NewAPI()

	must(api.RegisterTypeSettings(valueHash[V]{},
		serix.TypeSettings{}.WithObjectType(uint8(2)),
	))

	must(api.RegisterTypeSettings(leafHash[V]{},
		serix.TypeSettings{}.WithObjectType(uint8(1)),
	))

	must(api.RegisterTypeSettings(Pair[V]{},
		serix.TypeSettings{}.WithObjectType(uint8(0)),
	))

	must(api.RegisterInterfaceObjects((*hashable[V])(nil), (*valueHash[V])(nil)))
	must(api.RegisterInterfaceObjects((*hashable[V])(nil), (*leafHash[V])(nil)))
	must(api.RegisterInterfaceObjects((*hashable[V])(nil), (*Pair[V])(nil)))

	return api
}

func (p *Proof[V]) JSONEncode() ([]byte, error) {
	return serixAPI[V]().JSONEncode(context.TODO(), p)
}

func ProofFromJSON[V Value](bytes []byte) (*Proof[V], error) {
	p := new(Proof[V])
	if err := serixAPI[V]().JSONDecode(context.TODO(), bytes, p); err != nil {
		return nil, err
	}

	return p, nil
}

func ProofFromBytes[V Value](bytes []byte) (*Proof[V], int, error) {
	p := new(Proof[V])
	count, err := serixAPI[V]().Decode(context.TODO(), bytes, p)
	if err != nil {
		return nil, 0, err
	}

	return p, count, nil
}

func (p *Proof[V]) Bytes() ([]byte, error) {
	return serixAPI[V]().Encode(context.TODO(), p)
}
