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

func ValidHashTrytes(ts string) bool {
	return (len(ts) == 81 || len(ts) == 90) && ValidTrytes(ts)
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
