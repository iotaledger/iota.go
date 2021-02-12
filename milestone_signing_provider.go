package iota

import (
	"github.com/iotaledger/iota.go/v2/ed25519"
)

// MilestoneSignerProvider provides milestone signers.
type MilestoneSignerProvider interface {
	// MilestoneIndexSigner returns a new signer for the milestone index.
	MilestoneIndexSigner(msIndex uint32) MilestoneIndexSigner
	// PublicKeysCount returns the amount of public keys in a milestone.
	PublicKeysCount() int
}

// MilestoneIndexSigner is a signer for a particular milestone.
type MilestoneIndexSigner interface {
	// PublicKeys returns a slice of the used public keys.
	PublicKeys() []MilestonePublicKey
	// PublicKeysSet returns a map of the used public keys.
	PublicKeysSet() MilestonePublicKeySet
	// SigningFunc returns a function to sign the particular milestone.
	SigningFunc() MilestoneSigningFunc
}

// InMemoryEd25519MilestoneSignerProvider provides InMemoryEd25519MilestoneIndexSigner.
type InMemoryEd25519MilestoneSignerProvider struct {
	privateKeys     []ed25519.PrivateKey
	keyManger       *MilestoneKeyManager
	publicKeysCount int
}

// NewInMemoryEd25519MilestoneSignerProvider create a new InMemoryEd25519MilestoneSignerProvider.
func NewInMemoryEd25519MilestoneSignerProvider(privateKeys []ed25519.PrivateKey, keyManager *MilestoneKeyManager, publicKeysCount int) *InMemoryEd25519MilestoneSignerProvider {

	return &InMemoryEd25519MilestoneSignerProvider{
		privateKeys:     privateKeys,
		keyManger:       keyManager,
		publicKeysCount: publicKeysCount,
	}
}

// MilestoneIndexSigner returns a new signer for the milestone index.
func (p *InMemoryEd25519MilestoneSignerProvider) MilestoneIndexSigner(index uint32) MilestoneIndexSigner {

	pubKeySet := p.keyManger.GetPublicKeysSetForMilestoneIndex(index)

	keyPairs := p.keyManger.GetMilestonePublicKeyMappingForMilestoneIndex(index, p.privateKeys, p.PublicKeysCount())
	pubKeys := make([]MilestonePublicKey, 0, len(keyPairs))
	for pubKey := range keyPairs {
		pubKeys = append(pubKeys, pubKey)
	}

	milestoneSignFunc := InMemoryEd25519MilestoneSigner(keyPairs)

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
	pubKeys     []MilestonePublicKey
	pubKeySet   MilestonePublicKeySet
	signingFunc MilestoneSigningFunc
}

// PublicKeys returns a slice of the used public keys.
func (s *InMemoryEd25519MilestoneIndexSigner) PublicKeys() []MilestonePublicKey {
	return s.pubKeys
}

// PublicKeysSet returns a map of the used public keys.
func (s *InMemoryEd25519MilestoneIndexSigner) PublicKeysSet() MilestonePublicKeySet {
	return s.pubKeySet
}

// SigningFunc returns a function to sign the particular milestone.
func (s *InMemoryEd25519MilestoneIndexSigner) SigningFunc() MilestoneSigningFunc {
	return s.signingFunc
}
