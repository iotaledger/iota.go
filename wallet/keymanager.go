package wallet

import (
	"crypto/ed25519"
	"fmt"

	"github.com/iotaledger/iota-crypto-demo/pkg/bip32path"
	"github.com/iotaledger/iota-crypto-demo/pkg/bip39"
	"github.com/iotaledger/iota-crypto-demo/pkg/slip10"
	"github.com/iotaledger/iota-crypto-demo/pkg/slip10/eddsa"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
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
func (k *KeyManager) KeyPair(index ...uint32) (ed25519.PrivateKey, ed25519.PublicKey) {
	curve := eddsa.Ed25519()
	key, err := slip10.DeriveKeyFromPath(k.seed, curve, k.Path(index...))
	if err != nil {
		panic(err)
	}

	pubKey, privKey := key.Key.(eddsa.Seed).Ed25519Key()

	return ed25519.PrivateKey(privKey), ed25519.PublicKey(pubKey)
}

func (k *KeyManager) Path(index ...uint32) bip32path.Path {
	if len(index) == 0 {
		// no additional index given, use the internal path
		return k.path
	}

	// new index given, check if the internal path contains the index part
	if len(k.path) < 5 {
		panic("invalid path length")
	}

	// copy the former path
	newPath := lo.CopySlice(k.path)

	// set the new index
	newPath[4] = index[0] | (1 << 31)

	return newPath
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
func (k *KeyManager) AddressSigner(indexes ...uint32) iotago.AddressSigner {
	privKeys := make([]ed25519.PrivateKey, 0)

	if len(indexes) == 0 {
		privKey, _ := k.KeyPair()
		privKeys = append(privKeys, privKey)
	} else {
		for _, index := range indexes {
			privKey, _ := k.KeyPair(index)
			privKeys = append(privKeys, privKey)
		}
	}

	return iotago.NewInMemoryAddressSignerFromEd25519PrivateKeys(privKeys...)
}

// Address calculates an address of the specified type.
func (k *KeyManager) Address(addressType iotago.AddressType, index ...uint32) iotago.DirectUnlockableAddress {
	_, pubKey := k.KeyPair(index...)

	//nolint:exhaustive // we only support two address types
	switch addressType {
	case iotago.AddressEd25519:
		return iotago.Ed25519AddressFromPubKey(pubKey)
	case iotago.AddressImplicitAccountCreation:
		return iotago.ImplicitAccountCreationAddressFromPubKey(pubKey)
	default:
		panic(fmt.Sprintf("address type %s is not supported", addressType))
	}
}
