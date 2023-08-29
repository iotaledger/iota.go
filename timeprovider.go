package iotago

import (
	"time"
)

// TimeProvider defines the perception of time, slots and epochs.
// It allows to convert slots to and from time, and epochs to and from slots.
// Slots are counted starting from 1 with 0 being reserved for times before the genesis, which has to be addressable as its own slot.
// Epochs are counted starting from 0.
//
// Example: with slotDurationSeconds = 10 and slotsPerEpochExponent = 3
// [] inclusive range boundary, () exclusive range boundary
// slot 0: [-inf; genesis)
// slot 1: [genesis; genesis+10)
// slot 2: [genesis+10; genesis+20)
// ...
// epoch 0: [slot 0; slot 8)
// epoch 1: [slot 8; slot 16)
// epoch 2: [slot 16; slot 24)
// ...
type TimeProvider struct {
	// genesisUnixTime is the time (Unix in seconds) of the genesis.
	genesisUnixTime int64

	genesisTime time.Time

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
	return &TimeProvider{
		genesisUnixTime:       genesisUnixTime,
		genesisTime:           time.Unix(genesisUnixTime, 0),
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
	return t.genesisTime
}

// SlotDurationSeconds is the slot duration in seconds.
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
// Note: The + 1 is required because slots are counted starting from 1 with 0 being reserved for times before the genesis,
// which has to be addressable as its own slot.
func (t *TimeProvider) SlotFromTime(targetTime time.Time) SlotIndex {
	elapsed := targetTime.Sub(t.genesisTime)
	if elapsed < 0 {
		return 0
	}

	return SlotIndex(int64(elapsed/time.Second)/t.slotDurationSeconds) + 1
}

// SlotStartTime calculates the start time of the given slot.
func (t *TimeProvider) SlotStartTime(slot SlotIndex) time.Time {
	if slot == 0 {
		return t.genesisTime.Add(-time.Nanosecond)
	}

	startUnix := t.genesisUnixTime + int64(slot-1)*t.slotDurationSeconds

	return time.Unix(startUnix, 0)
}

// SlotEndTime returns the latest possible timestamp for a slot. Anything with higher timestamp will belong to the next slot.
func (t *TimeProvider) SlotEndTime(slot SlotIndex) time.Time {
	if slot == 0 {
		return t.genesisTime.Add(-time.Nanosecond)
	}

	endUnix := t.genesisUnixTime + int64(slot)*t.slotDurationSeconds
	// we subtract 1 nanosecond from the next slot to get the latest possible timestamp for slot i
	return time.Unix(endUnix, 0).Add(-time.Nanosecond)
}

// EpochFromSlot calculates the EpochIndex from the given slot.
func (t *TimeProvider) EpochFromSlot(slot SlotIndex) EpochIndex {
	return EpochIndex(slot >> t.slotsPerEpochExponent)
}

// EpochStart calculates the start slot of the given epoch.
func (t *TimeProvider) EpochStart(epoch EpochIndex) SlotIndex {
	return SlotIndex(epoch << t.slotsPerEpochExponent)
}

// EpochEnd calculates the end included slot of the given epoch.
func (t *TimeProvider) EpochEnd(epoch EpochIndex) SlotIndex {
	return SlotIndex((epoch+1)<<t.slotsPerEpochExponent - 1)
}

// SlotsBeforeNextEpoch calculates the slots before the start of the next epoch.
func (t *TimeProvider) SlotsBeforeNextEpoch(slot SlotIndex) SlotIndex {
	return t.EpochStart(t.EpochFromSlot(slot)+1) - slot
}

// SlotsSinceEpochStart calculates the slots since the start of the epoch.
func (t *TimeProvider) SlotsSinceEpochStart(slot SlotIndex) SlotIndex {
	return slot - t.EpochStart(t.EpochFromSlot(slot))
}
