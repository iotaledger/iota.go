package iotago_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
)

func TestTimeProvider(t *testing.T) {
	genesisUnixTime := int64(1630000000) // Replace with an appropriate Unix timestamp
	genesisTime := time.Unix(genesisUnixTime, 0)
	slotDurationSeconds := int64(10)
	slotsPerEpochExponent := uint8(3) // 2^3 = 8 slots per epoch
	slotsPerEpoch := 1 << slotsPerEpochExponent

	tp := iotago.NewTimeProvider(genesisUnixTime, slotDurationSeconds, slotsPerEpochExponent)

	t.Run("Test Getters", func(t *testing.T) {
		require.EqualValues(t, genesisUnixTime, tp.GenesisUnixTime())
		require.EqualValues(t, genesisTime, tp.GenesisTime())
		require.EqualValues(t, slotDurationSeconds, tp.SlotDurationSeconds())
		require.EqualValues(t, slotsPerEpoch, tp.EpochDurationSlots())
		require.EqualValues(t, slotDurationSeconds*int64(slotsPerEpoch), tp.EpochDurationSeconds())
	})

	t.Run("Test SlotFromTime", func(t *testing.T) {
		slot0StartTime := genesisTime.Add(-time.Nanosecond)
		require.EqualValues(t, 0, tp.SlotFromTime(slot0StartTime))

		slot1StartTime := genesisTime
		require.EqualValues(t, 1, tp.SlotFromTime(slot1StartTime))

		slot2StartTime := genesisTime.Add(time.Duration(slotDurationSeconds) * time.Second)
		require.EqualValues(t, 2, tp.SlotFromTime(slot2StartTime))

		arbitraryTime := genesisTime.Add(time.Duration(slotDurationSeconds*3)*time.Second + 5*time.Second + 300*time.Millisecond)
		require.EqualValues(t, 4, tp.SlotFromTime(arbitraryTime))
	})

	t.Run("Test SlotStartTime", func(t *testing.T) {
		slot0StartTime := genesisTime.Add(-time.Nanosecond)
		require.EqualValues(t, slot0StartTime, tp.SlotStartTime(0))

		slot1StartTime := genesisTime
		require.EqualValues(t, slot1StartTime, tp.SlotStartTime(1))

		slot2StartTime := genesisTime.Add(time.Duration(slotDurationSeconds) * time.Second)
		require.EqualValues(t, slot2StartTime, tp.SlotStartTime(2))

		slot4000StartTime := genesisTime.Add(time.Duration(slotDurationSeconds*3999) * time.Second)
		require.EqualValues(t, slot4000StartTime, tp.SlotStartTime(4000))
	})

	t.Run("Test SlotEndTime", func(t *testing.T) {
		slot0EndTime := genesisTime.Add(-time.Nanosecond)
		require.EqualValues(t, slot0EndTime, tp.SlotEndTime(0))

		slot1EndTime := genesisTime.Add(time.Duration(slotDurationSeconds) * time.Second).Add(-time.Nanosecond)
		require.EqualValues(t, slot1EndTime, tp.SlotEndTime(1))

		slot2EndTime := genesisTime.Add(time.Duration(slotDurationSeconds*2) * time.Second).Add(-time.Nanosecond)
		require.EqualValues(t, slot2EndTime, tp.SlotEndTime(2))

		slot4000EndTime := genesisTime.Add(time.Duration(slotDurationSeconds*4000) * time.Second).Add(-time.Nanosecond)
		require.EqualValues(t, slot4000EndTime, tp.SlotEndTime(4000))
	})

	t.Run("Test EpochFromSlot", func(t *testing.T) {
		require.EqualValues(t, 0, tp.EpochFromSlot(0))
		require.EqualValues(t, 0, tp.EpochFromSlot(7))
		require.EqualValues(t, 1, tp.EpochFromSlot(8))
		require.EqualValues(t, 1, tp.EpochFromSlot(15))
		require.EqualValues(t, 4000, tp.EpochFromSlot(32000))
		require.EqualValues(t, 4000, tp.EpochFromSlot(32007))
	})

	t.Run("Test EpochStart", func(t *testing.T) {
		require.EqualValues(t, 0, tp.EpochStart(0))
		require.EqualValues(t, 8, tp.EpochStart(1))
		require.EqualValues(t, 16, tp.EpochStart(2))
		require.EqualValues(t, 32000, tp.EpochStart(4000))
	})

	t.Run("Test EpochEnd", func(t *testing.T) {
		require.EqualValues(t, 7, tp.EpochEnd(0))
		require.EqualValues(t, 15, tp.EpochEnd(1))
		require.EqualValues(t, 23, tp.EpochEnd(2))
		require.EqualValues(t, 32007, tp.EpochEnd(4000))
	})

	t.Run("Test SlotsBeforeNextEpoch", func(t *testing.T) {
		require.EqualValues(t, 8, tp.SlotsBeforeNextEpoch(0))
		require.EqualValues(t, 1, tp.SlotsBeforeNextEpoch(7))
		require.EqualValues(t, 1, tp.SlotsBeforeNextEpoch(15))
		require.EqualValues(t, 8, tp.SlotsBeforeNextEpoch(32000))
		require.EqualValues(t, 1, tp.SlotsBeforeNextEpoch(32007))
	})

	t.Run("Test SlotsSinceEpochStart", func(t *testing.T) {
		require.EqualValues(t, 0, tp.SlotsSinceEpochStart(0))
		require.EqualValues(t, 7, tp.SlotsSinceEpochStart(7))
		require.EqualValues(t, 0, tp.SlotsSinceEpochStart(8))
		require.EqualValues(t, 7, tp.SlotsSinceEpochStart(15))
		require.EqualValues(t, 0, tp.SlotsSinceEpochStart(32000))
		require.EqualValues(t, 7, tp.SlotsSinceEpochStart(32007))
	})
}
