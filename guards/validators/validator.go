package validators

import (
	"github.com/iotaledger/iota.go/bundle"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/guards"
	. "github.com/iotaledger/iota.go/trinary"
	"github.com/pkg/errors"
	"net/url"
)

type Validatable = func() error

func Validate(validators ...Validatable) error {
	for i := range validators {
		if err := validators[i](); err != nil {
			return err
		}
	}
	return nil
}

func ValidateNonEmptyStrings(err error, slice ...string) Validatable {
	return func() error {
		if slice == nil || len(slice) == 0 {
			return err
		}
		return nil
	}
}

func ValidateTransactionHashes(hashes ...Hash) Validatable {
	return func() error {
		for i := range hashes {
			if !IsTransactionHash(hashes[i]) {
				return errors.Wrapf(ErrInvalidTransactionHash, "%s at index %d", hashes[i], i)
			}
		}
		return nil
	}
}

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

func ValidateTransactionTrytes(trytes ...Trytes) Validatable {
	return func() error {
		for i := range trytes {
			if !IsTransactionTrytes(trytes[i]) {
				return errors.Wrapf(ErrInvalidTransactionTrytes, "at index %d", i)
			}
		}
		return nil
	}
}

func ValidateAttachedTransactionTrytes(trytes ...Trytes) Validatable {
	return func() error {
		for i := range trytes {
			if !IsAttachedTrytes(trytes[i]) {
				return errors.Wrapf(ErrInvalidAttachedTrytes, "at index %d", i)
			}
		}
		return nil
	}
}

func ValidateTags(tags ...Trytes) Validatable {
	return func() error {
		for i := range tags {
			if !IsTag(tags[i]) {
				return errors.Wrapf(ErrInvalidTag, "%s at index %d", tags[i], i)
			}
		}
		return nil
	}
}

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
			if _, err := url.Parse(uri[6:]); err != nil {
				return errors.Wrapf(ErrInvalidURI, "%s at index %d", uris[i], i)
			}
		}
		return nil
	}
}

func ValidateSecurityLevel(secLvl SecurityLevel) Validatable {
	return func() error {
		if secLvl > 3 || secLvl < 1 {
			return ErrInvalidSecurityLevel
		}
		return nil
	}
}

func ValidateSeed(seed Trytes) Validatable {
	return func() error {
		if !IsTrytesOfExactLength(seed, HashTrytesSize) {
			return ErrInvalidSeed
		}
		return nil
	}
}

const MaxIndexDiff = 1000

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

func ValidateTransfers(transfers ...bundle.Transfer) Validatable {
	return func() error {
		for i := range transfers {
			transfer := &transfers[i]
			if IsHash(transfer.Address) &&
				(transfer.Message == "" || IsTrytes(transfer.Message)) ||
				(transfer.Tag == "" || IsTag(transfer.Tag)) {
				continue
			}
			return ErrInvalidTransfer
		}
		return nil
	}
}
