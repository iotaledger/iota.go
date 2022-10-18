//#nosec G404

package signingprovider_test

import (
	"crypto/ed25519"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/keymanager"
	"github.com/iotaledger/iota.go/v3/signingprovider"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

func TestInMemoryEd25519MilestoneSignerProvider(t *testing.T) {

	privateKeys := make([]ed25519.PrivateKey, 0)

	pubKey1, privKey1, err := ed25519.GenerateKey(nil)
	assert.NoError(t, err)
	privateKeys = append(privateKeys, privKey1)

	pubKey2, privKey2, err := ed25519.GenerateKey(nil)
	assert.NoError(t, err)
	privateKeys = append(privateKeys, privKey2)

	pubKey3, privKey3, err := ed25519.GenerateKey(nil)
	assert.NoError(t, err)
	privateKeys = append(privateKeys, privKey3)

	pubKey4, privKey4, err := ed25519.GenerateKey(nil)
	assert.NoError(t, err)
	privateKeys = append(privateKeys, privKey4)

	var msPubKey1 iotago.MilestonePublicKey
	copy(msPubKey1[:], pubKey1)

	var msPubKey2 iotago.MilestonePublicKey
	copy(msPubKey2[:], pubKey2)

	var msPubKey3 iotago.MilestonePublicKey
	copy(msPubKey3[:], pubKey3)

	var msPubKey4 iotago.MilestonePublicKey
	copy(msPubKey4[:], pubKey4)

	km := keymanager.New()
	km.AddKeyRange(pubKey1, 0, 9)
	km.AddKeyRange(pubKey2, 3, 10)
	km.AddKeyRange(pubKey3, 8, 15)
	km.AddKeyRange(pubKey4, 17, 19)

	signer := signingprovider.NewInMemoryEd25519MilestoneSignerProvider(privateKeys, km, 2)
	require.Equal(t, 2, signer.PublicKeysCount())

	type test struct {
		name            string
		ms              *iotago.Milestone
		signingErr      error
		verificationErr error
	}

	tests := []test{
		func() test {
			msPayload := &iotago.Milestone{
				MilestoneEssence: iotago.MilestoneEssence{
					Parents:           tpkg.SortedRandMSParents(1 + rand.Intn(7)),
					Index:             4,
					Timestamp:         uint32(time.Now().Unix()),
					AppliedMerkleRoot: tpkg.Rand32ByteArray(),
				},
			}

			return test{
				name:            "ok - 2 of 2 from applicable set",
				ms:              msPayload,
				signingErr:      nil,
				verificationErr: nil,
			}
		}(),
		func() test {
			msPayload := &iotago.Milestone{
				MilestoneEssence: iotago.MilestoneEssence{
					Parents:           tpkg.SortedRandMSParents(1 + rand.Intn(7)),
					Index:             8,
					Timestamp:         uint32(time.Now().Unix()),
					AppliedMerkleRoot: tpkg.Rand32ByteArray(),
				},
			}

			return test{
				name:            "ok - 2 of 3 from applicable set",
				ms:              msPayload,
				signingErr:      nil,
				verificationErr: nil,
			}
		}(),
		func() test {
			msPayload := &iotago.Milestone{
				MilestoneEssence: iotago.MilestoneEssence{
					Parents:           tpkg.SortedRandMSParents(1 + rand.Intn(7)),
					Index:             20,
					Timestamp:         uint32(time.Now().Unix()),
					AppliedMerkleRoot: tpkg.Rand32ByteArray(),
				},
			}

			return test{
				name:            "err - too few signatures for signing",
				ms:              msPayload,
				signingErr:      iotago.ErrMilestoneTooFewSignatures,
				verificationErr: nil,
			}
		}(),
		func() test {
			msPayload := &iotago.Milestone{
				MilestoneEssence: iotago.MilestoneEssence{
					Parents:           tpkg.SortedRandMSParents(1 + rand.Intn(7)),
					Index:             17,
					Timestamp:         uint32(time.Now().Unix()),
					AppliedMerkleRoot: tpkg.Rand32ByteArray(),
				},
			}

			return test{
				name:            "err - too few signatures for verification",
				ms:              msPayload,
				signingErr:      nil,
				verificationErr: iotago.ErrMilestoneTooFewSignaturesForVerificationThreshold,
			}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			indexSigner := signer.MilestoneIndexSigner(tt.ms.Index)

			err := tt.ms.Sign(indexSigner.PublicKeys(), indexSigner.SigningFunc())
			if tt.signingErr != nil {
				assert.True(t, errors.Is(err, tt.signingErr))
				return
			}
			assert.NoError(t, err)

			err = tt.ms.VerifySignatures(signer.PublicKeysCount(), indexSigner.PublicKeysSet())
			if tt.verificationErr != nil {
				println(err.Error())
				assert.True(t, errors.Is(err, tt.verificationErr))
				return
			}
			assert.NoError(t, err)
		})
	}
}
