package iotago

import (
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

const (
	// NFTIDLength is the byte length of an NFTID.
	NFTIDLength = blake2b.Size256
)

var (
	emptyNFTID = [NFTIDLength]byte{}
)

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

	return addr.Chain()
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
	// The native tokens held by the output.
	NativeTokens NativeTokens `serix:"2,mapKey=nativeTokens,omitempty"`
	// The identifier of this NFT.
	NFTID NFTID `serix:"3,mapKey=nftId"`
	// The unlock conditions on this output.
	Conditions NFTOutputUnlockConditions `serix:"4,mapKey=unlockConditions,omitempty"`
	// The feature on the output.
	Features NFTOutputFeatures `serix:"5,mapKey=features,omitempty"`
	// The immutable feature on the output.
	ImmutableFeatures NFTOutputImmFeatures `serix:"6,mapKey=immutableFeatures,omitempty"`
}

func (n *NFTOutput) Clone() Output {
	return &NFTOutput{
		Amount:            n.Amount,
		Mana:              n.Mana,
		NativeTokens:      n.NativeTokens.Clone(),
		NFTID:             n.NFTID,
		Conditions:        n.Conditions.Clone(),
		Features:          n.Features.Clone(),
		ImmutableFeatures: n.ImmutableFeatures.Clone(),
	}
}

func (n *NFTOutput) Ident() Address {
	return n.Conditions.MustSet().Address().Address
}

func (n *NFTOutput) UnlockableBy(ident Address, pastBoundedSlotIndex SlotIndex, futureBoundedSlotIndex SlotIndex) bool {
	ok, _ := outputUnlockableBy(n, nil, ident, pastBoundedSlotIndex, futureBoundedSlotIndex)
	return ok
}

func (n *NFTOutput) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return outputOffsetVByteCost(rentStruct) +
		// prefix + amount + stored mana
		rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+BaseTokenSize+ManaSize) +
		n.NativeTokens.VBytes(rentStruct, nil) +
		rentStruct.VBFactorData.Multiply(NFTIDLength) +
		n.Conditions.VBytes(rentStruct, nil) +
		n.Features.VBytes(rentStruct, nil) +
		n.ImmutableFeatures.VBytes(rentStruct, nil)
}

func (n *NFTOutput) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
	workScoreNativeTokens, err := n.NativeTokens.WorkScore(workScoreStructure)
	if err != nil {
		return 0, err
	}

	workScoreConditions, err := n.Conditions.WorkScore(workScoreStructure)
	if err != nil {
		return 0, err
	}

	workScoreFeatures, err := n.Features.WorkScore(workScoreStructure)
	if err != nil {
		return 0, err
	}

	workScoreImmutableFeatures, err := n.ImmutableFeatures.WorkScore(workScoreStructure)
	if err != nil {
		return 0, err
	}

	return workScoreNativeTokens.Add(workScoreConditions, workScoreFeatures, workScoreImmutableFeatures)
}

func (n *NFTOutput) syntacticallyValidate() error {
	// Address should never be nil.
	address := n.Conditions.MustSet().Address().Address

	if address.Type() == AddressImplicitAccountCreation {
		return ErrImplicitAccountCreationAddressInInvalidOutput
	}

	return nil
}

func (n *NFTOutput) Chain() ChainID {
	return n.NFTID
}

func (n *NFTOutput) NativeTokenList() NativeTokens {
	return n.NativeTokens
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
		n.NativeTokens.Size() +
		NFTIDLength +
		n.Conditions.Size() +
		n.Features.Size() +
		n.ImmutableFeatures.Size()
}
