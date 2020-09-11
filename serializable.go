package iota

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Serializable is something which knows how to serialize/deserialize itself from/into bytes.
// This is almost analogous to BinaryMarshaler/BinaryUnmarshaler.
type Serializable interface {
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
)

// HasMode checks whether the de/serialization mode includes the given mode.
func (sm DeSerializationMode) HasMode(mode DeSerializationMode) bool {
	return sm&mode == 1
}

// ArrayRules defines rules around a to be deserialized array.
// Min and Max at 0 define an unbounded array.
type ArrayRules struct {
	// The min array bound.
	Min uint16
	// The max array bound.
	Max uint16
	// The error returned if the min bound is violated.
	MinErr error
	// The error returned if the max bound is violated.
	MaxErr error
	// Whether the bytes of the elements have to be in lexical order.
	ElementBytesLexicalOrder bool
	// The error returned if the element bytes lexical order is violated.
	ElementBytesLexicalOrderErr error
}

// CheckBounds checks whether the given count violates the array bounds.
func (ar *ArrayRules) CheckBounds(count uint16) error {
	if ar.Min != 0 && count < ar.Min {
		return fmt.Errorf("%w: min is %d but count is %d", ar.MinErr, ar.Min, count)
	}
	if ar.Max != 0 && count > ar.Max {
		return fmt.Errorf("%w: max is %d but count is %d", ar.MaxErr, ar.Max, count)
	}
	return nil
}

// LexicalOrderFunc is a function which runs during lexical order validation.
type LexicalOrderFunc func(int, []byte) error

// LexicalOrderValidator returns a LexicalOrderFunc which returns an error if the given byte slices
// are not ordered lexicographically.
func (ar *ArrayRules) LexicalOrderValidator() LexicalOrderFunc {
	var prev []byte
	var prevIndex int
	return func(index int, next []byte) error {
		switch {
		case prev == nil:
			prev = next
			prevIndex = index
		case bytes.Compare(prev, next) > 0:
			return fmt.Errorf("%w: element %d should have been before element %d", ar.ElementBytesLexicalOrderErr, index, prevIndex)
		default:
			prev = next
			prevIndex = index
		}
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

// DeserializeArrayOfObjects deserializes the given data into Serializables.
// The data is expected to start with the count denoting varint, followed by the actual structs.
// An optional ArrayRules can be passed in to return an error in case it is violated.
func DeserializeArrayOfObjects(data []byte, deSeriMode DeSerializationMode, typeDen TypeDenotationType, serSel SerializableSelectorFunc, arrayRules *ArrayRules) (Serializables, int, error) {
	var bytesReadTotal int

	if len(data) < StructArrayLengthByteSize {
		return nil, 0, fmt.Errorf("%w: not enough data to deserialize struct array", ErrDeserializationNotEnoughData)
	}

	seriCount := binary.LittleEndian.Uint16(data)
	bytesReadTotal += StructArrayLengthByteSize

	if arrayRules != nil && deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := arrayRules.CheckBounds(seriCount); err != nil {
			return nil, 0, err
		}
	}

	// advance to objects
	var seris Serializables
	data = data[StructArrayLengthByteSize:]

	var lexicalOrderValidator LexicalOrderFunc
	if arrayRules != nil && arrayRules.ElementBytesLexicalOrder {
		lexicalOrderValidator = arrayRules.LexicalOrderValidator()
	}

	var offset int
	for i := 0; i < int(seriCount); i++ {
		seri, seriBytesConsumed, err := DeserializeObject(data[offset:], deSeriMode, typeDen, serSel)
		if err != nil {
			return nil, 0, err
		}
		// check lexical order against previous element
		if lexicalOrderValidator != nil {
			if err := lexicalOrderValidator(i, data[offset:offset+seriBytesConsumed]); err != nil {
				return nil, 0, err
			}
		}
		seris = append(seris, seri)
		offset += seriBytesConsumed
	}
	bytesReadTotal += offset

	return seris, bytesReadTotal, nil
}

// DeserializeObject deserializes the given data into a Serializable.
// The data is expected to start with the type denotation.
func DeserializeObject(data []byte, deSeriMode DeSerializationMode, typeDen TypeDenotationType, serSel SerializableSelectorFunc) (Serializable, int, error) {
	var ty uint32
	switch typeDen {
	case TypeDenotationUint32:
		if len(data) < UInt32ByteSize+1 {
			return nil, 0, ErrDeserializationNotEnoughData
		}
		ty = binary.LittleEndian.Uint32(data)
	case TypeDenotationByte:
		if len(data) < OneByte+1 {
			return nil, 0, ErrDeserializationNotEnoughData
		}
		ty = uint32(data[0])
	}
	seri, err := serSel(ty)
	if err != nil {
		return nil, 0, err
	}
	seriBytesConsumed, err := seri.Deserialize(data, deSeriMode)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to deserialize %T: %w", seri, err)
	}
	return seri, seriBytesConsumed, nil
}

// ReadStringFromBytes reads a string from data by first reading the string length by reading a uint16
// and then consuming that length from data.
func ReadStringFromBytes(data []byte) (string, int, error) {
	if len(data) < UInt16ByteSize {
		return "", 0, fmt.Errorf("%w: can't read string length", ErrDeserializationNotEnoughData)
	}

	strLen := binary.LittleEndian.Uint16(data)
	data = data[UInt16ByteSize:]

	if len(data) < int(strLen) {
		return "", 0, fmt.Errorf("%w: data is smaller than (%d) denoted string length of %d", ErrDeserializationNotEnoughData, len(data), strLen)
	}

	return string(data[:strLen]), int(strLen) + UInt16ByteSize, nil
}
