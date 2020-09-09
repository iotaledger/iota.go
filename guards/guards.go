// Package guards provides validation functions which are used throughout the entire library.
package guards

import (
	. "github.com/iotaledger/iota.go/consts"
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