package mam

import (
	"fmt"
	"strings"
)

// ChannelMode is an enum of channel modes.
type ChannelMode int

// Definition of possible channel modes.
const (
	ChannelModePublic ChannelMode = iota
	ChannelModePrivate
	ChannelModeRestricted
)

// ParseChannelMode parses a channel mode from the given string.
func ParseChannelMode(input string) (ChannelMode, error) {
	switch cm := strings.TrimSpace(strings.ToLower(input)); cm {
	case "public":
		return ChannelModePublic, nil
	case "private":
		return ChannelModePrivate, nil
	case "restricted":
		return ChannelModeRestricted, nil
	default:
		return ChannelModePublic, fmt.Errorf("channel mode %q is unknown", cm)
	}
}

func (cm ChannelMode) String() string {
	switch cm {
	case ChannelModePublic:
		return "public"
	case ChannelModePrivate:
		return "private"
	case ChannelModeRestricted:
		return "restricted"
	default:
		return "unknown"
	}
}
