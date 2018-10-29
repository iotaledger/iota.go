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
	provider     Provider
	localPoWfunc pow.PowFunc
}

// A function which creates a new Provider.
type CreateProviderFunc func(settings interface{}) (Provider, error)

// Settings can supply different options for Provider creation.
type Settings interface {
	PowFunc() pow.PowFunc
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
		provider, err = NewHttpClient(settings)
	}
	if err != nil {
		return nil, err
	}

	api := &API{provider: provider}
	if settings.PowFunc() != nil {
		api.localPoWfunc = settings.PowFunc()
	}

	return api, nil
}
