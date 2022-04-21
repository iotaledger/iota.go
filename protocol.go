package iotago

import (
	"encoding/binary"
	"errors"

	"golang.org/x/crypto/blake2b"
)

var (
	// ErrMissingProtocolParas is returned when ProtocolParameters are missing for operations which require them.
	ErrMissingProtocolParas = errors.New("missing protocol parameters")
)

// NetworkID defines the ID of the network on which entities operate on.
type NetworkID = uint64

// NetworkIDFromString returns the network ID string's numerical representation.
func NetworkIDFromString(networkIDStr string) NetworkID {
	networkIDBlakeHash := blake2b.Sum256([]byte(networkIDStr))
	return binary.LittleEndian.Uint64(networkIDBlakeHash[:])
}

// ProtocolParameters defines the parameters of the protocol.
type ProtocolParameters struct {
	// The version of the protocol running.
	Version byte
	// The human friendly name of the network.
	NetworkName string
	// The HRP prefix used for Bech32 addresses in the network.
	Bech32HRP string
	// The minimum pow score of the network.
	MinPowScore float64
	// The rent structure used by given node/network.
	RentStructure RentStructure
	// TokenSupply defines the current token supply on the network.
	TokenSupply uint64
}
