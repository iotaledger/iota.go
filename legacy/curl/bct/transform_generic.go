// +build !amd64 appengine gccgo

package bct

import (
	"github.com/iotaledger/iota.go/legacy/curl"
)

func transform(lto, hto, lfrom, hfrom *[curl.StateSize]uint, rounds uint) {
	transformGeneric(lto, hto, lfrom, hfrom, rounds)
}
