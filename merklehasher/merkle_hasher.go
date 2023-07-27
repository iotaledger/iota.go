package merklehasher

import (
	"crypto"
	"math/bits"

	"github.com/iotaledger/hive.go/serializer/v2"
)

// Domain separation prefixes.
const (
	LeafHashPrefix = 0
	NodeHashPrefix = 1
)

type Value interface {
	serializer.Byter
}

// Hasher implements the hashing algorithm described in the IOTA protocol RFC-12.
type Hasher[V Value] struct {
	//nolint:structcheck // false positive
	hash crypto.Hash
}

// NewHasher creates a new Hasher using the provided hash function.
func NewHasher[V Value](h crypto.Hash) *Hasher[V] {
	return &Hasher[V]{hash: h}
}

// Size returns the length, in bytes, of a digest resulting from the given hash function.
func (t *Hasher[V]) Size() int {
	return t.hash.Size()
}

// EmptyRoot returns a special case for an empty tree.
// This is equivalent to Hash(nil).
func (t *Hasher[V]) EmptyRoot() []byte {
	return t.hash.New().Sum(nil)
}

// HashValues computes the Merkle tree hash of the provided BlockIDs.
func (t *Hasher[V]) HashValues(values []V) ([]byte, error) {
	data := make([][]byte, len(values))
	for i := range values {
		value, err := values[i].Bytes()
		if err != nil {
			panic(err)
		}
		data[i] = value
	}

	return t.Hash(data), nil
}

// Hash computes the Merkle tree hash of the provided data.
func (t *Hasher[V]) Hash(data [][]byte) []byte {
	if len(data) == 0 {
		return t.EmptyRoot()
	}
	if len(data) == 1 {
		l := data[0]

		return t.hashLeaf(l)
	}

	k := largestPowerOfTwo(len(data))
	l := t.Hash(data[:k])
	r := t.Hash(data[k:])

	return t.hashNode(l, r)
}

// hashLeaf returns the Merkle tree leafValue hash of data.
func (t *Hasher[V]) hashLeaf(l []byte) []byte {
	h := t.hash.New()
	h.Write([]byte{LeafHashPrefix})
	h.Write(l)

	return h.Sum(nil)
}

// hashNode returns the inner Merkle tree node hash of the two child nodes l and r.
func (t *Hasher[V]) hashNode(l, r []byte) []byte {
	h := t.hash.New()
	h.Write([]byte{NodeHashPrefix})
	h.Write(l)
	h.Write(r)

	return h.Sum(nil)
}

// largestPowerOfTwo returns the largest power of two less than n.
func largestPowerOfTwo(x int) int {
	if x < 2 {
		panic("invalid value")
	}

	return 1 << (bits.Len(uint(x-1)) - 1)
}
