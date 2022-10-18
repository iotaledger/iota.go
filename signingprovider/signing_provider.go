package signingprovider

import (
	"crypto/ed25519"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/keymanager"
)

// MilestoneSignerProvider provides milestone signers.
type MilestoneSignerProvider interface {
	// MilestoneIndexSigner returns a new signer for the milestone index.
	MilestoneIndexSigner(index iotago.MilestoneIndex) MilestoneIndexSigner
	// PublicKeysCount returns the amount of public keys in a milestone.
	PublicKeysCount() int
}

// MilestoneIndexSigner is a signer for a particular milestone.
type MilestoneIndexSigner interface {
	// PublicKeys returns a slice of the used public keys.
	PublicKeys() []iotago.MilestonePublicKey
	// PublicKeysSet returns a map of the used public keys.
	PublicKeysSet() iotago.MilestonePublicKeySet
	// SigningFunc returns a function to sign the particular milestone.
	SigningFunc() iotago.MilestoneSigningFunc
}

// InMemoryEd25519MilestoneSignerProvider provides InMemoryEd25519MilestoneIndexSigner.
type InMemoryEd25519MilestoneSignerProvider struct {
	privateKeys     []ed25519.PrivateKey
	keyManger       *keymanager.KeyManager
	publicKeysCount int
}

// NewInMemoryEd25519MilestoneSignerProvider creates a new InMemoryEd25519MilestoneSignerProvider.
func NewInMemoryEd25519MilestoneSignerProvider(privateKeys []ed25519.PrivateKey, keyManager *keymanager.KeyManager, publicKeysCount int) *InMemoryEd25519MilestoneSignerProvider {

	return &InMemoryEd25519MilestoneSignerProvider{
		privateKeys:     privateKeys,
		keyManger:       keyManager,
		publicKeysCount: publicKeysCount,
	}
}

// MilestoneIndexSigner returns a new signer for the milestone index.
func (p *InMemoryEd25519MilestoneSignerProvider) MilestoneIndexSigner(index iotago.MilestoneIndex) MilestoneIndexSigner {

	pubKeySet := p.keyManger.PublicKeysSetForMilestoneIndex(index)

	keyPairs := p.keyManger.MilestonePublicKeyMappingForMilestoneIndex(index, p.privateKeys, p.PublicKeysCount())
	pubKeys := make([]iotago.MilestonePublicKey, 0, len(keyPairs))
	for pubKey := range keyPairs {
		pubKeys = append(pubKeys, pubKey)
	}

	milestoneSignFunc := iotago.InMemoryEd25519MilestoneSigner(keyPairs)

	return &InMemoryEd25519MilestoneIndexSigner{
		pubKeys:     pubKeys,
		pubKeySet:   pubKeySet,
		signingFunc: milestoneSignFunc,
	}
}

// PublicKeysCount returns the amount of public keys in a milestone.
func (p *InMemoryEd25519MilestoneSignerProvider) PublicKeysCount() int {
	return p.publicKeysCount
}

// InMemoryEd25519MilestoneIndexSigner is an in memory signer for a particular milestone.
type InMemoryEd25519MilestoneIndexSigner struct {
	pubKeys     []iotago.MilestonePublicKey
	pubKeySet   iotago.MilestonePublicKeySet
	signingFunc iotago.MilestoneSigningFunc
}

// PublicKeys returns a slice of the used public keys.
func (s *InMemoryEd25519MilestoneIndexSigner) PublicKeys() []iotago.MilestonePublicKey {
	return s.pubKeys
}

// PublicKeysSet returns a map of the used public keys.
func (s *InMemoryEd25519MilestoneIndexSigner) PublicKeysSet() iotago.MilestonePublicKeySet {
	return s.pubKeySet
}

// SigningFunc returns a function to sign the particular milestone.
func (s *InMemoryEd25519MilestoneIndexSigner) SigningFunc() iotago.MilestoneSigningFunc {
	return s.signingFunc
}

// InsecureRemoteEd25519MilestoneSignerProvider provides InsecureRemoteEd25519MilestoneIndexSigner.
type InsecureRemoteEd25519MilestoneSignerProvider struct {
	signingFunc     iotago.MilestoneSigningFunc
	keyManger       *keymanager.KeyManager
	publicKeysCount int
}

// NewInsecureRemoteEd25519MilestoneSignerProvider creates a new InsecureRemoteEd25519MilestoneSignerProvider.
func NewInsecureRemoteEd25519MilestoneSignerProvider(remoteEndpoint string, keyManager *keymanager.KeyManager, publicKeysCount int) *InsecureRemoteEd25519MilestoneSignerProvider {

	return &InsecureRemoteEd25519MilestoneSignerProvider{
		signingFunc:     iotago.InsecureRemoteEd25519MilestoneSigner(remoteEndpoint),
		keyManger:       keyManager,
		publicKeysCount: publicKeysCount,
	}
}

// MilestoneIndexSigner returns a new signer for the milestone index.
func (p *InsecureRemoteEd25519MilestoneSignerProvider) MilestoneIndexSigner(index iotago.MilestoneIndex) MilestoneIndexSigner {

	return &InsecureRemoteEd25519MilestoneIndexSigner{
		pubKeys:     p.keyManger.PublicKeysForMilestoneIndex(index),
		pubKeySet:   p.keyManger.PublicKeysSetForMilestoneIndex(index),
		signingFunc: p.signingFunc,
	}
}

// PublicKeysCount returns the amount of public keys in a milestone.
func (p *InsecureRemoteEd25519MilestoneSignerProvider) PublicKeysCount() int {
	return p.publicKeysCount
}

// InsecureRemoteEd25519MilestoneIndexSigner is an in memory signer for a particular milestone.
type InsecureRemoteEd25519MilestoneIndexSigner struct {
	pubKeys     []iotago.MilestonePublicKey
	pubKeySet   iotago.MilestonePublicKeySet
	signingFunc iotago.MilestoneSigningFunc
}

// PublicKeys returns a slice of the used public keys.
func (s *InsecureRemoteEd25519MilestoneIndexSigner) PublicKeys() []iotago.MilestonePublicKey {
	return s.pubKeys
}

// PublicKeysSet returns a map of the used public keys.
func (s *InsecureRemoteEd25519MilestoneIndexSigner) PublicKeysSet() iotago.MilestonePublicKeySet {
	return s.pubKeySet
}

// SigningFunc returns a function to sign the particular milestone.
func (s *InsecureRemoteEd25519MilestoneIndexSigner) SigningFunc() iotago.MilestoneSigningFunc {
	return s.signingFunc
}
