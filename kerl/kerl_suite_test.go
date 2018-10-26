package kerl_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKerl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kerl Suite")
}
