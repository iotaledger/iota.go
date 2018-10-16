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

var _ = Describe("GetTrytes()", func() {

	var api *API
	BeforeEach(func() {
		a, err := ComposeAPI(HttpClientSettings{}, nil)
		if err != nil {
			panic(err)
		}
		api = a
	})

	Context("call", func() {

		tx := "EXWYCOOGTORCOPFDQB9DGQQAMKXPSPLNCETD99TMIVGMCZEJXFHABMXVYGNABUVBWVARSQQSHPGWA9999"

		It("resolves to correct response", func() {
			defer gock.Flush()
			gock.New(DefaultLocalIRIURI).
				Post("/").
				MatchType("json").
				JSON(GetTrytesCommand{
					Command: GetTrytesCmd,
					Hashes:  Hashes{tx},
				}).
				Reply(200).
				JSON(GetTrytesResponse{Trytes: []Trytes{"ABCDEFG"}})

			trytes, err := api.GetTrytes(tx)
			Expect(err).ToNot(HaveOccurred())
			Expect(trytes[0]).To(Equal("ABCDEFG"))
		})
	})

	Context("invalid input", func() {
		It("returns an error for invalid transaction hashes", func() {
			_, err := api.GetTrytes("")
			Expect(errors.Cause(err)).To(Equal(ErrInvalidTransactionHash))
		})
	})

})
