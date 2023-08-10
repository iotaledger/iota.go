package iotago_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/hexutil"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestBlock_DeSerialize(t *testing.T) {
	// TODO: what does this test actually do?
	tests := []deSerializeTest{
		{
			name:   "ok - no payload",
			source: tpkg.RandProtocolBlock(tpkg.RandBasicBlock(1337), tpkg.TestAPI),
			target: &iotago.ProtocolBlock{},
		},
		{
			name:   "ok - transaction",
			source: tpkg.RandProtocolBlock(tpkg.RandBasicBlock(iotago.PayloadTransaction), tpkg.TestAPI),
			target: &iotago.ProtocolBlock{},
		},
		{
			name:   "ok - milestone",
			source: tpkg.RandProtocolBlock(tpkg.RandBasicBlock(iotago.PayloadMilestone), tpkg.TestAPI),
			target: &iotago.ProtocolBlock{},
		},
		{
			name:   "ok - tagged data",
			source: tpkg.RandProtocolBlock(tpkg.RandBasicBlock(iotago.PayloadTaggedData), tpkg.TestAPI),
			target: &iotago.ProtocolBlock{},
		},
		{
			name:   "ok - validation block",
			source: tpkg.RandProtocolBlock(tpkg.ValidationBlock(), tpkg.TestAPI),
			target: &iotago.ProtocolBlock{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestProtocolBlock_ProtocolVersionSyntactical(t *testing.T) {
	block := &iotago.ProtocolBlock{
		BlockHeader: iotago.BlockHeader{
			ProtocolVersion:  tpkg.TestAPI.Version() + 1,
			SlotCommitmentID: iotago.NewEmptyCommitment(tpkg.TestAPI.Version()).MustID(),
		},
		Signature: tpkg.RandEd25519Signature(),
		Block: &iotago.BasicBlock{
			StrongParents: tpkg.SortedRandBlockIDs(1),
			Payload:       nil,
		},
	}

	_, err := tpkg.TestAPI.Encode(block, serix.WithValidation())
	require.ErrorContains(t, err, "mismatched protocol version")
}

func TestProtocolBlock_DeserializationNotEnoughData(t *testing.T) {
	blockBytes := []byte{byte(tpkg.TestAPI.Version()), 1}

	block := &iotago.ProtocolBlock{}
	_, err := tpkg.TestAPI.Decode(blockBytes, block)
	require.ErrorIs(t, err, serializer.ErrDeserializationNotEnoughData)
}

func TestBasicBlock_MinSize(t *testing.T) {
	minProtocolBlock := &iotago.ProtocolBlock{
		BlockHeader: iotago.BlockHeader{
			ProtocolVersion:  tpkg.TestAPI.Version(),
			IssuingTime:      tpkg.RandUTCTime(),
			SlotCommitmentID: iotago.NewEmptyCommitment(tpkg.TestAPI.Version()).MustID(),
		},
		Signature: tpkg.RandEd25519Signature(),
		Block: &iotago.BasicBlock{
			StrongParents: tpkg.SortedRandBlockIDs(1),
			Payload:       nil,
		},
	}

	blockBytes, err := tpkg.TestAPI.Encode(minProtocolBlock)
	require.NoError(t, err)

	block2 := &iotago.ProtocolBlock{}
	consumedBytes, err := tpkg.TestAPI.Decode(blockBytes, block2, serix.WithValidation())
	require.NoError(t, err)
	require.Equal(t, minProtocolBlock, block2)
	require.Equal(t, len(blockBytes), consumedBytes)
}

func TestValidationBlock_MinSize(t *testing.T) {
	minProtocolBlock := &iotago.ProtocolBlock{
		BlockHeader: iotago.BlockHeader{
			ProtocolVersion:  tpkg.TestAPI.Version(),
			IssuingTime:      tpkg.RandUTCTime(),
			SlotCommitmentID: iotago.NewEmptyCommitment(tpkg.TestAPI.Version()).MustID(),
		},
		Signature: tpkg.RandEd25519Signature(),
		Block: &iotago.ValidationBlock{
			StrongParents:           tpkg.SortedRandBlockIDs(1),
			HighestSupportedVersion: tpkg.TestAPI.Version(),
		},
	}

	blockBytes, err := tpkg.TestAPI.Encode(minProtocolBlock)
	require.NoError(t, err)

	block2 := &iotago.ProtocolBlock{}
	consumedBytes, err := tpkg.TestAPI.Decode(blockBytes, block2, serix.WithValidation())
	require.NoError(t, err)
	require.Equal(t, minProtocolBlock, block2)
	require.Equal(t, len(blockBytes), consumedBytes)
}

func TestValidationBlock_HighestSupportedVersion(t *testing.T) {
	protocolBlock := &iotago.ProtocolBlock{
		BlockHeader: iotago.BlockHeader{
			ProtocolVersion:  tpkg.TestAPI.Version(),
			IssuingTime:      tpkg.RandUTCTime(),
			SlotCommitmentID: iotago.NewEmptyCommitment(tpkg.TestAPI.Version()).MustID(),
		},
		Signature: tpkg.RandEd25519Signature(),
	}

	// Invalid HighestSupportedVersion.
	{
		protocolBlock.Block = &iotago.ValidationBlock{
			StrongParents:           tpkg.SortedRandBlockIDs(1),
			HighestSupportedVersion: tpkg.TestAPI.Version() - 1,
		}
		blockBytes, err := tpkg.TestAPI.Encode(protocolBlock)
		require.NoError(t, err)

		block2 := &iotago.ProtocolBlock{}
		_, err = tpkg.TestAPI.Decode(blockBytes, block2, serix.WithValidation())
		require.ErrorContains(t, err, "highest supported version")
	}

	// Valid HighestSupportedVersion.
	{
		protocolBlock.Block = &iotago.ValidationBlock{
			StrongParents:           tpkg.SortedRandBlockIDs(1),
			HighestSupportedVersion: tpkg.TestAPI.Version(),
		}
		blockBytes, err := tpkg.TestAPI.Encode(protocolBlock)
		require.NoError(t, err)

		block2 := &iotago.ProtocolBlock{}
		consumedBytes, err := tpkg.TestAPI.Decode(blockBytes, block2, serix.WithValidation())
		require.NoError(t, err)
		require.Equal(t, protocolBlock, block2)
		require.Equal(t, len(blockBytes), consumedBytes)
	}
}

func TestBlockJSONMarshalling(t *testing.T) {
	networkID := iotago.NetworkIDFromString("xxxNetwork")
	issuingTime := tpkg.RandUTCTime()
	commitmentID := iotago.NewEmptyCommitment(tpkg.TestAPI.Version()).MustID()
	issuerID := tpkg.RandAccountID()
	signature := tpkg.RandEd25519Signature()
	strongParents := tpkg.SortedRandBlockIDs(1)
	validationBlock := &iotago.ProtocolBlock{
		BlockHeader: iotago.BlockHeader{
			ProtocolVersion:  tpkg.TestAPI.Version(),
			IssuingTime:      issuingTime,
			IssuerID:         issuerID,
			NetworkID:        networkID,
			SlotCommitmentID: commitmentID,
		},
		Block: &iotago.ValidationBlock{
			StrongParents:           strongParents,
			HighestSupportedVersion: tpkg.TestAPI.Version(),
		},
		Signature: signature,
	}

	blockJSON := fmt.Sprintf(`{"protocolVersion":%d,"networkId":"%d","issuingTime":"%s","slotCommitment":"%s","latestFinalizedSlot":"0","issuerId":"%s","block":{"type":%d,"strongParents":["%s"],"weakParents":[],"shallowLikeParents":[],"highestSupportedVersion":%d,"protocolParametersHash":"0x0000000000000000000000000000000000000000000000000000000000000000"},"signature":{"type":%d,"publicKey":"%s","signature":"%s"}}`,
		tpkg.TestAPI.Version(),
		networkID,
		issuingTime.Format(time.RFC3339Nano),
		commitmentID.ToHex(),
		issuerID.ToHex(),
		iotago.BlockTypeValidation,
		strongParents[0].ToHex(),
		tpkg.TestAPI.Version(),
		iotago.SignatureEd25519,
		hexutil.EncodeHex(signature.PublicKey[:]),
		hexutil.EncodeHex(signature.Signature[:]),
	)

	jsonEncode, err := tpkg.TestAPI.JSONEncode(validationBlock)

	fmt.Println(string(jsonEncode))

	require.NoError(t, err)
	require.Equal(t, blockJSON, string(jsonEncode))
}

// TODO: add tests
//  - max size
//  - parents parameters basic block
//  - parents parameters validator block
//  - decode/encode protocol parameters
