package signing

import (
	. "github.com/iotaledger/giota/trinary"
	"testing"
)

func TestNewAddressFromTrytes(t *testing.T) {
	tests := []struct {
		name          string
		address       Trytes
		validAddr     bool
		checksum      Trytes
		validChecksum bool
	}{
		{
			name:          "valid address and checksium",
			address:       "RGVOWCDJAGSO9TNLBBPUVYE9KHBOAZNVFRVKVYYCHRKQRKRNKGGWBF9WCRJVROKLVKWZUMBABVJGAALWU",
			validAddr:     true,
			checksum:      "NPJ9QIHFW",
			validChecksum: true,
		},
		{
			name:          "test blank address fails",
			address:       "",
			validAddr:     false,
			checksum:      "",
			validChecksum: true,
		},
		{
			name:          "valid address and checksum",
			address:       "999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999",
			validAddr:     true,
			checksum:      "A9BEONKZW",
			validChecksum: true,
		},
		{
			name:          "valid address with invalid checksum",
			address:       "RGVOWCDJAGSO9TNLBBPUVYE9KHBOAZNVFRVKVYYCHRKQRKRNKGGWBF9WCRJVROKLVKWZUMBABVJGAALWU",
			validAddr:     true,
			checksum:      "A9BEONKZW",
			validChecksum: false,
		},
	}

	for _, tt := range tests {
		adr, err := NewAddressHashFromTrytes(tt.address)
		adrChecksum, adrChecksumErr := AddressChecksum(adr)
		switch {
		case (err != nil) == tt.validAddr:
			t.Fatalf("%s: NewAddressFromTrytes(%q) expected (err != nil) to be %#v\nerr: %#v",
				tt.name, tt.address, tt.validAddr, err)
		case (err == nil && adrChecksumErr == nil && adrChecksum != tt.checksum) == tt.validChecksum:
			t.Fatalf("NewAddressFromTrytes(%q) checksum mismatch\nwant: %s\nhave: %s",
				tt.address, tt.checksum, adrChecksum)
		case !tt.validAddr || !tt.validChecksum:
			continue
		}

		wcs, err := AddressWithChecksum(adr)
		if err != nil {
			t.Errorf("WithChecksum returned an error: %v", err)
		}

		if wcs != Trytes(adr)+adrChecksum {
			t.Error("WithChecksum is incorrect")
		}

		adr2, err := NewAddressHashFromTrytes(tt.address)
		if err != nil {
			t.Error(err)
		}

		if adr != adr2 {
			t.Error("NewAddressHash is incorrect")
		}
	}
}

func TestAddress(t *testing.T) {
	tests := []struct {
		name         Trytes
		seed         Trytes
		seedIndex    uint
		seedSecurity SecurityLevel
		address      Trytes
		addressValid bool
	}{
		{
			name:         "test valid address 1",
			seed:         "WQNZOHUT99PWKEBFSKQSYNC9XHT9GEBMOSJAQDQAXPEZPJNDIUB9TSNWVMHKWICW9WVZXSMDFGISOD9FZ",
			seedIndex:    0,
			seedSecurity: 2,
			address:      "AYYNHWWNZQOFYXNQSLVULU9ARZCSXNWWAFYEWEL9LIXYDFS9KDSRZF9ZID9AQWSLAEUAJSTQKGPGXNWCD",
		},
		{
			name:         "test valid address 2",
			seed:         "WQNZOHUT99PWKEBFSKQSYNC9XHT9GEBMOSJAQDQAXPEZPJNDIUB9TSNWVMHKWICW9WVZXSMDFGISOD9FZ",
			seedIndex:    1,
			seedSecurity: 2,
			address:      "9CTFIAYOFLOKXVNDFKNERQQEFR9FCIXQQHNRDKHIVVGFZQKTBWPCOIHCCQIU9ASJQECGPHDBAREDXIRCX",
		},
	}

	for _, tt := range tests {
		address, err := NewAddressHash(Address{tt.seed, tt.seedIndex, tt.seedSecurity})
		if err != nil {
			t.Errorf("%s: NewAddressHash failed with error: %s", tt.name, err)
		}

		addressCheck, err := NewAddressHashFromTrytes(tt.address)
		if err != nil {
			t.Errorf("%s: NewAddressHashFromTrytes failed with err: %s", tt.name, err)
		}

		if address != addressCheck {
			t.Errorf("%s: address: %s != address: %s", tt.name, address, addressCheck)
		}

		if err := ValidAddressHash(address); err != nil {
			t.Errorf("%s: address failed to validate: %s", tt.name, err)
		}
	}
}
