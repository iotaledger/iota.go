package keymanager_test

import (
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/assert"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/keymanager"
)

func TestMilestoneKeyManager(t *testing.T) {

	pubKey1, privKey1, err := ed25519.GenerateKey(nil)
	assert.NoError(t, err)

	pubKey2, privKey2, err := ed25519.GenerateKey(nil)
	assert.NoError(t, err)

	pubKey3, privKey3, err := ed25519.GenerateKey(nil)
	assert.NoError(t, err)

	pubKey4, privKey4, err := ed25519.GenerateKey(nil)
	assert.NoError(t, err)

	var msPubKey1 iotago.MilestonePublicKey
	copy(msPubKey1[:], pubKey1)

	var msPubKey2 iotago.MilestonePublicKey
	copy(msPubKey2[:], pubKey2)

	var msPubKey3 iotago.MilestonePublicKey
	copy(msPubKey3[:], pubKey3)

	var msPubKey4 iotago.MilestonePublicKey
	copy(msPubKey4[:], pubKey4)

	km := keymanager.New()
	km.AddKeyRange(pubKey1, 0, 0)
	km.AddKeyRange(pubKey2, 3, 10)
	km.AddKeyRange(pubKey3, 8, 15)
	km.AddKeyRange(pubKey4, 17, 0)

	keysIndex0 := km.PublicKeysForMilestoneIndex(0)
	assert.Len(t, keysIndex0, 1)

	keysIndex3 := km.PublicKeysForMilestoneIndex(3)
	assert.Len(t, keysIndex3, 2)

	keysIndex7 := km.PublicKeysForMilestoneIndex(7)
	assert.Len(t, keysIndex7, 2)

	keysIndex8 := km.PublicKeysForMilestoneIndex(8)
	assert.Len(t, keysIndex8, 3)

	keysIndex10 := km.PublicKeysForMilestoneIndex(10)
	assert.Len(t, keysIndex10, 3)

	keysIndex11 := km.PublicKeysForMilestoneIndex(11)
	assert.Len(t, keysIndex11, 2)

	keysIndex15 := km.PublicKeysForMilestoneIndex(15)
	assert.Len(t, keysIndex15, 2)

	keysIndex16 := km.PublicKeysForMilestoneIndex(16)
	assert.Len(t, keysIndex16, 1)

	keysIndex17 := km.PublicKeysForMilestoneIndex(17)
	assert.Len(t, keysIndex17, 2)

	keysIndex1000 := km.PublicKeysForMilestoneIndex(1000)
	assert.Len(t, keysIndex1000, 2)

	keysSet8 := km.PublicKeysSetForMilestoneIndex(8)
	assert.Len(t, keysSet8, 3)

	assert.Contains(t, keysSet8, msPubKey1)
	assert.Contains(t, keysSet8, msPubKey2)
	assert.Contains(t, keysSet8, msPubKey3)

	keyMapping8 := km.MilestonePublicKeyMappingForMilestoneIndex(8, []ed25519.PrivateKey{privKey1, privKey2, privKey3}, 2)
	assert.Len(t, keyMapping8, 2)

	assert.Equal(t, keyMapping8[msPubKey1], privKey1)
	assert.Equal(t, keyMapping8[msPubKey2], privKey2)

	keyMapping17 := km.MilestonePublicKeyMappingForMilestoneIndex(17, []ed25519.PrivateKey{privKey1, privKey4}, 2)
	assert.Len(t, keyMapping17, 2)

	assert.Equal(t, keyMapping17[msPubKey1], privKey1)
	assert.Equal(t, keyMapping17[msPubKey4], privKey4)

	keysSet17 := km.PublicKeysSetForMilestoneIndex(17)
	assert.Len(t, keysSet17, 2)
}
