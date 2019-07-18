package mam

/*
#cgo CFLAGS: -Imam -Ientangled -Iuthash/include
#cgo LDFLAGS: -L. -lmam -lkeccak
#include <mam/api/api.h>
*/
import "C"
import (
	"unsafe"
)

type Trint9 int16

type PolyCoeff Trint9

const (
	PolyN             = 1 << 10
	NTRUIDSize        = 81
	NTRUPublicKeySize = 9216
	NTRUSecretKeySize = 1024
	// NTRU session symmetric key size
	NTRUKeySize = 243
	// NTRU encrypted key size
	NNTRUEncryptedKeySize = 9216
)

/*
typedef struct mam_ntru_pk_s {
  trit_t key[MAM_NTRU_PK_SIZE];
} mam_ntru_pk_t;
*/
type c_NTRUPublicKey = C.struct_mam_ntru_pk_s

func ntruPKSliceToCSet(ntruPKs []NTRUPK) C.mam_ntru_pk_t_set_t {
	var c_ntruPKSet C.mam_ntru_pk_t_set_t
	for i := range ntruPKs {
		c_ntruPK := ntruPKs[i].c()
		C.mam_ntru_pk_t_set_add(&c_ntruPKSet, &c_ntruPK)
	}
	return c_ntruPKSet
}

type NTRUPK struct {
	Key [NTRUPublicKeySize]int8
}

func NewNTRUPK(publicKey [NTRUPublicKeySize]int8) NTRUPK {
	return NTRUPK{Key: publicKey}
}

func (ntruPK *NTRUPK) c() c_NTRUPublicKey {
	c_ntruPK := C.mam_ntru_pk_t{}
	// TODO: convert via unsafe
	for i := 0; i < NTRUPublicKeySize; i++ {
		c_ntruPK.key[i] = C.schar(ntruPK.Key[i])
	}
	return c_ntruPK
}

/*
typedef struct mam_ntru_sk_s {
  // Associated public key
  mam_ntru_pk_t public_key;
  // Secret key - small coefficients of polynomial f
  trit_t secret_key[MAM_NTRU_SK_SIZE];
  // Internal representation of a private key: NTT(1+3f)
  poly_t f;
} mam_ntru_sk_t;
*/
type c_NTRUSecretKey = C.struct_mam_ntru_sk_s

type NTRUSK struct {
	PK        NTRUPK
	SecretKey [NTRUSecretKeySize]int8
	Poly      [PolyN]PolyCoeff
}

func NewNTRUSK(mam *MAM, ntruNonce string) *NTRUSK {
	c_ntruSK := C.mam_ntru_sk_t{}
	c_ntruNonceStr := (*C.char)(unsafe.Pointer(&[]byte(ntruNonce)[0]))

	// convert string nonce to trits (trits refers to trits_t here)
	c_ntruNonce := C.trits_alloc(C.size_t(len(ntruNonce) * 3))
	C.trits_from_str(c_ntruNonce, c_ntruNonceStr)

	// hack: the pointer to a C struct refers to the first field of the struct
	// which happens to be the prng field, therefore we can convert the pointer to
	// the C MAM API to a pointer of an mam_prng_t
	// TODO: don't use this hack because it relies on the ordering of the prng within the C struct
	// TODO: check error
	C.ntru_sk_gen(&c_ntruSK, (*C.mam_prng_t)(unsafe.Pointer(&mam.c_mamAPI)), c_ntruNonce)

	var ntruSK NTRUSK
	ntruSK.PK.Key = *(*[NTRUPublicKeySize]int8)(unsafe.Pointer(&c_ntruSK.public_key.key[0]))
	ntruSK.SecretKey = *(*[NTRUSecretKeySize]int8)(unsafe.Pointer(&c_ntruSK.secret_key[0]))
	ntruSK.Poly = *(*[PolyN]PolyCoeff)(unsafe.Pointer(&c_ntruSK.f[0]))

	return &ntruSK
}

func (ntruSK *NTRUSK) c() c_NTRUSecretKey {
	c_ntruSK := C.mam_ntru_sk_t{}
	c_ntruSK.public_key.key = *(*[NTRUPublicKeySize]C.schar)(unsafe.Pointer(&ntruSK.PK.Key[0]))
	c_ntruSK.secret_key = *(*[NTRUSecretKeySize]C.schar)(unsafe.Pointer(&ntruSK.SecretKey[0]))
	c_ntruSK.f = *(*[PolyN]C.int16_t)(unsafe.Pointer(&ntruSK.Poly[0]))
	return c_ntruSK
}
