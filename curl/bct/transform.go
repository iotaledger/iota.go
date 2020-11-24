package bct

import (
	"github.com/iotaledger/iota.go/curl"
)

func transformGeneric(lto, hto, lfrom, hfrom *[curl.StateSize]uint, rounds uint) {
	for r := rounds; r > 0; r-- {
		// three iterations unrolled
		for i := 0; i <= curl.StateSize-3; i += 3 {
			t0 := curl.Indices[i+0]
			t1 := curl.Indices[i+1]
			t2 := curl.Indices[i+2]
			t3 := curl.Indices[i+3]

			l0, h0 := lfrom[t0], hfrom[t0]
			l1, h1 := lfrom[t1], hfrom[t1]
			l2, h2 := lfrom[t2], hfrom[t2]
			l3, h3 := lfrom[t3], hfrom[t3]

			lto[i+0], hto[i+0] = sBox(l0, h0, l1, h1)
			lto[i+1], hto[i+1] = sBox(l1, h1, l2, h2)
			lto[i+2], hto[i+2] = sBox(l2, h2, l3, h3)
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
