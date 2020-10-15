package iota_test

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"errors"
	"testing"
	"time"

	"github.com/iotaledger/iota.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMilestone_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target iota.Serializable
		err    error
	}
	tests := []test{
		func() test {
			msPayload, msPayloadData := randMilestone()
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
			msPayload, msPayloadData := randMilestone()
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

func TestMilestone_SignatureInput(t *testing.T) {
	var msgVersion byte = 1
	parent1 := rand32ByteHash()
	parent2 := rand32ByteHash()
	var msIndex, msTimestamp uint64 = 1000, 133713771377
	var msInclMerkleProof [iota.MilestoneInclusionMerkleProofLength]byte
	copy(msInclMerkleProof[:], randBytes(iota.MilestoneInclusionMerkleProofLength))
	msSigs := [][iota.MilestoneSignatureLength]byte{
		randMilestoneSig(), randMilestoneSig(), randMilestoneSig(),
	}

	msg := &iota.Message{Version: msgVersion, Parent1: parent1, Parent2: parent2}
	msPayload := &iota.Milestone{
		Index: msIndex, Timestamp: msTimestamp,
		InclusionMerkleProof: msInclMerkleProof, Signatures: msSigs,
	}
	msPayloadBytes, err := msPayload.Serialize(iota.DeSeriModeNoValidation)
	require.NoError(t, err)

	sigInput, err := msPayload.SignatureInput(msg, 3)
	assert.NoError(t, err)

	var offset int
	assert.EqualValues(t, msgVersion, sigInput[0])
	offset += iota.OneByte
	assert.True(t, bytes.Equal(parent1[:], sigInput[offset:offset+iota.MessageIDLength]))
	offset += iota.MessageIDLength
	assert.True(t, bytes.Equal(parent2[:], sigInput[offset:offset+iota.MessageIDLength]))
	offset += iota.MessageIDLength
	assert.EqualValues(t, len(msPayloadBytes), binary.LittleEndian.Uint32(sigInput[offset:offset+iota.PayloadLengthByteSize]))
	offset += iota.PayloadLengthByteSize
	assert.EqualValues(t, iota.MilestonePayloadTypeID, binary.LittleEndian.Uint32(sigInput[offset:offset+iota.TypeDenotationByteSize]))
	offset += iota.TypeDenotationByteSize
	assert.EqualValues(t, msIndex, binary.LittleEndian.Uint64(sigInput[offset:offset+iota.UInt64ByteSize]))
	offset += iota.UInt64ByteSize
	assert.EqualValues(t, msTimestamp, binary.LittleEndian.Uint64(sigInput[offset:offset+iota.UInt64ByteSize]))
	offset += iota.UInt64ByteSize
	assert.True(t, bytes.Equal(msInclMerkleProof[:], sigInput[offset:offset+iota.MilestoneInclusionMerkleProofLength]))
	offset += iota.MilestoneInclusionMerkleProofLength
	assert.EqualValues(t, len(msSigs), sigInput[offset : offset+iota.OneByte][0])
	offset += iota.OneByte
	assert.Len(t, sigInput, offset)
}

func TestMilestoneSigning(t *testing.T) {
	type test struct {
		name                 string
		msg                  *iota.Message
		ms                   *iota.Milestone
		signer               iota.MilestoneSigningFunc
		sigsCount            byte
		minValid             int
		pubKeys              []ed25519.PublicKey
		pubKeysWhichVerified []ed25519.PublicKey
		signErr              error
		verfErr              error
	}

	tests := []test{
		func() test {

			signer := randEd25519PrivateKey()
			msg := &iota.Message{Version: 1, Parent1: rand32ByteHash(), Parent2: rand32ByteHash()}
			msPayload := &iota.Milestone{Index: 1000, Timestamp: uint64(time.Now().Unix())}

			return test{
				name:      "ok",
				msg:       msg,
				ms:        msPayload,
				signer:    iota.InMemoryEd25519MilestoneSigner(signer),
				sigsCount: 1,
				minValid:  1,
				pubKeys: []ed25519.PublicKey{
					signer.Public().(ed25519.PublicKey),
				},
				pubKeysWhichVerified: []ed25519.PublicKey{
					signer.Public().(ed25519.PublicKey),
				},
				signErr: nil,
				verfErr: nil,
			}
		}(),
		func() test {

			signer1 := randEd25519PrivateKey()
			signer2 := randEd25519PrivateKey()
			signer3 := randEd25519PrivateKey()
			msg := &iota.Message{Version: 1, Parent1: rand32ByteHash(), Parent2: rand32ByteHash()}
			msPayload := &iota.Milestone{Index: 1000, Timestamp: uint64(time.Now().Unix())}

			return test{
				name:      "ok - 2 of 3",
				msg:       msg,
				ms:        msPayload,
				signer:    iota.InMemoryEd25519MilestoneSigner(signer1, signer2, signer3),
				sigsCount: 3,
				minValid:  2,
				pubKeys: []ed25519.PublicKey{
					signer1.Public().(ed25519.PublicKey),
					signer2.Public().(ed25519.PublicKey),
					randEd25519PrivateKey().Public().(ed25519.PublicKey),
				},
				pubKeysWhichVerified: []ed25519.PublicKey{
					signer1.Public().(ed25519.PublicKey),
					signer2.Public().(ed25519.PublicKey),
				},
				signErr: nil,
				verfErr: nil,
			}
		}(),
		func() test {

			signer := randEd25519PrivateKey()
			msg := &iota.Message{Version: 1, Parent1: rand32ByteHash(), Parent2: rand32ByteHash()}
			msPayload := &iota.Milestone{Index: 1000, Timestamp: uint64(time.Now().Unix())}

			return test{
				name:      "err - invalid signature",
				msg:       msg,
				ms:        msPayload,
				signer:    iota.InMemoryEd25519MilestoneSigner(signer),
				sigsCount: 1,
				minValid:  1,
				pubKeys: []ed25519.PublicKey{
					randEd25519PrivateKey().Public().(ed25519.PublicKey),
				},
				signErr: nil,
				verfErr: iota.ErrMilestoneMinSigsThresholdNotReached,
			}
		}(),
		func() test {

			signer1 := randEd25519PrivateKey()
			signer2 := randEd25519PrivateKey()
			signer3 := randEd25519PrivateKey()
			msg := &iota.Message{Version: 1, Parent1: rand32ByteHash(), Parent2: rand32ByteHash()}
			msPayload := &iota.Milestone{Index: 1000, Timestamp: uint64(time.Now().Unix())}

			return test{
				name:      "err - 1 of wanted min. 2",
				msg:       msg,
				ms:        msPayload,
				signer:    iota.InMemoryEd25519MilestoneSigner(signer1, signer2, signer3),
				sigsCount: 3,
				minValid:  2,
				pubKeys: []ed25519.PublicKey{
					signer1.Public().(ed25519.PublicKey),
					randEd25519PrivateKey().Public().(ed25519.PublicKey),
					randEd25519PrivateKey().Public().(ed25519.PublicKey),
				},
				signErr: nil,
				verfErr: iota.ErrMilestoneMinSigsThresholdNotReached,
			}
		}(),
		func() test {
			signer1 := randEd25519PrivateKey()
			signer2 := randEd25519PrivateKey()
			msg := &iota.Message{Version: 1, Parent1: rand32ByteHash(), Parent2: rand32ByteHash()}
			msPayload := &iota.Milestone{Index: 1000, Timestamp: uint64(time.Now().Unix())}

			return test{
				name:      "err - same public key",
				msg:       msg,
				ms:        msPayload,
				signer:    iota.InMemoryEd25519MilestoneSigner(signer1, signer2),
				sigsCount: 2,
				minValid:  2,
				pubKeys: []ed25519.PublicKey{
					signer1.Public().(ed25519.PublicKey),
					signer1.Public().(ed25519.PublicKey),
				},
				signErr: nil,
				verfErr: iota.ErrMilestoneSigVerificationDupPublicKeys,
			}
		}(),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.ms.Sign(test.msg, test.sigsCount, test.signer)
			if test.signErr != nil {
				assert.True(t, errors.Is(err, test.signErr))
				return
			}
			assert.NoError(t, err)

			pubKeysWhichVerified, err := test.ms.VerifySignatures(test.msg, test.minValid, test.pubKeys)
			if test.verfErr != nil {
				assert.True(t, errors.Is(err, test.verfErr))
				return
			}

			assert.NoError(t, err)
			assert.EqualValues(t, test.pubKeysWhichVerified, pubKeysWhichVerified)
		})
	}
}
