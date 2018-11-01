// Package converter provides functions for ASCII to Trytes conversion (and vice-versa).
package converter

import (
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/trinary"
	"regexp"
	"strings"
)

var asciiRegex = regexp.MustCompile("^[\x00-\x7F]*$")

// ASCIIToTrytes converts an ascii encoded string to trytes.
func ASCIIToTrytes(s string) (Trytes, error) {
	if !asciiRegex.MatchString(s) {
		return "", ErrInvalidASCIIInput
	}

	trytesStr := ""

	for _, c := range s {
		trytesStr += string(TryteAlphabet[c%27])
		trytesStr += string(TryteAlphabet[(c-c%27)/27])
	}

	return NewTrytes(trytesStr)
}

// TrytesToASCII converts trytes of even length to an ascii string.
func TrytesToASCII(trytes Trytes) (string, error) {
	if err := ValidTrytes(trytes); err != nil {
		return "", err
	}

	if len(trytes)%2 != 0 {
		return "", ErrInvalidOddLength
	}

	ascii := ""
	for i := 0; i < len(trytes); i += 2 {
		ascii += string(strings.IndexRune(TryteAlphabet, rune(trytes[i])) + (strings.IndexRune(TryteAlphabet, rune(trytes[i+1])) * 27))
	}

	return ascii, nil
}
