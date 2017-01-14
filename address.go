package giota

import (
	"errors"
)

type Address struct {
	trytes string
}

var (
	ErrInvalidAddressTrytes = errors.New("addresses are either 81 or 90 trytes in length")
	ErrInvalidAddressTrits  = errors.New("addresses are either 243 or 270 trits in length")
)

func NewAddressFromTrytes(addr string) (*Address, error) {
	if !ValidAddressTrytes(addr) {
		return nil, ErrInvalidAddressTrytes
	}

	if len(addr) == 90 {
		return &Address{trytes: addr}, nil
	}

	addrTrits := TrytesToTrits(addr)
	c := &Curl{}
	c.Init(addrTrits)
	_ = c.Squeeze()
	checksum := TritsToTrytes(c.State())[:9]

	return &Address{trytes: addr + checksum}, nil
}

func NewAddressFromTrits(addr []int) (*Address, error) {
	if !ValidAddressTrits(addr) {
		return nil, ErrInvalidAddressTrits
	}

	addrTrytes := TritsToTrytes(addr)
	if len(addr) == 90*3 {
		return &Address{trytes: addrTrytes}, nil
	}

	c := &Curl{}
	c.Init(addr)
	_ = c.Squeeze()
	checksum := TritsToTrytes(c.State())[:9]

	return &Address{trytes: addrTrytes + checksum}, nil
}

func (a *Address) String() string {
	return a.trytes[:81] // The default seems to be to print it without checksum
}

func (a *Address) MarshalJSON() ([]byte, error) {
	return []byte(a.String()), nil
}

func (a *Address) TritsWithChecksum() []int {
	return TrytesToTrits(a.TrytesWithChecksum())
}

func (a *Address) TritsWithoutChecksum() []int {
	return TrytesToTrits(a.TrytesWithoutChecksum())
}

func (a *Address) Checksum() string {
	return a.trytes[81:]
}

func (a *Address) TrytesWithoutChecksum() string {
	return a.trytes[:81]
}

func (a *Address) TrytesWithChecksum() string {
	return a.trytes
}
