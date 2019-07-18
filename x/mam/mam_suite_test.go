package mam_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGoMam(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MAM Suite")
}
