// Package pow implements the Curl-based proof of work for arbitrary binary data.
package pow

import (
	"crypto"
	"encoding/binary"
	"math"

	legacy "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/curl"
	"github.com/iotaledger/iota.go/encoding/b1t6"
	"github.com/iotaledger/iota.go/trinary"
	_ "golang.org/x/crypto/blake2b" // BLAKE2b_256 is the default hash function for the PoW digest
)

// Hash defines the hash function that is used to compute the PoW digest.
var Hash = crypto.BLAKE2b_256

const (
	nonceBytes = 8 // len(uint64)
)

// Score returns the PoW score of msg.
func Score(msg []byte) float64 {
	if len(msg) < nonceBytes {
		panic("pow: invalid message length")
	}

	h := Hash.New()
	dataLen := len(msg) - nonceBytes
	// the PoW digest is the hash of msg without the nonce
	h.Write(msg[:dataLen])
	powDigest := h.Sum(nil)

	// extract the nonce from msg and compute the number of trailing zeros
	nonce := binary.LittleEndian.Uint64(msg[dataLen:])
	zeros := trailingZeros(powDigest, nonce)

	// compute the score
	return math.Pow(legacy.TrinaryRadix, float64(zeros)) / float64(len(msg))
}

func trailingZeros(powDigest []byte, nonce uint64) int {
	// allocate exactly one Curl block
	buf := make(trinary.Trits, legacy.HashTrinarySize)
	n := b1t6.Encode(buf, powDigest)
	// add the nonce to the trit buffer
	encodeNonce(buf[n:], nonce)

	c := curl.NewCurlP81()
	if err := c.Absorb(buf); err != nil {
		panic(err)
	}
	digest, _ := c.Squeeze(legacy.HashTrinarySize)
	return trinary.TrailingZeros(digest)
}

// encodeNonce encodes nonce as 48 trits using the b1t6 encoding.
func encodeNonce(dst trinary.Trits, nonce uint64) {
	var nonceBuf [nonceBytes]byte
	binary.LittleEndian.PutUint64(nonceBuf[:], nonce)
	b1t6.Encode(dst, nonceBuf[:])
}
