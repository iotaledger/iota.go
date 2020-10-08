package curl_test

import (
	"strings"

	"github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/curl"
	"github.com/iotaledger/iota.go/curl/classic"
	"github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Curl", func() {

	DescribeTable("Hash",
		func(in trinary.Trytes, expSqueeze trinary.Trytes, rounds ...CurlRounds) {

			By("tryte", func() {
				c := NewCurl(rounds...)
				err := c.AbsorbTrytes(trinary.MustPad(in, consts.HashTrytesSize))
				Expect(err).ToNot(HaveOccurred())
				squeeze, err := c.SqueezeTrytes(len(expSqueeze) * consts.TritsPerTryte)
				Expect(err).ToNot(HaveOccurred())
				Expect(squeeze).To(Equal(expSqueeze))
			})

			By("trits", func() {
				c := NewCurl(rounds...)
				err := c.Absorb(trinary.MustPadTrits(trinary.MustTrytesToTrits(in), consts.HashTrinarySize))
				Expect(err).ToNot(HaveOccurred())
				squeeze, err := c.Squeeze(len(expSqueeze) * consts.TritsPerTryte)
				Expect(err).ToNot(HaveOccurred())
				Expect(squeeze).To(Equal(trinary.MustTrytesToTrits(expSqueeze)))
			})
		},
		Entry("Curl-P-81: empty trytes", "", consts.NullHashTrytes, CurlP81),
		Entry("Curl-P-81: normal trytes", "A", "TJVKPMTAMIZVBVHIVQUPTKEMPROEKV9SB9COEDQYRHYPTYSKQIAN9PQKMZHCPO9TS9BHCORFKW9CQXZEE", CurlP81),
		Entry("Curl-P-81: normal trytes #2", "Z", "FA9WYZSJJWSD9AEEBOGGDHFTMIZVHFURFLJLFBTNENDDCMSXGAGLXFMYZTAMKVIYDQSZEDKXSWVAOPZMK"),
		Entry("Curl-P-81: normal trytes #3", "NOPQRSTUVWXYZ9ABSDEFGHIJKLM", "GWFZSXPZPAFSVPEGEIVWOTD9MY9KVP9HYVCIWSJEITEGVOVGQGV99RONTWDXOPUBIQPIWXK9L9OHZYFUB", CurlP81),
		Entry("Curl-P-81: long absorb", strings.Repeat("ABC", consts.TransactionTrytesSize/3), "UHZVKZCGDIPNGFNPBNFZGIM9GAKYLCPTHTRFRXMNDJLZNXSGRPREFWTBKZWVTKV9BISPXEECVIXFJERAC", CurlP81),
		Entry("Curl-P-81: long squeeze", "ABC", "LRJMQXFSZSLCIMKZTWFTEIHKWJZMUOHPSOVXZOHOEVHC9D9DROUQGRPTBZWOIJFTMGMXEYKXEJROQLWNUPSFJJRVTLUUJYW9GBQVXNCAUEGEBV9IJQ9TWFDHCFPUUYPCYLACTAIK9UZAJLVXLI9NPGCJN9ICFTEIYY", CurlP81),
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
	It("shows, curl/classic and curl(sb) behave the same, with unusual rounds, cloning, varying absorb and squeeze lengths", func() {
		t1 := trinary.MustPad("GOOGLEABBCDEVILPROOFOFFRONTENDDEVINTEGRATIONCURL", consts.HashTrytesSize)
		t2 := t1
		for _, rounds := range [...]CurlRounds{0, 1, 2, 3, 4, 7, 8, 9, 15, 16, 26, 27, 28, 32, 64, 80, 81, 82, 128, 242, 243, 244, 255, 256, 323, 324, 325, 511, 512} {
			cc := classic.NewCurl(rounds)
			c := NewCurl(rounds)
			err := cc.AbsorbTrytes(t1)
			Expect(err).ToNot(HaveOccurred())
			err = c.AbsorbTrytes(t2)
			Expect(err).ToNot(HaveOccurred())
			err = cc.AbsorbTrytes(t1) // once more
			Expect(err).ToNot(HaveOccurred())
			err = c.AbsorbTrytes(t2)
			Expect(err).ToNot(HaveOccurred())
			cc = cc.Clone()
			c = c.Clone()
			t1, err = cc.SqueezeTrytes((int(rounds) % 5 + 1) * consts.HashTrinarySize)
			Expect(err).ToNot(HaveOccurred())
			t2, err = c.SqueezeTrytes((int(rounds) % 5 + 1) * consts.HashTrinarySize)
			Expect(err).ToNot(HaveOccurred())
			t1, err = cc.SqueezeTrytes((int(rounds) % 5 + 1) * consts.HashTrinarySize) // and again
			Expect(err).ToNot(HaveOccurred())
			t2, err = c.SqueezeTrytes((int(rounds) % 5 + 1) * consts.HashTrinarySize)
			Expect(err).ToNot(HaveOccurred())
		}
		Expect(t1).To(Equal(t2))
	})
	It("on wrong trytes, curl/classic and curl(sb) behave the same", func() {
		t := trinary.MustPad("666", consts.HashTrytesSize)
		c := NewCurl()
		cc := classic.NewCurl()
		err := c.AbsorbTrytes(t)
		Expect(err).To(HaveOccurred())
		err = cc.AbsorbTrytes(t)
		Expect(err).To(HaveOccurred())
	})
	// Testing on 'wrong trits' would fail, because even curl/classic/asm (runs undefined)
	// and curl/glassic/go (panics) show different behaviour,
	// while curl(cb), like bct, takes false trits as zero.
})
