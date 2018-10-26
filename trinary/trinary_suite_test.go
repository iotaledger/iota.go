package trinary_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTrinary(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Trinary Suite")
}
