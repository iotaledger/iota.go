package iotago

import "time"

type APIProvider interface {
	// APIForVersion returns the API for the given version.
	APIForVersion(Version) (API, error)

	// APIForTime returns the API for the given time.
	APIForTime(time.Time) API

	// APIForSlot returns the API for the given slot.
	APIForSlot(SlotIndex) API

	// APIForEpoch returns the API for the given epoch.
	APIForEpoch(EpochIndex) API

	// CommittedAPI returns the API for the last committed slot.
	CommittedAPI() API

	// LatestAPI returns the API for the latest supported protocol version.
	LatestAPI() API
}
