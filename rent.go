package iotago

import (
	"errors"
	"fmt"
)

// VByteCostFactor defines the type of the virtual byte cost factor.
type VByteCostFactor byte

const (
	// VByteCostFactorData defines the multiplier for data fields.
	VByteCostFactorData VByteCostFactor = 1
	// VByteCostFactorKey defines the multiplier for fields which can act as keys for lookups.
	VByteCostFactorKey VByteCostFactor = 10
)

var (
	// ErrVByteRentNotCovered gets returned when a NonEphemeralObject does not cover the state rent
	// cost which are calculated from its virtual byte costs.
	ErrVByteRentNotCovered = errors.New("virtual byte rent costs not covered")
	// ErrTypeIsNotSupportedRentStructure gets returned when a serializable was found to not be a supported RentStructure.
	ErrTypeIsNotSupportedRentStructure = errors.New("serializable is not a supported rent structure")
)

// Multiply multiplies in with this factor.
func (factor VByteCostFactor) Multiply(in uint64) uint64 {
	return uint64(factor) * in
}

// With joins two factors with each other.
func (factor VByteCostFactor) With(other VByteCostFactor) VByteCostFactor {
	return factor + other
}

// RentStructure defines the parameters of rent cost calculations on objects which take node resources.
type RentStructure struct {
	// Defines the rent of a single virtual byte denoted in IOTA tokens.
	VByteCost uint32 `serix:"0,mapKey=vByteCost"`
	// Defines the factor to be used for data only fields.
	VBFactorData VByteCostFactor `serix:"1,mapKey=vByteFactorData"`
	// defines the factor to be used for key/lookup generating fields.
	VBFactorKey VByteCostFactor `serix:"2,mapKey=vByteFactorKey"`
}

// CoversStateRent tells whether given this NonEphemeralObject, the given rent fulfills the renting costs
// by examining the virtual bytes cost of the object.
// Returns the minimum rent computed and an error if it is not covered by rent.
func (r *RentStructure) CoversStateRent(object NonEphemeralObject, rent uint64) (uint64, error) {
	minRent := r.MinRent(object)
	if rent < minRent {
		return 0, fmt.Errorf("%w: needed %d but only got %d", ErrVByteRentNotCovered, minRent, rent)
	}
	return minRent, nil
}

// MinRent returns the minimum rent to cover a given object.
func (r *RentStructure) MinRent(object NonEphemeralObject) uint64 {
	return uint64(r.VByteCost) * object.VBytes(r, nil)
}

// MinStorageDepositForReturnOutput returns the minimum renting costs for an BasicOutput which returns
// a StorageDepositReturnUnlockCondition amount back to the origin sender.
func (r *RentStructure) MinStorageDepositForReturnOutput(sender Address) uint64 {
	return (&BasicOutput{Conditions: UnlockConditions[BasicOutputUnlockCondition]{&AddressUnlockCondition{Address: sender}}, Amount: 0}).VBytes(r, nil)
}

// NonEphemeralObject is an object which can not be pruned by nodes as it
// makes up an integral part to execute the IOTA protocol. This kind of objects are associated
// with costs in terms of the resources they take up.
type NonEphemeralObject interface {
	// VBytes returns the cost this object has in terms of taking up
	// virtual and physical space within the data set needed to implement the IOTA protocol.
	// The override parameter acts as an escape hatch in case the cost needs to be adjusted
	// according to some external properties outside the NonEphemeralObject.
	VBytes(rentStruct *RentStructure, override VBytesFunc) uint64
}

// VBytesFunc is a function which computes the virtual byte cost of a NonEphemeralObject.
type VBytesFunc func(rentStruct *RentStructure) uint64
