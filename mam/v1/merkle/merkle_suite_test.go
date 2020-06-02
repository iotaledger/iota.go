package merkle_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMerkle(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Merkle Suite")
}
