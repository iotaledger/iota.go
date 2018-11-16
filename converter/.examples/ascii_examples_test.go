package converter_examples_test

import (
	"fmt"
	"github.com/iotaledger/iota.go/converter"
)

// i req: s, The ASCII string to convert to Trytes.
// o: Trytes, The Trytes representation of the input ASCII string.
// o: error, Returned for non ASCII string inputs.
func ExampleASCIIToTrytes() {
	trytes, err := converter.ASCIIToTrytes("IOTA")
	if err != nil {
		// handle error
		return
	}
	fmt.Println(trytes) // output: "SBYBCCKB"
}

// i req: trytes, The input Trytes to convert to an ASCII string.
// o: string, The computed ASCII string.
// o: error, Returned for invalid Trytes or odd length inputs.
func ExampleTrytesToASCII() {
	ascii, err := converter.TrytesToASCII("SBYBCCKB")
	if err != nil {
		// handle error
		return
	}
	fmt.Println(ascii) // output: IOTA
}
