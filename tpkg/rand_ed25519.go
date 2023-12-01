package tpkg

import (
	"bytes"
	"crypto/ed25519"
	cryptorand "crypto/rand"
	"fmt"
	"slices"

	hiveEd25519 "github.com/iotaledger/hive.go/crypto/ed25519"
	iotago "github.com/iotaledger/iota.go/v4"
)

// RandEd25519Signature returns a random Ed25519 signature.
func RandEd25519Signature() *iotago.Ed25519Signature {
	edSig := &iotago.Ed25519Signature{}
	pub := RandBytes(ed25519.PublicKeySize)
	sig := RandBytes(ed25519.SignatureSize)
	copy(edSig.PublicKey[:], pub)
	copy(edSig.Signature[:], sig)

	return edSig
}

// RandEd25519PrivateKey returns a random Ed25519 private key.
func RandEd25519PrivateKey() ed25519.PrivateKey {
	seed := RandEd25519Seed()

	return ed25519.NewKeyFromSeed(seed[:])
}

func RandEd25519PublicKey() hiveEd25519.PublicKey {
	//nolint:forcetypeassert // we can safely assume that this is an ed25519.PublicKey
	return hiveEd25519.PublicKey(RandEd25519PrivateKey().Public().(ed25519.PublicKey))
}

// RandEd25519Seed returns a random Ed25519 seed.
func RandEd25519Seed() [ed25519.SeedSize]byte {
	var b [ed25519.SeedSize]byte
	read, err := cryptorand.Read(b[:])
	if read != ed25519.SeedSize {
		panic(fmt.Sprintf("could not read %d required bytes from secure RNG", ed25519.SeedSize))
	}
	if err != nil {
		panic(err)
	}

	return b
}

// RandEd25519Identity produces a random Ed25519 identity.
func RandEd25519Identity() (ed25519.PrivateKey, *iotago.Ed25519Address, iotago.AddressKeys) {
	edSk := RandEd25519PrivateKey()
	//nolint:forcetypeassert // we can safely assume that this is an ed25519.PublicKey
	edAddr := iotago.Ed25519AddressFromPubKey(edSk.Public().(ed25519.PublicKey))
	addrKeys := iotago.NewAddressKeysForEd25519Address(edAddr, edSk)

	return edSk, edAddr, addrKeys
}

// RandEd25519IdentitiesSortedByAddress returns random Ed25519 identities and keys lexically sorted by the address.
func RandEd25519IdentitiesSortedByAddress(count int) ([]iotago.Address, []iotago.AddressKeys) {
	addresses := make([]iotago.Address, count)
	addressKeys := make([]iotago.AddressKeys, count)
	for i := 0; i < count; i++ {
		_, addresses[i], addressKeys[i] = RandEd25519Identity()
	}

	// addressses need to be lexically ordered in the MultiAddress
	slices.SortFunc(addresses, func(a iotago.Address, b iotago.Address) int {
		return bytes.Compare(a.ID(), b.ID())
	})

	// addressses need to be lexically ordered in the MultiAddress
	slices.SortFunc(addressKeys, func(a iotago.AddressKeys, b iotago.AddressKeys) int {
		return bytes.Compare(a.Address.ID(), b.Address.ID())
	})

	return addresses, addressKeys
}

// RandImplicitAccountIdentity produces a random Implicit Account identity.
func RandImplicitAccountIdentity() (ed25519.PrivateKey, *iotago.ImplicitAccountCreationAddress, iotago.AddressKeys) {
	edSk := RandEd25519PrivateKey()
	//nolint:forcetypeassert // we can safely assume that this is an ed25519.PublicKey
	implicitAccAddr := iotago.ImplicitAccountCreationAddressFromPubKey(edSk.Public().(ed25519.PublicKey))
	addrKeys := iotago.NewAddressKeysForImplicitAccountCreationAddress(implicitAccAddr, edSk)

	return edSk, implicitAccAddr, addrKeys
}

func RandBlockIssuerKey() iotago.BlockIssuerKey {
	return iotago.Ed25519PublicKeyBlockIssuerKeyFromPublicKey(RandEd25519PublicKey())
}

func RandBlockIssuerKeys(count ...int) iotago.BlockIssuerKeys {
	// We always generate at least one key.
	length := RandInt(10) + 1

	if len(count) > 0 {
		length = count[0]
	}

	blockIssuerKeys := iotago.NewBlockIssuerKeys()
	for i := 0; i < length; i++ {
		blockIssuerKeys.Add(RandBlockIssuerKey())
	}
	blockIssuerKeys.Sort()

	return blockIssuerKeys
}
