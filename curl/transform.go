package curl

import (
	"math"

	"github.com/iotaledger/iota.go/consts"
)

const rotationOffset = 364

// stateRotations stores the chunk offset and the bit shift of the state after each round.
// Since the modulo operations are rather costly, they are pre-computed.
var stateRotations [NumRounds]struct {
	offset, shift uint
}

func init() {
	var rotation uint = rotationOffset
	for r := 0; r < NumRounds; r++ {
		// the state is organized as chunks of 243 trits each
		stateRotations[r].offset = rotation / consts.HashTrinarySize
		stateRotations[r].shift = rotation % consts.HashTrinarySize
		rotation = (rotation * rotationOffset) % StateSize // the rotation offset is applied every round
	}
}

// transform performs the Curl transformation.
// According to the specification, one Curl round performs the following transformation:
//   for i ← 1 to 729
//     x ← S[1]
//     S ← rot(S)
//     y ← S[1]
//     N[i] ← g(x,y)
//   S ← N
// Each element of the state S is combined with its rotated counterpart using the S-box g.
// This is equivalent to rotating just once and applying the S-box on the entire state:
//   N ← rot(S)
//   S ← g(S,N)
// The only difference then is, that the trits are at the wrong position. Successive trits are now an opposite rotation
// apart. This rotation offset adds up over the rounds and needs to be reverted in the end.
func transform(p, n *[3]uint256) {
	for r := 0; r < NumRounds; r++ {
		p2, n2 := rotateState(p, n, stateRotations[r].offset, stateRotations[r].shift)
		// unrolled S-box computation on each uint64 of the current state
		p[0][0], n[0][0] = batchBox(p[0][0], n[0][0], p2[0][0], n2[0][0])
		p[0][1], n[0][1] = batchBox(p[0][1], n[0][1], p2[0][1], n2[0][1])
		p[0][2], n[0][2] = batchBox(p[0][2], n[0][2], p2[0][2], n2[0][2])
		p[0][3], n[0][3] = batchBox(p[0][3], n[0][3], p2[0][3], n2[0][3])
		p[1][0], n[1][0] = batchBox(p[1][0], n[1][0], p2[1][0], n2[1][0])
		p[1][1], n[1][1] = batchBox(p[1][1], n[1][1], p2[1][1], n2[1][1])
		p[1][2], n[1][2] = batchBox(p[1][2], n[1][2], p2[1][2], n2[1][2])
		p[1][3], n[1][3] = batchBox(p[1][3], n[1][3], p2[1][3], n2[1][3])
		p[2][0], n[2][0] = batchBox(p[2][0], n[2][0], p2[2][0], n2[2][0])
		p[2][1], n[2][1] = batchBox(p[2][1], n[2][1], p2[2][1], n2[2][1])
		p[2][2], n[2][2] = batchBox(p[2][2], n[2][2], p2[2][2], n2[2][2])
		p[2][3], n[2][3] = batchBox(p[2][3], n[2][3], p2[2][3], n2[2][3])
		// only the first 243 bits of each uint256 are used
		p[0].norm243()
		p[1].norm243()
		p[2].norm243()
		n[0].norm243()
		n[1].norm243()
		n[2].norm243()
	}
	// successive trits are now 364⁸¹ mod 729 = 244 positions apart and need to be reordered
	reorder(p, n)
}

// rotateState rotates the Curl state by offset * 243 + s.
// It performs a left rotation of the state elements towards lower indices.
func rotateState(p, n *[3]uint256, offset, s uint) (p2, n2 [3]uint256) {
	// rotate the positive part
	p2[0].shrInto(&p[(0+offset)%3], s).shlInto(&p[(1+offset)%3], 243-s)
	p2[1].shrInto(&p[(1+offset)%3], s).shlInto(&p[(2+offset)%3], 243-s)
	p2[2].shrInto(&p[(2+offset)%3], s).shlInto(&p[(3+offset)%3], 243-s)
	// rotate the negative part
	n2[0].shrInto(&n[(0+offset)%3], s).shlInto(&n[(1+offset)%3], 243-s)
	n2[1].shrInto(&n[(1+offset)%3], s).shlInto(&n[(2+offset)%3], 243-s)
	n2[2].shrInto(&n[(2+offset)%3], s).shlInto(&n[(3+offset)%3], 243-s)
	return p2, n2
}

// batchBox applies the Curl S-box on 64 trits encoded as positive and negative bits.
func batchBox(xP, xN, yP, yN uint64) (uint64, uint64) {
	tmp := xN ^ yP
	return tmp &^ xP, ^tmp &^ (xP ^ yN)
}

// reorder arranges the state so that the trit at index (244 * k) % 729 becomes the trit at index k.
// Since the state is organized as 3 chunks of 243 trits each, the 1st output trit lies at index (0,0), 2nd at (1,1),
// 3rd at (2,2), 4th at (0,3), 5th at (1,4)...
// Thus, in order to rearrange the 1st chunk, copy trits 3*k from the 1st chunk, trits 3*k+1 from the 2nd chunk and
// trits 3*k+2 from the 3rd chunk.
func reorder(p, n *[3]uint256) {
	const (
		m0 = 0x9249249249249249       // every 3rd bit set, bit at index 0 set
		m1 = m0 << 1 & math.MaxUint64 // every 3rd bit set, bit at index 1 set
		m2 = m0 << 2 & math.MaxUint64 // every 3rd bit set, bit at index 2 set
	)
	var p2, n2 [3]uint256
	for i := uint(0); i < 3; i++ { // the uint hints to the compiler that mod 3 will never be negative
		p2[i][0] = p[i][0]&m0 | p[(1+i)%3][0]&m1 | p[(2+i)%3][0]&m2
		p2[i][1] = p[i][1]&m2 | p[(1+i)%3][1]&m0 | p[(2+i)%3][1]&m1
		p2[i][2] = p[i][2]&m1 | p[(1+i)%3][2]&m2 | p[(2+i)%3][2]&m0
		p2[i][3] = p[i][3]&m0 | p[(1+i)%3][3]&m1 | p[(2+i)%3][3]&m2

		n2[i][0] = n[i][0]&m0 | n[(1+i)%3][0]&m1 | n[(2+i)%3][0]&m2
		n2[i][1] = n[i][1]&m2 | n[(1+i)%3][1]&m0 | n[(2+i)%3][1]&m1
		n2[i][2] = n[i][2]&m1 | n[(1+i)%3][2]&m2 | n[(2+i)%3][2]&m0
		n2[i][3] = n[i][3]&m0 | n[(1+i)%3][3]&m1 | n[(2+i)%3][3]&m2
	}
	*p, *n = p2, n2
}
