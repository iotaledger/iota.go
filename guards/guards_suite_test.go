package guards_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGuards(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Guards Suite")
}
