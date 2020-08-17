package t5b1_examples_test

import (
	"fmt"

	"github.com/iotaledger/iota.go/encoding/t5b1"
	"github.com/iotaledger/iota.go/trinary"
)

// i req: dst, The target slice for the encoded bytes.
// i req: src, The trits to encode.
// o: int, The number of bytes written.
func ExampleEncode() {
	src := trinary.Trits{1, 1, 1}
	// allocate a slice for the output
	dst := make([]byte, t5b1.EncodedLen(len(src)))
	t5b1.Encode(dst, src)
	fmt.Println(dst)
	// output: [13]
}

// i req: dst, The target slice for the decoded trits.
// i req: src, The bytes to decode.
// o: int, The number of trits written.
// o: error, Returned for non t5b1 encoded inputs.
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
	// output: [1 1 1 0 0]
}

// i req: src, The trytes to encode.
// o: []byte, The encoded bytes.
func ExampleEncodeTrytes() {
	dst := t5b1.EncodeTrytes("MM")
	fmt.Println(dst)
	// output: [121 1]
}

// i req: src, The bytes to decode.
// o: Trytes, The decoded trytes.
// o: error, Returned for non t5b1 encoded inputs.
func ExampleDecodeToTrytes() {
	dst, err := t5b1.DecodeToTrytes([]byte{121, 1})
	if err != nil {
		// handle error
		return
	}
	// as the corresponding trit length will always be a multiple of 5,
	// the trytes might also be padded
	fmt.Println(dst)
	// output: MM99
}
