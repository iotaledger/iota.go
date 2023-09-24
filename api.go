package iotago

import (
	"encoding/binary"
	"time"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
)

type Version byte

const VersionLength = 1

func (v Version) Bytes() ([]byte, error) {
	return []byte{byte(v)}, nil
}

func VersionFromBytes(b []byte) (Version, int, error) {
	if len(b) < 1 {
		return 0, 0, ierrors.New("invalid version bytes length")
	}

	return Version(b[0]), 1, nil
}

// VersionSignaling defines the parameters used by signaling protocol parameters upgrade.
type VersionSignaling struct {
	// WindowSize is the size of the window in epochs to find which version of protocol parameters was most signaled, from currentEpoch - windowSize to currentEpoch.
	WindowSize uint8 `serix:"0,mapKey=windowSize"`
	// WindowTargetRatio is the target number of supporters for a version to win in a windowSize.
	WindowTargetRatio uint8 `serix:"1,mapKey=windowTargetRatio"`
	// ActivationOffset is the offset in epochs to activate the new version of protocol parameters.
	ActivationOffset uint8 `serix:"2,mapKey=activationOffset"`
}

func (s VersionSignaling) Equals(signaling VersionSignaling) bool {
	return s.WindowSize == signaling.WindowSize &&
		s.WindowTargetRatio == signaling.WindowTargetRatio &&
		s.ActivationOffset == signaling.ActivationOffset
}

// API handles en/decoding of IOTA protocol objects.
type API interface {
	// Encode encodes the given object to bytes.
	Encode(obj any, opts ...serix.Option) ([]byte, error)
	// Decode decodes the given bytes into object.
	Decode(b []byte, obj any, opts ...serix.Option) (int, error)
	// JSONEncode encodes the given object to its json representation.
	JSONEncode(obj any, opts ...serix.Option) ([]byte, error)
	// JSONDecode decodes the json data into object.
	JSONDecode(jsonData []byte, obj any, opts ...serix.Option) error
	// Underlying returns the underlying serix.API instance.
	Underlying() *serix.API
	// Version returns the version of the protocol this API is used with.
	Version() Version
	// ProtocolParameters returns the protocol parameters this API is used with.
	ProtocolParameters() ProtocolParameters
	// TimeProvider returns the underlying time provider used.
	TimeProvider() *TimeProvider
	// ManaDecayProvider returns the underlying mana decay provider used.
	ManaDecayProvider() *ManaDecayProvider
	// LivenessThresholdDuration returns the liveness threshold duration.
	LivenessThresholdDuration() time.Duration
	// MaxBlockWork returns the maximum block work score.
	MaxBlockWork() WorkScore
	// ComputedInitialReward returns the initial reward calculated from the parameters.
	ComputedInitialReward() uint64
	// ComputedFinalReward returns the final reward calculated from the parameters.
	ComputedFinalReward() uint64
}

func LatestProtocolVersion() Version {
	return apiV3Version
}

// LatestAPI creates a new API instance conforming to the latest IOTA protocol version.
func LatestAPI(protoParams ProtocolParameters) API {
	//nolint:forcetypeassert // we can safely assume that these are V3ProtocolParameters
	return V3API(protoParams.(*V3ProtocolParameters))
}

// NetworkID defines the ID of the network on which entities operate on.
type NetworkID = uint64

// NetworkIDFromString returns the network ID string's numerical representation.
func NetworkIDFromString(networkIDStr string) NetworkID {
	networkIDBlakeHash := blake2b.Sum256([]byte(networkIDStr))

	return binary.LittleEndian.Uint64(networkIDBlakeHash[:])
}

// ProtocolParametersType defines the type of protocol parameters.
type ProtocolParametersType byte

const (
	// ProtocolParametersV3 denotes a V3ProtocolParameters.
	ProtocolParametersV3 ProtocolParametersType = iota
)

// ProtocolParameters defines the parameters of the protocol.
type ProtocolParameters interface {
	// Version defines the version of the protocol running.
	Version() Version
	// NetworkName defines the human friendly name of the network.
	NetworkName() string
	// NetworkID defines the ID of the network which is derived from the network name.
	NetworkID() NetworkID
	// Bech32HRP defines the HRP prefix used for Bech32 addresses in the network.
	Bech32HRP() NetworkPrefix
	// RentStructure defines the rent structure used by given node/network.
	RentStructure() *RentStructure
	// WorkScoreStructure defines the work score structure used by the given network.
	WorkScoreStructure() *WorkScoreStructure
	// TokenSupply defines the current token supply on the network.
	TokenSupply() BaseToken

	TimeProvider() *TimeProvider

	ManaDecayProvider() *ManaDecayProvider

	StakingUnbondingPeriod() EpochIndex

	ValidationBlocksPerSlot() uint16

	PunishmentEpochs() EpochIndex

	LivenessThreshold() SlotIndex

	MinCommittableAge() SlotIndex

	MaxCommittableAge() SlotIndex

	// EpochNearingThreshold is used by the epoch orchestrator to detect the slot that should trigger a new committee
	// selection for the next and upcoming epoch.
	EpochNearingThreshold() SlotIndex

	// CongestionControlParameters returns the parameters used to calculate reference Mana cost.
	CongestionControlParameters() *CongestionControlParameters

	VersionSignaling() *VersionSignaling

	RewardsParameters() *RewardsParameters

	ImplicitAccountCreationParameters() *ImplicitAccountCreationParameters

	Bytes() ([]byte, error)

	Hash() (Identifier, error)

	Equals(other ProtocolParameters) bool
}

// Sizer is an object knowing its own byte size.
type Sizer interface {
	// Size returns the size of the object in terms of bytes.
	Size() int
}
