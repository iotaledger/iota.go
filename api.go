package iotago

import (
	"context"
	"encoding/binary"
	"errors"
	"sync"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/core/serix"
)

var (
	// ErrMissingProtocolParas is returned when ProtocolParameters are missing for operations which require them.
	ErrMissingProtocolParas = errors.New("missing protocol parameters")

	// internal API instance used to encode/decode objects where protocol parameters don't matter.
	_internalAPI   API
	_internalAPIMu = sync.RWMutex{}
)

func init() {
	_internalAPI = V2API(&ProtocolParameters{})
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
}

// LatestAPI creates a new API instance conforming to the latest IOTA protocol version.
func LatestAPI(protoPras *ProtocolParameters) API {
	return V2API(protoPras)
}

// calls the internally instantiated API to encode the given object.
func internalEncode(obj any, opts ...serix.Option) ([]byte, error) {
	_internalAPIMu.RLock()
	defer _internalAPIMu.RUnlock()
	return _internalAPI.Encode(obj, opts...)
}

// SwapInternalAPI swaps the internally used API of this lib with new.
func SwapInternalAPI(newAPI API) {
	_internalAPIMu.Lock()
	defer _internalAPIMu.Unlock()
	_internalAPI = newAPI
}

// NetworkID defines the ID of the network on which entities operate on.
type NetworkID = uint64

// NetworkIDFromString returns the network ID string's numerical representation.
func NetworkIDFromString(networkIDStr string) NetworkID {
	networkIDBlakeHash := blake2b.Sum256([]byte(networkIDStr))
	return binary.LittleEndian.Uint64(networkIDBlakeHash[:])
}

type protocolAPIContext string

// ProtocolAPIContextKey defines the key to use for a context containing a *ProtocolParameters.
const ProtocolAPIContextKey protocolAPIContext = "protocolParameters"

// ProtocolParameters defines the parameters of the protocol.
type ProtocolParameters struct {
	// The version of the protocol running.
	Version byte `serix:"0,mapKey=version"`
	// The human friendly name of the network.
	NetworkName string `serix:"1,lengthPrefixType=uint8,mapKey=networkName"`
	// The HRP prefix used for Bech32 addresses in the network.
	Bech32HRP NetworkPrefix `serix:"2,lengthPrefixType=uint8,mapKey=bech32Hrp"`
	// The minimum pow score of the network.
	MinPoWScore uint32 `serix:"3,mapKey=minPowScore"`
	// The below max depth parameter of the network.
	BelowMaxDepth uint8 `serix:"4,mapKey=belowMaxDepth"`
	// The rent structure used by given node/network.
	RentStructure RentStructure `serix:"5,mapKey=rentStructure"`
	// TokenSupply defines the current token supply on the network.
	TokenSupply uint64 `serix:"6,mapKey=tokenSupply"`
}

func (p ProtocolParameters) AsSerixContext() context.Context {
	return context.WithValue(context.Background(), ProtocolAPIContextKey, &p)
}

func (p ProtocolParameters) NetworkID() NetworkID {
	return NetworkIDFromString(p.NetworkName)
}

// Sizer is an object knowing its own byte size.
type Sizer interface {
	// Size returns the size of the object in terms of bytes.
	Size() int
}
