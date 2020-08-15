package t5b1

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/trinary"
	"github.com/stretchr/testify/assert"
)

var encDecTest = []*struct {
	trytes trinary.Trytes
	bytes  []byte
}{
	{"", []byte{}},
	{"9NOPQRSTUVWXYZ9", []byte{0x94, 0x2c, 0xa2, 0x12, 0xea, 0xd1, 0xab, 0xa9, 0x00}},
	{"9ABCDEFGHIJKLM9", []byte{0x1b, 0x06, 0x25, 0xb4, 0xc5, 0x54, 0x40, 0x76, 0x04}},
	{strings.Repeat("YZ9AB", 20), bytes.Repeat([]byte{0xe3, 0x51, 0x12}, 20)}, // long
	{"M", []byte{0x0d}},                 // 2 trit padding
	{"MM", []byte{0x79, 0x01}},          // 4 trit padding
	{"MMM", []byte{0x79, 0x28}},         // 1 trit padding
	{"MMMM", []byte{0x79, 0x79, 0x04}},  // 3 trit padding
	{"MMMMM", []byte{0x79, 0x79, 0x79}}, // no padding
}

func TestEncode(t *testing.T) {
	for _, tt := range encDecTest {
		t.Run(tt.trytes, func(t *testing.T) {
			src := trinary.MustTrytesToTrits(tt.trytes)
			dst := make([]byte, EncodedLen(len(src)))
			n := Encode(dst, src)
			assert.Equal(t, len(dst), n)
			assert.Equal(t, tt.bytes, dst)
		})
	}
}

func TestEncodeTrytes(t *testing.T) {
	for _, tt := range encDecTest {
		t.Run(tt.trytes, func(t *testing.T) {
			dst := EncodeTrytes(tt.trytes)
			assert.Equal(t, tt.bytes, dst)
		})
	}
}

func TestDecode(t *testing.T) {
	for _, tt := range encDecTest {
		t.Run(fmt.Sprintf("%x", tt.bytes), func(t *testing.T) {
			dst := make(trinary.Trits, DecodedLen(len(tt.bytes)))
			n, err := Decode(dst, tt.bytes)
			if assert.NoError(t, err) {
				assert.Equal(t, len(dst), n)
				// add expected padding
				paddedLen := ((len(tt.trytes)*consts.TritsPerTryte + tritsInByte - 1) / tritsInByte) * tritsInByte
				expDec := trinary.MustPadTrits(trinary.MustTrytesToTrits(tt.trytes), paddedLen)
				assert.Equal(t, expDec, dst)
			}
		})
	}
}

func TestDecodeToTrytes(t *testing.T) {
	for _, tt := range encDecTest {
		t.Run(fmt.Sprintf("%x", tt.bytes), func(t *testing.T) {
			dst, err := DecodeToTrytes(tt.bytes)
			if assert.NoError(t, err) {
				// add expected padding
				padLend := (((len(tt.trytes)*3+tritsInByte-1)/tritsInByte)*tritsInByte)/3 - (len(tt.trytes))
				expDec := trinary.MustPad(tt.trytes, len(tt.trytes)+padLend)
				assert.Equal(t, expDec, dst)
			}
		})
	}
}

func TestDecodeErr(t *testing.T) {
	var tests = []*struct {
		bytes []byte
		trits trinary.Trits
		err   error
	}{
		{[]byte{0x7a}, []int8{}, consts.ErrInvalidByte},
		{[]byte{0x80}, []int8{}, consts.ErrInvalidByte},
		{[]byte{0x86}, []int8{}, consts.ErrInvalidByte},
		{[]byte{0x79, 0x7a}, []int8{1, 1, 1, 1, 1}, consts.ErrInvalidByte},
		{[]byte{0x00, 0x01, 0x7a}, []int8{0, 0, 0, 0, 0, 1, 0, 0, 0, 0}, consts.ErrInvalidByte},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%x", tt.bytes), func(t *testing.T) {
			dst := make(trinary.Trits, DecodedLen(len(tt.bytes))+10)
			n, err := Decode(dst, tt.bytes)
			assert.Truef(t, errors.Is(err, tt.err), "unexpected error: %v", err)
			assert.Equal(t, tt.trits, dst[:n])
		})
	}
}

func TestDecodeToTrytesErr(t *testing.T) {
	var tests = []*struct {
		bytes []byte
		err   error
	}{
		{[]byte{0x7a}, consts.ErrInvalidByte},
		{[]byte{0x80}, consts.ErrInvalidByte},
		{[]byte{0x86}, consts.ErrInvalidByte},
		{[]byte{0x79, 0x7a}, consts.ErrInvalidByte},
		{[]byte{0x00, 0x01, 0x7a}, consts.ErrInvalidByte},
		{[]byte{0x5e}, ErrNonZeroPadding}, // 1, 1, 1, 0, 1
		{[]byte{0xc1}, ErrNonZeroPadding}, // 1, 1, 1, 0, -1
		{[]byte{0x28}, ErrNonZeroPadding}, //1, 1, 1, 1, 0
		{[]byte{0xf2}, ErrNonZeroPadding}, // 1, 1, 1, -1, 0
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%x", tt.bytes), func(t *testing.T) {
			dst, err := DecodeToTrytes(tt.bytes)
			assert.Truef(t, errors.Is(err, tt.err), "unexpected error: %v", err)
			assert.Zero(t, dst)
		})
	}
}

var (
	benchBytesLen = 1000
	benchTritsLen = DecodedLen(benchBytesLen)
)

func BenchmarkEncode(b *testing.B) {
	data := make([]trinary.Trits, b.N)
	for i := range data {
		data[i] = randomTrits(benchTritsLen)
	}
	b.ResetTimer()

	dst := make([]byte, benchBytesLen)
	for i := range data {
		_ = Encode(dst, data[i])
	}
}

func BenchmarkEncodeTrytes(b *testing.B) {
	data := make([]trinary.Trytes, b.N)
	for i := range data {
		data[i] = randomTrytes(benchTritsLen / consts.TritsPerTryte)
	}
	b.ResetTimer()

	for i := range data {
		_ = EncodeTrytes(data[i])
	}
}

func BenchmarkDecode(b *testing.B) {
	data := make([][]byte, b.N)
	for i := range data {
		data[i] = make([]byte, benchBytesLen)
		Encode(data[i], randomTrits(benchTritsLen))
	}
	b.ResetTimer()

	dst := make(trinary.Trits, benchTritsLen)
	for i := range data {
		_, _ = Decode(dst, data[i])
	}
}

func BenchmarkDecodeToTrytes(b *testing.B) {
	data := make([][]byte, b.N)
	for i := range data {
		data[i] = make([]byte, benchBytesLen)
		Encode(data[i], randomTrits(benchTritsLen))
	}
	b.ResetTimer()

	for i := range data {
		_, _ = DecodeToTrytes(data[i])
	}
}

func randomTrits(n int) trinary.Trits {
	trytes := randomTrytes(n/3 + 1)
	return trinary.MustTrytesToTrits(trytes)[:n]
}

func randomTrytes(n int) trinary.Trytes {
	var result strings.Builder
	result.Grow(n)
	for i := 0; i < n; i++ {
		result.WriteByte(consts.TryteAlphabet[rand.Intn(len(consts.TryteAlphabet))])
	}
	return result.String()
}
