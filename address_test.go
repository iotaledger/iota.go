//nolint:scopelint
package iotago_test

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestAddressDeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok - Ed25519Address",
			source: tpkg.RandEd25519Address(),
			target: &iotago.Ed25519Address{},
		},
		{
			name:   "ok - RestrictedEd25519Address with capabilities",
			source: tpkg.RandRestrictedEd25519Address(iotago.AddressCapabilitiesBitMask{0xff}),
			target: &iotago.RestrictedEd25519Address{},
		},
		{
			name:   "ok - RestrictedEd25519Address without capabilities",
			source: tpkg.RandRestrictedEd25519Address(iotago.AddressCapabilitiesBitMask{0x0}),
			target: &iotago.RestrictedEd25519Address{},
		},
		{
			name:   "ok - AccountAddress",
			source: tpkg.RandAccountAddress(),
			target: &iotago.AccountAddress{},
		},
		{
			name:   "ok - RestrictedAccountAddress with capabilities",
			source: tpkg.RandRestrictedAccountAddress(iotago.AddressCapabilitiesBitMask{0xff}),
			target: &iotago.RestrictedAccountAddress{},
		},
		{
			name:   "ok - RestrictedAccountAddress without capabilities",
			source: tpkg.RandRestrictedAccountAddress(iotago.AddressCapabilitiesBitMask{0x0}),
			target: &iotago.RestrictedAccountAddress{},
		},
		{
			name:   "ok - NFTAddress",
			source: tpkg.RandNFTAddress(),
			target: &iotago.NFTAddress{},
		},
		{
			name:   "ok - RestrictedNFTAddress with capabilities",
			source: tpkg.RandRestrictedNFTAddress(iotago.AddressCapabilitiesBitMask{0xff}),
			target: &iotago.RestrictedNFTAddress{},
		},
		{
			name:   "ok - RestrictedNFTAddress without capabilities",
			source: tpkg.RandRestrictedNFTAddress(iotago.AddressCapabilitiesBitMask{0x0}),
			target: &iotago.RestrictedNFTAddress{},
		},
		{
			name:   "ok - ImplicitAccountCreationAddress",
			source: tpkg.RandImplicitAccountCreationAddress(),
			target: &iotago.ImplicitAccountCreationAddress{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
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
		iotago.PrefixDevnet,
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

func TestNonRestrictedAddressCapabilities(t *testing.T) {
	pubKey := ed25519.PublicKey(tpkg.Rand32ByteArray()).ToEd25519()
	outputID := tpkg.RandOutputID(1)

	addresses := []iotago.Address{
		iotago.Ed25519AddressFromPubKey(pubKey),
		iotago.AccountAddressFromOutputID(outputID),
		iotago.NFTAddressFromOutputID(outputID),
		iotago.ImplicitAccountCreationAddressFromPubKey(pubKey),
	}

	for _, addr := range addresses {
		switch addr.Type() {
		case iotago.AddressEd25519:
			require.False(t, addr.CannotReceiveNativeTokens())
			require.False(t, addr.CannotReceiveMana())
			require.False(t, addr.CannotReceiveOutputsWithTimelockUnlockCondition())
			require.False(t, addr.CannotReceiveOutputsWithExpirationUnlockCondition())
			require.False(t, addr.CannotReceiveOutputsWithStorageDepositReturnUnlockCondition())
			require.False(t, addr.CannotReceiveAccountOutputs())
			require.False(t, addr.CannotReceiveNFTOutputs())
			require.False(t, addr.CannotReceiveDelegationOutputs())
		case iotago.AddressAccount:
			require.False(t, addr.CannotReceiveNativeTokens())
			require.False(t, addr.CannotReceiveMana())
			require.False(t, addr.CannotReceiveOutputsWithTimelockUnlockCondition())
			require.False(t, addr.CannotReceiveOutputsWithExpirationUnlockCondition())
			require.False(t, addr.CannotReceiveOutputsWithStorageDepositReturnUnlockCondition())
			require.False(t, addr.CannotReceiveAccountOutputs())
			require.False(t, addr.CannotReceiveNFTOutputs())
			require.False(t, addr.CannotReceiveDelegationOutputs())
		case iotago.AddressNFT:
			require.False(t, addr.CannotReceiveNativeTokens())
			require.False(t, addr.CannotReceiveMana())
			require.False(t, addr.CannotReceiveOutputsWithTimelockUnlockCondition())
			require.False(t, addr.CannotReceiveOutputsWithExpirationUnlockCondition())
			require.False(t, addr.CannotReceiveOutputsWithStorageDepositReturnUnlockCondition())
			require.False(t, addr.CannotReceiveAccountOutputs())
			require.False(t, addr.CannotReceiveNFTOutputs())
			require.False(t, addr.CannotReceiveDelegationOutputs())
		case iotago.AddressImplicitAccountCreation:
			require.True(t, addr.CannotReceiveNativeTokens())
			require.False(t, addr.CannotReceiveMana())
			require.True(t, addr.CannotReceiveOutputsWithTimelockUnlockCondition())
			require.True(t, addr.CannotReceiveOutputsWithExpirationUnlockCondition())
			require.True(t, addr.CannotReceiveOutputsWithStorageDepositReturnUnlockCondition())
			require.True(t, addr.CannotReceiveAccountOutputs())
			require.True(t, addr.CannotReceiveNFTOutputs())
			require.True(t, addr.CannotReceiveDelegationOutputs())
		default:
			t.Fail()
		}
	}

}

func assertRestrictedAddresses(t *testing.T, addresses []iotago.RestrictedAddress) {
	for i, addr := range addresses {
		fmt.Println(addr.Bech32(iotago.PrefixMainnet))

		j, err := tpkg.TestAPI.JSONEncode(addr)
		require.NoError(t, err)
		fmt.Println(string(j))

		b, err := tpkg.TestAPI.Encode(addr)
		require.NoError(t, err)
		fmt.Println(hexutil.Encode(b))

		addrChecks := []func() bool{
			addr.CannotReceiveNativeTokens,
			addr.CannotReceiveMana,
			addr.CannotReceiveOutputsWithTimelockUnlockCondition,
			addr.CannotReceiveOutputsWithExpirationUnlockCondition,
			addr.CannotReceiveOutputsWithStorageDepositReturnUnlockCondition,
			addr.CannotReceiveAccountOutputs,
			addr.CannotReceiveNFTOutputs,
			addr.CannotReceiveDelegationOutputs,
		}

		require.Equal(t, addr.Size(), len(b))

		switch i {
		default:
			for checkIndex, check := range addrChecks {
				require.Equal(t, check(), i != checkIndex)
			}
			require.Equal(t, addr.CapabilitiesBitMask(), iotago.AddressCapabilitiesBitMask{0 | 1<<i})
			require.Equal(t, addr.Size(), 35)
		case 8:
			for _, check := range addrChecks {
				require.False(t, check())
			}
			require.Equal(t, addr.CapabilitiesBitMask(), iotago.AddressCapabilitiesBitMask{0xFF})
			require.Equal(t, addr.Size(), 35)
		case 9:
			for _, check := range addrChecks {
				require.True(t, check())
			}
			require.Equal(t, addr.CapabilitiesBitMask(), iotago.AddressCapabilitiesBitMask(nil))
			require.Equal(t, addr.Size(), 34)
		}
	}
}

func TestRestrictedEd25519AddressCapabilities(t *testing.T) {
	pubKey := ed25519.PublicKey(tpkg.Rand32ByteArray()).ToEd25519()
	addresses := []iotago.RestrictedAddress{
		iotago.RestrictedEd25519AddressFromPubKey(pubKey, true, false, false, false, false, false, false, false),
		iotago.RestrictedEd25519AddressFromPubKey(pubKey, false, true, false, false, false, false, false, false),
		iotago.RestrictedEd25519AddressFromPubKey(pubKey, false, false, true, false, false, false, false, false),
		iotago.RestrictedEd25519AddressFromPubKey(pubKey, false, false, false, true, false, false, false, false),
		iotago.RestrictedEd25519AddressFromPubKey(pubKey, false, false, false, false, true, false, false, false),
		iotago.RestrictedEd25519AddressFromPubKey(pubKey, false, false, false, false, false, true, false, false),
		iotago.RestrictedEd25519AddressFromPubKey(pubKey, false, false, false, false, false, false, true, false),
		iotago.RestrictedEd25519AddressFromPubKey(pubKey, false, false, false, false, false, false, false, true),
		iotago.RestrictedEd25519AddressFromPubKey(pubKey, true, true, true, true, true, true, true, true),
		iotago.RestrictedEd25519AddressFromPubKey(pubKey, false, false, false, false, false, false, false, false),
	}

	assertRestrictedAddresses(t, addresses)
}

func TestRestrictedNFTAddressCapabilities(t *testing.T) {
	outputID := tpkg.RandOutputID(1)
	addresses := []iotago.RestrictedAddress{
		iotago.RestrictedNFTAddressFromOutputID(outputID, true, false, false, false, false, false, false, false),
		iotago.RestrictedNFTAddressFromOutputID(outputID, false, true, false, false, false, false, false, false),
		iotago.RestrictedNFTAddressFromOutputID(outputID, false, false, true, false, false, false, false, false),
		iotago.RestrictedNFTAddressFromOutputID(outputID, false, false, false, true, false, false, false, false),
		iotago.RestrictedNFTAddressFromOutputID(outputID, false, false, false, false, true, false, false, false),
		iotago.RestrictedNFTAddressFromOutputID(outputID, false, false, false, false, false, true, false, false),
		iotago.RestrictedNFTAddressFromOutputID(outputID, false, false, false, false, false, false, true, false),
		iotago.RestrictedNFTAddressFromOutputID(outputID, false, false, false, false, false, false, false, true),
		iotago.RestrictedNFTAddressFromOutputID(outputID, true, true, true, true, true, true, true, true),
		iotago.RestrictedNFTAddressFromOutputID(outputID, false, false, false, false, false, false, false, false),
	}

	assertRestrictedAddresses(t, addresses)
}

func TestRestrictedAccountAddressCapabilities(t *testing.T) {
	outputID := tpkg.RandOutputID(1)
	addresses := []iotago.RestrictedAddress{
		iotago.RestrictedAccountAddressFromOutputID(outputID, true, false, false, false, false, false, false, false),
		iotago.RestrictedAccountAddressFromOutputID(outputID, false, true, false, false, false, false, false, false),
		iotago.RestrictedAccountAddressFromOutputID(outputID, false, false, true, false, false, false, false, false),
		iotago.RestrictedAccountAddressFromOutputID(outputID, false, false, false, true, false, false, false, false),
		iotago.RestrictedAccountAddressFromOutputID(outputID, false, false, false, false, true, false, false, false),
		iotago.RestrictedAccountAddressFromOutputID(outputID, false, false, false, false, false, true, false, false),
		iotago.RestrictedAccountAddressFromOutputID(outputID, false, false, false, false, false, false, true, false),
		iotago.RestrictedAccountAddressFromOutputID(outputID, false, false, false, false, false, false, false, true),
		iotago.RestrictedAccountAddressFromOutputID(outputID, true, true, true, true, true, true, true, true),
		iotago.RestrictedAccountAddressFromOutputID(outputID, false, false, false, false, false, false, false, false),
	}

	assertRestrictedAddresses(t, addresses)
}
