// Package address provides primitives for generating and validating addresses (with and without checksum).
package address

import (
	"bytes"
	"strings"

	"github.com/iotaledger/iota.go/checksum"
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/encoding/b1t6"
	"github.com/iotaledger/iota.go/kerl"
	. "github.com/iotaledger/iota.go/signing"
	"github.com/iotaledger/iota.go/signing/key"
	. "github.com/iotaledger/iota.go/trinary"
	"github.com/pkg/errors"
	"golang.org/x/crypto/blake2b"
)

const (
	// The prefix of migration addresses.
	MigrationAddressPrefix = "TRANSFER"
)

// Checksum returns the checksum of the given address.
func Checksum(address Hash) (Trytes, error) {
	if len(address) < 81 {
		return "", ErrInvalidAddress
	}

	addressWithChecksum, err := checksum.AddChecksum(address[:81], true, 9)
	if err != nil {
		return "", err
	}
	return addressWithChecksum[AddressWithChecksumTrytesSize-AddressChecksumTrytesSize : 90], nil
}

// ValidAddress checks whether the given address is valid.
func ValidAddress(address Hash) error {
	switch len(address) {
	case 90:
		if err := ValidChecksum(address[:81], address[81:]); err != nil {
			return err
		}
	case 81:
	default:
		return ErrInvalidAddress
	}
	return ValidTrytes(address)
}

// ValidChecksum checks whether the given checksum corresponds to the given address.
func ValidChecksum(address Hash, checksum Trytes) error {
	actualChecksum, err := Checksum(address)
	if err != nil {
		return err
	}
	if checksum != actualChecksum {
		return ErrInvalidChecksum
	}
	return nil
}

// GenerateAddress generates an address deterministically, according to the given seed, index and security level.
func GenerateAddress(seed Trytes, index uint64, secLvl SecurityLevel, addChecksum ...bool) (Hash, error) {
	for len(seed)%81 != 0 {
		seed += "9"
	}

	if secLvl == 0 {
		secLvl = SecurityLevelMedium
	}

	// use Kerl for the entire address generation
	h := kerl.NewKerl()

	subseed, err := Subseed(seed, index, h)
	if err != nil {
		return "", err
	}

	prvKey, err := key.Sponge(subseed, secLvl, h)
	if err != nil {
		return "", err
	}

	digests, err := Digests(prvKey, h)
	if err != nil {
		return "", err
	}

	addressTrits, err := Address(digests, h)
	if err != nil {
		return "", err
	}

	address := MustTritsToTrytes(addressTrits)

	if len(addChecksum) > 0 && addChecksum[0] {
		return checksum.AddChecksum(address, true, 9)
	}

	return address, nil
}

// GenerateAddresses generates N new addresses from the given seed, indices and security level.
func GenerateAddresses(seed Trytes, start uint64, count uint64, secLvl SecurityLevel, addChecksum ...bool) (Hashes, error) {
	addresses := make(Hashes, count)

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

// GenerateMigrationAddress generates a migration address from the given Ed25519 raw bytes.
func GenerateMigrationAddress(ed25519Addr [32]byte) (Hash, error) {
	ed25519Checksum := blake2b.Sum256(ed25519Addr[:])
	ed25519Part := append(ed25519Addr[:], ed25519Checksum[:4]...)
	migrAddr := MigrationAddressPrefix + b1t6.EncodeToTrytes(ed25519Part) + "9"
	migrAddrChecksum, err := Checksum(migrAddr)
	if err != nil {
		return "", err
	}
	return migrAddr + migrAddrChecksum, nil
}

// IsMigrationAddress checks whether the given address is a valid migration address by checking that:
//	- it starts with the prefix 'TRANSFER'
//	- it ends with '9'
//	- the 72 trytes after 'TRANSFER' converted with B1T6 resulting in 36 bytes resolves to:
//		- the 32 bytes being the Ed25519 address
//		- the last 4 bytes of the 36 bytes being the Blake2b-256 hash of the Ed25519 address
func IsMigrationAddress(addr Hash) error {
	if !strings.HasPrefix(addr, MigrationAddressPrefix) {
		return errors.Wrapf(ErrInvalidMigrationAddress, "does not start with prefix '%s'", MigrationAddressPrefix)
	}
	if addr[len(addr)-1] != '9' {
		return errors.Wrap(ErrInvalidMigrationAddress, "does not end with '9'")
	}
	ed25519PartTrytes := addr[len(MigrationAddressPrefix) : len(MigrationAddressPrefix)+72]
	ed25519Part, err := b1t6.DecodeTrytes(ed25519PartTrytes)
	if err != nil {
		return errors.Wrapf(ErrInvalidMigrationAddress, "Ed25519 part is not valid B1T6: %s", err)
	}

	// compute hash
	computedChecksum := blake2b.Sum256(ed25519Part[:32])
	if !bytes.Equal(computedChecksum[:4], ed25519Part[32:36]) {
		return errors.Wrap(ErrInvalidMigrationAddress, "Ed25519 checksum doesn't match computed")
	}

	return nil
}
