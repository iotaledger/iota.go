package api

import (
	iotago "github.com/iotaledger/iota.go/v4"
)

type (
	// HealthResponse defines the health response.
	HealthResponse struct {
		// Whether the node is healthy.
		IsHealthy bool `serix:""`
	}

	// RoutesResponse defines the response of a GET routes REST API call.
	RoutesResponse struct {
		Routes []iotago.PrefixedStringUint8 `serix:",lenPrefix=uint8"`
	}
)
