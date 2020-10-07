package iota_test

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/iotaledger/iota.go"
	"github.com/stretchr/testify/assert"
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
	rand.Read(msInclMerkleProof[:])
	var msSignature [iota.MilestoneSignatureLength]byte
	rand.Read(msSignature[:])

	msg := &iota.Message{Version: msgVersion, Parent1: parent1, Parent2: parent2}
	msPayload := &iota.Milestone{
		Index: msIndex, Timestamp: msTimestamp,
		InclusionMerkleProof: msInclMerkleProof, Signature: msSignature,
	}

	sigInput, extSig, err := msPayload.SignatureInput(msg)
	assert.NoError(t, err)

	assert.True(t, bytes.Equal(msSignature[:], extSig[:]))

	var offset int
	assert.EqualValues(t, msgVersion, sigInput[0])
	offset += iota.OneByte
	assert.True(t, bytes.Equal(parent1[:], sigInput[offset:offset+iota.MessageIDLength]))
	offset += iota.MessageIDLength
	assert.True(t, bytes.Equal(parent2[:], sigInput[offset:offset+iota.MessageIDLength]))
	offset += iota.MessageIDLength
	assert.EqualValues(t, iota.MilestoneBinSerializedSize, binary.LittleEndian.Uint32(sigInput[offset:offset+iota.PayloadLengthByteSize]))
	offset += iota.PayloadLengthByteSize
	assert.EqualValues(t, iota.MilestonePayloadTypeID, binary.LittleEndian.Uint32(sigInput[offset:offset+iota.TypeDenotationByteSize]))
	offset += iota.TypeDenotationByteSize
	assert.EqualValues(t, msIndex, binary.LittleEndian.Uint64(sigInput[offset:offset+iota.UInt64ByteSize]))
	offset += iota.UInt64ByteSize
	assert.EqualValues(t, msTimestamp, binary.LittleEndian.Uint64(sigInput[offset:offset+iota.UInt64ByteSize]))
	offset += iota.UInt64ByteSize
	assert.True(t, bytes.Equal(msInclMerkleProof[:], sigInput[offset:offset+iota.MilestoneInclusionMerkleProofLength]))
	offset += iota.MilestoneInclusionMerkleProofLength
	assert.Len(t, sigInput, offset)
}

func TestMilestoneSigning(t *testing.T) {
	type test struct {
		name    string
		msg     *iota.Message
		ms      *iota.Milestone
		signer  ed25519.PrivateKey
		pubKey  ed25519.PublicKey
		signErr error
		verfErr error
	}

	tests := []test{
		func() test {

			signer := randEd25519PrivateKey()
			msg := &iota.Message{Version: 1, Parent1: rand32ByteHash(), Parent2: rand32ByteHash()}
			msPayload := &iota.Milestone{Index: 1000, Timestamp: uint64(time.Now().Unix())}

			return test{
				name:    "ok",
				msg:     msg,
				ms:      msPayload,
				signer:  signer,
				pubKey:  signer.Public().(ed25519.PublicKey),
				signErr: nil,
				verfErr: nil,
			}
		}(),
		func() test {

			signer := randEd25519PrivateKey()
			msg := &iota.Message{Version: 1, Parent1: rand32ByteHash(), Parent2: rand32ByteHash()}
			msPayload := &iota.Milestone{Index: 1000, Timestamp: uint64(time.Now().Unix())}

			return test{
				name:    "err - invalid signature",
				msg:     msg,
				ms:      msPayload,
				signer:  signer,
				pubKey:  randEd25519PrivateKey().Public().(ed25519.PublicKey),
				signErr: nil,
				verfErr: iota.ErrInvalidMilestoneSignature,
			}
		}(),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.ms.Sign(test.msg, test.signer)
			if test.signErr != nil {
				assert.True(t, errors.Is(err, test.signErr))
				return
			}
			assert.NoError(t, err)

			err = test.ms.VerifySignature(test.msg, test.pubKey)
			if test.verfErr != nil {
				assert.True(t, errors.Is(err, test.verfErr))
				return
			}
			assert.NoError(t, err)
		})
	}
}
