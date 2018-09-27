// +build cgo
// +build pow_avx
// +build amd64

package pow

import (
	"testing"
)

func TestPowAVX(t *testing.T) {
	sp := testPoW(t, powAVX)
	t.Logf("%d kH/sec on AVX PoW", int(sp))
}

func TestPowAVX_1(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	proc := PowProcs
	PowProcs = 1
	sp := testPoW(t, powAVX)
	t.Logf("%d kH/sec on AVX PoW", int(sp))
	PowProcs = proc
}

func TestPowAVX_32(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	proc := PowProcs
	PowProcs = 32
	sp := testPoW(t, powAVX)
	t.Logf("%d kH/sec on AVX PoW", int(sp))
	PowProcs = proc
}

func TestPowAVX_64(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	proc := PowProcs
	PowProcs = 64
	sp := testPoW(t, powAVX)
	t.Logf("%d kH/sec on AVX PoW", int(sp))
	PowProcs = proc
}
