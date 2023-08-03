package iotago

import (
	"github.com/iotaledger/hive.go/ierrors"
)

// VBytes defines the type of the virtual byte costs.
type VBytes uint64

// VByteCostFactor defines the type of the virtual byte cost factor.
type VByteCostFactor byte

const (
	// VByteCostFactorData defines the multiplier for data fields.
	VByteCostFactorData VByteCostFactor = 1
	// VByteCostFactorKey defines the multiplier for fields which can act as keys for lookups.
	VByteCostFactorKey VByteCostFactor = 10
	// VByteCostFactorBlockIssuerKeys defines the multiplier for block issuer public keys.
	VByteCostFactorBlockIssuerKeys VByteCostFactor = 10
	// VByteCostFactorStakingFeature defines the multiplier for account output's staking feature.
	VByteCostFactorStakingFeature
)

var (
	// ErrVByteDepositNotCovered gets returned when a NonEphemeralObject does not cover the minimum deposit
	// which is calculated from its virtual byte costs.
	ErrVByteDepositNotCovered = ierrors.New("virtual byte minimum deposit not covered")
	// ErrTypeIsNotSupportedRentStructure gets returned when a serializable was found to not be a supported RentStructure.
	ErrTypeIsNotSupportedRentStructure = ierrors.New("serializable is not a supported rent structure")
)

// Multiply multiplies in with this factor.
func (factor VByteCostFactor) Multiply(in VBytes) VBytes {
	return VBytes(factor) * in
}

// With joins two factors with each other.
func (factor VByteCostFactor) With(other VByteCostFactor) VByteCostFactor {
	return factor + other
}

// RentStructure defines the parameters of rent cost calculations on objects which take node resources.
// This structure defines the minimum base token deposit required on an object. This deposit does not
// generate Mana, which serves as a rent payment in Mana for storing the object.
type RentStructure struct {
	// Defines the rent of a single virtual byte denoted in IOTA tokens.
	VByteCost uint32 `serix:"0,mapKey=vByteCost"`
	// Defines the factor to be used for data only fields.
	VBFactorData VByteCostFactor `serix:"1,mapKey=vByteFactorData"`
	// Defines the factor to be used for key/lookup generating fields.
	VBFactorKey VByteCostFactor `serix:"2,mapKey=vByteFactorKey"`
	// Defines the factor to be used for block issuer feature public keys.
	VBFactorIssuerKeys VByteCostFactor `serix:"3,mapKey=vByteFactorIssuerKeys"`
	// Defines the factor to be used for staking feature.
	VBFactorStakingFeature VByteCostFactor `serix:"4,mapKey=vByteFactorStakingFeature"`
}

// CoversMinDeposit tells whether given this NonEphemeralObject, the base token amount fulfills the deposit requirements
// by examining the virtual bytes cost of the object.
// Returns the minimum deposit computed and an error if it is not covered by the base token amount of the object.
func (r *RentStructure) CoversMinDeposit(object NonEphemeralObject, amount BaseToken) (BaseToken, error) {
	minDeposit := r.MinDeposit(object)
	if amount < minDeposit {
		return 0, ierrors.Wrapf(ErrVByteDepositNotCovered, "needed %d but only got %d", minDeposit, amount)
	}

	return minDeposit, nil
}

// MinDeposit returns the minimum deposit to cover a given object.
func (r *RentStructure) MinDeposit(object NonEphemeralObject) BaseToken {
	return BaseToken(r.VByteCost) * BaseToken(object.VBytes(r, nil))
}

// MinStorageDepositForReturnOutput returns the minimum renting costs for an BasicOutput which returns
// a StorageDepositReturnUnlockCondition amount back to the origin sender.
func (r *RentStructure) MinStorageDepositForReturnOutput(sender Address) BaseToken {
	return BaseToken(r.VByteCost) * BaseToken((&BasicOutput{Conditions: UnlockConditions[basicOutputUnlockCondition]{&AddressUnlockCondition{Address: sender}}, Amount: 0}).VBytes(r, nil))
}

func (r RentStructure) Equals(other RentStructure) bool {
	return r.VByteCost == other.VByteCost &&
		r.VBFactorData == other.VBFactorData &&
		r.VBFactorKey == other.VBFactorKey
}

// NonEphemeralObject is an object which can not be pruned by nodes as it
// makes up an integral part to execute the IOTA protocol. This kind of objects are associated
// with costs in terms of the resources they take up.
type NonEphemeralObject interface {
	// VBytes returns the cost this object has in terms of taking up
	// virtual and physical space within the data set needed to implement the IOTA protocol.
	// The override parameter acts as an escape hatch in case the cost needs to be adjusted
	// according to some external properties outside the NonEphemeralObject.
	VBytes(rentStruct *RentStructure, override VBytesFunc) VBytes
}

// VBytesFunc is a function which computes the virtual byte cost of a NonEphemeralObject.
type VBytesFunc func(rentStruct *RentStructure) VBytes
