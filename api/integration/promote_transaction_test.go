package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("PromoteTransaction()", func() {
	api, err := ComposeAPI(HttpClientSettings{})
	if err != nil {
		panic(err)
	}
	_ = api

	/*
	Context("call", func() {
		It("resolves to correct response", func() {
			bndl, err := api.PromoteTransaction(Bundle[0].Hash, 3, 14, nil, PromoteTransactionOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(bndl).To(Equal(Bundle))
		})
	})

	Context("invalid input", func() {
		It("returns an error for invalid transaction hash", func() {
			_, err := api.ReplayBundle("asdf", 3, 14)
			Expect(errors.Cause(err)).To(Equal(ErrInvalidTransactionHash))
		})
	})
	*/

})
