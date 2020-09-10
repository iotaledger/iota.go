// Package checksums provides functions for adding/removing checksums from supplied Trytes.
package checksum

import (
	"github.com/iotaledger/iota.go/guards"
	"github.com/iotaledger/iota.go/legacy"
	"github.com/iotaledger/iota.go/legacy/kerl"
	. "github.com/iotaledger/iota.go/legacy/trinary"
)

// AddChecksum computes the checksums and returns the given trytes with the appended checksums.
// If isAddress is true, then the input trytes must be of length HashTrytesSize.
// Specified checksums length must be at least MinChecksumTrytesSize long or it must be
// AddressChecksumTrytesSize if isAddress is true.
func AddChecksum(input Trytes, isAddress bool, checksumLength uint64) (Trytes, error) {
	if isAddress && len(input) != legacy.HashTrytesSize {
		if len(input) == legacy.AddressWithChecksumTrytesSize {
			return input, nil
		}
		return "", legacy.ErrInvalidAddress
	}

	if checksumLength < legacy.MinChecksumTrytesSize ||
		(isAddress && checksumLength != legacy.AddressChecksumTrytesSize) {
		return "", legacy.ErrInvalidChecksum
	}

	inputCopy := input

	for len(inputCopy)%legacy.HashTrytesSize != 0 {
		inputCopy += "9"
	}

	k := kerl.NewKerl()
	if err := k.AbsorbTrytes(inputCopy); err != nil {
		return "", err
	}
	checksumTrytes, err := k.SqueezeTrytes(legacy.HashTrinarySize)
	if err != nil {
		return "", err
	}
	input += checksumTrytes[legacy.HashTrytesSize-checksumLength : legacy.HashTrytesSize]
	return input, nil
}

// AddChecksums is a wrapper function around AddChecksum for multiple trytes strings.
func AddChecksums(inputs []Trytes, isAddress bool, checksumLength uint64) ([]Trytes, error) {
	withChecksums := make([]Trytes, len(inputs))
	for i, s := range inputs {
		t, err := AddChecksum(s, isAddress, checksumLength)
		if err != nil {
			return nil, err
		}
		withChecksums[i] = t
	}
	return withChecksums, nil
}

// RemoveChecksum removes the checksums from the given trytes.
// The input trytes must be of length HashTrytesSize or AddressWithChecksumTrytesSize.
func RemoveChecksum(input Trytes) (Trytes, error) {
	if !guards.IsTrytesOfExactLength(input, legacy.HashTrytesSize) &&
		!guards.IsTrytesOfExactLength(input, legacy.AddressWithChecksumTrytesSize) {
		return "", legacy.ErrInvalidAddress
	}
	return input[:legacy.HashTrytesSize], nil
}

// RemoveChecksums is a wrapper function around RemoveChecksum for multiple trytes strings.
func RemoveChecksums(inputs []Trytes) ([]Trytes, error) {
	withoutChecksums := make([]Trytes, len(inputs))
	for i, s := range inputs {
		t, err := RemoveChecksum(s)
		if err != nil {
			return nil, err
		}
		withoutChecksums[i] = t
	}
	return withoutChecksums, nil
}
