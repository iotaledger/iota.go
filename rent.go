package iotago

import (
	"errors"
	"fmt"
)

// VByteCostFactor defines the type of the virtual byte cost factor.
type VByteCostFactor uint64

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

	// ZeroRentParas are test parameters for de/serialization using zero vbyte rent cost.
	// Only use this var in testing. Do not modify.
	ZeroRentParas = &DeSerializationParameters{RentStructure: &RentStructure{
		VByteCost:    0,
		VBFactorData: 0,
		VBFactorKey:  0,
	}}
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
	VByteCost uint64
	// Defines the factor to be used for data only fields.
	VBFactorData VByteCostFactor
	// defines the factor to be used for key/lookup generating fields.
	VBFactorKey VByteCostFactor
}

// CoversStateRent tells whether given this NonEphemeralObject, the given rent fulfils the renting costs
// by examining the virtual bytes cost of the object.
// Returns the minimum rent computed and an error if it is not covered by rent.
func (vbcs *RentStructure) CoversStateRent(object NonEphemeralObject, rent uint64) (uint64, error) {
	minRent := vbcs.VByteCost * object.VByteCost(vbcs, nil)
	if rent < minRent {
		return 0, fmt.Errorf("%w: needed %d but only got %d", ErrVByteRentNotCovered, minRent, rent)
	}
	return minRent, nil
}

// MinDustDeposit returns the minimum renting costs for an ExtendedOutput which returns
// a DustDepositReturnUnlockCondition amount back to the origin sender.
func (vbcs *RentStructure) MinDustDeposit(sender Address) uint64 {
	return (&ExtendedOutput{Conditions: UnlockConditions{&AddressUnlockCondition{Address: sender}}, Amount: 0}).VByteCost(vbcs, nil)
}

// NonEphemeralObject is an object which can not be pruned by nodes as it
// makes up an integral part to execute the IOTA protocol. This kind of objects are associated
// with costs in terms of the resources they take up.
type NonEphemeralObject interface {
	// VByteCost returns the cost this object has in terms of taking up
	// virtual and physical space within the data set needed to implement the IOTA protocol.
	// The override parameter acts as an escape hatch in case the cost needs to be adjusted
	// according to some external properties outside the NonEphemeralObject.
	VByteCost(costStruct *RentStructure, override VByteCostFunc) uint64
}

// VByteCostFunc is a function which computes the virtual byte cost of a NonEphemeralObject.
type VByteCostFunc func(costStruct *RentStructure) uint64
