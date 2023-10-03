package pow

import (
	"crypto"
	"encoding/binary"
	"math/bits"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
)

const (
	// Hash defines the hash function that is used to compute the PoW digest.
	Hash = crypto.BLAKE2b_256 //nolint:nosnakecase

	// HashLength defines the length of the hash function in bytes.
	HashLength = blake2b.Size256

	// NonceLength defines the length of the nonce in bytes.
	NonceLength = serializer.UInt64ByteSize

	// MaxTrailingZeros defines the maximum amount of trailing zeros.
	MaxTrailingZeros = serializer.UInt64ByteSize * 8
)

// TrailingZeros returns amount of trailing zeros for the hash of the given msg and nonce.
func TrailingZeros(msgBytes []byte, nonce uint64) int {
	nonceData := make([]byte, NonceLength)
	binary.LittleEndian.PutUint64(nonceData, nonce)

	// calculate the hash of the concatenation of the msg and the nonce.
	h := Hash.New()
	h.Write(byteutils.ConcatBytes(msgBytes, nonceData))
	hash := h.Sum(nil)

	// calculate the amount of trailing zeros
	return bits.TrailingZeros64(binary.LittleEndian.Uint64(hash[HashLength-serializer.UInt64ByteSize : HashLength]))
}
