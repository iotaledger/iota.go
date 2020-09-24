package bct

import (
	"github.com/iotaledger/iota.go/curl"
)

func transformGeneric(lto, hto, lfrom, hfrom *[curl.StateSize]uint, rounds uint) {
	for r := rounds; r > 0; r-- {
		// three Curl-P rounds unrolled
		for i := 0; i < curl.StateSize-2; i += 3 {
			t0 := curl.Indices[i+0]
			t1 := curl.Indices[i+1]
			t2 := curl.Indices[i+2]
			t3 := curl.Indices[i+3]

			l0, h0 := lfrom[t0], hfrom[t0]
			l1, h1 := lfrom[t1], hfrom[t1]
			l2, h2 := lfrom[t2], hfrom[t2]
			l3, h3 := lfrom[t3], hfrom[t3]

			v0 := (h0 ^ l1) & l0
			lto[i+0], hto[i+0] = ^v0, (l0^h1)|v0
			v1 := (h1 ^ l2) & l1
			lto[i+1], hto[i+1] = ^v1, (l1^h2)|v1
			v2 := (h2 ^ l3) & l2
			lto[i+2], hto[i+2] = ^v2, (l2^h3)|v2
		}
		// swap buffers
		lfrom, lto = lto, lfrom
		hfrom, hto = hto, hfrom
	}
}
