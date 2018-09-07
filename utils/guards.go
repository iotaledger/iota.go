package utils

import (
	"github.com/iotaledger/giota/curl"
	"github.com/iotaledger/giota/signing"
	"github.com/iotaledger/giota/transaction"
	"github.com/iotaledger/giota/trinary"

	"regexp"
)

// Checks if input is correct trytes consisting of [9A-Z] and given length
func IsTrytesOfExactLength(trytes trinary.Trytes, length int) bool {
	if len(trytes) != length {
		return false
	}
	match, _ := regexp.MatchString("^[9A-Z]+$", string(trytes))
	return match
}

// Checks if input is correct trytes consisting of [9A-Z] and length <= maxLength
func IsTrytesOfMaxLength(trytes trinary.Trytes, max int) bool {
	if len(trytes) > max {
		return false
	}
	match, _ := regexp.MatchString("^[9A-Z]+$", string(trytes))
	return match
}

var onlyNinesRegex = regexp.MustCompile("^[9]+$")

// Checks if input is null (all 9s trytes)
func IsEmptyTrytes(trytes trinary.Trytes) bool {
	return onlyNinesRegex.MatchString(string(trytes))
}

// alias
var IsNineTrytes = IsEmptyTrytes

// Checks if input is correct hash (81 trytes)
func IsHash(trytes trinary.Trytes) bool {
	return IsTrytesOfExactLength(trytes, curl.HashSizeTrytes) || IsTrytesOfExactLength(trytes, curl.HashSizeTrytes+signing.ChecksumSize)
}

// Checks that input is valid tag trytes.
func IsTag(trytes trinary.Trytes) bool {
	return IsTrytesOfExactLength(trytes, transaction.TagTrinarySize/3)
}

// Checks if input is correct transaction hash (81 trytes)
func IsTxHash(trytes trinary.Trytes) bool {
	return IsTrytesOfExactLength(trytes, curl.HashSizeTrytes)
}

// Checks if input is correct transaction hash (81 trytes) with given MWM
func IsTxHashWithMWM(trytes trinary.Trytes, mwm uint) bool {
	correctLength := IsTrytesOfExactLength(trytes, curl.HashSizeTrytes)
	if !correctLength {
		return false
	}

	trits := trytes.Trits()
	for _, trit := range trits[len(trits)-int(mwm):] {
		if trit != 0 {
			return false
		}
	}
	return true
}

// Checks if input is correct transaction trytes (2673 trytes)
func IsTransactionTrytes(trytes trinary.Trytes) bool {
	return IsTrytesOfExactLength(trytes, transaction.TransactionTrinarySize/3)
}

// Checks if input is correct transaction trytes (2673 trytes) with given MWM
func IsTransactionTrytesWithMWM(trytes trinary.Trytes, mwm uint) (bool, error) {
	correctLength := IsTrytesOfExactLength(trytes, transaction.TransactionTrinarySize/3)
	if !correctLength {
		return false, nil
	}

	tx, err := transaction.NewTransaction(trytes)
	if err != nil {
		return false, err
	}
	hashTrits := tx.Hash().Trits()
	for _, trit := range hashTrits[len(hashTrits)-int(mwm):] {
		if trit != 0 {
			return false, nil
		}
	}
	return true, nil
}

// Checks if input is valid attached transaction trytes. For attached transactions last 241 trytes are non-zero.
func IsAttachedTrytes(trytes trinary.Trytes) bool {
	return IsTrytesOfExactLength(trytes, transaction.TransactionTrinarySize/3) && !IsEmptyTrytes(trytes[len(trytes)-3*curl.HashSizeTrytes:])
}
