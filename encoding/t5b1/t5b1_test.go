package t5b1_test

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/iotaledger/iota.go/encoding/t5b1"
	"github.com/iotaledger/iota.go/legacy"
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
		{name: "empty", args: args{trytes: "", bytes: []byte{}}},
		{name: "positive tryte values", args: args{trytes: "9ABCDEFGHIJKLM9", bytes: []byte{0x1b, 0x06, 0x25, 0xb4, 0xc5, 0x54, 0x40, 0x76, 0x04}}},
		{name: "negative tryte values", args: args{trytes: "9NOPQRSTUVWXYZ9", bytes: []byte{0x94, 0x2c, 0xa2, 0x12, 0xea, 0xd1, 0xab, 0xa9, 0x00}}},
		{name: "long", args: args{trytes: strings.Repeat("YZ9AB", 20), bytes: bytes.Repeat([]byte{0xe3, 0x51, 0x12}, 20)}},
		{name: "no padding", args: args{trytes: "MMMMM", bytes: []byte{0x79, 0x79, 0x79}}},
		{name: "1 trit padding", args: args{trytes: "MMM", bytes: []byte{0x79, 0x28}}},
		{name: "2 trit padding", args: args{trytes: "M", bytes: []byte{0x0d}}},
		{name: "3 trit padding", args: args{trytes: "MMMM", bytes: []byte{0x79, 0x79, 0x04}}},
		{name: "4 trit padding", args: args{trytes: "MM", bytes: []byte{0x79, 0x01}}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Run("Encode()", func(t *testing.T) {
				src := trinary.MustTrytesToTrits(test.args.trytes)
				dst := make([]byte, t5b1.EncodedLen(len(src)))
				n := t5b1.Encode(dst, src)
				assert.Len(t, dst, n)
				assert.Equal(t, test.args.bytes, dst)
			})

			t.Run("EncodeTrytes()", func(t *testing.T) {
				dst := t5b1.EncodeTrytes(test.args.trytes)
				assert.Equal(t, test.args.bytes, dst)
			})

			t.Run("Decode()", func(t *testing.T) {
				dst := make(trinary.Trits, t5b1.DecodedLen(len(test.args.bytes)))
				n, err := t5b1.Decode(dst, test.args.bytes)
				assert.NoError(t, err)
				assert.Len(t, dst, n)
				// add expected padding
				paddedLen := t5b1.DecodedLen(t5b1.EncodedLen(len(test.args.trytes) * legacy.TritsPerTryte))
				assert.Equal(t, trinary.MustPadTrits(trinary.MustTrytesToTrits(test.args.trytes), paddedLen), dst)
			})

			t.Run("DecodeToTrytes()", func(t *testing.T) {
				dst, err := t5b1.DecodeToTrytes(test.args.bytes)
				assert.NoError(t, err)
				// add expected padding
				paddedTritLen := t5b1.DecodedLen(t5b1.EncodedLen(len(test.args.trytes) * legacy.TritsPerTryte))
				paddedTryteLen := int(math.Ceil(float64(paddedTritLen) / legacy.TritsPerTryte))
				assert.Equal(t, trinary.MustPad(test.args.trytes, paddedTryteLen), dst)
			})
		})
	}
}

func TestInvalidEncodings(t *testing.T) {
	type args struct {
		bytes []byte
		trits []int8
	}
	tests := []struct {
		name string
		args args
		err  error
	}{
		{
			name: "invalid group value",
			args: args{
				bytes: []byte{0x80},
				trits: []int8{},
			},
			err: legacy.ErrInvalidByte,
		},
		{
			name: "above max group value",
			args: args{
				bytes: []byte{0x7a},
				trits: []int8{},
			},
			err: legacy.ErrInvalidByte,
		},
		{
			name: "below min group value",
			args: args{
				bytes: []byte{0x86},
				trits: []int8{},
			},
			err: legacy.ErrInvalidByte,
		},
		{
			name: "2nd group invalid",
			args: args{
				bytes: []byte{0x79, 0x7a},
				trits: []int8{1, 1, 1, 1, 1},
			},
			err: legacy.ErrInvalidByte,
		},
		{
			name: "3rd group invalid",
			args: args{
				bytes: []byte{0x00, 0x01, 0x7a},
				trits: []int8{0, 0, 0, 0, 0, 1, 0, 0, 0, 0},
			},
			err: legacy.ErrInvalidByte,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {

			t.Run("Decode()", func(t *testing.T) {
				dst := make(trinary.Trits, t5b1.DecodedLen(len(test.args.bytes))+10)
				n, err := t5b1.Decode(dst, test.args.bytes)
				assert.True(t, errors.Is(err, test.err))
				assert.LessOrEqual(t, n, t5b1.DecodedLen(len(test.args.bytes)))
				assert.Equal(t, test.args.trits, dst[:n])
			})

			t.Run("DecodeToTrytes()", func(t *testing.T) {
				dst, err := t5b1.DecodeToTrytes(test.args.bytes)
				assert.True(t, errors.Is(err, test.err))
				assert.Empty(t, dst)
			})
		})
	}
}

func ExampleEncode() {
	src := trinary.Trits{1, 1, 1}
	// allocate a slice for the output
	dst := make([]byte, t5b1.EncodedLen(len(src)))
	t5b1.Encode(dst, src)
	fmt.Println(dst)
	// Output: [13]
}

func ExampleDecode() {
	src := []byte{13}
	// allocate a slice for the output
	dst := make(trinary.Trits, t5b1.DecodedLen(len(src)))
	_, err := t5b1.Decode(dst, src)
	if err != nil {
		// handle error
		return
	}
	// the output length will always be a multiple of 5
	fmt.Println(dst)
	// Output: [1 1 1 0 0]
}

func ExampleEncodeTrytes() {
	dst := t5b1.EncodeTrytes("MM")
	fmt.Println(dst)
	// Output: [121 1]
}

func ExampleDecodeToTrytes() {
	dst, err := t5b1.DecodeToTrytes([]byte{121, 1})
	if err != nil {
		// handle error
		return
	}
	// as the corresponding trit length will always be a multiple of 5,
	// the trytes might also be padded
	fmt.Println(dst)
	// Output: MM99
}
