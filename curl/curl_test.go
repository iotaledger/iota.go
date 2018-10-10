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
		func(in Trytes, expected Trytes) {
			Expect(HashTrytes(in)).To(Equal(expected))
		},
		Entry("normal trytes", "A", "TJVKPMTAMIZVBVHIVQUPTKEMPROEKV9SB9COEDQYRHYPTYSKQIAN9PQKMZHCPO9TS9BHCORFKW9CQXZEE"),
		Entry("normal trytes #2", "B", "QFZXTJUJNLAOSZKXXMMGJJLFACVLRQMRBKOJLMTZXPLPVDSWWWXLBX9CDZWHMDMSDMDQKXQGEWPC9BJHN"),
		Entry("normal trytes #3", "ABCDEFGHIJ", "JKSGOZW9WFTALAYESGNJYRGCKIMZSVBMFIIHYBFCUCSLWDI9EEPTZBLGWNPJOMW9HZWNOFGBR9RNHKCYI"),
		Entry("empty trytes", "", "999999999999999999999999999999999999999999999999999999999999999999999999999999999"),
	)

})
