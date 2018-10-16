package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("GetTips()", func() {

	var api *API
	BeforeEach(func() {
		a, err := ComposeAPI(HttpClientSettings{}, nil)
		if err != nil {
			panic(err)
		}
		api = a
	})

	It("resolves to correct response", func() {
		defer gock.Flush()
		gock.New(DefaultLocalIRIURI).
			Post("/").
			MatchType("json").
			JSON(GetTipsCommand{Command: GetTipsCmd}).
			Reply(200).
			JSON(GetTipsResponse{Hashes: Hashes{
				"HOUERMDAGZQ9BAN9SLRSBSBXJDMTTWWSLJURLQVYVJOSZQQGRWOCOZNBSJZSTCYOQWWHMEW9M9UWZ9999",
				"AHIRYFHBCIVRNLDPYGPASEVGWRBSEQOEKJJNZLLZXWH9KFKMH9HSBMFBN9WTTOIVXOGNEJKWUXSDA9999",
			}})

		tips, err := api.GetTips()
		Expect(err).ToNot(HaveOccurred())
		Expect(tips).To(Equal(Hashes{
			"HOUERMDAGZQ9BAN9SLRSBSBXJDMTTWWSLJURLQVYVJOSZQQGRWOCOZNBSJZSTCYOQWWHMEW9M9UWZ9999",
			"AHIRYFHBCIVRNLDPYGPASEVGWRBSEQOEKJJNZLLZXWH9KFKMH9HSBMFBN9WTTOIVXOGNEJKWUXSDA9999",
		}))
	})

})
