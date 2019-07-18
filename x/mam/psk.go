package mam

/*
#cgo CFLAGS: -Imam -Ientangled -Iuthash/include
#cgo LDFLAGS: -L. -lmam -lkeccak
#include <mam/api/api.h>
*/
import "C"
import "unsafe"

/*
typedef struct mam_psk_s {
  trit_t id[MAM_PSK_ID_SIZE];
  trit_t key[MAM_PSK_KEY_SIZE];
} mam_psk_t;
*/
type c_PreSharedKey = C.struct_mam_psk_s

type PSK struct {
	ID  [PSKIDSize]int8
	Key [PSKKeySize]int8
}

func (psk *PSK) c() c_PreSharedKey {
	c_psk := C.struct_mam_psk_s{}
	// TODO: convert via unsafe
	for i := 0; i < PSKIDSize; i++ {
		c_psk.id[i] = C.schar(psk.ID[i])
	}
	for i := 0; i < PSKKeySize; i++ {
		c_psk.key[i] = C.schar(psk.Key[i])
	}
	return c_psk
}

const (
	PSKIDSize  = 81
	PSKKeySize = 243
)

func NewPSK(mam *MAM, key string, nonce string) *PSK {
	c_psk := &C.mam_psk_t{}
	c_key := (*C.tryte_t)(unsafe.Pointer(&[]byte(key)[0]))
	c_nonce := (*C.tryte_t)(unsafe.Pointer(&[]byte(key)[0]))

	// hack: the pointer to a C struct refers to the first field of the struct
	// which happens to be the prng field, therefore we can convert the pointer to
	// the C MAM API to a pointer of an mam_prng_t
	// TODO: don't use this hack because it relies on the ordering of the prng within the C struct
	// TODO: check error
	C.mam_psk_gen(c_psk, (*C.mam_prng_t)(unsafe.Pointer(&mam.c_mamAPI)), c_key, c_nonce, C.size_t(len(nonce)))

	var psk PSK
	psk.ID = *(*[PSKIDSize]int8)(unsafe.Pointer(&c_psk.id[0]))
	psk.Key = *(*[PSKKeySize]int8)(unsafe.Pointer(&c_psk.key[0]))
	return &psk
}

func pskSliceToCSet(psks []PSK) C.mam_psk_t_set_t {
	var c_pskSet C.mam_psk_t_set_t
	for i := range psks {
		c_psk := psks[i].c()
		C.mam_psk_t_set_add(&c_pskSet, &c_psk)
	}
	return c_pskSet
}
