package iotago_test

import (
	"testing"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
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
	data, err := test.source.Serialize(serializer.DeSeriModePerformValidation, iotago.ZeroRentParas)
	if test.seriErr != nil {
		require.Error(t, err, test.seriErr)
		return
	}
	assert.NoError(t, err)

	bytesRead, err := test.target.Deserialize(data, serializer.DeSeriModePerformValidation, iotago.ZeroRentParas)
	if test.deSeriErr != nil {
		require.Error(t, err, test.deSeriErr)
		return
	}
	assert.NoError(t, err)
	require.Len(t, data, bytesRead)
	assert.EqualValues(t, test.source, test.target)
}
