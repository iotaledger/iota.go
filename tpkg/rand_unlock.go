package tpkg

import (
	"github.com/iotaledger/hive.go/ierrors"
	iotago "github.com/iotaledger/iota.go/v4"
)

// RandUnlock returns a random unlock (except Signature, Reference, Account, Anchor, NFT).
func RandUnlock(allowEmptyUnlock bool) iotago.Unlock {
	unlockTypes := []iotago.UnlockType{iotago.UnlockSignature, iotago.UnlockReference, iotago.UnlockAccount, iotago.UnlockAnchor, iotago.UnlockNFT}

	if allowEmptyUnlock {
		unlockTypes = append(unlockTypes, iotago.UnlockEmpty)
	}

	unlockType := unlockTypes[RandInt(len(unlockTypes))]

	//nolint:exhaustive
	switch unlockType {
	case iotago.UnlockSignature:
		return RandEd25519SignatureUnlock()
	case iotago.UnlockReference:
		return RandReferenceUnlock()
	case iotago.UnlockAccount:
		return RandAccountUnlock()
	case iotago.UnlockAnchor:
		return RandAnchorUnlock()
	case iotago.UnlockNFT:
		return RandNFTUnlock()
	case iotago.UnlockEmpty:
		return &iotago.EmptyUnlock{}
	default:
		panic(ierrors.Wrapf(iotago.ErrUnknownUnlockType, "type %d", unlockType))
	}
}

// RandEd25519SignatureUnlock returns a random Ed25519 signature unlock.
func RandEd25519SignatureUnlock() *iotago.SignatureUnlock {
	return &iotago.SignatureUnlock{Signature: RandEd25519Signature()}
}

// RandReferenceUnlock returns a random reference unlock.
func RandReferenceUnlock() *iotago.ReferenceUnlock {
	return ReferenceUnlock(uint16(RandInt(1000)))
}

// RandAccountUnlock returns a random account unlock.
func RandAccountUnlock() *iotago.AccountUnlock {
	return &iotago.AccountUnlock{Reference: uint16(RandInt(1000))}
}

// RandAnchorUnlock returns a random anchor unlock.
func RandAnchorUnlock() *iotago.AnchorUnlock {
	return &iotago.AnchorUnlock{Reference: uint16(RandInt(1000))}
}

// RandNFTUnlock returns a random account unlock.
func RandNFTUnlock() *iotago.NFTUnlock {
	return &iotago.NFTUnlock{Reference: uint16(RandInt(1000))}
}

// RandMultiUnlock returns a random multi unlock.
func RandMultiUnlock() *iotago.MultiUnlock {
	unlockCnt := RandInt(10) + 1
	unlocks := make([]iotago.Unlock, 0, unlockCnt)

	for i := 0; i < unlockCnt; i++ {
		unlocks = append(unlocks, RandUnlock(true))
	}

	return &iotago.MultiUnlock{
		Unlocks: unlocks,
	}
}
