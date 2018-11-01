// Package api provides an API object for interacting with IRI nodes.
package api

import (
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/pow"
)

// A Provider is able to send API commands.
type Provider interface {
	// Send sends the given command and injects the result into the given out parameter.
	Send(cmd interface{}, out interface{}) error
	// SetSettings sets the settings for the provider.
	SetSettings(settings interface{}) error
}

// API defines an object encapsulating the communication to connected nodes and API calls.
type API struct {
	provider             Provider
	localProofOfWorkFunc pow.ProofOfWorkFunc
}

// CreateProviderFunc is a function which creates a new Provider given some settings.
type CreateProviderFunc func(settings interface{}) (Provider, error)

// Settings can supply different options for Provider creation.
type Settings interface {
	ProofOfWorkFunc() pow.ProofOfWorkFunc
}

// ComposeAPI composes a new API from the given settings and provider.
// If no provider function is supplied, then the default http provider is used.
// Settings must not be nil.
func ComposeAPI(settings Settings, createProvider ...CreateProviderFunc) (*API, error) {
	if settings == nil {
		return nil, ErrSettingsNil
	}
	var provider Provider
	var err error
	if len(createProvider) > 0 && createProvider[0] != nil {
		provider, err = createProvider[0](settings)
	} else {
		provider, err = NewHTTPClient(settings)
	}
	if err != nil {
		return nil, err
	}

	api := &API{provider: provider}
	if settings.ProofOfWorkFunc() != nil {
		api.localProofOfWorkFunc = settings.ProofOfWorkFunc()
	}

	return api, nil
}
