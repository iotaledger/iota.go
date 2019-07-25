package mam

import (
	"errors"
	"strings"

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

// Transmitter defines the MAM facade state the is used in various functions.
type Transmitter struct {
	api           API
	subscriptions map[trinary.Trytes]*Subscription
	channel       *Channel
	seed          trinary.Trytes
}

// NewTransmitter returns a new state.
func NewTransmitter(api API, seed trinary.Trytes, securityLevel consts.SecurityLevel) Transmitter {
	return Transmitter{
		api:           api,
		subscriptions: make(map[trinary.Trytes]*Subscription),
		channel:       newChannel(securityLevel),
		seed:          seed,
	}
}

// Channel returns the channel of the state.
func (t *Transmitter) Channel() *Channel {
	return t.channel
}

// Subscribe subscribs the state the channel defines by the given `channelRoot`.
func (t *Transmitter) Subscribe(cr trinary.Trytes, cm ChannelMode, ck trinary.Trytes) {
	t.subscriptions[cr] = newSubscription(cr, cm, ck)
}

// SubscriptionCount returns the number of subscriptions.
func (t *Transmitter) SubscriptionCount() int {
	return len(t.subscriptions)
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
func (t *Transmitter) Transmit(message string) (bundle.Bundle, error) {
	payload, _, address, err := t.createMessage(message)
	if err != nil {
		return nil, err
	}

	bundle, err := t.attachMessage(payload, address)
	if err != nil {
		return nil, err
	}

	return bundle, nil
}

func (t *Transmitter) createMessage(message string) (trinary.Trytes, trinary.Trytes, trinary.Trytes, error) {
	nextStart := t.channel.Start + t.channel.Count

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

	nextRoot, err := merkle.MerkleCreate(t.channel.NextCount, t.seed, nextStart, t.channel.SecurityLevel, curl.NewCurlP27())
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

	if t.channel.Mode == ChannelModePublic {
		return payloadTrytes, rootTrytes, rootTrytes, nil
	}

	return "", "", "", err
}

func (t *Transmitter) attachMessage(payload, address trinary.Trytes) (bundle.Bundle, error) {
	transfers := bundle.Transfers{bundle.Transfer{
		Address: address,
		Value:   0,
		Message: payload,
		Tag:     "",
	}}

	trytes, err := t.api.PrepareTransfers(strings.Repeat("9", 81), transfers, api.PrepareTransfersOptions{})
	if err != nil {
		return nil, err
	}

	bundle, err := t.api.SendTrytes(trytes, 3, 9)
	if err != nil {
		return nil, err
	}

	return bundle, nil
}
