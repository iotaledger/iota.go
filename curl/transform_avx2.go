// +build cgo
// +build amd64
// +build linux darwin windows

package curl

// #cgo LDFLAGS:
// #cgo CFLAGS: -mavx2 -Wall
/*
#ifdef _MSC_VER
#include <intrin.h>
#else
#include <x86intrin.h>
#endif

#define PERMUTE(a,b,c,d) (a*1+b*4+c*16+d*64)
#define NUMBER_OF_ROUNDS 81
#define STATESIZE 729

typedef unsigned long long uint64_t;

static const __m256i m_0000 = {0llu, 0llu, 0llu, 0llu};
static const __m256i m_1000 = {~0llu, 0llu, 0llu, 0llu};
static const __m256i m_1100 = {~0llu, ~0llu, 0llu, 0llu};
static const __m256i m_1110 = {~0llu, ~0llu, ~0llu, 0llu};
static const __m256i m_1111 = {~0llu, ~0llu, ~0llu, ~0llu};
static const __m256i m_0111 = {0llu, ~0llu, ~0llu, ~0llu};
static const __m256i m_0011 = {0llu, 0llu, ~0llu, ~0llu};
static const __m256i m_0001 = {0llu, 0llu, 0llu, ~0llu};
static const __m256i m_243 = {~0llu, ~0llu, ~0llu, 0x0007ffffffffffffllu};
typedef uint64_t u64x4[4];

static inline void lshift256_into_avx2(__m256i *dst, const __m256i *src, const int n){
	if (n >= 256) return;
	const int lo = n % 64, hi = n / 64;
	__m256i y = _mm256_slli_epi64(*src, lo);
	__m256i z = _mm256_srli_epi64(*src, (64 - lo));
	if (hi == 0) {
		*dst |= _mm256_permute4x64_epi64(y, PERMUTE(0,1,2,3)) & m_1111;
		*dst |= _mm256_permute4x64_epi64(z, PERMUTE(3,0,1,2)) & m_0111;
	} else if (hi == 1) {
		*dst |= _mm256_permute4x64_epi64(y, PERMUTE(3,0,1,2)) & m_0111;
		*dst |= _mm256_permute4x64_epi64(z, PERMUTE(2,3,0,1)) & m_0011;
	} else if (hi == 2) {
		*dst |= _mm256_permute4x64_epi64(y, PERMUTE(2,3,0,1)) & m_0011;
		*dst |= _mm256_permute4x64_epi64(z, PERMUTE(1,2,3,0)) & m_0001;
	} else if (hi == 3) {
		*dst |= _mm256_permute4x64_epi64(y, PERMUTE(1,2,3,0)) & m_0001;
	}
}

static inline void rshift256_into_avx2(__m256i *dst, const __m256i *src, const int n){
	if (n >= 256) return;
	const int lo = n % 64, hi = n / 64;
	__m256i y = _mm256_srli_epi64(*src, lo);
	__m256i z = _mm256_slli_epi64(*src, (64 - lo));
	if (hi == 0) {
		*dst |= _mm256_permute4x64_epi64(y, PERMUTE(0,1,2,3)) & m_1111;
		*dst |= _mm256_permute4x64_epi64(z, PERMUTE(1,2,3,0)) & m_1110;
	} else if (hi == 1) {
		*dst |= _mm256_permute4x64_epi64(y, PERMUTE(1,2,3,0)) & m_1110;
		*dst |= _mm256_permute4x64_epi64(z, PERMUTE(2,3,0,1)) & m_1100;
	} else if (hi == 2) {
		*dst |= _mm256_permute4x64_epi64(y, PERMUTE(2,3,0,1)) & m_1100;
		*dst |= _mm256_permute4x64_epi64(z, PERMUTE(3,0,1,2)) & m_1000;
	} else if (hi == 3) {
		*dst |= _mm256_permute4x64_epi64(y, PERMUTE(3,0,1,2)) & m_1000;
	}
}

static inline void curl256_avx(const __m256i *an,const __m256i* ap,  const __m256i* bn,  const __m256i* bp, __m256i* rn, __m256i* rp){
	__m256i tmp = *an ^ *bp;
	*rp = tmp & ~ *ap;
	*rn = ~tmp & ~(*ap ^ *bn);
	*rp &= m_243; // clean up
	*rn &= m_243;
}

void transform_avx2(uint64_t inout_n[3][4], uint64_t inout_p[3][4]){
	int rotation_offset = 364;
	__m256i an[3], ap[3], cn[3], cp[3];
	for (int i = 0; i < 3; i++){
		void *ioN = &inout_n[i];
		void *ioP = &inout_p[i];
		an[i] = _mm256_load_si256(ioN);
		ap[i] = _mm256_load_si256(ioP);
	}
	for (int i = 0; i < NUMBER_OF_ROUNDS; i++) {
		int hi = rotation_offset / 243, lo = rotation_offset % 243;
		__m256i bn[3] = {m_0000, m_0000, m_0000}, bp[3] = {m_0000, m_0000, m_0000};
		for (int j = 0; j < 3; j++){
			rshift256_into_avx2(&bn[(j + 3 - hi) % 3], &an[j], lo);
			lshift256_into_avx2(&bn[(j + 2 - hi) % 3], &an[j], 243 - lo);
			rshift256_into_avx2(&bp[(j + 3 - hi) % 3], &ap[j], lo);
			lshift256_into_avx2(&bp[(j + 2 - hi) % 3], &ap[j], 243 - lo);
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
	if cpu.X86.HasAVX2 {
		availableTransformFuncs[useAvx2] = transformAvx2
		if fastestTransformIndex < useAvx2 {
			fastestTransformIndex = useAvx2
			fastestTransformFunc = transformAvx2
		}
	}
}

// transformAvx2 does Curl-P-81 transformation using AVX2 instructions.
func transformAvx2(n, p *[3][4]uint64, _, _ int) {
	C.transform_avx2((*[4]C.ulonglong)(unsafe.Pointer(n)), (*[4]C.ulonglong)(unsafe.Pointer(p)))
}
