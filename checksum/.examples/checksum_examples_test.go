package checksum_examples_test

import (
	"fmt"
	"github.com/iotaledger/iota.go/checksum"
	"github.com/iotaledger/iota.go/consts"
)

// i req: input, The Trytes of which to compute the checksum of.
// i req: isAddress, Whether to validate the input as an address.
// i req: checksumLength, The wanted length of the checksum. Must be 9 when isAddress is true.
// o: Trytes, The input Trytes with the appended checksum.
// o: error, Returned for invalid addresses and other inputs.
func ExampleAddChecksum() {
	addr := "ZGPO9BSVZHJBLWHHRPOCKMRHLIEIOXQPOMGSDETZINIJGCDEP9QVJED9D9IUHNPPVDINQ9GOSLY9KWZGC"
	addrWithChecksum, err := checksum.AddChecksum(addr, true, consts.AddressChecksumTrytesSize)
	if err != nil {
		// handle error
		return
	}
	fmt.Println(addrWithChecksum) // output: JHCYLIGUW
}

// i req: inputs, The Trytes slice of which to compute the checksum of.
// i req: isAddress, Whether to validate the inputs as addresses.
// o: error, Returned for invalid addresses and other inputs.
// i req: checksumLength, The wanted length of each checksum. Must be 9 when isAddress is true.
func ExampleAddChecksums() {}

// i req: input, The Trytes (which must be 81/90 in length) of which to remove the checksum.
// o: Trytes, The input Trytes without the checksum.
// o: error, Returned for inputs of invalid length.
func ExampleRemoveChecksum() {
	cs := "JHCYLIGUW"
	addr := "ZGPO9BSVZHJBLWHHRPOCKMRHLIEIOXQPOMGSDETZINIJGCDEP9QVJED9D9IUHNPPVDINQ9GOSLY9KWZGC"
	addr += cs
	addrWithoutChecksum, err := checksum.RemoveChecksum(addr)
	if err != nil {
		// handle error
		return
	}
	fmt.Println(addrWithoutChecksum) // output: ZGPO9BSVZHJBLWHHRPOCKMRHLIEIOXQPOMGSDETZINIJGCDEP9QVJED9D9IUHNPPVDINQ9GOSLY9KWZGC
}

// i req: inputs, The Trytes (which must be 81/90 in length) slice of which to remove the checksums.
// o: []Trytes, The input Trytes slice without the checksums.
// o: error, Returned for inputs of invalid length.
func ExampleRemoveChecksums() {}
