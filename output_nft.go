package iotago

import (
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

var (
	// ErrInvalidNFTStateTransition gets returned when a NFT is doing an invalid state transition.
	ErrInvalidNFTStateTransition = ierrors.New("invalid NFT state transition")
)

const (
	// NFTIDLength is the byte length of an NFTID.
	NFTIDLength = blake2b.Size256
)

var (
	emptyNFTID = [NFTIDLength]byte{}
)

func EmptyNFTID() NFTID {
	return emptyNFTID
}

// NFTID is the identifier for an NFT.
// It is computed as the Blake2b-256 hash of the OutputID of the output which created the NFT.
type NFTID [NFTIDLength]byte

func (nftID NFTID) ToHex() string {
	return hexutil.EncodeHex(nftID[:])
}

// NFTIDs are NFTID(s).
type NFTIDs []NFTID

func (nftID NFTID) Addressable() bool {
	return true
}

func (nftID NFTID) Key() interface{} {
	return nftID.String()
}

func (nftID NFTID) FromOutputID(id OutputID) ChainID {
	addr := NFTAddressFromOutputID(id)

	return addr.ChainID()
}

func (nftID NFTID) Empty() bool {
	return nftID == emptyNFTID
}

func (nftID NFTID) Matches(other ChainID) bool {
	otherNFTID, isNFTID := other.(NFTID)
	if !isNFTID {
		return false
	}

	return nftID == otherNFTID
}

func (nftID NFTID) ToAddress() ChainAddress {
	var addr NFTAddress
	copy(addr[:], nftID[:])

	return &addr
}

func (nftID NFTID) String() string {
	return hexutil.EncodeHex(nftID[:])
}

func NFTIDFromOutputID(o OutputID) NFTID {
	ret := NFTID{}
	addr := NFTAddressFromOutputID(o)
	copy(ret[:], addr[:])

	return ret
}

type (
	nftOutputUnlockCondition  interface{ UnlockCondition }
	nftOutputFeature          interface{ Feature }
	nftOutputImmFeature       interface{ Feature }
	NFTOutputUnlockConditions = UnlockConditions[nftOutputUnlockCondition]
	NFTOutputFeatures         = Features[nftOutputFeature]
	NFTOutputImmFeatures      = Features[nftOutputImmFeature]
)

// NFTOutput is an output type used to implement non-fungible tokens.
type NFTOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount BaseToken `serix:"0,mapKey=amount"`
	// The stored mana held by the output.
	Mana Mana `serix:"1,mapKey=mana"`
	// The identifier of this NFT.
	NFTID NFTID `serix:"2,mapKey=nftId"`
	// The unlock conditions on this output.
	Conditions NFTOutputUnlockConditions `serix:"3,mapKey=unlockConditions,omitempty"`
	// The feature on the output.
	Features NFTOutputFeatures `serix:"4,mapKey=features,omitempty"`
	// The immutable feature on the output.
	ImmutableFeatures NFTOutputImmFeatures `serix:"5,mapKey=immutableFeatures,omitempty"`
}

func (n *NFTOutput) Clone() Output {
	return &NFTOutput{
		Amount:            n.Amount,
		Mana:              n.Mana,
		NFTID:             n.NFTID,
		Conditions:        n.Conditions.Clone(),
		Features:          n.Features.Clone(),
		ImmutableFeatures: n.ImmutableFeatures.Clone(),
	}
}

func (n *NFTOutput) Equal(other Output) bool {
	otherOutput, isSameType := other.(*NFTOutput)
	if !isSameType {
		return false
	}

	if n.Amount != otherOutput.Amount {
		return false
	}

	if n.Mana != otherOutput.Mana {
		return false
	}

	if n.NFTID != otherOutput.NFTID {
		return false
	}

	if !n.Conditions.Equal(otherOutput.Conditions) {
		return false
	}

	if !n.Features.Equal(otherOutput.Features) {
		return false
	}

	if !n.ImmutableFeatures.Equal(otherOutput.ImmutableFeatures) {
		return false
	}

	return true
}

func (n *NFTOutput) Ident() Address {
	return n.Conditions.MustSet().Address().Address
}

func (n *NFTOutput) UnlockableBy(ident Address, pastBoundedSlotIndex SlotIndex, futureBoundedSlotIndex SlotIndex) bool {
	ok, _ := outputUnlockableBy(n, nil, ident, pastBoundedSlotIndex, futureBoundedSlotIndex)
	return ok
}

func (n *NFTOutput) StorageScore(rentStruct *RentStructure, _ StorageScoreFunc) StorageScore {
	return offsetOutput(rentStruct) +
		rentStruct.StorageScoreFactorData().Multiply(StorageScore(n.Size())) +
		n.Conditions.StorageScore(rentStruct, nil) +
		n.Features.StorageScore(rentStruct, nil) +
		n.ImmutableFeatures.StorageScore(rentStruct, nil)
}

func (n *NFTOutput) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	workScoreConditions, err := n.Conditions.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	workScoreFeatures, err := n.Features.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	workScoreImmutableFeatures, err := n.ImmutableFeatures.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	return workScoreConditions.Add(workScoreFeatures, workScoreImmutableFeatures)
}

func (n *NFTOutput) syntacticallyValidate() error {
	// Address should never be nil.
	address := n.Conditions.MustSet().Address().Address

	if address.Type() == AddressImplicitAccountCreation {
		return ErrImplicitAccountCreationAddressInInvalidOutput
	}

	return nil
}

func (n *NFTOutput) ChainID() ChainID {
	return n.NFTID
}

func (n *NFTOutput) FeatureSet() FeatureSet {
	return n.Features.MustSet()
}

func (n *NFTOutput) UnlockConditionSet() UnlockConditionSet {
	return n.Conditions.MustSet()
}

func (n *NFTOutput) ImmutableFeatureSet() FeatureSet {
	return n.ImmutableFeatures.MustSet()
}

func (n *NFTOutput) BaseTokenAmount() BaseToken {
	return n.Amount
}

func (n *NFTOutput) StoredMana() Mana {
	return n.Mana
}

func (n *NFTOutput) Type() OutputType {
	return OutputNFT
}

func (n *NFTOutput) Size() int {
	// OutputType
	return serializer.OneByte +
		BaseTokenSize +
		ManaSize +
		NFTIDLength +
		n.Conditions.Size() +
		n.Features.Size() +
		n.ImmutableFeatures.Size()
}
