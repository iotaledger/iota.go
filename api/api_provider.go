package api

import iotago "github.com/iotaledger/iota.go/v4"

type Provider interface {
	// APIForVersion returns the API for the given version.
	APIForVersion(iotago.Version) (iotago.API, error)

	// APIForSlot returns the API for the given slot.
	APIForSlot(iotago.SlotIndex) iotago.API

	// APIForEpoch returns the API for the given epoch.
	APIForEpoch(iotago.EpochIndex) iotago.API

	// CurrentAPI returns the API for the current slot.
	CurrentAPI() iotago.API

	// LatestAPI returns the API for the latest supported protocol version.
	LatestAPI() iotago.API
}
