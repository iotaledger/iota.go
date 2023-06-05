package iotago_test

import (
	"math"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/serix"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

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
		StrongParents:   tpkg.SortedRandBlockIDs(1),
		SlotCommitment:  iotago.NewEmptyCommitment(),
		Signature:       tpkg.RandEd25519Signature(),
		Payload:         nil,
	}

	msgBytes, err := v3API.Encode(block)
	require.NoError(t, err)

	block2 := &iotago.Block{}
	_, err = v3API.Decode(msgBytes, block2, serix.WithValidation())
	require.NoError(t, err)
	require.Equal(t, block, block2)
}

func TestBlock_ProtocolVersionSyntactical(t *testing.T) {

	block := &iotago.Block{
		ProtocolVersion: tpkg.TestProtocolVersion + 1,
		StrongParents:   tpkg.SortedRandBlockIDs(1),
		Payload:         nil,
	}

	_, err := v3API.Encode(block, serix.WithValidation())
	require.Error(t, err)
}

func TestBlock_DeserializationNotEnoughData(t *testing.T) {

	blockBytes := []byte{tpkg.TestProtocolVersion, 1}

	block := &iotago.Block{}
	_, err := v3API.Decode(blockBytes, block)
	require.ErrorIs(t, err, serializer.ErrDeserializationNotEnoughData)
}

func TestBlock_TimestampRoundtrip(t *testing.T) {
	// Strip the monotonic clock value by rounding.
	now := time.Now().Round(0)
	block := &iotago.Block{}
	block.SetIssuingTime(now)
	issuanceTime := block.IssuingTime()

	assert.Equal(t, now, issuanceTime)
}

func TestBlock_TimestampNegative(t *testing.T) {
	unixTime := time.Unix(-1, -1000)
	block := &iotago.Block{}
	block.SetIssuingTime(unixTime)
	issuanceTime := block.IssuingTime()

	expected := time.Unix(0, 0)
	assert.Equal(t, issuanceTime, expected)
}

const (
	maxUnixInt64Sec  = math.MaxInt64 / 1_000_000_000
	maxUnixInt64Nano = math.MaxInt64 % 1_000_000_000
)

func TestBlock_TimestampExceedsMaxRepresentableNanos(t *testing.T) {
	// Maximum time value representable in nanoseconds in an int64 + 10 seconds.
	unixTime := time.Unix(int64(maxUnixInt64Sec+10), int64(maxUnixInt64Nano))

	block := &iotago.Block{}
	block.SetIssuingTime(unixTime)
	issuanceTime := block.IssuingTime()

	expected := time.Unix(maxUnixInt64Sec, maxUnixInt64Nano)
	assert.Equal(t, issuanceTime, expected)
}

func TestBlock_TimestampExceedsInt64Nanoseconds(t *testing.T) {
	block := iotago.Block{IssuingTimestamp: math.MaxUint64}
	issuanceTime := block.IssuingTime()

	expected := time.Unix(maxUnixInt64Sec, maxUnixInt64Nano)
	assert.Equal(t, issuanceTime, expected)
}
