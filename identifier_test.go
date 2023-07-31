package iotago_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
)

func TestIdentifier_Bytes(t *testing.T) {
	foo := iotago.IdentifierFromData([]byte("foo"))
	bytes, err := foo.Bytes()
	require.NoError(t, err)
	require.Len(t, bytes, iotago.IdentifierLength)

	decoded, i, err := iotago.IdentifierFromBytes(bytes)
	require.NoError(t, err)
	require.Equal(t, i, iotago.IdentifierLength)
	require.Equal(t, decoded, foo)
}
