package math_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/iota.go/v4/math"
)

func TestMain(m *testing.M) {
	// call the tests
	os.Exit(m.Run())
}

func TestAbsInt64(t *testing.T) {
	assert.EqualValues(t, int64(9223372036854775807), math.AbsInt64(-9223372036854775807))
}
