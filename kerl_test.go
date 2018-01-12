package giota

import (
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
		trytes         Trytes
		expectedTrytes Trytes
		squeezeSize    int
	}{
		{
			name:           "test squeeze HashSize",
			trytes:         Trytes("EMIDYNHBWMBCXVDEFOFWINXTERALUKYYPPHKP9JJFGJEIUY9MUDVNFZHMMWZUYUSWAIOWEVTHNWMHANBH"),
			expectedTrytes: Trytes("EJEAOOZYSAWFPZQESYDHZCGYNSTWXUMVJOVDWUNZJXDGWCLUFGIMZRMGCAZGKNPLBRLGUNYWKLJTYEAQX"),
			squeezeSize:    HashSize,
		},
		{
			name:           "test squeeze HashSize * 2",
			trytes:         Trytes("9MIDYNHBWMBCXVDEFOFWINXTERALUKYYPPHKP9JJFGJEIUY9MUDVNFZHMMWZUYUSWAIOWEVTHNWMHANBH"),
			expectedTrytes: Trytes("G9JYBOMPUXHYHKSNRNMMSSZCSHOFYOYNZRSZMAAYWDYEIMVVOGKPJBVBM9TDPULSFUNMTVXRKFIDOHUXXVYDLFSZYZTWQYTE9SPYYWYTXJYQ9IFGYOLZXWZBKWZN9QOOTBQMWMUBLEWUEEASRHRTNIQWJQNDWRYLCA"),
			squeezeSize:    HashSize * 2,
		},
		{
			name:           "test longer trytes with HashSize * 2",
			trytes:         Trytes("G9JYBOMPUXHYHKSNRNMMSSZCSHOFYOYNZRSZMAAYWDYEIMVVOGKPJBVBM9TDPULSFUNMTVXRKFIDOHUXXVYDLFSZYZTWQYTE9SPYYWYTXJYQ9IFGYOLZXWZBKWZN9QOOTBQMWMUBLEWUEEASRHRTNIQWJQNDWRYLCA"),
			expectedTrytes: Trytes("LUCKQVACOGBFYSPPVSSOXJEKNSQQRQKPZC9NXFSMQNRQCGGUL9OHVVKBDSKEQEBKXRNUJSRXYVHJTXBPDWQGNSCDCBAIRHAQCOWZEBSNHIJIGPZQITIBJQ9LNTDIBTCQ9EUWKHFLGFUVGGUWJONK9GBCDUIMAYMMQX"),
			squeezeSize:    HashSize * 2,
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
