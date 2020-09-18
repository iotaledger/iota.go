package curl_test

import (
	"strings"
	"testing"

	"github.com/iotaledger/iota.go/legacy"
	"github.com/iotaledger/iota.go/legacy/curl"
	"github.com/iotaledger/iota.go/legacy/trinary"
	"github.com/stretchr/testify/assert"
)

func TestCurlHash(t *testing.T) {
	type test struct {
		name       string
		in         trinary.Trytes
		expSqueeze trinary.Trytes
		rounds     []curl.CurlRounds
	}

	tests := []test{
		{
			name:       "Curl-P-81: empty trytes",
			in:         "",
			expSqueeze: legacy.NullHashTrytes,
			rounds:     []curl.CurlRounds{curl.CurlP81},
		},
		{
			name:       "Curl-P-81: normal trytes",
			in:         "A",
			expSqueeze: "TJVKPMTAMIZVBVHIVQUPTKEMPROEKV9SB9COEDQYRHYPTYSKQIAN9PQKMZHCPO9TS9BHCORFKW9CQXZEE",
			rounds:     []curl.CurlRounds{curl.CurlP81},
		},
		{
			name:       "Curl-P-81: normal trytes #2",
			in:         "Z",
			expSqueeze: "FA9WYZSJJWSD9AEEBOGGDHFTMIZVHFURFLJLFBTNENDDCMSXGAGLXFMYZTAMKVIYDQSZEDKXSWVAOPZMK",
		},
		{
			name:       "Curl-P-81: normal trytes #3",
			in:         "NOPQRSTUVWXYZ9ABSDEFGHIJKLM",
			expSqueeze: "GWFZSXPZPAFSVPEGEIVWOTD9MY9KVP9HYVCIWSJEITEGVOVGQGV99RONTWDXOPUBIQPIWXK9L9OHZYFUB",
			rounds:     []curl.CurlRounds{curl.CurlP81},
		},
		{
			name:       "Curl-P-81: long absorb",
			in:         strings.Repeat("ABC", legacy.TransactionTrytesSize/3),
			expSqueeze: "UHZVKZCGDIPNGFNPBNFZGIM9GAKYLCPTHTRFRXMNDJLZNXSGRPREFWTBKZWVTKV9BISPXEECVIXFJERAC",
			rounds:     []curl.CurlRounds{curl.CurlP81},
		},
		{
			name:       "Curl-P-81: long squeeze",
			in:         "ABC",
			expSqueeze: "LRJMQXFSZSLCIMKZTWFTEIHKWJZMUOHPSOVXZOHOEVHC9D9DROUQGRPTBZWOIJFTMGMXEYKXEJROQLWNUPSFJJRVTLUUJYW9GBQVXNCAUEGEBV9IJQ9TWFDHCFPUUYPCYLACTAIK9UZAJLVXLI9NPGCJN9ICFTEIYY",
			rounds:     []curl.CurlRounds{curl.CurlP81},
		},
		{
			name:       "Curl-P-27: empty trytes",
			in:         "",
			expSqueeze: legacy.NullHashTrytes,
			rounds:     []curl.CurlRounds{curl.CurlP27},
		},
		{
			name:       "Curl-P-27: normal trytes",
			in:         "TWENTYSEVEN",
			expSqueeze: "RQPYXJPRXEEPLYLAHWTTFRXXUZTV9SZPEVOQ9FZATCXJOZLZ9A9BFXTUBSHGXN9OOA9GWIPGAAWEDVNPN",
			rounds:     []curl.CurlRounds{curl.CurlP27},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("tryte", func(t *testing.T) {
				c := curl.NewCurl(tt.rounds...)
				err := c.AbsorbTrytes(trinary.MustPad(tt.in, legacy.HashTrytesSize))
				assert.NoError(t, err)
				squeeze, err := c.SqueezeTrytes(len(tt.expSqueeze) * legacy.TritsPerTryte)
				assert.NoError(t, err)
				assert.Equal(t, squeeze, tt.expSqueeze)
			})
			t.Run("trits", func(t *testing.T) {
				c := curl.NewCurl(tt.rounds...)
				err := c.Absorb(trinary.MustPadTrits(trinary.MustTrytesToTrits(tt.in), legacy.HashTrinarySize))
				assert.NoError(t, err)
				squeeze, err := c.Squeeze(len(tt.expSqueeze) * legacy.TritsPerTryte)
				assert.NoError(t, err)
				assert.Equal(t, squeeze, trinary.MustTrytesToTrits(tt.expSqueeze))
			})
		})
	}
}

func TestCurlClone(t *testing.T) {
	a := strings.Repeat("A", legacy.HashTrytesSize)
	b := strings.Repeat("B", legacy.HashTrytesSize)

	c1 := curl.NewCurlP81()
	err := c1.AbsorbTrytes(a)
	assert.NoError(t, err)

	c2 := c1.Clone()
	err = c1.AbsorbTrytes(b)
	assert.NoError(t, err)
	err = c2.AbsorbTrytes(b)
	assert.NoError(t, err)

	assert.Equal(t, c2.MustSqueezeTrytes(legacy.HashTrinarySize), c1.MustSqueezeTrytes(legacy.HashTrinarySize))
}
