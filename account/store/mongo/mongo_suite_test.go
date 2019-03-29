package mongo_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMongo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mongo Suite")
}
