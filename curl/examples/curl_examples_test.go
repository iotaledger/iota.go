package examples

import (
	"fmt"
	"github.com/iotaledger/iota.go/curl"
	"github.com/iotaledger/iota.go/trinary"
)

func ExampleNewCurl() {}

// o: Trytes, The 81 Trytes long hash.
func ExampleSqueeze() {}

// i req: in, The Trytes to absorb.
func ExampleAbsorb() {}

func ExampleTransform() {}

func ExampleReset() {}

// i req: trits, The Trits of which to compute the hash of.
func ExampleHashTrits() {
	trytes := "PDFIDVWRXONZSPJJQVZVVMLGSVB"
	trits := trinary.MustTrytesToTrits(trytes)
	tritsHash := curl.HashTrits(trits)
	fmt.Println(tritsHash) // output: [0 1 -1 0 -1 0 -1 1 ...]
}

// i req: trytes, The Trytes of which to compute the hash of.
func ExampleHashTrytes() {
	trytes := "PDFIDVWRXONZSPJJQVZVVMLGSVB"
	hash := curl.HashTrytes(trytes)
	fmt.Println(hash) // output: UXBXSI9LHCPYFFZXOWALCBTUIVXYKMCEDDIFXXGXJ9ZLEWKOTXSGYHPEAD9SXSRAWM9TPPXWZMZSIEKGX
}
