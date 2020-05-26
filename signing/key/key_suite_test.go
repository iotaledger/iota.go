package key_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKey(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Key Suite")
}
