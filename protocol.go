package iotago

import (
	"encoding/binary"
	"errors"

	"golang.org/x/crypto/blake2b"
)

var (
	// ErrMissingDeSerializationParas is returned when DeSerializationParameters are missing while
	// performing de/serialization on an object which requires them.
	ErrMissingDeSerializationParas = errors.New("missing de/serialization parameters")
)

// NetworkID defines the ID of the network on which entities operate on.
type NetworkID = uint64

// NetworkIDFromString returns the network ID string's numerical representation.
func NetworkIDFromString(networkIDStr string) NetworkID {
	networkIDBlakeHash := blake2b.Sum256([]byte(networkIDStr))
	return binary.LittleEndian.Uint64(networkIDBlakeHash[:])
}

// DeSerializationParameters defines parameters which must be given into de/serialization context
// if de/serialization is executed with syntactical validation.
type DeSerializationParameters struct {
	// Used to determine the validity of Outputs by checking whether
	// they fulfil the virtual byte rent cost given their deposit value.
	RentStructure *RentStructure
	// MinDustDeposit defines the minimum dust deposit.
	MinDustDeposit uint64
}
