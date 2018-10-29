package units

import (
	"math"
	"strconv"
)

// Defines units of IOTAs.
type Unit float64

const (
	// The smallest Unit.
	I = Unit(1)
	// Kiloiota. 1000 iotas.
	Ki = Unit(1000)
	// Megaiota. 1 million iotas.
	Mi = Unit(1000000)
	// Gigaiota. 1 billion iotas.
	Gi = Unit(1000000000)
	// Teraiota. 1 trillion iotas.
	Ti = Unit(1000000000000)
	// Petaiota. 1 quadrillion iotas.
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

// ConvertUnits converts the given string value in the base Unit to the given new Unit.
func ConvertUnitString(value string, from Unit, to Unit) (float64, error) {
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}
	return ConvertUnits(floatValue, from, to), nil
}
