package curl_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/curl"
	"github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Curl", func() {

	Context("golden", func() {
		var tests []Test

		BeforeSuite(func() {
			b, err := ioutil.ReadFile(filepath.Join("testdata", goldenName+".json"))
			Expect(err).ToNot(HaveOccurred())
			err = json.Unmarshal(b, &tests)
			Expect(err).ToNot(HaveOccurred())
		})

		It("absorb and squeeze trytes", func() {
			for i, tt := range tests {
				By(fmt.Sprintf("test vector: %d", i), func() {
					c := NewCurlP81()
					err := c.AbsorbTrytes(tt.In)
					Expect(err).ToNot(HaveOccurred())
					squeeze, err := c.SqueezeTrytes(len(tt.Hash) * consts.TritsPerTryte)
					Expect(err).ToNot(HaveOccurred())
					Expect(squeeze).To(Equal(tt.Hash))
				})
			}
		})

		It("absorb and squeeze trits", func() {
			for i, tt := range tests {
				By(fmt.Sprintf("test vector: %d", i), func() {
					c := NewCurlP81()
					err := c.Absorb(trinary.MustTrytesToTrits(tt.In))
					Expect(err).ToNot(HaveOccurred())
					squeeze, err := c.Squeeze(len(tt.Hash) * consts.TritsPerTryte)
					Expect(err).ToNot(HaveOccurred())
					Expect(squeeze).To(Equal(trinary.MustTrytesToTrits(tt.Hash)))
				})
			}
		})
	})

	DescribeTable("Hash",
		func(in trinary.Trytes, expHash trinary.Trytes) {

			By("tryte", func() {
				hash, err := HashTrytes(trinary.MustPad(in, consts.HashTrytesSize))
				Expect(err).ToNot(HaveOccurred())
				Expect(hash).To(Equal(expHash))
			})

			By("trits", func() {
				hash, err := HashTrits(trinary.MustPadTrits(trinary.MustTrytesToTrits(in), consts.HashTrinarySize))
				Expect(err).ToNot(HaveOccurred())
				Expect(hash).To(Equal(trinary.MustTrytesToTrits(expHash)))
			})
		},
		Entry("empty trytes", "", consts.NullHashTrytes),
		Entry("normal trytes", "A", "TJVKPMTAMIZVBVHIVQUPTKEMPROEKV9SB9COEDQYRHYPTYSKQIAN9PQKMZHCPO9TS9BHCORFKW9CQXZEE"),
		Entry("normal trytes #2", "Z", "FA9WYZSJJWSD9AEEBOGGDHFTMIZVHFURFLJLFBTNENDDCMSXGAGLXFMYZTAMKVIYDQSZEDKXSWVAOPZMK"),
		Entry("normal trytes #3", "NOPQRSTUVWXYZ9ABSDEFGHIJKLM", "GWFZSXPZPAFSVPEGEIVWOTD9MY9KVP9HYVCIWSJEITEGVOVGQGV99RONTWDXOPUBIQPIWXK9L9OHZYFUB"),
		Entry("long absorb", strings.Repeat("ABC", consts.TransactionTrytesSize/3), "UHZVKZCGDIPNGFNPBNFZGIM9GAKYLCPTHTRFRXMNDJLZNXSGRPREFWTBKZWVTKV9BISPXEECVIXFJERAC"),
	)

	It("CopyState", func() {
		a := strings.Repeat("A", consts.HashTrytesSize)

		c := NewCurlP81().(*Curl)
		err := c.AbsorbTrytes(a)
		Expect(err).ToNot(HaveOccurred())

		state := make(trinary.Trits, StateSize)
		c.CopyState(state[:])

		Expect(c.MustSqueeze(consts.HashTrinarySize)).To(Equal(state[:consts.HashTrinarySize]))
	})

	It("Reset", func() {
		a := strings.Repeat("A", consts.HashTrytesSize)
		b := strings.Repeat("B", consts.HashTrytesSize)

		c1 := NewCurlP81()
		err := c1.AbsorbTrytes(a)
		Expect(err).ToNot(HaveOccurred())
		_, err = c1.SqueezeTrytes(consts.HashTrinarySize)
		Expect(err).ToNot(HaveOccurred())

		c1.Reset()
		c2 := NewCurlP81()

		err = c1.AbsorbTrytes(b)
		Expect(err).ToNot(HaveOccurred())
		err = c2.AbsorbTrytes(b)
		Expect(err).ToNot(HaveOccurred())

		Expect(c2.MustSqueezeTrytes(consts.HashTrinarySize)).To(Equal(c1.MustSqueezeTrytes(consts.HashTrinarySize)))
	})

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

	It("absorb after squeeze should panic", func() {
		a := strings.Repeat("A", consts.HashTrytesSize)

		c := NewCurlP81()
		err := c.AbsorbTrytes(a)
		Expect(err).ToNot(HaveOccurred())
		_, err = c.SqueezeTrytes(consts.HashTrinarySize)
		Expect(err).ToNot(HaveOccurred())

		absorb := func() { _ = c.AbsorbTrytes(a) }
		Expect(absorb).To(Panic())
	})
})
