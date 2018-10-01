package signing

import (
	"errors"
	"github.com/iotaledger/giota/curl"
	"github.com/iotaledger/giota/kerl"
	. "github.com/iotaledger/giota/trinary"
)

const (
	AddressChecksumSize = 9
)

// Error types for address
var (
	ErrInvalidAddressLength = errors.New("addresses without checksum must be 81/243 trytes/trits in length")
	ErrInvalidChecksum      = errors.New("checksum doesn't match address")
)

// Address holds the information needed to create an address hash.
type Address struct {
	Seed     Trytes
	Index    uint
	Security SecurityLevel
}
type Addresses []Address

// AddressHash generates the address hash.
func (a *Address) Hash() (AddressHash, error) {
	return NewAddressHash(*a)
}

// PrivateKey generates the private key of the address.
func (a *Address) PrivateKey() (Trytes, error) {
	return NewPrivateKey(a.Seed, a.Index, a.Security)
}

// NewAddresses generates addresses out of the given seed, indices and security level
func NewAddresses(seed Trytes, start uint, end uint, secLvl SecurityLevel) Addresses {
	infos := Addresses{}
	for i := start; i < end; i++ {
		infos = append(infos, Address{Seed: seed, Index: i, Security: secLvl})
	}
	return infos
}

// AddressHash represents an address hash without a checksum.
// Use NewAddressHash() instead of explicit type conversion.
type AddressHash = Hash
type AddressHashes []AddressHash

// NewAddress generates a new address from the given seed without the checksum
func NewAddressHash(a Address) (AddressHash, error) {
	k, err := NewPrivateKeyTrits(a.Seed, a.Index, a.Security)
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

	trytes, err := TritsToTrytes(addr)
	if err != nil {
		return "", err
	}

	return NewAddressHashFromTrytes(trytes)
}

// NewAddressHashes generates N new address hashes from the given seed without a checksum
func NewAddressHashes(seed Trytes, start, count uint, security SecurityLevel) ([]AddressHash, error) {
	as := make([]AddressHash, count)

	var err error
	var i uint
	for ; i < count; i++ {
		as[i], err = NewAddressHash(Address{seed, start + i, security})
		if err != nil {
			return nil, err
		}
	}
	return as, nil
}

// AddressFromDigests calculates the address from the given digests
func AddressFromDigests(digests Trits) (Trits, error) {
	k := kerl.NewKerl()
	if err := k.Absorb(digests); err != nil {
		return nil, err
	}
	return k.Squeeze(curl.HashSize)
}

// NewAddressHashFromTrytes converts trytes (with and without checksum) to an address and checks its validity.
func NewAddressHashFromTrytes(trytes Trytes) (AddressHash, error) {
	if len(trytes) == 90 {
		trytes = trytes[:81]
	}

	a := AddressHash(trytes)
	if err := ValidAddressHash(a); err != nil {
		return "", err
	}

	// validate the checksum
	if len(trytes) == 90 {
		if err := ValidAddressChecksum(a, trytes[81:]); err != nil {
			return "", err
		}
	}

	return a, nil
}

// ValidAddressHash checks whether the given address is valid.
func ValidAddressHash(a AddressHash) error {
	if !(len(a) == 81) {
		return ErrInvalidAddressLength
	}
	return ValidTrytes(a)
}

// ValidAddressChecksum checks whether the given checksum corresponds to the given address.
func ValidAddressChecksum(a AddressHash, checksum Trytes) error {
	checksumFromAddress, err := AddressChecksum(a)
	if err != nil {
		return err
	}
	if checksumFromAddress != checksum {
		return ErrInvalidChecksum
	}
	return nil
}

// AddressChecksum returns checksum trytes.
func AddressChecksum(a AddressHash) (Trytes, error) {
	if len(a) != 81 {
		return "", ErrInvalidAddressLength
	}

	checksumHash, err := AddressChecksumHash(a)
	if err != nil {
		return "", err
	}
	return checksumHash[81-9 : 81], nil
}

// AddressChecksumHash hashes the address hash and returns the 81 trytes long checksum hash.
func AddressChecksumHash(a AddressHash) (Trytes, error) {
	k := kerl.NewKerl()
	t := TrytesToTrits(a)
	if err := k.Absorb(t); err != nil {
		return "", err
	}
	hashTrits, err := k.Squeeze(curl.HashSize)
	if err != nil {
		return "", err
	}
	return TritsToTrytes(hashTrits)
}

// AddressWithChecksum returns the address hash together with the checksum. (90 trytes)
func AddressWithChecksum(a AddressHash) (Trytes, error) {
	if len(a) != 81 {
		return "", ErrInvalidAddressLength
	}

	cu, err := AddressChecksum(a)
	if err != nil {
		return "", err
	}

	return Trytes(a) + cu, nil
}
