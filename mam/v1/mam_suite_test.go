package mam_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMAM(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MAM Suite")
}
