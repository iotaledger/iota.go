package deposit

import (
	"fmt"
	"github.com/iotaledger/iota.go/bundle"
	"github.com/iotaledger/iota.go/checksum"
	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/curl"
	. "github.com/iotaledger/iota.go/trinary"
	"github.com/pkg/errors"
	"net/url"
	"strconv"
	"time"
)

// ErrAddressInvalid is returned when an address is invalid when parsed from a serialized form.
var ErrAddressInvalid = errors.New("invalid address")
var ErrMagnetLinkChecksumInvalid = errors.New("magnet-link checksum is invalid")

// Conditions defines a conditional deposit request.
type Conditions struct {
	Request
	Address Hash `json:"address"`
}

// Defines the names of the condition fields in a magnet link.
const (
	MagnetLinkTimeoutField        = "timeout_at"
	MagnetLinkMultiUseField       = "multi_use"
	MagnetLinkExpectedAmountField = "expected_amount"
	MagnetLinkFormat = "iota://%s%s/?%s=%d&%s=%d&%s=%d"
)

// AsMagnetLink converts the conditions into a magnet link URL.
func (dc *Conditions) AsMagnetLink() (string, error) {
	var expectedAmount uint64
	if dc.ExpectedAmount != nil {
		expectedAmount = *dc.ExpectedAmount
	}
	checksum, err := dc.Checksum()
	if err != nil {
		return "", err
	}
	var multiUse int
	if dc.MultiUse {
		multiUse = 1
	}
	return fmt.Sprintf(MagnetLinkFormat,
		dc.Address[:consts.HashTrytesSize],
		checksum[consts.HashTrytesSize-consts.AddressChecksumTrytesSize:consts.HashTrytesSize],
		MagnetLinkTimeoutField, dc.TimeoutAt.Unix(),
		MagnetLinkMultiUseField, multiUse,
		MagnetLinkExpectedAmountField, expectedAmount), nil
}

// AsTransfer converts the conditions into a transfer object.
func (dc *Conditions) AsTransfer() bundle.Transfer {
	return bundle.Transfer{
		Address: dc.Address,
		Value: func() uint64 {
			if dc.ExpectedAmount == nil {
				return 0
			}
			return *dc.ExpectedAmount
		}(),
	}
}

// Checksum returns the checksum of the the CDR.
func (dc *Conditions) Checksum() (Trytes, error) {
	// checksum formula:
	// Checksum = CurlHash(
	// 	CurlHash(address_trits)[:134] +
	// 	PadTrits27(timeout_value_trits) +
	// 	MultiUse(0/1) +
	// 	PadTrits81(amount_value_trits)
	// )
	addrTrits, err := TrytesToTrits(dc.Address[:consts.HashTrytesSize])
	if err != nil {
		return "", err
	}
	c := curl.NewCurl()
	if err := c.Absorb(addrTrits); err != nil {
		return "", err
	}
	addrChecksumTrits, err := c.Squeeze(consts.HashTrinarySize)
	if err != nil {
		return "", err
	}
	timeoutAtTrits := PadTrits(IntToTrits(dc.TimeoutAt.Unix()), 27)
	var expectedAmountTrits Trits
	if dc.ExpectedAmount != nil {
		expectedAmountTrits = PadTrits(IntToTrits(int64(*dc.ExpectedAmount)), 81)
	} else {
		expectedAmountTrits = PadTrits(expectedAmountTrits, 81)
	}
	var multiUse int8
	if dc.MultiUse {
		multiUse = 1
	}
	input := make(Trits, 243)
	input = append(input, addrChecksumTrits[:134]...)
	input = append(input, timeoutAtTrits...)
	input = append(input, multiUse)
	input = append(input, expectedAmountTrits...)
	c.Reset()
	if err := c.Absorb(input); err != nil {
		return "", err
	}
	checksumTrits, err := c.Squeeze(consts.HashTrinarySize)
	if err != nil {
		return "", err
	}
	return TritsToTrytes(checksumTrits)
}

// ParseMagnetLink parses the given magnet link URL.
func ParseMagnetLink(s string) (*Conditions, error) {
	link, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	query := link.Query()
	cond := &Conditions{}
	if len(link.Host) != consts.AddressWithChecksumTrytesSize {
		return nil, errors.Wrap(ErrAddressInvalid, "host/address part of magnet-link must be 90 trytes long")
	}
	addrWithChecksum, err := checksum.AddChecksum(link.Host[:consts.HashTrytesSize], true, consts.AddressChecksumTrytesSize)
	if err != nil {
		return nil, err
	}
	cond.Address = addrWithChecksum
	expiresSeconds, err := strconv.ParseInt(query.Get(MagnetLinkTimeoutField), 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "invalid expire timestamp")
	}
	expire := time.Unix(expiresSeconds, 0).UTC()
	cond.TimeoutAt = &expire
	cond.MultiUse = query.Get(MagnetLinkMultiUseField) == "1"
	expectedAmount, err := strconv.ParseUint(query.Get(MagnetLinkExpectedAmountField), 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "invalid expected amount")
	}
	cond.ExpectedAmount = &expectedAmount

	computedChecksum, err := cond.Checksum()
	if err != nil {
		return nil, err
	}
	magnetLinkChecksum := link.Host[consts.HashTrytesSize:]
	if computedChecksum[consts.HashTrytesSize-consts.AddressChecksumTrytesSize:consts.HashTrytesSize] != magnetLinkChecksum {
		return nil, ErrMagnetLinkChecksumInvalid
	}
	return cond, nil
}

// Request defines a new deposit request against the account.
type Request struct {
	// The time after this deposit address becomes invalid.
	TimeoutAt *time.Time `json:"timeout_at,omitempty" bson:"timeout_at,omitempty"`
	// Whether to expect multiple deposits to this address
	// in the given timeout.
	// If this flag is false, the deposit address is considered
	// in the input selection as soon as one deposit is available
	// (if the expected amount is set and also fulfilled)
	MultiUse bool `json:"multi_use,omitempty" bson:"multi_use,omitempty"`
	// The expected amount which gets deposited.
	// If the timeout is hit, the address is automatically
	// considered in the input selection.
	ExpectedAmount *uint64 `json:"expected_amount,omitempty" bson:"expected_amount,omitempty"`
}
