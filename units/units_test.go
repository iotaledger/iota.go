package units_test

import (
	"fmt"
	"testing"

	"github.com/iotaledger/iota.go/v2/units"
	"github.com/stretchr/testify/assert"
)

func TestConvertUnits(t *testing.T) {
	tests := []struct {
		name     string
		in       float64
		from     units.Unit
		to       units.Unit
		expected float64
	}{
		{name: "Mi to I", in: float64(100), from: units.Mi, to: units.I, expected: float64(100000000)},
		{name: "Gi to I", in: float64(10.1), from: units.Gi, to: units.I, expected: float64(10100000000)},
		{name: "I to Ti", in: float64(1), from: units.I, to: units.Ti, expected: float64(0.000000000001)},
		{name: "Ti to I", in: float64(1), from: units.Ti, to: units.I, expected: float64(1000000000000)},
		{name: "Gi to Ti", in: float64(1000), from: units.Gi, to: units.Ti, expected: float64(1)},
		{name: "Gi to I", in: float64(133.999111111), from: units.Gi, to: units.I, expected: float64(133999111111)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, units.ConvertUnits(test.in, test.from, test.to))
		})
	}
}

func ExampleConvertUnits() {
	conv := units.ConvertUnits(float64(100), units.Mi, units.I)
	fmt.Println(conv)
	// Output: 1e+08
}

func TestConvertUnitsString(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		from     units.Unit
		to       units.Unit
		expected float64
	}{
		{name: "Mi to I", in: "10.1", from: units.Gi, to: units.I, expected: float64(10100000000)},
		{name: "Gi to I", in: "133.999111111", from: units.Gi, to: units.I, expected: float64(133999111111)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := units.ConvertUnitsString(test.in, test.from, test.to)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, result)
		})
	}
}

func ExampleConvertUnitsString() {
	conv, err := units.ConvertUnitsString("10.1", units.Gi, units.I)
	if err != nil {
		// handle error
		return
	}
	fmt.Println(conv)
	// Output: 1.01e+10
}
