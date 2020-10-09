package iota_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/iota.go"
	"github.com/stretchr/testify/assert"
)

func TestWOTSAddress_Deserialize(t *testing.T) {
	tests := []struct {
		name         string
		wotsAddrData []byte
		err          error
	}{
		{
			"ok",
			func() []byte {
				_, wotsAddrData := randWOTSAddr()
				return wotsAddrData
			}(),
			nil,
		},
		{
			"not enough bytes",
			func() []byte {
				_, wotsAddrData := randWOTSAddr()
				return wotsAddrData[:iota.WOTSAddressSerializedBytesSize-1]
			}(),
			iota.ErrDeserializationNotEnoughData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wotsAddr := &iota.WOTSAddress{}
			bytesRead, err := wotsAddr.Deserialize(tt.wotsAddrData, iota.DeSeriModePerformValidation)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, len(tt.wotsAddrData), bytesRead)
			assert.Equal(t, tt.wotsAddrData[iota.SmallTypeDenotationByteSize:], wotsAddr[:])
		})
	}
}

func TestWOTSAddress_Serialize(t *testing.T) {
	originWOTSAddr, originData := randWOTSAddr()
	tests := []struct {
		name   string
		source *iota.WOTSAddress
		target []byte
	}{
		{
			"ok", originWOTSAddr, originData,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wotsData, err := tt.source.Serialize(iota.DeSeriModePerformValidation)
			assert.NoError(t, err)
			assert.Equal(t, tt.target, wotsData)
		})
	}
}

func TestEd25519Address_Deserialize(t *testing.T) {
	tests := []struct {
		name       string
		edAddrData []byte
		err        error
	}{
		{
			"ok",
			func() []byte {
				_, edAddrData := randEd25519Addr()
				return edAddrData
			}(),
			nil,
		},
		{
			"not enough bytes",
			func() []byte {
				_, edAddrData := randEd25519Addr()
				return edAddrData[:iota.Ed25519AddressSerializedBytesSize-1]
			}(),
			iota.ErrDeserializationNotEnoughData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edAddr := &iota.Ed25519Address{}
			bytesRead, err := edAddr.Deserialize(tt.edAddrData, iota.DeSeriModePerformValidation)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, len(tt.edAddrData), bytesRead)
			assert.Equal(t, tt.edAddrData[iota.SmallTypeDenotationByteSize:], edAddr[:])
		})
	}
}

func TestEd25519Address_Serialize(t *testing.T) {
	originEdAddr, originData := randEd25519Addr()
	tests := []struct {
		name   string
		source *iota.Ed25519Address
		target []byte
	}{
		{
			"ok", originEdAddr, originData,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edData, err := tt.source.Serialize(iota.DeSeriModePerformValidation)
			assert.NoError(t, err)
			assert.Equal(t, tt.target, edData)
		})
	}
}

var bech32Tests = []struct {
	name    string
	network iota.NetworkPrefix
	addr    iota.Address
	bech32  string
}{
	{
		"RFC example: W-OTS mainnet",
		iota.PrefixMainnet,
		iota.WOTSAddressFromTrytes("EQSAUZXULTTYZCLNJNTXQTQHOMOFZERHTCGTXOLTVAHKSA9OGAZDEKECURBRIXIJWNPFCQIOVFVVXJVD9"),
		"iot1qr4r3j4wamzu8ltdp6y7xtysj77vqtwua3q23hx9zmrmcqfdpd4muv25jwctsmxlh5g60w8s6k4x7gsq28c8da",
	},
	{
		"RFC example: Ed25519 mainnet",
		iota.PrefixMainnet,
		&iota.Ed25519Address{0x52, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
		"iot1q9f0mlq8yxpx2nck8a0slxnzr4ef2ek8f5gqxlzd0wasgp73utryjtzcp98",
	},
	{
		"RFC example: W-OTS testnet",
		iota.PrefixTestnet,
		iota.WOTSAddressFromTrytes("EQSAUZXULTTYZCLNJNTXQTQHOMOFZERHTCGTXOLTVAHKSA9OGAZDEKECURBRIXIJWNPFCQIOVFVVXJVD9"),
		"tio1qr4r3j4wamzu8ltdp6y7xtysj77vqtwua3q23hx9zmrmcqfdpd4muv25jwctsmxlh5g60w8s6k4x7gsqcmnkzh",
	},
	{
		"RFC example: Ed25519 testnet",
		iota.PrefixTestnet,
		&iota.Ed25519Address{0x52, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
		"tio1q9f0mlq8yxpx2nck8a0slxnzr4ef2ek8f5gqxlzd0wasgp73utryj3qemv4",
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
			network, addr, err := iota.ParseBech32(tt.bech32)
			assert.NoError(t, err)
			assert.Equal(t, tt.network, network)
			assert.Equal(t, tt.addr, addr)
		})
	}
}
