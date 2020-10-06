package iota_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"sort"
	"testing"

	"github.com/iotaledger/iota.go"
	"github.com/stretchr/testify/assert"
)

const (
	TypeA       byte = 0
	TypeB       byte = 1
	aKeyLength       = 16
	bNameLength      = 32
	typeALength      = iota.SmallTypeDenotationByteSize + aKeyLength
	typeBLength      = iota.SmallTypeDenotationByteSize + bNameLength
)

var (
	ErrUnknownDummyType = errors.New("unknown example type")
)

func DummyTypeSelector(dummyType uint32) (iota.Serializable, error) {
	var seri iota.Serializable
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

func (a *A) Deserialize(data []byte, deSeriMode iota.DeSerializationMode) (int, error) {
	data = data[iota.SmallTypeDenotationByteSize:]
	copy(a.Key[:], data[:aKeyLength])
	return typeALength, nil
}

func (a *A) Serialize(deSeriMode iota.DeSerializationMode) ([]byte, error) {
	var b [typeALength]byte
	b[0] = TypeA
	copy(b[iota.SmallTypeDenotationByteSize:], a.Key[:])
	return b[:], nil
}

func randSerializedA() []byte {
	var b [typeALength]byte
	b[0] = TypeA
	keyData := randBytes(aKeyLength)
	copy(b[iota.SmallTypeDenotationByteSize:], keyData)
	return b[:]
}

func randA() *A {
	var k [aKeyLength]byte
	copy(k[:], randBytes(aKeyLength))
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

func (b *B) Deserialize(data []byte, deSeriMode iota.DeSerializationMode) (int, error) {
	data = data[iota.SmallTypeDenotationByteSize:]
	copy(b.Name[:], data[:bNameLength])
	return typeBLength, nil
}

func (b *B) Serialize(deSeriMode iota.DeSerializationMode) ([]byte, error) {
	var bf [typeBLength]byte
	bf[0] = TypeB
	copy(bf[iota.SmallTypeDenotationByteSize:], b.Name[:])
	return bf[:], nil
}

func randSerializedB() []byte {
	var bf [typeBLength]byte
	bf[0] = TypeB
	nameData := randBytes(bNameLength)
	copy(bf[iota.SmallTypeDenotationByteSize:], nameData)
	return bf[:]
}

func randB() *B {
	var n [bNameLength]byte
	copy(n[:], randBytes(bNameLength))
	return &B{Name: n}
}

func TestDeserializeA(t *testing.T) {
	seriA := randSerializedA()
	objA := &A{}
	bytesRead, err := objA.Deserialize(seriA, iota.DeSeriModePerformValidation)
	assert.NoError(t, err)
	assert.Equal(t, len(seriA), bytesRead)
	assert.Equal(t, seriA[iota.SmallTypeDenotationByteSize:], objA.Key[:])
}

func TestDeserializeObject(t *testing.T) {
	seriA := randSerializedA()
	objA, bytesRead, err := iota.DeserializeObject(seriA, iota.DeSeriModePerformValidation, iota.TypeDenotationByte, DummyTypeSelector)
	assert.NoError(t, err)
	assert.Equal(t, len(seriA), bytesRead)
	assert.IsType(t, &A{}, objA)
	assert.Equal(t, seriA[iota.SmallTypeDenotationByteSize:], objA.(*A).Key[:])
}

func TestDeserializeArrayOfObjects(t *testing.T) {
	var buf bytes.Buffer
	originObjs := iota.Serializables{
		randA(), randA(), randB(), randA(), randB(), randB(),
	}
	assert.NoError(t, binary.Write(&buf, binary.LittleEndian, uint16(len(originObjs))))

	for _, seri := range originObjs {
		seriBytes, err := seri.Serialize(iota.DeSeriModePerformValidation)
		assert.NoError(t, err)
		written, err := buf.Write(seriBytes)
		assert.NoError(t, err)
		assert.Equal(t, len(seriBytes), written)
	}

	data := buf.Bytes()
	seris, serisByteRead, err := iota.DeserializeArrayOfObjects(data, iota.DeSeriModePerformValidation, iota.TypeDenotationByte, DummyTypeSelector, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(data), serisByteRead)
	assert.EqualValues(t, originObjs, seris)
}

func TestLexicalOrderedByteSlices(t *testing.T) {
	type test struct {
		name   string
		source iota.LexicalOrderedByteSlices
		target iota.LexicalOrderedByteSlices
	}
	tests := []test{
		{
			name: "ok - order by first ele",
			source: iota.LexicalOrderedByteSlices{
				{3, 2, 1},
				{2, 3, 1},
				{1, 2, 3},
			},
			target: iota.LexicalOrderedByteSlices{
				{1, 2, 3},
				{2, 3, 1},
				{3, 2, 1},
			},
		},
		{
			name: "ok - order by last ele",
			source: iota.LexicalOrderedByteSlices{
				{1, 1, 3},
				{1, 1, 2},
				{1, 1, 1},
			},
			target: iota.LexicalOrderedByteSlices{
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

func TestSerializationMode_HasMode(t *testing.T) {
	type args struct {
		mode iota.DeSerializationMode
	}
	tests := []struct {
		name string
		sm   iota.DeSerializationMode
		args args
		want bool
	}{
		{
			"has no validation",
			iota.DeSeriModeNoValidation,
			args{mode: iota.DeSeriModePerformValidation},
			false,
		},
		{
			"has validation",
			iota.DeSeriModePerformValidation,
			args{mode: iota.DeSeriModePerformValidation},
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

func TestReadStringFromBytes(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				data: []byte{17, 0, 72, 101, 108, 108, 111, 44, 32, 112, 108, 97, 121, 103, 114, 111, 117, 110, 100},
			},
			want:    "Hello, playground",
			wantErr: false,
		},
		{
			name: "not enough (length denotation)",
			args: args{
				data: []byte{0, 1},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "not enough (actual length)",
			args: args{
				data: []byte{17, 0, 72, 101, 108, 108, 111, 44, 32, 112},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := iota.ReadStringFromBytes(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadStringFromBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReadStringFromBytes() got = %v, want %v", got, tt.want)
			}
		})
	}
}
