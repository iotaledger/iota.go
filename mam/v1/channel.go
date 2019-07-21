package mam

import (
	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/trinary"
)

// Channel defines a MAM channel.
type Channel struct {
	Mode          ChannelMode
	SideKey       trinary.Trytes
	NextRoot      string
	SecurityLevel consts.SecurityLevel
	Start         int
	Count         int
	NextCount     int
	Index         int
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
