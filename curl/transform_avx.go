// +build cgo
// +build amd64
// +build linux darwin windows

package curl

// #cgo LDFLAGS:
// #cgo CFLAGS: -mavx -Wall
/*
#ifdef _MSC_VER
#include <intrin.h>
#else
#include <x86intrin.h>
#endif

#define NUMBER_OF_ROUNDS 81
#define STATESIZE 729

typedef unsigned long long uint64_t;

static const __m256i m_0000 = {0llu, 0llu, 0llu, 0llu};
static const __m256i m_243 = {~0llu, ~0llu, ~0llu, 0x0007ffffffffffffllu};

static inline void lshift256_into_arr(uint64_t dst[4], uint64_t src[4], const int n){
	if (n >= 256) return;
	const int lo = n % 64, hi = n / 64;
	if (hi == 0) {
		dst[0] |= src[0] << lo;
		dst[1] |= src[1] << lo | src[0] >> (64 - lo);
		dst[2] |= src[2] << lo | src[1] >> (64 - lo);
		dst[3] |= src[3] << lo | src[2] >> (64 - lo);
	} else if (hi == 1) {
		dst[1] |= src[0] << lo;
		dst[2] |= src[1] << lo | src[0] >> (64 - lo);
		dst[3] |= src[2] << lo | src[1] >> (64 - lo);
	} else if (hi == 2) {
		dst[2] |= src[0] << lo;
		// why is 'lo == 0 ? ...' needed here?
		dst[3] |= lo == 0 ? src[1] : src[1] << lo | src[0] >> (64 - lo);
	} else if (hi == 3) {
		dst[3] |= src[0] << lo;
	}
}

static inline void rshift256_into_arr(uint64_t dst[4], uint64_t src[4], const int n){
	if (n >= 256) return;
	const int lo = n % 64, hi = n / 64;
	if (hi == 0) {
		dst[0] |= src[0] >> lo | src[1] << (64 - lo);
		dst[1] |= src[1] >> lo | src[2] << (64 - lo);
		dst[2] |= src[2] >> lo | src[3] << (64 - lo);
		dst[3] |= src[3] >> lo;
	} else if (hi == 1) {
		dst[0] |= lo == 0 ? src[1] : src[1] >> lo | src[2] << (64 - lo); // why '? :'
		dst[1] |= lo == 0 ? src[2] : src[2] >> lo | src[3] << (64 - lo); // same
		dst[2] |= src[3] >> lo;
	} else if (hi == 2) {
		dst[0] |= src[2] >> lo | src[3] << (64 - lo);
		dst[1] |= src[3] >> lo;
	} else if (hi == 3) {
		dst[0] |= src[3] >> lo;
	}
}

static inline void curl256_avx(const __m256i *an,const __m256i* ap,  const __m256i* bn,  const __m256i* bp, __m256i* rn, __m256i* rp){
	__m256i tmp = *an ^ *bp;
	*rp = tmp & ~ *ap;
	*rn = ~tmp & ~(*ap ^ *bn);
	*rp &= m_243; // clean up
	*rn &= m_243;
}

void transform_avx(uint64_t inout_n[3][4], uint64_t inout_p[3][4]){
	int rotation_offset = 364;
	__m256i an[3], ap[3], cn[3], cp[3];
	for (int i = 0; i < 3; i++){
		for (int j = 0; j < 4; j++){
			an[i][j] = inout_n[i][j];
			ap[i][j] = inout_p[i][j];
		}
	}
	for (int i = 0; i < NUMBER_OF_ROUNDS; i++) {
		int hi = rotation_offset / 243, lo = rotation_offset % 243;
		__m256i bn[3] = {m_0000, m_0000, m_0000}, bp[3] = {m_0000, m_0000, m_0000};
		for (int j = 0; j < 3; j++){ // rotate
			rshift256_into_arr((uint64_t *)&bn[(j + 3 - hi) % 3], (uint64_t *)&an[j], lo);
			lshift256_into_arr((uint64_t *)&bn[(j + 2 - hi) % 3], (uint64_t *)&an[j], 243 - lo);
			rshift256_into_arr((uint64_t *)&bp[(j + 3 - hi) % 3], (uint64_t *)&ap[j], lo);
			lshift256_into_arr((uint64_t *)&bp[(j + 2 - hi) % 3], (uint64_t *)&ap[j], 243 - lo);
		}
		for (int j = 0; j < 3; j++) {
			curl256_avx(&(an[j]), &(ap[j]), &(bn[j]), &(bp[j]), &(cn[j]), &(cp[j]));
		}
		for (int j = 0; j < 3; j++){
			an[j] = cn[j];
			ap[j] = cp[j];
		}
		rotation_offset = (rotation_offset * 364) % STATESIZE;
	}
	// resort: trits are now sorted like 0, 487, 245, 3, 490, 248, ...
	const uint64_t m9 = 0x9249249249249249llu, m4 = 0x4924924924924924llu, m2 = 0x2492492492492492llu;
	const __m256i mask[3] = {{m9, m4, m2, m9}, {m2, m9, m4, m2}, {m4, m2, m9, m4}};
	for (int i = 0; i < 3; i++) {
		cn[i] = (an[i] & mask[0]) | (an[(i + 1) % 3] & mask[1]) | (an[(i + 2) % 3] & mask[2]);
		cp[i] = (ap[i] & mask[0]) | (ap[(i + 1) % 3] & mask[1]) | (ap[(i + 2) % 3] & mask[2]);
	}
	for (int i = 0; i < 3; i++){
		for (int j = 0; j < 4; j++){
			inout_n[i][j] = cn[i][j];
			inout_p[i][j] = cp[i][j];
		}
	}
}
*/
import "C"
import (
	"golang.org/x/sys/cpu"
	"unsafe"
)

func init() {
	// Add transform func if the CPU supports AVX2
	if cpu.X86.HasAVX {
		availableTransformFuncs[useAvx] = transformAvx
		if fastestTransformIndex < useAvx {
			fastestTransformIndex = useAvx
			fastestTransformFunc = transformAvx
		}
	}
}

// transformAvx does Curl-P-81 transformation using AVX instructions.
func transformAvx(n, p *[3][4]uint64, _, _ int) {
	//fmt.Println("using AVX")
	C.transform_avx((*[4]C.ulonglong)(unsafe.Pointer(n)), (*[4]C.ulonglong)(unsafe.Pointer(p)))
}
