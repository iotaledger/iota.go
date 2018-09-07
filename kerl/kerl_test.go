package kerl

import (
	"github.com/iotaledger/giota/curl"
	"github.com/iotaledger/giota/trinary"
	"testing"
)

func TestNewKerl(t *testing.T) {
	k := NewKerl()
	if k == nil {
		t.Error("could not initialize kerl instance")
	}
}

func TestKerl(t *testing.T) {
	tests := []struct {
		name           string
		trytes         trinary.Trytes
		expectedTrytes trinary.Trytes
		squeezeSize    int
	}{
		{
			name:           "test squeeze HashSize",
			trytes:         trinary.Trytes("EMIDYNHBWMBCXVDEFOFWINXTERALUKYYPPHKP9JJFGJEIUY9MUDVNFZHMMWZUYUSWAIOWEVTHNWMHANBH"),
			expectedTrytes: trinary.Trytes("EJEAOOZYSAWFPZQESYDHZCGYNSTWXUMVJOVDWUNZJXDGWCLUFGIMZRMGCAZGKNPLBRLGUNYWKLJTYEAQX"),
			squeezeSize:    curl.HashSize,
		},
		{
			name:           "test squeeze HashSize * 2",
			trytes:         trinary.Trytes("9MIDYNHBWMBCXVDEFOFWINXTERALUKYYPPHKP9JJFGJEIUY9MUDVNFZHMMWZUYUSWAIOWEVTHNWMHANBH"),
			expectedTrytes: trinary.Trytes("G9JYBOMPUXHYHKSNRNMMSSZCSHOFYOYNZRSZMAAYWDYEIMVVOGKPJBVBM9TDPULSFUNMTVXRKFIDOHUXXVYDLFSZYZTWQYTE9SPYYWYTXJYQ9IFGYOLZXWZBKWZN9QOOTBQMWMUBLEWUEEASRHRTNIQWJQNDWRYLCA"),
			squeezeSize:    curl.HashSize * 2,
		},
		{
			name:           "test longer trytes with HashSize * 2",
			trytes:         trinary.Trytes("G9JYBOMPUXHYHKSNRNMMSSZCSHOFYOYNZRSZMAAYWDYEIMVVOGKPJBVBM9TDPULSFUNMTVXRKFIDOHUXXVYDLFSZYZTWQYTE9SPYYWYTXJYQ9IFGYOLZXWZBKWZN9QOOTBQMWMUBLEWUEEASRHRTNIQWJQNDWRYLCA"),
			expectedTrytes: trinary.Trytes("LUCKQVACOGBFYSPPVSSOXJEKNSQQRQKPZC9NXFSMQNRQCGGUL9OHVVKBDSKEQEBKXRNUJSRXYVHJTXBPDWQGNSCDCBAIRHAQCOWZEBSNHIJIGPZQITIBJQ9LNTDIBTCQ9EUWKHFLGFUVGGUWJONK9GBCDUIMAYMMQX"),
			squeezeSize:    curl.HashSize * 2,
		},
	}

	for _, tt := range tests {
		k := NewKerl()
		if k == nil {
			t.Errorf("could not initialize Kerl instance")

		}

		err := k.Absorb(tt.trytes.Trits())
		if err != nil {
			t.Errorf("Absorb(%q) failed: %s", tt.trytes, err)
		}

		ts, err := k.Squeeze(tt.squeezeSize)
		if err != nil {
			t.Errorf("Squeeze() failed: %s", err)
		}

		if ts.Trytes() != tt.expectedTrytes {
			if err != nil {
				t.Errorf("%s: tryte output: %s != expected output: %s", tt.name, ts.Trytes(), tt.expectedTrytes)
			}
		}
	}
}
