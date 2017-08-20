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
	k := NewKerl()
	if k == nil {
		t.Error("could not initialize kerl instance")
	}

	tr := Trytes("EMIDYNHBWMBCXVDEFOFWINXTERALUKYYPPHKP9JJFGJEIUY9MUDVNFZHMMWZUYUSWAIOWEVTHNWMHANBH")
	err := k.Absorb(tr.Trits())
	if err != nil {
		t.Errorf("Absorb(%q) failed: %s", tr, err)
	}

	ts, err := k.Squeeze(HashSize)
	if err != nil {
		t.Errorf("Squeeze() failed: %s", err)
	}

	h1Out := Trytes("EJEAOOZYSAWFPZQESYDHZCGYNSTWXUMVJOVDWUNZJXDGWCLUFGIMZRMGCAZGKNPLBRLGUNYWKLJTYEAQX")
	if ts.Trytes() != h1Out {
		t.Errorf("Expected %q but got %q", h1Out, ts.Trytes())
	}

	k = NewKerl()
	tr = Trytes("9MIDYNHBWMBCXVDEFOFWINXTERALUKYYPPHKP9JJFGJEIUY9MUDVNFZHMMWZUYUSWAIOWEVTHNWMHANBH")
	err = k.Absorb(tr.Trits())
	if err != nil {
		t.Errorf("Absorb(%q) failed: %s", tr, err)
	}

	ts, err = k.Squeeze(HashSize * 2)
	if err != nil {
		t.Errorf("Squeeze() failed: %s", err)
	}

	h2Out := Trytes("G9JYBOMPUXHYHKSNRNMMSSZCSHOFYOYNZRSZMAAYWDYEIMVVOGKPJBVBM9TDPULSFUNMTVXRKFIDOHUXXVYDLFSZYZTWQYTE9SPYYWYTXJYQ9IFGYOLZXWZBKWZN9QOOTBQMWMUBLEWUEEASRHRTNIQWJQNDWRYLCA")
	if ts.Trytes() != h2Out {
		t.Errorf("Expected %q but got %q", h2Out, ts.Trytes())
	}

	k = NewKerl()
	tr = Trytes("G9JYBOMPUXHYHKSNRNMMSSZCSHOFYOYNZRSZMAAYWDYEIMVVOGKPJBVBM9TDPULSFUNMTVXRKFIDOHUXXVYDLFSZYZTWQYTE9SPYYWYTXJYQ9IFGYOLZXWZBKWZN9QOOTBQMWMUBLEWUEEASRHRTNIQWJQNDWRYLCA")
	err = k.Absorb(tr.Trits())
	if err != nil {
		t.Errorf("Absorb(%q) failed: %s", tr, err)
	}

	ts, err = k.Squeeze(HashSize * 2)
	if err != nil {
		t.Errorf("Squeeze() failed: %s", err)
	}

	h3Out := Trytes("LUCKQVACOGBFYSPPVSSOXJEKNSQQRQKPZC9NXFSMQNRQCGGUL9OHVVKBDSKEQEBKXRNUJSRXYVHJTXBPDWQGNSCDCBAIRHAQCOWZEBSNHIJIGPZQITIBJQ9LNTDIBTCQ9EUWKHFLGFUVGGUWJONK9GBCDUIMAYMMQX")
	if ts.Trytes() != h3Out {
		t.Errorf("Expected %q but got %q", h3Out, ts.Trytes())
	}
}
