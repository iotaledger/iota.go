package iotago

import (
	"github.com/iotaledger/hive.go/core/safemath"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
)

// StorageScore defines the type of storage score.
type StorageScore uint64

// StorageScoreFactor defines the type of the storage score factor.
type StorageScoreFactor byte

var (
	// ErrStorageDepositNotCovered gets returned when a NonEphemeralObject does not cover the minimum deposit
	// which is calculated from its storage score.
	ErrStorageDepositNotCovered = ierrors.New("minimum storage deposit not covered")
	// ErrTypeIsNotSupportedRentStructure gets returned when a serializable was found to not be a supported RentStructure.
	ErrTypeIsNotSupportedRentStructure = ierrors.New("serializable is not a supported rent structure")
)

// Multiply multiplies in with this factor.
func (factor StorageScoreFactor) Multiply(in StorageScore) StorageScore {
	return StorageScore(factor) * in
}

// With joins two factors with each other.
func (factor StorageScoreFactor) With(other StorageScoreFactor) StorageScoreFactor {
	return factor + other
}

// RentParameters defines the parameters of rent cost calculations on objects which take node resources.
// This structure defines the minimum base token deposit required on an object. This deposit does not
// generate Mana, which serves as a rent payment in Mana for storing the object.
type RentParameters struct {
	// Defines the number of IOTA tokens required per unit of storage score.
	StorageCost BaseToken `serix:"0,mapKey=storageCost"`
	// Defines the factor to be used for data only fields.
	StorageScoreFactorData StorageScoreFactor `serix:"1,mapKey=storageScoreFactorData"`
	// Defines the offset to be used for key/lookup generating fields.
	StorageScoreOffsetOutput StorageScore `serix:"2,mapKey=storageScoreOffsetOutput"`
	// Defines the offset to be used for block issuer feature public keys.
	StorageScoreOffsetEd25519BlockIssuerKey StorageScore `serix:"3,mapKey=storageScoreOffsetEd25519BlockIssuerKey"`
	// Defines the offset to be used for staking feature.
	StorageScoreOffsetStakingFeature StorageScore `serix:"4,mapKey=storageScoreOffsetStakingFeature"`
	// Defines the offset to be used for delegation output.
	StorageScoreOffsetDelegation StorageScore `serix:"5,mapKey=storageScoreOffsetDelegation"`
}

// RentStructure includes the rent parameters and the additional factors/offsets computed from these parameters.
type RentStructure struct {
	RentParameters *RentParameters
	// The storage score that a minimal block issuer account needs to have minus the wrapping Basic Output.
	// Since this value is used for implicit account creation addresses, this value plus the wrapping
	// Basic Output (in which the Implicit Account Creation Address is contained in) results in the
	// minimum storage score of a block issuer account.
	StorageScoreOffsetImplicitAccountCreationAddress StorageScore
}

// StorageCost returns the cost of a single unit of storage score denoted in base tokens.
func (r *RentStructure) StorageCost() BaseToken {
	return r.RentParameters.StorageCost
}

// StorageScoreFactorData returns the factor to be used for data only fields.
func (r *RentStructure) StorageScoreFactorData() StorageScoreFactor {
	return r.RentParameters.StorageScoreFactorData
}

// StorageScoreOffsetOutput returns the offset to be used for all outputs to account for metadata created for the output.
func (r *RentStructure) StorageScoreOffsetOutput() StorageScore {
	return r.RentParameters.StorageScoreOffsetOutput
}

// StorageScoreOffsetEd25519BlockIssuerKey returns the offset to be used for block issuer feature public keys.
func (r *RentStructure) StorageScoreOffsetEd25519BlockIssuerKey() StorageScore {
	return r.RentParameters.StorageScoreOffsetEd25519BlockIssuerKey
}

// StorageScoreOffsetStakingFeature returns the offset to be used for staking feature.
func (r *RentStructure) StorageScoreOffsetStakingFeature() StorageScore {
	return r.RentParameters.StorageScoreOffsetStakingFeature
}

// StorageScoreOffsetDelegation returns the offset to be used for delegation output.
func (r *RentStructure) StorageScoreOffsetDelegation() StorageScore {
	return r.RentParameters.StorageScoreOffsetDelegation
}

// NewRentStructure creates a new RentStructure.
func NewRentStructure(rentParameters *RentParameters) *RentStructure {
	// create a dummy account with a block issuer feature to calculate the storage score.
	dummyAccountOutput := &AccountOutput{
		Amount:         0,
		Mana:           0,
		AccountID:      EmptyAccountID,
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

	// Set the storage score offset for implicit account creation addresses as
	// the difference between the storage score of the dummy account and the storage
	// score of the dummy basic output.
	storageScoreAccountOutput := dummyAccountOutput.StorageScore(rentStructure, nil)
	storageScoreBasicOutput := dummyBasicOutput.StorageScore(rentStructure, nil)
	rentStructure.StorageScoreOffsetImplicitAccountCreationAddress = lo.PanicOnErr(
		safemath.SafeSub(storageScoreAccountOutput, storageScoreBasicOutput),
	)

	return rentStructure
}

// CoversMinDeposit tells whether given this NonEphemeralObject, the base token amount fulfills the deposit requirements
// by examining the storage score of the object.
// Returns the minimum deposit computed and an error if it is not covered by the base token amount of the object.
func (r *RentStructure) CoversMinDeposit(object NonEphemeralObject, amount BaseToken) (BaseToken, error) {
	minDeposit, err := r.MinDeposit(object)
	if err != nil {
		return 0, ierrors.Wrap(err, "failed to compute min deposit")
	}
	if amount < minDeposit {
		return 0, ierrors.Wrapf(ErrStorageDepositNotCovered, "needed %d but only got %d", minDeposit, amount)
	}

	return minDeposit, nil
}

// MinDeposit returns the minimum deposit to cover a given object.
func (r *RentStructure) MinDeposit(object NonEphemeralObject) (BaseToken, error) {
	return safemath.SafeMul(r.StorageCost(), BaseToken(object.StorageScore(r, nil)))
}

// MinStorageDepositForReturnOutput returns the minimum renting costs for an BasicOutput which returns
// a StorageDepositReturnUnlockCondition amount back to the origin sender.
func (r *RentStructure) MinStorageDepositForReturnOutput(sender Address) (BaseToken, error) {
	return safemath.SafeMul(r.StorageCost(), BaseToken((&BasicOutput{Conditions: UnlockConditions[basicOutputUnlockCondition]{&AddressUnlockCondition{Address: sender}}, Amount: 0}).StorageScore(r, nil)))

}

func (r RentParameters) Equals(other RentParameters) bool {
	return r.StorageCost == other.StorageCost &&
		r.StorageScoreFactorData == other.StorageScoreFactorData &&
		r.StorageScoreOffsetOutput == other.StorageScoreOffsetOutput &&
		r.StorageScoreOffsetEd25519BlockIssuerKey == other.StorageScoreOffsetEd25519BlockIssuerKey &&
		r.StorageScoreOffsetStakingFeature == other.StorageScoreOffsetStakingFeature &&
		r.StorageScoreOffsetDelegation == other.StorageScoreOffsetDelegation
}

// NonEphemeralObject is an object which can not be pruned by nodes as it
// makes up an integral part to execute the IOTA protocol. This kind of objects are associated
// with costs in terms of the resources they take up.
type NonEphemeralObject interface {
	// StorageScore returns the cost this object has in terms of taking up
	// virtual and physical space within the data set needed to implement the protocol.
	// The override parameter acts as an escape hatch in case the cost needs to be adjusted
	// according to some external properties outside the NonEphemeralObject.
	StorageScore(rentStruct *RentStructure, override StorageScoreFunc) StorageScore
}

// StorageScoreFunc is a function which computes the storage score of a NonEphemeralObject.
type StorageScoreFunc func(rentStruct *RentStructure) StorageScore
