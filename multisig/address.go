package multisig

import (
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/signing/utils"
	. "github.com/iotaledger/iota.go/trinary"
)

// NewMultisigAddress creates a new multisig address object.
func NewMultisigAddress(digests Trytes, spongeFunc ...SpongeFunction) (*MultisigAddress, error) {
	h := GetSpongeFunc(spongeFunc, defaultCreator)

	m := &MultisigAddress{h: h}
	if len(digests) != 0 {
		if err := m.Absorb(digests); err != nil {
			return nil, err
		}
	}
	return m, nil
}

// MultisigAddress represents a multisig address.
type MultisigAddress struct {
	h SpongeFunction
}

// Absorb absorbs the given key digests.
func (m *MultisigAddress) Absorb(digests ...Trytes) error {
	for i := range digests {
		if err := m.h.AbsorbTrytes(digests[i]); err != nil {
			return err
		}
	}
	return nil
}

// Finalize finalizes and returns the multisig address as trytes.
func (m *MultisigAddress) Finalize(digest *string) (Trytes, error) {
	if digest != nil {
		if err := m.Absorb(*digest); err != nil {
			return "", err
		}
	}

	addressTrytes, err := m.h.SqueezeTrytes(HashTrinarySize)
	if err != nil {
		return "", err
	}

	return addressTrytes, nil
}
