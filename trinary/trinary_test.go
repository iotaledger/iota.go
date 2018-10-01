package trinary

import (
	"bytes"
	"testing"
)

func TestValidTryte(t *testing.T) {
	type validTryteTC struct {
		in    rune
		valid bool
	}

	var validTryteCases = []validTryteTC{
		{in: 'A', valid: true},
		{in: 'Z', valid: true},
		{in: '9', valid: true},
		{in: '8', valid: false},
		{in: 'a', valid: false},
		{in: '-', valid: false},
		{in: 'Ɩ', valid: false},
	}

	for _, tc := range validTryteCases {
		if (ValidTryte(tc.in) == nil) != tc.valid {
			t.Fatalf("ValidTryte(%q) should be %#v but is not", tc.in, tc.valid)
		}
	}
}

func TestValidTrytes(t *testing.T) {
	type validtrytetestcase struct {
		in    Trytes
		valid bool
	}

	var validTryteCases = []validtrytetestcase{
		{in: "ABCDEFGHIJKLMNOPQRSTUVWXYZ9", valid: true},
		{in: "ABCDEFGHIJKLMNOPQRSTUVWXYZ90", valid: false},
		{in: "ABCDEFGHIJKLMNOPQRSTUVWXYZ9 ", valid: false},
		{in: "Ɩ", valid: false},
	}

	for _, tc := range validTryteCases {
		if (ValidTrytes(tc.in) == nil) != tc.valid {
			t.Fatalf("ValidTrytes(%q) should be %#v but is not", tc.in, tc.valid)
		}
	}
}

func TestValidTrit(t *testing.T) {
	type validtrittestcase struct {
		in    int8
		valid bool
	}

	var validTritCases = []validtrittestcase{
		{in: -1, valid: true},
		{in: 0, valid: true},
		{in: 1, valid: true},
		{in: -2, valid: false},
		{in: 2, valid: false},
	}

	for _, tc := range validTritCases {
		if (ValidTrit(tc.in) == nil) != tc.valid {
			t.Fatalf("ValidTrit(%q) should be %#v but is not", tc.in, tc.valid)
		}
	}
}

func TestValidTrits(t *testing.T) {
	type validtritstestcase struct {
		in    Trits
		valid bool
	}

	var validTritsCases = []validtritstestcase{
		{in: Trits{0}, valid: true},
		{in: Trits{-1}, valid: true},
		{in: Trits{1}, valid: true},
		{in: Trits{0, -1, 1}, valid: true},
		{in: Trits{2, -1, 1}, valid: false},
	}

	for _, tc := range validTritsCases {
		if (ValidTrits(tc.in) == nil) != tc.valid {
			t.Fatalf("ValidTrits(%q) should be %#v but is not", tc.in, tc.valid)
		}
	}
}

func TestTritByteTrit(t *testing.T) {
	ts := []Trits{
		TrytesToTrits("NNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNN"),
		TrytesToTrits("MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM"),
		TrytesToTrits("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
		TrytesToTrits("ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ"),
		TrytesToTrits("999999999999999999999999999999999999999999999999999999999999999999999999999999999"),
		TrytesToTrits("SCYLJDWIM9LIXCSLETSHLQOOFDKYOVFZLAHQYCCNMYHRTNIKBZRIRACFYPOWYNSOWDNXFZUG9OEOZPOTD"),
	}

	for _, trits := range ts {
		trits[TritHashLength-1] = 0
		b, err := TritsToBytes(trits)
		if err != nil {
			t.Errorf("Bytes() failed: %s", err)
		}

		tb, err := BytesToTrits(b)
		if err != nil {
			t.Errorf("BytesToTrits() failed: %s", err)
		}

		if !TritsEqual(tb, trits) {
			t.Errorf("Trits->Bytes->Trits roundtrip failed\nwanted:\t%#v\ngot:\t%#v", trits, tb)
		}
	}
}

func TestAllBytes(t *testing.T) {
	for i := 0; i < 256; i++ {
		bs := bytes.Repeat([]byte{uint8(i)}, ByteLength)

		fst, err := BytesToTrits(bs)
		if err != nil {
			t.Errorf("BytesToTrits() failed: %s", err)
		}

		bs, err = TritsToBytes(fst)
		if err != nil {
			t.Errorf("Bytes() failed: %s", err)
		}

		snd, err := BytesToTrits(bs)
		if err != nil {
			t.Errorf("BytesToTrits() failed: %s", err)
		}

		if !TritsEqual(fst, snd) {
			t.Errorf("Bytes->Trits->Bytes->Trits roundtrip failed\nwanted:\t%#v\ngot:\t%#v", fst, snd)
		}
	}
}

func TestConvert(t *testing.T) {
	trits := Trits{0, 1, -1, 1, 1, -1, -1, 1, 1, 0, 0, 1, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	invalid := []int8{1, -1, 2, 0, 1, -1}

	if _, err := NewTrits(invalid); err == nil {
		t.Error("expected NewTrits() return an error but got none")
	}

	if _, err := NewTrytes("A_AAA"); err == nil {
		t.Error("expected NewTrytes() return an error but got none")
	}

	var target int64 = 6562317
	i := TritsToInt(trits)
	if i != target {
		t.Errorf("expected Int() to return %d but got %d", target, i)
	}

	st := Trits{0, 1, -1, 1, 1, -1, -1, 1, 1, 0, 0, 1, 0, 1, 1}
	trits2 := IntToTrits(target, 15)
	if !TritsEqual(st, trits2) {
		t.Error("IntToTrits() is illegal.", trits2)
	}

	trits22 := IntToTrits(-1024, 7)
	if !TritsEqual(trits22, Trits{-1, 1, 0, 1, -1, -1, -1}) {
		t.Error("IntToTrits() is illegal.")
	}

	try := MustTritsToTrytes(st)
	if try != "UVKIL" {
		t.Error("Int() is illegal.", try)
	}

	trits3 := TrytesToTrits(try)
	if !TritsEqual(st, trits3) {
		t.Error("Trits() is illegal.", trits3)
	}
}

func TestTrytes_Normalize(t *testing.T) {
	var bundleHash Trytes = "DEXRPLKGBROUQMKCLMRPG9HFKCACDZ9AB9HOJQWERTYWERJNOYLW9PKLOGDUPC9DLGSUH9UHSKJOASJRU"
	no := []int8{-13, -13, -13, -13, -11, 12, 11, 7, 2, -9, -12, -6, -10, 13, 11, 3, 12, 13, -9, -11, 7, 0, 8, 6,
		11, 3, 1, 13, 13, 13, 7, 1, 2, 0, 8, -12, 10, -10, -4, 5, -9, -7, -2, -4, 5, -9, 10, -13, -12, -2, 12, -4,
		0, -11, -5, 12, -12, 7, 4, -6, -11, 3, 0, 4, 12, 7, -8, -6, 8, 0, -6, 8, -8, 11, 10, -12, 1, -8, 10, -9, -6}

	norm := Normalize(bundleHash)
	for i := range no {
		if no[i] != norm[i] {
			t.Fatal("normalization is incorrect.")
		}
	}
}

func TestTrits_Value(t *testing.T) {
	trits := Trits{1, 1, 1, 0, 0, 0, 1, -1, 1, -1, 1}
	const expected = 44482
	x := TritsToInt(trits)
	if x != expected {
		t.Fatalf("expected value %d but got %d\n", expected, x)
	}
}
