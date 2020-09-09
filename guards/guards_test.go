package guards_test

import (
	"fmt"
	"testing"

	"github.com/iotaledger/iota.go/guards"
	"github.com/iotaledger/iota.go/trinary"
	"github.com/stretchr/testify/assert"
)

func TestIsTrytes(t *testing.T) {
	tests := []struct {
		name   string
		trytes trinary.Trytes
		is     bool
	}{
		{name: "valid hash", trytes: "ABC", is: true},
		{name: "invalid - spaces", trytes: "A B C", is: false},
		{name: "invalid - lowercased", trytes: "abc", is: false},
		{name: "invalid - empty", trytes: "", is: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			is := guards.IsTrytes(test.trytes)
			assert.Equal(t, test.is, is)
		})
	}
}

func TestIsTrytesOfExactLength(t *testing.T) {
	tests := []struct {
		name   string
		trytes trinary.Trytes
		length int
		is     bool
	}{
		{name: "valid", trytes: "ABC", length: 3, is: true},
		{name: "invalid - lowercased", trytes: "abc", length: 3, is: false},
		{name: "invalid - empty", trytes: "", length: 0, is: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			is := guards.IsTrytesOfExactLength(test.trytes, test.length)
			assert.Equal(t, test.is, is)
		})
	}
}

func TestIsTrytesOfMaxLength(t *testing.T) {
	tests := []struct {
		name   string
		trytes trinary.Trytes
		length int
		is     bool
	}{
		{name: "valid", trytes: "A", length: 3, is: true},
		{name: "invalid - lowercased", trytes: "abc", length: 3, is: false},
		{name: "invalid - empty", trytes: "", length: 0, is: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			is := guards.IsTrytesOfMaxLength(test.trytes, test.length)
			assert.Equal(t, test.is, is)
		})
	}
}

func TestIsEmptyTrytes(t *testing.T) {
	tests := []struct {
		name   string
		trytes trinary.Trytes
		is     bool
	}{
		{name: "valid empty hash", trytes: "9999", is: true},
		{name: "non empty", trytes: "A99", is: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			is := guards.IsEmptyTrytes(test.trytes)
			assert.Equal(t, test.is, is)
		})
	}
}

func TestIsHash(t *testing.T) {
	tests := []struct {
		name string
		hash trinary.Trytes
		is   bool
	}{
		{name: "valid", hash: "TXBGJB9NORCEHAAWVCQRC9GQSLQCWUIKDOBYTDKVYY9GUQHPJQMKHGNWRWIFLEBPJNAAIOMUFRFLDQUEC", is: true},
		{name: "valid - with checksum", hash: "TXBGJB9NORCEHAAWVCQRC9GQSLQCWUIKDOBYTDKVYY9GUQHPJQMKHGNWRWIFLEBPJNAAIOMUFRFLDQUECB9UMGFVBD", is: true},
		{name: "invalid", hash: "ABCD", is: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			is := guards.IsHash(test.hash)
			assert.Equal(t, test.is, is)
		})
	}
}

func ExampleIsTrytes() {
	is := guards.IsTrytes("ABCD")
	fmt.Println(is)
	// Output: true
}

func ExampleIsTrytesOfExactLength() {
	fmt.Println(guards.IsTrytesOfExactLength("ABCD", 4), guards.IsTrytesOfExactLength("ABCD", 2))
	// Output: true false
}

func ExampleIsTrytesOfMaxLength() {
	fmt.Println(guards.IsTrytesOfMaxLength("ABCD", 5), guards.IsTrytesOfMaxLength("ABCD", 2))
	// Output: true false
}

func ExampleIsEmptyTrytes() {
	fmt.Println(guards.IsEmptyTrytes("99999999"), guards.IsEmptyTrytes("ABCD"))
	// Output: true false
}

func ExampleIsHash() {
	hash := "ZFPPXWSTIYJCPPMVCCBZR9TISFJALXEXVYMADGTERQLTHAZJMHGWWFIXVCVPJRBUYLKMTLLKMTWMA9999"
	fmt.Println(guards.IsHash(hash))
	// Output: true
}
