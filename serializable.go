package iota

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Serializable is something which knows how to serialize/deserialize itself from/into bytes.
// This is almost analogous to BinaryMarshaler/BinaryUnmarshaler.
type Serializable interface {
	json.Marshaler
	json.Unmarshaler
	// Deserialize deserializes the given data (by copying) into the object and returns the amount of bytes consumed from data.
	// If the passed data is not big enough for deserialization, an error must be returned.
	// During deserialization additional validation may be performed if the given modes are set.
	Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error)
	// Serialize returns a serialized byte representation.
	// This function does not check the serialized data for validity.
	// During serialization additional validation may be performed if the given modes are set.
	Serialize(deSeriMode DeSerializationMode) ([]byte, error)
}

// Serializables is a slice of Serializable.
type Serializables []Serializable

// SerializableSelectorFunc is a function that given a type byte, returns an empty instance of the given underlying type.
// If the type doesn't resolve, an error is returned.
type SerializableSelectorFunc func(ty uint32) (Serializable, error)

// DeSerializationMode defines the mode of de/serialization.
type DeSerializationMode byte

const (
	// Instructs de/serialization to perform no validation.
	DeSeriModeNoValidation DeSerializationMode = 0
	// Instructs de/serialization to perform validation.
	DeSeriModePerformValidation DeSerializationMode = 1 << 0
	// Instructs de/deserialization to perform ordering of certain struct arrays by their lexical serialized form.
	DeSeriModePerformLexicalOrdering DeSerializationMode = 1 << 1
)

// HasMode checks whether the de/serialization mode includes the given mode.
func (sm DeSerializationMode) HasMode(mode DeSerializationMode) bool {
	return sm&mode > 0
}

// ArrayValidationMode defines the mode of array elements validation.
type ArrayValidationMode byte

const (
	// Instructs the array element validation to perform no validation.
	ArrayValidModeNoValidation ArrayValidationMode = 0
	// Instructs the array element validation to check for duplicates.
	ArrayValidModeDuplicates ArrayValidationMode = 1 << 0
	// Instructs the array element validation to check for lexical order.
	ArrayValidModeLexicalOrdering ArrayValidationMode = 1 << 1
)

// HasMode checks whether the array element validation mode includes the given mode.
func (av ArrayValidationMode) HasMode(mode ArrayValidationMode) bool {
	return av&mode > 0
}

// ArrayRules defines rules around a to be deserialized array.
// Min and Max at 0 define an unbounded array.
type ArrayRules struct {
	// The min array bound.
	Min uint16
	// The max array bound.
	Max uint16
	// The mode of array elements validation.
	ValidationMode ArrayValidationMode
}

// CheckBounds checks whether the given count violates the array bounds.
func (ar *ArrayRules) CheckBounds(count uint16) error {
	if ar.Min != 0 && count < ar.Min {
		return fmt.Errorf("%w: min is %d but count is %d", ErrArrayValidationMinElementsNotReached, ar.Min, count)
	}
	if ar.Max != 0 && count > ar.Max {
		return fmt.Errorf("%w: max is %d but count is %d", ErrArrayValidationMaxElementsExceeded, ar.Max, count)
	}
	return nil
}

// ElementValidationFunc is a function which runs during array element validation (e.g. lexical ordering).
type ElementValidationFunc func(index int, next []byte) error

// ElementUniqueValidator returns a ElementValidationFunc which returns an error if the given element is not unique.
func (ar *ArrayRules) ElementUniqueValidator() ElementValidationFunc {
	set := map[string]struct{}{}
	return func(index int, next []byte) error {
		k := string(next)
		if j, has := set[k]; has {
			return fmt.Errorf("%w: element %d and %d are duplicates", ErrArrayValidationViolatesUniqueness, j, index)
		}
		set[k] = struct{}{}
		return nil
	}
}

// LexicalOrderValidator returns a ElementValidationFunc which returns an error if the given byte slices
// are not ordered lexicographically.
func (ar *ArrayRules) LexicalOrderValidator() ElementValidationFunc {
	var prev []byte
	var prevIndex int
	return func(index int, next []byte) error {
		switch {
		case prev == nil:
			prev = next
			prevIndex = index
		case bytes.Compare(prev, next) > 0:
			return fmt.Errorf("%w: element %d should have been before element %d", ErrArrayValidationOrderViolatesLexicalOrder, index, prevIndex)
		default:
			prev = next
			prevIndex = index
		}
		return nil
	}
}

// LexicalOrderWithoutDupsValidator returns a ElementValidationFunc which returns an error if the given byte slices
// are not ordered lexicographically or any elements are duplicated.
func (ar *ArrayRules) LexicalOrderWithoutDupsValidator() ElementValidationFunc {
	var prev []byte
	var prevIndex int
	return func(index int, next []byte) error {
		if prev == nil {
			prev = next
			prevIndex = index
			return nil
		}
		switch bytes.Compare(prev, next) {
		case 1:
			return fmt.Errorf("%w: element %d should have been before element %d", ErrArrayValidationOrderViolatesLexicalOrder, index, prevIndex)
		case 0:
			// dup
			return fmt.Errorf("%w: element %d and %d are duplicates", ErrArrayValidationViolatesUniqueness, index, prevIndex)
		}
		prev = next
		prevIndex = index
		return nil
	}
}

// LexicalOrderedByteSlices are byte slices ordered in lexical order.
type LexicalOrderedByteSlices [][]byte

func (l LexicalOrderedByteSlices) Len() int {
	return len(l)
}

func (l LexicalOrderedByteSlices) Less(i, j int) bool {
	return bytes.Compare(l[i], l[j]) < 0
}

func (l LexicalOrderedByteSlices) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// LexicalOrdered32ByteArrays are 32 byte arrays ordered in lexical order.
type LexicalOrdered32ByteArrays [][32]byte

func (l LexicalOrdered32ByteArrays) Len() int {
	return len(l)
}

func (l LexicalOrdered32ByteArrays) Less(i, j int) bool {
	return bytes.Compare(l[i][:], l[j][:]) < 0
}

func (l LexicalOrdered32ByteArrays) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// SortedSerializables are Serializables sorted by their serialized form.
type SortedSerializables Serializables

func (ss SortedSerializables) Len() int {
	return len(ss)
}

func (ss SortedSerializables) Less(i, j int) bool {
	iData, _ := ss[i].Serialize(DeSeriModeNoValidation)
	jData, _ := ss[j].Serialize(DeSeriModeNoValidation)
	return bytes.Compare(iData, jData) < 0
}

func (ss SortedSerializables) Swap(i, j int) {
	ss[i], ss[j] = ss[j], ss[i]
}
