package iota_test

import (
	"encoding/json"
	"errors"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/iotaledger/iota.go/v2"
	"github.com/iotaledger/iota.go/v2/ed25519"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMilestone_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target *iota.Milestone
		err    error
	}
	tests := []test{
		func() test {
			msPayload, msPayloadData := randMilestone(nil)
			return test{"ok", msPayloadData, msPayload, nil}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msPayload := &iota.Milestone{}
			bytesRead, err := msPayload.Deserialize(tt.source, iota.DeSeriModePerformValidation)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, len(tt.source), bytesRead)
			assert.EqualValues(t, tt.target, msPayload)
		})
	}
}

func TestMilestone_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iota.Milestone
		target []byte
	}
	tests := []test{
		func() test {
			msPayload, msPayloadData := randMilestone(nil)
			return test{"ok", msPayload, msPayloadData}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edData, err := tt.source.Serialize(iota.DeSeriModePerformValidation)
			assert.NoError(t, err)
			assert.Equal(t, tt.target, edData)
		})
	}
}

func TestMilestone_MarshalUnmarshalJSON(t *testing.T) {
	ms := &iota.Milestone{
		Index:                1337,
		Timestamp:            13371337,
		Parents:              sortedRand32ByteHashes(2),
		InclusionMerkleProof: rand32ByteHash(),
		PublicKeys:           sortedRand32ByteHashes(3),
		Signatures: []iota.MilestoneSignature{
			rand64ByteHash(),
			rand64ByteHash(),
			rand64ByteHash(),
		},
	}

	msJSON, err := json.Marshal(ms)
	require.NoError(t, err)

	desMs := &iota.Milestone{}
	require.NoError(t, json.Unmarshal(msJSON, desMs))

	require.EqualValues(t, ms, desMs)
}

func TestMilestoneSigning(t *testing.T) {
	type test struct {
		name            string
		ms              *iota.Milestone
		signer          iota.MilestoneSigningFunc
		minSigThreshold int
		pubKeySet       iota.MilestonePublicKeySet
		signingErr      error
		verificationErr error
	}

	pubKeyFromPrv := func(prvKey ed25519.PrivateKey) iota.MilestonePublicKey {
		var pubKey iota.MilestonePublicKey
		copy(pubKey[:], prvKey.Public().(ed25519.PublicKey))
		return pubKey
	}

	tests := []test{
		func() test {
			prvKey := randEd25519PrivateKey()
			pubKey1 := pubKeyFromPrv(prvKey)

			pubKeys := []iota.MilestonePublicKey{pubKey1}

			msPayload := &iota.Milestone{
				Parents:              sortedRand32ByteHashes(1 + rand.Intn(7)),
				Index:                1000,
				Timestamp:            uint64(time.Now().Unix()),
				PublicKeys:           pubKeys,
				InclusionMerkleProof: rand32ByteHash(),
			}

			return test{
				name: "ok",
				ms:   msPayload,
				signer: iota.InMemoryEd25519MilestoneSigner(iota.MilestonePublicKeyMapping{
					pubKey1: prvKey,
				}),
				minSigThreshold: 1,
				pubKeySet:       map[iota.MilestonePublicKey]struct{}{pubKey1: {}},
				signingErr:      nil,
				verificationErr: nil,
			}
		}(),
		func() test {

			prvKey1 := randEd25519PrivateKey()
			prvKey2 := randEd25519PrivateKey()
			prvKey3 := randEd25519PrivateKey()
			pubKey1 := pubKeyFromPrv(prvKey1)
			pubKey2 := pubKeyFromPrv(prvKey2)
			pubKey3 := pubKeyFromPrv(prvKey3)

			// only 1 and 2
			pubKeys := iota.LexicalOrdered32ByteArrays{pubKey1, pubKey2}
			sort.Sort(pubKeys)

			msPayload := &iota.Milestone{
				Parents:              sortedRand32ByteHashes(1 + rand.Intn(7)),
				Index:                1000,
				Timestamp:            uint64(time.Now().Unix()),
				PublicKeys:           pubKeys,
				InclusionMerkleProof: rand32ByteHash(),
			}

			return test{
				name: "ok - 2 of 3 from applicable set",
				ms:   msPayload,
				signer: iota.InMemoryEd25519MilestoneSigner(iota.MilestonePublicKeyMapping{
					pubKey1: prvKey1,
					pubKey2: prvKey2,
					pubKey3: prvKey3,
				}),
				minSigThreshold: 2,
				pubKeySet:       map[iota.MilestonePublicKey]struct{}{pubKey1: {}, pubKey2: {}, pubKey3: {}},
				signingErr:      nil,
				verificationErr: nil,
			}
		}(),
		func() test {
			prvKey := randEd25519PrivateKey()
			pubKey1 := pubKeyFromPrv(prvKey)

			pubKeys := []iota.MilestonePublicKey{pubKey1}

			msPayload := &iota.Milestone{
				Parents:              sortedRand32ByteHashes(1 + rand.Intn(7)),
				Index:                1000,
				Timestamp:            uint64(time.Now().Unix()),
				PublicKeys:           pubKeys,
				InclusionMerkleProof: rand32ByteHash(),
			}

			return test{
				name: "err - invalid signature",
				ms:   msPayload,
				signer: iota.InMemoryEd25519MilestoneSigner(iota.MilestonePublicKeyMapping{
					// signature will be signed with a non matching private key
					pubKey1: randEd25519PrivateKey(),
				}),
				minSigThreshold: 1,
				pubKeySet:       map[iota.MilestonePublicKey]struct{}{pubKey1: {}},
				signingErr:      nil,
				verificationErr: iota.ErrMilestoneInvalidSignature,
			}
		}(),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.ms.Sign(test.signer)
			if test.signingErr != nil {
				assert.True(t, errors.Is(err, test.signingErr))
				return
			}
			assert.NoError(t, err)

			err = test.ms.VerifySignatures(test.minSigThreshold, test.pubKeySet)
			if test.verificationErr != nil {
				assert.True(t, errors.Is(err, test.verificationErr))
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestNewMilestone(t *testing.T) {
	parents := sortedRand32ByteHashes(1 + rand.Intn(7))
	inclusionMerkleProof := rand32ByteHash()
	const msIndex, timestamp = 1000, 133713371337
	unsortedPubKeys := []iota.MilestonePublicKey{{3}, {2}, {1}, {5}}

	ms, err := iota.NewMilestone(msIndex, timestamp, parents, inclusionMerkleProof, unsortedPubKeys)
	assert.NoError(t, err)

	assert.EqualValues(t, &iota.Milestone{
		Index:                msIndex,
		Timestamp:            timestamp,
		Parents:              parents,
		InclusionMerkleProof: inclusionMerkleProof,
		PublicKeys:           []iota.MilestonePublicKey{{1}, {2}, {3}, {5}},
		Signatures:           nil,
	}, ms)
}
