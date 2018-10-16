package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/consts"

	. "github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("GetInclusionStates()", func() {

	var api *API
	BeforeEach(func() {
		a, err := ComposeAPI(HttpClientSettings{}, nil)
		if err != nil {
			panic(err)
		}
		api = a
	})

	tx := "UNLJBFREZYWGBYLUGMOAOXO9GLNZMBFDVWDLVLKEVWMBCSXTNUATMCAMWVQFQAWLPKFCOLD9JJXCA9999"
	tipMil := "ZIJGAJ9AADLRPWNCYNNHUHRRAC9QOUDATEDQUMTNOTABUVRPTSTFQDGZKFYUUIE9ZEBIVCCXXXLKX9999"

	Context("call", func() {

		BeforeEach(func() {
			gock.New(DefaultLocalIRIURI).
				Post("/").
				MatchType("json").
				JSON(GetInclusionStatesCommand{
					Command: GetInclusionStatesCmd,
					Transactions: Hashes{tx}, Tips: Hashes{tipMil},
				}).
				Reply(200).
				JSON(GetInclusionStatesResponse{States: []bool{true}})
		})

		It("resolves to correct response", func() {
			defer gock.Flush()
			states, err := api.GetInclusionStates(Hashes{tx}, tipMil)
			Expect(err).ToNot(HaveOccurred())
			Expect(states[0]).To(BeTrue())
		})
	})

	Context("invalid input", func() {
		It("returns an error for invalid transaction hashes", func() {
			_, err := api.GetInclusionStates(Hashes{"balalaika"}, tipMil)
			Expect(errors.Cause(err)).To(Equal(ErrInvalidTransactionHash))
		})

		It("returns an error for invalid tip hashes", func() {
			_, err := api.GetInclusionStates(Hashes{tx}, "balalaika")
			Expect(errors.Cause(err)).To(Equal(ErrInvalidTransactionHash))
		})
	})

})
