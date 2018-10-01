package giota_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGiota(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Giota Suite")
}
