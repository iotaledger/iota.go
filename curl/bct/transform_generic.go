// +build !amd64 appengine gccgo

package bct

import (
	"github.com/iotaledger/iota.go/curl"
)

func transform(lto, hto, lfrom, hfrom *[curl.StateSize]uint) {
	transformGeneric(lto, hto, lfrom, hfrom)
}
