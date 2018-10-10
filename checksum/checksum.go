package checksum

import (
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/kerl"
	. "github.com/iotaledger/iota.go/trinary"
	"github.com/iotaledger/iota.go/utils"
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
	if isAddress && len(input) != HashTrytesSize{
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
	if !utils.IsTrytesOfExactLength(input, HashTrytesSize) &&
		!utils.IsTrytesOfExactLength(input, AddressWithChecksumTrytesSize) {
		return "", ErrInvalidAddress
	}
	return input[:HashTrytesSize], nil
}
