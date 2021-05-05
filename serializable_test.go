package iotago_test

import (
	"errors"
	"github.com/iotaledger/iota.go/v2/tpkg"
	"sort"
	"testing"

	"github.com/iotaledger/iota.go/v2"
	"github.com/stretchr/testify/assert"
)

const (
	TypeA       byte = 0
	TypeB       byte = 1
	aKeyLength       = 16
	bNameLength      = 32
	typeALength      = iotago.SmallTypeDenotationByteSize + aKeyLength
	typeBLength      = iotago.SmallTypeDenotationByteSize + bNameLength
)

var (
	ErrUnknownDummyType = errors.New("unknown example type")
)

func DummyTypeSelector(dummyType uint32) (iotago.Serializable, error) {
	var seri iotago.Serializable
	switch byte(dummyType) {
	case TypeA:
		seri = &A{}
	case TypeB:
		seri = &B{}
	default:
		return nil, ErrUnknownDummyType
	}
	return seri, nil
}

type A struct {
	Key [aKeyLength]byte
}

func (a *A) MarshalJSON() ([]byte, error) {
	panic("implement me")
}

func (a *A) UnmarshalJSON(i []byte) error {
	panic("implement me")
}

func (a *A) Deserialize(data []byte, deSeriMode iotago.DeSerializationMode) (int, error) {
	data = data[iotago.SmallTypeDenotationByteSize:]
	copy(a.Key[:], data[:aKeyLength])
	return typeALength, nil
}

func (a *A) Serialize(deSeriMode iotago.DeSerializationMode) ([]byte, error) {
	var b [typeALength]byte
	b[0] = TypeA
	copy(b[iotago.SmallTypeDenotationByteSize:], a.Key[:])
	return b[:], nil
}

func randSerializedA() []byte {
	var b [typeALength]byte
	b[0] = TypeA
	keyData := tpkg.RandBytes(aKeyLength)
	copy(b[iotago.SmallTypeDenotationByteSize:], keyData)
	return b[:]
}

func randA() *A {
	var k [aKeyLength]byte
	copy(k[:], tpkg.RandBytes(aKeyLength))
	return &A{Key: k}
}

type B struct {
	Name [bNameLength]byte
}

func (b *B) MarshalJSON() ([]byte, error) {
	panic("implement me")
}

func (b *B) UnmarshalJSON(i []byte) error {
	panic("implement me")
}

func (b *B) Deserialize(data []byte, deSeriMode iotago.DeSerializationMode) (int, error) {
	data = data[iotago.SmallTypeDenotationByteSize:]
	copy(b.Name[:], data[:bNameLength])
	return typeBLength, nil
}

func (b *B) Serialize(deSeriMode iotago.DeSerializationMode) ([]byte, error) {
	var bf [typeBLength]byte
	bf[0] = TypeB
	copy(bf[iotago.SmallTypeDenotationByteSize:], b.Name[:])
	return bf[:], nil
}

func randSerializedB() []byte {
	var bf [typeBLength]byte
	bf[0] = TypeB
	nameData := tpkg.RandBytes(bNameLength)
	copy(bf[iotago.SmallTypeDenotationByteSize:], nameData)
	return bf[:]
}

func randB() *B {
	var n [bNameLength]byte
	copy(n[:], tpkg.RandBytes(bNameLength))
	return &B{Name: n}
}

func TestDeserializeA(t *testing.T) {
	seriA := randSerializedA()
	objA := &A{}
	bytesRead, err := objA.Deserialize(seriA, iotago.DeSeriModePerformValidation)
	assert.NoError(t, err)
	assert.Equal(t, len(seriA), bytesRead)
	assert.Equal(t, seriA[iotago.SmallTypeDenotationByteSize:], objA.Key[:])
}

func TestLexicalOrderedByteSlices(t *testing.T) {
	type test struct {
		name   string
		source iotago.LexicalOrderedByteSlices
		target iotago.LexicalOrderedByteSlices
	}
	tests := []test{
		{
			name: "ok - order by first ele",
			source: iotago.LexicalOrderedByteSlices{
				{3, 2, 1},
				{2, 3, 1},
				{1, 2, 3},
			},
			target: iotago.LexicalOrderedByteSlices{
				{1, 2, 3},
				{2, 3, 1},
				{3, 2, 1},
			},
		},
		{
			name: "ok - order by last ele",
			source: iotago.LexicalOrderedByteSlices{
				{1, 1, 3},
				{1, 1, 2},
				{1, 1, 1},
			},
			target: iotago.LexicalOrderedByteSlices{
				{1, 1, 1},
				{1, 1, 2},
				{1, 1, 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sort.Sort(tt.source)
			assert.Equal(t, tt.target, tt.source)
		})
	}
}

func TestRemoveDupsAndSortByLexicalOrderArrayOf32Bytes(t *testing.T) {
	type test struct {
		name   string
		source iotago.LexicalOrdered32ByteArrays
		target iotago.LexicalOrdered32ByteArrays
	}
	tests := []test{
		{
			name: "ok - dups removed and order by first ele",
			source: iotago.LexicalOrdered32ByteArrays{
				{3, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
				{3, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
				{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
				{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
			},
			target: iotago.LexicalOrdered32ByteArrays{
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
				{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
				{3, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
			},
		},
		{
			name: "ok - dups removed and order by last ele",
			source: iotago.LexicalOrdered32ByteArrays{
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 34},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 34},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 33},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
			},
			target: iotago.LexicalOrdered32ByteArrays{
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 33},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 34},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.source = iotago.RemoveDupsAndSortByLexicalOrderArrayOf32Bytes(tt.source)
			assert.Equal(t, tt.target, tt.source)
		})
	}
}

func TestSerializationMode_HasMode(t *testing.T) {
	type args struct {
		mode iotago.DeSerializationMode
	}
	tests := []struct {
		name string
		sm   iotago.DeSerializationMode
		args args
		want bool
	}{
		{
			"has no validation",
			iotago.DeSeriModeNoValidation,
			args{mode: iotago.DeSeriModePerformValidation},
			false,
		},
		{
			"has validation",
			iotago.DeSeriModePerformValidation,
			args{mode: iotago.DeSeriModePerformValidation},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sm.HasMode(tt.args.mode); got != tt.want {
				t.Errorf("HasMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArrayValidationMode_HasMode(t *testing.T) {
	type args struct {
		mode iotago.ArrayValidationMode
	}
	tests := []struct {
		name string
		sm   iotago.ArrayValidationMode
		args args
		want bool
	}{
		{
			"has no validation",
			iotago.ArrayValidationModeNone,
			args{mode: iotago.ArrayValidationModeNoDuplicates},
			false,
		},
		{
			"has mode duplicates",
			iotago.ArrayValidationModeNoDuplicates,
			args{mode: iotago.ArrayValidationModeNoDuplicates},
			true,
		},
		{
			"has mode lexical order",
			iotago.ArrayValidationModeLexicalOrdering,
			args{mode: iotago.ArrayValidationModeLexicalOrdering},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sm.HasMode(tt.args.mode); got != tt.want {
				t.Errorf("HasMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArrayRules_ElementUniqueValidator(t *testing.T) {
	type test struct {
		name  string
		args  [][]byte
		valid bool
	}

	arrayRules := iotago.ArrayRules{}

	tests := []test{
		{
			name: "ok - no dups",
			args: [][]byte{
				{1, 2, 3},
				{2, 3, 1},
				{3, 2, 1},
			},
			valid: true,
		},
		{
			name: "not ok - dups",
			args: [][]byte{
				{1, 1, 1},
				{1, 1, 2},
				{1, 1, 3},
				{1, 1, 3},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arrayElementValidator := arrayRules.ElementUniqueValidator()

			valid := true
			for i := range tt.args {
				element := tt.args[i]

				if err := arrayElementValidator(i, element); err != nil {
					valid = false
				}
			}

			assert.Equal(t, tt.valid, valid)
		})
	}
}

func TestArrayRules_Bounds(t *testing.T) {
	type test struct {
		name  string
		args  [][]byte
		min   int
		max   int
		valid bool
	}

	arrayRules := iotago.ArrayRules{}

	tests := []test{
		{
			name: "ok - min",
			args: [][]byte{
				{1},
			},
			min:   1,
			max:   3,
			valid: true,
		},
		{
			name: "ok - max",
			args: [][]byte{
				{1},
				{2},
				{3},
			},
			min:   1,
			max:   3,
			valid: true,
		},
		{
			name: "not ok - min",
			args: [][]byte{
				{1},
				{2},
				{3},
			},
			min:   4,
			max:   5,
			valid: false,
		},
		{
			name: "not ok - max",
			args: [][]byte{
				{1},
				{2},
				{3},
			},
			min:   1,
			max:   2,
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arrayRules.Min = uint16(tt.min)
			arrayRules.Max = uint16(tt.max)
			err := arrayRules.CheckBounds(uint16(len(tt.args)))
			assert.Equal(t, tt.valid, err == nil)
		})
	}
}

func TestArrayRules_LexicalOrderValidator(t *testing.T) {
	type test struct {
		name  string
		args  [][]byte
		valid bool
	}

	arrayRules := iotago.ArrayRules{}

	tests := []test{
		{
			name: "ok - order by first ele",
			args: [][]byte{
				{1, 2, 3},
				{2, 3, 1},
				{3, 2, 1},
			},
			valid: true,
		},
		{
			name: "ok - order by last ele",
			args: [][]byte{
				{1, 1, 1},
				{1, 1, 2},
				{1, 1, 3},
			},
			valid: true,
		},
		{
			name: "not ok",
			args: [][]byte{
				{2, 1, 1},
				{1, 1, 2},
				{3, 1, 3},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arrayElementValidator := arrayRules.LexicalOrderValidator()

			valid := true
			for i := range tt.args {
				element := tt.args[i]

				if err := arrayElementValidator(i, element); err != nil {
					valid = false
				}
			}

			assert.Equal(t, tt.valid, valid)
		})
	}
}

func TestArrayRules_LexicalOrderWithoutDupsValidator(t *testing.T) {
	type test struct {
		name  string
		args  [][]byte
		valid bool
	}

	arrayRules := iotago.ArrayRules{}

	tests := []test{
		{
			name: "ok - order by first ele - no dups",
			args: [][]byte{
				{1, 2, 3},
				{2, 3, 1},
				{3, 2, 1},
			},
			valid: true,
		},
		{
			name: "ok - order by last ele - no dups",
			args: [][]byte{
				{1, 1, 1},
				{1, 1, 2},
				{1, 1, 3},
			},
			valid: true,
		},
		{
			name: "not ok - dups",
			args: [][]byte{
				{1, 1, 1},
				{1, 1, 2},
				{1, 1, 3},
				{1, 1, 3},
			},
			valid: false,
		},
		{
			name: "not ok - order",
			args: [][]byte{
				{2, 1, 1},
				{1, 1, 2},
				{3, 1, 3},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arrayElementValidator := arrayRules.LexicalOrderWithoutDupsValidator()

			valid := true
			for i := range tt.args {
				element := tt.args[i]

				if err := arrayElementValidator(i, element); err != nil {
					valid = false
				}
			}

			assert.Equal(t, tt.valid, valid)
		})
	}
}
