package giota

func ValidTryte(t rune) bool {
	return ('A' <= t && t <= 'Z') || t == '9'
}

func ValidTrytes(ts string) bool {
	for _, t := range ts {
		if !ValidTryte(t) {
			return false
		}
	}

	return true
}

func ValidTransactionTrytes(ts string) bool {
	return ValidTrytes(ts) && len(ts) == TransactionTryteSize
}

func ValidTransactionTrits(ts []int) bool {
	return ValidTrits(ts) && len(ts) == TransactionTritSize
}

func ValidTrit(t int) bool {
	return t >= MinTritValue && t <= MaxTritValue
}

func ValidTrits(ts []int) bool {
	for _, t := range ts {
		if !ValidTrit(t) {
			return false
		}
	}

	return true
}

func ValidAddressTrytes(ts string) bool {
	return (len(ts) == 81 || len(ts) == 90) && ValidTrytes(ts)
}

func ValidAddressTrits(tr []int) bool {
	return (len(tr) == 81 || len(tr) == 90) && ValidTrits(tr)
}

func EqualTrits(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	for i, _ := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
