package keymanager

import (
	"crypto/ed25519"
	"sort"

	iotago "github.com/iotaledger/iota.go/v3"
)

// KeyRange defines a public key of a milestone including the range it is valid.
type KeyRange struct {
	PublicKey  iotago.MilestonePublicKey
	StartIndex uint32
	EndIndex   uint32
}

// KeyManager provides public and private keys for ranges of milestone indexes.
type KeyManager struct {
	keyRanges []*KeyRange
}

// New returns a new KeyManager.
func New() *KeyManager {
	return &KeyManager{}
}

// AddKeyRange adds a new public key to the MilestoneKeyManager including its valid range.
func (k *KeyManager) AddKeyRange(publicKey ed25519.PublicKey, startIndex uint32, endIndex uint32) {

	var msPubKey iotago.MilestonePublicKey
	copy(msPubKey[:], publicKey)

	k.keyRanges = append(k.keyRanges, &KeyRange{PublicKey: msPubKey, StartIndex: startIndex, EndIndex: endIndex})

	// sort by start index
	sort.Slice(k.keyRanges, func(i int, j int) bool {
		return k.keyRanges[i].StartIndex < k.keyRanges[j].StartIndex
	})
}

func (k *KeyManager) KeyRanges() []*KeyRange {
	keyRanges := []*KeyRange{}
	for _, r := range k.keyRanges {
		keyRanges = append(keyRanges, &KeyRange{
			PublicKey:  r.PublicKey,
			StartIndex: r.StartIndex,
			EndIndex:   r.EndIndex,
		})
	}
	return keyRanges
}

// PublicKeysForMilestoneIndex returns the valid public keys for a certain milestone index.
func (k *KeyManager) PublicKeysForMilestoneIndex(msIndex uint32) []iotago.MilestonePublicKey {
	var pubKeys []iotago.MilestonePublicKey

	for _, pubKeyRange := range k.keyRanges {
		if pubKeyRange.StartIndex <= msIndex {
			if pubKeyRange.EndIndex >= msIndex || pubKeyRange.EndIndex == 0 {
				// EndIndex == 0 means the key is valid forever
				pubKeys = append(pubKeys, pubKeyRange.PublicKey)
			}
			continue
		}

		// no need to search further because the keys are sorted by StartIndex
		break
	}

	return pubKeys
}

// PublicKeysSetForMilestoneIndex returns a set of valid public keys for a certain milestone index.
func (k *KeyManager) PublicKeysSetForMilestoneIndex(msIndex uint32) iotago.MilestonePublicKeySet {
	pubKeys := k.PublicKeysForMilestoneIndex(msIndex)

	result := iotago.MilestonePublicKeySet{}

	for _, pubKey := range pubKeys {
		result[pubKey] = struct{}{}
	}

	return result
}

// MilestonePublicKeyMappingForMilestoneIndex returns a MilestonePublicKeyMapping for a certain milestone index.
func (k *KeyManager) MilestonePublicKeyMappingForMilestoneIndex(msIndex uint32, privateKeys []ed25519.PrivateKey, milestonePublicKeysCount int) iotago.MilestonePublicKeyMapping {
	pubKeySet := k.PublicKeysSetForMilestoneIndex(msIndex)

	result := iotago.MilestonePublicKeyMapping{}

	for _, privKey := range privateKeys {
		pubKey := privKey.Public().(ed25519.PublicKey)

		var msPubKey iotago.MilestonePublicKey
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
