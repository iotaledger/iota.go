package wallet

import (
	"crypto/ed25519"

	"github.com/wollac/iota-crypto-demo/pkg/bip32path"
	"github.com/wollac/iota-crypto-demo/pkg/bip39"
	"github.com/wollac/iota-crypto-demo/pkg/slip10"
	"github.com/wollac/iota-crypto-demo/pkg/slip10/eddsa"

	"github.com/iotaledger/hive.go/ierrors"
	iotago "github.com/iotaledger/iota.go/v4"
)

const (
	DefaultIOTAPath    = "m/44'/4218'/0'/0'/0'"
	DefaultShimmerPath = "m/44'/4219'/0'/0'/0'"
)

// KeyManager is a hierarchical deterministic key manager.
// NOTE: The seed is stored in memory and is not protected against memory dumps.
type KeyManager struct {
	seed []byte
	path bip32path.Path
}

// NewKeyManagerFromRandom creates a new key manager from random entropy.
func NewKeyManagerFromRandom(path string) (*KeyManager, error) {
	// Generate random entropy by using ed25519 key generation and using the private key seed (32 bytes)
	_, random, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, ierrors.Wrap(err, "failed to generate random entropy")
	}

	return NewKeyManager(random.Seed(), path)
}

// NewKeyManagerFromMnemonic creates a new key manager from a mnemonic.
func NewKeyManagerFromMnemonic(mnemonic string, path string) (*KeyManager, error) {
	mnemonicSentence := bip39.ParseMnemonic(mnemonic)
	if len(mnemonicSentence) != 24 {
		return nil, ierrors.Errorf("mnemomic contains an invalid sentence length. Mnemonic should be 24 words")
	}

	seed, err := bip39.MnemonicToSeed(mnemonicSentence, "")
	if err != nil {
		return nil, ierrors.Wrap(err, "failed to convert mnemonic to seed")
	}

	return NewKeyManager(seed, path)
}

// NewKeyManager creates a new key manager.
func NewKeyManager(seed []byte, path string) (*KeyManager, error) {
	bip32Path, err := bip32path.ParsePath(path)
	if err != nil {
		return nil, ierrors.Wrap(err, "failed to parse path")
	}

	return &KeyManager{
		seed: seed,
		path: bip32Path,
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

func (k *KeyManager) Path() bip32path.Path {
	return k.path
}

// Mnemonic returns the mnemonic of the key manager.
func (k *KeyManager) Mnemonic() bip39.Mnemonic {
	mnemonic, err := bip39.EntropyToMnemonic(k.seed)
	if err != nil {
		panic(ierrors.Wrap(err, "failed to convert seed to mnemonic"))
	}

	return mnemonic
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
