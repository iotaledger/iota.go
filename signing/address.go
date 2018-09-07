package signing

import (
	"errors"
	"github.com/iotaledger/giota/curl"
	"github.com/iotaledger/giota/kerl"
	"github.com/iotaledger/giota/trinary"
)

type Addresses []Address

// Address represents an address without a checksum.
// Don't type cast, use ToAddress instead to check validity.
type Address trinary.Trytes

const (
	ChecksumSize = 9
)

// Error types for address
var (
	ErrInvalidAddressLength = errors.New("addresses without checksum must be 81/243 trytes/trits in length")
	ErrInvalidChecksum      = errors.New("checksum doesn't match address")
)

// NewAddress generates a new address from the given seed without the checksum
func NewAddress(seed trinary.Trytes, index uint, security SecurityLevel) (Address, error) {
	k, err := NewKeyTrits(seed, index, security)
	if err != nil {
		return "", err
	}

	dg, err := Digests(k)
	if err != nil {
		return "", err
	}

	addr, err := AddressFromDigests(dg)
	if err != nil {
		return "", err
	}

	return ToAddress(addr.Trytes())
}

// NewAddresses generates N new addresses from the given seed without a checksum
func NewAddresses(seed trinary.Trytes, start, count uint, security SecurityLevel) ([]Address, error) {
	as := make([]Address, count)

	var err error
	var i uint
	for ; i < count; i++ {
		as[i], err = NewAddress(seed, start+i, security)
		if err != nil {
			return nil, err
		}
	}
	return as, nil
}

// AddressFromDigests calculates the address from the given digests
func AddressFromDigests(digests trinary.Trits) (trinary.Trits, error) {
	k := kerl.NewKerl()
	if err := k.Absorb(digests); err != nil {
		return nil, err
	}
	return k.Squeeze(curl.HashSize)
}

// ToAddress convert trytes (with and without checksum) to address and checks the validity
func ToAddress(t trinary.Trytes) (Address, error) {
	if len(t) == 90 {
		t = t[:81]
	}

	a := Address(t)
	if err := a.IsValid(); err != nil {
		return "", err
	}

	// validate the checksum
	if len(t) == 90 {
		if err := a.IsValidChecksum(t[81:]); err != nil {
			return "", err
		}
	}

	return a, nil
}

// IsValid returns nil if address is valid
func (a Address) IsValid() error {
	if !(len(a) == 81) {
		return ErrInvalidAddressLength
	}

	return trinary.Trytes(a).IsValid()
}

func (a Address) IsValidChecksum(checksum trinary.Trytes) error {
	checksumFromAddress, err := a.Checksum()
	if err != nil {
		return err
	}
	if checksumFromAddress != checksum {
		return ErrInvalidChecksum
	}
	return nil
}

// Checksum returns checksum trytes
func (a Address) Checksum() (trinary.Trytes, error) {
	if len(a) != 81 {
		return "", ErrInvalidAddressLength
	}

	checksumHash, err := a.ChecksumHash()
	if err != nil {
		return "", err
	}
	return checksumHash[81-9 : 81], nil
}

// ChecksumHash hashes the address and returns the 81 trytes long checksum hash
func (a Address) ChecksumHash() (trinary.Trytes, error) {
	k := kerl.NewKerl()
	t := trinary.Trytes(a).Trits()
	if err := k.Absorb(t); err != nil {
		return "", err
	}
	h, err := k.Squeeze(curl.HashSize)
	if err != nil {
		return "", err
	}
	return h.Trytes(), nil
}

// WithChecksum returns the address together with the checksum. (90 trytes)
func (a Address) WithChecksum() (trinary.Trytes, error) {
	if len(a) != 81 {
		return "", ErrInvalidAddressLength
	}

	cu, err := a.Checksum()
	if err != nil {
		return "", err
	}

	return trinary.Trytes(a) + cu, nil
}
