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
	NFTOutputUnlockCondition  interface{ UnlockCondition }
	NFTOutputFeature          interface{ Feature }
	NFTOutputImmFeature       interface{ Feature }
	NFTOutputUnlockConditions = UnlockConditions[NFTOutputUnlockCondition]
	NFTOutputFeatures         = Features[NFTOutputFeature]
	NFTOutputImmFeatures      = Features[NFTOutputImmFeature]
)

// NFTOutput is an output type used to implement non-fungible tokens.
type NFTOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount BaseToken `serix:""`
	// The stored mana held by the output.
	Mana Mana `serix:""`
	// The identifier of this NFT.
	NFTID NFTID `serix:""`
	// The unlock conditions on this output.
	UnlockConditions NFTOutputUnlockConditions `serix:",omitempty"`
	// The feature on the output.
	Features NFTOutputFeatures `serix:",omitempty"`
	// The immutable feature on the output.
	ImmutableFeatures NFTOutputImmFeatures `serix:",omitempty"`
}

func (n *NFTOutput) Clone() Output {
	return &NFTOutput{
		Amount:            n.Amount,
		Mana:              n.Mana,
		NFTID:             n.NFTID,
		UnlockConditions:  n.UnlockConditions.Clone(),
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

	if !n.UnlockConditions.Equal(otherOutput.UnlockConditions) {
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

func (n *NFTOutput) Owner() Address {
	return n.UnlockConditions.MustSet().Address().Address
}

func (n *NFTOutput) UnlockableBy(addr Address, pastBoundedSlotIndex SlotIndex, futureBoundedSlotIndex SlotIndex) bool {
	ok, _ := outputUnlockableBy(n, nil, addr, pastBoundedSlotIndex, futureBoundedSlotIndex)
	return ok
}

func (n *NFTOutput) StorageScore(storageScoreStruct *StorageScoreStructure, _ StorageScoreFunc) StorageScore {
	return storageScoreStruct.OffsetOutput +
		storageScoreStruct.FactorData().Multiply(StorageScore(n.Size())) +
		n.UnlockConditions.StorageScore(storageScoreStruct, nil) +
		n.Features.StorageScore(storageScoreStruct, nil) +
		n.ImmutableFeatures.StorageScore(storageScoreStruct, nil)
}

func (n *NFTOutput) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	workScoreConditions, err := n.UnlockConditions.WorkScore(workScoreParameters)
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

	return workScoreParameters.Output.Add(workScoreConditions, workScoreFeatures, workScoreImmutableFeatures)
}

func (n *NFTOutput) ChainID() ChainID {
	return n.NFTID
}

func (n *NFTOutput) FeatureSet() FeatureSet {
	return n.Features.MustSet()
}

func (n *NFTOutput) UnlockConditionSet() UnlockConditionSet {
	return n.UnlockConditions.MustSet()
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
		n.UnlockConditions.Size() +
		n.Features.Size() +
		n.ImmutableFeatures.Size()
}
