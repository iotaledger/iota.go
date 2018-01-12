/*
MIT License

Copyright (c) 2016 Sascha Hanse
Copyright (c) 2017 Shinya Yagyu

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package giota

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
		validTryteTC{in: 'A', valid: true},
		validTryteTC{in: 'Z', valid: true},
		validTryteTC{in: '9', valid: true},
		validTryteTC{in: '8', valid: false},
		validTryteTC{in: 'a', valid: false},
		validTryteTC{in: '-', valid: false},
		validTryteTC{in: 'Ɩ', valid: false},
	}

	for _, tc := range validTryteCases {
		if (IsValidTryte(tc.in) == nil) != tc.valid {
			t.Fatalf("ValidTryte(%q) should be %#v but is not", tc.in, tc.valid)
		}
	}
}

func TestValidTrytes(t *testing.T) {
	type validTryteTC struct {
		in    Trytes
		valid bool
	}

	var validTryteCases = []validTryteTC{
		validTryteTC{in: "ABCDEFGHIJKLMNOPQRSTUVWXYZ9", valid: true},
		validTryteTC{in: "ABCDEFGHIJKLMNOPQRSTUVWXYZ90", valid: false},
		validTryteTC{in: "ABCDEFGHIJKLMNOPQRSTUVWXYZ9 ", valid: false},
		validTryteTC{in: "Ɩ", valid: false},
	}

	for _, tc := range validTryteCases {
		if (tc.in.IsValid() == nil) != tc.valid {
			t.Fatalf("ValidTrytes(%q) should be %#v but is not", tc.in, tc.valid)
		}
	}
}

func TestValidTrit(t *testing.T) {
	type validTritTC struct {
		in    int8
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
		if (IsValidTrit(tc.in) == nil) != tc.valid {
			t.Fatalf("ValidTrit(%q) should be %#v but is not", tc.in, tc.valid)
		}
	}
}

func TestValidTrits(t *testing.T) {
	type validTritsTC struct {
		in    Trits
		valid bool
	}

	var validTritsCases = []validTritsTC{
		validTritsTC{in: Trits{0}, valid: true},
		validTritsTC{in: Trits{-1}, valid: true},
		validTritsTC{in: Trits{1}, valid: true},
		validTritsTC{in: Trits{0, -1, 1}, valid: true},
		validTritsTC{in: Trits{2, -1, 1}, valid: false},
	}

	for _, tc := range validTritsCases {
		if (tc.in.IsValid() == nil) != tc.valid {
			t.Fatalf("ValidTrits(%q) should be %#v but is not", tc.in, tc.valid)
		}
	}
}

func TestTritByteTrit(t *testing.T) {
	ts := []Trits{
		Trytes("NNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNN").Trits(),
		Trytes("MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM").Trits(),
		Trytes("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA").Trits(),
		Trytes("ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ").Trits(),
		Trytes("999999999999999999999999999999999999999999999999999999999999999999999999999999999").Trits(),
		Trytes("SCYLJDWIM9LIXCSLETSHLQOOFDKYOVFZLAHQYCCNMYHRTNIKBZRIRACFYPOWYNSOWDNXFZUG9OEOZPOTD").Trits(),
	}

	for _, trits := range ts {
		trits[TritHashLength-1] = 0
		b, err := trits.Bytes()
		if err != nil {
			t.Errorf("Bytes() failed: %s", err)
		}

		tb, err := BytesToTrits(b)
		if err != nil {
			t.Errorf("BytesToTrits() failed: %s", err)
		}

		if !tb.Equal(trits) {
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

		bs, err = fst.Bytes()
		if err != nil {
			t.Errorf("Bytes() failed: %s", err)
		}

		snd, err := BytesToTrits(bs)
		if err != nil {
			t.Errorf("BytesToTrits() failed: %s", err)
		}

		if !fst.Equal(snd) {
			t.Errorf("Bytes->Trits->Bytes->Trits roundtrip failed\nwanted:\t%#v\ngot:\t%#v", fst, snd)
		}
	}
}

func TestConvert(t *testing.T) {
	trits := Trits{0, 1, -1, 1, 1, -1, -1, 1, 1, 0, 0, 1, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	invalid := []int8{1, -1, 2, 0, 1, -1}

	if _, err := ToTrits(invalid); err == nil {
		t.Error("expected ToTrits() return an error but got none")
	}

	if _, err := ToTrytes("A_AAA"); err == nil {
		t.Error("expected ToTrytes() return an error but got none")
	}

	var target int64 = 6562317
	i := trits.Int()
	if i != target {
		t.Errorf("expected Int() to return %d but got %d", target, i)
	}

	st := Trits{0, 1, -1, 1, 1, -1, -1, 1, 1, 0, 0, 1, 0, 1, 1}
	trits2 := Int2Trits(target, 15)
	if !st.Equal(trits2) {
		t.Error("Int2Trits() is illegal.", trits2)
	}

	trits22 := Int2Trits(-1024, 7)
	if !trits22.Equal(Trits{-1, 1, 0, 1, -1, -1, -1}) {
		t.Error("Int2Trits() is illegal.")
	}

	try := st.Trytes()
	if try != "UVKIL" {
		t.Error("Int() is illegal.", try)
	}

	trits3 := try.Trits()
	if !st.Equal(trits3) {
		t.Error("Trits() is illegal.", trits3)
	}
}

func TestNormalize(t *testing.T) {
	var bundleHash Trytes = "DEXRPLKGBROUQMKCLMRPG9HFKCACDZ9AB9HOJQWERTYWERJNOYLW9PKLOGDUPC9DLGSUH9UHSKJOASJRU"
	no := []int8{-13, -13, -13, -13, -11, 12, 11, 7, 2, -9, -12, -6, -10, 13, 11, 3, 12, 13, -9, -11, 7, 0, 8, 6,
		11, 3, 1, 13, 13, 13, 7, 1, 2, 0, 8, -12, 10, -10, -4, 5, -9, -7, -2, -4, 5, -9, 10, -13, -12, -2, 12, -4,
		0, -11, -5, 12, -12, 7, 4, -6, -11, 3, 0, 4, 12, 7, -8, -6, 8, 0, -6, 8, -8, 11, 10, -12, 1, -8, 10, -9, -6}

	norm := bundleHash.Normalize()
	for i := range no {
		if no[i] != norm[i] {
			t.Fatal("normalization is incorrect.")
		}
	}
}
