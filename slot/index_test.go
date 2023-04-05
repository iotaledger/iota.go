package slot

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSlot(t *testing.T) {
	timeProvider := NewTimeProvider(time.Now().Unix(), 10)
	genesisTime := timeProvider.GenesisTime()

	{
		endOfSlotTime := genesisTime.Add(time.Duration(timeProvider.Duration()) * time.Second).Add(-1)

		require.Equal(t, Index(1), timeProvider.IndexFromTime(endOfSlotTime))
		require.False(t, timeProvider.EndTime(Index(1)).Before(endOfSlotTime))

		startOfSlotTime := genesisTime.Add(time.Duration(timeProvider.Duration()) * time.Second)

		require.Equal(t, Index(2), timeProvider.IndexFromTime(startOfSlotTime))
		require.False(t, timeProvider.StartTime(Index(2)).After(startOfSlotTime))
	}

	{
		testTime := genesisTime.Add(5 * time.Second)
		index := timeProvider.IndexFromTime(testTime)
		require.Equal(t, index, Index(1))

		startTime := timeProvider.StartTime(index)
		require.Equal(t, startTime, time.Unix(genesisTime.Unix(), 0))
		endTime := timeProvider.EndTime(index)
		require.Equal(t, endTime, timeProvider.StartTime(index+1).Add(-1))
	}

	{
		testTime := genesisTime.Add(10 * time.Second)
		index := timeProvider.IndexFromTime(testTime)
		require.Equal(t, index, Index(2))

		startTime := timeProvider.StartTime(index)
		require.Equal(t, startTime, time.Unix(genesisTime.Add(10*time.Second).Unix(), 0))
		endTime := timeProvider.EndTime(index)
		require.Equal(t, endTime, timeProvider.StartTime(index+1).Add(-1))
	}

	{
		testTime := genesisTime.Add(35 * time.Second)
		index := timeProvider.IndexFromTime(testTime)
		require.Equal(t, index, Index(4))

		startTime := timeProvider.StartTime(index)
		require.Equal(t, startTime, time.Unix(genesisTime.Add(30*time.Second).Unix(), 0))
		endTime := timeProvider.EndTime(index)
		require.Equal(t, endTime, timeProvider.StartTime(index+1).Add(-1))
	}

	{
		testTime := genesisTime.Add(49 * time.Second)
		index := timeProvider.IndexFromTime(testTime)
		require.Equal(t, index, Index(5))
	}

	{
		// a time before genesis time, index = 0
		testTime := genesisTime.Add(-10 * time.Second)
		index := timeProvider.IndexFromTime(testTime)
		require.Equal(t, index, Index(0))
	}
}
