package api

import (
	"sync"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/runtime/options"
	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	iotago "github.com/iotaledger/iota.go/v4"
)

type EpochBasedProvider struct {
	mutex                             sync.RWMutex
	protocolParametersByVersion       map[iotago.Version]iotago.ProtocolParameters
	futureProtocolParametersByVersion map[iotago.Version]iotago.Identifier
	protocolVersions                  *ProtocolEpochVersions

	latestAPIMutex sync.RWMutex
	latestAPI      iotago.API

	currentAPIMutex sync.RWMutex
	currentAPI      iotago.API

	currentSlotMutex sync.RWMutex
	currentSlot      iotago.SlotIndex

	optsAPIForMissingVersionCallback func(version iotago.Version) (iotago.API, error)
}

func WithAPIForMissingVersionCallback(callback func(version iotago.Version) (iotago.API, error)) options.Option[EpochBasedProvider] {
	return func(provider *EpochBasedProvider) {
		provider.optsAPIForMissingVersionCallback = callback
	}
}

func NewEpochBasedProvider(opts ...options.Option[EpochBasedProvider]) *EpochBasedProvider {
	return options.Apply(&EpochBasedProvider{
		protocolParametersByVersion:       make(map[iotago.Version]iotago.ProtocolParameters),
		futureProtocolParametersByVersion: make(map[iotago.Version]iotago.Identifier),
		protocolVersions:                  NewProtocolEpochVersions(),
	}, opts)
}

func (e *EpochBasedProvider) SetCurrentSlot(slot iotago.SlotIndex) {
	e.currentSlotMutex.Lock()
	e.currentSlot = slot
	e.currentSlotMutex.Unlock()

	e.updateCurrentAPI(slot)
}

func (e *EpochBasedProvider) updateCurrentAPI(slot iotago.SlotIndex) {
	e.currentAPIMutex.Lock()
	defer e.currentAPIMutex.Unlock()

	latestAPI := e.LatestAPI()
	if latestAPI == nil {
		return
	}

	epoch := latestAPI.TimeProvider().EpochFromSlot(slot)
	version := e.VersionForEpoch(epoch)

	if e.currentAPI == nil || version > e.currentAPI.ProtocolParameters().Version() {
		e.currentAPI = lo.PanicOnErr(e.apiForVersion(version))
	}
}

func (e *EpochBasedProvider) AddProtocolParametersAtEpoch(protocolParameters iotago.ProtocolParameters, epoch iotago.EpochIndex) {
	e.AddProtocolParameters(protocolParameters)
	e.AddVersion(protocolParameters.Version(), epoch)
}

func (e *EpochBasedProvider) AddProtocolParameters(protocolParameters iotago.ProtocolParameters) {
	e.mutex.Lock()
	e.protocolParametersByVersion[protocolParameters.Version()] = protocolParameters
	delete(e.futureProtocolParametersByVersion, protocolParameters.Version())
	e.mutex.Unlock()

	e.latestAPIMutex.Lock()
	defer e.latestAPIMutex.Unlock()

	if e.latestAPI == nil || e.latestAPI.Version() < protocolParameters.Version() {
		e.latestAPI = lo.PanicOnErr(e.apiForVersion(protocolParameters.Version()))
	}
}

func (e *EpochBasedProvider) AddVersion(version iotago.Version, epoch iotago.EpochIndex) {
	e.mutex.Lock()
	e.protocolVersions.Add(version, epoch)
	e.mutex.Unlock()

	e.currentSlotMutex.Lock()
	defer e.currentSlotMutex.Unlock()

	e.updateCurrentAPI(e.currentSlot)
}

func (e *EpochBasedProvider) AddFutureVersion(version iotago.Version, protocolParamsHash iotago.Identifier, epoch iotago.EpochIndex) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.protocolVersions.Add(version, epoch)
	e.futureProtocolParametersByVersion[version] = protocolParamsHash
}

func (e *EpochBasedProvider) apiForVersion(version iotago.Version) (iotago.API, error) {
	protocolParams, exists := e.protocolParametersByVersion[version]
	if !exists {
		return nil, ierrors.Errorf("protocol parameters for version %d are not set", version)
	}

	//nolint:gocritic // further version will be added here
	switch protocolParams.Version() {
	case 3:
		return iotago.V3API(protocolParams), nil
	}

	if e.optsAPIForMissingVersionCallback != nil {
		return e.optsAPIForMissingVersionCallback(version)
	}

	return nil, ierrors.Errorf("no api available for parameters with version %d", protocolParams.Version())
}

func (e *EpochBasedProvider) APIForVersion(version iotago.Version) (iotago.API, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	return e.apiForVersion(version)
}

func (e *EpochBasedProvider) APIForSlot(slot iotago.SlotIndex) iotago.API {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	epoch := e.latestAPI.TimeProvider().EpochFromSlot(slot)

	return lo.PanicOnErr(e.apiForVersion(e.protocolVersions.VersionForEpoch(epoch)))
}

func (e *EpochBasedProvider) APIForEpoch(epoch iotago.EpochIndex) iotago.API {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	return lo.PanicOnErr(e.apiForVersion(e.protocolVersions.VersionForEpoch(epoch)))
}

func (e *EpochBasedProvider) LatestAPI() iotago.API {
	e.latestAPIMutex.RLock()
	defer e.latestAPIMutex.RUnlock()

	return e.latestAPI
}

func (e *EpochBasedProvider) CurrentAPI() iotago.API {
	e.currentAPIMutex.RLock()
	defer e.currentAPIMutex.RUnlock()

	return e.currentAPI
}

func (e *EpochBasedProvider) VersionsAndProtocolParametersHash() (iotago.Identifier, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	util := marshalutil.New()
	for _, version := range e.protocolVersions.Slice() {
		util.WriteBytes(lo.PanicOnErr(version.Version.Bytes()))
		util.WriteUint64(uint64(version.StartEpoch))

		var paramsHash iotago.Identifier
		params, paramsExist := e.protocolParametersByVersion[version.Version]
		if paramsExist {
			var err error
			if paramsHash, err = params.Hash(); err != nil {
				return iotago.Identifier{}, ierrors.Wrap(err, "failed to get protocol parameters hash")
			}
		} else {
			var hashExists bool
			if paramsHash, hashExists = e.futureProtocolParametersByVersion[version.Version]; !hashExists {
				return iotago.Identifier{}, ierrors.Errorf("protocol parameters for version %d are not set", version.Version)
			}
		}

		hashBytes, err := paramsHash.Bytes()
		if err != nil {
			return iotago.Identifier{}, ierrors.Wrap(err, "failed to get protocol parameters hash bytes")
		}
		util.WriteBytes(hashBytes)
	}

	return iotago.IdentifierFromData(util.Bytes()), nil
}

func (e *EpochBasedProvider) ProtocolParameters(version iotago.Version) iotago.ProtocolParameters {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	return e.protocolParametersByVersion[version]
}

func (e *EpochBasedProvider) ProtocolParametersHash(version iotago.Version) iotago.Identifier {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	if params, exists := e.protocolParametersByVersion[version]; exists {
		return lo.PanicOnErr(params.Hash())
	}

	return e.futureProtocolParametersByVersion[version]
}

func (e *EpochBasedProvider) ProtocolEpochVersions() []ProtocolEpochVersion {
	return e.protocolVersions.Slice()
}

func (e *EpochBasedProvider) EpochForVersion(version iotago.Version) (iotago.EpochIndex, bool) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	return e.protocolVersions.EpochForVersion(version)
}

func (e *EpochBasedProvider) VersionForEpoch(epoch iotago.EpochIndex) iotago.Version {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	return e.protocolVersions.VersionForEpoch(epoch)
}

func (e *EpochBasedProvider) VersionForSlot(slot iotago.SlotIndex) iotago.Version {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	epoch := e.latestAPI.TimeProvider().EpochFromSlot(slot)

	return e.protocolVersions.VersionForEpoch(epoch)
}

func (e *EpochBasedProvider) IsFutureVersion(version iotago.Version) bool {
	return e.CurrentAPI().Version() < version
}
