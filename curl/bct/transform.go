package bct

import (
	"github.com/iotaledger/iota.go/curl"
)

func transformGeneric(lto, hto, lfrom, hfrom *[curl.StateSize]uint) {
	for r := curl.NumRounds; r > 0; r-- {
		l0, h0 := lfrom[0], hfrom[0]
		l1, h1 := lfrom[364], hfrom[364]
		lto[0], hto[0] = sBox(l0, h0, l1, h1)

		t := 364
		for i := 1; i < curl.StateSize-1; i += 2 {
			t += 364
			l0, h0 = lfrom[t], hfrom[t]
			lto[i+0], hto[i+0] = sBox(l1, h1, l0, h0)

			t -= 365
			l1, h1 = lfrom[t], hfrom[t]
			lto[i+1], hto[i+1] = sBox(l0, h0, l1, h1)
		}
		// swap buffers
		lfrom, lto = lto, lfrom
		hfrom, hto = hto, hfrom
	}
}

func sBox(la, ha, lb, hb uint) (uint, uint) {
	tmp := (ha ^ lb) & la
	return ^tmp, (la ^ hb) | tmp
}
