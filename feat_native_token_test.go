//nolint:dupl,scopelint
package iotago_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestNativeTokenDeSerialization(t *testing.T) {
	ntIn := &iotago.NativeTokenFeature{
		ID:     tpkg.Rand38ByteArray(),
		Amount: new(big.Int).SetUint64(1000),
	}

	ntBytes, err := tpkg.TestAPI.Encode(ntIn, serix.WithValidation())
	require.NoError(t, err)

	ntOut := &iotago.NativeTokenFeature{}
	_, err = tpkg.TestAPI.Decode(ntBytes, ntOut, serix.WithValidation())
	require.NoError(t, err)

	require.EqualValues(t, ntIn, ntOut)
}
