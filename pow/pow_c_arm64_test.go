// +build cgo
// +build pow_arm_c128
// +build linux
// +build arm64

package pow

import (
	"testing"
)

func TestPowCARM64(t *testing.T) {
	sp := testPoW(t, powCARM64)
	t.Logf("%d kH/sec on CARM64 PoW", int(sp))
}

func TestPowCARM64_1(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	proc := PowProcs
	PowProcs = 1
	sp := testPoW(t, powCARM64)
	t.Logf("%d kH/sec on CARM64 PoW", int(sp))
	PowProcs = proc
}

func TestPowCARM64_32(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	proc := PowProcs
	PowProcs = 32
	sp := testPoW(t, powCARM64)
	t.Logf("%d kH/sec on CARM64 PoW", int(sp))
	PowProcs = proc
}

func TestPowCARM64_64(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	proc := PowProcs
	PowProcs = 64
	sp := testPoW(t, powCARM64)
	t.Logf("%d kH/sec on CARM64 PoW", int(sp))
	PowProcs = proc
}
