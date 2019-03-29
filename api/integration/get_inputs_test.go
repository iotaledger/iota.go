package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/samples"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("GetInputs()", func() {

	api, err := ComposeAPI(HTTPClientSettings{}, nil)
	if err != nil {
		panic(err)
	}

	var inputs = Inputs{
		Inputs: []Input{
			{
				Address:  SampleAddressesWithChecksum[0],
				Balance:  99,
				KeyIndex: 0,
				Security: SecurityLevelMedium,
			},
			{
				Address:  SampleAddressesWithChecksum[2],
				Balance:  1,
				KeyIndex: 2,
				Security: SecurityLevelMedium,
			},
		},
		TotalBalance: 100,
	}

	Context("call", func() {
		It("resolves to correct balance", func() {
			var threshold uint64 = 100
			ins, err := api.GetInputs(Seed, GetInputsOptions{Start: 0, Threshold: &threshold})
			Expect(err).ToNot(HaveOccurred())
			Expect(*ins).To(Equal(inputs))
		})
	})

	Context("invalid input", func() {
		It("returns an error for invalid seed", func() {
			var threshold uint64 = 100
			_, err := api.GetInputs("asdf", GetInputsOptions{Start: 0, Threshold: &threshold})
			Expect(errors.Cause(err)).To(Equal(ErrInvalidSeed))
		})

		It("returns an error for invalid start end option", func() {
			var threshold uint64 = 100
			var end uint64 = 9
			_, err := api.GetInputs(Seed, GetInputsOptions{Start: 10, End: &end, Threshold: &threshold})
			Expect(errors.Cause(err)).To(Equal(ErrInvalidStartEndOptions))
		})
	})

})
