//nolint:golint // false positives
package merklehasher

import (
	"bytes"
	"context"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

type MerkleHashableType = uint8

const (
	MerkleHashableTypeNode      MerkleHashableType = iota
	MerkleHashableTypeLeafHash  MerkleHashableType = 1
	MerkleHashableTypeValueHash MerkleHashableType = 2
)

var (
	// ErrProofValueNotFound gets returned when the value for which to compute the proof was not found.
	ErrProofValueNotFound = ierrors.New("the value for which to compute the inclusion proof was not found in the supplied list")
)

type MerkleHashable[V Value] interface {
	hash(hasher *Hasher[V]) []byte
}

// ValueHash contains the hash of the value for which the proof is being computed.
type ValueHash[V Value] struct {
	Hash []byte `serix:",lenPrefix=uint8"`
}

// LeafHash contains the hash of a leaf in the tree.
type LeafHash[V Value] struct {
	Hash []byte `serix:",lenPrefix=uint8"`
}

// Node contains the hashes of the left and right children of a node in the tree.
type Node[V Value] struct {
	Left  MerkleHashable[V] `serix:"l"`
	Right MerkleHashable[V] `serix:"r"`
}

// nolint: tagliatelle // Does not understand generics
type Proof[V Value] struct {
	MerkleHashable[V] `serix:",inlined"`
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
		return nil, ierrors.WithMessagef(ErrProofValueNotFound, "value %s is not contained in the given values list", hexutil.EncodeHex(valueToProofBytes))
	}

	return t.ComputeProofForIndex(values, index)
}

// ComputeProofForIndex computes the audit path given the values and the index of the value we want to create the inclusion proof for.
func (t *Hasher[V]) ComputeProofForIndex(values []V, index int) (*Proof[V], error) {
	if len(values) < 1 {
		return nil, ierrors.New("at least one item is needed to create an inclusion proof")
	}
	if index >= len(values) {
		return nil, ierrors.Errorf("index %d out of bounds for 'values' of len %d", index, len(values))
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

	return &Proof[V]{MerkleHashable: p}, nil
}

func (t *Hasher[V]) computeProof(data [][]byte, index int) (MerkleHashable[V], error) {
	if len(data) < 2 {
		leaf := data[0]

		return &ValueHash[V]{t.hashLeaf(leaf)}, nil
	}

	if len(data) == 2 {
		left := data[0]
		right := data[1]
		if index == 0 {
			return &Node[V]{
				Left:  &ValueHash[V]{t.hashLeaf(left)},
				Right: &LeafHash[V]{t.hashLeaf(right)},
			}, nil
		}

		return &Node[V]{
			Left:  &LeafHash[V]{t.hashLeaf(left)},
			Right: &ValueHash[V]{t.hashLeaf(right)},
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

		return &Node[V]{
			Left:  left,
			Right: &LeafHash[V]{right},
		}, nil
	}

	// Inside right half
	left := t.Hash(data[:k])
	right, err := t.computeProof(data[k:], index-k)
	if err != nil {
		return nil, err
	}

	return &Node[V]{
		Left:  &LeafHash[V]{left},
		Right: right,
	}, nil
}

//nolint:unused // False positive
func (l *ValueHash[V]) hash(_ *Hasher[V]) []byte {
	return l.Hash
}

//nolint:unused // False positive
func (h *LeafHash[V]) hash(_ *Hasher[V]) []byte {
	return h.Hash
}

//nolint:unused // False positive
func (p *Node[V]) hash(hasher *Hasher[V]) []byte {
	return hasher.hashNode(p.Left.hash(hasher), p.Right.hash(hasher))
}

func (p *Proof[V]) Hash(hasher *Hasher[V]) []byte {
	return p.MerkleHashable.hash(hasher)
}

func containsValueHash[V Value](hashable MerkleHashable[V], hashedValue []byte) bool {
	switch t := hashable.(type) {
	case *LeafHash[V]:
		return false
	case *ValueHash[V]:
		return bytes.Equal(hashedValue, t.Hash)
	case *Node[V]:
		return containsValueHash[V](t.Right, hashedValue) || containsValueHash[V](t.Left, hashedValue)
	}

	return false
}

func (p *Proof[V]) ContainsValue(value V, hasher *Hasher[V]) (bool, error) {
	valueBytes, err := value.Bytes()
	if err != nil {
		return false, err
	}

	return containsValueHash[V](p.MerkleHashable, hasher.hashLeaf(valueBytes)), nil
}

func RegisterSerixRules[V Value](api *serix.API) {
	must := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	must(api.RegisterTypeSettings(ValueHash[V]{},
		serix.TypeSettings{}.WithObjectType(MerkleHashableTypeValueHash),
	))

	must(api.RegisterTypeSettings(LeafHash[V]{},
		serix.TypeSettings{}.WithObjectType(MerkleHashableTypeLeafHash),
	))

	must(api.RegisterTypeSettings(Node[V]{},
		serix.TypeSettings{}.WithObjectType(MerkleHashableTypeNode),
	))

	must(api.RegisterInterfaceObjects((*MerkleHashable[V])(nil), (*ValueHash[V])(nil)))
	must(api.RegisterInterfaceObjects((*MerkleHashable[V])(nil), (*LeafHash[V])(nil)))
	must(api.RegisterInterfaceObjects((*MerkleHashable[V])(nil), (*Node[V])(nil)))
}

func serixAPI[V Value]() *serix.API {
	api := serix.NewAPI()
	RegisterSerixRules[V](api)
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
