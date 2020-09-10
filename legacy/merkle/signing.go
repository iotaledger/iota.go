package merkle

import (
	"math/bits"

	"github.com/iotaledger/iota.go/guards"
	"github.com/iotaledger/iota.go/legacy"
	"github.com/iotaledger/iota.go/legacy/kerl"
	"github.com/iotaledger/iota.go/legacy/signing"
	sponge "github.com/iotaledger/iota.go/legacy/signing/utils"
	. "github.com/iotaledger/iota.go/legacy/trinary"
)

// the default SpongeFunction creator
var defaultCreator = func() sponge.SpongeFunction { return kerl.NewKerl() }

// MerkleRoot computes the Merkle tree root from a leaf hash and the given audit path.
// Optionally takes the SpongeFunction to use. Default is Kerl.
func MerkleRoot(leafHash Hash, leafIndex uint32, auditPath []Hash, spongeFunc ...sponge.SpongeFunction) (Hash, error) {
	if len(auditPath) > 32 {
		return "", ErrInvalidAuditPathLength
	}
	if bits.Len32(leafIndex) > len(auditPath) {
		return "", ErrInvalidAuditPathLength
	}

	h := sponge.GetSpongeFunc(spongeFunc, defaultCreator)
	defer h.Reset()

	var (
		j    uint32 = 1
		hash        = leafHash
	)
	for i := range auditPath {
		if (leafIndex & j) != 0 { // if index is a right node, absorb the sibling (left) and then the leafHash
			if err := h.AbsorbTrytes(auditPath[i]); err != nil {
				return "", err
			}
			if err := h.AbsorbTrytes(hash); err != nil {
				return "", err
			}
		} else { // if index is a left node, absorb the leafHash and then the sibling (right)
			if err := h.AbsorbTrytes(hash); err != nil {
				return "", err
			}
			if err := h.AbsorbTrytes(auditPath[i]); err != nil {
				return "", err
			}
		}

		hash = h.MustSqueezeTrytes(legacy.HashTrinarySize)
		h.Reset()

		j *= 2
	}
	return hash, nil
}

// SignatureFragments returns the signed fragments of hashToSign for the given seed and leafIndex.
// Optionally takes the SpongeFunction to use. Default is Kerl.
func SignatureFragments(seed Hash, leafIndex uint32, securityLvl legacy.SecurityLevel, hashToSign Hash, spongeFunc ...sponge.SpongeFunction) ([]Trytes, error) {
	if !guards.IsHash(seed) {
		return nil, legacy.ErrInvalidSeed
	}

	h := sponge.GetSpongeFunc(spongeFunc, defaultCreator)
	defer h.Reset()

	keyTrits, err := computeKey(seed, leafIndex, securityLvl, h)
	if err != nil {
		return nil, err
	}

	// normalize the hash before signing
	normalized := signing.NormalizedBundleHash(hashToSign)

	fragments := make([]Trytes, securityLvl)
	for i := 0; i < int(securityLvl); i++ {
		frag, err := signing.SignatureFragment(
			normalized[i*legacy.HashTrytesSize/3:(i+1)*legacy.HashTrytesSize/3],
			keyTrits[i*legacy.KeyFragmentLength:(i+1)*legacy.KeyFragmentLength],
			h,
		)
		if err != nil {
			return nil, err
		}
		fragments[i] = MustTritsToTrytes(frag)
	}
	return fragments, nil
}

// ValidateSignatureFragments validates the given signature fragments by checking whether the
// address computed from hashToSign and fragments validates the Merkle proof.
// Optionally takes the SpongeFunction to use. Default is Kerl.
func ValidateSignatureFragments(expectedRoot Hash, leafIndex uint32, auditPath []Hash, fragments []Trytes, hashToSign Hash, spongeFunc ...sponge.SpongeFunction) (bool, error) {
	h := sponge.GetSpongeFunc(spongeFunc, defaultCreator)
	defer h.Reset()

	address, err := signing.SignatureAddress(fragments, hashToSign, h)
	if err != nil {
		return false, err
	}

	root, err := MerkleRoot(address, leafIndex, auditPath, h)
	if err != nil {
		return false, err
	}

	return expectedRoot == root, nil
}
