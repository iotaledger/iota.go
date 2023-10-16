package iotago

type APIProvider interface {
	// APIForVersion returns the API for the given version.
	APIForVersion(Version) (API, error)

	// APIForSlot returns the API for the given slot.
	APIForSlot(SlotIndex) API

	// APIForEpoch returns the API for the given epoch.
	APIForEpoch(EpochIndex) API

	// CurrentAPI returns the API for the last committed slot.
	CurrentAPI() API

	// LatestAPI returns the API for the latest supported protocol version.
	LatestAPI() API
}
