package iota

const (
	// The byte size of a single byte.
	OneByte = 1
	// The byte size of an int16.
	Int16ByteSize = 2
	// The byte size of a uint16.
	UInt16ByteSize = 2
	// The byte size of an int32.
	Int32ByteSize = 4
	// The byte size of a uint32.
	UInt32ByteSize = 4
	// The byte size of a float32.
	Float32ByteSize = 4
	// The byte size of an int64.
	Int64ByteSize = 8
	// The byte size of an uint64.
	UInt64ByteSize = 8
	// The byte size of a float64.
	Float64ByteSize = 8
	// The size of a type denotation.
	TypeDenotationByteSize = UInt32ByteSize
	// The size of a type denotation for a small range of possible values.
	SmallTypeDenotationByteSize = OneByte
	// // The byte size of struct array lengths.
	StructArrayLengthByteSize = UInt16ByteSize
	// // The byte size of byte array lengths.
	ByteArrayLengthByteSize = UInt32ByteSize
	// The size of the payload length denoting bytes.
	PayloadLengthByteSize = UInt32ByteSize
	// The minimum size of a payload (together with its length denotation).
	MinPayloadByteSize = UInt32ByteSize + OneByte
	// The IOTA token supply.
	TokenSupply = 2_779_530_283_277_761
)

type TypeDenotationType byte

const (
	// Defines a denotation which defines a type ID by a uint32.
	TypeDenotationUint32 TypeDenotationType = iota
	// Defines a denotation which defines a type ID by a byte.
	TypeDenotationByte
	// Defines that there is no type denotation.
	TypeDenotationNone
)
