package mam

import (
	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/curl"
	"github.com/iotaledger/iota.go/trinary"
)

const checksumTrinarySize = consts.AddressChecksumTrytesSize * consts.TritsPerTryte

func makeAddress(mode ChannelMode, root trinary.Trits, sideKey trinary.Trytes) (trinary.Trytes, error) {
	if mode == ChannelModePublic {
		return toAddress(root)
	}

	sideKeyTrits, err := trinary.TrytesToTrits(sideKey)
	if err != nil {
		return "", err
	}

	h := curl.NewCurlP81()
	if err := h.Absorb(sideKeyTrits); err != nil {
		return "", err
	}
	if err := h.Absorb(root); err != nil {
		return "", err
	}
	hashedRoot, err := h.Squeeze(consts.HashTrinarySize)
	if err != nil {
		return "", err
	}

	return toAddress(hashedRoot)
}

// create the Curl-P-81 checksum and return the trytes
func toAddress(trits trinary.Trits) (trinary.Trytes, error) {
	h := curl.NewCurlP81()
	if err := h.Absorb(trits); err != nil {
		return "", err
	}
	checksum := h.MustSqueeze(consts.HashTrinarySize)[consts.HashTrinarySize-checksumTrinarySize:]
	return trinary.TritsToTrytes(append(trits, checksum...))
}
