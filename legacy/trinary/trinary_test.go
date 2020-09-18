package trinary_test

import (
	"strings"
	"testing"

	"github.com/iotaledger/iota.go/legacy"
	"github.com/iotaledger/iota.go/legacy/trinary"
	"github.com/stretchr/testify/assert"
)

func TestValidTrit(t *testing.T) {
	t.Run("should return true for valid trits", func(t *testing.T) {
		assert.True(t, trinary.ValidTrit(-1))
		assert.True(t, trinary.ValidTrit(1))
		assert.True(t, trinary.ValidTrit(0))
	})

	t.Run("should return false for invalid trits", func(t *testing.T) {
		assert.False(t, trinary.ValidTrit(2))
		assert.False(t, trinary.ValidTrit(-2))
	})
}

func TestValidTrits(t *testing.T) {
	t.Run("should not return an error for valid trits", func(t *testing.T) {
		assert.NoError(t, trinary.ValidTrits(trinary.Trits{0, -1, 1, -1, 0, 0, 1, 1}))
	})

	t.Run("should return an error for invalid trits", func(t *testing.T) {
		assert.Error(t, trinary.ValidTrits(trinary.Trits{-1, 0, 3, -1, 0, 0, 1}))
	})
}

func TestNewTrits(t *testing.T) {
	t.Run("should return trits and no error with valid trits", func(t *testing.T) {
		trits, err := trinary.NewTrits([]int8{0, 0, 0, 0, -1, 1, 1, 0})
		assert.NoError(t, err)
		assert.Equal(t, []int8{0, 0, 0, 0, -1, 1, 1, 0}, trits)
	})

	t.Run("should return an error for invalid trits", func(t *testing.T) {
		_, err := trinary.NewTrits([]int8{122, 0, -1, 60, -10, -50})
		assert.Error(t, err)
	})
}

func TestTritsEqual(t *testing.T) {
	t.Run("should return true for equal trits", func(t *testing.T) {
		a := trinary.Trits{0, 1, 0}
		b := trinary.Trits{0, 1, 0}
		equal, err := trinary.TritsEqual(a, b)
		assert.NoError(t, err)
		assert.True(t, equal)
	})

	t.Run("should return false for unequal trits", func(t *testing.T) {
		a := trinary.Trits{0, 1, 0}
		b := trinary.Trits{1, 0, 0}
		equal, err := trinary.TritsEqual(a, b)
		assert.NoError(t, err)
		assert.False(t, equal)
	})

	t.Run("should return an error for invalid trits", func(t *testing.T) {
		a := trinary.Trits{120, 50, -33}
		equal, err := trinary.TritsEqual(a, a)
		assert.Error(t, err)
		assert.False(t, equal)
	})
}

func TestIntToTrits(t *testing.T) {
	t.Run("should return correct trits representation for positive int64", func(t *testing.T) {
		assert.Equal(t, trinary.Trits{0, 1, 1}, trinary.IntToTrits(12))
		assert.Equal(t, trinary.Trits{-1, 1}, trinary.IntToTrits(2))
		assert.Equal(t, trinary.Trits{0, 0, 1, -1, 0, -1, 0, 0, 1, 1, -1, 1, 0, -1, 1}, trinary.IntToTrits(3332727))
		assert.Equal(t, trinary.Trits{0}, trinary.IntToTrits(0))
	})

	t.Run("should return correct trits representation for negative int64", func(t *testing.T) {
		assert.Equal(t, trinary.Trits{-1, 1, -1}, trinary.IntToTrits(-7))
		assert.Equal(t, trinary.Trits{0, -1, 1, 0, 1, -1, -1, 1, 1, 1, -1, 0, 1, -1}, trinary.IntToTrits(-1094385))
	})
}

func TestTritsToInt(t *testing.T) {
	t.Run("should return correct nums for positive trits", func(t *testing.T) {
		assert.Equal(t, int64(12), trinary.TritsToInt(trinary.Trits{0, 1, 1}))
		assert.Equal(t, int64(2), trinary.TritsToInt(trinary.Trits{-1, 1}))
		assert.Equal(t, int64(3332727), trinary.TritsToInt(trinary.Trits{0, 0, 1, -1, 0, -1, 0, 0, 1, 1, -1, 1, 0, -1, 1}))
		assert.Equal(t, int64(0), trinary.TritsToInt(trinary.Trits{0}))
	})

	t.Run("should return correct nums for negative trits", func(t *testing.T) {
		assert.Equal(t, int64(-7), trinary.TritsToInt(trinary.Trits{-1, 1, -1}))
		assert.Equal(t, int64(-1094385), trinary.TritsToInt(trinary.Trits{0, -1, 1, 0, 1, -1, -1, 1, 1, 1, -1, 0, 1, -1}))
	})
}

func TestCanTritsToTrytes(t *testing.T) {
	t.Run("returns true for valid lengths", func(t *testing.T) {
		assert.True(t, trinary.CanTritsToTrytes(trinary.Trits{1, 1, 1}))
		assert.True(t, trinary.CanTritsToTrytes(trinary.Trits{1, 1, 1, 1, 1, 1}))
	})

	t.Run("returns false for invalid lengths", func(t *testing.T) {
		assert.False(t, trinary.CanTritsToTrytes(trinary.Trits{1, 1}))
		assert.False(t, trinary.CanTritsToTrytes(trinary.Trits{1, 1, 1, 1}))
	})

	t.Run("returns false for empty trits slices", func(t *testing.T) {
		assert.False(t, trinary.CanTritsToTrytes(trinary.Trits{}))
	})
}

func TestTrailingZeros(t *testing.T) {
	t.Run("should return count of zeroes", func(t *testing.T) {
		assert.Equal(t, 3, trinary.TrailingZeros(trinary.Trits{1, 0, 0, 0}))
		assert.Equal(t, 4, trinary.TrailingZeros(trinary.Trits{0, 0, 0, 0}))
	})
}

func TestTritsToTrytes(t *testing.T) {
	t.Run("should return trytes and no errors for valid trits", func(t *testing.T) {
		trytes, err := trinary.TritsToTrytes(trinary.Trits{1, 1, 1})
		assert.NoError(t, err)
		assert.Equal(t, "M", trytes)
	})

	t.Run("should return an error for invalid trits slice length", func(t *testing.T) {
		_, err := trinary.TritsToTrytes(trinary.Trits{1, 1})
		assert.Error(t, err)
	})

	t.Run("should return an error for invalid trits", func(t *testing.T) {
		_, err := trinary.TritsToTrytes(trinary.Trits{12, -45})
		assert.Error(t, err)
	})
}

func TestMustTritsToTrytes(t *testing.T) {
	t.Run("should return trytes and not panic for valid trits", func(t *testing.T) {
		trytes := trinary.MustTritsToTrytes(trinary.Trits{1, 1, 1})
		assert.Equal(t, "M", trytes)
	})
}

func TestCanBeHash(t *testing.T) {
	t.Run("should return true for a valid trits slice", func(t *testing.T) {
		assert.True(t, trinary.CanBeHash(make(trinary.Trits, legacy.HashTrinarySize)))
	})
	t.Run("should return false for an invalid trits slice", func(t *testing.T) {
		assert.False(t, trinary.CanBeHash(make(trinary.Trits, 100)))
		assert.False(t, trinary.CanBeHash(make(trinary.Trits, 250)))
	})
}

func TestReverseTrits(t *testing.T) {
	t.Run("should correctly reverse trits", func(t *testing.T) {
		assert.Equal(t, trinary.Trits{-1, 0, 1}, trinary.ReverseTrits(trinary.Trits{1, 0, -1}))
	})

	t.Run("should return an empty trits slice for empty trits slice", func(t *testing.T) {
		assert.Equal(t, trinary.Trits{}, trinary.ReverseTrits(trinary.Trits{}))
	})
}

func TestValidTryte(t *testing.T) {
	t.Run("should return true for valid tryte", func(t *testing.T) {
		assert.NoError(t, trinary.ValidTryte('A'))
		assert.NoError(t, trinary.ValidTryte('X'))
		assert.NoError(t, trinary.ValidTryte('F'))
	})

	t.Run("should return false for invalid tryte", func(t *testing.T) {
		assert.Error(t, trinary.ValidTryte('a'))
		assert.Error(t, trinary.ValidTryte('x'))
		assert.Error(t, trinary.ValidTryte('f'))
	})
}

func TestValidTrytes(t *testing.T) {
	t.Run("should not return any error for valid trytes", func(t *testing.T) {
		assert.NoError(t, trinary.ValidTrytes("AAA"))
		assert.NoError(t, trinary.ValidTrytes("XXX"))
		assert.NoError(t, trinary.ValidTrytes("FFF"))
	})

	t.Run("should return an error for invalid trytes", func(t *testing.T) {
		assert.Error(t, trinary.ValidTrytes("f"))
		assert.Error(t, trinary.ValidTrytes("xx"))
		assert.Error(t, trinary.ValidTrytes("203984"))
		assert.Error(t, trinary.ValidTrytes(""))
	})
}

func TestNewTrytes(t *testing.T) {
	t.Run("should return trytes for valid string input", func(t *testing.T) {
		trytes, err := trinary.NewTrytes("BLABLABLA")
		assert.NoError(t, err)
		assert.Equal(t, "BLABLABLA", trytes)
	})

	t.Run("should return an error for invalid string input", func(t *testing.T) {
		_, err := trinary.NewTrytes("abcd")
		assert.Error(t, err)
	})

	t.Run("should return an error for empty string input", func(t *testing.T) {
		_, err := trinary.NewTrytes("")
		assert.Error(t, err)
	})
}

func TestTrytesToTrits(t *testing.T) {
	t.Run("should return trits for valid trytes", func(t *testing.T) {
		trits, err := trinary.TrytesToTrits("M")
		assert.NoError(t, err)
		assert.Equal(t, trinary.Trits{1, 1, 1}, trits)
		trits, err = trinary.TrytesToTrits("O")
		assert.NoError(t, err)
		assert.Equal(t, trinary.Trits{0, -1, -1}, trits)
	})

	t.Run("should return an error for empty trytes", func(t *testing.T) {
		_, err := trinary.TrytesToTrits("")
		assert.Error(t, err)
	})

	t.Run("should return an error for invalid trytes", func(t *testing.T) {
		_, err := trinary.TrytesToTrits("abcd")
		assert.Error(t, err)
	})
}

func TestMinTrits(t *testing.T) {
	t.Run("should return correct length", func(t *testing.T) {
		v := trinary.MinTrits(1)
		assert.Equal(t, 1, v)
		v = trinary.MinTrits(4)
		assert.Equal(t, 2, v)
	})
}

func TestIntToTrytes(t *testing.T) {
	t.Run("should return correct trytes", func(t *testing.T) {
		v := trinary.IntToTrytes(-1, 1)
		assert.Equal(t, "Z", v)

		v = trinary.IntToTrytes(500, 5)
		assert.Equal(t, "NSA99", v)
	})
}

func TestTrytesToInt(t *testing.T) {
	t.Run("should return correct int", func(t *testing.T) {
		v := trinary.TrytesToInt("ABCD")
		assert.Equal(t, int64(80974), v)

		v = trinary.TrytesToInt("ABCDEFGH")
		assert.Equal(t, int64(86483600668), v)
	})
}

func TestMustTrytesToTrits(t *testing.T) {
	t.Run("should return trits for valid trytes", func(t *testing.T) {
		trits := trinary.MustTrytesToTrits("M")
		assert.Equal(t, trinary.Trits{1, 1, 1}, trits)
		trits = trinary.MustTrytesToTrits("O")
		assert.Equal(t, trinary.Trits{0, -1, -1}, trits)
	})

	t.Run("should panic for invalid trytes", func(t *testing.T) {
		assert.Panics(t, func() {
			trinary.MustTrytesToTrits("abcd")
		})
	})
}

func TestMustPad(t *testing.T) {
	t.Run("should pad up to the given size", func(t *testing.T) {
		assert.Equal(t, "A9999", trinary.MustPad("A", 5))
		assert.Equal(t, strings.Repeat("9", 81), trinary.MustPad("", 81))
	})
}

func TestMustPadTrits(t *testing.T) {
	t.Run("should pad up to the given size", func(t *testing.T) {
		assert.Equal(t, trinary.Trits{0, 0, 0, 0, 0}, trinary.MustPadTrits(trinary.Trits{}, 5))
		assert.Equal(t, trinary.Trits{1, 1, 0, 0, 0}, trinary.MustPadTrits(trinary.Trits{1, 1}, 5))
		assert.Equal(t, trinary.Trits{1, -1, 0, 1, 0}, trinary.MustPadTrits(trinary.Trits{1, -1, 0, 1}, 5))
	})
}

func TestAddTrits(t *testing.T) {
	t.Run("should correctly add trits together (positive)", func(t *testing.T) {
		assert.Equal(t, int64(10), trinary.TritsToInt(trinary.AddTrits(trinary.IntToTrits(5), trinary.IntToTrits(5))))
		assert.Equal(t, int64(0), trinary.TritsToInt(trinary.AddTrits(trinary.IntToTrits(0), trinary.IntToTrits(0))))
		assert.Equal(t, int64(-120), trinary.TritsToInt(trinary.AddTrits(trinary.IntToTrits(-100), trinary.IntToTrits(-20))))
	})
}
