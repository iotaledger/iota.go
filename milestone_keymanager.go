package iota

import (
	"sort"

	"github.com/iotaledger/iota.go/v2/ed25519"
)

// MilestoneKeyRange defines a public key of a milestone including the range it is valid.
type MilestoneKeyRange struct {
	PublicKey  MilestonePublicKey
	StartIndex uint32
	EndIndex   uint32
}

// MilestoneKeyManager provides public and private keys for ranges of milestone indexes.
type MilestoneKeyManager struct {
	keyRanges []*MilestoneKeyRange
}

// NewMilestoneKeyManager returns a new MilestoneKeyManager.
func NewMilestoneKeyManager() *MilestoneKeyManager {
	return &MilestoneKeyManager{}
}

// AddKeyRange adds a new public key to the MilestoneKeyManager including its valid range.
func (k *MilestoneKeyManager) AddKeyRange(publicKey ed25519.PublicKey, startIndex uint32, endIndex uint32) error {

	var msPubKey MilestonePublicKey
	copy(msPubKey[:], publicKey)

	k.keyRanges = append(k.keyRanges, &MilestoneKeyRange{PublicKey: msPubKey, StartIndex: startIndex, EndIndex: endIndex})

	// sort by start index
	sort.Slice(k.keyRanges, func(i int, j int) bool {
		return k.keyRanges[i].StartIndex < k.keyRanges[j].StartIndex
	})

	return nil
}

// GetPublicKeysForMilestoneIndex returns the valid public keys for a certain milestone index.
func (k *MilestoneKeyManager) GetPublicKeysForMilestoneIndex(msIndex uint32) []MilestonePublicKey {
	var pubKeys []MilestonePublicKey

	for _, pubKeyRange := range k.keyRanges {
		if pubKeyRange.StartIndex <= msIndex {
			if pubKeyRange.EndIndex >= msIndex || pubKeyRange.StartIndex == pubKeyRange.EndIndex {
				// startIndex == endIndex means the key is valid forever
				pubKeys = append(pubKeys, pubKeyRange.PublicKey)
			}
			continue
		}
		break
	}

	return pubKeys
}

// GetPublicKeysSetForMilestoneIndex returns a set of valid public keys for a certain milestone index.
func (k *MilestoneKeyManager) GetPublicKeysSetForMilestoneIndex(msIndex uint32) MilestonePublicKeySet {
	pubKeys := k.GetPublicKeysForMilestoneIndex(msIndex)

	result := MilestonePublicKeySet{}

	for _, pubKey := range pubKeys {
		result[pubKey] = struct{}{}
	}

	return result
}

// GetMilestonePublicKeyMappingForMilestoneIndex returns a MilestonePublicKeyMapping for a certain milestone index.
func (k *MilestoneKeyManager) GetMilestonePublicKeyMappingForMilestoneIndex(msIndex uint32, privateKeys []ed25519.PrivateKey, milestonePublicKeysCount int) MilestonePublicKeyMapping {
	pubKeySet := k.GetPublicKeysSetForMilestoneIndex(msIndex)

	result := MilestonePublicKeyMapping{}

	for _, privKey := range privateKeys {
		pubKey := privKey.Public().(ed25519.PublicKey)

		var msPubKey MilestonePublicKey
		copy(msPubKey[:], pubKey)

		if _, exists := pubKeySet[msPubKey]; exists {
			result[msPubKey] = privKey

			if len(result) == len(pubKeySet) {
				break
			}

			if len(result) == milestonePublicKeysCount {
				break
			}
		}
	}

	return result
}
