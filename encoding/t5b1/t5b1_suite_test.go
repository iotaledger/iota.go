package t5b1_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestConverter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "T5B1 Encoding Suite")
}
