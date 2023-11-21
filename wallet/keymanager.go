package wallet

import (
	"crypto/ed25519"
	"fmt"

	"github.com/wollac/iota-crypto-demo/pkg/bip32path"
	"github.com/wollac/iota-crypto-demo/pkg/slip10"
	"github.com/wollac/iota-crypto-demo/pkg/slip10/eddsa"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	iotago "github.com/iotaledger/iota.go/v4"
)

const (
	defaultPathPrefix = "44'/4218'/0'"
)

// KeyManager is a hierarchical deterministic key manager.
type KeyManager struct {
	seed  []byte
	index uint64
	path  bip32path.Path
}

// NewKeyManager creates a new key manager.
func NewKeyManager(seed []byte, index uint64, pathPrefix ...string) (*KeyManager, error) {
	bip32Path, err := bip32path.ParsePath(fmt.Sprintf("%s/%d'", lo.Cond(len(pathPrefix) > 0, pathPrefix[0], defaultPathPrefix), index))
	if err != nil {
		return nil, ierrors.Wrap(err, "failed to parse path")
	}

	return &KeyManager{
		seed:  seed,
		index: index,
		path:  bip32Path,
	}, nil
}

// KeyPair calculates an ed25519 key pair by using slip10.
func (k *KeyManager) KeyPair() (ed25519.PrivateKey, ed25519.PublicKey) {
	curve := eddsa.Ed25519()
	key, err := slip10.DeriveKeyFromPath(k.seed, curve, k.path)
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
