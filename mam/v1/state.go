package mam

import (
	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/trinary"
)

// State defines the MAM facade state the is used in various functions.
type State struct {
	subscriptions map[trinary.Trytes]*Subscription
	channel       *Channel
	seed          trinary.Trytes
}

// NewState returns a new state.
func NewState(settings Settings, seed trinary.Trytes, securityLevel consts.SecurityLevel) State {
	return State{
		subscriptions: make(map[trinary.Trytes]*Subscription),
		channel:       newChannel(securityLevel),
		seed:          seed,
	}
}

// Channel returns the channel of the state.
func (s *State) Channel() *Channel {
	return s.channel
}

// Subscribe subscribs the state the channel defines by the given `channelRoot`.
func (s *State) Subscribe(cr trinary.Trytes, cm ChannelMode, ck trinary.Trytes) {
	s.subscriptions[cr] = newSubscription(cr, cm, ck)
}

// SubscriptionCount returns the number of subscriptions.
func (s *State) SubscriptionCount() int {
	return len(s.subscriptions)
}

// SetMode sets the mode of the state.
func (s *State) SetMode(m ChannelMode, ck trinary.Trytes) error {
	if ck != "" {
		s.channel.SideKey = ck
	}
	s.channel.Mode = m
	return nil
}
