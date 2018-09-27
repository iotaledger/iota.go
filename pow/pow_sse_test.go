// +build cgo
// +build pow_sse
// +build amd64

package pow

import (
	"testing"
)

func TestPowSSE(t *testing.T) {
	sp := testPoW(t, powSSE)
	t.Logf("%d kH/sec on SSE PoW", int(sp))
}

func TestPowSSE_1(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	proc := PowProcs
	PowProcs = 1
	sp := testPoW(t, powSSE)
	t.Logf("%d kH/sec on SSE PoW", int(sp))
	PowProcs = proc
}

func TestPowSSE_32(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	proc := PowProcs
	PowProcs = 32
	sp := testPoW(t, powSSE)
	t.Logf("%d kH/sec on SSE PoW", int(sp))
	PowProcs = proc
}

func TestPowSSE_64(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	proc := PowProcs
	PowProcs = 64
	sp := testPoW(t, powSSE)
	t.Logf("%d kH/sec on SSE PoW", int(sp))
	PowProcs = proc
}
