package giota

import ()

const (
	SIZE_IN_BYTES = 49
)

var (
	NullHashTrits = make([]int, 243, 243)
)

type Hash struct {
	ts []int
}

func HashFromTrits(in []int) *Hash {
	c := NewCurl()
	c.Absorb(in)
	return &Hash{ts: c.Squeeze()}
}

func (h *Hash) String() string {
	return TritsToTrytes(h.Trits())
}

func (h *Hash) Trits() []int {
	return h.ts
}

func (h *Hash) Equal(g Hash) bool {
	return EqualTrits(h.Trits(), g.Trits())
}
