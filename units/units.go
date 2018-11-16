// Package units provides functions for converting different units of IOTAs.
package units

import (
	"math"
	"strconv"
)

// Unit a unit of IOTAs.
type Unit float64

const (
	// I is the smallest Unit.
	I = Unit(1)
	// Ki = Kiloiota. 1000 iotas.
	Ki = Unit(1000)
	// Mi = Megaiota. 1 million iotas.
	Mi = Unit(1000000)
	// Gi = Gigaiota. 1 billion iotas.
	Gi = Unit(1000000000)
	// Ti = Teraiota. 1 trillion iotas.
	Ti = Unit(1000000000000)
	// Pi = Petaiota. 1 quadrillion iotas.
	Pi = Unit(1000000000000000)
)

// ConvertUnits converts the given value in the base Unit to the given new Unit.
func ConvertUnits(val float64, from Unit, to Unit) float64 {
	value := Unit(val)
	// convert to I unit by multiplying with the current unit
	value *= from
	// convert to the target unit by dividing by it
	if to == I {
		return math.Round(float64(value))
	}
	value /= to
	return float64(value)
}

// ConvertUnitsString converts the given string value in the base Unit to the given new Unit.
func ConvertUnitsString(val string, from Unit, to Unit) (float64, error) {
	floatValue, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, err
	}
	return ConvertUnits(floatValue, from, to), nil
}
