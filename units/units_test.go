package units_test

import (
	. "github.com/iotaledger/iota.go/units"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo/extensions/table"
)

var _ = Describe("Units", func() {

	DescribeTable("float conversion",
		func(in float64, from Unit, to Unit, expected float64) {
			Expect(ConvertUnits(in, from, to)).To(Equal(expected))
		},
		Entry("Mi to I", float64(100), Mi, I, float64(100000000)),
		Entry("Gi to I", float64(10.1), Gi, I, float64(10100000000)),
		Entry("I to Ti", float64(1), I, Ti, float64(0.000000000001)),
		Entry("Ti to I", float64(1), Ti, I, float64(1000000000000)),
		Entry("Gi to Ti", float64(1000), Gi, Ti, float64(1)),
		Entry("Gi to I", float64(133.999111111), Gi, I, float64(133999111111)),
	)

	DescribeTable("string conversion",
		func(in string, from Unit, to Unit, expected float64) {
			Expect(ConvertUnitsString(in, from, to)).To(Equal(expected))
		},
		Entry("Gi to I", "10.1", Gi, I, float64(10100000000)),
		Entry("Gi to I", "133.999111111", Gi, I, float64(133999111111)),
	)

})
