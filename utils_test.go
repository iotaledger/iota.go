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

func TestValidTrit(t *testing.T) {
	type validTritTC struct {
		in    int
		valid bool
	}

	var validTritCases = []validTritTC{
		validTritTC{in: -1, valid: true},
		validTritTC{in: 0, valid: true},
		validTritTC{in: 1, valid: true},
		validTritTC{in: -2, valid: false},
		validTritTC{in: 2, valid: false},
	}

	for _, tc := range validTritCases {
		if ValidTrit(tc.in) != tc.valid {
			t.Fatalf("ValidTrit(%q) should be %#v but is not", tc.in, tc.valid)
		}
	}
}

func TestValidTrits(t *testing.T) {
	type validTritsTC struct {
		in    []int
		valid bool
	}

	var validTritsCases = []validTritsTC{
		validTritsTC{in: []int{0}, valid: true},
		validTritsTC{in: []int{-1}, valid: true},
		validTritsTC{in: []int{1}, valid: true},
		validTritsTC{in: []int{0, -1, 1}, valid: true},
		validTritsTC{in: []int{2, -1, 1}, valid: false},
	}

	for _, tc := range validTritsCases {
		if ValidTrits(tc.in) != tc.valid {
			t.Fatalf("ValidTrits(%q) should be %#v but is not", tc.in, tc.valid)
		}
	}
}
