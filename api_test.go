//nolint:scopelint
package iotago_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

const (
	OneIOTA iotago.BaseToken = 1_000_000
)

type deSerializeTest struct {
	name      string
	source    any
	target    any
	seriErr   error
	deSeriErr error
}

func (test *deSerializeTest) deSerialize(t *testing.T) {
	t.Helper()

	serixData, err := tpkg.ZeroCostTestAPI.Encode(test.source, serix.WithValidation())
	if test.seriErr != nil {
		require.ErrorIs(t, err, test.seriErr)

		return
	}
	require.NoError(t, err)

	if src, ok := test.source.(iotago.Sizer); ok {
		require.Equal(t, src.Size(), len(serixData))
	}

	serixTarget := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	bytesRead, err := tpkg.ZeroCostTestAPI.Decode(serixData, serixTarget)
	if test.deSeriErr != nil {
		require.ErrorIs(t, err, test.deSeriErr)

		return
	}
	require.NoError(t, err)
	require.Len(t, serixData, bytesRead)
	require.EqualValues(t, test.source, serixTarget)

	sourceJSON, err := tpkg.ZeroCostTestAPI.JSONEncode(test.source)
	require.NoError(t, err)

	jsonDest := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	require.NoError(t, tpkg.ZeroCostTestAPI.JSONDecode(sourceJSON, jsonDest))

	require.EqualValues(t, test.source, jsonDest)
}

func TestProtocolParameters_DeSerialize(t *testing.T) {
	tests := []*deSerializeTest{
		{
			name:      "ok",
			source:    tpkg.RandProtocolParameters(),
			target:    &iotago.V3ProtocolParameters{},
			seriErr:   nil,
			deSeriErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}
