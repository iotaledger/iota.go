package guards_test

import (
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/guards"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strings"
)

var _ = Describe("Guards", func() {

	attachedTrytes := "INBFSTXKHWXTZPUEQ9GXZDMAOHDFTTUWCHPPSNMFEKDIWYYWOSANUCFPV9UBFXTVVHNGSNWOOUHOYOTAWZCQ9FXLYHSONPJSWCGEQHACFXQJYNBPYOW9VVKMJJAYZRUIE9UYJWIPXNMYTBHSRYHDJEKKCVWLUJGLHAZCUHGR9AEEHURNADDA9JVFATSIYAMJ9SITXDLNAPPTRBJFMJJXIYZLMUXRCETSFWTPECLFRAELDBNUDIDQSLYOUIZXBT9OFWPSBGZNWSMYCWGUNOEOLP9NTXIHG9VIHXRQDZTBVXQYFAVUQVAIAEEWAQHLZMAVVWBCIYJJQEXK9HHQBXAZ9WATFSJT9ERNWI9OWBIWWZRRCTDNMTUYZWGAGQLYSQUDWOEQYSVIGYRIYN9RXPYOXPHAZVP9PXQIQATFRYYPJUHZHJLABWTBTMJOGAQDJXMYXJLEMRRWEQNXPMAG9ZZYEKFXCLJWLLSMNJ9DFWPKDRQHPPLQDETHZOFRIUBOT9DSWZBHENSAEFEWFYDKDLTTLFQMPGMAQCBLVHUVPMOBANKMQFJNPAQSNVXBVDRCASHRJOVHPUT9FLZLZIJCQFUO9UARFPNQDPXBZMZIZIA9RXECUOXUHZIRXDIWPDCFEFMCTBDASCLXLVROXIKPJDUWDL9SEMHLLVENYT9UUAWHMSLSHIGYWODACORCEJK9SLRLDJURXECBWC9JDLJIRIRPMFRDBOXJAQORKNJQGFDQMUOCAPVPVVFXPZHUGJSHSQS99CB9HCGCUYPCKOVQXNXURHEDRUZKNSMITRXYHOSHE9NUH99EFSULTYUNCFE9PZEKAVYOMYGIPXPDLLUWSHEHYBTCEON9WVWMNBIANMSIJKOWJDRWVIHJTGIZZUWTFVQZSOLBCBMWHLLTEPOTGDFOYVGRNHAFIVLJQZASNFNGSIMKWPTQRYG9KLCNDHDSBOCF9H9NBOEUTGYCMVIAKPQITP9WOPFGJYZETHOHX9FOODSPIRYITXNCWKJJIAZDFZMHWYZYHKOQRGERYIQDFAMAX9WFFGWD9Q9QCNTQVCRDHWYLORAGOVHJFXHJGRKBTAVJGPHQBKAQXKRFFYCAAHAKCRVEHLIUGLHPMNKVSA9DJZMYOXSEXTB9UQQNDELWJORFDXROIQQTBSZYMVEMM9TI9BLOXXAGOPFABBBCBGEZEWBY9XFLOPA9DTNCSLTJOVDKPTLAMXYDRBDJDZZXLLV9JVMJJPEVAEKRJOV9VXMHZLNPFMPQK9SQBVDONJZLFUCRWVZQIMN9NOAC9MV9NUYVJXCMMXRWHMRVX9YCNF9NHACGFVILU9VWVKMSDPTBJFNNXPYVCOGIL9OGNHOQMWSDOBDXCXUNZX9UYPPAUCLJPMICZMQWUR9OBLNFSOHYDISSKXAJHKKOPQTBYVYAEMGKMMCTTXKKKSBLYKHIFIEZELIAFDAXKANJEKWYFLB9GWGCQJGEYNXSIGSIWNEKRUIYOCHJJFD9GXTVOLQOWZNDZNITGUEKZUMQKBDIVTBDSYMMY9ANOQRHGDVKKHEG9BFEGYZNWT9JKYESQFKNXHCBRHDNNBPANL9CNVLNDJHPAHBAXUFA9AIBFIQFBJTEYSUYFPHEAGMWUJQCASBEYCGAOWSWINOKVPUBRLZPCYPKDLPRAGBFHOVJNGIWISKZYTNMNOPMZNWYLRPWDMJGDTHPWKVAWUJUOVGXJVPWERLYOSPBOBXFLTGZMLAUSIMBITAFIGZSPYVROSFOMFAHWHHUGQC9FES9XGNRDDIDTEPMJRXVIGNXXWSHNNTZIEJKIWKUMBUULZLJUI9PAWBKECEDZJAMPLMAKNVJAVZHXFGDGXNDMSLJSBDKVDCGJRZZSZSVXHDUKTOZSMMQNMJEUJPZBUPHOYUCHODRYJATFIDNGDLDONLQQYADNLKMWGESOC9RYKUXSJAZYQM9JCZMECASHQYVVHBATQKYJT9ZMNUAIIFAZREKOPFWLGZIIYFBBTREWRWYOKVDQIUDSMR9BSKEWTWOF9FXOEJHARRFZVKRAILAUDXYQKVCGKSRVINCZTJHONLFSDBZWDY9YYHOBGBMJSHGAETZBLZ9EN99SXTIYLNKYFJMCHPEQJPAT99JSPXGURDFRWVHVUPQLDREFMVRNZBTXPCNXKIXCZZWXARHSHWZFFNVHCWT9Y9BNRHBYGLIOOHXJRRYWGJTVJWUHWSMAATLDXYKMCUHKQGNOUTTCIVIY9SKZYPDJKVAPZWFLTWACBVZDGW9999999999999999999999999999MINEIOTADOTCOM9999999999999FS9NHZD99B99999999C99999999ZEEAAWQY9KR9WEJNGXJXRTWKLKIAKPVTVDW9YTXOXBXLYHJEQE9UGFJBDKDHADWIBRYOQLNLQA9OILCXWQFDYZ9UOLIAPXUVKSCMMCWINAZABFASQIFQYMQNNOFU9NDGQLAWBXKRXJFIFEFDROXMGSVGFFQVPA9999RVJUYNUXQCSEMMAWZKKEVEDONAFLNJKEHDRPDJJQTTKTSRMQOUSOCEDXGFSUEVOAUHXAHRBTUADUZ9999MINEIOTADOTCOM9999999999999YGNDLGDLE999999999MMMMMMMMMPMJVUCLYDATAMYVEILQUFGLYZ9X"

	Context("IsTrytes()", func() {
		It("should return true for valid trytes", func() {
			Expect(IsTrytes("ABC")).To(BeTrue())
		})

		It("should return false for trytes with spaces", func() {
			Expect(IsTrytes("A B C")).To(BeFalse())
		})

		It("should return false for invalid trytes", func() {
			Expect(IsTrytes("abc")).To(BeFalse())
		})

		It("should return false for empty trytes", func() {
			Expect(IsTrytes("")).To(BeFalse())
		})
	})

	Context("IsTrytesOfExactLength()", func() {
		It("should return true for valid trytes and length", func() {
			Expect(IsTrytesOfExactLength("ABC", 3)).To(BeTrue())
		})

		It("should return false for invalid trytes", func() {
			Expect(IsTrytesOfExactLength("abc", 3)).To(BeFalse())
		})

		It("should return false for empty trytes", func() {
			Expect(IsTrytesOfExactLength("", 0)).To(BeFalse())
		})
	})

	Context("IsTrytesOfMaxLength()", func() {
		It("should return true for valid trytes and length", func() {
			Expect(IsTrytesOfMaxLength("A", 3)).To(BeTrue())
		})

		It("should return false for invalid trytes", func() {
			Expect(IsTrytesOfMaxLength("abc", 3)).To(BeFalse())
		})

		It("should return false for empty trytes", func() {
			Expect(IsTrytesOfMaxLength("", 0)).To(BeFalse())
		})
	})

	Context("IsEmptyTrytes()", func() {
		It("should return true for empty trytes", func() {
			Expect(IsEmptyTrytes("9999")).To(BeTrue())
		})

		It("should return false for non empty trytes", func() {
			Expect(IsEmptyTrytes("A99")).To(BeFalse())
		})
	})

	Context("IsHash()", func() {
		It("should return true for valid hash", func() {
			Expect(IsHash("TXBGJB9NORCEHAAWVCQRC9GQSLQCWUIKDOBYTDKVYY9GUQHPJQMKHGNWRWIFLEBPJNAAIOMUFRFLDQUEC")).To(BeTrue())
		})

		It("should return true for address with checksum", func() {
			Expect(IsHash("TXBGJB9NORCEHAAWVCQRC9GQSLQCWUIKDOBYTDKVYY9GUQHPJQMKHGNWRWIFLEBPJNAAIOMUFRFLDQUECB9UMGFVBD")).To(BeTrue())
		})

		It("should return false for invalid hash", func() {
			Expect(IsHash("ABCD")).To(BeFalse())
		})
	})

	Context("IsTransactionHash()", func() {
		It("should return true for valid transaction hash", func() {
			Expect(IsTransactionHash("USECPDPCNHC9KUZWH9STNLLLMDCAZUNVCSF9BFKVLBZNMATXIFHYH9NWNYUOBO9YNUEDXFWZIDSMZ9999")).To(BeTrue())
		})

		It("should return false for invalid transaction hash", func() {
			Expect(IsTransactionHash("USECPDPCNHC9KUZWH9STNLLLMDCAZUNVCSF9")).To(BeFalse())
		})
	})

	Context("IsTag()", func() {
		It("should return true for valid tag", func() {
			Expect(IsTag("MINEIOTADOTCOM9999999999999")).To(BeTrue())
		})

		It("should return false for invalid tag", func() {
			Expect(IsTag("MIeEIoTADOTCoM9999999999999")).To(BeFalse())
		})
	})

	Context("IsTransactionHashWithMWM()", func() {
		It("should return true for hash with mwm 14", func() {
			Expect(IsTransactionHashWithMWM("USECPDPCNHC9KUZWH9STNLLLMDCAZUNVCSF9BFKVLBZNMATXIFHYH9NWNYUOBO9YNUEDXFWZIDSMZ9999", 14)).To(BeTrue())
		})

		It("should return false for hash with mwm 15 but given mwm of 14", func() {
			Expect(IsTransactionHashWithMWM("USECPDPCNHC9KUZWH9STNLLLMDCAZUNVCSF9BFKVLBZNMATXIFHYH9NWNYUOBO9YNUEDXFWZIDSMZ9999", 15)).To(BeFalse())
		})
	})

	Context("IsTransactionTrytes()", func() {
		It("should return true for valid transaction trytes", func() {
			Expect(IsTransactionTrytes(attachedTrytes)).To(BeTrue())
		})

		It("should return false for invalid transaction trytes", func() {
			trytes := "abc"
			Expect(IsTransactionTrytes(trytes)).To(BeFalse())
		})
	})

	Context("IsTransactionTrytesWithMWM", func() {
		It("should return true for parameters", func() {

			Expect(IsTransactionTrytesWithMWM(attachedTrytes, 14)).To(BeTrue())
		})

		It("should return false for wrong parameters", func() {
			Expect(IsTransactionTrytesWithMWM(attachedTrytes, 15)).To(BeFalse())
		})

		It("should return false for invalid parameters", func() {
			trytes := "abc"
			Expect(IsTransactionTrytesWithMWM(trytes, 15)).To(BeFalse())
		})
	})

	Context("IsAttachedTrytes()", func() {
		It("should return true for attached trytes", func() {
			Expect(IsAttachedTrytes(attachedTrytes)).To(BeTrue())
		})

		It("should return false for non attached trytes", func() {
			trytesCopy := attachedTrytes[(TransactionTrytesSize)-3*HashTrytesSize:] + strings.Repeat("9", 243)
			Expect(IsAttachedTrytes(trytesCopy)).To(BeFalse())
		})

	})

})
