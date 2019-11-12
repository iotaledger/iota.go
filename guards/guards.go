// Package guards provides validation functions which are used throughout the entire library.
package guards

import (
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/curl"
	. "github.com/iotaledger/iota.go/trinary"

)

// IsTrytes checks if input is correct trytes consisting of [9A-Z]
func IsTrytes(trytes Trytes) bool {
	if len(trytes) < 1 {
		return false
	}
	for _, runeVal := range trytes {
		if (runeVal < 'A' || runeVal > 'Z') && runeVal != '9' {
			return false
		}
	}
	return true
}

// IsTrytesOfExactLength checks if input is correct trytes consisting of [9A-Z] and given length
func IsTrytesOfExactLength(trytes Trytes, length int) bool {
	if len(trytes) != length || len(trytes) == 0 {
		return false
	}
        for _, runeVal := range trytes {
                if (runeVal < 'A' || runeVal > 'Z') && runeVal != '9' {
                        return false
                }
        }
        return true
}

// IsTrytesOfMaxLength checks if input is correct trytes consisting of [9A-Z] and length <= maxLength
func IsTrytesOfMaxLength(trytes Trytes, max int) bool {
	if len(trytes) > max || len(trytes) < 1 {
		return false
	}
        for _, runeVal := range trytes {
                if (runeVal < 'A' || runeVal > 'Z') && runeVal != '9' {
                        return false
                }
        }
        return true
}

// IsEmptyTrytes checks if input is null (all 9s trytes)
func IsEmptyTrytes(trytes Trytes) bool {
        for _, runeVal := range trytes {
                if runeVal != '9' {
                        return false
                }
        }

        return true
}

// IsHash checks if input is correct hash (81 trytes or 90)
func IsHash(trytes Trytes) bool {
	return IsTrytesOfExactLength(trytes, HashTrytesSize) || IsTrytesOfExactLength(trytes, AddressWithChecksumTrytesSize)
}

// IsAddressWithChecksum checks if the given address is exactly 90 trytes long.
func IsAddressWithChecksum(addr Trytes) bool {
	return IsTrytesOfExactLength(addr, AddressWithChecksumTrytesSize)
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

// IsAttachedTrytes checks if input is valid attached transaction trytes.
// For attached transactions the last 243 trytes are non-zero.
func IsAttachedTrytes(trytes Trytes) bool {
	return IsTrytesOfExactLength(trytes, TransactionTrytesSize) && !IsEmptyTrytes(trytes[(TransactionTrytesSize)-3*HashTrytesSize:])
}
