package iotago

import (
	"github.com/iotaledger/hive.go/ierrors"
)

// VBytes defines the type of the virtual byte costs.
type VBytes uint64

// VByteCostFactor defines the type of the virtual byte cost factor.
type VByteCostFactor byte

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

// RentParameters defines the parameters of rent cost calculations on objects which take node resources.
// This structure defines the minimum base token deposit required on an object. This deposit does not
// generate Mana, which serves as a rent payment in Mana for storing the object.
type RentParameters struct {
	// Defines the rent of a single virtual byte denoted in IOTA tokens.
	VByteCost uint32 `serix:"0,mapKey=vByteCost"`
	// Defines the factor to be used for data only fields.
	VBFactorData VByteCostFactor `serix:"1,mapKey=vByteFactorData"`
	// Defines the factor to be used for key/lookup generating fields.
	VBFactorKey VByteCostFactor `serix:"2,mapKey=vByteFactorKey"`
	// Defines the factor to be used for block issuer feature public keys.
	VBFactorBlockIssuerKey VByteCostFactor `serix:"3,mapKey=vByteFactorBlockIssuerKey"`
	// Defines the factor to be used for staking feature.
	VBFactorStakingFeature VByteCostFactor `serix:"4,mapKey=vByteFactorStakingFeature"`
	// Defines the factor to be used for delegation output.
	VBFactorDelegation VByteCostFactor `serix:"5,mapKey=vByteFactorDelegation"`
}

// RentStructure includes the rent parameters and the additional factors computed from these parameters.
type RentStructure struct {
	RentParameters                         *RentParameters
	VBFactorImplicitAccountCreationAddress VByteCostFactor
}

// VByteCost returns the cost of a single virtual byte denoted in IOTA tokens.
func (r *RentStructure) VByteCost() uint32 {
	return r.RentParameters.VByteCost
}

// VBFactorData returns the factor to be used for data only fields.
func (r *RentStructure) VBFactorData() VByteCostFactor {
	return r.RentParameters.VBFactorData
}

// VBFactorKey returns the factor to be used for key/lookup generating fields.
func (r *RentStructure) VBFactorKey() VByteCostFactor {
	return r.RentParameters.VBFactorKey
}

// VBFactorBlockIssuerKey returns the factor to be used for block issuer feature public keys.
func (r *RentStructure) VBFactorBlockIssuerKey() VByteCostFactor {
	return r.RentParameters.VBFactorBlockIssuerKey
}

// VBFactorStakingFeature returns the factor to be used for staking feature.
func (r *RentStructure) VBFactorStakingFeature() VByteCostFactor {
	return r.RentParameters.VBFactorStakingFeature
}

// VBFactorDelegation returns the factor to be used for delegation output.
func (r *RentStructure) VBFactorDelegation() VByteCostFactor {
	return r.RentParameters.VBFactorDelegation
}

// NewRentStructure creates a new RentStructure.
func NewRentStructure(rentParameters *RentParameters) *RentStructure {
	// create a dummy account with a block issuer feature to calculate the vbytes cost.
	dummyAccountOutput := &AccountOutput{
		Amount:         0,
		Mana:           0,
		NativeTokens:   NativeTokens{},
		AccountID:      EmptyAccountID(),
		StateIndex:     0,
		StateMetadata:  []byte{},
		FoundryCounter: 0,
		Conditions: AccountOutputUnlockConditions{
			&GovernorAddressUnlockCondition{
				Address: &Ed25519Address{},
			},
			&StateControllerAddressUnlockCondition{
				Address: &Ed25519Address{},
			},
		},
		Features: AccountOutputFeatures{
			&BlockIssuerFeature{
				BlockIssuerKeys: BlockIssuerKeys{
					&Ed25519PublicKeyHashBlockIssuerKey{},
				},
			},
		},
		ImmutableFeatures: AccountOutputImmFeatures{},
	}
	// create a rent structure with the provided rent parameters.
	rentStructure := &RentStructure{
		RentParameters: rentParameters,
	}

	// set the vbyte cost factor for implicit account creation addresses as the vbyte cost of the dummy account.
	vBFactorImplicitAccountCreationAddress := dummyAccountOutput.VBytes(rentStructure, nil)
	rentStructure.VBFactorImplicitAccountCreationAddress = VByteCostFactor(vBFactorImplicitAccountCreationAddress)

	return rentStructure
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
	return BaseToken(r.RentParameters.VByteCost) * BaseToken(object.VBytes(r, nil))
}

// MinStorageDepositForReturnOutput returns the minimum renting costs for an BasicOutput which returns
// a StorageDepositReturnUnlockCondition amount back to the origin sender.
func (r *RentStructure) MinStorageDepositForReturnOutput(sender Address) BaseToken {
	return BaseToken(r.RentParameters.VByteCost) * BaseToken((&BasicOutput{Conditions: UnlockConditions[basicOutputUnlockCondition]{&AddressUnlockCondition{Address: sender}}, Amount: 0}).VBytes(r, nil))
}

func (r RentParameters) Equals(other RentParameters) bool {
	return r.VByteCost == other.VByteCost &&
		r.VBFactorData == other.VBFactorData &&
		r.VBFactorKey == other.VBFactorKey &&
		r.VBFactorBlockIssuerKey == other.VBFactorBlockIssuerKey &&
		r.VBFactorStakingFeature == other.VBFactorStakingFeature &&
		r.VBFactorDelegation == other.VBFactorDelegation
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
