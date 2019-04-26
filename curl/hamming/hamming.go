// Package curl implements the Curl hashing function.
package hamming

import (
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/curl"
	"github.com/iotaledger/iota.go/pow"
	. "github.com/iotaledger/iota.go/trinary"
)

func check(low *[curl.StateSize]uint64, high *[curl.StateSize]uint64, security int) int {
	var sum int16

	for i := 0; i < 64; i++ {
		sum = 0

		for j := 0; j < security; j++ {
			for k := j * HashTrinarySize / 3; k < (j+1)*HashTrinarySize/3; k++ {
				if (low[k] & (1 << uint64(i))) == 0 {
					sum--
				} else if (high[k] & (1 << uint64(i))) == 0 {
					sum++
				}
			}

			if sum == 0 && j < security-1 {
				sum = 1
				break
			}
		}

		if sum == 0 {
			return i
		}
	}

	return -1
}

// Hamming calculates the hamming nonce.
func Hamming(c *curl.Curl, offset, end, security int) Trits {
	lmid, hmid := pow.Para(c.State)

	lmid[offset] = pow.PearlDiverMidStateLow0
	hmid[offset] = pow.PearlDiverMidStateHigh0
	lmid[offset+1] = pow.PearlDiverMidStateLow1
	hmid[offset+1] = pow.PearlDiverMidStateHigh1
	lmid[offset+2] = pow.PearlDiverMidStateLow2
	hmid[offset+2] = pow.PearlDiverMidStateHigh2
	lmid[offset+3] = pow.PearlDiverMidStateLow3
	hmid[offset+3] = pow.PearlDiverMidStateHigh3

	cancelled := false
	nonce, _, foundIndex := pow.Loop(lmid, hmid, security, &cancelled, check, int(c.Rounds))
	if foundIndex >= 0 {
		copy(c.State[offset:], ptritsToTrits(lmid, hmid, uint64(foundIndex), len(c.State)-offset))
		return nonce
	}
	return nil
}

func ptritsToTrits(low *[curl.StateSize]uint64, high *[curl.StateSize]uint64, index uint64, length int) (out Trits) {
	out = make(Trits, length)

	for j := 0; j < length; j++ {
		h := ((high[j] >> index) & 1) == 1
		l := ((low[j] >> index) & 1) == 1

		// out[j] = l ? (h ? 0 : -1) : 1
		if l {
			if h {
				out[j] = 0
			} else {
				out[j] = -1
			}
		} else {
			out[j] = 1
		}
	}
	return out
}
