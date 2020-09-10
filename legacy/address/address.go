// Package address provides primitives for generating and validating addresses (with and without checksum).
package address

import (
	"github.com/iotaledger/iota.go/legacy"
	"github.com/iotaledger/iota.go/legacy/checksum"
	"github.com/iotaledger/iota.go/legacy/kerl"
	"github.com/iotaledger/iota.go/legacy/signing"
	"github.com/iotaledger/iota.go/legacy/signing/key"
	"github.com/iotaledger/iota.go/legacy/trinary"
)

// Checksum returns the checksum of the given address.
func Checksum(address trinary.Hash) (trinary.Trytes, error) {
	if len(address) < 81 {
		return "", legacy.ErrInvalidAddress
	}

	addressWithChecksum, err := checksum.AddChecksum(address[:81], true, 9)
	if err != nil {
		return "", err
	}
	return addressWithChecksum[legacy.AddressWithChecksumTrytesSize-legacy.AddressChecksumTrytesSize : 90], nil
}

// ValidAddress checks whether the given address is valid.
func ValidAddress(address trinary.Hash) error {
	switch len(address) {
	case 90:
		if err := ValidChecksum(address[:81], address[81:]); err != nil {
			return err
		}
	case 81:
	default:
		return legacy.ErrInvalidAddress
	}
	return trinary.ValidTrytes(address)
}

// ValidChecksum checks whether the given checksum corresponds to the given address.
func ValidChecksum(address trinary.Hash, checksum trinary.Trytes) error {
	actualChecksum, err := Checksum(address)
	if err != nil {
		return err
	}
	if checksum != actualChecksum {
		return legacy.ErrInvalidChecksum
	}
	return nil
}

// GenerateAddress generates an address deterministically, according to the given seed, index and security level.
func GenerateAddress(seed trinary.Trytes, index uint64, secLvl legacy.SecurityLevel, addChecksum ...bool) (trinary.Hash, error) {
	for len(seed)%81 != 0 {
		seed += "9"
	}

	if secLvl == 0 {
		secLvl = legacy.SecurityLevelMedium
	}

	// use Kerl for the entire address generation
	h := kerl.NewKerl()

	subseed, err := signing.Subseed(seed, index, h)
	if err != nil {
		return "", err
	}

	prvKey, err := key.Sponge(subseed, secLvl, h)
	if err != nil {
		return "", err
	}

	digests, err := signing.Digests(prvKey, h)
	if err != nil {
		return "", err
	}

	addressTrits, err := signing.Address(digests, h)
	if err != nil {
		return "", err
	}

	address := trinary.MustTritsToTrytes(addressTrits)

	if len(addChecksum) > 0 && addChecksum[0] {
		return checksum.AddChecksum(address, true, 9)
	}

	return address, nil
}

// GenerateAddresses generates N new addresses from the given seed, indices and security level.
func GenerateAddresses(seed trinary.Trytes, start uint64, count uint64, secLvl legacy.SecurityLevel, addChecksum ...bool) (trinary.Hashes, error) {
	addresses := make(trinary.Hashes, count)

	var withChecksum bool
	if len(addChecksum) > 0 && addChecksum[0] {
		withChecksum = true
	}

	var err error
	for i := 0; i < int(count); i++ {
		addresses[i], err = GenerateAddress(seed, start+uint64(i), secLvl, withChecksum)
		if err != nil {
			return nil, err
		}
	}
	return addresses, nil
}
