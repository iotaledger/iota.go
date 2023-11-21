package wallet

import (
	"crypto/ed25519"
	"fmt"

	"github.com/wollac/iota-crypto-demo/pkg/bip32path"
	"github.com/wollac/iota-crypto-demo/pkg/slip10"
	"github.com/wollac/iota-crypto-demo/pkg/slip10/eddsa"

	"github.com/iotaledger/hive.go/ierrors"
	iotago "github.com/iotaledger/iota.go/v4"
)

const (
	pathString = "44'/4218'/0'/%d'"
)

// KeyManager is a hierarchical deterministic key manager.
type KeyManager struct {
	seed  []byte
	index uint64
}

func NewKeyManager(seed []byte, index uint64) *KeyManager {
	return &KeyManager{
		seed:  seed,
		index: index,
	}
}

// KeyPair calculates an ed25519 key pair by using slip10.
func (k *KeyManager) KeyPair() (ed25519.PrivateKey, ed25519.PublicKey) {
	path, err := bip32path.ParsePath(fmt.Sprintf(pathString, k.index))
	if err != nil {
		panic(err)
	}

	curve := eddsa.Ed25519()
	key, err := slip10.DeriveKeyFromPath(k.seed, curve, path)
	if err != nil {
		panic(err)
	}

	pubKey, privKey := key.Key.(eddsa.Seed).Ed25519Key()

	return ed25519.PrivateKey(privKey), ed25519.PublicKey(pubKey)
}

// AddressSigner returns an address signer.
func (k *KeyManager) AddressSigner() iotago.AddressSigner {
	privKey, pubKey := k.KeyPair()

	// add both address types for simplicity in tests
	ed25519Address := iotago.Ed25519AddressFromPubKey(pubKey)
	ed25519AddressKey := iotago.NewAddressKeysForEd25519Address(ed25519Address, privKey)
	implicitAccountCreationAddress := iotago.ImplicitAccountCreationAddressFromPubKey(pubKey)
	implicitAccountCreationAddressKey := iotago.NewAddressKeysForImplicitAccountCreationAddress(implicitAccountCreationAddress, privKey)

	return iotago.NewInMemoryAddressSigner(ed25519AddressKey, implicitAccountCreationAddressKey)
}

// Address calculates an address of the specified type.
func (k *KeyManager) Address(addressType iotago.AddressType) iotago.DirectUnlockableAddress {
	_, pubKey := k.KeyPair()

	//nolint:exhaustive // we only support two address types
	switch addressType {
	case iotago.AddressEd25519:
		return iotago.Ed25519AddressFromPubKey(pubKey)
	case iotago.AddressImplicitAccountCreation:
		return iotago.ImplicitAccountCreationAddressFromPubKey(pubKey)
	default:
		panic(ierrors.Wrapf(iotago.ErrUnknownAddrType, "type %d", addressType))
	}
}
