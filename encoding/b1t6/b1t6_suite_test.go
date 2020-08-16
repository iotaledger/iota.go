package b1t6_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestConverter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "B1T6 Encoding Suite")
}
