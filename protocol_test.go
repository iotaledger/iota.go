package iotago_test

import (
	"github.com/iotaledger/iota.go/v3/tpkg"
	"testing"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	OneMi = 1_000_000
)

type deSerializeTest struct {
	name      string
	source    serializer.Serializable
	target    serializer.Serializable
	seriErr   error
	deSeriErr error
}

func (test *deSerializeTest) deSerialize(t *testing.T) {
	data, err := test.source.Serialize(serializer.DeSeriModePerformValidation, tpkg.TestProtoParas)
	if test.seriErr != nil {
		require.Error(t, err, test.seriErr)
		return
	}
	assert.NoError(t, err)
	if src, ok := test.source.(serializer.SerializableWithSize); ok {
		assert.Equal(t, len(data), src.Size())
	}

	bytesRead, err := test.target.Deserialize(data, serializer.DeSeriModePerformValidation, tpkg.TestProtoParas)
	if test.deSeriErr != nil {
		require.Error(t, err, test.deSeriErr)
		return
	}
	assert.NoError(t, err)
	require.Len(t, data, bytesRead)
	assert.EqualValues(t, test.source, test.target)
}
