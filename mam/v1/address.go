package mam

import (
	"github.com/iotaledger/iota.go/address"
	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/curl"
	"github.com/iotaledger/iota.go/trinary"
)

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

func toAddress(root trinary.Trits) (trinary.Trytes, error) {
	rootTrytes, err := trinary.TritsToTrytes(root)
	if err != nil {
		return "", err
	}

	chkSum, err := address.Checksum(rootTrytes)
	if err != nil {
		return "", err
	}

	return rootTrytes + chkSum, nil
}
