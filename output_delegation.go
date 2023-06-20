package iotago

import (
	"golang.org/x/crypto/blake2b"
)

const (
	// 	DelegationIDLength is the byte length of an NFTID.
	DelegationIDLength = blake2b.Size256
)

var (
	emptyDelegationID = [DelegationIDLength]byte{}
)

// NFTID is the identifier for an NFT.
// It is computed as the Blake2b-256 hash of the OutputID of the output which created the NFT.
type DelegationID [DelegationIDLength]byte

// DelegationIDs are DelegationID(s).
type DelegationIDs []DelegationID

func (delegationId DelegationID) Addressable() bool {
	return false
}

type (
	delegationOutputUnlockCondition  interface{ UnlockCondition }
	delegationOutputImmFeature       interface{ Feature }
	DelegationOutputUnlockConditions = UnlockConditions[nftOutputUnlockCondition]
	DelegationOutputImmFeatures      = Features[nftOutputImmFeature]
)

// DelegationOutput is an output type used to implement delegation.
type DelegationOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount uint64 `serix:"0,mapKey=amount"`
	// The amount of IOTA tokens that were delegated when the output was created.
	DelegatedAmount uint64 `serix:"1,mapKey=delegatedAmount"`
	// The Account ID of the validator to which this output is delegating.
	ValidatorId AccountID `serix:"2,mapKey=validatorId"`
	// The index of the first epoch for which this output delegates.
	StartEpoch uint64
	// The index of the last epoch for which this output delegates.
	EndEpoch uint64
	// The unlock conditions on this output.
	Conditions NFTOutputUnlockConditions `serix:"3,mapKey=unlockConditions,omitempty"`
	// The feature on the output.
	Features NFTOutputFeatures `serix:"4,mapKey=features,omitempty"`
	// The immutable feature on the output.
	ImmutableFeatures NFTOutputImmFeatures `serix:"5,mapKey=immutableFeatures,omitempty"`
	// The stored mana held by the output.
	Mana uint64 `serix:"6,mapKey=mana"`
}
