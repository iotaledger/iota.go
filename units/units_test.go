package units

import (
	"testing"
)

type testcasefloat struct {
	value    float64
	fromUnit Unit
	toUnit   Unit
	expected float64
}

type testcasestring struct {
	value    string
	fromUnit Unit
	toUnit   Unit
	expected float64
}

var testcasesFloats = []testcasefloat{
	{
		value: 100, fromUnit: Mi, toUnit: I,
		expected: 100000000,
	},
	{
		value: 10.1, fromUnit: Gi, toUnit: I,
		expected: 10100000000,
	},
	{
		value: 1, fromUnit: I, toUnit: Ti,
		expected: 0.000000000001,
	},
	{
		value: 1, fromUnit: Ti, toUnit: I,
		expected: 1000000000000,
	},
	{
		value: 1000, fromUnit: Gi, toUnit: Ti,
		expected: 1,
	},
	{
		value: 133.999111111, fromUnit: Gi, toUnit: I,
		expected: 133999111111,
	},
}

var testcasesString = []testcasestring{
	{
		value: "10.1", fromUnit: Gi, toUnit: I,
		expected: 10100000000,
	},
	{
		value: "133.999111111", fromUnit: Gi, toUnit: I,
		expected: 133999111111,
	},
}

func TestConvertUnitsFloats(t *testing.T) {
	for i, test := range testcasesFloats {
		result := ConvertUnits(test.value, test.fromUnit, test.toUnit)
		if result != test.expected {
			t.Fatalf("#%d expected input value %v to be %.10f but was %v\n", i, test.value, test.expected, result)
		}
	}
}

func TestConvertUnitsString(t *testing.T) {
	for i, test := range testcasesString {
		result, err := ConvertUnitString(test.value, test.fromUnit, test.toUnit)
		if err != nil {
			t.Fatalf("expected no error but got %v\n", err)
		}
		if result != test.expected {
			t.Fatalf("#%d expected input value %v to be %.10f but was %v\n", i, test.value, test.expected, result)
		}
	}
}
