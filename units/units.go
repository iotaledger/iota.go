package units

import (
	"math"
	"strconv"
)

type Unit float64

const (
	I  = Unit(1)
	Ki = Unit(1000)
	Mi = Unit(1000000)
	Gi = Unit(1000000000)
	Ti = Unit(1000000000000)
	Pi = Unit(1000000000000000)
)

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

func ConvertUnitString(value string, from Unit, to Unit) (float64, error) {
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}
	return ConvertUnits(floatValue, from, to), nil
}
