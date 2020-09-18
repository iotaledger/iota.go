// +build !gccgo,!appengine
// +build amd64

package bct

import "github.com/iotaledger/iota.go/legacy/curl"

var Indices = curl.Indices

//go:noescape
func transform(lto, hto, lfrom, hfrom *[curl.StateSize]uint, rounds uint)
