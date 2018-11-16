package trinary_examples_test

import (
	"fmt"
	"github.com/iotaledger/iota.go/trinary"
)

// i req: t, The Trit value to check.
// o: bool, Whether the Trit is valid.
func ExampleValidTrit() {}

// i req: t, The Trits to check.
// o: bool, Whether the Trits are valid.
func ExampleValidTrits() {}

// i req: t, The Trits value to convert.
// o: Trits, The validated Trits.
// o: error, Returned for invalid Trit values.
func ExampleNewTrits() {}

// i req: a, First Trits slice input.
// i req: b, Second Trits slice input.
// o: bool, Whether the Trits slices are equal.
// o: error, Returned for invalid Trits.
func ExampleTritsEqual() {}

// i req: value, The int value to convert to Trits.
// o: Trits, The Trits representation of the int value.
func ExampleIntToTrits() {
	trits := trinary.IntToTrits(12)
	fmt.Println(trits)
	// output: {0, 1, 1}
}

// i req: t, The Trits to convert to an int value.
// o: Trits, The int of the Trits.
func ExampleTritsToInt() {
	val := trinary.TritsToInt(trinary.Trits{0, 1, 1})
	fmt.Println(val)
	// output: 12
}

// i req: trits, The Trits to check for conversion.
// o: bool, Whether the Trits can be converted to Trytes.
func ExampleCanTritsToTrytes() {}

// i req: trits, The Trits to check for trailing zeros.
// o: int64, The amunt of trailing zeros.
func ExampleTrailingZeros() {}

// i req: trits, The Trits to convert to Trytes.
// o: Trytes, The Trytes representation of the Trits.
// o: error, Returned for invalid Trits.
func ExampleTritsToTrytes() {}

// i req: trits, The Trits to convert to Trytes.
// o: Trytes, The Trytes representation of the Trits.
func ExampleMustTritsToTrytes() {}

// i req: trits, The Trits to check for conversion.
// o: bool, Whether the Trits can be a Hash.
func ExampleCanBeHash() {}

// i req: trytes, The Trytes to convert to []byte.
// o: []byte, The []byte representation of the Trytes.
// o: error, Returned for invalid Trytes.
func ExampleTrytesToBytes() {}

// i req: bytes, The []byte to convert to Trytes.
// o: Trytes, The Trytes representation of the []byte.
// o: error, Returned for invalid []byte.
func ExampleBytesToTrytes() {}

// i req: trits, The Trits to convert to []byte.
// o: []byte, The []byte representation of the Trits.
// o: error, Returned for invalid Trits.
func ExampleTritsToBytes() {}

// i req: b, The []byte to convert to Trits.
// o: Trits, The Trits representation of []byte.
// o: error, Returned for invalid []byte.
func ExampleBytesToTrits() {}

// i req: trits, The Trits slice to reverse.
// o: Trits, The reversed order Trits slice.
func ExampleReverseTrits() {}

// i req: trytes, The Trytes to validate.
// o: error, Whether the Trytes are valid.
func ExampleValidTrytes() {}

// i req: t, The Trytes to validate.
// o: error, Whether the Tryte is valid.
func ExampleValidTryte() {}

// i req: s, The string to convert to Trytes.
// o: Trytes, The converted Trytes.
// o: error, Returned for invalid trytes.
func ExampleNewTrytes() {}

// i req: trytes, The Trytes to convert to Trits.
// o: Trits, The Trits representation of the given Trytes.
// o: error, Returned for invalid Trytes.
func ExampleTrytesToTrits() {}

// i req: trytes, The Trytes to convert to Trits.
// o: Trits, The Trits representation of the given Trytes.
func ExampleMustTrytesToTrits() {}

// i req: trytes, The Trytes to pad.
// i req: size, The size up to which to pad.
// o: Trytes, The padded Trytes.
func ExamplePad() {}

// i req: trits, The Trits to pad.
// i req: size, The size up to which to pad.
// o: Trits, The padded Trits.
func ExamplePadTrits() {}

// i req: a, First Trits slice.
// i req: b, Second Trits slice.
// o: Trits, The sum of both Trits slices.
func ExampleAddTrits() {}
