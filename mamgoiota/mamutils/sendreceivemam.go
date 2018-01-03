/*
MIT License
Copyright (c) 2017 Harry Boer, Jonah Polack

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

//Package mamutils is a utility to create Masked Authenticated Messages (MAM's)
package mamutils

import (
	"errors"
	"strings"

	"github.com/iotaledger/giota"
)

//ToMAMTrytes checks its validity and casts to giota.Trytes.
//All ASCII characters greater than number 255 will be converted to a space.
func ToMAMTrytes(message string) (giota.Trytes, error) {

	trytes := ""
	TryteValues := giota.TryteAlphabet

	for _, number := range message {
		if number > 255 {
			//make it a space
			number = 32
		}
		firstValue := number % 27
		secondValue := (number - firstValue) / 27
		trytesValue := string(TryteValues[firstValue]) + string(TryteValues[secondValue])
		trytes = trytes + trytesValue
	}

	newTrytes, err := giota.ToTrytes(trytes)
	if err != nil {
		return "", err
	}

	return newTrytes, nil
}

//FromMAMTrytes converts the MAM from giota.Trytes to a readable string
func FromMAMTrytes(inputTrytes giota.Trytes) (string, error) {

	trimmed := strings.TrimRight(string(inputTrytes), "9")
	var err error
	inputTrytes, err = giota.ToTrytes(trimmed)
	if err != nil {
		return "", err
	}

	outputString := ""
	TryteValues := giota.TryteAlphabet

	// Check if input is an even number of giota.Trytes
	err = IsValidTrytes(inputTrytes)
	if err != nil {
		err := errors.New("wrong trytes input")
		return "", err

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
	//fmt.Println("Output string is: ", outputString)
	return outputString, nil
}

//IsValidTrytes checks wether type and length of Trytes are valid
func IsValidTrytes(t giota.Trytes) error {
	if len(t)%2 != 0 {
		err := errors.New("wrong number of trytes; number should be even")
		return err
	}
	return nil
}
