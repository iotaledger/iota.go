package mam

import (
	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/trinary"
)

type Channel struct {
	Mode          ChannelMode          `json:"mode"`
	SideKey       trinary.Trytes       `json:"side_key"`
	NextRoot      trinary.Trits        `json:"next_root"`
	SecurityLevel consts.SecurityLevel `json:"security_level"`
	Start         uint64               `json:"start"`
	Count         uint64               `json:"count"`
	NextCount     uint64               `json:"next_count"`
	Index         uint64               `json:"index"`
}

func NewChannel(securityLevel consts.SecurityLevel) *Channel {
	return &Channel{
		Mode:          ChannelModePublic,
		SideKey:       consts.NullHashTrytes,
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

func (c *Channel) nextStart() uint64 {
	return c.Start + c.Count
}
