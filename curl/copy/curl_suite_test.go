package curl_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCurl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Curl Suite")
}
