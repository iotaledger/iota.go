package giota

import (
	"testing"
)

type addressTestCase struct {
	addr          string
	validAddr     bool
	checksum      string
	validChecksum bool
}

var addressTCs = []addressTestCase{
	addressTestCase{
		addr:          "RGVOWCDJAGSO9TNLBBPUVYE9KHBOAZNVFRVKVYYCHRKQRKRNKGGWBF9WCRJVROKLVKWZUMBABVJGAALWU",
		validAddr:     true,
		checksum:      "QNXFPRSPG",
		validChecksum: true,
	},
	addressTestCase{
		addr:          "",
		validAddr:     false,
		checksum:      "",
		validChecksum: true,
	},
	addressTestCase{
		addr:          "999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999",
		validAddr:     true,
		checksum:      "999999999",
		validChecksum: true,
	},
	addressTestCase{
		addr:          "RGVOWCDJAGSO9TNLBBPUVYE9KHBOAZNVFRVKVYYCHRKQRKRNKGGWBF9WCRJVROKLVKWZUMBABVJGAALWU",
		validAddr:     true,
		checksum:      "999999999",
		validChecksum: false,
	},
}

func TestNewAddressFromTrytes(t *testing.T) {
	for _, tc := range addressTCs {
		addr, err := NewAddressFromTrytes(tc.addr)
		if (err != nil) == tc.validAddr {
			t.Fatalf("NewAddressFromTrytes(%q) expected (err != nil) to be %#v\nerr: %#v", tc.addr, tc.validAddr, err)
		} else if addr != nil && (addr.Checksum() != tc.checksum) == tc.validChecksum {
			t.Fatalf("NewAddressFromTrytes(%q) checksum mismatch\nwant: %s\nhave: %s", tc.addr, tc.checksum, addr.Checksum())
		}
	}
}
