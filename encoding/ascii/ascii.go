// Package ascii implements a ternary encoding ASCII strings.
package ascii

import (
	"regexp"
	"strings"

	"github.com/iotaledger/iota.go/legacy"
	. "github.com/iotaledger/iota.go/legacy/trinary"
)

var asciiRegex = regexp.MustCompile("^[\x00-\x7F]*$")

// EncodeToTrytes returns the encoding of ASCII string src converted into trytes.
// If src is not a valid ASCII string, an error is returned.
func EncodeToTrytes(src string) (Trytes, error) {
	if !asciiRegex.MatchString(src) {
		return "", legacy.ErrInvalidASCIIInput
	}

	var dst strings.Builder
	dst.Grow(2 * len(src))

	for _, c := range []byte(src) {
		quo, rem := c/legacy.TryteRadix, c%legacy.TryteRadix
		dst.WriteByte(legacy.TryteAlphabet[rem])
		dst.WriteByte(legacy.TryteAlphabet[quo])
	}
	return dst.String(), nil
}

// DecodeTrytes returns the ASCII string represented by the encoded trytes.
// DecodeTrytes expects that src contains a valid ascii encoding and that in has even length,
// it returns an error otherwise. If src does not contain trytes, the behavior of DecodeTrytes is undefined.
func DecodeTrytes(src Trytes) (string, error) {
	if len(src)%2 != 0 {
		return "", legacy.ErrInvalidOddLength
	}

	var dst strings.Builder
	dst.Grow(len(src) / 2)

	for i := 0; i <= len(src)-2; i += 2 {
		v := strings.IndexByte(legacy.TryteAlphabet, src[i]) + strings.IndexByte(legacy.TryteAlphabet, src[i+1])*legacy.TryteRadix
		dst.WriteByte(byte(v))
	}
	return dst.String(), nil
}
