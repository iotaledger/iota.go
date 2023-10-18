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
	// ErrTypeIsNotSupportedStorageScoreStructure gets returned when a serializable was found to not be a supported StorageScoreStructure.
	ErrTypeIsNotSupportedStorageScoreStructure = ierrors.New("serializable is not a supported storage score structure")
)

// Multiply multiplies in with this factor.
func (factor StorageScoreFactor) Multiply(in StorageScore) StorageScore {
	return StorageScore(factor) * in
}

// With joins two factors with each other.
func (factor StorageScoreFactor) With(other StorageScoreFactor) StorageScoreFactor {
	return factor + other
}

// StorageScoreParameters defines the parameters of storage cost calculations on objects which take node resources.
// This structure defines the minimum base token deposit required on an object. This deposit does not
// generate Mana, which serves as a rent payment in Mana for storing the object.
type StorageScoreParameters struct {
	// Defines the number of IOTA tokens required per unit of storage score.
	StorageCost BaseToken `serix:"0,mapKey=storageCost"`
	// Defines the factor to be used for data only fields.
	FactorData StorageScoreFactor `serix:"1,mapKey=factorData"`
	// Defines the offset to be used for key/lookup generating fields.
	OffsetOutput StorageScore `serix:"2,mapKey=offsetOutput"`
	// Defines the offset to be used for block issuer feature public keys.
	OffsetEd25519BlockIssuerKey StorageScore `serix:"3,mapKey=offsetEd25519BlockIssuerKey"`
	// Defines the offset to be used for staking feature.
	OffsetStakingFeature StorageScore `serix:"4,mapKey=offsetStakingFeature"`
	// Defines the offset to be used for delegation output.
	OffsetDelegation StorageScore `serix:"5,mapKey=offsetDelegation"`
}

// StorageScoreStructure includes the storage score parameters and the additional factors/offsets computed from these parameters.
type StorageScoreStructure struct {
	StorageScoreParameters *StorageScoreParameters
	// The storage score that a minimal block issuer account needs to have minus the wrapping Basic Output.
	// Since this value is used for implicit account creation addresses, this value plus the wrapping
	// Basic Output (the Implicit Account Creation Address is contained in) results in the
	// minimum storage score of a block issuer account.
	OffsetImplicitAccountCreationAddress StorageScore
}

// StorageCost returns the cost of a single unit of storage score denoted in base tokens.
func (r *StorageScoreStructure) StorageCost() BaseToken {
	return r.StorageScoreParameters.StorageCost
}

// FactorData returns the factor to be used for data only fields.
func (r *StorageScoreStructure) FactorData() StorageScoreFactor {
	return r.StorageScoreParameters.FactorData
}

// OffsetOutput returns the offset to be used for all outputs to account for metadata created for the output.
func (r *StorageScoreStructure) OffsetOutput() StorageScore {
	return r.StorageScoreParameters.OffsetOutput
}

// OffsetEd25519BlockIssuerKey returns the offset to be used for block issuer feature public keys.
func (r *StorageScoreStructure) OffsetEd25519BlockIssuerKey() StorageScore {
	return r.StorageScoreParameters.OffsetEd25519BlockIssuerKey
}

// OffsetStakingFeature returns the offset to be used for staking feature.
func (r *StorageScoreStructure) OffsetStakingFeature() StorageScore {
	return r.StorageScoreParameters.OffsetStakingFeature
}

// OffsetDelegation returns the offset to be used for delegation output.
func (r *StorageScoreStructure) OffsetDelegation() StorageScore {
	return r.StorageScoreParameters.OffsetDelegation
}

// NewStorageScoreStructure creates a new StorageScoreStructure.
func NewStorageScoreStructure(storageScoreParameters *StorageScoreParameters) *StorageScoreStructure {
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

	// create a storage score structure with the provided storage score parameters.
	storageScoreStructure := &StorageScoreStructure{
		StorageScoreParameters: storageScoreParameters,
	}

	// Set the storage score offset for implicit account creation addresses as
	// the difference between the storage score of the dummy account and the storage
	// score of the dummy basic output minus the storage score of the dummy address.
	storageScoreAccountOutput := dummyAccountOutput.StorageScore(storageScoreStructure, nil)
	storageScoreBasicOutputWithoutAddress := dummyBasicOutput.StorageScore(storageScoreStructure, nil) - dummyAddress.StorageScore(storageScoreStructure, nil)
	storageScoreStructure.OffsetImplicitAccountCreationAddress = lo.PanicOnErr(
		safemath.SafeSub(storageScoreAccountOutput, storageScoreBasicOutputWithoutAddress),
	)

	return storageScoreStructure
}

// CoversMinDeposit tells whether given this NonEphemeralObject, the base token amount fulfills the deposit requirements
// by examining the storage score of the object.
// Returns the minimum deposit computed and an error if it is not covered by the base token amount of the object.
func (r *StorageScoreStructure) CoversMinDeposit(object NonEphemeralObject, amount BaseToken) (BaseToken, error) {
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
func (r *StorageScoreStructure) MinDeposit(object NonEphemeralObject) (BaseToken, error) {
	return safemath.SafeMul(r.StorageCost(), BaseToken(object.StorageScore(r, nil)))
}

// MinStorageDepositForReturnOutput returns the minimum storage costs for an BasicOutput which returns
// a StorageDepositReturnUnlockCondition amount back to the origin sender.
func (r *StorageScoreStructure) MinStorageDepositForReturnOutput(sender Address) (BaseToken, error) {
	return safemath.SafeMul(r.StorageCost(), BaseToken((&BasicOutput{Conditions: UnlockConditions[basicOutputUnlockCondition]{&AddressUnlockCondition{Address: sender}}, Amount: 0}).StorageScore(r, nil)))

}

func (r StorageScoreParameters) Equals(other StorageScoreParameters) bool {
	return r.StorageCost == other.StorageCost &&
		r.FactorData == other.FactorData &&
		r.OffsetOutput == other.OffsetOutput &&
		r.OffsetEd25519BlockIssuerKey == other.OffsetEd25519BlockIssuerKey &&
		r.OffsetStakingFeature == other.OffsetStakingFeature &&
		r.OffsetDelegation == other.OffsetDelegation
}

// NonEphemeralObject is an object which can not be pruned by nodes as it
// makes up an integral part to execute the IOTA protocol. This kind of objects are associated
// with costs in terms of the resources they take up.
type NonEphemeralObject interface {
	// StorageScore returns the cost this object has in terms of taking up
	// virtual and physical space within the data set needed to implement the protocol.
	// The override parameter acts as an escape hatch in case the cost needs to be adjusted
	// according to some external properties outside the NonEphemeralObject.
	StorageScore(storageScoreStruct *StorageScoreStructure, override StorageScoreFunc) StorageScore
}

// StorageScoreFunc is a function which computes the storage score of a NonEphemeralObject.
type StorageScoreFunc func(storageScoreStruct *StorageScoreStructure) StorageScore
