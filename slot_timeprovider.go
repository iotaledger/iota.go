package iotago

import (
	"time"
)

// SlotTimeProvider defines the genesis time of slot 0 and allows to convert index to and from time.
type SlotTimeProvider struct {
	// genesisUnixTime is the time (Unix in seconds) of the genesis.
	genesisUnixTime int64

	// duration is the default slot duration in seconds.
	duration int64
}

// NewSlotTimeProvider creates a new time provider.
func NewSlotTimeProvider(genesisUnixTime int64, slotDuration int64) *SlotTimeProvider {
	return &SlotTimeProvider{
		genesisUnixTime: genesisUnixTime,
		duration:        slotDuration,
	}
}

// GenesisUnixTime is the time (Unix in seconds) of the genesis.
func (t *SlotTimeProvider) GenesisUnixTime() int64 {
	return t.genesisUnixTime
}

// GenesisTime is the time  of the genesis.
func (t *SlotTimeProvider) GenesisTime() time.Time {
	return time.Unix(t.genesisUnixTime, 0)
}

// Duration is the slot duration in seconds.
func (t *SlotTimeProvider) Duration() int64 {
	return t.duration
}

// IndexFromTime calculates the SlotIndex from the given time.
//
// Note: slots are counted starting from 1 because 0 is reserved for the genesis which has to be addressable as its own
// slot as part of the commitment chains.
func (t *SlotTimeProvider) IndexFromTime(time time.Time) SlotIndex {
	elapsedSeconds := time.Unix() - t.genesisUnixTime
	if elapsedSeconds < 0 {
		return 0
	}

	return SlotIndex(elapsedSeconds/t.duration + 1)
}

// StartTime calculates the start time of the given slot.
func (t *SlotTimeProvider) StartTime(i SlotIndex) time.Time {
	startUnix := t.genesisUnixTime + int64(i-1)*t.duration
	return time.Unix(startUnix, 0)
}

// EndTime returns the latest possible timestamp for a slot. Anything with higher timestamp will belong to the next slot.
func (t *SlotTimeProvider) EndTime(i SlotIndex) time.Time {
	endUnix := t.genesisUnixTime + int64(i)*t.duration
	// we subtract 1 nanosecond from the next slot to get the latest possible timestamp for slot i
	return time.Unix(endUnix, 0).Add(-1)
}
