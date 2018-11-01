// Package checksum provides functions for adding/removing checksums from supplied Trytes.
package checksum

import (
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/guards"
	"github.com/iotaledger/iota.go/kerl"
	. "github.com/iotaledger/iota.go/trinary"
)

// AddChecksum computes the checksum and returns the given trytes with the appended checksum.
// If isAddress is true, then the input trytes must be of length HashTrytesSize.
// Specified checksum length must be at least MinChecksumTrytesSize long or it must be
// AddressChecksumTrytesSize if isAddress is true.
func AddChecksum(input Trytes, isAddress bool, checksumLength uint64) (Trytes, error) {
	if isAddress && len(input) != HashTrytesSize {
		if len(input) == AddressWithChecksumTrytesSize {
			return input, nil
		}
		return "", ErrInvalidAddress
	}

	if checksumLength < MinChecksumTrytesSize ||
		(isAddress && checksumLength != AddressChecksumTrytesSize) {
		return "", ErrInvalidChecksum
	}

	inputCopy := input

	for len(inputCopy)%HashTrytesSize != 0 {
		inputCopy += "9"
	}

	inputTrits := MustTrytesToTrits(inputCopy)
	k := kerl.NewKerl()
	if err := k.Absorb(inputTrits); err != nil {
		return "", err
	}
	checksumTrits, err := k.Squeeze(HashTrinarySize)
	if err != nil {
		return "", err
	}
	input += MustTritsToTrytes(checksumTrits[HashTrinarySize-checksumLength*3 : HashTrinarySize])
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

// RemoveChecksum removes the checksum from the given trytes.
// The input trytes must be of length HashTrytesSize or AddressWithChecksumTrytesSize.
func RemoveChecksum(input Trytes) (Trytes, error) {
	if !guards.IsTrytesOfExactLength(input, HashTrytesSize) &&
		!guards.IsTrytesOfExactLength(input, AddressWithChecksumTrytesSize) {
		return "", ErrInvalidAddress
	}
	return input[:HashTrytesSize], nil
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
