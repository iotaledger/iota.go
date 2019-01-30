package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/samples"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("GetNewAddress()", func() {

	api, err := ComposeAPI(HTTPClientSettings{}, nil)
	if err != nil {
		panic(err)
	}

	Context("call", func() {
		It("resolves to correct address", func() {
			addresses, err := api.GetNewAddress(Seed, GetNewAddressOptions{Index: 0})
			Expect(err).ToNot(HaveOccurred())
			// third address because previous ones are spent or have transactions
			Expect(addresses[0]).To(Equal(SampleAddressesWithChecksum[2]))
		})

		It("resolves to correct addresses with total option", func() {
			var total uint64 = 2
			addresses, err := api.GetNewAddress(Seed, GetNewAddressOptions{Index: 0, Total: &total})
			Expect(err).ToNot(HaveOccurred())
			Expect(addresses[0]).To(Equal(SampleAddressesWithChecksum[0]))
			Expect(addresses[1]).To(Equal(SampleAddressesWithChecksum[1]))
		})

		It("resolves to correct addresses with return all option", func() {
			addresses, err := api.GetNewAddress(Seed, GetNewAddressOptions{Index: 1, ReturnAll: true})
			Expect(err).ToNot(HaveOccurred())
			// index 1 has transactions, 2 is new
			Expect(len(addresses)).To(Equal(2))
			Expect(addresses[0]).To(Equal(SampleAddressesWithChecksum[1]))
			Expect(addresses[1]).To(Equal(SampleAddressesWithChecksum[2]))
		})

		It("resolves to correct addresses with total option from different index", func() {
			addresses, err := api.GetNewAddress(Seed, GetNewAddressOptions{Index: 1, ReturnAll: true})
			Expect(err).ToNot(HaveOccurred())
			Expect(addresses[0]).To(Equal(SampleAddressesWithChecksum[1]))
		})
	})

	Context("invalid input", func() {
		It("returns an error for invalid seed", func() {
			_, err := api.GetNewAddress("asdf", GetNewAddressOptions{})
			Expect(errors.Cause(err)).To(Equal(ErrInvalidSeed))
		})

		It("returns an error for invalid total option", func() {
			var total uint64 = 0
			_, err := api.GetNewAddress(Seed, GetNewAddressOptions{Total: &total})
			Expect(errors.Cause(err)).To(Equal(ErrInvalidTotalOption))
		})
	})

})
