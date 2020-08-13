// +build !gccgo,!appengine
// +build amd64

package curl

//go:noescape
func transform(dst, src *[StateSize]int8, rounds int)
