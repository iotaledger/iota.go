package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("CheckConsistency()", func() {

	var api *API
	a, err := ComposeAPI(HttpClientSettings{}, nil)
	if err != nil {
		panic(err)
	}
	api = a

	Context("call", func() {
		It("resolves to correct response", func() {
			defer gock.Flush()

			hash := "UFKDPIQSIGJCKXWJZXAPPOWGSTCENJERGMUKJOWXQDUVNXRKXMEAJCTTZDEC9DUNXKUXEOBLULCBA9999"
			gock.New(DefaultLocalIRIURI).
				Post("/").
				MatchType("json").
				JSON(CheckConsistencyCommand{Command: CheckConsistencyCmd, Tails: Hashes{hash}}).
				Reply(200).
				JSON(CheckConsistencyResponse{State: true})

			state, _, err := api.CheckConsistency(hash)
			Expect(err).ToNot(HaveOccurred())
			Expect(state).To(BeTrue())
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
