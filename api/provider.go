package api

import (
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/pow"
)

type Provider interface {
	Send(cmd interface{}, out interface{}) error
	SetSettings(settings interface{}) error
}

type API struct {
	provider     Provider
	localPoWfunc pow.PowFunc
}

type CreateProviderFunc func(settings interface{}) (Provider, error)

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
