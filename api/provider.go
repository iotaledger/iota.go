package api

import (
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/pow"
	. "github.com/iotaledger/iota.go/trinary"
)

type Provider interface {
	Send(cmd interface{}, out interface{}) error
	SetSettings(settings interface{}) error
}

type API struct {
	provider       Provider
	attachToTangle AttachToTangleFunc
}

type CreateProviderFunc func(settings interface{}) (Provider, error)

type AttachToTangleFunc = func(trunkTxHash Hash, branchTxHash Hash, mwm uint64, trytes []Trytes) ([]Trytes, error)

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

	var attachToTangle AttachToTangleFunc
	if settings.PowFunc() != nil {
		attachToTangle = func(trunkTxHash Hash, branchTxHash Hash, mwm uint64, trytes []Trytes) ([]Trytes, error) {
			return pow.DoPoW(trunkTxHash, branchTxHash, trytes, mwm, settings.PowFunc())
		}
	} else {
		attachToTangle = nil
	}

	return &API{attachToTangle: attachToTangle, provider: provider}, nil
}
