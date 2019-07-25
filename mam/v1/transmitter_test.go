package mam_test

import (
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/bundle"
	"github.com/iotaledger/iota.go/consts"
	mam "github.com/iotaledger/iota.go/mam/v1"
	"github.com/iotaledger/iota.go/trinary"
)

var _ = Describe("Transmitter", func() {

	Context("NewTransmitter", func() {

		It("Should return a new transmitter", func() {
			transmitter := mam.NewTransmitter(newFakeAPI(), "seed", consts.SecurityLevelLow)

			Expect(transmitter.Channel().Mode).To(Equal(mam.ChannelModePublic))
			Expect(transmitter.Channel().SideKey).To(Equal(trinary.Trytes("")))
			Expect(transmitter.Channel().SecurityLevel).To(Equal(consts.SecurityLevelLow))
			Expect(transmitter.Channel().Start).To(Equal(uint64(0)))
			Expect(transmitter.Channel().Count).To(Equal(uint64(1)))
			Expect(transmitter.Channel().NextCount).To(Equal(uint64(1)))
			Expect(transmitter.Channel().Index).To(Equal(uint64(0)))
		})

	})

	Context("Subcribe", func() {

		It("Should add a subscription to the transmitter", func() {
			transmitter := mam.NewTransmitter(newFakeAPI(), "seed", consts.SecurityLevelLow)

			transmitter.Subscribe("", mam.ChannelModePublic, "")

			Expect(transmitter.SubscriptionCount()).To(Equal(1))
		})

	})

	Context("SetMode", func() {

		It("Should set the mode to private", func() {
			transmitter := mam.NewTransmitter(newFakeAPI(), "seed", consts.SecurityLevelLow)

			err := transmitter.SetMode(mam.ChannelModePrivate, "")

			Expect(err).NotTo(HaveOccurred())
			Expect(transmitter.Channel().Mode).To(Equal(mam.ChannelModePrivate))
		})

		It("Should set the mode to restricted", func() {
			transmitter := mam.NewTransmitter(newFakeAPI(), "seed", consts.SecurityLevelLow)

			err := transmitter.SetMode(mam.ChannelModeRestricted, "ABC")

			Expect(err).NotTo(HaveOccurred())
			Expect(transmitter.Channel().Mode).To(Equal(mam.ChannelModeRestricted))
			Expect(transmitter.Channel().SideKey).To(Equal(trinary.Trytes("ABC")))
		})

		It("Should error on undefined mode", func() {
			transmitter := mam.NewTransmitter(newFakeAPI(), "seed", consts.SecurityLevelLow)

			err := transmitter.SetMode(555, "")

			Expect(err).To(Equal(mam.ErrUnknownChannelMode))
			Expect(transmitter.Channel().Mode).To(Equal(mam.ChannelModePublic))
			Expect(transmitter.Channel().SideKey).To(BeEmpty())
		})

		It("Should error on missing side key in restricted mode", func() {
			transmitter := mam.NewTransmitter(newFakeAPI(), "seed", consts.SecurityLevelLow)

			err := transmitter.SetMode(mam.ChannelModeRestricted, "")

			Expect(err).To(Equal(mam.ErrNoSideKey))
			Expect(transmitter.Channel().Mode).To(Equal(mam.ChannelModePublic))
			Expect(transmitter.Channel().SideKey).To(BeEmpty())
		})

	})

	Context("Transmit", func() {

		const seed = "TX9XRR9SRCOBMTYDTMKNEIJCSZIMEUPWCNLC9DPDZKKAEMEFVSTEVUFTRUZXEHLULEIYJIEOWIC9STAHW"

		It("Should transmit the given message", func(done Done) {
			fakeAPI := newFakeAPI()
			transmitter := mam.NewTransmitter(fakeAPI, seed, consts.SecurityLevelLow)

			wg := sync.WaitGroup{}
			go func() {
				defer GinkgoRecover()
				wg.Add(1)
				defer wg.Done()

				call := <-fakeAPI.calls
				Expect(call.method).To(Equal("PrepareTransfers"))
				Expect(call.arguments).To(Equal([]interface{}{
					"999999999999999999999999999999999999999999999999999999999999999999999999999999999",
					bundle.Transfers{bundle.Transfer{
						Address: "YKOLXUAMTJGGIVPEWHUCSKKJWIY9PWEFABXYHAWDBTMOPKWNXOOQCKNHADSZP9SOSDFEOXPVTWUWFDQNH",
						Value:   0,
						Message: "AZBCERWNZGTGNPUJYARSZHHYKJYXVRWNZODXYRP9TQVT9XX9MMQIDEJUFWNQGTTIMAMENDXVQYHSUVSBXGRAGYSYMFQQTKBKBBNTPABQVVUIYFAQYZGYQLGALVCJCSFNRRXUEODCO9DXHHNMITGJPLISKXXMTRLGRWDTQMHBXZPRWUCPIJEBPWVERLRNGCZPGERXWVKFLCDSZDQ9WQZIYY9MZHRRUJ9EACMREZSZYPMGQV9DUBFTAJEFPCDKWGQTANVYMIJIJHZGMWFCNCBUIXHXG9CTLDTZBSMBWEHNITUOABCWWCOEPAKRFOSZLGTFRNUFKCYOEPKFUSZQPFYODAEJRSNKAUOOPXPBWNXASWXDLMNCFEQWXDQBPKWIZBXNWBHCKTCUPERYDESMGOEZYUODWMHENRWMVKQEQ9AEQPNAFLNKZUMXMQFLGDZBNVQDZCAGQUFCFFCAFYUVNKQLTAO9BTSOIQ9IQLEAWFYACSNEIYVYWNVKCGCKZNGPLXTDNE9PKJBTXXQBDBKH9ICLBUVDVIGXUVVCSCVSLBKVRXBCJHJCRKEYNHSPXIYMWEIWQAIMTYWLOZQTBQNEXFWUOAOJ9LMBXSQQA9DCUOWOGW9JQYWXIVULMDLFVAFPYE9JGRJQLNQTOYNPK9JSB9BERTEICKVMXMLNNOCHTKVACIKEXUXZTQXGTXQMEB9SLDTOK9IFLRXYSKAFTCZVLPR9ECQVNQN9LWYRD9XLYPCKNBBCIJKCWUVGWMXMFOJLVONBKQGIISFWXAOHGUOWCM9YTCUKXKIEO9PCRBPMIPTQXQCLWMWGJHQJMAZBVHIOS9CZBAIGBX9VEKWGEQSMZVBCZNTKNJG9FEPSNLNYAVXTHLVHDTXGJCBRAXGWKBUPIGMRGFOUYFCLUXNNDDUYTFJPHTZYFUNTZXRJRTNSFYQHBSGRJPAEPFIEQJPYXKUXOUCQOXLPKRDJWAKCGWRWGJGWCOUMDDBCTVVOPQQYQOIIQIBKGMBVBPDFKNJVXVPSMZWVWXPWLUZPSXOEGTBXKJTG9ORUJUSMZXFDMMZWYRNMKGTBLXIPHVW9BSDR9LOWGSBHHBUOSZYLJJDTNJKZQWYMNCCMIVLITMFWXAWWHDCOWBTVCOGLTBILLT9KLHECWIGZVAVR99BJEMTXJRXGLXAZYHIGQOJEPMDVYAVSAPQU9JMSIXKISVTRIEI9KM9UGZKBSRLEGZU9UTJNXQRHTXZNAUDA9DFYPWEFDOOMTZICIXGIJELIUNRHSUVJJQJITOOFBFCPQSLKQKNASCTKIPGJQHVGJWETMQGUVJBISTVXNS9BLGRUWNKZQKUPSFOLO9IJQNDLXVVNZH9NRIYKJGPRJZACHECFQSQXMSGKMSTRRAFJWJJBFZCQAOBITPTMLTVLFAGBBZOYEXDFKWBJNXLUUIOSDWNLWQWUXNXOELSHBVMIMHUOPLJLNFBOQ9VPBXRRNHHEKUHNDLZ9GZZRKEXNGMCXSOPZKEUBHYNTTIHIGHDLBLSCGEQHLWKGHDKJTPXYPYNDWNSJOKKKPAUBDTVLMYBT9ITCIH9TJOPTICXTYQSFTHYX9NCZGM9TOBJLEMIVI9EPJAKKRYSDPSXLHQDWJGHJZFTWXVFVZGJKGGMYJHCW9XLFEN9UUYUUCJGEZIQVT9HKXOSBFINKEV9LWTXDDFDIG9YPWPQODFXHSGNIOBFDROPEJZZECPKOAVXVECWABNPWMQSXYSCJALKMF9HUQXAPWTEXHEJ9EQPZZKCUUBYZSNYGBL9JLIRJNZSVVYUGSVTYGUKQDQNCJFM9L9GW9XKIUGAHDUSAFW9ZURVLYGXWHRLLAKHHDRDF9BVTSNNTRJBPDAUWWFKYAHHUUFCQEEPNVPRFPDAOAZSOXKOBYXQIXER9ZH9JKUTAWAAASZ9NBXNEDSYXGARIMOEOMTENPOVGCZTPQLPAYWHUYOFQKUMMYXKTRYCESVPDUJMVDMVPCXC9GYQ9FSNDQYPDIZJZPMRBXUSKOSDDENXFWMQMELOYSFT9GAPEYRUMHRHNBNJUAIZDUOZPVIFBEMVWXDUPTQQWEQIZTVMOKGXIHACDFFRPUX9CZYLKOAVJFQMJZ9YDLSFFSQAJANI9XSGERSUIFTMQ9AZXHGJGKIGLSXBFZMQVQ9XBOHEQJAYDJBZVIZGSAAUHX9ESEDALVQKHRYCPFEAWBQVTAUBQOANWPFYHKGVQKVOYJQFIYSLRAZPLZJLDZXMYRBN9WTYTOZKHSAAIOFVBITRHRMPXNZXZYEVMKOXDMASIBMQEFBAWISR99IVOUDJFTFENZ",
						Tag:     "",
					}},
					api.PrepareTransfersOptions{},
				}))
				call.returns([]trinary.Trytes{"TRYTES"}, nil)

				call = <-fakeAPI.calls
				Expect(call.method).To(Equal("SendTrytes"))
				Expect(call.arguments).To(Equal([]interface{}{[]trinary.Trytes{"TRYTES"}, uint64(3), uint64(9), ([]trinary.Hash)(nil)}))
				call.returns(bundle.Bundle{}, nil)
			}()

			_, err := transmitter.Transmit("Hello!")
			Expect(err).NotTo(HaveOccurred())

			wg.Wait()
			close(done)
		})

	})

})
