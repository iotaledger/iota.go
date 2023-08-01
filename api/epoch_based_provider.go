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

	latestVersionMutex sync.RWMutex
	latestVersion      iotago.Version

	currentSlotMutex sync.RWMutex
	currentSlot      iotago.SlotIndex

	optsAPIForMissingVersionCallback func(version iotago.Version) (iotago.API, error)
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
	defer e.currentSlotMutex.Unlock()

	e.currentSlot = slot
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

	e.latestVersionMutex.Lock()
	defer e.latestVersionMutex.Unlock()

	if e.latestVersion < protocolParameters.Version() {
		e.latestVersion = protocolParameters.Version()
	}
}

func (e *EpochBasedProvider) AddVersion(version iotago.Version, epoch iotago.EpochIndex) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.protocolVersions.Add(version, epoch)
}

func (e *EpochBasedProvider) AddFutureVersion(version iotago.Version, protocolParamsHash iotago.Identifier, epoch iotago.EpochIndex) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.protocolVersions.Add(version, epoch)
	e.futureProtocolParametersByVersion[version] = protocolParamsHash
}

func (e *EpochBasedProvider) APIForVersion(version iotago.Version) (iotago.API, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

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

func (e *EpochBasedProvider) APIForSlot(slot iotago.SlotIndex) iotago.API {
	epoch := e.LatestAPI().TimeProvider().EpochFromSlot(slot)
	return lo.PanicOnErr(e.APIForVersion(e.VersionForEpoch(epoch)))
}

func (e *EpochBasedProvider) APIForEpoch(epoch iotago.EpochIndex) iotago.API {
	return lo.PanicOnErr(e.APIForVersion(e.VersionForEpoch(epoch)))
}

func (e *EpochBasedProvider) LatestAPI() iotago.API {
	e.latestVersionMutex.RLock()
	defer e.latestVersionMutex.RUnlock()

	api, err := e.APIForVersion(e.latestVersion)
	if err != nil {
		panic(err)
	}

	return api
}

func (e *EpochBasedProvider) CurrentAPI() iotago.API {
	e.currentSlotMutex.RLock()
	defer e.currentSlotMutex.RUnlock()

	return e.APIForSlot(e.currentSlot)
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
			paramsBytes, err := params.Bytes()
			if err != nil {
				return iotago.Identifier{}, ierrors.Wrap(err, "failed to get protocol parameters bytes")
			}

			paramsHash = iotago.IdentifierFromData(paramsBytes)
		} else {
			var hashExists bool
			paramsHash, hashExists = e.futureProtocolParametersByVersion[version.Version]
			if !hashExists {
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

func WithAPIForMissingVersionCallback(callback func(version iotago.Version) (iotago.API, error)) options.Option[EpochBasedProvider] {
	return func(provider *EpochBasedProvider) {
		provider.optsAPIForMissingVersionCallback = callback
	}
}
