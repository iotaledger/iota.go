package address_example_test

import (
	"fmt"
	"github.com/iotaledger/iota.go/address"
	"github.com/iotaledger/iota.go/consts"
	"log"
	"strings"
)

// i req: address, The address from which to compute the checksum.
// o: Trytes, The checksum of the address.
// o: error, Returned for invalid addresses or checksum errors.
func ExampleChecksum() {
	addr := "ZGPO9BSVZHJBLWHHRPOCKMRHLIEIOXQPOMGSDETZINIJGCDEP9QVJED9D9IUHNPPVDINQ9GOSLY9KWZGC"
	addrWithChecksum, err := address.Checksum(addr)
	if err != nil {
		log.Fatalf("unable to compute checksum: %s", err.Error())
	}
	fmt.Println(addrWithChecksum) // output: JHCYLIGUW
}

// i req: address, The address to check.
// o: error, Returned if the address is invalid.
func ExampleValidAddress() {
	addr := "ZGPO9BSVZHJBLWHHRPOCKMRHLIEIOXQPOMGSDETZINIJGCDEP9QVJED9D9IUHNPPVDINQ9GOSLY9KWZGC"
	if err := address.ValidAddress(addr); err != nil {
		log.Fatalf("invalid address: %s", err.Error())
	}
}

// i req: address, The address to check.
// i req: checksum, The checksum to compare against.
// o: error, Returned if checksums don't match.
func ExampleValidChecksum() {
	addr := "ZGPO9BSVZHJBLWHHRPOCKMRHLIEIOXQPOMGSDETZINIJGCDEP9QVJED9D9IUHNPPVDINQ9GOSLY9KWZGC"
	checksum := "JHCYLIGUW"
	if err := address.ValidChecksum(addr, checksum); err != nil {
		log.Fatal(err.Error())
	}
}

// i req: seed, The seed used for address generation.
// i req: index, The index from which to generate the address from.
// i req: secLvl, The security level used for address generation.
// i: addChecksum, Whether to append the checksum on the returned address.
// o: Hash, The generated address.
// o: error, Returned for any error occurring during address generation.
func ExampleGenerateAddress() {
	seed := strings.Repeat("9", 81)
	var index uint64 = 0
	secLvl := consts.SecurityLevelMedium
	addr, err := address.GenerateAddress(seed, index, secLvl)
	if err != nil {
		log.Fatalf("unable to generate address: %s", err.Error())
	}
	fmt.Println(addr)
	// output: GPB9PBNCJTPGFZ9CCAOPCZBFMBSMMFMARZAKBMJFMTSECEBRWMGLPTYZRAFKUFOGJQVWVUPPABLTTLCIA
}

// i req: seed, The seed used for address generation.
// i req: start, The index from which to start generating addresses.
// i req: count, The amount of addresses to generate.
// i req: secLvl, The security level used for address generation.
// i: addChecksum, Whether to append the checksum on the returned address.
// o: Hashes, The generated addresses.
// o: error, Returned for any error occurring during address generation.
func ExampleGenerateAddresses() {
	seed := strings.Repeat("9", 81)
	var index uint64 = 0
	secLvl := consts.SecurityLevelMedium
	addrs, err := address.GenerateAddresses(seed, index, 2, secLvl)
	if err != nil {
		log.Fatalf("unable to generate addresses: %s", err.Error())
	}
	fmt.Println(addrs)
	// output:
	// [
	// 	GPB9PBNCJTPGFZ9CCAOPCZBFMBSMMFMARZAKBMJFMTSECEBRWMGLPTYZRAFKUFOGJQVWVUPPABLTTLCIA,
	//  GMLRCFYRCWPZTORXSFCEGKXTVQGPFI9W9EJLERYJMEJGIPLNCLIKCCAOKQEFYUYCEUGIZKCSSJL9JD9SC,
	// ]
}
