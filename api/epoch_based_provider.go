package api

import (
	"sync"
	"time"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/runtime/options"
	"github.com/iotaledger/hive.go/serializer/v2/stream"
	iotago "github.com/iotaledger/iota.go/v4"
)

type EpochBasedProvider struct {
	mutex                             sync.RWMutex
	protocolParametersByVersion       map[iotago.Version]iotago.ProtocolParameters
	futureProtocolParametersByVersion map[iotago.Version]iotago.Identifier
	protocolVersions                  *ProtocolEpochVersions

	latestAPI iotago.API

	committedAPI iotago.API

	committedSlotMutex sync.RWMutex
	committedSlot      iotago.SlotIndex

	optsAPIForMissingVersionCallback func(protocolParameters iotago.ProtocolParameters) (iotago.API, error)
}

func WithAPIForMissingVersionCallback(callback func(protocolParameters iotago.ProtocolParameters) (iotago.API, error)) options.Option[EpochBasedProvider] {
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

func (e *EpochBasedProvider) SetCommittedSlot(slot iotago.SlotIndex) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.committedSlotMutex.Lock()
	e.committedSlot = slot
	e.committedSlotMutex.Unlock()

	e.updateCommittedAPI(slot)
}

func (e *EpochBasedProvider) AddProtocolParametersAtEpoch(protocolParameters iotago.ProtocolParameters, epoch iotago.EpochIndex) {
	e.AddProtocolParameters(protocolParameters)
	e.AddVersion(protocolParameters.Version(), epoch)
}

func (e *EpochBasedProvider) AddProtocolParameters(protocolParameters iotago.ProtocolParameters) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.protocolParametersByVersion[protocolParameters.Version()] = protocolParameters
	delete(e.futureProtocolParametersByVersion, protocolParameters.Version())

	if e.latestAPI == nil || e.latestAPI.Version() < protocolParameters.Version() {
		e.latestAPI = lo.PanicOnErr(e.apiForVersion(protocolParameters.Version()))
	}
}

func (e *EpochBasedProvider) AddVersion(version iotago.Version, epoch iotago.EpochIndex) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.protocolVersions.Add(version, epoch)

	e.committedSlotMutex.Lock()
	defer e.committedSlotMutex.Unlock()

	e.updateCommittedAPI(e.committedSlot)
}

func (e *EpochBasedProvider) AddFutureVersion(version iotago.Version, protocolParamsHash iotago.Identifier, epoch iotago.EpochIndex) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.protocolVersions.Add(version, epoch)
	e.futureProtocolParametersByVersion[version] = protocolParamsHash
}

func (e *EpochBasedProvider) apiForVersion(version iotago.Version) (iotago.API, error) {
	if e.latestAPI != nil && e.latestAPI.Version() == version {
		return e.latestAPI, nil
	}

	if e.committedAPI != nil && e.committedAPI.Version() == version {
		return e.committedAPI, nil
	}

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
		return e.optsAPIForMissingVersionCallback(protocolParams)
	}

	return nil, ierrors.Errorf("no api available for parameters with version %d", protocolParams.Version())
}

func (e *EpochBasedProvider) APIForVersion(version iotago.Version) (iotago.API, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	return e.apiForVersion(version)
}

func (e *EpochBasedProvider) APIForTime(t time.Time) iotago.API {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	slot := e.latestAPI.TimeProvider().SlotFromTime(t)
	epoch := e.latestAPI.TimeProvider().EpochFromSlot(slot)

	return lo.PanicOnErr(e.apiForVersion(e.protocolVersions.VersionForEpoch(epoch)))
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
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	return e.latestAPI
}

func (e *EpochBasedProvider) CommittedAPI() iotago.API {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	return e.committedAPI
}

func (e *EpochBasedProvider) VersionsAndProtocolParametersHash() (iotago.Identifier, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	byteBuffer := stream.NewByteBuffer()

	for _, version := range e.protocolVersions.Slice() {
		if err := stream.Write(byteBuffer, version.Version); err != nil {
			return iotago.Identifier{}, ierrors.Wrap(err, "failed to write Version")
		}
		if err := stream.Write(byteBuffer, version.StartEpoch); err != nil {
			return iotago.Identifier{}, ierrors.Wrap(err, "failed to write StartEpoch")
		}

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

		if err := stream.Write(byteBuffer, paramsHash); err != nil {
			return iotago.Identifier{}, ierrors.Wrap(err, "failed to write protocol parameters hash bytes")
		}
	}

	return iotago.IdentifierFromData(lo.PanicOnErr(byteBuffer.Bytes())), nil
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

func (e *EpochBasedProvider) VersionForSlot(slot iotago.SlotIndex) iotago.Version {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	epoch := e.latestAPI.TimeProvider().EpochFromSlot(slot)

	return e.protocolVersions.VersionForEpoch(epoch)
}

func (e *EpochBasedProvider) updateCommittedAPI(slot iotago.SlotIndex) {
	if e.latestAPI == nil {
		return
	}

	epoch := e.latestAPI.TimeProvider().EpochFromSlot(slot)
	version := e.versionForEpoch(epoch)

	if e.committedAPI == nil || version > e.committedAPI.ProtocolParameters().Version() {
		e.committedAPI = lo.PanicOnErr(e.apiForVersion(version))
	}
}

func (e *EpochBasedProvider) versionForEpoch(epoch iotago.EpochIndex) iotago.Version {
	return e.protocolVersions.VersionForEpoch(epoch)
}
