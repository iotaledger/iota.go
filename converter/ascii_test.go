package converter

import (
	. "github.com/iotaledger/iota.go/trinary"
	"testing"
)

func TestASCIIToTrytes(t *testing.T) {
	const ascii = "IOTA"
	const utf8 = "Γιώτα"
	const expected = "SBYBCCKB"

	trytes, err := ASCIIToTrytes(ascii)
	if err != nil {
		t.Fatalf("didn't expect an error for valid ascii input but got error: %v\n", err)
	}

	for i := range trytes {
		if trytes[i] != expected[i] {
			t.Fatalf("char at %d is %v but expected %v\n", i, trytes[i], expected[i])
		}
	}

	if _, err := ASCIIToTrytes(utf8); err == nil {
		t.Fatalf("expected an error for the invalid ascii input of %v", utf8)
	}
}

func TestTrytesToASCII(t *testing.T) {
	const trytes = Trytes("SBYBCCKB")
	const expected = "IOTA"

	asciiVal, err := TrytesToASCII(trytes)
	if err != nil {
		t.Fatalf("didn't expect an error for valid tryte values for ascii conversion but got error: %v\n", err)
	}

	if asciiVal != expected {
		t.Fatalf("got converted ascii value %s but expected %s\n", asciiVal, expected)
	}

	const invalidTrytes = Trytes("AAAfasds")
	const trytesWithOddLength = Trytes("AAA")

	_, err = TrytesToASCII(invalidTrytes)
	if err == nil {
		t.Fatalf("expected an error for non convertible tryte value %s", invalidTrytes)
	}

	if err != ErrInvalidTryteCharacter {
		t.Fatalf("expected invalid tryte char error but got: %v", err)
	}

	_, err = TrytesToASCII(trytesWithOddLength)
	if err == nil {
		t.Fatalf("expected an error for non convertible tryte value %s", trytesWithOddLength)
	}

	if err != ErrInvalidLengthForToASCIIConversion {
		t.Fatalf("expected invalid trytes length error but got: %v", err)
	}
}
