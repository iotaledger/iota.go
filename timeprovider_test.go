package iotago

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSlot(t *testing.T) {
	timeProvider := NewTimeProvider(time.Now().Unix(), 10, 10)
	genesisTime := timeProvider.GenesisTime()

	{
		endOfSlotTime := genesisTime.Add(time.Duration(timeProvider.SlotDuration()) * time.Second).Add(-1)

		require.Equal(t, SlotIndex(1), timeProvider.SlotIndexFromTime(endOfSlotTime))
		require.False(t, timeProvider.SlotEndTime(SlotIndex(1)).Before(endOfSlotTime))

		startOfSlotTime := genesisTime.Add(time.Duration(timeProvider.SlotDuration()) * time.Second)

		require.Equal(t, SlotIndex(2), timeProvider.SlotIndexFromTime(startOfSlotTime))
		require.False(t, timeProvider.SlotStartTime(SlotIndex(2)).After(startOfSlotTime))
	}

	{
		testTime := genesisTime.Add(5 * time.Second)
		index := timeProvider.SlotIndexFromTime(testTime)
		require.Equal(t, index, SlotIndex(1))

		startTime := timeProvider.SlotStartTime(index)
		require.Equal(t, startTime, time.Unix(genesisTime.Unix(), 0))
		endTime := timeProvider.SlotEndTime(index)
		require.Equal(t, endTime, timeProvider.SlotStartTime(index+1).Add(-1))
	}

	{
		testTime := genesisTime.Add(10 * time.Second)
		index := timeProvider.SlotIndexFromTime(testTime)
		require.Equal(t, index, SlotIndex(2))

		startTime := timeProvider.SlotStartTime(index)
		require.Equal(t, startTime, time.Unix(genesisTime.Add(10*time.Second).Unix(), 0))
		endTime := timeProvider.SlotEndTime(index)
		require.Equal(t, endTime, timeProvider.SlotStartTime(index+1).Add(-1))
	}

	{
		testTime := genesisTime.Add(35 * time.Second)
		index := timeProvider.SlotIndexFromTime(testTime)
		require.Equal(t, index, SlotIndex(4))

		startTime := timeProvider.SlotStartTime(index)
		require.Equal(t, startTime, time.Unix(genesisTime.Add(30*time.Second).Unix(), 0))
		endTime := timeProvider.SlotEndTime(index)
		require.Equal(t, endTime, timeProvider.SlotStartTime(index+1).Add(-1))
	}

	{
		testTime := genesisTime.Add(49 * time.Second)
		index := timeProvider.SlotIndexFromTime(testTime)
		require.Equal(t, index, SlotIndex(5))
	}

	{
		// a time before genesis time, index = 0
		testTime := genesisTime.Add(-10 * time.Second)
		index := timeProvider.SlotIndexFromTime(testTime)
		require.Equal(t, index, SlotIndex(0))
	}

	{
		endOfEpochTime := genesisTime.Add(time.Duration(timeProvider.EpochDuration()*timeProvider.SlotDuration()) * time.Second).Add(-1)
		preEndSlot := timeProvider.SlotIndexFromTime(endOfEpochTime) - 1
		require.Equal(t, EpochIndex(1), timeProvider.EpochsFromSlot(preEndSlot))

		endSlot := timeProvider.SlotIndexFromTime(endOfEpochTime)
		require.Equal(t, EpochIndex(2), timeProvider.EpochsFromSlot(endSlot))

		startSlot := SlotIndex(timeProvider.EpochDuration())
		require.Equal(t, EpochIndex(2), timeProvider.EpochsFromSlot(startSlot))

		nextEpochStart := startSlot + SlotIndex(timeProvider.EpochDuration())
		require.Equal(t, EpochIndex(3), timeProvider.EpochsFromSlot(nextEpochStart))
	}

	{
		require.Equal(t, SlotIndex(5), timeProvider.SlotsBeforeNextEpoch(15))
		require.Equal(t, SlotIndex(10), timeProvider.SlotsBeforeNextEpoch(20))
		require.Equal(t, SlotIndex(0), timeProvider.SlotsSinceEpochStart(20))
		require.Equal(t, SlotIndex(1), timeProvider.SlotsSinceEpochStart(21))
	}
}
