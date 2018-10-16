package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/checksum"
	. "github.com/iotaledger/iota.go/consts"

	. "github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("GetBalances()", func() {

	var api *API
	BeforeEach(func() {
		a, err := ComposeAPI(HttpClientSettings{}, nil)
		if err != nil {
			panic(err)
		}
		api = a
	})

	hash := "UFKDPIQSIGJCKXWJZXAPPOWGSTCENJERGMUKJOWXQDUVNXRKXMEAJCTTZDEC9DUNXKUXEOBLULCBA9999"

	Context("call", func() {

		BeforeEach(func() {
			gock.New(DefaultLocalIRIURI).
				Post("/").
				MatchType("json").
				JSON(GetBalancesCommand{
					Command:   GetBalancesCmd,
					Addresses: Hashes{hash},
					Threshold: 100,
				}).
				Reply(200).
				JSON(GetBalancesResponse{
					Duration: 100, Balances: []string{"100"},
					Milestone:      "PCCO9LDWVGHCOOQIBJQRZHXQKOFHVJSBSBE9V9TCXXZPEAYCLAHHBQKHY9SFUTH9KIV9KYQUHORIA9999",
					MilestoneIndex: 123456,
				})
		})

		It("resolves to correct response", func() {
			defer gock.Flush()

			balances, err := api.GetBalances(Hashes{hash}, 100)
			Expect(err).ToNot(HaveOccurred())
			Expect(*balances).To(Equal(Balances{
				Balances:       []uint64{100},
				Milestone:      "PCCO9LDWVGHCOOQIBJQRZHXQKOFHVJSBSBE9V9TCXXZPEAYCLAHHBQKHY9SFUTH9KIV9KYQUHORIA9999",
				MilestoneIndex: 123456,
			}))
		})

		It("removes the checksum from the addresses", func() {
			defer gock.Flush()

			hashWithChecksum, err := checksum.AddChecksum(hash, true, AddressChecksumTrytesSize)
			Expect(err).ToNot(HaveOccurred())

			balances, err := api.GetBalances(Hashes{hashWithChecksum}, 100)
			Expect(err).ToNot(HaveOccurred())
			Expect(*balances).To(Equal(Balances{
				Balances:       []uint64{100},
				Milestone:      "PCCO9LDWVGHCOOQIBJQRZHXQKOFHVJSBSBE9V9TCXXZPEAYCLAHHBQKHY9SFUTH9KIV9KYQUHORIA9999",
				MilestoneIndex: 123456,
			}))
		})
	})

	Context("invalid input", func() {
		It("returns an error for invalid addresses", func() {
			_, err := api.GetBalances(Hashes{"balalaika"}, 100)
			Expect(errors.Cause(err)).To(Equal(ErrInvalidHash))
		})

		It("returns an error for invalid threshold", func() {
			_, err := api.GetBalances(Hashes{hash}, 101)
			Expect(errors.Cause(err)).To(Equal(ErrInvalidThreshold))
		})
	})

})
