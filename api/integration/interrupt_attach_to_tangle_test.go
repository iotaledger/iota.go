package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InterruptAttachToTangle()", func() {

	api, err := ComposeAPI(HTTPClientSettings{}, nil)
	if err != nil {
		panic(err)
	}

	Context("call", func() {
		It("resolves to correct response", func() {
			Expect(api.InterruptAttachToTangle()).ToNot(HaveOccurred())
		})
	})

})
