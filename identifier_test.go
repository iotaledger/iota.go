package iotago

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentifier_Bytes(t *testing.T) {
	foo := IdentifierFromData([]byte("foo"))
	bytes, err := foo.Bytes()
	require.NoError(t, err)
	require.Len(t, bytes, IdentifierLength)

	var decoded Identifier
	i, err := decoded.FromBytes(bytes)
	require.NoError(t, err)
	require.Equal(t, i, IdentifierLength)
	require.Equal(t, decoded, foo)
}
