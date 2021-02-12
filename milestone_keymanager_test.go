package iota_test

import (
	"testing"

	"github.com/iotaledger/iota.go/v2"
	"github.com/iotaledger/iota.go/v2/ed25519"

	"github.com/stretchr/testify/assert"
)

func TestMilestoneKeyManager(t *testing.T) {

	pubKey1, privKey1, err := ed25519.GenerateKey(nil)
	assert.NoError(t, err)

	pubKey2, privKey2, err := ed25519.GenerateKey(nil)
	assert.NoError(t, err)

	pubKey3, privKey3, err := ed25519.GenerateKey(nil)
	assert.NoError(t, err)

	km := iota.NewMilestoneKeyManager()
	km.AddKeyRange(pubKey1, 0, 0)
	km.AddKeyRange(pubKey2, 3, 10)
	km.AddKeyRange(pubKey3, 8, 15)

	keysIndex_0 := km.GetPublicKeysForMilestoneIndex(0)
	assert.Len(t, keysIndex_0, 1)

	keysIndex_3 := km.GetPublicKeysForMilestoneIndex(3)
	assert.Len(t, keysIndex_3, 2)

	keysIndex_7 := km.GetPublicKeysForMilestoneIndex(7)
	assert.Len(t, keysIndex_7, 2)

	keysIndex_8 := km.GetPublicKeysForMilestoneIndex(8)
	assert.Len(t, keysIndex_8, 3)

	keysIndex_10 := km.GetPublicKeysForMilestoneIndex(10)
	assert.Len(t, keysIndex_10, 3)

	keysIndex_11 := km.GetPublicKeysForMilestoneIndex(11)
	assert.Len(t, keysIndex_11, 2)

	keysIndex_15 := km.GetPublicKeysForMilestoneIndex(15)
	assert.Len(t, keysIndex_15, 2)

	keysIndex_16 := km.GetPublicKeysForMilestoneIndex(16)
	assert.Len(t, keysIndex_16, 1)

	keysIndex_1000 := km.GetPublicKeysForMilestoneIndex(1000)
	assert.Len(t, keysIndex_1000, 1)

	keysSet_8 := km.GetPublicKeysSetForMilestoneIndex(8)
	assert.Len(t, keysSet_8, 3)

	var msPubKey1 iota.MilestonePublicKey
	copy(msPubKey1[:], pubKey1)

	var msPubKey2 iota.MilestonePublicKey
	copy(msPubKey2[:], pubKey2)

	var msPubKey3 iota.MilestonePublicKey
	copy(msPubKey3[:], pubKey3)

	assert.Contains(t, keysSet_8, msPubKey1)
	assert.Contains(t, keysSet_8, msPubKey2)
	assert.Contains(t, keysSet_8, msPubKey3)

	keyMapping_8 := km.GetMilestonePublicKeyMappingForMilestoneIndex(8, []ed25519.PrivateKey{privKey1, privKey2, privKey3}, 2)
	assert.Len(t, keyMapping_8, 2)

	assert.Equal(t, keyMapping_8[msPubKey1], privKey1)
	assert.Equal(t, keyMapping_8[msPubKey2], privKey2)
}
