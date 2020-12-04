package curl_test

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/curl"
	"github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

func TestCurl(t *testing.T) {
	if *update {
		t.Log("update golden file")
		generateGolden(t)
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Curl Suite")
}

func generateGolden(t *testing.T) {
	rng := rand.New(rand.NewSource(goldenSeed))

	var data []Test
	// single absorb, single squeeze
	for i := 0; i < goldenTestsPerCase; i++ {
		data = append(data, generateTest(rng, consts.HashTrytesSize, consts.HashTrytesSize))
	}
	// multi absorb, single squeeze
	for i := 0; i < goldenTestsPerCase; i++ {
		data = append(data, generateTest(rng, consts.HashTrytesSize*3, consts.HashTrytesSize))
	}
	// single absorb, multi squeeze
	for i := 0; i < goldenTestsPerCase; i++ {
		data = append(data, generateTest(rng, consts.HashTrytesSize, consts.HashTrytesSize*3))
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
	hash := c.MustSqueezeTrytes(hashTrytes * consts.TritsPerTryte)
	return Test{input, hash}
}

func randomTrytes(rng *rand.Rand, n int) trinary.Trytes {
	var result strings.Builder
	result.Grow(n)
	for i := 0; i < n; i++ {
		result.WriteByte(consts.TryteAlphabet[rng.Intn(len(consts.TryteAlphabet))])
	}
	return result.String()
}
