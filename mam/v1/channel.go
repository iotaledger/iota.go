package mam

import (
	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/trinary"
)

type channel struct {
	mode          ChannelMode
	sideKey       trinary.Trytes
	nextRoot      trinary.Trits
	securityLevel consts.SecurityLevel
	start         uint64
	count         uint64
	nextCount     uint64
	index         uint64
}

func newChannel(securityLevel consts.SecurityLevel) *channel {
	return &channel{
		mode:          ChannelModePublic,
		sideKey:       consts.NullHashTrytes,
		securityLevel: securityLevel,
		start:         0,
		count:         1,
		nextCount:     1,
		index:         0,
	}
}

func (c *channel) incIndex() {
	if c.index == c.count-1 {
		c.start += c.nextCount
		c.index = 0
	} else {
		c.index++
	}
}

func (c *channel) nextStart() uint64 {
	return c.start + c.count
}
