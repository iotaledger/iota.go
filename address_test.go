package iotago_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/hive.go/serializer"
	"github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/stretchr/testify/assert"
)

func TestAddressDeSerialization(t *testing.T) {
	tests := []struct {
		name       string
		sourceData []byte
		target     serializer.Serializable
		checkBytes func(target serializer.Serializable) []byte
		err        error
	}{
		{
			"ok - Ed25519Address",
			func() []byte {
				_, data := tpkg.RandEd25519AddressAndBytes()
				return data
			}(),
			&iotago.Ed25519Address{},
			func(target serializer.Serializable) []byte {
				return target.(*iotago.Ed25519Address)[:]
			},
			nil,
		},
		{
			"ok - BLSAddress",
			func() []byte {
				_, data := tpkg.RandBLSAddressAndBytes()
				return data
			}(),
			&iotago.BLSAddress{},
			func(target serializer.Serializable) []byte {
				return target.(*iotago.BLSAddress)[:]
			},
			nil,
		},
		{
			"ok - AliasAddress",
			func() []byte {
				_, data := tpkg.RandAliasAddressAndBytes()
				return data
			}(),
			&iotago.AliasAddress{},
			func(target serializer.Serializable) []byte {
				return target.(*iotago.AliasAddress)[:]
			},
			nil,
		},
		{
			"ok - NFTAddress",
			func() []byte {
				_, data := tpkg.RandNFTAddressAndBytes()
				return data
			}(),
			&iotago.NFTAddress{},
			func(target serializer.Serializable) []byte {
				return target.(*iotago.NFTAddress)[:]
			},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytesRead, err := tt.target.Deserialize(tt.sourceData, serializer.DeSeriModePerformValidation)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, len(tt.sourceData), bytesRead)
			assert.Equal(t, tt.sourceData[serializer.SmallTypeDenotationByteSize:], tt.checkBytes(tt.target))

			outputData, err := tt.target.Serialize(serializer.DeSeriModePerformValidation)
			assert.NoError(t, err)
			assert.Equal(t, tt.sourceData, outputData)
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
