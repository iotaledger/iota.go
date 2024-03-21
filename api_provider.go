package iotago

import "time"

type APIProvider interface {
	// APIForVersion returns the API for the given version.
	APIForVersion(version Version) (API, error)

	// APIForTime returns the API for the given time.
	APIForTime(ts time.Time) API

	// APIForSlot returns the API for the given slot.
	APIForSlot(slot SlotIndex) API

	// APIForEpoch returns the API for the given epoch.
	APIForEpoch(epoch EpochIndex) API

	// CommittedAPI returns the API for the last committed slot.
	CommittedAPI() API

	// LatestAPI returns the API for the latest supported protocol version.
	LatestAPI() API
}
