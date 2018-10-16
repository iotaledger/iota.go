package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/consts"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("GetTransactionsToApprove()", func() {

	var api *API
	BeforeEach(func() {
		a, err := ComposeAPI(HttpClientSettings{}, nil)
		if err != nil {
			panic(err)
		}
		api = a
	})


	Context("call", func() {

		trunk := "EXWYCOOGTORCOPFDQB9DGQQAMKXPSPLNCETD99TMIVGMCZEJXFHABMXVYGNABUVBWVARSQQSHPGWA9999"
		branch := "JGBOORVOPMYC9BGOJRKHGICCQWLWYNLXCQTNNHYNMTRDLSSSNAQDFOHZBYFL9R9EVYPYHQQRADFXZ9999"

		It("resolves to correct response", func() {
			defer gock.Flush()
			gock.New(DefaultLocalIRIURI).
				Post("/").
				MatchType("json").
				JSON(GetTransactionsToApproveCommand{
					Command: GetTransactionsToApproveCmd,
					Depth:   3, Reference: "",
				}).
				Reply(200).
				JSON(GetTransactionsToApproveResponse{
					Duration: 100,
					TransactionsToApprove: TransactionsToApprove{
						TrunkTransaction:  trunk,
						BranchTransaction: branch,
					},
				})
			gtta, err := api.GetTransactionsToApprove(3)
			Expect(err).ToNot(HaveOccurred())
			Expect(*gtta).To(Equal(TransactionsToApprove{
				TrunkTransaction:  trunk,
				BranchTransaction: branch,
			}))
		})

		It("resolves to correct response with reference option", func() {
			defer gock.Flush()
			gock.New(DefaultLocalIRIURI).
				Post("/").
				MatchType("json").
				JSON(GetTransactionsToApproveCommand{
					Command: GetTransactionsToApproveCmd,
					Depth:   3, Reference: "RZNYHJLXSLRBJIBWXZKWTFZLZGB9QPGCPHOZYASPQVGAGDWZEKDRNMBXRSUYAYBUTBC9GPOSSKTRA9999",
				}).
				Reply(200).
				JSON(GetTransactionsToApproveResponse{
					Duration: 100,
					TransactionsToApprove: TransactionsToApprove{
						TrunkTransaction:  branch, // don't get confused
						BranchTransaction: trunk,
					},
				})
			gtta, err := api.GetTransactionsToApprove(3, "RZNYHJLXSLRBJIBWXZKWTFZLZGB9QPGCPHOZYASPQVGAGDWZEKDRNMBXRSUYAYBUTBC9GPOSSKTRA9999")
			Expect(err).ToNot(HaveOccurred())
			Expect(*gtta).To(Equal(TransactionsToApprove{
				TrunkTransaction:  branch,
				BranchTransaction: trunk,
			}))
		})
	})

	Context("invalid input", func() {
		It("returns an error for invalid reference", func() {
			_, err := api.GetTransactionsToApprove(3, "")
			Expect(errors.Cause(err)).To(Equal(ErrInvalidReferenceHash))
		})
	})

})
