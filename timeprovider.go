package iotago

import (
	"time"
)

// TimeProvider defines the genesis time of slot 0 and allows to convert index to and from time.
type TimeProvider struct {
	// genesisUnixTime is the time (Unix in seconds) of the genesis.
	genesisUnixTime int64

	// duration is the default slot duration in seconds.
	slotDuration int64
	// epochDuration is the default epoch duration in seconds.
	epochDuration int64
}

// NewTimeProvider creates a new time provider.
func NewTimeProvider(genesisUnixTime, slotDuration, epochDuration int64) *TimeProvider {
	return &TimeProvider{
		genesisUnixTime: genesisUnixTime,
		slotDuration:    slotDuration,
		epochDuration:   epochDuration,
	}
}

// GenesisUnixTime is the time (Unix in seconds) of the genesis.
func (t *TimeProvider) GenesisUnixTime() int64 {
	return t.genesisUnixTime
}

// GenesisTime is the time  of the genesis.
func (t *TimeProvider) GenesisTime() time.Time {
	return time.Unix(t.genesisUnixTime, 0)
}

// SlotDuration is the slot duration in seconds.
func (t *TimeProvider) SlotDuration() int64 {
	return t.slotDuration
}

func (t *TimeProvider) EpochDuration() int64 {
	return t.epochDuration
}

// SlotIndexFromTime calculates the SlotIndex from the given time.
//
// Note: slots are counted starting from 1 because 0 is reserved for the genesis which has to be addressable as its own
// slot as part of the commitment chains.
func (t *TimeProvider) SlotIndexFromTime(time time.Time) SlotIndex {
	elapsedSeconds := time.Unix() - t.genesisUnixTime
	if elapsedSeconds < 0 {
		return 0
	}

	return SlotIndex(elapsedSeconds/t.slotDuration + 1)
}

// SlotStartTime calculates the start time of the given slot.
func (t *TimeProvider) SlotStartTime(i SlotIndex) time.Time {
	if i == 0 {
		return time.Unix(t.genesisUnixTime, 0)
	}

	startUnix := t.genesisUnixTime + int64(i-1)*t.slotDuration
	return time.Unix(startUnix, 0)
}

// SlotEndTime returns the latest possible timestamp for a slot. Anything with higher timestamp will belong to the next slot.
func (t *TimeProvider) SlotEndTime(i SlotIndex) time.Time {
	if i == 0 {
		return time.Unix(t.genesisUnixTime, 0)
	}

	endUnix := t.genesisUnixTime + int64(i)*t.slotDuration
	// we subtract 1 nanosecond from the next slot to get the latest possible timestamp for slot i
	return time.Unix(endUnix, 0).Add(-1)
}

// EpochsFromSlot calculates the EpochIndex from the given slot.
func (t *TimeProvider) EpochsFromSlot(slot SlotIndex) EpochIndex {
	return EpochIndex(slot/SlotIndex(t.epochDuration)) + 1
}

// EpochStart calculates the start slot of the given epoch.
func (t *TimeProvider) EpochStart(epoch EpochIndex) SlotIndex {
	return SlotIndex((epoch-1) * EpochIndex(t.epochDuration))
}

// EpochEnd calculates the end included slot of the given epoch.
func (t *TimeProvider) EpochEnd(epoch EpochIndex) SlotIndex {
	return SlotIndex(epoch * EpochIndex(t.epochDuration) - 1)
}
