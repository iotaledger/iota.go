package mam

import (
	"strings"

	"github.com/pkg/errors"
)

// Error definitions
var (
	ErrCouldNotParseChannelMode = errors.New("could not parse channel mode")
)

// ChannelMode is an enum of channel modes.
type ChannelMode string

// Definition of possible channel modes.
const (
	ChannelModePublic     ChannelMode = "public"
	ChannelModePrivate    ChannelMode = "private"
	ChannelModeRestricted ChannelMode = "restricted"
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
		return "", errors.Wrapf(ErrCouldNotParseChannelMode, "input %q", cm)
	}
}
