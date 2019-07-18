package mam

/*
#cgo CFLAGS: -Imam -Ientangled -Iuthash/src
#cgo LDFLAGS: -L. -lmam -lkeccak
#include <mam/api/api.h>
*/
import "C"


type MsgPubKey int

const (
	MsgPubKeyChannelID   MsgPubKey = 0
	MsgPubKeyEndpointID  MsgPubKey = 1
	MsgPubKeyChannelID1  MsgPubKey = 2
	MsgPubKeyEndpointID1 MsgPubKey = 3
)

func goMsgPubKeyToC(msgPubKey MsgPubKey) C.mam_msg_type_t {
	switch (msgPubKey) {
	case MsgPubKeyChannelID:
		return C.MAM_MSG_PUBKEY_CHID
	case MsgPubKeyEndpointID:
		return C.MAM_MSG_PUBKEY_EPID
	case MsgPubKeyChannelID1:
		return C.MAM_MSG_PUBKEY_CHID1
	case MsgPubKeyEndpointID1:
		return C.MAM_MSG_PUBKEY_EPID1
	default:
		panic("invalid message public key type")
	}
}

type MsgType int

const (
	MsgTypeUnstructured = 0
	MsgTypePubKeyCert   = 1
)

func goMsgTypeToC(msgType MsgType) C.mam_msg_type_t {
	switch (msgType) {
	case MsgTypeUnstructured:
		return C.MAM_MSG_TYPE_UNSTRUCTURED
	case MsgTypePubKeyCert:
		return C.MAM_MSG_TYPE_PK_CERT
	default:
		panic("invalid message type")
	}
}

type MsgKeyload int

const (
	MsgKeyloadPublic MsgKeyload = 0
	MsgKeyloadPSK    MsgKeyload = 1
	MsgKeyloadNTRU   MsgKeyload = 2
)

func goMsgKeyloadToC(msgKeyload MsgKeyload) C.mam_msg_keyload_t {
	switch (msgKeyload) {
	case MsgKeyloadPublic:
		return C.MAM_MSG_KEYLOAD_PUBLIC
	case MsgKeyloadPSK:
		return C.MAM_MSG_KEYLOAD_PSK
	case MsgKeyloadNTRU:
		return C.MAM_MSG_KEYLOAD_NTRU
	default:
		panic("invalid message keyload type")
	}
}

type MsgChecksum int

const (
	MsgChecksumNone MsgChecksum = 0
	MsgChecksumMAC  MsgChecksum = 1
	MsgChecksumSig  MsgChecksum = 2
)

func goMsgChecksumToC(msgChecksum MsgChecksum) C.mam_msg_checksum_t {
	switch (msgChecksum) {
	case MsgChecksumNone:
		return C.MAM_MSG_CHECKSUM_NONE
	case MsgChecksumMAC:
		return C.MAM_MSG_CHECKSUM_MAC
	case MsgChecksumSig:
		return C.MAM_MSG_CHECKSUM_SIG
	default:
		panic("invalid message checksum type")
	}
}
