package iotago

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
	"strconv"
)

// EncodeHex encodes the bytes string to a hex string. It always adds the 0x prefix.
func EncodeHex(b []byte) string {
	return hexutil.Encode(b)
}

// DecodeHex decodes the given hex string to bytes. It expects the 0x prefix.
func DecodeHex(s string) ([]byte, error) {
	b, err := hexutil.Decode(s)
	if err != nil {
		if err == hexutil.ErrEmptyString {
			return []byte{}, nil
		}
		return nil, err
	}
	return b, nil
}

// EncodeUint256 encodes the uint256 to a little-endian encoded hex string.
func EncodeUint256(n *big.Int) string {
	numBytes := n.Bytes()
	for i, j := 0, len(numBytes)-1; i < j; i, j = i+1, j-1 {
		numBytes[i], numBytes[j] = numBytes[j], numBytes[i]
	}
	return EncodeHex(append(numBytes, make([]byte, 32-len(numBytes))...))
}

// DecodeUint256 decodes the little-endian hex encoded string to an uint256.
func DecodeUint256(s string) (*big.Int, error) {
	source, err := DecodeHex(s)
	if err != nil {
		return nil, err
	}
	for i, j := 0, len(source)-1; i < j; i, j = i+1, j-1 {
		source[i], source[j] = source[j], source[i]
	}
	return new(big.Int).SetBytes(source), nil
}

// EncodeUint64 encodes the uint64 to a base 10 string.
func EncodeUint64(n uint64) string {
	return strconv.FormatUint(n, 10)
}

// DecodeUint64 decodes the base 10 string to an uint64.
func DecodeUint64(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}
