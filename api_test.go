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

func (test *deSerializeTest) assertBinaryEncodeDecode(t *testing.T) {
	t.Helper()

	serixData, err := tpkg.ZeroCostTestAPI.Encode(test.source, serix.WithValidation())
	if test.seriErr != nil {
		require.ErrorIs(t, err, test.seriErr, "binary encoding")

		// Encode again without validation so we can check that deserialization would also fail.
		serixData, err = tpkg.ZeroCostTestAPI.Encode(test.source)
		require.NoError(t, err, "binary encoding")
	} else {
		require.NoError(t, err, "binary encoding")
	}

	if src, ok := test.source.(iotago.Sizer); ok {
		require.Equal(t, src.Size(), len(serixData), "binary encoding")
	}

	serixTarget := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	bytesRead, err := tpkg.ZeroCostTestAPI.Decode(serixData, serixTarget, serix.WithValidation())
	if test.deSeriErr != nil {
		require.ErrorIs(t, err, test.deSeriErr, "binary decoding")

		return
	}
	require.NoError(t, err, "binary decoding")
	require.Len(t, serixData, bytesRead, "binary decoding")
	require.EqualValues(t, test.source, serixTarget, "binary decoding")
}

func (test *deSerializeTest) assertJSONEncodeDecode(t *testing.T) {
	t.Helper()

	sourceJSON, err := tpkg.ZeroCostTestAPI.JSONEncode(test.source, serix.WithValidation())
	if test.seriErr != nil {
		require.ErrorIs(t, err, test.seriErr, "JSON encoding")

		// Encode again without validation so we can check that deserialization would also fail.
		sourceJSON, err = tpkg.ZeroCostTestAPI.JSONEncode(test.source)
		require.NoError(t, err, "JSON encoding")
	} else {
		require.NoError(t, err, "JSON encoding")
	}

	jsonDest := reflect.New(reflect.TypeOf(test.target).Elem()).Interface()
	err = tpkg.ZeroCostTestAPI.JSONDecode(sourceJSON, jsonDest, serix.WithValidation())
	if test.deSeriErr != nil {
		require.ErrorIs(t, err, test.deSeriErr, "JSON decoding")

		return
	}
	require.NoError(t, err, "JSON decoding")
	require.EqualValues(t, test.source, jsonDest, "JSON decoding")
}

func (test *deSerializeTest) deSerialize(t *testing.T) {
	t.Helper()

	if reflect.TypeOf(test.target).Kind() != reflect.Ptr {
		// This is required for the serixTarget creation hack to work.
		t.Fatal("test target must be a pointer")
	}

	test.assertBinaryEncodeDecode(t)
	test.assertJSONEncodeDecode(t)
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
