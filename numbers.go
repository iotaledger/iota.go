package iotago

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
	"strconv"
)

// EncodeHex encodes the bytes string to a hex string. It always adds the 0x prefix if bytes are not empty.
func EncodeHex(b []byte) string {
	if len(b) == 0 {
		return ""
	}
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
	return hexutil.EncodeBig(n)
}

// DecodeUint256 decodes the little-endian hex encoded string to an uint256.
func DecodeUint256(s string) (*big.Int, error) {
	return hexutil.DecodeBig(s)
}

// EncodeUint64 encodes the uint64 to a base 10 string.
func EncodeUint64(n uint64) string {
	return strconv.FormatUint(n, 10)
}

// DecodeUint64 decodes the base 10 string to an uint64.
func DecodeUint64(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}
