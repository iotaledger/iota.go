package key_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func TestKey(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Key Suite")
}

func ReadTestJSONFile(name string, v interface{}, failureHandler types.GomegaFailHandler) {
	path := filepath.Join("testdata", name)
	f, err := os.Open(path)
	if err != nil {
		failureHandler(err.Error())
		panic("unreachable")
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	err = dec.Decode(v)
	if err != nil {
		failureHandler(err.Error())
	}
}
