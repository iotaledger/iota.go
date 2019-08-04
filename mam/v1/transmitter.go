package mam

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/iotaledger/iota.go/address"
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

// Transmitter defines the MAM facade state that is used in various functions to transmit MAM-Messages.
type Transmitter struct {
	api     API
	channel *Channel
	seed    trinary.Trytes
}

// NewTransmitter returns a new state.
func NewTransmitter(api API, seed trinary.Trytes, securityLevel consts.SecurityLevel) *Transmitter {
	return &Transmitter{
		api:     api,
		channel: newChannel(securityLevel),
		seed:    seed,
	}
}

// Channel returns the channel of the state.
func (t *Transmitter) Channel() *Channel {
	return t.channel
}

// SetMode sets the mode of the state.
func (t *Transmitter) SetMode(m ChannelMode, ck trinary.Trytes) error {
	if m != ChannelModePublic && m != ChannelModePrivate && m != ChannelModeRestricted {
		return ErrUnknownChannelMode
	}
	if m == ChannelModeRestricted && ck == "" {
		return ErrNoSideKey
	}
	if ck != "" {
		t.channel.SideKey = ck
	}
	t.channel.Mode = m
	return nil
}

// Transmit creates a MAM message using the given string and transmits it.
func (t *Transmitter) Transmit(message string, mwm uint64) (trinary.Trytes, error) {
	address, payload, err := t.createMessage(message)
	if err != nil {
		return "", errors.Wrapf(err, "create message")
	}

	if err := t.attachMessage(address, payload, mwm); err != nil {
		return "", errors.Wrapf(err, "attach message")
	}

	return address, nil
}

func (t *Transmitter) createMessage(message string) (trinary.Trytes, trinary.Trytes, error) {
	treeSize := merkle.MerkleSize(t.channel.Count)
	messageTrytes, err := converter.ASCIIToTrytes(message)
	if err != nil {
		return "", "", err
	}

	payloadLength := PayloadMinLength(uint64(len(messageTrytes)*3), treeSize*uint64(consts.HashTrinarySize), t.channel.Index, t.channel.SecurityLevel)

	root, err := merkle.MerkleCreate(t.channel.Count, t.seed, t.channel.Start, t.channel.SecurityLevel, curl.NewCurlP27())
	if err != nil {
		return "", "", err
	}
	rootTrytes, err := trinary.TritsToTrytes(root)
	if err != nil {
		return "", "", err
	}

	nextRoot, err := merkle.MerkleCreate(t.channel.NextCount, t.seed, t.channel.Start+t.channel.Count, t.channel.SecurityLevel, curl.NewCurlP27())
	if err != nil {
		return "", "", err
	}

	sideKey := t.channel.SideKey
	if sideKey == "" {
		sideKey = "999999999999999999999999999999999999999999999999999999999999999999999999999999999"
	}
	payload, payloadLength, err := MAMCreate(payloadLength, messageTrytes, t.channel.SideKey, root, treeSize*consts.HashTrinarySize,
		t.channel.Count, t.channel.Index, nextRoot, t.channel.Start, t.seed, t.channel.SecurityLevel)
	if err != nil {
		return "", "", err
	}
	payload = trinary.PadTrits(payload, len(payload)+(3-len(payload)%3))
	payloadTrytes, err := trinary.TritsToTrytes(payload)
	if err != nil {
		return "", "", err
	}

	t.channel.incIndex()
	t.channel.NextRoot = nextRoot

	if t.channel.Mode == ChannelModePublic {
		chkSum, err := address.Checksum(rootTrytes)
		if err != nil {
			return "", "", err
		}
		return rootTrytes + chkSum, payloadTrytes, nil
	}

	return "", "", err
}

func (t *Transmitter) attachMessage(address, payload trinary.Trytes, mwm uint64) error {
	if err := trinary.ValidTrytes(address); err != nil {
		return errors.Wrapf(err, "invalid address")
	}

	transfers := bundle.Transfers{bundle.Transfer{
		Address: address,
		Value:   0,
		Message: payload,
		Tag:     "",
	}}

	trytes, err := t.api.PrepareTransfers(strings.Repeat("9", 81), transfers, api.PrepareTransfersOptions{})
	if err != nil {
		return errors.Wrapf(err, "prepare transfers")
	}

	if _, err = t.api.SendTrytes(trytes, 3, mwm); err != nil {
		return errors.Wrapf(err, "send trytes")
	}

	return nil
}
