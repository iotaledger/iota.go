package curl

import (
	"math"

	"github.com/iotaledger/iota.go/consts"
)

const rotationOffset = 364

var stateRotations [NumberOfRounds]struct {
	offset, shift uint
}

func init() {
	var rotation uint = rotationOffset
	for r := 0; r < int(NumberOfRounds); r++ {
		stateRotations[r].offset = rotation / consts.HashTrinarySize
		stateRotations[r].shift = rotation % consts.HashTrinarySize
		rotation = (rotation * rotationOffset) % StateSize
	}
}

func transform(p, n *[3]uint256) {
	for r := 0; r < int(NumberOfRounds); r++ {
		p2, n2 := rotateState(p, n, stateRotations[r].offset, stateRotations[r].shift)
		for i := 0; i < 3; i++ {
			p[i][0], n[i][0] = batchBox(p[i][0], n[i][0], p2[i][0], n2[i][0])
			p[i][1], n[i][1] = batchBox(p[i][1], n[i][1], p2[i][1], n2[i][1])
			p[i][2], n[i][2] = batchBox(p[i][2], n[i][2], p2[i][2], n2[i][2])
			p[i][3], n[i][3] = batchBox(p[i][3], n[i][3], p2[i][3], n2[i][3])
			n[i].norm243()
			p[i].norm243()
		}
	}

	// after 81 rounds the trits are in the wrong order; successive trits are 364⁸¹ mod 729 = 244 bits apart
	reorder(p, n)
}

// rotateState rotates the Curl state by offset * 243 + s
func rotateState(p, n *[3]uint256, offset, s uint) (p2, n2 [3]uint256) {
	// rotate p to the left
	p2[0].shrInto(&p[(0+offset)%3], s).shlInto(&p[(1+offset)%3], 243-s)
	p2[1].shrInto(&p[(1+offset)%3], s).shlInto(&p[(2+offset)%3], 243-s)
	p2[2].shrInto(&p[(2+offset)%3], s).shlInto(&p[(3+offset)%3], 243-s)
	// rotate n to the left
	n2[0].shrInto(&n[(0+offset)%3], s).shlInto(&n[(1+offset)%3], 243-s)
	n2[1].shrInto(&n[(1+offset)%3], s).shlInto(&n[(2+offset)%3], 243-s)
	n2[2].shrInto(&n[(2+offset)%3], s).shlInto(&n[(3+offset)%3], 243-s)
	return p2, n2
}

// batchBox computes the Curl S-box on 64 trits.
func batchBox(xP, xN, yP, yN uint64) (uint64, uint64) {
	tmp := xN ^ yP
	return tmp &^ xP, ^tmp &^ (xP ^ yN)
}

// reorder arranges the state so that the trit at index (244 * k) % 729 becomes the trit at index k.
// Since the state is organized as 3 chunks of 243 trits each, the 1st output trit lies at index (0,0), 2nd at (1,1),
// 3rd at (2,2), 4th at (0,3), 5th at (1,4)...
// Thus, in order to rearrange the 1st chunk, copy every 3rd trit, starting with 0, from the 1st chunk, every 3rd trit,
// starting with 1, from the 2nd chunk and every 3rd trit, starting with 2, from the 3rd chunk.
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
	*n, *p = n2, p2
}
