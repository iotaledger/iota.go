package b1t6_examples_test

import (
	"fmt"

	"github.com/iotaledger/iota.go/encoding/b1t6"
	"github.com/iotaledger/iota.go/trinary"
)

// i req: dst, The target slice for the encoded trits.
// i req: src, The bytes to encode.
// o: int, The number of trits written.
func ExampleEncode() {
	src := []byte{127}
	// allocate a slice for the output
	dst := make(trinary.Trits, b1t6.EncodedLen(len(src)))
	b1t6.Encode(dst, src)
	fmt.Println(dst)
	// output: [1 0 -1 -1 -1 1]
}

// i req: dst, The target slice for the decoded bytes.
// i req: src, The trits to decode.
// o: int, The number of bytes written.
// o: error, Returned for non b1t6 encoded inputs.
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
	// output: [127]
}

// i req: src, The bytes to encode.
// o: Trytes, The encoded trytes.
func ExampleEncodeToTrytes() {
	dst := b1t6.EncodeToTrytes([]byte{127})
	fmt.Println(dst)
	// output: SE
}

// i req: src, The trytes to decode.
// o: []byte, The decoded bytes.
// o: error, Returned for non b1t6 encoded inputs.
func ExampleDecodeTrytes() {
	dst, err := b1t6.DecodeTrytes("SE")
	if err != nil {
		// handle error
		return
	}
	fmt.Println(dst)
	// output: [127]
}
