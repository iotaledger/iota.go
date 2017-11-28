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

	// Check if input is a string
	if !IsString(t) {
		fmt.Println("Input is not a string. Please provide a valid string.")
	}

	trytes := ""
	TryteValues := "9ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	for i := 0; i < len(t); i++ {
		asciiValue := CharCodeAt(string(t[i]), 0)
		if asciiValue > 255 {
			asciiValue = 32
		}
		firstValue := asciiValue % 27
		secondValue := (asciiValue - firstValue) / 27
		trytesValue := string(TryteValues[firstValue]) + string(TryteValues[secondValue])
		trytes = trytes + trytesValue
	}
	fmt.Println("trytes is: ", trytes)

	newTrytes := giota.Trytes(trytes)

	return newTrytes
}

//FromMAMTrytes converts the MAM from giota.Trytes to a readable string
func FromMAMTrytes(inputTrytes giota.Trytes) string {
	// character := ""
	outputString := ""
	TryteValues := "9ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// Check if input is a string
	if IsString(inputTrytes) {
		fmt.Println("Input is not a giota.Trytes. Please provide a valid argument.")
	}
	// Check of we have an even length
	if len(inputTrytes)%2 != 0 {
		fmt.Println("Error. Wrong number of trytes")
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

	return outputString
}
