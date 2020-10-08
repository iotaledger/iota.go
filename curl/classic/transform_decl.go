// +build !gccgo,!appengine
// +build amd64

package classic

import "github.com/iotaledger/iota.go/curl"

//go:noescape
func transform(dst, src *[curl.StateSize]int8, rounds uint)
