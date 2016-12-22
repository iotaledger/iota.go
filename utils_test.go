package giota

import (
	"testing"
)

func TestValidTryte(t *testing.T) {
	type validTryteTC struct {
		in    rune
		valid bool
	}

	var validTryteCases = []validTryteTC{
		validTryteTC{in: 'A', valid: true},
		validTryteTC{in: 'Z', valid: true},
		validTryteTC{in: '9', valid: true},
		validTryteTC{in: '8', valid: false},
		validTryteTC{in: 'a', valid: false},
		validTryteTC{in: '-', valid: false},
		validTryteTC{in: 'Ɩ', valid: false},
	}

	for _, tc := range validTryteCases {
		if ValidTryte(tc.in) != tc.valid {
			t.Fatalf("ValidTryte(%q) should be %#v but is not", tc.in, tc.valid)
		}
	}
}

func TestValidTrytes(t *testing.T) {
	type validTryteTC struct {
		in    string
		valid bool
	}

	var validTryteCases = []validTryteTC{
		validTryteTC{in: "ABCDEFGHIJKLMNOPQRSTUVWXYZ9", valid: true},
		validTryteTC{in: "ABCDEFGHIJKLMNOPQRSTUVWXYZ90", valid: false},
		validTryteTC{in: "ABCDEFGHIJKLMNOPQRSTUVWXYZ9 ", valid: false},
		validTryteTC{in: "Ɩ", valid: false},
	}

	for _, tc := range validTryteCases {
		if ValidTrytes(tc.in) != tc.valid {
			t.Fatalf("ValidTrytes(%q) should be %#v but is not", tc.in, tc.valid)
		}
	}
}
