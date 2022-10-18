package iotago_test

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/core/serix"
	"github.com/iotaledger/hive.go/serializer/v2"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())

	// call the tests
	os.Exit(m.Run())
}

func TestBlock_DeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok - no payload",
			source: tpkg.RandBlock(1337),
			target: &iotago.Block{},
		},
		{
			name:   "ok - transaction",
			source: tpkg.RandBlock(iotago.PayloadTransaction),
			target: &iotago.Block{},
		},
		{
			name:   "ok - milestone",
			source: tpkg.RandBlock(iotago.PayloadMilestone),
			target: &iotago.Block{},
		},
		{
			name:   "ok - tagged data",
			source: tpkg.RandBlock(iotago.PayloadTaggedData),
			target: &iotago.Block{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestBlock_MinSize(t *testing.T) {

	block := &iotago.Block{
		ProtocolVersion: tpkg.TestProtocolVersion,
		Parents:         tpkg.SortedRandBlockIDs(1),
		Payload:         nil,
	}

	msgBytes, err := v2API.Encode(block)
	require.NoError(t, err)

	block2 := &iotago.Block{}
	_, err = v2API.Decode(msgBytes, block2, serix.WithValidation())
	require.NoError(t, err)
	require.Equal(t, block, block2)
}

func TestBlock_ProtocolVersionSyntactical(t *testing.T) {

	block := &iotago.Block{
		ProtocolVersion: tpkg.TestProtocolVersion + 1,
		Parents:         tpkg.SortedRandBlockIDs(1),
		Payload:         nil,
	}

	_, err := v2API.Encode(block, serix.WithValidation())
	require.Error(t, err)
}

func TestBlock_DeserializationNotEnoughData(t *testing.T) {

	blockBytes := []byte{tpkg.TestProtocolVersion, 1}

	block := &iotago.Block{}
	_, err := v2API.Decode(blockBytes, block)
	require.ErrorIs(t, err, serializer.ErrDeserializationNotEnoughData)
}
