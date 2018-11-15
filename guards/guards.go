// Package guards provides validation functions which are used throughout the entire library.
package guards

import (
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/curl"
	. "github.com/iotaledger/iota.go/trinary"

	"regexp"
)

// IsTrytes checks if input is correct trytes consisting of [9A-Z]
func IsTrytes(trytes Trytes) bool {
	if len(trytes) < 1 {
		return false
	}
	match, _ := regexp.MatchString("^[9A-Z]+$", string(trytes))
	return match
}

// IsTrytesOfExactLength checks if input is correct trytes consisting of [9A-Z] and given length
func IsTrytesOfExactLength(trytes Trytes, length int) bool {
	if len(trytes) != length {
		return false
	}
	match, _ := regexp.MatchString("^[9A-Z]+$", string(trytes))
	return match
}

// IsTrytesOfMaxLength checks if input is correct trytes consisting of [9A-Z] and length <= maxLength
func IsTrytesOfMaxLength(trytes Trytes, max int) bool {
	if len(trytes) > max || len(trytes) < 1 {
		return false
	}
	match, _ := regexp.MatchString("^[9A-Z]+$", string(trytes))
	return match
}

var onlyNinesRegex = regexp.MustCompile("^[9]+$")

// IsEmptyTrytes checks if input is null (all 9s trytes)
func IsEmptyTrytes(trytes Trytes) bool {
	return onlyNinesRegex.MatchString(string(trytes))
}

// IsHash checks if input is correct hash (81 trytes or 90)
func IsHash(trytes Trytes) bool {
	return IsTrytesOfExactLength(trytes, HashTrytesSize) || IsTrytesOfExactLength(trytes, AddressWithChecksumTrytesSize)
}

// IsTransactionHash checks whether the given trytes can be a transaction hash.
func IsTransactionHash(trytes Trytes) bool {
	return IsTrytesOfExactLength(trytes, HashTrytesSize)
}

// IsTag checks that input is valid tag trytes.
func IsTag(trytes Trytes) bool {
	return IsTrytesOfExactLength(trytes, TagTrinarySize/3)
}

// IsTransactionHashWithMWM checks if input is correct transaction hash (81 trytes) with given MWM
func IsTransactionHashWithMWM(trytes Trytes, mwm uint) bool {
	correctLength := IsTrytesOfExactLength(trytes, HashTrytesSize)
	if !correctLength {
		return false
	}

	trits := MustTrytesToTrits(trytes)
	for _, trit := range trits[len(trits)-int(mwm):] {
		if trit != 0 {
			return false
		}
	}
	return true
}

// IsTransactionTrytes checks if input is correct transaction trytes (2673 trytes)
func IsTransactionTrytes(trytes Trytes) bool {
	return IsTrytesOfExactLength(trytes, TransactionTrytesSize)
}

// IsTransactionTrytesWithMWM checks if input is correct transaction trytes (2673 trytes) with given MWM
func IsTransactionTrytesWithMWM(trytes Trytes, mwm uint) (bool, error) {
	correctLength := IsTrytesOfExactLength(trytes, TransactionTrytesSize)
	if !correctLength {
		return false, nil
	}

	trits, err := TrytesToTrits(trytes)
	if err != nil {
		return false, err
	}

	hashTrits, err := curl.HashTrits(trits)
	if err != nil {
		return false, err
	}

	for _, trit := range hashTrits[len(hashTrits)-int(mwm):] {
		if trit != 0 {
			return false, nil
		}
	}
	return true, nil
}

// IsAttachedTrytes checks if input is valid attached transaction trytes. For attached transactions last 243 trytes are non-zero.
func IsAttachedTrytes(trytes Trytes) bool {
	return IsTrytesOfExactLength(trytes, TransactionTrytesSize) && !IsEmptyTrytes(trytes[(TransactionTrytesSize)-3*HashTrytesSize:])
}
