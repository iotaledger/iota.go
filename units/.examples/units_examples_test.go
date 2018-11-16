package units_examples_test

import (
	"fmt"
	"github.com/iotaledger/iota.go/units"
)

// i req: val, The source value.
// i req: from, The Unit format of the source value.
// i req: to, The Unit format of the target value.
// o: float64, The float64 representation of the target value.
func ExampleConvertUnits() {
	conv := units.ConvertUnits(float64(100), units.Mi, units.I)
	fmt.Println(conv)
	// output: 100000000
}

// i req: val, The source string value.
// i req: from, The Unit format of the source value.
// i req: to, The Unit format of the target value.
// o: float64, The float64 representation of the target value.
// o: error, Returned for invalid string values.
func ExampleConvertUnitsString() {
	conv, err := units.ConvertUnitsString("10.1", units.Gi, units.I)
	if err != nil {
		// handle error
		return
	}
	fmt.Println(conv)
	// output: 10100000000
}
