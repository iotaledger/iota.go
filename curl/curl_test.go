package curl_test

import (
	"strings"

	"github.com/iotaledger/iota.go/consts"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/iotaledger/iota.go/curl"
	. "github.com/iotaledger/iota.go/trinary"
)

var _ = Describe("Curl", func() {

	DescribeTable("hash",
		func(in Trytes, expected Trytes, rounds ...CurlRounds) {
			Expect(MustHashTrytes(MustPad(in, consts.HashTrytesSize), rounds...)).To(Equal(expected))
		},
		Entry("Curl-P-81: empty trytes", "", consts.NullHashTrytes, CurlP81),
		Entry("Curl-P-81: normal trytes", "A", "TJVKPMTAMIZVBVHIVQUPTKEMPROEKV9SB9COEDQYRHYPTYSKQIAN9PQKMZHCPO9TS9BHCORFKW9CQXZEE", CurlP81),
		Entry("Curl-P-81: normal trytes #2", "Z", "FA9WYZSJJWSD9AEEBOGGDHFTMIZVHFURFLJLFBTNENDDCMSXGAGLXFMYZTAMKVIYDQSZEDKXSWVAOPZMK"),
		Entry("Curl-P-81: normal trytes #3", "NOPQRSTUVWXYZ9ABSDEFGHIJKLM", "GWFZSXPZPAFSVPEGEIVWOTD9MY9KVP9HYVCIWSJEITEGVOVGQGV99RONTWDXOPUBIQPIWXK9L9OHZYFUB", CurlP81),
		Entry("Curl-P-27: empty trytes", "", consts.NullHashTrytes, CurlP27),
		Entry("Curl-P-27: normal trytes", "TWENTYSEVEN", "RQPYXJPRXEEPLYLAHWTTFRXXUZTV9SZPEVOQ9FZATCXJOZLZ9A9BFXTUBSHGXN9OOA9GWIPGAAWEDVNPN", CurlP27),
	)

	It("Clone", func() {
		a := strings.Repeat("A", consts.HashTrytesSize)
		b := strings.Repeat("B", consts.HashTrytesSize)

		c1 := NewCurlP81()
		err := c1.AbsorbTrytes(a)
		Expect(err).ToNot(HaveOccurred())

		c2 := c1.Clone()
		err = c1.AbsorbTrytes(b)
		Expect(err).ToNot(HaveOccurred())
		err = c2.AbsorbTrytes(b)
		Expect(err).ToNot(HaveOccurred())

		Expect(c2.MustSqueezeTrytes(consts.HashTrinarySize)).To(Equal(c1.MustSqueezeTrytes(consts.HashTrinarySize)))
	})

})
