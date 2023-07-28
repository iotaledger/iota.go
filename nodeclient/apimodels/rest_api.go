package apimodels

import (
	iotago "github.com/iotaledger/iota.go/v4"
)

// TODO: use the API instance from Client instead.
//
//nolint:nosnakecase
var _internalAPI = iotago.V3API(iotago.NewV3ProtocolParameters())

type (
	// RoutesResponse defines the response of a GET routes REST API call.
	RoutesResponse struct {
		Routes []string `json:"routes"`
	}
)
