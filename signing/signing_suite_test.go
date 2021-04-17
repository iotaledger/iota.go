package signing_test

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/signing"
	"github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	goldenName            = "wots"
	goldenSeed            = 42
	goldenMaxKeyFragments = 6
	goldenTestsPerCase    = 50
)

var update = flag.Bool("update", false, "update golden files")

type Test struct {
	Key       trinary.Trytes
	Address   trinary.Hash
	Hash      trinary.Hash
	Signature trinary.Trytes
}

func TestSigning(t *testing.T) {
	if *update {
		t.Log("update golden file")
		generateGolden(t)
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Signing Suite")
}

func generateGolden(t *testing.T) {
	rng := rand.New(rand.NewSource(goldenSeed))

	var data []Test
	for keyFragments := 1; keyFragments <= goldenMaxKeyFragments; keyFragments++ {
		for i := 0; i < goldenTestsPerCase; i++ {
			data = append(data, generateTest(rng, keyFragments))
		}
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

func generateTest(rng *rand.Rand, keyFragments int) Test {
	// generate a random private key
	key := trinary.MustTrytesToTrits(randomTrytes(rng, keyFragments*consts.KeyFragmentLength/3))
	for i := consts.HashTrinarySize - 1; i < len(key); i += consts.HashTrinarySize {
		key[i] = 0 // every 243rd trit needs to be zero
	}

	// compute public key / address
	digests, _ := signing.Digests(key)
	address, _ := signing.Address(digests)

	// generate random message to sign
	hashTrytes := randomTrytes(rng, consts.HashTrytesSize)
	normalized := signing.NormalizedBundleHash(hashTrytes)

	// compute signature
	var signature trinary.Trits
	for i := 0; i < len(key)/consts.KeyFragmentLength; i++ {
		hashFragmentIndex := i % (consts.HashTrytesSize / consts.KeySegmentsPerFragment)
		signatureFragment, _ := signing.SignatureFragment(
			normalized[hashFragmentIndex*consts.KeySegmentsPerFragment:(hashFragmentIndex+1)*consts.KeySegmentsPerFragment],
			key[i*consts.KeyFragmentLength:(i+1)*consts.KeyFragmentLength],
		)
		signature = append(signature, signatureFragment...)
	}

	return Test{
		Key:       trinary.MustTritsToTrytes(key),
		Address:   trinary.MustTritsToTrytes(address),
		Hash:      hashTrytes,
		Signature: trinary.MustTritsToTrytes(signature),
	}
}

func randomTrytes(rng *rand.Rand, n int) trinary.Trytes {
	var result strings.Builder
	result.Grow(n)
	for i := 0; i < n; i++ {
		result.WriteByte(consts.TryteAlphabet[rng.Intn(len(consts.TryteAlphabet))])
	}
	return result.String()
}
