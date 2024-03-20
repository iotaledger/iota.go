package frameworks

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

type DeSerializeTest struct {
	Name      string
	Source    any
	Target    any
	SeriErr   error
	DeSeriErr error
}

func (test *DeSerializeTest) assertBinaryEncodeDecode(t *testing.T) {
	t.Helper()

	serixData, err := tpkg.ZeroCostTestAPI.Encode(test.Source, serix.WithValidation())
	if test.SeriErr != nil {
		require.ErrorIs(t, err, test.SeriErr, "binary encoding")

		// Encode again without validation so we can check that deserialization would also fail.
		serixData, err = tpkg.ZeroCostTestAPI.Encode(test.Source)
		require.NoError(t, err, "binary encoding")
	} else {
		require.NoError(t, err, "binary encoding")
	}

	if src, ok := test.Source.(iotago.Sizer); ok {
		require.Len(t, serixData, src.Size(), "binary encoding")
	}

	serixTarget := reflect.New(reflect.TypeOf(test.Target).Elem()).Interface()
	bytesRead, err := tpkg.ZeroCostTestAPI.Decode(serixData, serixTarget, serix.WithValidation())
	if test.DeSeriErr != nil {
		require.ErrorIs(t, err, test.DeSeriErr, "binary decoding")

		return
	}
	require.NoError(t, err, "binary decoding")
	require.Len(t, serixData, bytesRead, "binary decoding")
	require.EqualValues(t, test.Source, serixTarget, "binary decoding")
}

func (test *DeSerializeTest) assertJSONEncodeDecode(t *testing.T) {
	t.Helper()

	sourceJSON, err := tpkg.ZeroCostTestAPI.JSONEncode(test.Source, serix.WithValidation())
	if test.SeriErr != nil {
		require.ErrorIs(t, err, test.SeriErr, "JSON encoding")

		// Encode again without validation so we can check that deserialization would also fail.
		sourceJSON, err = tpkg.ZeroCostTestAPI.JSONEncode(test.Source)
		require.NoError(t, err, "JSON encoding")
	} else {
		require.NoError(t, err, "JSON encoding")
	}

	jsonDest := reflect.New(reflect.TypeOf(test.Target).Elem()).Interface()
	err = tpkg.ZeroCostTestAPI.JSONDecode(sourceJSON, jsonDest, serix.WithValidation())
	if test.DeSeriErr != nil {
		require.ErrorIs(t, err, test.DeSeriErr, "JSON decoding")

		return
	}
	require.NoError(t, err, "JSON decoding")
	require.EqualValues(t, test.Source, jsonDest, "JSON decoding")
}

func (test *DeSerializeTest) Run(t *testing.T) {
	t.Helper()

	if reflect.TypeOf(test.Target).Kind() != reflect.Ptr {
		// This is required for the serixTarget creation hack to work.
		t.Fatal("test target must be a pointer")
	}

	test.assertBinaryEncodeDecode(t)
	test.assertJSONEncodeDecode(t)
}
