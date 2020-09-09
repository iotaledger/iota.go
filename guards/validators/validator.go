// Package validators leverages package guards to provide easy to use validation functions.
package validators

import (
	"net/url"

	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/guards"
	. "github.com/iotaledger/iota.go/trinary"
	"github.com/pkg/errors"
)

// Validatable is a function which validates something and returns an error if the validation fails.
type Validatable = func() error

// Validate calls all given validators or returns the first occurred error.
func Validate(validators ...Validatable) error {
	for i := range validators {
		if err := validators[i](); err != nil {
			return err
		}
	}
	return nil
}

// ValidateNonEmptyStrings checks for non empty string slices.
func ValidateNonEmptyStrings(err error, slice ...string) Validatable {
	return func() error {
		if slice == nil || len(slice) == 0 {
			return err
		}
		return nil
	}
}

// ValidateHashes validates the given hashes.
func ValidateHashes(hashes ...Hash) Validatable {
	return func() error {
		for i := range hashes {
			if !IsHash(hashes[i]) {
				return errors.Wrapf(ErrInvalidHash, "%s at index %d", hashes[i], i)
			}
		}
		return nil
	}
}

// ValidateURIs validates the given URIs for neighbor addition/removal.
func ValidateURIs(uris ...string) Validatable {
	return func() error {
		for i := range uris {
			uri := uris[i]
			if len(uri) < 7 {
				return errors.Wrapf(ErrInvalidURI, "%s at index %d", uris[i], i)
			}
			schema := uri[:6]
			if schema != "tcp://" && schema != "udp://" {
				return errors.Wrapf(ErrInvalidURI, "%s at index %d", uris[i], i)
			}
			if _, err := url.Parse(uri); err != nil {
				return errors.Wrapf(ErrInvalidURI, "%s at index %d", uris[i], i)
			}
		}
		return nil
	}
}

// ValidateSecurityLevel validates the given security level.
func ValidateSecurityLevel(secLvl SecurityLevel) Validatable {
	return func() error {
		if secLvl > 3 || secLvl < 1 {
			return ErrInvalidSecurityLevel
		}
		return nil
	}
}

// ValidateSeed validates the given seed.
func ValidateSeed(seed Trytes) Validatable {
	return func() error {
		if !IsTrytesOfExactLength(seed, HashTrytesSize) {
			return ErrInvalidSeed
		}
		return nil
	}
}

// MaxIndexDiff is the max delta between start and end options.
const MaxIndexDiff = 1000

// ValidateStartEndOptions validates the given start and optional end option.
func ValidateStartEndOptions(start uint64, end *uint64) Validatable {
	return func() error {
		if end == nil {
			return nil
		}
		e := *end
		if start > e || e > start+MaxIndexDiff {
			return ErrInvalidStartEndOptions
		}
		return nil
	}
}
