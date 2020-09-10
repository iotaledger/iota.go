package b1t6_test

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/iotaledger/iota.go/encoding/b1t6"
	"github.com/iotaledger/iota.go/legacy/trinary"
	"github.com/stretchr/testify/assert"
)

func TestValidEncodings(t *testing.T) {
	type args struct {
		bytes  []byte
		trytes trinary.Trytes
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "empty", args: args{bytes: []byte{}, trytes: ""}},
		{name: "normal", args: args{bytes: []byte{1}, trytes: "A9"}},
		{name: "min byte value", args: args{bytes: []byte{0}, trytes: "99"}},
		{name: "max byte value", args: args{bytes: []byte{255}, trytes: "Z9"}},
		{name: "max trytes value", args: args{bytes: []byte{127}, trytes: "SE"}},
		{name: "min trytes value", args: args{bytes: []byte{128}, trytes: "GV"}},
		{name: "endianness", args: args{bytes: []byte{0, 1}, trytes: "99A9"}},
		{name: "long", args: args{
			bytes:  bytes.Repeat([]byte{0, 1}, 25),
			trytes: strings.Repeat("99A9", 25)},
		},
		{name: "RFC example I", args: args{
			bytes:  MustDecodeHex("00"),
			trytes: "99"},
		},
		{name: "RFC example II", args: args{
			bytes:  MustDecodeHex("0001027e7f8081fdfeff"),
			trytes: "99A9B9RESEGVHVX9Y9Z9"},
		},
		{name: "RFC example III", args: args{
			bytes:  MustDecodeHex("9ba06c78552776a596dfe360cc2b5bf644c0f9d343a10e2e71debecd30730d03"),
			trytes: "GWLW9DLDDCLAJDQXBWUZYZODBYPBJCQ9NCQYT9IYMBMWNASBEDTZOYCYUBGDM9C9"},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {

			t.Run("Encode()", func(t *testing.T) {
				dst := make(trinary.Trits, b1t6.EncodedLen(len(test.args.bytes)))
				n := b1t6.Encode(dst, test.args.bytes)
				assert.Len(t, dst, n)
				assert.Equal(t, trinary.MustTrytesToTrits(test.args.trytes), dst)
			})

			t.Run("EncodeToTrytes()", func(t *testing.T) {
				dst := b1t6.EncodeToTrytes(test.args.bytes)
				assert.Equal(t, test.args.trytes, dst)
			})

			t.Run("Decode()", func(t *testing.T) {
				src := trinary.MustTrytesToTrits(test.args.trytes)
				dst := make([]byte, b1t6.DecodedLen(len(src)))
				n, err := b1t6.Decode(dst, src)
				assert.NoError(t, err)
				assert.Len(t, dst, n)
				assert.Equal(t, test.args.bytes, dst)
			})

			t.Run("DecodeTrytes()", func(t *testing.T) {
				dst, err := b1t6.DecodeTrytes(test.args.trytes)
				assert.NoError(t, err)
				assert.Equal(t, test.args.bytes, dst)
			})
		})
	}
}

func TestInvalidEncodings(t *testing.T) {
	type args struct {
		bytes  []byte
		trytes trinary.Trytes
	}
	tests := []struct {
		name string
		args args
		err  error
	}{
		{name: "one tryte", args: args{bytes: []byte{}, trytes: "A"}, err: b1t6.ErrInvalidLength},
		{name: "3 trytes", args: args{bytes: []byte{1}, trytes: "A9A"}, err: b1t6.ErrInvalidLength},
		{name: "5 trytes", args: args{bytes: []byte{0, 1}, trytes: "99A9A"}, err: b1t6.ErrInvalidLength},
		{name: "above max group value", args: args{bytes: []byte{}, trytes: "TE"}, err: b1t6.ErrInvalidTrits},
		{name: "below min group value", args: args{bytes: []byte{}, trytes: "FV"}, err: b1t6.ErrInvalidTrits},
		{name: "max trytes value", args: args{bytes: []byte{}, trytes: "MM"}, err: b1t6.ErrInvalidTrits},
		{name: "min trytes value", args: args{bytes: []byte{}, trytes: "NN"}, err: b1t6.ErrInvalidTrits},
		{name: "2nd group invalid", args: args{bytes: []byte{255}, trytes: "Z9TE"}, err: b1t6.ErrInvalidTrits},
		{name: "3rd group invalid", args: args{bytes: []byte{0, 1}, trytes: "99A9AFV"}, err: b1t6.ErrInvalidTrits},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {

			t.Run("Decode()", func(t *testing.T) {
				trits := trinary.MustTrytesToTrits(test.args.trytes)
				dst := make([]byte, b1t6.DecodedLen(len(trits))+10)
				n, err := b1t6.Decode(dst, trits)
				assert.True(t, errors.Is(err, test.err))
				assert.LessOrEqual(t, n, b1t6.DecodedLen(len(trits)))
				assert.Equal(t, test.args.bytes, dst[:n])
			})

			t.Run("DecodeTrytes()", func(t *testing.T) {
				dst, err := b1t6.DecodeTrytes(test.args.trytes)
				assert.True(t, errors.Is(err, test.err))
				assert.Nil(t, dst)
			})
		})
	}
}

func MustDecodeHex(s string) []byte {
	dst, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return dst
}

func ExampleEncode() {
	src := []byte{127}
	// allocate a slice for the output
	dst := make(trinary.Trits, b1t6.EncodedLen(len(src)))
	b1t6.Encode(dst, src)
	fmt.Println(dst)
	// Output: [1 0 -1 -1 -1 1]
}

func ExampleDecode() {
	src := trinary.Trits{1, 0, -1, -1, -1, 1}
	// allocate a slice for the output
	dst := make([]byte, b1t6.DecodedLen(len(src)))
	_, err := b1t6.Decode(dst, src)
	if err != nil {
		// handle error
		return
	}
	fmt.Println(dst)
	// Output: [127]
}

func ExampleEncodeToTrytes() {
	dst := b1t6.EncodeToTrytes([]byte{127})
	fmt.Println(dst)
	// Output: SE
}

func ExampleDecodeTrytes() {
	dst, err := b1t6.DecodeTrytes("SE")
	if err != nil {
		// handle error
		return
	}
	fmt.Println(dst)
	// Output: [127]
}
