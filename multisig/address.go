package multisig

import "github.com/iotaledger/iota.go/kerl"
import (
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/trinary"
)

// NewMultisigAddress creates a new multisig address object.
func NewMultisigAddress(digests ...Trytes) (*MultisigAddress, error) {
	m := &MultisigAddress{k: kerl.NewKerl()}
	if len(digests) != 0 {
		if err := m.Absorb(digests...); err != nil {
			return nil, err
		}
	}
	return m, nil
}

// MultisigAddress represents a multisig address.
type MultisigAddress struct {
	k *kerl.Kerl
}

// Absorb absorbs the given key digests.
func (m *MultisigAddress) Absorb(digests ...Trytes) error {
	for i := range digests {
		if err := m.k.Absorb(MustTrytesToTrits(digests[i])); err != nil {
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

	addressTrits, err := m.k.Squeeze(HashTrinarySize)
	if err != nil {
		return "", err
	}

	return MustTritsToTrytes(addressTrits), nil
}
