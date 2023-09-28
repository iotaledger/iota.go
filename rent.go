package iotago

import (
	"github.com/iotaledger/hive.go/ierrors"
)

// VBytes defines the type of the virtual byte costs.
type VBytes uint64

// VByteFactor defines the type of the virtual byte cost factor.
type VByteFactor byte

var (
	// ErrVByteDepositNotCovered gets returned when a NonEphemeralObject does not cover the minimum deposit
	// which is calculated from its virtual byte costs.
	ErrVByteDepositNotCovered = ierrors.New("virtual byte minimum deposit not covered")
	// ErrTypeIsNotSupportedRentStructure gets returned when a serializable was found to not be a supported RentStructure.
	ErrTypeIsNotSupportedRentStructure = ierrors.New("serializable is not a supported rent structure")
)

// Multiply multiplies in with this factor.
func (factor VByteFactor) Multiply(in VBytes) VBytes {
	return VBytes(factor) * in
}

// With joins two factors with each other.
func (factor VByteFactor) With(other VByteFactor) VByteFactor {
	return factor + other
}

// RentParameters defines the parameters of rent cost calculations on objects which take node resources.
// This structure defines the minimum base token deposit required on an object. This deposit does not
// generate Mana, which serves as a rent payment in Mana for storing the object.
type RentParameters struct {
	// Defines the rent of a single virtual byte denoted in IOTA tokens.
	VByteCost uint32 `serix:"0,mapKey=vByteCost"`
	// Defines the factor to be used for data only fields.
	VBFactorData VByteFactor `serix:"1,mapKey=vByteFactorData"`
	// Defines the offset to be used for key/lookup generating fields.
	VBOffsetKey VBytes `serix:"2,mapKey=vByteOffsetKey"`
	// Defines the offset to be used for block issuer feature public keys.
	VBOffsetEd25519BlockIssuerKey VBytes `serix:"3,mapKey=vByteOffsetBlockIssuerKey"`
	// Defines the offset to be used for staking feature.
	VBOffsetStakingFeature VBytes `serix:"4,mapKey=vByteOffsetStakingFeature"`
	// Defines the offset to be used for delegation output.
	VBOffsetDelegation VBytes `serix:"5,mapKey=vByteOffsetDelegation"`
}

// RentStructure includes the rent parameters and the additional factors/offsets computed from these parameters.
type RentStructure struct {
	RentParameters                         *RentParameters
	VBOffsetImplicitAccountCreationAddress VBytes
}

// VByteCost returns the cost of a single virtual byte denoted in IOTA tokens.
func (r *RentStructure) VByteCost() uint32 {
	return r.RentParameters.VByteCost
}

// VBFactorData returns the factor to be used for data only fields.
func (r *RentStructure) VBFactorData() VByteFactor {
	return r.RentParameters.VBFactorData
}

// VBOffsetOutput returns the offset to be used for all outputs to account for metadata created for the output.
func (r *RentStructure) VBOffsetOutput() VBytes {
	return r.RentParameters.VBOffsetKey
}

// VBOffsetEd25519BlockIssuerKey returns the offset to be used for block issuer feature public keys.
func (r *RentStructure) VBOffsetEd25519BlockIssuerKey() VBytes {
	return r.RentParameters.VBOffsetEd25519BlockIssuerKey
}

// VBOffsetStakingFeature returns the offset to be used for staking feature.
func (r *RentStructure) VBOffsetStakingFeature() VBytes {
	return r.RentParameters.VBOffsetStakingFeature
}

// VBOffsetDelegation returns the offset to be used for delegation output.
func (r *RentStructure) VBOffsetDelegation() VBytes {
	return r.RentParameters.VBOffsetDelegation
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
					&Ed25519PublicKeyBlockIssuerKey{},
				},
			},
		},
		ImmutableFeatures: AccountOutputImmFeatures{},
	}

	dummyAddress := &Ed25519Address{}
	dummyBasicOutput := &BasicOutput{
		Conditions: UnlockConditions[basicOutputUnlockCondition]{
			&AddressUnlockCondition{
				Address: dummyAddress,
			},
		},
	}

	// create a rent structure with the provided rent parameters.
	rentStructure := &RentStructure{
		RentParameters: rentParameters,
	}

	// set the vbyte cost offset for implicit account creation addresses as the vbyte cost of the dummy account.
	vBDummyAccountOutput := dummyAccountOutput.VBytes(rentStructure, nil)
	vBDummyBasicOutput := dummyBasicOutput.VBytes(rentStructure, nil)
	vBDummyAddress := dummyAddress.VBytes(rentStructure, nil)
	rentStructure.VBOffsetImplicitAccountCreationAddress = vBDummyAccountOutput - vBDummyBasicOutput + vBDummyAddress

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
	return BaseToken(r.VByteCost()) * BaseToken(object.VBytes(r, nil))
}

// MinStorageDepositForReturnOutput returns the minimum renting costs for an BasicOutput which returns
// a StorageDepositReturnUnlockCondition amount back to the origin sender.
func (r *RentStructure) MinStorageDepositForReturnOutput(sender Address) BaseToken {
	return BaseToken(r.VByteCost()) * BaseToken((&BasicOutput{Conditions: UnlockConditions[basicOutputUnlockCondition]{&AddressUnlockCondition{Address: sender}}, Amount: 0}).VBytes(r, nil))
}

func (r RentParameters) Equals(other RentParameters) bool {
	return r.VByteCost == other.VByteCost &&
		r.VBFactorData == other.VBFactorData &&
		r.VBOffsetKey == other.VBOffsetKey &&
		r.VBOffsetEd25519BlockIssuerKey == other.VBOffsetEd25519BlockIssuerKey &&
		r.VBOffsetStakingFeature == other.VBOffsetStakingFeature &&
		r.VBOffsetDelegation == other.VBOffsetDelegation
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
