package iotago_test

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/tpkg"

	"github.com/iotaledger/iota.go/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMilestone__DeSerialize(t *testing.T) {

	milestoneWithoutMetadata := tpkg.RandMilestone(nil)
	milestoneWithoutMetadata.Metadata = nil

	tests := []deSerializeTest{
		{
			name:   "ok",
			source: tpkg.RandMilestone(nil),
			target: &iotago.Milestone{},
		},
		{
			name:   "empty-metadata",
			source: milestoneWithoutMetadata,
			target: &iotago.Milestone{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestMilestone_MarshalUnmarshalJSON(t *testing.T) {
	ms := &iotago.Milestone{
		Index:             1337,
		Timestamp:         13371337,
		Parents:           tpkg.SortedRand32BytArray(2),
		AppliedMerkleRoot: tpkg.Rand32ByteArray(),
		Metadata:          tpkg.RandBytes(10),
		Opts:              iotago.MilestoneOpts{},
		Signatures: iotago.Signatures{
			tpkg.RandEd25519Signature(),
			tpkg.RandEd25519Signature(),
			tpkg.RandEd25519Signature(),
		},
	}

	msJSON, err := json.Marshal(ms)
	require.NoError(t, err)

	desMs := &iotago.Milestone{}
	require.NoError(t, json.Unmarshal(msJSON, desMs))

	require.EqualValues(t, ms, desMs)
}

func TestMilestoneSigning(t *testing.T) {
	type test struct {
		name            string
		ms              *iotago.Milestone
		pubKeys         []iotago.MilestonePublicKey
		signer          iotago.MilestoneSigningFunc
		minSigThreshold int
		pubKeySet       iotago.MilestonePublicKeySet
		signingErr      error
		verificationErr error
	}

	pubKeyFromPrv := func(prvKey ed25519.PrivateKey) iotago.MilestonePublicKey {
		var pubKey iotago.MilestonePublicKey
		copy(pubKey[:], prvKey.Public().(ed25519.PublicKey))
		return pubKey
	}

	tests := []test{
		func() test {
			prvKey := tpkg.RandEd25519PrivateKey()
			pubKey1 := pubKeyFromPrv(prvKey)

			pubKeys := []iotago.MilestonePublicKey{pubKey1}

			msPayload := &iotago.Milestone{
				Parents:           tpkg.SortedRand32BytArray(1 + rand.Intn(7)),
				Index:             1000,
				Timestamp:         uint64(time.Now().Unix()),
				AppliedMerkleRoot: tpkg.Rand32ByteArray(),
			}

			return test{
				name: "ok",
				ms:   msPayload,
				signer: iotago.InMemoryEd25519MilestoneSigner(iotago.MilestonePublicKeyMapping{
					pubKey1: prvKey,
				}),
				minSigThreshold: 1,
				pubKeys:         pubKeys,
				pubKeySet:       map[iotago.MilestonePublicKey]struct{}{pubKey1: {}},
				signingErr:      nil,
				verificationErr: nil,
			}
		}(),
		func() test {

			prvKey1 := tpkg.RandEd25519PrivateKey()
			prvKey2 := tpkg.RandEd25519PrivateKey()
			prvKey3 := tpkg.RandEd25519PrivateKey()
			pubKey1 := pubKeyFromPrv(prvKey1)
			pubKey2 := pubKeyFromPrv(prvKey2)
			pubKey3 := pubKeyFromPrv(prvKey3)

			// only 1 and 2
			pubKeys := serializer.LexicalOrdered32ByteArrays{pubKey1, pubKey2}
			sort.Sort(pubKeys)

			msPayload := &iotago.Milestone{
				Parents:           tpkg.SortedRand32BytArray(1 + rand.Intn(7)),
				Index:             1000,
				Timestamp:         uint64(time.Now().Unix()),
				AppliedMerkleRoot: tpkg.Rand32ByteArray(),
			}

			return test{
				name: "ok - 2 of 3 from applicable set",
				ms:   msPayload,
				signer: iotago.InMemoryEd25519MilestoneSigner(iotago.MilestonePublicKeyMapping{
					pubKey1: prvKey1,
					pubKey2: prvKey2,
					pubKey3: prvKey3,
				}),
				minSigThreshold: 2,
				pubKeys:         pubKeys,
				pubKeySet:       map[iotago.MilestonePublicKey]struct{}{pubKey1: {}, pubKey2: {}, pubKey3: {}},
				signingErr:      nil,
				verificationErr: nil,
			}
		}(),
		func() test {
			prvKey := tpkg.RandEd25519PrivateKey()
			pubKey1 := pubKeyFromPrv(prvKey)

			pubKeys := []iotago.MilestonePublicKey{pubKey1}

			msPayload := &iotago.Milestone{
				Parents:           tpkg.SortedRand32BytArray(1 + rand.Intn(7)),
				Index:             1000,
				Timestamp:         uint64(time.Now().Unix()),
				AppliedMerkleRoot: tpkg.Rand32ByteArray(),
			}

			return test{
				name: "err - invalid signature",
				ms:   msPayload,
				signer: iotago.InMemoryEd25519MilestoneSigner(iotago.MilestonePublicKeyMapping{
					// signature will be signed with a non matching private key
					pubKey1: tpkg.RandEd25519PrivateKey(),
				}),
				minSigThreshold: 1,
				pubKeys:         pubKeys,
				pubKeySet:       map[iotago.MilestonePublicKey]struct{}{pubKey1: {}},
				signingErr:      nil,
				verificationErr: iotago.ErrMilestoneInvalidSignature,
			}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ms.Sign(tt.pubKeys, tt.signer)
			if tt.signingErr != nil {
				assert.True(t, errors.Is(err, tt.signingErr))
				return
			}
			assert.NoError(t, err)

			err = tt.ms.VerifySignatures(tt.minSigThreshold, tt.pubKeySet)
			if tt.verificationErr != nil {
				assert.True(t, errors.Is(err, tt.verificationErr))
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestNewMilestone(t *testing.T) {
	parents := tpkg.SortedRand32BytArray(1 + rand.Intn(7))
	lastMs := tpkg.Rand32ByteArray()
	pastConeMerkleProof := tpkg.Rand32ByteArray()
	inclusionMerkleProof := tpkg.Rand32ByteArray()
	const msIndex, timestamp = 1000, 133713371337

	ms, err := iotago.NewMilestone(msIndex, timestamp, lastMs, parents, pastConeMerkleProof, inclusionMerkleProof)
	assert.NoError(t, err)

	assert.EqualValues(t, &iotago.Milestone{
		Index:               msIndex,
		Timestamp:           timestamp,
		LastMilestone:       lastMs,
		Parents:             parents,
		ConfirmedMerkleRoot: pastConeMerkleProof,
		AppliedMerkleRoot:   inclusionMerkleProof,
		Signatures:          nil,
	}, ms)
}
