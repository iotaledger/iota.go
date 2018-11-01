package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/samples"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("SendTrytes()", func() {
	api, err := ComposeAPI(HTTPClientSettings{})
	if err != nil {
		panic(err)
	}

	var reqTrytes []Trytes
	BeforeEach(func() {
		reqTrytes = make([]Trytes, len(BundleTrytes))
		copy(reqTrytes, BundleTrytes)
		for i, j := 0, len(reqTrytes)-1; i < j; i, j = i+1, j-1 {
			reqTrytes[i], reqTrytes[j] = reqTrytes[j], reqTrytes[i]
		}
	})

	Context("call", func() {
		It("resolves to correct response", func() {
			bndl, err := api.SendTrytes(reqTrytes, 3, 14)
			Expect(err).ToNot(HaveOccurred())
			Expect(bndl).To(Equal(Bundle))
		})
	})

	Context("invalid input", func() {
		It("returns an error for invalid trytes", func() {
			_, err := api.SendTrytes([]Trytes{"asdf"}, 3, 14)
			Expect(errors.Cause(err)).To(Equal(ErrInvalidTransactionTrytes))
		})
	})

})
