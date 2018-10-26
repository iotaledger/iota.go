package units_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestUnits(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Units Suite")
}
