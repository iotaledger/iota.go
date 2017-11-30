/*
MIT License
Copyright (c) 2017 Harry Boer

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

//Package stringutils is a utility to create Masked Authenticated Messages (MAM's)
package stringutils

import (
	"fmt"
	"strings"

	"github.com/iotaledger/giota"
)

//ToMAMTrytes checks its validity and casts to giota.Trytes.
func ToMAMTrytes(t string) (tr giota.Trytes) {

	trytes := ""
	TryteValues := giota.TryteAlphabet

	for _, number := range t {
		asciiValue := CharCodeAt(string(number), 0)
		if asciiValue > 255 {
			//make it a space
			asciiValue = 32
		}
		firstValue := asciiValue % 27
		secondValue := (asciiValue - firstValue) / 27
		trytesValue := string(TryteValues[firstValue]) + string(TryteValues[secondValue])
		trytes = trytes + trytesValue
		// fmt.Println()
		// for iter, number := range t {
		// 	fmt.Println(iter, number, string(number))
		// }
	}

	newTrytes := giota.Trytes(trytes)

	return newTrytes
}

//FromMAMTrytes converts the MAM from giota.Trytes to a readable string
func FromMAMTrytes(inputTrytes giota.Trytes) string {

	outputString := ""
	TryteValues := giota.TryteAlphabet
	// Check if input is giota.Trytes
	err := IsValidTrytes(inputTrytes)
	if err != nil {
		fmt.Println("Error: ", err)
	}

	for i := 0; i < len(inputTrytes); i += 2 {
		// get a trytes pair
		trytes := string(inputTrytes[i]) + string(inputTrytes[i+1])
		firstValue := strings.Index(TryteValues, string(trytes[0]))
		secondValue := strings.Index(TryteValues, string(trytes[1]))
		decimalValue := firstValue + secondValue*27
		character := string(decimalValue)
		outputString = outputString + character
	}
	fmt.Println("Output string is: ", outputString)
	return outputString
}
