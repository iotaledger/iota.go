package mam

/*
#cgo CFLAGS: -Imam -Ientangled -Iuthash/src
#cgo LDFLAGS: -L. -lmam -lkeccak
#include <mam/api/api.h>

int mam_mss_max_skn(int depth) {
	return MAM_MSS_MAX_SKN(depth);
}
*/
import "C"

// MSSMaxSKN returns the maximum amount of secret keys for the given MSS depth.
func MSSMaxSKN(depth uint) int {
	return int(C.mam_mss_max_skn(C.int(depth)))
}