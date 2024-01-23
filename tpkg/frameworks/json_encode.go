package frameworks

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2/serix"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

// JSONEncodeTest is used to check if the JSON encoding is equal to a manually provided JSON string.
type JSONEncodeTest struct {
	Name   string
	Source any
	// the Target should be an indented JSON string (tabs instead of spaces)
	Target string
}

func (test *JSONEncodeTest) Run(t *testing.T) {
	t.Helper()

	sourceJSON, err := tpkg.ZeroCostTestAPI.JSONEncode(test.Source, serix.WithValidation())
	require.NoError(t, err, "JSON encoding")

	var b bytes.Buffer
	err = json.Indent(&b, sourceJSON, "", "\t")
	require.NoError(t, err, "JSON indenting")
	indentedJSON := b.String()

	require.EqualValues(t, test.Target, string(indentedJSON))
}
