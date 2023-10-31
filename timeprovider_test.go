package iotago_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestTimeProvider(t *testing.T) {
	testTimeProviderWithGenesisSlot(t, 0)
	testTimeProviderWithGenesisSlot(t, 102848302)
	testTimeProviderWithGenesisSlot(t, tpkg.RandSlot())
}

func testTimeProviderWithGenesisSlot(t *testing.T, genesisSlot iotago.SlotIndex) {
	genesisUnixTime := int64(1630000000) // Replace with an appropriate Unix timestamp
	genesisTime := time.Unix(genesisUnixTime, 0)
	slotDurationSeconds := int64(10)
	slotsPerEpochExponent := uint8(3) // 2^3 = 8 slots per epoch
	slotsPerEpoch := 1 << slotsPerEpochExponent

	tp := iotago.NewTimeProvider(genesisSlot, genesisUnixTime, slotDurationSeconds, slotsPerEpochExponent)

	t.Run(fmt.Sprintf("Test Getters %d", genesisSlot), func(t *testing.T) {
		require.EqualValues(t, genesisUnixTime, tp.GenesisUnixTime())
		require.EqualValues(t, genesisTime, tp.GenesisTime())
		require.EqualValues(t, slotDurationSeconds, tp.SlotDurationSeconds())
		require.EqualValues(t, slotsPerEpoch, tp.EpochDurationSlots())
		require.EqualValues(t, slotDurationSeconds*int64(slotsPerEpoch), tp.EpochDurationSeconds())
	})

	if genesisSlot > 0 {
		t.Run(fmt.Sprintf("Test Below Genesis %d", genesisSlot), func(t *testing.T) {
			firstEpoch := iotago.EpochIndex(0)
			belowGenesisTime := genesisTime.Add(-time.Nanosecond)

			require.EqualValues(t, genesisSlot, tp.SlotFromTime(belowGenesisTime))

			require.EqualValues(t, belowGenesisTime, tp.SlotStartTime(genesisSlot-1))
			require.EqualValues(t, belowGenesisTime, tp.SlotStartTime(0))

			require.EqualValues(t, belowGenesisTime, tp.SlotEndTime(genesisSlot-1))
			require.EqualValues(t, belowGenesisTime, tp.SlotEndTime(0))

			require.EqualValues(t, firstEpoch, tp.EpochFromSlot(genesisSlot-1))
			require.EqualValues(t, firstEpoch, tp.EpochFromSlot(0))

			require.EqualValues(t, 0, tp.SlotsBeforeNextEpoch(genesisSlot-1))
			require.EqualValues(t, 0, tp.SlotsBeforeNextEpoch(0))

			require.EqualValues(t, 0, tp.SlotsSinceEpochStart(genesisSlot-1))
			require.EqualValues(t, 0, tp.SlotsSinceEpochStart(0))
		})
	}

	t.Run(fmt.Sprintf("Test SlotFromTime %d", genesisSlot), func(t *testing.T) {
		slot0StartTime := genesisTime.Add(-time.Nanosecond)
		require.EqualValues(t, genesisSlot, tp.SlotFromTime(slot0StartTime))

		slot1StartTime := genesisTime
		require.EqualValues(t, genesisSlot+1, tp.SlotFromTime(slot1StartTime))

		slot2StartTime := genesisTime.Add(time.Duration(slotDurationSeconds) * time.Second)
		require.EqualValues(t, genesisSlot+2, tp.SlotFromTime(slot2StartTime))

		arbitraryTime := genesisTime.Add(time.Duration(slotDurationSeconds*3)*time.Second + 5*time.Second + 300*time.Millisecond)
		require.EqualValues(t, genesisSlot+4, tp.SlotFromTime(arbitraryTime))
	})

	t.Run(fmt.Sprintf("Test SlotStartTime %d", genesisSlot), func(t *testing.T) {
		slot0StartTime := genesisTime.Add(-time.Nanosecond)
		require.EqualValues(t, slot0StartTime, tp.SlotStartTime(genesisSlot))

		slot1StartTime := genesisTime
		require.EqualValues(t, slot1StartTime, tp.SlotStartTime(genesisSlot+1))

		slot2StartTime := genesisTime.Add(time.Duration(slotDurationSeconds) * time.Second)
		require.EqualValues(t, slot2StartTime, tp.SlotStartTime(genesisSlot+2))

		slot4000StartTime := genesisTime.Add(time.Duration(slotDurationSeconds*3999) * time.Second)
		require.EqualValues(t, slot4000StartTime, tp.SlotStartTime(genesisSlot+4000))
	})

	t.Run(fmt.Sprintf("Test SlotEndTime %d", genesisSlot), func(t *testing.T) {
		slot0EndTime := genesisTime.Add(-time.Nanosecond)
		require.EqualValues(t, slot0EndTime, tp.SlotEndTime(genesisSlot))

		slot1EndTime := genesisTime.Add(time.Duration(slotDurationSeconds) * time.Second).Add(-time.Nanosecond)
		require.EqualValues(t, slot1EndTime, tp.SlotEndTime(genesisSlot+1))

		slot2EndTime := genesisTime.Add(time.Duration(slotDurationSeconds*2) * time.Second).Add(-time.Nanosecond)
		require.EqualValues(t, slot2EndTime, tp.SlotEndTime(genesisSlot+2))

		slot4000EndTime := genesisTime.Add(time.Duration(slotDurationSeconds*4000) * time.Second).Add(-time.Nanosecond)
		require.EqualValues(t, slot4000EndTime, tp.SlotEndTime(genesisSlot+4000))
	})

	t.Run(fmt.Sprintf("Test EpochFromSlot %d", genesisSlot), func(t *testing.T) {
		require.EqualValues(t, 0, tp.EpochFromSlot(genesisSlot))
		require.EqualValues(t, 0, tp.EpochFromSlot(genesisSlot+7))
		require.EqualValues(t, 1, tp.EpochFromSlot(genesisSlot+8))
		require.EqualValues(t, 1, tp.EpochFromSlot(genesisSlot+15))
		require.EqualValues(t, 4000, tp.EpochFromSlot(genesisSlot+32000))
		require.EqualValues(t, 4000, tp.EpochFromSlot(genesisSlot+32007))
	})

	t.Run(fmt.Sprintf("Test EpochStart %d", genesisSlot), func(t *testing.T) {
		require.EqualValues(t, genesisSlot, tp.EpochStart(0))
		require.EqualValues(t, genesisSlot+8, tp.EpochStart(1))
		require.EqualValues(t, genesisSlot+16, tp.EpochStart(2))
		require.EqualValues(t, genesisSlot+32000, tp.EpochStart(4000))
	})

	t.Run(fmt.Sprintf("Test EpochEnd %d", genesisSlot), func(t *testing.T) {
		require.EqualValues(t, genesisSlot+7, tp.EpochEnd(0))
		require.EqualValues(t, genesisSlot+15, tp.EpochEnd(1))
		require.EqualValues(t, genesisSlot+23, tp.EpochEnd(2))
		require.EqualValues(t, genesisSlot+32007, tp.EpochEnd(4000))
	})

	t.Run(fmt.Sprintf("Test SlotsBeforeNextEpoch %d", genesisSlot), func(t *testing.T) {
		require.EqualValues(t, genesisSlot+8, tp.SlotsBeforeNextEpoch(genesisSlot))
		require.EqualValues(t, genesisSlot+1, tp.SlotsBeforeNextEpoch(genesisSlot+7))
		require.EqualValues(t, genesisSlot+1, tp.SlotsBeforeNextEpoch(genesisSlot+15))
		require.EqualValues(t, genesisSlot+8, tp.SlotsBeforeNextEpoch(genesisSlot+32000))
		require.EqualValues(t, genesisSlot+1, tp.SlotsBeforeNextEpoch(genesisSlot+32007))
	})

	t.Run(fmt.Sprintf("Test SlotsSinceEpochStart %d", genesisSlot), func(t *testing.T) {
		require.EqualValues(t, genesisSlot, tp.SlotsSinceEpochStart(genesisSlot))
		require.EqualValues(t, genesisSlot+7, tp.SlotsSinceEpochStart(genesisSlot+7))
		require.EqualValues(t, genesisSlot, tp.SlotsSinceEpochStart(genesisSlot+8))
		require.EqualValues(t, genesisSlot+7, tp.SlotsSinceEpochStart(genesisSlot+15))
		require.EqualValues(t, genesisSlot+0, tp.SlotsSinceEpochStart(genesisSlot+32000))
		require.EqualValues(t, genesisSlot+7, tp.SlotsSinceEpochStart(genesisSlot+32007))
	})
}
