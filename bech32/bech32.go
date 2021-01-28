// Package bech32 implements bech32 encoding and decoding.
package bech32

import (
	"errors"
	"fmt"
	"strings"

	"github.com/iotaledger/iota.go/v2/bech32/internal/base32"
)

const (
	maxStringLength = 90
	checksumLength  = 6
	separator       = '1'
)

var charset = newEncoding("qpzry9x8gf2tvdw0s3jn54khce6mua7l")

// Encode encodes the String string and the src data as a Bech32 string.
// It returns an error when the input is invalid.
func Encode(hrp string, src []byte) (string, error) {
	dataLen := base32.EncodedLen(len(src))
	if len(hrp)+dataLen+checksumLength+1 > maxStringLength {
		return "", fmt.Errorf("%w: String length=%d, data length=%d", ErrInvalidLength, len(hrp), dataLen)
	}
	// validate the human-readable part
	if len(hrp) < 1 {
		return "", fmt.Errorf("%w: String must not be empty", ErrInvalidLength)
	}
	for _, c := range hrp {
		if !isValidHRPChar(c) {
			return "", fmt.Errorf("%w: not US-ASCII character in human-readable part", ErrInvalidCharacter)
		}
	}
	if err := validateCase(hrp); err != nil {
		return "", err
	}

	// convert the human-readable part to lower for the checksum
	hrpLower := strings.ToLower(hrp)

	// convert to base32 and add the checksum
	data := make([]uint8, base32.EncodedLen(len(src))+checksumLength)
	base32.Encode(data, src)
	copy(data[dataLen:], bech32CreateChecksum(hrpLower, data[:dataLen]))

	// enc the data part using the charset
	chars := charset.encode(data)

	// convert to a string using the corresponding charset
	var res strings.Builder
	res.WriteString(hrp)
	res.WriteByte(separator)
	res.WriteString(chars)

	// return with the correct case
	if hrp == hrpLower {
		return res.String(), nil
	}
	return strings.ToUpper(res.String()), nil
}

// Decode decodes the Bech32 string s into its human-readable and data part.
// It returns an error when s does not represent a valid Bech32 encoding.
// An SyntaxError is returned when the error can be matched to a certain position in s.
func Decode(s string) (string, []byte, error) {
	if len(s) > maxStringLength {
		return "", nil, &SyntaxError{fmt.Errorf("%w: maximum length exceeded", ErrInvalidLength), maxStringLength}
	}
	// validate the separator
	hrpLen := strings.LastIndex(s, string(separator))
	if hrpLen == -1 {
		return "", nil, ErrMissingSeparator
	}
	if hrpLen < 1 || hrpLen+checksumLength > len(s) {
		return "", nil, &SyntaxError{fmt.Errorf("%w: invalid position", ErrInvalidSeparator), hrpLen}
	}
	// validate characters in human-readable part
	for i, c := range s[:hrpLen] {
		if !isValidHRPChar(c) {
			return "", nil, &SyntaxError{fmt.Errorf("%w: not US-ASCII character in human-readable part", ErrInvalidCharacter), i}
		}
	}
	// validate that the case of the entire string is consistent
	if err := validateCase(s); err != nil {
		return "", nil, err
	}

	// convert everything to lower
	s = strings.ToLower(s)
	hrp := s[:hrpLen]
	chars := s[hrpLen+1:]

	// decode the data part
	data, err := charset.decode(chars)
	if err != nil {
		return "", nil, &SyntaxError{fmt.Errorf("%w: non-charset character in data part", ErrInvalidCharacter), hrpLen + 1 + len(data)}
	}

	// validate the checksum
	if len(data) < checksumLength || !bech32VerifyChecksum(hrp, data) {
		return "", nil, &SyntaxError{ErrInvalidChecksum, len(s) - checksumLength}
	}
	data = data[:len(data)-checksumLength]

	// decode the data part
	dst := make([]byte, base32.DecodedLen(len(data)))
	if _, err := base32.Decode(dst, data); err != nil {
		var e *base32.CorruptInputError
		if errors.As(err, &e) {
			return "", nil, &SyntaxError{e.Unwrap(), hrpLen + 1 + e.Offset}
		}
		return "", nil, err
	}
	return hrp, dst, nil
}

func isValidHRPChar(r rune) bool {
	// it must only contain US-ASCII characters, with each character having a value in the range [33-126]
	return r >= 33 && r <= 126
}

func validateCase(s string) error {
	upper, lower := firstUpper(s), firstLower(s)
	if upper < lower && upper >= 0 {
		return &SyntaxError{ErrMixedCase, lower}
	}
	if lower < upper && lower >= 0 {
		return &SyntaxError{ErrMixedCase, upper}
	}
	return nil
}

func firstUpper(s string) int {
	lower := strings.ToLower(s)
	for i := range s {
		if lower[i] != s[i] {
			return i
		}
	}
	return -1
}

func firstLower(s string) int {
	lower := strings.ToUpper(s)
	for i := range s {
		if lower[i] != s[i] {
			return i
		}
	}
	return -1
}
