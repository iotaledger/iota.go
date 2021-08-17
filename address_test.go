package iotago_test

import (
	"errors"
	"github.com/iotaledger/hive.go/serializer"
	"github.com/iotaledger/iota.go/v2/tpkg"
	"testing"

	"github.com/iotaledger/iota.go/v2"
	"github.com/stretchr/testify/assert"
)

func TestEd25519Address_Deserialize(t *testing.T) {
	tests := []struct {
		name       string
		edAddrData []byte
		err        error
	}{
		{
			"ok",
			func() []byte {
				_, edAddrData := tpkg.RandEd25519Address()
				return edAddrData
			}(),
			nil,
		},
		{
			"not enough bytes",
			func() []byte {
				_, edAddrData := tpkg.RandEd25519Address()
				return edAddrData[:iotago.Ed25519AddressSerializedBytesSize-1]
			}(),
			serializer.ErrDeserializationNotEnoughData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edAddr := &iotago.Ed25519Address{}
			bytesRead, err := edAddr.Deserialize(tt.edAddrData, serializer.DeSeriModePerformValidation)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, len(tt.edAddrData), bytesRead)
			assert.Equal(t, tt.edAddrData[serializer.SmallTypeDenotationByteSize:], edAddr[:])
		})
	}
}

func TestEd25519Address_Serialize(t *testing.T) {
	originEdAddr, originData := tpkg.RandEd25519Address()
	tests := []struct {
		name   string
		source *iotago.Ed25519Address
		target []byte
	}{
		{
			"ok", originEdAddr, originData,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edData, err := tt.source.Serialize(serializer.DeSeriModePerformValidation)
			assert.NoError(t, err)
			assert.Equal(t, tt.target, edData)
		})
	}
}

var bech32Tests = []struct {
	name    string
	network iotago.NetworkPrefix
	addr    iotago.Address
	bech32  string
}{
	{
		"RFC example: Ed25519 mainnet",
		iotago.PrefixMainnet,
		&iotago.Ed25519Address{0x52, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
		"iota1qpf0mlq8yxpx2nck8a0slxnzr4ef2ek8f5gqxlzd0wasgp73utryj430ldu",
	},
	{
		"RFC example: Ed25519 testnet",
		iotago.PrefixTestnet,
		&iotago.Ed25519Address{0x52, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
		"atoi1qpf0mlq8yxpx2nck8a0slxnzr4ef2ek8f5gqxlzd0wasgp73utryjjl77h3",
	},
}

func TestBech32(t *testing.T) {
	for _, tt := range bech32Tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.bech32, tt.addr.Bech32(tt.network))
		})
	}
}

func TestParseBech32(t *testing.T) {
	for _, tt := range bech32Tests {
		t.Run(tt.name, func(t *testing.T) {
			network, addr, err := iotago.ParseBech32(tt.bech32)
			assert.NoError(t, err)
			assert.Equal(t, tt.network, network)
			assert.Equal(t, tt.addr, addr)
		})
	}
}
