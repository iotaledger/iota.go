package mam

import (
	"strings"

	"github.com/pkg/errors"
)

// Error definitions
var (
	ErrCouldNotParseChannelMode = errors.New("could not parse Channel mode")
)

// ChannelMode is an enum of Channel modes.
type ChannelMode string

// Definition of possible Channel modes.
const (
	ChannelModePublic     ChannelMode = "public"
	ChannelModePrivate    ChannelMode = "private"
	ChannelModeRestricted ChannelMode = "restricted"
)

// ParseChannelMode parses a Channel mode from the given string.
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
