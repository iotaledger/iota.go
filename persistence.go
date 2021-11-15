package iotago

// VirtualByteCostFactor defines the type of the virtual byte cost factor.
type VirtualByteCostFactor uint64

const (
	// VirtualByteCostFactorData defines the multiplier for data fields.
	VirtualByteCostFactorData VirtualByteCostFactor = 1
	// VirtualByteCostFactorKey defines the multiplier for fields which can act as keys for lookups.
	VirtualByteCostFactorKey VirtualByteCostFactor = 10
)

// VirtualByteCostStructure defines the parameters of virtual byte cost calculations.
type VirtualByteCostStructure struct {
	FactorData VirtualByteCostFactor
	FactorKey  VirtualByteCostFactor
}

// NonEphemeralObject is an object which can not be pruned by nodes as it
// makes up an integral part of the protocol. This kind of objects are associated
// with a costs in terms of the resources they take up.
type NonEphemeralObject interface {
	// VirtualByteCost returns the cost this object has in terms of taking up
	// virtual and physical space within the ledger state.
	VirtualByteCost(costStruct *VirtualByteCostStructure) uint64
}
