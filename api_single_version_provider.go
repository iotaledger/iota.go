package iotago

import (
	"time"
)

func SingleVersionProvider(api API) APIProvider {
	return &singleVersionProvider{api: api}
}

type singleVersionProvider struct {
	api API
}

func (t *singleVersionProvider) APIForVersion(Version) (API, error) {
	return t.api, nil
}

func (t *singleVersionProvider) APIForTime(time.Time) API {
	return t.api
}

func (t *singleVersionProvider) APIForSlot(SlotIndex) API {
	return t.api
}

func (t *singleVersionProvider) APIForEpoch(EpochIndex) API {
	return t.api
}

func (t *singleVersionProvider) LatestAPI() API {
	return t.api
}

func (t *singleVersionProvider) CommittedAPI() API {
	return t.api
}
