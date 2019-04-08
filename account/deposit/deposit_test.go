package deposit_test

import (
	"github.com/iotaledger/iota.go/account/deposit"
	"github.com/iotaledger/iota.go/checksum"
	"github.com/iotaledger/iota.go/consts"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
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

	actualMagnetLink := "iota://AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAWOWRMBLMD/?timeout_at=1552847640&multi_use=0&expected_amount=1000"

	var expAmount uint64 = 1000

	conds := &deposit.CDA{
		Address: addrWithChecksum,
		Conditions: deposit.Conditions{
			MultiUse:       false,
			ExpectedAmount: &expAmount,
			TimeoutAt:      &timeoutAt,
		},
	}

	invalidConds := &deposit.CDA{
		Address: addrWithChecksum,
		Conditions: deposit.Conditions{
			MultiUse:       true,
			ExpectedAmount: &expAmount,
			TimeoutAt:      &timeoutAt,
		},
	}

	Context("Creating a magnet-link from conditions", func() {
		It("returns an error when both multi use and expected amount are set", func() {
			_, err := invalidConds.AsMagnetLink()
			Expect(errors.Cause(err)).To(Equal(deposit.ErrInvalidDepositAddressOptions))
		})

		It("works", func() {
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
