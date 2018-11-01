package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/samples"
	"github.com/iotaledger/iota.go/consts"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"strings"
)

var _ = Describe("CheckConsistency()", func() {

	var api *API
	a, err := ComposeAPI(HTTPClientSettings{}, nil)
	if err != nil {
		panic(err)
	}
	api = a

	Context("call", func() {
		It("resolves to correct response", func() {
			state, _, err := api.CheckConsistency(DefaultHashes()...)
			Expect(err).ToNot(HaveOccurred())
			Expect(state).To(BeTrue())
		})

		It("inconsistent transactions returns an info and false", func() {
			state, info, err := api.CheckConsistency(append(DefaultHashes(), strings.Repeat("C", 81))...)
			Expect(err).ToNot(HaveOccurred())
			Expect(state).To(BeFalse())
			Expect(info).To(Equal("test response"))
		})
	})

	Context("invalid input", func() {
		It("returns an error for invalid hash", func() {
			_, _, err := api.CheckConsistency("balalaika")
			Expect(errors.Cause(err)).To(Equal(consts.ErrInvalidTransactionHash))
		})

		It("returns an error for empty hash", func() {
			_, _, err := api.CheckConsistency()
			Expect(errors.Cause(err)).To(Equal(consts.ErrInvalidTransactionHash))
		})
	})

})
