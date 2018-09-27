// +build cgo
// +build pow_c128
// +build linux darwin windows
// +build amd64

package pow

import (
	"testing"
)


func TestPowC128(t *testing.T) {
	sp := testPoW(t, powC128)
	t.Logf("%d kH/sec on C 128 PoW", int(sp))
}

func TestPowC128_1(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	proc := PowProcs
	PowProcs = 1
	sp := testPoW(t, powC128)
	t.Logf("%d kH/sec on C 128 PoW", int(sp))
	PowProcs = proc
}

func TestPowC128_32(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	proc := PowProcs
	PowProcs = 32
	sp := testPoW(t, powC128)
	t.Logf("%d kH/sec on C 128 PoW", int(sp))
	PowProcs = proc
}

func TestPowC128_64(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	proc := PowProcs
	PowProcs = 64
	sp := testPoW(t, powC128)
	t.Logf("%d kH/sec on C 128 PoW", int(sp))
	PowProcs = proc
}