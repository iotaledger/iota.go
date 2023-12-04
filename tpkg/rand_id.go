package tpkg

import (
	iotago "github.com/iotaledger/iota.go/v4"
)

func RandIdentifier() iotago.Identifier {
	return Rand32ByteArray()
}

// RandBlockID produces a random block ID.
func RandBlockID() iotago.BlockID {
	return Rand36ByteArray()
}

// SortedRandBlockIDs returned random block IDs.
func SortedRandBlockIDs(count int) iotago.BlockIDs {
	slice := make([]iotago.BlockID, count)
	for i, ele := range SortedRand36ByteArray(count) {
		slice[i] = ele
	}

	return slice
}

func RandAccountID() iotago.AccountID {
	alias := iotago.AccountID{}
	copy(alias[:], RandBytes(iotago.AccountIDLength))

	return alias
}

func RandAnchorID() iotago.AnchorID {
	anchorID := iotago.AnchorID{}
	copy(anchorID[:], RandBytes(iotago.AnchorIDLength))

	return anchorID
}

func RandNFTID() iotago.NFTID {
	nft := iotago.NFTID{}
	copy(nft[:], RandBytes(iotago.NFTIDLength))

	return nft
}

func RandDelegationID() iotago.DelegationID {
	delegation := iotago.DelegationID{}
	copy(delegation[:], RandBytes(iotago.DelegationIDLength))

	return delegation
}

func RandNativeTokenID() iotago.NativeTokenID {
	var nativeTokenID iotago.NativeTokenID
	copy(nativeTokenID[:], RandBytes(iotago.NativeTokenIDLength))

	// the underlying address needs to be an account address
	nativeTokenID[0] = byte(iotago.AddressAccount)

	// set the simple token scheme type
	nativeTokenID[iotago.FoundryIDLength-iotago.FoundryTokenSchemeLength] = byte(iotago.TokenSchemeSimple)

	return nativeTokenID
}
