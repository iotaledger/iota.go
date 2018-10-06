package api

import (
	"github.com/iotaledger/iota.go/bundle"
	"github.com/iotaledger/iota.go/pow"
	. "github.com/iotaledger/iota.go/trinary"
	"github.com/pkg/errors"
)

var (
	ErrInvalidSettingsType = errors.New("incompatible settings type supplied")
)

type API struct {
	provider       Provider
	attachToTangle AttachToTangleFunc
}

type CreateProviderFunc func(settings interface{}) (Provider, error)

type AttachToTangleFunc = func(trunkTxHash Hash, branchTxHash Hash, mwm uint64, trytes []Trytes) ([]Trytes, error)

type Settings interface {
	PowFunc() pow.PowFunc
}

func ComposeAPI(settings Settings, createProvider *CreateProviderFunc) (*API, error) {
	var provider Provider
	var err error
	if createProvider != nil {
		provider, err = (*createProvider)(settings)
	} else {
		provider, err = NewHttpClient(settings)
	}
	if err != nil {
		return nil, err
	}

	var attachToTangle AttachToTangleFunc
	if settings.PowFunc() != nil {
		attachToTangle = func(trunkTxHash Hash, branchTxHash Hash, mwm uint64, trytes []Trytes) ([]Trytes, error) {
			return bundle.DoPoW(trunkTxHash, branchTxHash, trytes, mwm, settings.PowFunc())
		}
	} else {
		attachToTangle = nil
	}

	return &API{attachToTangle: attachToTangle, provider: provider}, nil
}

type Provider interface {
	Send(cmd interface{}, out interface{}) error
	SetSettings(settings interface{}) error
}
