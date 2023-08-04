package iotago_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/api"
	"github.com/iotaledger/iota.go/v4/builder"
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

func createBlockAtSlotWithVersion(t *testing.T, index iotago.SlotIndex, version iotago.Version, apiProvider *api.EpochBasedProvider) error {
	t.Helper()

	apiForSlot := apiProvider.APIForSlot(index)
	block, err := builder.NewBasicBlockBuilder(apiForSlot).
		ProtocolVersion(version).
		StrongParents(iotago.BlockIDs{iotago.BlockID{}}).
		IssuingTime(apiForSlot.TimeProvider().SlotStartTime(index)).
		SlotCommitmentID(iotago.NewCommitment(apiForSlot.Version(), index-apiForSlot.ProtocolParameters().MinCommittableAge(), iotago.CommitmentID{}, iotago.Identifier{}, 0).MustID()).
		Build()
	require.NoError(t, err)

	return lo.Return2(apiForSlot.Encode(block, serix.WithValidation()))
}

func TestProtocolBlock_ProtocolVersionSyntactical(t *testing.T) {
	apiProvider := api.NewEpochBasedProvider(
		api.WithAPIForMissingVersionCallback(
			func(version iotago.Version) (iotago.API, error) {
				return iotago.V3API(iotago.NewV3ProtocolParameters(iotago.WithVersion(version))), nil
			},
		),
	)
	apiProvider.AddProtocolParametersAtEpoch(iotago.NewV3ProtocolParameters(), 0)
	apiProvider.AddProtocolParametersAtEpoch(iotago.NewV3ProtocolParameters(iotago.WithVersion(4)), 3)

	timeProvider := apiProvider.CurrentAPI().TimeProvider()

	require.ErrorIs(t, createBlockAtSlotWithVersion(t, timeProvider.EpochStart(1), 2, apiProvider), iotago.ErrInvalidBlockVersion)

	require.NoError(t, createBlockAtSlotWithVersion(t, timeProvider.EpochEnd(1), 3, apiProvider))

	require.NoError(t, createBlockAtSlotWithVersion(t, timeProvider.EpochEnd(2), 3, apiProvider))

	require.ErrorIs(t, createBlockAtSlotWithVersion(t, timeProvider.EpochStart(3), 3, apiProvider), iotago.ErrInvalidBlockVersion)

	require.NoError(t, createBlockAtSlotWithVersion(t, timeProvider.EpochStart(3), 4, apiProvider))

	require.NoError(t, createBlockAtSlotWithVersion(t, timeProvider.EpochEnd(3), 4, apiProvider))

	require.NoError(t, createBlockAtSlotWithVersion(t, timeProvider.EpochStart(5), 4, apiProvider))

	apiProvider.AddProtocolParametersAtEpoch(iotago.NewV3ProtocolParameters(iotago.WithVersion(5)), 10)

	require.NoError(t, createBlockAtSlotWithVersion(t, timeProvider.EpochEnd(9), 4, apiProvider))

	require.ErrorIs(t, createBlockAtSlotWithVersion(t, timeProvider.EpochStart(10), 4, apiProvider), iotago.ErrInvalidBlockVersion)

	require.NoError(t, createBlockAtSlotWithVersion(t, timeProvider.EpochStart(10), 5, apiProvider))
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
	// TODO: finish this test.
	validationBlock := &iotago.ProtocolBlock{
		BlockHeader: iotago.BlockHeader{
			ProtocolVersion:  tpkg.TestAPI.Version(),
			SlotCommitmentID: iotago.NewEmptyCommitment(tpkg.TestAPI.Version()).MustID(),
		},
		Signature: tpkg.RandEd25519Signature(),
		Block: &iotago.ValidationBlock{
			StrongParents:           tpkg.SortedRandBlockIDs(1),
			HighestSupportedVersion: tpkg.TestAPI.Version(),
		},
	}

	// protoParamsJSON := `{"type":0,"version":3,"networkName":"xxxNetwork","bech32Hrp":"xxx","rentStructure":{"vByteCost":6,"vByteFactorData":7,"vByteFactorKey":8},"tokenSupply":"1234567890987654321","genesisUnixTimestamp":"1681373293","slotDurationInSeconds":10,"slotsPerEpochExponent":13,"manaGenerationRate":1,"manaGenerationRateExponent":27,"manaDecayFactors":[10,20],"manaDecayFactorsExponent":32,"manaDecayFactorEpochsSum":1337,"manaDecayFactorEpochsSumExponent":20,"stakingUnbondingPeriod":"11","evictionAge":"10","livenessThreshold":"3"}`

	jsonEncode, err := tpkg.TestAPI.JSONEncode(validationBlock)
	fmt.Println(string(jsonEncode))
	require.NoError(t, err)
	// require.Equal(t, protoParamsJSON, string(jsonEncode))
	//
	// var decodedProtoParams iotago.ProtocolParameters
	// err = tpkg.TestAPI.JSONDecode([]byte(protoParamsJSON), &decodedProtoParams)
	// require.NoError(t, err)
	//
	// require.Equal(t, protoParams, decodedProtoParams)
}

// TODO: add tests
//  - max size
//  - parents parameters basic block
//  - parents parameters validator block
