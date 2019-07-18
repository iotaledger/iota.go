// Package mam provides functions for creating Masked Authentication Messaging messages
package mam

import (
	"github.com/iotaledger/iota.go/curl"

	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/trinary"
)

// Masks a given message with a curl instance state.
//	dest is the destination of the masked message
//	message is the message to be masked
//	length is the length of the message
//	c is the curl instance used to mask the message
func Mask(dest Trits, message Trits, length uint64, c *curl.Curl) {
	var chunkLength uint64

	chunk := make(Trits, HashTrinarySize)

	for i := uint64(0); i < length; i += HashTrinarySize {
		if length-i < HashTrinarySize {
			chunkLength = length - i
		} else {
			chunkLength = HashTrinarySize
		}

		copy(chunk, message[i:i+chunkLength])

		for j := uint64(0); j < chunkLength; j++ {
			dest[i+j] = Sum(chunk[j], c.State[j])
		}
		c.Absorb(chunk[:chunkLength])
	}
}

// Unmasks a given cipher with a curl instance state.
//	cipher is the cipher to be unmasked
//	length is the length of the cipher
//	c is the curl instance used to unmask the cipher
//	returns the unmasked cipher
func Unmask(cipher Trits, length uint64, c *curl.Curl) (dest Trits) {
	var chunkLength uint64
	dest = make(Trits, length)

	for i := uint64(0); i < length; i += HashTrinarySize {
		if length-i < HashTrinarySize {
			chunkLength = length - i
		} else {
			chunkLength = HashTrinarySize
		}

		for j := uint64(0); j < chunkLength; j++ {
			dest[i+j] = Sum(cipher[i+j], -c.State[j])
		}
		c.Absorb(dest[i : i+chunkLength])
	}
	return dest
}
