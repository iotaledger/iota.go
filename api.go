package iotago

import (
	"encoding/binary"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/serializer/v2/serix"
)

type Version = byte

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
}

func LatestProtocolVersion() Version {
	return apiV3Version
}

// LatestAPI creates a new API instance conforming to the latest IOTA protocol version.
func LatestAPI(protoParams ProtocolParameters) API {
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
	// TokenSupply defines the current token supply on the network.
	TokenSupply() BaseToken

	TimeProvider() *TimeProvider

	ManaDecayProvider() *ManaDecayProvider

	StakingUnbondingPeriod() EpochIndex

	LivenessThreshold() SlotIndex

	MinCommittableAge() SlotIndex

	MaxCommittableAge() SlotIndex

	// EpochNearingThreshold is used by the epoch orchestrator to detect the slot that should trigger a new committee
	// selection for the next and upcoming epoch.
	EpochNearingThreshold() SlotIndex

	Bytes() ([]byte, error)
}

// Sizer is an object knowing its own byte size.
type Sizer interface {
	// Size returns the size of the object in terms of bytes.
	Size() int
}
