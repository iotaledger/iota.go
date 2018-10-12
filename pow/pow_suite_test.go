package pow_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPow(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PoW Suite")
}
