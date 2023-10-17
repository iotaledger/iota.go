package iotago

import (
	"fmt"

	"github.com/iotaledger/hive.go/constraints"
)

const (
	// MaxPayloadSize defines the maximum size of a basic block payload.
	// MaxPayloadSize = MaxBlockSize - block header - empty basic block - one strong parent - block signature.
	MaxPayloadSize = MaxBlockSize - BlockHeaderLength - BasicBlockSizeEmptyParentsAndEmptyPayload - BlockIDLength - Ed25519SignatureSerializedBytesSize
)

// PayloadType denotes the type of a payload.
type PayloadType uint8

const (
	// PayloadTaggedData denotes a TaggedData payload.
	PayloadTaggedData PayloadType = iota
	// PayloadSignedTransaction denotes a SignedTransaction.
	PayloadSignedTransaction PayloadType = 1
	// PayloadCandidacyAnnouncement denotes a CandidacyAnnouncement.
	PayloadCandidacyAnnouncement PayloadType = 2
)

func (payloadType PayloadType) String() string {
	if int(payloadType) >= len(payloadNames) {
		return fmt.Sprintf("unknown payload type: %d", payloadType)
	}

	return payloadNames[payloadType]
}

var (
	payloadNames = [PayloadSignedTransaction + 1]string{
		"TaggedData",
		"SignedTransaction",
	}
)

// Payload is an object which can be embedded into other objects.
type Payload interface {
	Sizer
	ProcessableObject
	constraints.Cloneable[Payload]

	// PayloadType returns the type of the payload.
	PayloadType() PayloadType
}
