package iotago

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
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
