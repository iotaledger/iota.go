package iotago_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
)

func TestSlot(t *testing.T) {
	timeProvider := iotago.NewTimeProvider(time.Now().Unix(), 10, 3)
	genesisTime := timeProvider.GenesisTime()

	{
		endOfSlotTime := genesisTime.Add(time.Duration(timeProvider.SlotDurationSeconds()) * time.Second).Add(-1)

		require.Equal(t, iotago.SlotIndex(1), timeProvider.SlotFromTime(endOfSlotTime))
		require.False(t, timeProvider.SlotEndTime(iotago.SlotIndex(1)).Before(endOfSlotTime))

		startOfSlotTime := genesisTime.Add(time.Duration(timeProvider.SlotDurationSeconds()) * time.Second)

		require.Equal(t, iotago.SlotIndex(2), timeProvider.SlotFromTime(startOfSlotTime))
		require.False(t, timeProvider.SlotStartTime(iotago.SlotIndex(2)).After(startOfSlotTime))
	}

	{
		testTime := genesisTime.Add(5 * time.Second)
		index := timeProvider.SlotFromTime(testTime)
		require.Equal(t, index, iotago.SlotIndex(1))

		startTime := timeProvider.SlotStartTime(index)
		require.Equal(t, startTime, time.Unix(genesisTime.Unix(), 0))
		endTime := timeProvider.SlotEndTime(index)
		require.Equal(t, endTime, timeProvider.SlotStartTime(index+1).Add(-1))
	}

	{
		testTime := genesisTime.Add(10 * time.Second)
		index := timeProvider.SlotFromTime(testTime)
		require.Equal(t, index, iotago.SlotIndex(2))

		startTime := timeProvider.SlotStartTime(index)
		require.Equal(t, startTime, time.Unix(genesisTime.Add(10*time.Second).Unix(), 0))
		endTime := timeProvider.SlotEndTime(index)
		require.Equal(t, endTime, timeProvider.SlotStartTime(index+1).Add(-1))
	}

	{
		testTime := genesisTime.Add(35 * time.Second)
		index := timeProvider.SlotFromTime(testTime)
		require.Equal(t, index, iotago.SlotIndex(4))

		startTime := timeProvider.SlotStartTime(index)
		require.Equal(t, startTime, time.Unix(genesisTime.Add(30*time.Second).Unix(), 0))
		endTime := timeProvider.SlotEndTime(index)
		require.Equal(t, endTime, timeProvider.SlotStartTime(index+1).Add(-1))
	}

	{
		testTime := genesisTime.Add(49 * time.Second)
		index := timeProvider.SlotFromTime(testTime)
		require.Equal(t, index, iotago.SlotIndex(5))
	}

	{
		// a time before genesis time, index = 0
		testTime := genesisTime.Add(-10 * time.Second)
		index := timeProvider.SlotFromTime(testTime)
		require.Equal(t, index, iotago.SlotIndex(0))
	}

	{
		endOfEpochTime := genesisTime.Add(time.Duration(timeProvider.EpochDurationSeconds()) * time.Second).Add(-1)
		preEndSlot := timeProvider.SlotFromTime(endOfEpochTime) - 1
		require.Equal(t, iotago.EpochIndex(1), timeProvider.EpochFromSlot(preEndSlot))

		endSlot := timeProvider.SlotFromTime(endOfEpochTime)
		require.Equal(t, iotago.EpochIndex(2), timeProvider.EpochFromSlot(endSlot))

		startSlot := timeProvider.EpochDurationSlots()
		require.Equal(t, iotago.EpochIndex(2), timeProvider.EpochFromSlot(startSlot))

		nextEpochStart := startSlot + timeProvider.EpochDurationSlots()
		require.Equal(t, iotago.EpochIndex(3), timeProvider.EpochFromSlot(nextEpochStart))
	}

	{
		require.Equal(t, iotago.SlotIndex(8), timeProvider.SlotsBeforeNextEpoch(16))
		require.Equal(t, iotago.SlotIndex(4), timeProvider.SlotsBeforeNextEpoch(20))
		require.Equal(t, iotago.SlotIndex(0), timeProvider.SlotsSinceEpochStart(24))
		require.Equal(t, iotago.SlotIndex(5), timeProvider.SlotsSinceEpochStart(21))
	}
}
