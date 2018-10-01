package giota_test

import (
	. "github.com/iotaledger/giota/mocks"
	. "github.com/iotaledger/giota/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/iotaledger/giota"
)

var _ = Describe("API", func() {

	Describe("AttachToTangle", func() {

		var api *API
		BeforeEach(func() {
			api = NewAPI(".", AttachToTangleMock{})
		})

		Context("with valid trytes", func() {

			It("should resolve to correct response", func() {
				responseTrytes, err := api.AttachToTangle(
					AttachToTangleCommand.TrunkTransaction,
					AttachToTangleCommand.BranchTransaction,
					AttachToTangleCommand.MinWeightMagnitude,
					AttachToTangleCommand.Trytes,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(responseTrytes).To(Equal(AttachToTangleResponseMock.Trytes))
			})

			It("should not mutate trytes", func() {
				trytes := make([]Trytes, len(AttachToTangleCommand.Trytes))
				copy(trytes, AttachToTangleCommand.Trytes)
				_, err := api.AttachToTangle(
					AttachToTangleCommand.TrunkTransaction,
					AttachToTangleCommand.BranchTransaction,
					AttachToTangleCommand.MinWeightMagnitude,
					AttachToTangleCommand.Trytes,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(trytes).To(Equal(AttachToTangleCommand.Trytes))
			})
		})

		Context("with invalid trytes", func() {
			invalidTrytes := []Trytes{"asdasDSFDAFD"}

			It("should be rejected with correct error for invalid trunk transaction", func() {
				_, err := api.AttachToTangle(
					Hash(invalidTrytes[0]),
					AttachToTangleCommand.BranchTransaction,
					AttachToTangleCommand.MinWeightMagnitude,
					AttachToTangleCommand.Trytes,
				)
				Expect(err).To(Equal(ErrInvalidTrunkTransaction))
			})

			It("should be rejected with correct error for invalid branch transaction", func() {
				_, err := api.AttachToTangle(
					AttachToTangleCommand.TrunkTransaction,
					Hash(invalidTrytes[0]),
					AttachToTangleCommand.MinWeightMagnitude,
					AttachToTangleCommand.Trytes,
				)
				Expect(err).To(Equal(ErrInvalidTrunkTransaction))
			})

			It("should be rejected with invalid transaction trytes", func() {
				_, err := api.AttachToTangle(
					AttachToTangleCommand.TrunkTransaction,
					AttachToTangleCommand.BranchTransaction,
					AttachToTangleCommand.MinWeightMagnitude,
					invalidTrytes,
				)
				Expect(err).To(Equal(ErrInvalidTrytes))
			})

		})
	})

	Describe("CheckConsistency", func() {

		var api *API
		BeforeEach(func() {
			api = NewAPI(".", AttachToTangleMock{})
		})

	})
})
