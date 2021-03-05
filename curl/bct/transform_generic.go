// +build !amd64 appengine gccgo

package bct

func transform(lto, hto, lfrom, hfrom *[729]uint) {
	transformGeneric(lto, hto, lfrom, hfrom)
}
