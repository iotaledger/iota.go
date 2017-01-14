package giota

import (
	"strings"
)

var (
	BYTE_TO_TRITS_MAPPINGS = [][]int{
		[]int{0, 0, 0, 0, 0}, []int{1, 0, 0, 0, 0}, []int{-1, 1, 0, 0, 0}, []int{0, 1, 0, 0, 0}, []int{1, 1, 0, 0, 0},
		[]int{-1, -1, 1, 0, 0}, []int{0, -1, 1, 0, 0}, []int{1, -1, 1, 0, 0}, []int{-1, 0, 1, 0, 0}, []int{0, 0, 1, 0, 0},
		[]int{1, 0, 1, 0, 0}, []int{-1, 1, 1, 0, 0}, []int{0, 1, 1, 0, 0}, []int{1, 1, 1, 0, 0}, []int{-1, -1, -1, 1, 0},
		[]int{0, -1, -1, 1, 0}, []int{1, -1, -1, 1, 0}, []int{-1, 0, -1, 1, 0}, []int{0, 0, -1, 1, 0}, []int{1, 0, -1, 1, 0},
		[]int{-1, 1, -1, 1, 0}, []int{0, 1, -1, 1, 0}, []int{1, 1, -1, 1, 0}, []int{-1, -1, 0, 1, 0}, []int{0, -1, 0, 1, 0},
		[]int{1, -1, 0, 1, 0}, []int{-1, 0, 0, 1, 0}, []int{0, 0, 0, 1, 0}, []int{1, 0, 0, 1, 0}, []int{-1, 1, 0, 1, 0},
		[]int{0, 1, 0, 1, 0}, []int{1, 1, 0, 1, 0}, []int{-1, -1, 1, 1, 0}, []int{0, -1, 1, 1, 0}, []int{1, -1, 1, 1, 0},
		[]int{-1, 0, 1, 1, 0}, []int{0, 0, 1, 1, 0}, []int{1, 0, 1, 1, 0}, []int{-1, 1, 1, 1, 0}, []int{0, 1, 1, 1, 0},
		[]int{1, 1, 1, 1, 0}, []int{-1, -1, -1, -1, 1}, []int{0, -1, -1, -1, 1}, []int{1, -1, -1, -1, 1}, []int{-1, 0, -1, -1, 1},
		[]int{0, 0, -1, -1, 1}, []int{1, 0, -1, -1, 1}, []int{-1, 1, -1, -1, 1}, []int{0, 1, -1, -1, 1}, []int{1, 1, -1, -1, 1},
		[]int{-1, -1, 0, -1, 1}, []int{0, -1, 0, -1, 1}, []int{1, -1, 0, -1, 1}, []int{-1, 0, 0, -1, 1}, []int{0, 0, 0, -1, 1},
		[]int{1, 0, 0, -1, 1}, []int{-1, 1, 0, -1, 1}, []int{0, 1, 0, -1, 1}, []int{1, 1, 0, -1, 1}, []int{-1, -1, 1, -1, 1},
		[]int{0, -1, 1, -1, 1}, []int{1, -1, 1, -1, 1}, []int{-1, 0, 1, -1, 1}, []int{0, 0, 1, -1, 1}, []int{1, 0, 1, -1, 1},
		[]int{-1, 1, 1, -1, 1}, []int{0, 1, 1, -1, 1}, []int{1, 1, 1, -1, 1}, []int{-1, -1, -1, 0, 1}, []int{0, -1, -1, 0, 1},
		[]int{1, -1, -1, 0, 1}, []int{-1, 0, -1, 0, 1}, []int{0, 0, -1, 0, 1}, []int{1, 0, -1, 0, 1}, []int{-1, 1, -1, 0, 1},
		[]int{0, 1, -1, 0, 1}, []int{1, 1, -1, 0, 1}, []int{-1, -1, 0, 0, 1}, []int{0, -1, 0, 0, 1}, []int{1, -1, 0, 0, 1},
		[]int{-1, 0, 0, 0, 1}, []int{0, 0, 0, 0, 1}, []int{1, 0, 0, 0, 1}, []int{-1, 1, 0, 0, 1}, []int{0, 1, 0, 0, 1},
		[]int{1, 1, 0, 0, 1}, []int{-1, -1, 1, 0, 1}, []int{0, -1, 1, 0, 1}, []int{1, -1, 1, 0, 1}, []int{-1, 0, 1, 0, 1},
		[]int{0, 0, 1, 0, 1}, []int{1, 0, 1, 0, 1}, []int{-1, 1, 1, 0, 1}, []int{0, 1, 1, 0, 1}, []int{1, 1, 1, 0, 1},
		[]int{-1, -1, -1, 1, 1}, []int{0, -1, -1, 1, 1}, []int{1, -1, -1, 1, 1}, []int{-1, 0, -1, 1, 1}, []int{0, 0, -1, 1, 1},
		[]int{1, 0, -1, 1, 1}, []int{-1, 1, -1, 1, 1}, []int{0, 1, -1, 1, 1}, []int{1, 1, -1, 1, 1}, []int{-1, -1, 0, 1, 1},
		[]int{0, -1, 0, 1, 1}, []int{1, -1, 0, 1, 1}, []int{-1, 0, 0, 1, 1}, []int{0, 0, 0, 1, 1}, []int{1, 0, 0, 1, 1},
		[]int{-1, 1, 0, 1, 1}, []int{0, 1, 0, 1, 1}, []int{1, 1, 0, 1, 1}, []int{-1, -1, 1, 1, 1}, []int{0, -1, 1, 1, 1},
		[]int{1, -1, 1, 1, 1}, []int{-1, 0, 1, 1, 1}, []int{0, 0, 1, 1, 1}, []int{1, 0, 1, 1, 1}, []int{-1, 1, 1, 1, 1},
		[]int{0, 1, 1, 1, 1}, []int{1, 1, 1, 1, 1}, []int{-1, -1, -1, -1, -1}, []int{0, -1, -1, -1, -1}, []int{1, -1, -1, -1, -1},
		[]int{-1, 0, -1, -1, -1}, []int{0, 0, -1, -1, -1}, []int{1, 0, -1, -1, -1}, []int{-1, 1, -1, -1, -1}, []int{0, 1, -1, -1, -1},
		[]int{1, 1, -1, -1, -1}, []int{-1, -1, 0, -1, -1}, []int{0, -1, 0, -1, -1}, []int{1, -1, 0, -1, -1}, []int{-1, 0, 0, -1, -1},
		[]int{0, 0, 0, -1, -1}, []int{1, 0, 0, -1, -1}, []int{-1, 1, 0, -1, -1}, []int{0, 1, 0, -1, -1}, []int{1, 1, 0, -1, -1},
		[]int{-1, -1, 1, -1, -1}, []int{0, -1, 1, -1, -1}, []int{1, -1, 1, -1, -1}, []int{-1, 0, 1, -1, -1}, []int{0, 0, 1, -1, -1},
		[]int{1, 0, 1, -1, -1}, []int{-1, 1, 1, -1, -1}, []int{0, 1, 1, -1, -1}, []int{1, 1, 1, -1, -1}, []int{-1, -1, -1, 0, -1},
		[]int{0, -1, -1, 0, -1}, []int{1, -1, -1, 0, -1}, []int{-1, 0, -1, 0, -1}, []int{0, 0, -1, 0, -1}, []int{1, 0, -1, 0, -1},
		[]int{-1, 1, -1, 0, -1}, []int{0, 1, -1, 0, -1}, []int{1, 1, -1, 0, -1}, []int{-1, -1, 0, 0, -1}, []int{0, -1, 0, 0, -1},
		[]int{1, -1, 0, 0, -1}, []int{-1, 0, 0, 0, -1}, []int{0, 0, 0, 0, -1}, []int{1, 0, 0, 0, -1}, []int{-1, 1, 0, 0, -1},
		[]int{0, 1, 0, 0, -1}, []int{1, 1, 0, 0, -1}, []int{-1, -1, 1, 0, -1}, []int{0, -1, 1, 0, -1}, []int{1, -1, 1, 0, -1},
		[]int{-1, 0, 1, 0, -1}, []int{0, 0, 1, 0, -1}, []int{1, 0, 1, 0, -1}, []int{-1, 1, 1, 0, -1}, []int{0, 1, 1, 0, -1},
		[]int{1, 1, 1, 0, -1}, []int{-1, -1, -1, 1, -1}, []int{0, -1, -1, 1, -1}, []int{1, -1, -1, 1, -1}, []int{-1, 0, -1, 1, -1},
		[]int{0, 0, -1, 1, -1}, []int{1, 0, -1, 1, -1}, []int{-1, 1, -1, 1, -1}, []int{0, 1, -1, 1, -1}, []int{1, 1, -1, 1, -1},
		[]int{-1, -1, 0, 1, -1}, []int{0, -1, 0, 1, -1}, []int{1, -1, 0, 1, -1}, []int{-1, 0, 0, 1, -1}, []int{0, 0, 0, 1, -1},
		[]int{1, 0, 0, 1, -1}, []int{-1, 1, 0, 1, -1}, []int{0, 1, 0, 1, -1}, []int{1, 1, 0, 1, -1}, []int{-1, -1, 1, 1, -1},
		[]int{0, -1, 1, 1, -1}, []int{1, -1, 1, 1, -1}, []int{-1, 0, 1, 1, -1}, []int{0, 0, 1, 1, -1}, []int{1, 0, 1, 1, -1},
		[]int{-1, 1, 1, 1, -1}, []int{0, 1, 1, 1, -1}, []int{1, 1, 1, 1, -1}, []int{-1, -1, -1, -1, 0}, []int{0, -1, -1, -1, 0},
		[]int{1, -1, -1, -1, 0}, []int{-1, 0, -1, -1, 0}, []int{0, 0, -1, -1, 0}, []int{1, 0, -1, -1, 0}, []int{-1, 1, -1, -1, 0},
		[]int{0, 1, -1, -1, 0}, []int{1, 1, -1, -1, 0}, []int{-1, -1, 0, -1, 0}, []int{0, -1, 0, -1, 0}, []int{1, -1, 0, -1, 0},
		[]int{-1, 0, 0, -1, 0}, []int{0, 0, 0, -1, 0}, []int{1, 0, 0, -1, 0}, []int{-1, 1, 0, -1, 0}, []int{0, 1, 0, -1, 0},
		[]int{1, 1, 0, -1, 0}, []int{-1, -1, 1, -1, 0}, []int{0, -1, 1, -1, 0}, []int{1, -1, 1, -1, 0}, []int{-1, 0, 1, -1, 0},
		[]int{0, 0, 1, -1, 0}, []int{1, 0, 1, -1, 0}, []int{-1, 1, 1, -1, 0}, []int{0, 1, 1, -1, 0}, []int{1, 1, 1, -1, 0},
		[]int{-1, -1, -1, 0, 0}, []int{0, -1, -1, 0, 0}, []int{1, -1, -1, 0, 0}, []int{-1, 0, -1, 0, 0}, []int{0, 0, -1, 0, 0},
		[]int{1, 0, -1, 0, 0}, []int{-1, 1, -1, 0, 0}, []int{0, 1, -1, 0, 0}, []int{1, 1, -1, 0, 0}, []int{-1, -1, 0, 0, 0},
		[]int{0, -1, 0, 0, 0}, []int{1, -1, 0, 0, 0}, []int{-1, 0, 0, 0, 0},
	}

	TRYTE_TO_TRITS_MAPPINGS = [][]int{
		[]int{0, 0, 0}, []int{1, 0, 0}, []int{-1, 1, 0}, []int{0, 1, 0},
		[]int{1, 1, 0}, []int{-1, -1, 1}, []int{0, -1, 1}, []int{1, -1, 1},
		[]int{-1, 0, 1}, []int{0, 0, 1}, []int{1, 0, 1}, []int{-1, 1, 1},
		[]int{0, 1, 1}, []int{1, 1, 1}, []int{-1, -1, -1}, []int{0, -1, -1},
		[]int{1, -1, -1}, []int{-1, 0, -1}, []int{0, 0, -1}, []int{1, 0, -1},
		[]int{-1, 1, -1}, []int{0, 1, -1}, []int{1, 1, -1}, []int{-1, -1, 0},
		[]int{0, -1, 0}, []int{1, -1, 0}, []int{-1, 0, 0},
	}
)

type Trit int
type Tryte byte

// tritsToIntWithOffset takes a slice of trits and converts them into an integer,
// starting at the given offset and using size entries.
// Assumes big-endian notation.
func tritsToIntWithOffset(trits []int, offset, size int) int64 {
	var val int64

	i := size
	for i > 0 {
		i -= 1
		val = val*3 + int64(trits[offset+i])
	}

	return val
}

func tritsToInt(trits []int) int64 {
	return tritsToIntWithOffset(trits, 0, len(trits))
}

//    public static String trytes(final int[] trits, final int offset, final int size) {
//        final StringBuilder trytes = new StringBuilder();
//        for (int i = 0; i < (size + 3 - 1) / 3; ++i) {
//            int j = trits[offset + i * 3] + trits[offset + i * 3 + 1] * 3 + trits[offset + i * 3 + 2] * 9;
//            if (j < 0) {
//                j += "9ABCDEFGHIJKLMNOPQRSTUVWXYZ".length();
//            }
//            trytes.append("9ABCDEFGHIJKLMNOPQRSTUVWXYZ".charAt(j));
//        }
//        return trytes.toString();
//    }
func tritsToTrytesWithOffset(trits []int, offset, size int) string {
	o := ""
	for i := 0; i < (size+3-1)/3; i += 1 {
		j := trits[offset+i*3] + trits[offset+i*3+1]*3 + trits[offset+i*3+2]*9
		if j < 0 {
			j += len(TryteAlphabet)
		}

		o = o + TryteAlphabet[j:j+1]
	}

	return o
}

func tritsToTrytes(trits []int) string {
	return tritsToTrytesWithOffset(trits, 0, len(trits))
}

func trytesToTrits(trytes string) []int {
	trits := make([]int, len(trytes)*3)
	for i, _ := range trytes {
		idx := strings.Index(TryteAlphabet, trytes[i:i+1])
		copy(trits[i*3:i*3+3], TRYTE_TO_TRITS_MAPPINGS[idx])
	}

	return trits
}

func bytesToTrits(b []int, size int) []int {
	trits := make([]int, size)
	i := 0
	offset := 0
	for (i+1) < len(b) && offset < len(trits) {
		idx := int(b[i])
		if b[i] < 0 {
			idx = int(b[i]) + len(BYTE_TO_TRITS_MAPPINGS)
		}
		length := 5
		if (len(trits) - offset) < 5 {
			length = len(trits) - offset
		}
		copy(trits[offset:offset+length], BYTE_TO_TRITS_MAPPINGS[idx][0:length])
		offset += 5
		i += 1
	}

	return trits
}

func tritsToBytesWithOffset(trits []int, offset int, size int) []int {
	bs := make([]int, (size+5-1)/5)
	for i, _ := range bs {
		value := 0
		j := 5
		if size-i*5 < 5 {
			j = size - i*5
		}
		for j > 0 {
			j -= 1
			value = value*3 + trits[offset+i*5+j]
		}
		bs[i] = int(value)
	}

	return bs
}

func tritsToBytes(trits []int) []int {
	return tritsToBytesWithOffset(trits, 0, len(trits))
}
