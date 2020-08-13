// +build !amd64 appengine gccgo

package curl

func transform(dst, src *[StateSize]int8, rounds int) {
	transformGeneric(dst, src, rounds)
}
