package mam

// ChannelMode is an enum of channel modes.
type ChannelMode int

// Definition of possible channel modes.
const (
	ChannelModePublic ChannelMode = iota
	ChannelModePrivate
	ChannelModeRestricted
)
