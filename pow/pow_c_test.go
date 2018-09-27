// +build cgo
// +build pow_c
// +build linux darwin windows

package pow

import (
	"testing"
)

func TestPowC(t *testing.T) {
	sp := testPoW(t, powC)
	t.Logf("%d kH/sec on C PoW", int(sp))
}

func TestPowC_1(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	proc := PowProcs
	PowProcs = 1
	sp := testPoW(t, powC)
	t.Logf("%d kH/sec on C PoW", int(sp))
	PowProcs = proc
}

func TestPowC_32(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	proc := PowProcs
	PowProcs = 32
	sp := testPoW(t, powC)
	t.Logf("%d kH/sec on C PoW", int(sp))
	PowProcs = proc
}

func TestPowC_64(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	proc := PowProcs
	PowProcs = 64
	sp := testPoW(t, powC)
	t.Logf("%d kH/sec on C PoW", int(sp))
	PowProcs = proc
}
