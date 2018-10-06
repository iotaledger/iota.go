package checksum

import (
	"github.com/iotaledger/iota.go/api_errors"
	"github.com/iotaledger/iota.go/curl"
	"github.com/iotaledger/iota.go/kerl"
	. "github.com/iotaledger/iota.go/trinary"
	"github.com/iotaledger/iota.go/utils"
)

const (
	HashTrytesLength                = 81
	AddressChecksumTrytesLength     = 9
	AddressWithChecksumTrytesLength = HashTrytesLength + AddressChecksumTrytesLength
	MinChecksumLength               = 3
)

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

func AddChecksum(input Trytes, isAddress bool, checksumLength uint64) (Trytes, error) {
	if isAddress && len(input) != curl.HashSizeTrytes {
		if len(input) == AddressWithChecksumTrytesLength {
			return input, nil
		}
		return "", api_errors.ErrInvalidAddress
	}

	if checksumLength < MinChecksumLength ||
		(isAddress && checksumLength != AddressChecksumTrytesLength) {
		return "", api_errors.ErrInvalidChecksum
	}

	inputCopy := input

	for len(inputCopy)%HashTrytesLength != 0 {
		inputCopy += "9"
	}

	inputTrits := TrytesToTrits(inputCopy)
	k := kerl.NewKerl()
	if err := k.Absorb(inputTrits); err != nil {
		return "", err
	}
	checksumTrits, err := k.Squeeze(curl.HashSize)
	if err != nil {
		return "", err
	}
	input += MustTritsToTrytes(checksumTrits[243-checksumLength*3 : 243])
	return input, nil
}

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

func RemoveChecksum(input Trytes) (Trytes, error) {
	if !utils.IsTrytesOfExactLength(input, HashTrytesLength) &&
		!utils.IsTrytesOfExactLength(input, AddressWithChecksumTrytesLength) {
		return "", api_errors.ErrInvalidAddress
	}
	return input[:HashTrytesLength], nil
}
