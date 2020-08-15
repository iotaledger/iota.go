package b1t6

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/iotaledger/iota.go/trinary"
	"github.com/stretchr/testify/assert"
)

var encDecTests = []*struct {
	bytes  []byte
	trytes trinary.Trytes
}{
	{[]byte{}, ""},
	{[]byte{1}, "A9"},
	{[]byte{127}, "SE"},
	{[]byte{128}, "GV"},
	{[]byte{255}, "Z9"},
	{[]byte{0, 1}, "99A9"}, // endianness
	{bytes.Repeat([]byte{0, 1}, 25), strings.Repeat("99A9", 25)}, // long
	// RFC examples
	{decodeHex("00"), "99"},
	{decodeHex("0001027e7f8081fdfeff"), "99A9B9RESEGVHVX9Y9Z9"},
	{decodeHex("9ba06c78552776a596dfe360cc2b5bf644c0f9d343a10e2e71debecd30730d03"), "GWLW9DLDDCLAJDQXBWUZYZODBYPBJCQ9NCQYT9IYMBMWNASBEDTZOYCYUBGDM9C9"},
}

func TestEncode(t *testing.T) {
	for _, tt := range encDecTests {
		t.Run(fmt.Sprintf("%x", tt.bytes), func(t *testing.T) {
			dst := make(trinary.Trits, EncodedLen(len(tt.bytes)))
			n := Encode(dst, tt.bytes)
			assert.Equal(t, len(dst), n)
			assert.Equal(t, trinary.MustTrytesToTrits(tt.trytes), dst)
		})
	}
}

func TestEncodeToTrytes(t *testing.T) {
	for _, tt := range encDecTests {
		t.Run(fmt.Sprintf("%x", tt.bytes), func(t *testing.T) {
			dst := EncodeToTrytes(tt.bytes)
			assert.Equal(t, tt.trytes, dst)
		})
	}
}

func TestDecode(t *testing.T) {
	for _, tt := range encDecTests {
		t.Run(tt.trytes, func(t *testing.T) {
			src := trinary.MustTrytesToTrits(tt.trytes)
			dst := make([]byte, DecodedLen(len(src)))
			n, err := Decode(dst, src)
			if assert.NoError(t, err) {
				assert.Equal(t, len(dst), n)
				assert.Equal(t, tt.bytes, dst)
			}
		})
	}
}

func TestDecodeTrytes(t *testing.T) {
	for _, tt := range encDecTests {
		t.Run(tt.trytes, func(t *testing.T) {
			dst, err := DecodeTrytes(tt.trytes)
			if assert.NoError(t, err) {
				assert.Equal(t, tt.bytes, dst)
			}
		})
	}
}

var errTests = []*struct {
	trytes trinary.Trytes
	bytes  []byte
	err    error
}{
	{"A", []byte{}, ErrInvalidLength},
	{"A9A", []byte{1}, ErrInvalidLength},
	{"99A9A", []byte{0, 1}, ErrInvalidLength},
	{"TE", []byte{}, ErrInvalidTrits},
	{"FV", []byte{}, ErrInvalidTrits},
	{"MM", []byte{}, ErrInvalidTrits},
	{"NN", []byte{}, ErrInvalidTrits},
	{"LI", []byte{}, ErrInvalidTrits},
	{"Z9TE", []byte{255}, ErrInvalidTrits},
	{"99A9AFV", []byte{0, 1}, ErrInvalidTrits},
}

func TestDecodeErr(t *testing.T) {
	for _, tt := range errTests {
		t.Run(fmt.Sprint(tt.trytes), func(t *testing.T) {
			trits := trinary.MustTrytesToTrits(tt.trytes)
			dst := make([]byte, DecodedLen(len(trits))+10)
			n, err := Decode(dst, trits)
			assert.Truef(t, errors.Is(err, tt.err), "unexpected error: %v", err)
			assert.Equal(t, tt.bytes, dst[:n])
		})
	}
}

func TestDecodeToTrytesErr(t *testing.T) {
	for _, tt := range errTests {
		t.Run(tt.trytes, func(t *testing.T) {
			dst, err := DecodeTrytes(tt.trytes)
			assert.Truef(t, errors.Is(err, tt.err), "unexpected error: %v", err)
			assert.Zero(t, dst)
		})
	}
}

func decodeHex(s string) []byte {
	dst, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return dst
}

var (
	benchBytesLen = 1000
	benchTritsLen = EncodedLen(benchBytesLen)
)

func BenchmarkEncode(b *testing.B) {
	data := make([][]byte, b.N)
	for i := range data {
		data[i] = randomBytes(benchBytesLen)
	}
	b.ResetTimer()

	dst := make(trinary.Trits, benchTritsLen)
	for i := range data {
		_ = Encode(dst, data[i])
	}
}

func BenchmarkEncodeToTrytes(b *testing.B) {
	data := make([][]byte, b.N)
	for i := range data {
		data[i] = randomBytes(benchBytesLen)
	}
	b.ResetTimer()

	for i := range data {
		_ = EncodeToTrytes(data[i])
	}
}

func BenchmarkDecode(b *testing.B) {
	data := make([]trinary.Trits, b.N)
	for i := range data {
		data[i] = make(trinary.Trits, benchTritsLen)
		Encode(data[i], randomBytes(benchBytesLen))
	}
	b.ResetTimer()

	dst := make([]byte, benchBytesLen)
	for i := range data {
		_, _ = Decode(dst, data[i])
	}
}

func BenchmarkDecodeTrytes(b *testing.B) {
	data := make([]trinary.Trytes, b.N)
	for i := range data {
		data[i] = EncodeToTrytes(randomBytes(benchBytesLen))
	}
	b.ResetTimer()

	for i := range data {
		_, _ = DecodeTrytes(data[i])
	}
}

func randomBytes(n int) []byte {
	result := make([]byte, n)
	rand.Read(result)
	return result
}
