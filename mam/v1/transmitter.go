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
	ErrUnknownChannelMode = errors.New("Channel mode must be ChannelModePublic, ChannelModePrivate or ChannelModeRestricted")
	ErrNoSideKey          = errors.New("A 81-trytes sideKey must be provided for the restricted mode")
)

// Transmitter defines the MAM facade transmitter.
type Transmitter struct {
	api     API
	channel *Channel
	seed    trinary.Trytes
	mwm     uint64
}

// NewTransmitter returns a new transmitter.
func NewTransmitter(api API, seed trinary.Trytes, mwm uint64, securityLevel consts.SecurityLevel) *Transmitter {
	return &Transmitter{
		api:     api,
		channel: NewChannel(securityLevel),
		seed:    seed,
		mwm:     mwm,
	}
}

// NewTransmitterWithChannel returns a new transmitter with an existing channel.
func NewTransmitterWithChannel(api API, seed trinary.Trytes, mwm uint64, channel *Channel) *Transmitter {
	return &Transmitter{
		api:     api,
		channel: channel,
		seed:    seed,
		mwm:     mwm,
	}
}

// SetMode sets the Channel mode.
func (t *Transmitter) SetMode(m ChannelMode, sideKey trinary.Trytes) error {
	switch m {
	case ChannelModePublic, ChannelModePrivate:
		t.channel.SideKey = consts.NullHashTrytes
	case ChannelModeRestricted:
		if l := len(sideKey); l != 81 {
			return errors.Wrapf(ErrNoSideKey, "sidekey of length %d", l)
		}
		t.channel.SideKey = sideKey
	default:
		return errors.Wrapf(ErrUnknownChannelMode, "Channel mode [%s]", m)
	}
	t.channel.Mode = m
	return nil
}

// Mode returns the Channel mode.
func (t *Transmitter) Mode() ChannelMode {
	return t.channel.Mode
}

// SideKey returns the Channel's side key.
func (t *Transmitter) SideKey() trinary.Trytes {
	return t.channel.SideKey
}

// SideKey returns the underlying Channel of this Transmitter.
func (t *Transmitter) Channel() *Channel {
	return t.channel
}

// Transmit creates a MAM message using the given string and transmits it. On success, it returns
// the addresses root.
func (t *Transmitter) Transmit(message string, params ...string) (trinary.Trytes, error) {
	root, address, payload, err := t.createMessage(message)
	if err != nil {
		return "", errors.Wrapf(err, "create message")
	}

	var tag = ""
	if len(params) > 0 {
		tag = params[0]
	}

	if err := t.attachMessage(address, payload, tag); err != nil {
		return "", errors.Wrapf(err, "attach message")
	}

	return root, nil
}

func (t *Transmitter) createMessage(message string) (trinary.Trytes, trinary.Trytes, trinary.Trytes, error) {
	treeSize := merkle.MerkleSize(t.channel.Count)
	messageTrytes, err := converter.ASCIIToTrytes(message)
	if err != nil {
		return "", "", "", err
	}

	payloadLength := PayloadMinLength(uint64(len(messageTrytes)*3), treeSize*uint64(consts.HashTrinarySize), t.channel.Index, t.channel.SecurityLevel)

	root, err := merkle.MerkleCreate(t.channel.Count, t.seed, t.channel.Start, t.channel.SecurityLevel, curl.NewCurlP27())
	if err != nil {
		return "", "", "", err
	}
	rootTrytes, err := trinary.TritsToTrytes(root)
	if err != nil {
		return "", "", "", err
	}

	nextRoot, err := merkle.MerkleCreate(t.channel.NextCount, t.seed, t.channel.nextStart(), t.channel.SecurityLevel, curl.NewCurlP27())
	if err != nil {
		return "", "", "", err
	}

	payload, payloadLength, err := MAMCreate(payloadLength, messageTrytes, t.channel.SideKey, root, treeSize*consts.HashTrinarySize,
		t.channel.Count, t.channel.Index, nextRoot, t.channel.Start, t.seed, t.channel.SecurityLevel)
	if err != nil {
		return "", "", "", err
	}
	payload = trinary.PadTrits(payload, len(payload)+(3-len(payload)%3))
	payloadTrytes, err := trinary.TritsToTrytes(payload)
	if err != nil {
		return "", "", "", err
	}

	t.channel.incIndex()
	t.channel.NextRoot = nextRoot

	address, err := makeAddress(t.channel.Mode, root, t.channel.SideKey)
	if err != nil {
		return "", "", "", err
	}

	return rootTrytes, address, payloadTrytes, nil
}

func (t *Transmitter) attachMessage(address, payload trinary.Trytes, tag string) error {
	if err := trinary.ValidTrytes(address); err != nil {
		return errors.Wrapf(err, "invalid address")
	}

	transfers := bundle.Transfers{bundle.Transfer{
		Address: address,
		Value:   0,
		Message: payload,
		Tag:     tag,
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
