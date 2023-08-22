package iotago

import (
	"time"
)

// TimeProvider defines the genesis time of slot 0 and allows to convert index to and from time.
type TimeProvider struct {
	// genesisUnixTime is the time (Unix in seconds) of the genesis.
	genesisUnixTime int64

	// slotDurationSeconds is the slot duration in seconds.
	slotDurationSeconds int64

	// slotsPerEpochExponent is the number of slots in an epoch expressed as an exponent of 2.
	// (2**SlotsPerEpochExponent) == slots in an epoch.
	slotsPerEpochExponent uint8

	// epochDurationSeconds is the epoch duration in seconds.
	epochDurationSeconds int64

	// epochDurationSlots is the epoch duration in slots.
	epochDurationSlots SlotIndex
}

// NewTimeProvider creates a new time provider.
func NewTimeProvider(genesisUnixTime int64, slotDurationSeconds int64, slotsPerEpochExponent uint8) *TimeProvider {
	// if slotDurationSeconds == 0 {
	//	panic("slot duration can't be zero")
	// }

	return &TimeProvider{
		genesisUnixTime:       genesisUnixTime,
		slotDurationSeconds:   slotDurationSeconds,
		slotsPerEpochExponent: slotsPerEpochExponent,
		epochDurationSeconds:  (1 << slotsPerEpochExponent) * slotDurationSeconds,
		epochDurationSlots:    1 << slotsPerEpochExponent,
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
func (t *TimeProvider) SlotDurationSeconds() int64 {
	return t.slotDurationSeconds
}

func (t *TimeProvider) EpochDurationSlots() SlotIndex {
	return t.epochDurationSlots
}

func (t *TimeProvider) EpochDurationSeconds() int64 {
	return t.epochDurationSeconds
}

// SlotFromTime calculates the SlotIndex from the given time.
//
// Note: slots are counted starting from 1 because 0 is reserved for the genesis which has to be addressable as its own
// slot as part of the commitment chains.
func (t *TimeProvider) SlotFromTime(targetTime time.Time) SlotIndex {
	elapsedSeconds := targetTime.Unix() - t.genesisUnixTime
	if elapsedSeconds < 0 {
		return 0
	}

	return SlotIndex((elapsedSeconds / t.slotDurationSeconds) + 1)
}

// SlotStartTime calculates the start time of the given slot.
func (t *TimeProvider) SlotStartTime(i SlotIndex) time.Time {
	if i == 0 {
		return time.Unix(t.genesisUnixTime, 0)
	}

	startUnix := t.genesisUnixTime + int64(i-1)*t.slotDurationSeconds

	return time.Unix(startUnix, 0)
}

// SlotEndTime returns the latest possible timestamp for a slot. Anything with higher timestamp will belong to the next slot.
func (t *TimeProvider) SlotEndTime(i SlotIndex) time.Time {
	if i == 0 {
		return time.Unix(t.genesisUnixTime, 0)
	}

	endUnix := t.genesisUnixTime + int64(i)*t.slotDurationSeconds
	// we subtract 1 nanosecond from the next slot to get the latest possible timestamp for slot i
	return time.Unix(endUnix, 0).Add(-1)
}

// EpochFromSlot calculates the EpochIndex from the given slot.
func (t *TimeProvider) EpochFromSlot(slot SlotIndex) EpochIndex {
	if slot == 0 {
		return 0
	}

	return EpochIndex(slot>>SlotIndex(t.slotsPerEpochExponent) + 1)
}

// EpochStart calculates the start slot of the given epoch.
func (t *TimeProvider) EpochStart(epoch EpochIndex) SlotIndex {
	if epoch == 0 {
		return 0
	} else if epoch == 1 {
		return 1
	}

	return SlotIndex((epoch - 1) << t.slotsPerEpochExponent)
}

// EpochEnd calculates the end included slot of the given epoch.
func (t *TimeProvider) EpochEnd(epoch EpochIndex) SlotIndex {
	if epoch == 0 {
		return 0
	}

	return SlotIndex(epoch<<EpochIndex(t.slotsPerEpochExponent) - 1)
}

// SlotsBeforeNextEpoch calculates the slots before the start of the next epoch.
func (t *TimeProvider) SlotsBeforeNextEpoch(slot SlotIndex) SlotIndex {
	return t.EpochStart(t.EpochFromSlot(slot)+1) - slot
}

// SlotsSinceEpochStart calculates the slots since the start of the epoch.
func (t *TimeProvider) SlotsSinceEpochStart(slot SlotIndex) SlotIndex {
	return slot - t.EpochStart(t.EpochFromSlot(slot))
}
