package curl_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/iotaledger/iota.go/curl"
	. "github.com/iotaledger/iota.go/trinary"
)

var _ = Describe("Curl", func() {

	DescribeTable("hash computation",
		func(in Trytes, expected Trytes, rounds ...CurlRounds) {
			Expect(MustHashTrytes(in, rounds...)).To(Equal(expected))
		},
		Entry("normal trytes", "A", "TJVKPMTAMIZVBVHIVQUPTKEMPROEKV9SB9COEDQYRHYPTYSKQIAN9PQKMZHCPO9TS9BHCORFKW9CQXZEE", CurlP81),
		Entry("normal trytes #2", "B", "QFZXTJUJNLAOSZKXXMMGJJLFACVLRQMRBKOJLMTZXPLPVDSWWWXLBX9CDZWHMDMSDMDQKXQGEWPC9BJHN"),
		Entry("normal trytes #3", "ABCDEFGHIJ", "JKSGOZW9WFTALAYESGNJYRGCKIMZSVBMFIIHYBFCUCSLWDI9EEPTZBLGWNPJOMW9HZWNOFGBR9RNHKCYI", CurlP81),
		Entry("empty trytes", "", "999999999999999999999999999999999999999999999999999999999999999999999999999999999", CurlP81),
		Entry("empty trytes", "TWENTYSEVEN", "RQPYXJPRXEEPLYLAHWTTFRXXUZTV9SZPEVOQ9FZATCXJOZLZ9A9BFXTUBSHGXN9OOA9GWIPGAAWEDVNPN", CurlP27),
	)

})
