//nolint:forcetypeassert
package builder_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/builder"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestBasicBlockBuilder(t *testing.T) {
	parents := tpkg.SortedRandBlockIDs(4)

	taggedDataPayload := &iotago.TaggedData{
		Tag:  []byte("hello world"),
		Data: []byte{1, 2, 3, 4},
	}
	block, err := builder.NewBasicBlockBuilder(tpkg.ZeroCostTestAPI).
		Payload(taggedDataPayload).
		StrongParents(parents).
		CalculateAndSetMaxBurnedMana(100).
		Build()
	require.NoError(t, err)

	require.Equal(t, iotago.BlockBodyTypeBasic, block.Body.Type())

	basicBlock := block.Body.(*iotago.BasicBlockBody)
	expectedBurnedMana, err := basicBlock.ManaCost(100, tpkg.ZeroCostTestAPI.ProtocolParameters().WorkScoreParameters())
	require.NoError(t, err)
	require.EqualValues(t, expectedBurnedMana, basicBlock.MaxBurnedMana)
}

func TestValidationBlockBuilder(t *testing.T) {
	parents := tpkg.SortedRandBlockIDs(4)

	block, err := builder.NewValidationBlockBuilder(tpkg.ZeroCostTestAPI).
		StrongParents(parents).
		HighestSupportedVersion(100).
		Build()
	require.NoError(t, err)

	require.Equal(t, iotago.BlockBodyTypeValidation, block.Body.Type())

	basicBlock := block.Body.(*iotago.ValidationBlockBody)
	require.EqualValues(t, 100, basicBlock.HighestSupportedVersion)
}
