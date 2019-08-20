package mam

import (
	"github.com/pkg/errors"

	"github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/bundle"
	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/converter"
	"github.com/iotaledger/iota.go/curl"
	"github.com/iotaledger/iota.go/merkle"
	"github.com/iotaledger/iota.go/trinary"
)

// Some error definitions.
var (
	ErrUnknownChannelMode = errors.New("channel mode must be ChannelModePublic, ChannelModePrivate or ChannelModeRestricted")
	ErrNoSideKey          = errors.New("A sideKey must be provided for the restricted mode")
)

// Transmitter defines the MAM facade transmitter.
type Transmitter struct {
	api     API
	channel *channel
	seed    trinary.Trytes
	mwm     uint64
}

// NewTransmitter returns a new transmitter.
func NewTransmitter(api API, seed trinary.Trytes, mwm uint64, securityLevel consts.SecurityLevel) *Transmitter {
	return &Transmitter{
		api:     api,
		channel: newChannel(securityLevel),
		seed:    seed,
		mwm:     mwm,
	}
}

// SetMode sets the channel mode.
func (t *Transmitter) SetMode(m ChannelMode, sideKey trinary.Trytes) error {
	if m != ChannelModePublic && m != ChannelModePrivate && m != ChannelModeRestricted {
		return ErrUnknownChannelMode
	}
	if m == ChannelModeRestricted {
		if sideKey == "" {
			return ErrNoSideKey
		}
		t.channel.sideKey = sideKey
	}
	t.channel.mode = m
	return nil
}

// Mode returns the channel mode.
func (t *Transmitter) Mode() ChannelMode {
	return t.channel.mode
}

// SideKey returns the channel's side key.
func (t *Transmitter) SideKey() trinary.Trytes {
	return t.channel.sideKey
}

// Transmit creates a MAM message using the given string and transmits it. On sucess, it returns
// the addresses root.
func (t *Transmitter) Transmit(message string) (trinary.Trytes, error) {
	root, address, payload, err := t.createMessage(message)
	if err != nil {
		return "", errors.Wrapf(err, "create message")
	}

	if err := t.attachMessage(address, payload); err != nil {
		return "", errors.Wrapf(err, "attach message")
	}

	return root, nil
}

func (t *Transmitter) createMessage(message string) (trinary.Trytes, trinary.Trytes, trinary.Trytes, error) {
	treeSize := merkle.MerkleSize(t.channel.count)
	messageTrytes, err := converter.ASCIIToTrytes(message)
	if err != nil {
		return "", "", "", err
	}

	payloadLength := PayloadMinLength(uint64(len(messageTrytes)*3), treeSize*uint64(consts.HashTrinarySize), t.channel.index, t.channel.securityLevel)

	root, err := merkle.MerkleCreate(t.channel.count, t.seed, t.channel.start, t.channel.securityLevel, curl.NewCurlP27())
	if err != nil {
		return "", "", "", err
	}
	rootTrytes, err := trinary.TritsToTrytes(root)
	if err != nil {
		return "", "", "", err
	}

	nextRoot, err := merkle.MerkleCreate(t.channel.nextCount, t.seed, t.channel.nextStart(), t.channel.securityLevel, curl.NewCurlP27())
	if err != nil {
		return "", "", "", err
	}

	payload, payloadLength, err := MAMCreate(payloadLength, messageTrytes, t.channel.sideKey, root, treeSize*consts.HashTrinarySize,
		t.channel.count, t.channel.index, nextRoot, t.channel.start, t.seed, t.channel.securityLevel)
	if err != nil {
		return "", "", "", err
	}
	payload = trinary.PadTrits(payload, len(payload)+(3-len(payload)%3))
	payloadTrytes, err := trinary.TritsToTrytes(payload)
	if err != nil {
		return "", "", "", err
	}

	t.channel.incIndex()
	t.channel.nextRoot = nextRoot

	address, err := makeAddress(t.channel.mode, root, t.channel.sideKey)
	if err != nil {
		return "", "", "", err
	}

	return rootTrytes, address, payloadTrytes, nil
}

func (t *Transmitter) attachMessage(address, payload trinary.Trytes) error {
	if err := trinary.ValidTrytes(address); err != nil {
		return errors.Wrapf(err, "invalid address")
	}

	transfers := bundle.Transfers{bundle.Transfer{
		Address: address,
		Value:   0,
		Message: payload,
		Tag:     "",
	}}

	trytes, err := t.api.PrepareTransfers(consts.NullHashTrytes, transfers, api.PrepareTransfersOptions{})
	if err != nil {
		return errors.Wrapf(err, "prepare transfers")
	}

	if _, err = t.api.SendTrytes(trytes, 3, t.mwm); err != nil {
		return errors.Wrapf(err, "send trytes")
	}

	return nil
}
