package mam

/*
#cgo CFLAGS: -Imam -Ientangled -Iuthash/src
#cgo LDFLAGS: -L. -lmam -lkeccak
#include <mam/api/api.h>
*/
import "C"
import (
	"github.com/iotaledger/iota.go/bundle"
	"github.com/iotaledger/iota.go/trinary"
	"unsafe"
)

/*
typedef struct mam_api_s {
  mam_prng_t prng;
  mam_ntru_sk_t_set_t ntru_sks;
  mam_ntru_pk_t_set_t ntru_pks;
  mam_psk_t_set_t psks;
  mam_channel_t_set_t channels;
  trint18_t channel_ord;
  trit_t_to_mam_msg_write_context_t_map_t write_ctxs;
  trit_t_to_mam_msg_read_context_t_map_t read_ctxs;
  mam_pk_t_set_t trusted_channel_pks;
  mam_pk_t_set_t trusted_endpoint_pks;
} mam_api_t;
*/
type c_MAMAPI = C.struct_mam_api_s

const (
	MAMMessageIDSize    = 63
	MAMMessageOrderSize = 18
)

type MAM struct {
	c_mamAPI c_MAMAPI
}

// Init initializes the API.
func (mam *MAM) Init(seed trinary.Trytes) error {
	mam.c_mamAPI = C.mam_api_t{}
	c_seedTrytes := (*C.tryte_t)(unsafe.Pointer(&[]byte(seed)[0]))
	c_errCode := C.mam_api_init(&mam.c_mamAPI, c_seedTrytes)
	return wrapError(int(c_errCode))
}

// Destroy clears the API and frees its allocated memory.
func (mam *MAM) Destroy() error {
	c_errCode := C.mam_api_destroy(&mam.c_mamAPI)
	return wrapError(int(c_errCode))
}

// AddTrustedChannel adds a new trusted channel id into the API's trusted channels set.
func (mam *MAM) AddTrustedChannel(channelID trinary.Trytes) error {
	c_keyTrytes := (*C.tryte_t)(unsafe.Pointer(&[]byte(channelID)[0]))
	c_errCode := C.mam_api_add_trusted_channel_pk(&mam.c_mamAPI, c_keyTrytes)
	return wrapError(int(c_errCode))
}

// AddTrustedEndpoint adds a new trusted endpoint id into the API's trusted endpoints set.
func (mam *MAM) AddTrustedEndpoint(endpointID trinary.Trytes) error {
	c_keyTrytes := (*C.tryte_t)(unsafe.Pointer(&[]byte(endpointID)[0]))
	c_errCode := C.mam_api_add_trusted_endpoint_pk(&mam.c_mamAPI, c_keyTrytes)
	return wrapError(int(c_errCode))
}

// AddNTRUSecretKey adds a NTRU secret key to the API's NTRU secret keys set.
func (mam *MAM) AddNTRUSecretKey(ntruSK *NTRUSK) error {
	c_ntruSK := ntruSK.c()
	c_errCode := C.mam_api_add_ntru_sk(&mam.c_mamAPI, &c_ntruSK)
	return wrapError(int(c_errCode))
}

// AddNTRUPublicKey adds a NTRU public key to the API's NTRU public keys set.
func (mam *MAM) AddNTRUPublicKey(ntruPK *NTRUPK) error {
	c_ntruPK := ntruPK.c()
	c_errCode := C.mam_api_add_ntru_pk(&mam.c_mamAPI, &c_ntruPK)
	return wrapError(int(c_errCode))
}

// AddPreSharedKey adds a pre shared key to the API's pre shared keys set.
func (mam *MAM) AddPreSharedKey(psk *PSK) error {
	c_psk := psk.c()
	c_errCode := C.mam_api_add_psk(&mam.c_mamAPI, &c_psk)
	return wrapError(int(c_errCode))
}

// ChannelCreate creates and adds a new channel to the API.
func (mam *MAM) ChannelCreate(height uint) (trinary.Trytes, error) {
	var channelIDStrBuf [81]byte
	c_channelIDTrytes := (*C.tryte_t)(unsafe.Pointer(&channelIDStrBuf[0]))
	c_errCode := C.mam_api_channel_create(&mam.c_mamAPI, C.size_t(height), c_channelIDTrytes)
	if err := wrapError(int(c_errCode)); err != nil {
		return "", err
	}
	return string(channelIDStrBuf[:]), nil
}

// ChannelRemainingSecretKeys returns the number of remaining secret keys of a channel.
func (mam *MAM) ChannelRemainingSecretKeys(channelID trinary.Trytes) int {
	c_channelIDTrytes := (*C.tryte_t)(unsafe.Pointer(&[]byte(channelID)[0]))
	c_remainingSks := C.mam_api_channel_remaining_sks(&mam.c_mamAPI, c_channelIDTrytes)
	return int(c_remainingSks)
}

// EndpointCreate creates and adds a new endpoint to the API.
func (mam *MAM) EndpointCreate(height uint, channelID trinary.Trytes) (trinary.Trytes, error) {
	var endpointIDStrBuf [81]byte
	// convert
	c_channelIDTrytes := (*C.tryte_t)(unsafe.Pointer(&[]byte(channelID)[0]))
	c_endpointIDTrytes := (*C.tryte_t)(unsafe.Pointer(&endpointIDStrBuf[0]))
	c_errCode := C.mam_api_endpoint_create(&mam.c_mamAPI, C.size_t(height), c_channelIDTrytes, c_endpointIDTrytes)
	if err := wrapError(int(c_errCode)); err != nil {
		return "", err
	}
	return string(endpointIDStrBuf[:]), nil
}

// EndpointRemainingSecretKeys returns the number of remaining secret keys of an endpoint.
func (mam *MAM) EndpointRemainingSecretKeys(channelID trinary.Trytes, endpointID trinary.Trytes) int {
	c_channelIDTrytes := (*C.tryte_t)(unsafe.Pointer(&[]byte(channelID)[0]))
	c_endpointIDTrytes := (*C.tryte_t)(unsafe.Pointer(&[]byte(endpointID)[0]))
	return int(C.mam_api_endpoint_remaining_sks(&mam.c_mamAPI, c_channelIDTrytes, c_endpointIDTrytes))
}

// WriteTag creates a MAM tag which can be embedded into a transaction.
func (mam *MAM) WriteTag(msgID trinary.Trits, order int32) (trinary.Trytes, error) {
	buf := [81]int8{}
	// convert
	c_tagTrits := (*C.schar)(unsafe.Pointer(&buf[0]))
	c_msgIDTrits := (*C.schar)(unsafe.Pointer(&msgID[0]))
	C.mam_api_write_tag(c_tagTrits, c_msgIDTrits, C.int(order))
	tagTrytes, err := trinary.TritsToTrytes(buf[:])
	return tagTrytes, err
}

// BundleWriteHeaderOnChannel writes a MAM header on a channel into a bundle.
func (mam *MAM) BundleWriteHeaderOnChannel(bndl bundle.Bundle, channelID trinary.Trytes, psks []PSK, ntruPKs []NTRUPK) (bundle.Bundle, trinary.Trits, error) {

	// convert
	c_channelIDTrytes := (*C.tryte_t)(unsafe.Pointer(&[]byte(channelID)[0]))
	c_pskSet := pskSliceToCSet(psks)
	c_ntruPKSet := ntruPKSliceToCSet(ntruPKs)
	c_bundle := goBundleToCBundle(bndl)

	// msg id buf
	var msgIDBuf [MAMMessageIDSize]int8
	c_msgID := (*C.schar)(unsafe.Pointer(&msgIDBuf[0]))

	errCode := C.mam_api_bundle_write_header_on_channel(&mam.c_mamAPI, c_channelIDTrytes, c_pskSet, c_ntruPKSet, c_bundle, c_msgID)
	if err := wrapError(int(errCode)); err != nil {
		return nil, nil, err
	}

	// write mutated bundle entries back to origin bundle
	destBndl, err := cBundleToGoBundle(c_bundle)
	if err != nil {
		return nil, nil, err
	}

	return destBndl, msgIDBuf[:], nil
}

// BundleWriteHeaderOnEndpoint writes a MAM header on an endpoint into a bundle.
func (mam *MAM) BundleWriteHeaderOnEndpoint(bndl bundle.Bundle, channelID trinary.Trytes, endpointID trinary.Trytes, psks []PSK, ntruPKs []NTRUPK) (bundle.Bundle, trinary.Trits, error) {

	// convert
	c_channelID := (*C.tryte_t)(unsafe.Pointer(&[]byte(channelID)[0]))
	c_endpointID := (*C.tryte_t)(unsafe.Pointer(&[]byte(endpointID)[0]))
	c_pskSet := pskSliceToCSet(psks)
	c_ntruPKSet := ntruPKSliceToCSet(ntruPKs)
	c_bundle := goBundleToCBundle(bndl)

	// msg id buf
	msgIDBuf := [MAMMessageIDSize]int8{}
	c_msgID := (*C.schar)(unsafe.Pointer(&msgIDBuf[0]))

	errCode := C.mam_api_bundle_write_header_on_endpoint(&mam.c_mamAPI, c_channelID, c_endpointID, c_pskSet, c_ntruPKSet, c_bundle, c_msgID)
	if err := wrapError(int(errCode)); err != nil {
		return nil, nil, err
	}

	// write mutated bundle entries back to origin bundle
	destBndl, err := cBundleToGoBundle(c_bundle)
	if err != nil {
		return nil, nil, err
	}

	return destBndl, msgIDBuf[:], nil
}

// BundleAnnounceChannel writes an announcement of a channel into a bundle.
func (mam *MAM) BundleAnnounceChannel(bndl bundle.Bundle, channelID trinary.Trytes, newChannelID trinary.Trytes, psks []PSK, ntruPKs []NTRUPK) (bundle.Bundle, trinary.Trits, error) {

	// convert
	c_channelID := (*C.tryte_t)(unsafe.Pointer(&[]byte(channelID)[0]))
	c_newChannelID := (*C.tryte_t)(unsafe.Pointer(&[]byte(newChannelID)[0]))
	c_pskSet := pskSliceToCSet(psks)
	c_ntruPKSet := ntruPKSliceToCSet(ntruPKs)
	c_bundle := goBundleToCBundle(bndl)

	// msg id buf
	msgIDBuf := [MAMMessageIDSize]int8{}
	c_msgID := (*C.schar)(unsafe.Pointer(&msgIDBuf[0]))

	errCode := C.mam_api_bundle_announce_channel(&mam.c_mamAPI, c_channelID, c_newChannelID, c_pskSet, c_ntruPKSet, c_bundle, c_msgID)
	if err := wrapError(int(errCode)); err != nil {
		return nil, nil, err
	}

	// write mutated bundle entries back to origin bundle
	destBndl, err := cBundleToGoBundle(c_bundle)
	if err != nil {
		return nil, nil, err
	}

	return destBndl, msgIDBuf[:], nil
}

// BundleAnnounceEndpoint writes an announcement of an endpoint into a bundle.
func (mam *MAM) BundleAnnounceEndpoint(bndl bundle.Bundle, channelID trinary.Trytes, endpointID trinary.Trytes, psks []PSK, ntruPKs []NTRUPK) (bundle.Bundle, trinary.Trits, error) {

	// convert
	c_channelID := (*C.tryte_t)(unsafe.Pointer(&[]byte(channelID)[0]))
	c_endpointID := (*C.tryte_t)(unsafe.Pointer(&[]byte(endpointID)[0]))
	c_pskSet := pskSliceToCSet(psks)
	c_ntruPKSet := ntruPKSliceToCSet(ntruPKs)
	c_bundle := goBundleToCBundle(bndl)

	// msg id buf
	msgIDBuf := [MAMMessageIDSize]int8{}
	c_msgID := (*C.schar)(unsafe.Pointer(&msgIDBuf[0]))

	errCode := C.mam_api_bundle_announce_endpoint(&mam.c_mamAPI, c_channelID, c_endpointID, c_pskSet, c_ntruPKSet, c_bundle, c_msgID)
	if err := wrapError(int(errCode)); err != nil {
		return nil, nil, err
	}

	// write mutated bundle entries back to origin bundle
	destBndl, err := cBundleToGoBundle(c_bundle)
	if err != nil {
		return nil, nil, err
	}

	return destBndl, msgIDBuf[:], nil
}

// BundleWritePacket writes a MAM packet into a bundle.
func (mam *MAM) BundleWritePacket(msgID trinary.Trits, payload trinary.Trytes, checksum MsgChecksum, lastPacket bool, bndl bundle.Bundle) (bundle.Bundle, error) {
	// convert
	c_bundle := goBundleToCBundle(bndl)
	c_msgIDTrits := (*C.schar)(unsafe.Pointer(&msgID[0]))
	c_payload := (*C.tryte_t)(unsafe.Pointer(&[]byte(payload)[0]))
	errCode := C.mam_api_bundle_write_packet(&mam.c_mamAPI, c_msgIDTrits, c_payload, C.size_t(len(payload)), goMsgChecksumToC(checksum), C._Bool(lastPacket), c_bundle)
	if err := wrapError(int(errCode)); err != nil {
		return nil, err
	}
	destBndl, err := cBundleToGoBundle(c_bundle)
	if err != nil {
		return nil, err
	}
	return destBndl, nil
}

// BundleReadPacket reads MAM's session key and potentially the first packet using an NTRU secret key.
func (mam *MAM) BundleRead(bndl bundle.Bundle) (trinary.Trytes, bool, error) {
	var payloadBuf *[]int8
	c_payloadBufPointer := (*C.schar)(unsafe.Pointer(payloadBuf))
	c_payloadBufPointerPointer := &c_payloadBufPointer
	c_bundle := goBundleToCBundle(bndl)
	var c_lastPacket C._Bool
	c_payloadSize := C.ulong(0)
	errCode := C.mam_api_bundle_read(&mam.c_mamAPI, c_bundle, c_payloadBufPointerPointer, &c_payloadSize, &c_lastPacket)
	if err := wrapError(int(errCode)); err != nil {
		return "", false, err
	}

	lastPacket := (*bool)(unsafe.Pointer(&c_lastPacket))

	c_msg := (*C.char)(unsafe.Pointer(*c_payloadBufPointerPointer))
	msg := C.GoStringN(c_msg, C.int(c_payloadSize))
	C.free(unsafe.Pointer(c_msg))

	return msg, *lastPacket, nil
}

// Serializes an API into its trits representation.
func (mam *MAM) Serialize(encrKey ...trinary.Trytes) trinary.Trits {
	var c_encrKey *C.tryte_t
	var encrKeyLen int

	if len(encrKey) > 0 {
		c_encrKey = (*C.tryte_t)(unsafe.Pointer(&[]byte(encrKey[0])[0]))
		encrKeyLen = len(encrKey[0])
	}

	// get the size of the trits serialized API instance
	c_tritsSize := C.mam_api_serialized_size(&mam.c_mamAPI)
	tritsBuf := make([]int8, int(c_tritsSize), int(c_tritsSize))
	c_tritsBuf := (*C.trit_t)(unsafe.Pointer(&tritsBuf[0]))

	// serialize
	C.mam_api_serialize(&mam.c_mamAPI, c_tritsBuf, c_encrKey, C.ulong(encrKeyLen))
	return tritsBuf
}

// Deserializes an API's trits representation in to this API instance.
func (mam *MAM) Deserialize(apiTrits trinary.Trits, decrKey ...trinary.Trytes) error {
	var c_decrKey *C.tryte_t
	var decrKeyLen int

	if len(decrKey) > 0 {
		c_decrKey = (*C.tryte_t)(unsafe.Pointer(&[]byte(decrKey[0])[0]))
		decrKeyLen = len(decrKey[0])
	}

	c_tritsBuf := (*C.trit_t)(unsafe.Pointer(&apiTrits[0]))
	errCode := C.mam_api_deserialize(c_tritsBuf, C.size_t(len(apiTrits)), &mam.c_mamAPI, c_decrKey, C.ulong(decrKeyLen))
	return wrapError(int(errCode))
}

// Saves the API instance into a file.
func (mam *MAM) Save(fileName string, encrKey ...trinary.Trytes) error {
	var c_encrKey *C.tryte_t
	var encrKeyLen int

	if len(encrKey) > 0 {
		c_encrKey = (*C.tryte_t)(unsafe.Pointer(&[]byte(encrKey[0])[0]))
		encrKeyLen = len(encrKey[0])
	}

	c_fileName := (*C.char)(unsafe.Pointer(&[]byte(fileName)[0]))

	errCode := C.mam_api_save(&mam.c_mamAPI, c_fileName, c_encrKey, C.ulong(encrKeyLen))
	return wrapError(int(errCode))
}

// Loads an API instance from a file.
func (mam *MAM) Load(fileName string, decrKey ...trinary.Trytes) error {
	var c_decrKey *C.tryte_t
	var decrKeyLen int

	if len(decrKey) > 0 {
		c_decrKey = (*C.tryte_t)(unsafe.Pointer(&[]byte(decrKey[0])[0]))
		decrKeyLen = len(decrKey[0])
	}

	c_fileName := (*C.char)(unsafe.Pointer(&[]byte(fileName)[0]))
	errCode := C.mam_api_load(c_fileName, &mam.c_mamAPI, c_decrKey, C.ulong(decrKeyLen))
	return wrapError(int(errCode))
}
