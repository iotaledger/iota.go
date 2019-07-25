package mam

import (
	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/trinary"
)

// Channel defines a MAM channel.
type Channel struct {
	Mode          ChannelMode
	SideKey       trinary.Trytes
	NextRoot      trinary.Trits
	SecurityLevel consts.SecurityLevel
	Start         uint64
	Count         uint64
	NextCount     uint64
	Index         uint64
}

func newChannel(securityLevel consts.SecurityLevel) *Channel {
	return &Channel{
		Mode:          ChannelModePublic,
		SideKey:       trinary.Trytes(""),
		SecurityLevel: securityLevel,
		Start:         0,
		Count:         1,
		NextCount:     1,
		Index:         0,
	}
}

func (c *Channel) incIndex() {
	if c.Index == c.Count-1 {
		c.Start += c.NextCount
		c.Index = 0
	} else {
		c.Index++
	}
}
