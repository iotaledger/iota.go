package iotago

const (
	// OneByte is the byte size of a single byte.
	OneByte = 1
	// Int16ByteSize is the byte size of an int16.
	Int16ByteSize = 2
	// UInt16ByteSize is the byte size of a uint16.
	UInt16ByteSize = 2
	// Int32ByteSize is the byte size of an int32.
	Int32ByteSize = 4
	// UInt32ByteSize is the byte size of a uint32.
	UInt32ByteSize = 4
	// Float32ByteSize is the byte size of a float32.
	Float32ByteSize = 4
	// Int64ByteSize is the byte size of an int64.
	Int64ByteSize = 8
	// UInt64ByteSize is the byte size of an uint64.
	UInt64ByteSize = 8
	// Float64ByteSize is the byte size of a float64.
	Float64ByteSize = 8
	// TypeDenotationByteSize is the size of a type denotation.
	TypeDenotationByteSize = UInt32ByteSize
	// SmallTypeDenotationByteSize is the size of a type denotation for a small range of possible values.
	SmallTypeDenotationByteSize = OneByte
	// StructArrayLengthByteSize is the byte size of struct array lengths.
	StructArrayLengthByteSize = UInt16ByteSize
	// ByteArrayLengthByteSize is the byte size of byte array lengths.
	ByteArrayLengthByteSize = UInt32ByteSize
	// PayloadLengthByteSize is the size of the payload length denoting bytes.
	PayloadLengthByteSize = UInt32ByteSize
	// MinPayloadByteSize is the minimum size of a payload (together with its length denotation).
	MinPayloadByteSize = UInt32ByteSize + OneByte
	// TokenSupply is the IOTA token supply.
	TokenSupply = 2_779_530_283_277_761
)

// TypeDenotationType defines a type denotation.
type TypeDenotationType byte

const (
	// TypeDenotationUint32 defines a denotation which defines a type ID by a uint32.
	TypeDenotationUint32 TypeDenotationType = iota
	// TypeDenotationByte defines a denotation which defines a type ID by a byte.
	TypeDenotationByte
	// TypeDenotationNone defines that there is no type denotation.
	TypeDenotationNone
)
