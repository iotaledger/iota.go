package deposit_test

import (
	"github.com/iotaledger/iota.go/account/deposit"
	"github.com/iotaledger/iota.go/checksum"
	"github.com/iotaledger/iota.go/consts"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strings"
	"time"
)

var _ = Describe("Deposit", func() {

	addr := strings.Repeat("A", 81)
	addrWithChecksum, err := checksum.AddChecksum(addr, true, consts.AddressChecksumTrytesSize)
	if err != nil {
		panic(err)
	}
	timeoutAt := time.Date(2019, time.March, 17, 18, 34, 0, 0, time.UTC)

	actualMagnetLink := "iota://AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAYJHYFZJYZ/?timeout_at=1552847640&multi_use=1&expected_amount=1000"

	var expAmount uint64 = 1000
	conds := &deposit.Conditions{
		Address: addrWithChecksum,
		Request: deposit.Request{
			MultiUse:       true,
			ExpectedAmount: &expAmount,
			TimeoutAt:      &timeoutAt,
		},
	}

	Context("Creating a magnet-link from conditions", func() {
		It("works", func() {
			Expect(err).ToNot(HaveOccurred())
			magnetLink, err := conds.AsMagnetLink()
			Expect(err).ToNot(HaveOccurred())
			Expect(magnetLink).To(Equal(actualMagnetLink))
		})
	})

	Context("A valid magnet-link", func() {

		It("parses", func() {
			condsFromMangetLink, err := deposit.ParseMagnetLink(actualMagnetLink)
			Expect(err).ToNot(HaveOccurred())
			Expect(condsFromMangetLink.Address).To(Equal(conds.Address))
			Expect(*condsFromMangetLink.ExpectedAmount).To(Equal(*conds.ExpectedAmount))
			Expect(*condsFromMangetLink.TimeoutAt).To(Equal(*conds.TimeoutAt))
		})

	})

})
