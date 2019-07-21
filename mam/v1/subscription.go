package mam

import (
	"time"

	"github.com/iotaledger/iota.go/trinary"
)

// Subscription defines a channel subscription.
type Subscription struct {
	Active          bool
	Timeout         time.Duration
	ChannelMode     ChannelMode
	ChannelKey      trinary.Trytes
	ChannelRoot     trinary.Trytes
	NextChannelRoot trinary.Trytes
}

func newSubscription(cr trinary.Trytes, cm ChannelMode, ck trinary.Trytes) *Subscription {
	return &Subscription{
		Active:      true,
		Timeout:     time.Second * 5,
		ChannelMode: cm,
		ChannelKey:  ck,
		ChannelRoot: cr,
	}
}
