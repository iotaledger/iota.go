package deposit_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDeposit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Deposit Suite")
}
