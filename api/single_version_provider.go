package api

import (
	iotago "github.com/iotaledger/iota.go/v4"
)

func SingleVersionProvider(api iotago.API) Provider {
	return &singleVersionProvider{api: api}
}

type singleVersionProvider struct {
	api iotago.API
}

func (t *singleVersionProvider) APIForVersion(iotago.Version) (iotago.API, error) {
	return t.api, nil
}

func (t *singleVersionProvider) APIForSlot(iotago.SlotIndex) iotago.API {
	return t.api
}

func (t *singleVersionProvider) APIForEpoch(iotago.EpochIndex) iotago.API {
	return t.api
}

func (t *singleVersionProvider) LatestAPI() iotago.API {
	return t.api
}

func (t *singleVersionProvider) CurrentAPI() iotago.API {
	return t.api
}
