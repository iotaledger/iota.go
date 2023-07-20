package apimodels

import (
	iotago "github.com/iotaledger/iota.go/v4"
)

// TODO: use the API instance from Client instead.
var _internalAPI = iotago.V3API(iotago.NewV3ProtocolParameters())

type (
	httpOutput interface{ iotago.Output }
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func init() {
	api := _internalAPI.Underlying()
	must(api.RegisterInterfaceObjects((*httpOutput)(nil), (*iotago.BasicOutput)(nil)))
	must(api.RegisterInterfaceObjects((*httpOutput)(nil), (*iotago.AccountOutput)(nil)))
	must(api.RegisterInterfaceObjects((*httpOutput)(nil), (*iotago.FoundryOutput)(nil)))
	must(api.RegisterInterfaceObjects((*httpOutput)(nil), (*iotago.NFTOutput)(nil)))
}

type (
	// RoutesResponse defines the response of a GET routes REST API call.
	RoutesResponse struct {
		Routes []string `json:"routes"`
	}
)
