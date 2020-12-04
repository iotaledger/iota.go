package curl_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iotaledger/iota.go/legacy"
	"github.com/iotaledger/iota.go/legacy/curl"
	"github.com/iotaledger/iota.go/legacy/trinary"
	"github.com/stretchr/testify/assert"
)

const (
	goldenName         = "curlp81"
	goldenSeed         = 42
	goldenTestsPerCase = 100
)

var update = flag.Bool("update", false, "update golden files")

type Test struct {
	In   trinary.Trytes `json:"in"`
	Hash trinary.Trytes `json:"hash"`
}

func TestCurlGolden(t *testing.T) {
	if *update {
		t.Log("updating golden file")
		generateGolden(t)
	}

	var tests []Test
	b, err := ioutil.ReadFile(filepath.Join("testdata", goldenName+".json"))
	assert.NoError(t, err)
	assert.NoError(t, json.Unmarshal(b, &tests))

	t.Run("absorb and squeeze trytes", func(t *testing.T) {
		for i, tt := range tests {
			t.Run(fmt.Sprintf("test vector: %d", i), func(t *testing.T) {
				c := curl.NewCurlP81()
				assert.NoError(t, c.AbsorbTrytes(tt.In))
				squeeze, err := c.SqueezeTrytes(len(tt.Hash) * legacy.TritsPerTryte)
				assert.NoError(t, err)
				assert.EqualValues(t, squeeze, tt.Hash)
			})
		}
	})

	t.Run("absorb and squeeze trits", func(t *testing.T) {
		for i, tt := range tests {
			t.Run(fmt.Sprintf("test vector: %d", i), func(t *testing.T) {
				c := curl.NewCurlP81()
				assert.NoError(t, c.Absorb(trinary.MustTrytesToTrits(tt.In)))
				squeeze, err := c.Squeeze(len(tt.Hash) * legacy.TritsPerTryte)
				assert.NoError(t, err)
				assert.EqualValues(t, squeeze, trinary.MustTrytesToTrits(tt.Hash))
			})
		}
	})

}

func generateGolden(t *testing.T) {
	rng := rand.New(rand.NewSource(goldenSeed))

	var data []Test
	// single absorb, single squeeze
	for i := 0; i < goldenTestsPerCase; i++ {
		data = append(data, generateTest(rng, legacy.HashTrytesSize, legacy.HashTrytesSize))
	}
	// multi absorb, single squeeze
	for i := 0; i < goldenTestsPerCase; i++ {
		data = append(data, generateTest(rng, legacy.HashTrytesSize*3, legacy.HashTrytesSize))
	}
	// single absorb, multi squeeze
	for i := 0; i < goldenTestsPerCase; i++ {
		data = append(data, generateTest(rng, legacy.HashTrytesSize, legacy.HashTrytesSize*3))
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	filename := filepath.Join("testdata", goldenName+".json")
	err = ioutil.WriteFile(filename, b, 0644)
	if err != nil {
		t.Fatalf("Error writing %s: %s", filename, err)
	}
}

func generateTest(rng *rand.Rand, inputTrytes, hashTrytes int) Test {
	input := randomTrytes(rng, inputTrytes)
	c := curl.NewCurlP81()
	c.MustAbsorbTrytes(input)
	hash := c.MustSqueezeTrytes(hashTrytes * legacy.TritsPerTryte)
	return Test{input, hash}
}

func randomTrytes(rng *rand.Rand, n int) trinary.Trytes {
	var result strings.Builder
	result.Grow(n)
	for i := 0; i < n; i++ {
		result.WriteByte(legacy.TryteAlphabet[rng.Intn(len(legacy.TryteAlphabet))])
	}
	return result.String()
}

func TestCurlHash(t *testing.T) {
	type test struct {
		name    string
		in      trinary.Trytes
		expHash trinary.Trytes
	}

	tests := []test{
		{
			name:    "Curl-P-81: empty trytes",
			in:      "",
			expHash: legacy.NullHashTrytes,
		},
		{
			name:    "Curl-P-81: normal trytes",
			in:      "A",
			expHash: "TJVKPMTAMIZVBVHIVQUPTKEMPROEKV9SB9COEDQYRHYPTYSKQIAN9PQKMZHCPO9TS9BHCORFKW9CQXZEE",
		},
		{
			name:    "Curl-P-81: normal trytes #2",
			in:      "Z",
			expHash: "FA9WYZSJJWSD9AEEBOGGDHFTMIZVHFURFLJLFBTNENDDCMSXGAGLXFMYZTAMKVIYDQSZEDKXSWVAOPZMK",
		},
		{
			name:    "Curl-P-81: normal trytes #3",
			in:      "NOPQRSTUVWXYZ9ABSDEFGHIJKLM",
			expHash: "GWFZSXPZPAFSVPEGEIVWOTD9MY9KVP9HYVCIWSJEITEGVOVGQGV99RONTWDXOPUBIQPIWXK9L9OHZYFUB",
		},
		{
			name:    "Curl-P-81: long absorb",
			in:      strings.Repeat("ABC", legacy.TransactionTrytesSize/3),
			expHash: "UHZVKZCGDIPNGFNPBNFZGIM9GAKYLCPTHTRFRXMNDJLZNXSGRPREFWTBKZWVTKV9BISPXEECVIXFJERAC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("tryte", func(t *testing.T) {
				hash, err := curl.HashTrytes(trinary.MustPad(tt.in, legacy.HashTrytesSize))
				assert.NoError(t, err)
				assert.EqualValues(t, tt.expHash, hash)
			})
			t.Run("trits", func(t *testing.T) {
				hash, err := curl.HashTrits(trinary.MustPadTrits(trinary.MustTrytesToTrits(tt.in), legacy.HashTrinarySize))
				assert.NoError(t, err)
				assert.EqualValues(t, tt.expHash, hash)
			})
		})
	}
}

func TestCurl_CopyState(t *testing.T) {
	a := strings.Repeat("A", legacy.HashTrytesSize)

	c := curl.NewCurlP81().(*curl.Curl)
	assert.NoError(t, c.AbsorbTrytes(a))

	state := make(trinary.Trits, curl.StateSize)
	c.CopyState(state[:])

	assert.EqualValues(t, state[:legacy.HashTrinarySize], c.MustSqueeze(legacy.HashTrinarySize))
}

func TestCurl_Reset(t *testing.T) {
	a := strings.Repeat("A", legacy.HashTrytesSize)
	b := strings.Repeat("B", legacy.HashTrytesSize)

	c1 := curl.NewCurlP81()
	assert.NoError(t, c1.AbsorbTrytes(a))
	_, err := c1.SqueezeTrytes(legacy.HashTrinarySize)
	assert.NoError(t, err)

	c1.Reset()
	c2 := curl.NewCurlP81()

	assert.NoError(t, c1.AbsorbTrytes(b))
	assert.NoError(t, c2.AbsorbTrytes(b))

	expected := c1.MustSqueezeTrytes(legacy.HashTrinarySize)
	actual := c2.MustSqueezeTrytes(legacy.HashTrinarySize)
	assert.EqualValues(t, expected, actual)
}

func TestCurl_Clone(t *testing.T) {
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

func TestCurlAbsorbAfterSqueeze(t *testing.T) {
	a := strings.Repeat("A", legacy.HashTrytesSize)

	c := curl.NewCurlP81()
	assert.NoError(t, c.AbsorbTrytes(a))
	_, err := c.SqueezeTrytes(legacy.HashTrinarySize)
	assert.NoError(t, err)
	assert.Panics(t, func() {
		_ = c.AbsorbTrytes(a)
	})
}
