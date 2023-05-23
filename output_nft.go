package iotago

import (
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/util"
)

const (
	// 	NFTIDLength is the byte length of an NFTID.
	NFTIDLength = blake2b.Size256
)

var (
	emptyNFTID = [NFTIDLength]byte{}
)

// NFTID is the identifier for an NFT.
// It is computed as the Blake2b-256 hash of the OutputID of the output which created the NFT.
type NFTID [NFTIDLength]byte

func (nftID NFTID) ToHex() string {
	return EncodeHex(nftID[:])
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
	return EncodeHex(nftID[:])
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
	Amount uint64 `serix:"0,mapKey=amount"`
	// The native tokens held by the output.
	NativeTokens NativeTokens `serix:"1,mapKey=nativeTokens,omitempty"`
	// The identifier of this NFT.
	NFTID NFTID `serix:"2,mapKey=nftId"`
	// The unlock conditions on this output.
	Conditions NFTOutputUnlockConditions `serix:"3,mapKey=unlockConditions,omitempty"`
	// The feature on the output.
	Features NFTOutputFeatures `serix:"4,mapKey=features,omitempty"`
	// The immutable feature on the output.
	ImmutableFeatures NFTOutputImmFeatures `serix:"5,mapKey=immutableFeatures,omitempty"`
	// The stored mana held by the output.
	Mana uint64 `serix:"6,mapKey=mana"`
}

func (n *NFTOutput) Clone() Output {
	return &NFTOutput{
		Amount:            n.Amount,
		NativeTokens:      n.NativeTokens.Clone(),
		NFTID:             n.NFTID,
		Conditions:        n.Conditions.Clone(),
		Features:          n.Features.Clone(),
		ImmutableFeatures: n.ImmutableFeatures.Clone(),
		Mana:              n.Mana,
	}
}

func (n *NFTOutput) Ident() Address {
	return n.Conditions.MustSet().Address().Address
}

func (n *NFTOutput) UnlockableBy(ident Address, extParams *ExternalUnlockParameters) bool {
	ok, _ := outputUnlockable(n, nil, ident, extParams)
	return ok
}

func (n *NFTOutput) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	return outputOffsetVByteCost(rentStruct) +
		// prefix + amount
		rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+serializer.UInt64ByteSize) +
		n.NativeTokens.VBytes(rentStruct, nil) +
		rentStruct.VBFactorData.Multiply(NFTIDLength) +
		n.Conditions.VBytes(rentStruct, nil) +
		n.Features.VBytes(rentStruct, nil) +
		n.ImmutableFeatures.VBytes(rentStruct, nil)
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

func (n *NFTOutput) Deposit() uint64 {
	return n.Amount
}

func (a *NFTOutput) StoredMana() uint64 {
	return a.Mana
}

func (n *NFTOutput) Type() OutputType {
	return OutputNFT
}

func (n *NFTOutput) Size() int {
	return util.NumByteLen(byte(OutputNFT)) +
		util.NumByteLen(n.Amount) +
		n.NativeTokens.Size() +
		NFTIDLength +
		n.Conditions.Size() +
		n.Features.Size() +
		n.ImmutableFeatures.Size()
}
