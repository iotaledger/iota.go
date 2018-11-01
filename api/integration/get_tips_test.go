package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strings"
)

var _ = Describe("GetTips()", func() {

	api, err := ComposeAPI(HTTPClientSettings{}, nil)
	if err != nil {
		panic(err)
	}

	It("resolves to correct response", func() {
		tips, err := api.GetTips()
		Expect(err).ToNot(HaveOccurred())
		Expect(tips).To(Equal(Hashes{strings.Repeat("T", 81), strings.Repeat("U", 81)}))
	})

})
