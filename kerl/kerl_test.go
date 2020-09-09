package kerl_test

import (
	"bytes"
	"encoding/hex"

	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/kerl"
	"github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Kerl", func() {

	Context("hash valid trits", func() {
		DescribeTable("hash computation",
			func(in trinary.Trytes, expected trinary.Trytes) {
				k := NewKerl()
				Expect(k.Absorb(trinary.MustTrytesToTrits(in))).ToNot(HaveOccurred())
				trits, err := k.Squeeze(len(expected) * HashTrinarySize / HashTrytesSize)
				Expect(err).ToNot(HaveOccurred())
				Expect(trinary.MustTritsToTrytes(trits)).To(Equal(expected))
			},
			Entry("normal trytes",
				"HHPELNTNJIOKLYDUW9NDULWPHCWFRPTDIUWLYUHQWWJVPAKKGKOAZFJPQJBLNDPALCVXGJLRBFSHATF9C",
				"DMJWZTDJTASXZTHZFXFZXWMNFHRTKWFUPCQJXEBJCLRZOM9LPVJSTCLFLTQTDGMLVUHOVJHBBUYFD9AXX",
			),
			Entry("normal trytes #2",
				"QAUGQZQKRAW9GKEFIBUD9BMJQOABXBTFELCT9GVSZCPTZOSFBSHPQRWJLLWURPXKNAOWCSVWUBNDSWMPW",
				"HOVOHFEPCIGTOFEAZVXAHQRFFRTPQEEKANKFKIHUKSGRICVADWDMBINDYKRCCIWBEOPXXIKMLNSOHEAQZ",
			),
			Entry("normal trytes #3",
				"MWBLYBSRKEKLDHUSRDSDYZRNV9DDCPN9KENGXIYTLDWPJPKBHQBOALSDH9LEJVACJAKJYPCFTJEROARRW",
				"KXBKXQUZBYZFSYSPDPCNILVUSXOEHQWWWFKZPFCQ9ABGIIQBNLSWLPIMV9LYNQDDYUS9L9GNUIYKYAGVZ",
			),
			Entry("output with non-zero 243rd trit",
				"GYOMKVTSNHVJNCNFBBAH9AAMXLPLLLROQY99QN9DLSJUHDPBLCFFAIQXZA9BKMBJCYSFHFPXAHDWZFEIZ",
				"OXJCNFHUNAHWDLKKPELTBFUCVW9KLXKOGWERKTJXQMXTKFKNWNNXYD9DMJJABSEIONOSJTTEVKVDQEWTW",
			),
			Entry("input with 243-trits",
				"EMIDYNHBWMBCXVDEFOFWINXTERALUKYYPPHKP9JJFGJEIUY9MUDVNFZHMMWZUYUSWAIOWEVTHNWMHANBZ",
				"EJEAOOZYSAWFPZQESYDHZCGYNSTWXUMVJOVDWUNZJXDGWCLUFGIMZRMGCAZGKNPLBRLGUNYWKLJTYEAQX",
			),
			Entry("output with more than 243-trits",
				"9MIDYNHBWMBCXVDEFOFWINXTERALUKYYPPHKP9JJFGJEIUY9MUDVNFZHMMWZUYUSWAIOWEVTHNWMHANBZ",
				"G9JYBOMPUXHYHKSNRNMMSSZCSHOFYOYNZRSZMAAYWDYEIMVVOGKPJBVBM9TDPULSFUNMTVXRKFIDOHUXXVYDLFSZYZTWQYTE9SPYYWYTXJYQ9IFGYOLZXWZBKWZN9QOOTBQMWMUBLEWUEEASRHRTNIQWJQNDWRYLCA",
			),
			Entry("input & output with more than 243-trits",
				"G9JYBOMPUXHYHKSNRNMMSSZCSHOFYOYNZRSZMAAYWDYEIMVVOGKPJBVBM9TDPULSFUNMTVXRKFIDOHUXXVYDLFSZYZTWQYTE9SPYYWYTXJYQ9IFGYOLZXWZBKWZN9QOOTBQMWMUBLEWUEEASRHRTNIQWJQNDWRYLCA",
				"LUCKQVACOGBFYSPPVSSOXJEKNSQQRQKPZC9NXFSMQNRQCGGUL9OHVVKBDSKEQEBKXRNUJSRXYVHJTXBPDWQGNSCDCBAIRHAQCOWZEBSNHIJIGPZQITIBJQ9LNTDIBTCQ9EUWKHFLGFUVGGUWJONK9GBCDUIMAYMMQX"),
		)
	})

	Context("hash valid trytes", func() {
		DescribeTable("hash computation",
			func(in trinary.Trytes, expected trinary.Trytes) {
				k := NewKerl()
				Expect(k.AbsorbTrytes(in)).ToNot(HaveOccurred())
				trytes, err := k.SqueezeTrytes(len(expected) * HashTrinarySize / HashTrytesSize)
				Expect(err).ToNot(HaveOccurred())
				Expect(trytes).To(Equal(expected))
			},
			Entry("normal trytes",
				"HHPELNTNJIOKLYDUW9NDULWPHCWFRPTDIUWLYUHQWWJVPAKKGKOAZFJPQJBLNDPALCVXGJLRBFSHATF9C",
				"DMJWZTDJTASXZTHZFXFZXWMNFHRTKWFUPCQJXEBJCLRZOM9LPVJSTCLFLTQTDGMLVUHOVJHBBUYFD9AXX",
			),
			Entry("normal trytes #2",
				"QAUGQZQKRAW9GKEFIBUD9BMJQOABXBTFELCT9GVSZCPTZOSFBSHPQRWJLLWURPXKNAOWCSVWUBNDSWMPW",
				"HOVOHFEPCIGTOFEAZVXAHQRFFRTPQEEKANKFKIHUKSGRICVADWDMBINDYKRCCIWBEOPXXIKMLNSOHEAQZ",
			),
			Entry("normal trytes #3",
				"MWBLYBSRKEKLDHUSRDSDYZRNV9DDCPN9KENGXIYTLDWPJPKBHQBOALSDH9LEJVACJAKJYPCFTJEROARRW",
				"KXBKXQUZBYZFSYSPDPCNILVUSXOEHQWWWFKZPFCQ9ABGIIQBNLSWLPIMV9LYNQDDYUS9L9GNUIYKYAGVZ",
			),
			Entry("output with non-zero 243rd trit",
				"GYOMKVTSNHVJNCNFBBAH9AAMXLPLLLROQY99QN9DLSJUHDPBLCFFAIQXZA9BKMBJCYSFHFPXAHDWZFEIZ",
				"OXJCNFHUNAHWDLKKPELTBFUCVW9KLXKOGWERKTJXQMXTKFKNWNNXYD9DMJJABSEIONOSJTTEVKVDQEWTW",
			),
			Entry("input with 243-trits",
				"EMIDYNHBWMBCXVDEFOFWINXTERALUKYYPPHKP9JJFGJEIUY9MUDVNFZHMMWZUYUSWAIOWEVTHNWMHANBZ",
				"EJEAOOZYSAWFPZQESYDHZCGYNSTWXUMVJOVDWUNZJXDGWCLUFGIMZRMGCAZGKNPLBRLGUNYWKLJTYEAQX",
			),
			Entry("output with more than 243-trits",
				"9MIDYNHBWMBCXVDEFOFWINXTERALUKYYPPHKP9JJFGJEIUY9MUDVNFZHMMWZUYUSWAIOWEVTHNWMHANBZ",
				"G9JYBOMPUXHYHKSNRNMMSSZCSHOFYOYNZRSZMAAYWDYEIMVVOGKPJBVBM9TDPULSFUNMTVXRKFIDOHUXXVYDLFSZYZTWQYTE9SPYYWYTXJYQ9IFGYOLZXWZBKWZN9QOOTBQMWMUBLEWUEEASRHRTNIQWJQNDWRYLCA",
			),
			Entry("input & output with more than 243-trits",
				"G9JYBOMPUXHYHKSNRNMMSSZCSHOFYOYNZRSZMAAYWDYEIMVVOGKPJBVBM9TDPULSFUNMTVXRKFIDOHUXXVYDLFSZYZTWQYTE9SPYYWYTXJYQ9IFGYOLZXWZBKWZN9QOOTBQMWMUBLEWUEEASRHRTNIQWJQNDWRYLCA",
				"LUCKQVACOGBFYSPPVSSOXJEKNSQQRQKPZC9NXFSMQNRQCGGUL9OHVVKBDSKEQEBKXRNUJSRXYVHJTXBPDWQGNSCDCBAIRHAQCOWZEBSNHIJIGPZQITIBJQ9LNTDIBTCQ9EUWKHFLGFUVGGUWJONK9GBCDUIMAYMMQX"),
		)
	})

	Context("hash valid bytes", func() {
		DescribeTable("hash computation",
			func(in string, expected string) {
				k := NewKerl()
				// absorb
				inBuf, err := hex.DecodeString(in)
				Expect(err).ToNot(HaveOccurred())
				written, err := k.Write(inBuf)
				Expect(written).To(Equal(len(inBuf)))
				Expect(err).ToNot(HaveOccurred())
				// squeeze
				outBuf := make([]byte, hex.DecodedLen(len(expected)))
				read, err := k.Read(outBuf)
				Expect(read).To(Equal(len(outBuf)))
				Expect(err).ToNot(HaveOccurred())
				Expect(hex.EncodeToString(outBuf)).To(Equal(expected))
			},
			Entry("input with less than 48 bytes",
				"ff",
				"08869cad3dc2429eb295195200ad22eb36188452ba65f0e31b2b21bd49b503a7f1d1d61a6df8bff569d3decc9810721b",
			),
			Entry("input with 48 bytes",
				"3f1b9727c967dda4f0ef98032f97483864e0dc9ed391dd5b8bc8133d9ce77fe182fef749de882dace92b655c6bba22df",
				"bec15753fa767d98b59de095a962f472f7b4e15da47d6dd987c4608ccd32f8e2231c3201faca0ee2591f11179c816e30",
			),
			Entry("output with more than 48 bytes",
				"ec3357c2b1f26b6567a80542a65159f3fdc5c4a7ff0d07ff52c14ed39df3cdee8e3b62250b04592ba0beef909e1c430e",
				"be8fc99ed01018cd8e1904d5188ddc62d85edf1aecc61609a820df347cfe7b8bfa928e9c0460854c638fa330cfd0e3f517a7a822b3d0a0fb3d7b05bbe86ae815c8e063b638363351ac87dd62784db4441e1beae32596a224699fd5aeeab61eab",
			),
			Entry("input & output with more than 48 bytes",
				"be8fc99ed01018cd8e1904d5188ddc62d85edf1aecc61609a820df347cfe7b8bfa928e9c0460854c638fa330cfd0e3f517a7a822b3d0a0fb3d7b05bbe86ae815c8e063b638363351ac87dd62784db4441e1beae32596a224699fd5aeeab61eab",
				"4b6e005b96685f95c22b94e5248b2fa06b22062124bea182a2e7d8834f9bcf1a3e0debe180a377f19207404916263040b9acbbca8851297604dcdf1ae0cb11f8444d9f387fdd80993e7158e4efea04630bcf931d90190dc98a4243265eb73c3f",
			),
			Entry("output with non-zero 243rd trit",
				"f229bc41fdbfbef56f0380f4a7c5ca34f640492ec097af2abd7eae8b8b19f08e13acbb5244becd4ee477cb3c17b38eeb",
				"a686e9c70af67a3892e65ea7b0340ee94c9a0a54cafb99aaed9b489500ba260e4f2eeb4161f1ccf0b153b3d9fc8ffd65",
			),
		)
	})

	Context("hash invalid bytes", func() {
		var k *Kerl
		BeforeEach(func() {
			k = NewKerl()
		})

		It("should return an error for Write after Read", func() {
			written, err := k.Write(make([]byte, HashBytesSize))
			Expect(written).To(Equal(HashBytesSize))
			Expect(err).ToNot(HaveOccurred())
			read, err := k.Read(make([]byte, HashBytesSize))
			Expect(read).To(Equal(HashBytesSize))
			Expect(err).ToNot(HaveOccurred())
			// write again
			written, err = k.Write(make([]byte, HashBytesSize))
			Expect(written).To(Equal(0))
			Expect(err).To(Equal(ErrAbsorbAfterSqueeze))
		})
	})

	Context("hash invalid trits", func() {
		var k *Kerl
		BeforeEach(func() {
			k = NewKerl()
		})

		It("should return an error with invalid trits slice length", func() {
			Expect(k.Absorb(trinary.Trits{1, 0, 0, 0, 0, -1})).To(HaveOccurred())
		})

		It("should return an error with 243rd trit set to 1", func() {
			in := make(trinary.Trits, 2*HashTrinarySize)
			in[HashTrinarySize-1] = 1
			Expect(k.Absorb(in)).To(HaveOccurred())
		})

		It("should return an error with 243rd trit set to -1", func() {
			in := make(trinary.Trits, 2*HashTrinarySize)
			in[HashTrinarySize-1] = -1
			Expect(k.Absorb(in)).To(HaveOccurred())
		})

		It("should return an error for Absorb after Squeeze", func() {
			Expect(k.Absorb(make(trinary.Trits, HashTrinarySize))).ToNot(HaveOccurred())
			_, err := k.Squeeze(HashTrinarySize)
			Expect(err).ToNot(HaveOccurred())
			// absorb again
			Expect(k.Absorb(make(trinary.Trits, HashTrinarySize))).To(Equal(ErrAbsorbAfterSqueeze))
		})
	})

	Context("hash invalid trytes", func() {
		var k *Kerl
		BeforeEach(func() {
			k = NewKerl()
		})

		It("should return an error with invalid trits slice length", func() {
			Expect(k.AbsorbTrytes("AR")).To(HaveOccurred())
		})

		It("should return an error with 243rd trit set to 1", func() {
			in := bytes.Repeat([]byte{'9'}, 2*HashTrytesSize)
			in[HashTrytesSize-1] = 'I'
			Expect(k.AbsorbTrytes(trinary.Trytes(in))).To(HaveOccurred())
		})

		It("should return an error with 243rd trit set to -1", func() {
			in := bytes.Repeat([]byte{'9'}, 2*HashTrytesSize)
			in[HashTrytesSize-1] = 'R'
			Expect(k.AbsorbTrytes(trinary.Trytes(in))).To(HaveOccurred())
		})

		It("should return an error for Absorb after Squeeze", func() {
			Expect(k.AbsorbTrytes(NullHashTrytes)).ToNot(HaveOccurred())
			_, err := k.SqueezeTrytes(HashTrinarySize)
			Expect(err).ToNot(HaveOccurred())
			// absorb again
			Expect(k.AbsorbTrytes(NullHashTrytes)).To(Equal(ErrAbsorbAfterSqueeze))
		})
	})

	Context("(*Kerl).Sum", func() {
		DescribeTable("hash computation", func(in string, expected string) {
			k := NewKerl()
			// absorb
			inBuf, err := hex.DecodeString(in)
			Expect(err).ToNot(HaveOccurred())
			written, err := k.Write(inBuf)
			Expect(written).To(Equal(len(inBuf)))
			Expect(err).ToNot(HaveOccurred())
			// append sum to inBuf
			outBuf := k.Sum(inBuf)
			Expect(hex.EncodeToString(outBuf)).To(Equal(in + expected))
			// absorb again
			written, err = k.Write(inBuf)
			Expect(written).To(Equal(len(inBuf)))
			Expect(err).ToNot(HaveOccurred())
			// compare against new Kerl
			k2 := NewKerl()
			k2.Write(inBuf[:hex.DecodedLen(len(in))])
			k2.Write(inBuf)
			Expect(k.Sum(nil)).To(Equal(k2.Sum(nil)))
		},
			Entry("input with less than 48 bytes",
				"ff",
				"08869cad3dc2429eb295195200ad22eb36188452ba65f0e31b2b21bd49b503a7f1d1d61a6df8bff569d3decc9810721b",
			),
			Entry("input with 48 bytes",
				"3f1b9727c967dda4f0ef98032f97483864e0dc9ed391dd5b8bc8133d9ce77fe182fef749de882dace92b655c6bba22df",
				"bec15753fa767d98b59de095a962f472f7b4e15da47d6dd987c4608ccd32f8e2231c3201faca0ee2591f11179c816e30",
			),
			Entry("input with more than 48 bytes",
				"be8fc99ed01018cd8e1904d5188ddc62d85edf1aecc61609a820df347cfe7b8bfa928e9c0460854c638fa330cfd0e3f517a7a822b3d0a0fb3d7b05bbe86ae815c8e063b638363351ac87dd62784db4441e1beae32596a224699fd5aeeab61eab",
				"4b6e005b96685f95c22b94e5248b2fa06b22062124bea182a2e7d8834f9bcf1a3e0debe180a377f19207404916263040",
			),
			Entry("output with non-zero 243rd trit",
				"f229bc41fdbfbef56f0380f4a7c5ca34f640492ec097af2abd7eae8b8b19f08e13acbb5244becd4ee477cb3c17b38eeb",
				"a686e9c70af67a3892e65ea7b0340ee94c9a0a54cafb99aaed9b489500ba260e4f2eeb4161f1ccf0b153b3d9fc8ffd65",
			),
		)
	})

	Context("(*Kerl).Reset", func() {
		It("reset during absorb", func() {
			k1 := NewKerl()
			Expect(k1.Absorb(make(trinary.Trits, HashTrinarySize))).ToNot(HaveOccurred())

			k1.Reset()
			k2 := NewKerl()
			Expect(k1.MustSqueeze(HashTrinarySize)).To(Equal(k2.MustSqueeze(HashTrinarySize)))
		})

		It("reset during squeeze", func() {
			k1 := NewKerl()
			Expect(k1.Absorb(make(trinary.Trits, HashTrinarySize))).ToNot(HaveOccurred())
			k1.MustSqueeze(HashTrinarySize)

			k1.Reset()
			k2 := NewKerl()
			Expect(k1.MustSqueeze(HashTrinarySize)).To(Equal(k2.MustSqueeze(HashTrinarySize)))
		})
	})

	Context("(*Kerl).Clone", func() {
		It("clone during absorb", func() {
			k1, k2 := NewKerl(), NewKerl()
			Expect(k1.Absorb(make(trinary.Trits, HashTrinarySize))).ToNot(HaveOccurred())
			Expect(k2.Absorb(make(trinary.Trits, HashTrinarySize))).ToNot(HaveOccurred())

			k1Clone := k1.Clone()
			Expect(k1.Absorb(make(trinary.Trits, HashTrinarySize))).ToNot(HaveOccurred())

			hash1 := k1.MustSqueeze(HashTrinarySize)
			hash2 := k2.MustSqueeze(HashTrinarySize)
			hash1Clone := k1Clone.MustSqueeze(HashTrinarySize)
			Expect(hash1Clone).To(Equal(hash2))
			Expect(hash1Clone).ToNot(Equal(hash1))
		})

		It("clone during squeeze", func() {
			k1, k2 := NewKerl(), NewKerl()
			Expect(k1.Absorb(make(trinary.Trits, HashTrinarySize))).ToNot(HaveOccurred())
			Expect(k2.Absorb(make(trinary.Trits, HashTrinarySize))).ToNot(HaveOccurred())

			k1.MustSqueeze(HashTrinarySize)
			k2.MustSqueeze(HashTrinarySize)

			k1Clone := k1.Clone()
			k1.MustSqueeze(HashTrinarySize)

			hash1 := k1.MustSqueeze(HashTrinarySize)
			hash2 := k2.MustSqueeze(HashTrinarySize)
			hash1Clone := k1Clone.MustSqueeze(HashTrinarySize)
			Expect(hash1Clone).To(Equal(hash2))
			Expect(hash1Clone).ToNot(Equal(hash1))
		})
	})

})
