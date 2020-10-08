// +build !amd64 appengine gccgo

package classic

import "github.com/iotaledger/iota.go/curl"

func transform(dst, src *[curl.StateSize]int8, rounds uint) {
	transformGeneric(dst, src, rounds)
}
