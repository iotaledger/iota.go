package kerl_test

import (
	"encoding/hex"
	"testing"

	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/kerl"
	"github.com/iotaledger/iota.go/trinary"
	"github.com/stretchr/testify/assert"
)

func TestHashValidTrits(t *testing.T) {
	tests := []struct {
		name     string
		in       trinary.Trytes
		expected trinary.Trytes
	}{
		{
			name:     "normal trytes",
			in:       "HHPELNTNJIOKLYDUW9NDULWPHCWFRPTDIUWLYUHQWWJVPAKKGKOAZFJPQJBLNDPALCVXGJLRBFSHATF9C",
			expected: "DMJWZTDJTASXZTHZFXFZXWMNFHRTKWFUPCQJXEBJCLRZOM9LPVJSTCLFLTQTDGMLVUHOVJHBBUYFD9AXX",
		},
		{
			name:     "normal trytes #2",
			in:       "QAUGQZQKRAW9GKEFIBUD9BMJQOABXBTFELCT9GVSZCPTZOSFBSHPQRWJLLWURPXKNAOWCSVWUBNDSWMPW",
			expected: "HOVOHFEPCIGTOFEAZVXAHQRFFRTPQEEKANKFKIHUKSGRICVADWDMBINDYKRCCIWBEOPXXIKMLNSOHEAQZ",
		},
		{
			name:     "normal trytes #3",
			in:       "MWBLYBSRKEKLDHUSRDSDYZRNV9DDCPN9KENGXIYTLDWPJPKBHQBOALSDH9LEJVACJAKJYPCFTJEROARRW",
			expected: "KXBKXQUZBYZFSYSPDPCNILVUSXOEHQWWWFKZPFCQ9ABGIIQBNLSWLPIMV9LYNQDDYUS9L9GNUIYKYAGVZ",
		},
		{
			name:     "output with non-zero 243rd trit",
			in:       "GYOMKVTSNHVJNCNFBBAH9AAMXLPLLLROQY99QN9DLSJUHDPBLCFFAIQXZA9BKMBJCYSFHFPXAHDWZFEIZ",
			expected: "OXJCNFHUNAHWDLKKPELTBFUCVW9KLXKOGWERKTJXQMXTKFKNWNNXYD9DMJJABSEIONOSJTTEVKVDQEWTW",
		},
		{
			name:     "input with 243-trits",
			in:       "EMIDYNHBWMBCXVDEFOFWINXTERALUKYYPPHKP9JJFGJEIUY9MUDVNFZHMMWZUYUSWAIOWEVTHNWMHANBH",
			expected: "EJEAOOZYSAWFPZQESYDHZCGYNSTWXUMVJOVDWUNZJXDGWCLUFGIMZRMGCAZGKNPLBRLGUNYWKLJTYEAQX",
		},
		{
			name:     "output with more than 243-trits",
			in:       "9MIDYNHBWMBCXVDEFOFWINXTERALUKYYPPHKP9JJFGJEIUY9MUDVNFZHMMWZUYUSWAIOWEVTHNWMHANBH",
			expected: "G9JYBOMPUXHYHKSNRNMMSSZCSHOFYOYNZRSZMAAYWDYEIMVVOGKPJBVBM9TDPULSFUNMTVXRKFIDOHUXXVYDLFSZYZTWQYTE9SPYYWYTXJYQ9IFGYOLZXWZBKWZN9QOOTBQMWMUBLEWUEEASRHRTNIQWJQNDWRYLCA",
		},
		{
			name:     "input & output with more than 243-trits",
			in:       "G9JYBOMPUXHYHKSNRNMMSSZCSHOFYOYNZRSZMAAYWDYEIMVVOGKPJBVBM9TDPULSFUNMTVXRKFIDOHUXXVYDLFSZYZTWQYTE9SPYYWYTXJYQ9IFGYOLZXWZBKWZN9QOOTBQMWMUBLEWUEEASRHRTNIQWJQNDWRYLCA",
			expected: "LUCKQVACOGBFYSPPVSSOXJEKNSQQRQKPZC9NXFSMQNRQCGGUL9OHVVKBDSKEQEBKXRNUJSRXYVHJTXBPDWQGNSCDCBAIRHAQCOWZEBSNHIJIGPZQITIBJQ9LNTDIBTCQ9EUWKHFLGFUVGGUWJONK9GBCDUIMAYMMQX",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			k := kerl.NewKerl()
			assert.NoError(t, k.Absorb(trinary.MustTrytesToTrits(test.in)))
			trits, err := k.Squeeze(len(test.expected) * consts.HashTrinarySize / consts.HashTrytesSize)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, trinary.MustTritsToTrytes(trits))
		})
	}
}

func TestHashValidTrytes(t *testing.T) {
	tests := []struct {
		name     string
		in       trinary.Trytes
		expected trinary.Trytes
	}{
		{
			name:     "normal trytes",
			in:       "HHPELNTNJIOKLYDUW9NDULWPHCWFRPTDIUWLYUHQWWJVPAKKGKOAZFJPQJBLNDPALCVXGJLRBFSHATF9C",
			expected: "DMJWZTDJTASXZTHZFXFZXWMNFHRTKWFUPCQJXEBJCLRZOM9LPVJSTCLFLTQTDGMLVUHOVJHBBUYFD9AXX",
		},
		{
			name:     "normal trytes #2",
			in:       "QAUGQZQKRAW9GKEFIBUD9BMJQOABXBTFELCT9GVSZCPTZOSFBSHPQRWJLLWURPXKNAOWCSVWUBNDSWMPW",
			expected: "HOVOHFEPCIGTOFEAZVXAHQRFFRTPQEEKANKFKIHUKSGRICVADWDMBINDYKRCCIWBEOPXXIKMLNSOHEAQZ",
		},
		{
			name:     "normal trytes #3",
			in:       "MWBLYBSRKEKLDHUSRDSDYZRNV9DDCPN9KENGXIYTLDWPJPKBHQBOALSDH9LEJVACJAKJYPCFTJEROARRW",
			expected: "KXBKXQUZBYZFSYSPDPCNILVUSXOEHQWWWFKZPFCQ9ABGIIQBNLSWLPIMV9LYNQDDYUS9L9GNUIYKYAGVZ",
		},
		{
			name:     "output with non-zero 243rd trit",
			in:       "GYOMKVTSNHVJNCNFBBAH9AAMXLPLLLROQY99QN9DLSJUHDPBLCFFAIQXZA9BKMBJCYSFHFPXAHDWZFEIZ",
			expected: "OXJCNFHUNAHWDLKKPELTBFUCVW9KLXKOGWERKTJXQMXTKFKNWNNXYD9DMJJABSEIONOSJTTEVKVDQEWTW",
		},
		{
			name:     "input with 243-trits",
			in:       "EMIDYNHBWMBCXVDEFOFWINXTERALUKYYPPHKP9JJFGJEIUY9MUDVNFZHMMWZUYUSWAIOWEVTHNWMHANBH",
			expected: "EJEAOOZYSAWFPZQESYDHZCGYNSTWXUMVJOVDWUNZJXDGWCLUFGIMZRMGCAZGKNPLBRLGUNYWKLJTYEAQX",
		},
		{
			name:     "output with more than 243-trits",
			in:       "9MIDYNHBWMBCXVDEFOFWINXTERALUKYYPPHKP9JJFGJEIUY9MUDVNFZHMMWZUYUSWAIOWEVTHNWMHANBH",
			expected: "G9JYBOMPUXHYHKSNRNMMSSZCSHOFYOYNZRSZMAAYWDYEIMVVOGKPJBVBM9TDPULSFUNMTVXRKFIDOHUXXVYDLFSZYZTWQYTE9SPYYWYTXJYQ9IFGYOLZXWZBKWZN9QOOTBQMWMUBLEWUEEASRHRTNIQWJQNDWRYLCA",
		},
		{
			name:     "input & output with more than 243-trits",
			in:       "G9JYBOMPUXHYHKSNRNMMSSZCSHOFYOYNZRSZMAAYWDYEIMVVOGKPJBVBM9TDPULSFUNMTVXRKFIDOHUXXVYDLFSZYZTWQYTE9SPYYWYTXJYQ9IFGYOLZXWZBKWZN9QOOTBQMWMUBLEWUEEASRHRTNIQWJQNDWRYLCA",
			expected: "LUCKQVACOGBFYSPPVSSOXJEKNSQQRQKPZC9NXFSMQNRQCGGUL9OHVVKBDSKEQEBKXRNUJSRXYVHJTXBPDWQGNSCDCBAIRHAQCOWZEBSNHIJIGPZQITIBJQ9LNTDIBTCQ9EUWKHFLGFUVGGUWJONK9GBCDUIMAYMMQX",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			k := kerl.NewKerl()
			assert.NoError(t, k.AbsorbTrytes(test.in))
			trytes, err := k.SqueezeTrytes(len(test.expected) * consts.HashTrinarySize / consts.HashTrytesSize)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, trytes)
		})
	}
}

func TestHashValidBytes(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		expected string
	}{
		{
			name:     "input with 48 bytes",
			in:       "3f1b9727c967dda4f0ef98032f97483864e0dc9ed391dd5b8bc8133d9ce77fe182fef749de882dace92b655c6bba22df",
			expected: "bec15753fa767d98b59de095a962f472f7b4e15da47d6dd987c4608ccd32f8e2231c3201faca0ee2591f11179c816e30",
		},
		{
			name:     "output with more than 48 bytes",
			in:       "ec3357c2b1f26b6567a80542a65159f3fdc5c4a7ff0d07ff52c14ed39df3cdee8e3b62250b04592ba0beef909e1c430e",
			expected: "be8fc99ed01018cd8e1904d5188ddc62d85edf1aecc61609a820df347cfe7b8bfa928e9c0460854c638fa330cfd0e3f517a7a822b3d0a0fb3d7b05bbe86ae815c8e063b638363351ac87dd62784db4441e1beae32596a224699fd5aeeab61eab",
		},
		{
			name:     "input & output with more than 48 bytes",
			in:       "be8fc99ed01018cd8e1904d5188ddc62d85edf1aecc61609a820df347cfe7b8bfa928e9c0460854c638fa330cfd0e3f517a7a822b3d0a0fb3d7b05bbe86ae815c8e063b638363351ac87dd62784db4441e1beae32596a224699fd5aeeab61eab",
			expected: "4b6e005b96685f95c22b94e5248b2fa06b22062124bea182a2e7d8834f9bcf1a3e0debe180a377f19207404916263040b9acbbca8851297604dcdf1ae0cb11f8444d9f387fdd80993e7158e4efea04630bcf931d90190dc98a4243265eb73c3f",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			k := kerl.NewKerl()
			// absorb
			inBuf, err := hex.DecodeString(test.in)
			assert.NoError(t, err)
			written, err := k.Write(inBuf)
			assert.Len(t, inBuf, written)
			assert.NoError(t, err)
			// squeeze
			outBuf := make([]byte, hex.DecodedLen(len(test.expected)))
			read, err := k.Read(outBuf)
			assert.Len(t, outBuf, read)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, hex.EncodeToString(outBuf))
		})
	}
}

func TestErrorForWriteAfterRead(t *testing.T) {
	k := kerl.NewKerl()
	written, err := k.Write(make([]byte, consts.HashBytesSize))
	assert.NoError(t, err)
	assert.Equal(t, consts.HashBytesSize, written)
	read, err := k.Read(make([]byte, consts.HashBytesSize))
	assert.NoError(t, err)
	assert.Equal(t, consts.HashBytesSize, read)
	// write again
	written, err = k.Write(make([]byte, consts.HashBytesSize))
	assert.Zero(t, written)
	assert.Equal(t, kerl.ErrAbsorbAfterSqueeze, err)
}

func TestHashInvalidTrits(t *testing.T) {

	t.Run("should return an error with empty trits slice", func(t *testing.T) {
		k := kerl.NewKerl()
		assert.Error(t, k.Absorb(trinary.Trits{}))
	})

	t.Run("should return an error with invalid trits slice length", func(t *testing.T) {
		k := kerl.NewKerl()
		assert.Error(t, k.Absorb(trinary.Trits{1, 0, 0, 0, 0, -1}))
	})

	t.Run("should return an error for Absorb after Squeeze", func(t *testing.T) {
		k := kerl.NewKerl()
		assert.NoError(t, k.Absorb(make(trinary.Trits, consts.HashTrinarySize)))
		_, err := k.Squeeze(consts.HashTrinarySize)
		assert.NoError(t, err)
		// absorb again
		assert.Equal(t, kerl.ErrAbsorbAfterSqueeze, k.Absorb(make(trinary.Trits, consts.HashTrinarySize)))
	})
}

func TestHashInvalidTrytes(t *testing.T) {
	t.Run("should return an error with empty tryte slice", func(t *testing.T) {
		k := kerl.NewKerl()
		assert.Error(t, k.AbsorbTrytes(""))
	})

	t.Run("should return an error with invalid trits slice length", func(t *testing.T) {
		k := kerl.NewKerl()
		assert.Error(t, k.AbsorbTrytes("AR"))
	})

	t.Run("should return an error for Absorb after Squeeze", func(t *testing.T) {
		k := kerl.NewKerl()
		assert.NoError(t, k.AbsorbTrytes(consts.NullHashTrytes))
		_, err := k.SqueezeTrytes(consts.HashTrinarySize)
		assert.NoError(t, err)
		// absorb again
		assert.Equal(t, kerl.ErrAbsorbAfterSqueeze, k.AbsorbTrytes(consts.NullHashTrytes))
	})
}

func TestKerlSum(t *testing.T) {

	in := "ff"
	hash := "08869cad3dc2429eb295195200ad22eb36188452ba65f0e31b2b21bd49b503a7f1d1d61a6df8bff569d3decc9810721b"

	t.Run("valid bytes", func(t *testing.T) {
		k := kerl.NewKerl()
		// absorb
		inBuf, err := hex.DecodeString(in)
		assert.NoError(t, err)
		written, err := k.Write(inBuf)
		assert.Len(t, inBuf, written)
		assert.NoError(t, err)
		// append sum to inBuf
		outBuf := k.Sum(inBuf)
		assert.Equal(t, in+hash, hex.EncodeToString(outBuf))
	})

	t.Run("absorb after sum", func(t *testing.T) {
		k := kerl.NewKerl()
		// absorb
		inBuf, err := hex.DecodeString(in)
		assert.NoError(t, err)
		written, err := k.Write(inBuf)
		assert.Len(t, inBuf, written)
		assert.NoError(t, err)
		// append sum to inBuf
		outBuf := k.Sum(inBuf)
		assert.Equal(t, in+hash, hex.EncodeToString(outBuf))
		// absorb again
		written, err = k.Write(inBuf)
		assert.Len(t, inBuf, written)
		assert.NoError(t, err)
	})
}

func TestKerlReset(t *testing.T) {
	t.Run("reset during absorb", func(t *testing.T) {
		k1 := kerl.NewKerl()
		assert.NoError(t, k1.Absorb(make(trinary.Trits, consts.HashTrinarySize)))

		k1.Reset()
		k2 := kerl.NewKerl()
		assert.Equal(t, k1.MustSqueeze(consts.HashTrinarySize), k2.MustSqueeze(consts.HashTrinarySize))
	})

	t.Run("reset during squeeze", func(t *testing.T) {
		k1 := kerl.NewKerl()
		assert.NoError(t, k1.Absorb(make(trinary.Trits, consts.HashTrinarySize)))
		k1.MustSqueeze(consts.HashTrinarySize)

		k1.Reset()
		k2 := kerl.NewKerl()
		assert.Equal(t, k1.MustSqueeze(consts.HashTrinarySize), k2.MustSqueeze(consts.HashTrinarySize))
	})
}

func TestKerlClone(t *testing.T) {
	t.Run("clone during absorb", func(t *testing.T) {
		k1, k2 := kerl.NewKerl(), kerl.NewKerl()
		assert.NoError(t, k1.Absorb(make(trinary.Trits, consts.HashTrinarySize)))
		assert.NoError(t, k2.Absorb(make(trinary.Trits, consts.HashTrinarySize)))

		k1Clone := k1.Clone()
		assert.NoError(t, k1.Absorb(make(trinary.Trits, consts.HashTrinarySize)))

		hash1 := k1.MustSqueeze(consts.HashTrinarySize)
		hash2 := k2.MustSqueeze(consts.HashTrinarySize)
		hash1Clone := k1Clone.MustSqueeze(consts.HashTrinarySize)
		assert.Equal(t, hash2, hash1Clone)
		assert.NotEqual(t, hash1, hash1Clone)
	})

	t.Run("clone during squeeze", func(t *testing.T) {
		k1, k2 := kerl.NewKerl(), kerl.NewKerl()
		assert.NoError(t, k1.Absorb(make(trinary.Trits, consts.HashTrinarySize)))
		assert.NoError(t, k2.Absorb(make(trinary.Trits, consts.HashTrinarySize)))

		k1.MustSqueeze(consts.HashTrinarySize)
		k2.MustSqueeze(consts.HashTrinarySize)

		k1Clone := k1.Clone()
		k1.MustSqueeze(consts.HashTrinarySize)

		hash1 := k1.MustSqueeze(consts.HashTrinarySize)
		hash2 := k2.MustSqueeze(consts.HashTrinarySize)
		hash1Clone := k1Clone.MustSqueeze(consts.HashTrinarySize)
		assert.Equal(t, hash2, hash1Clone)
		assert.NotEqual(t, hash1, hash1Clone)
	})
}
