package mam_test

import (
	"github.com/iotaledger/iota.go/curl"
	. "github.com/iotaledger/iota.go/mam/v1"
	. "github.com/iotaledger/iota.go/trinary"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	payload = "AAMESSAGEFORYOU9AMESSAGEFORYOU9AMESSAGEFORYOU9AMESSAGEFORYOU9AMESSAGEFORYOU9AMESSAGEFORYOU9AMESSAGEFORYOU9AMESSAGEFORYOU9AMESSAGEFORYOU9AMESSAGEFORYOU9AMESSAGEFORYOU9AMESSAGEFORYOU9AMESSAGEFORYOU9AMESSAGEFORYOU9AMESSAGEFORYOU9AMESSAGEFORYOU9AMESSAGEFORYOU9AMESSAGEFORYOU9AMESSAGEFORYOU9AMESSAGEFORYOU9MESSAGEFORYOU9"
	authID  = "MYMERKLEROOTHASH"
)

var _ = Describe("Mask", func() {

	Context("Mask()", func() {
		It("Mask", func() {
			payloadSize := uint64(len(payload))

			payloadTrits := MustTrytesToTrits(payload)
			authIDTrits := MustTrytesToTrits(authID)
			cipherTrits := make(Trits, 3*payloadSize)

			var index int64 = 5

			indexTrits := IntToTrits(index)
			authIDTrits = AddTrits(authIDTrits, indexTrits)

			c := curl.NewCurlP27().(*curl.Curl)
			c.Absorb(authIDTrits)
			Mask(cipherTrits, payloadTrits, 3*payloadSize, c)

			c.Reset()
			c.Absorb(authIDTrits)
			cipherTrits = Unmask(cipherTrits, 3*payloadSize, c)

			Expect(payloadTrits).To(Equal(cipherTrits))
		})
	})
})
