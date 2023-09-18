//nolint:scopelint,dupl
package iotago_test

import (
	"testing"

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
			name:   "ok - AccountAddress",
			source: tpkg.RandAccountAddress(),
			target: &iotago.AccountAddress{},
		},
		{
			name:   "ok - NFTAddress",
			source: tpkg.RandNFTAddress(),
			target: &iotago.NFTAddress{},
		},
		{
			name:   "ok - ImplicitAccountCreationAddress",
			source: tpkg.RandImplicitAccountCreationAddress(),
			target: &iotago.ImplicitAccountCreationAddress{},
		},
		{
			name:   "ok - MultiAddress",
			source: tpkg.RandMultiAddress(),
			target: &iotago.MultiAddress{},
		},
		{
			name:   "ok - RestrictedEd25519Address without capabilities",
			source: tpkg.RandRestrictedEd25519Address(iotago.AddressCapabilitiesBitMask{0x0}),
			target: &iotago.RestrictedAddress{},
		},
		{
			name:   "ok - RestrictedEd25519Address with capabilities",
			source: tpkg.RandRestrictedEd25519Address(iotago.AddressCapabilitiesBitMask{0xff}),
			target: &iotago.RestrictedAddress{},
		},
		{
			name:   "ok - RestrictedAccountAddress with capabilities",
			source: tpkg.RandRestrictedAccountAddress(iotago.AddressCapabilitiesBitMask{0xff}),
			target: &iotago.RestrictedAddress{},
		},
		{
			name:   "ok - RestrictedNFTAddress with capabilities",
			source: tpkg.RandRestrictedNFTAddress(iotago.AddressCapabilitiesBitMask{0xff}),
			target: &iotago.RestrictedAddress{},
		},
		{
			name:   "ok - RestrictedMultiAddress with capabilities",
			source: tpkg.RandRestrictedMultiAddress(iotago.AddressCapabilitiesBitMask{0xff}),
			target: &iotago.RestrictedAddress{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

var bech32Tests = []struct {
	name     string
	network  iotago.NetworkPrefix
	addr     iotago.Address
	bech32   string
	parseErr error
}{
	{
		"RFC example: Ed25519 mainnet",
		iotago.PrefixMainnet,
		&iotago.Ed25519Address{0x52, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
		"iota1qpf0mlq8yxpx2nck8a0slxnzr4ef2ek8f5gqxlzd0wasgp73utryj430ldu",
		nil,
	},
	{
		"RFC example: Ed25519 testnet",
		iotago.PrefixDevnet,
		&iotago.Ed25519Address{0x52, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
		"atoi1qpf0mlq8yxpx2nck8a0slxnzr4ef2ek8f5gqxlzd0wasgp73utryjjl77h3",
		nil,
	},
	{
		"Multi Address",
		iotago.PrefixDevnet,
		&iotago.MultiAddress{
			Addresses: []*iotago.AddressWithWeight{
				{
					Address: &iotago.Ed25519Address{0x52, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
					Weight:  1,
				},
				{
					Address: &iotago.Ed25519Address{0x53, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
					Weight:  1,
				},
				{
					Address: &iotago.Ed25519Address{0x54, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
					Weight:  1,
				},
				{
					Address: &iotago.AccountAddress{0x55, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
					Weight:  2,
				},
				{
					Address: &iotago.NFTAddress{0x56, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
					Weight:  3,
				},
			},
			Threshold: 2,
		},
		"atoi1yz4qe5j4s44a7qpnz4lkd0nuepc9xkchznae90gy78ht8m9g9epxwaq3k3k",
		iotago.ErrMultiAddrCannotBeReconstructedViaBech32,
	},
	{
		"Restricted Ed25519 Address",
		iotago.PrefixDevnet,
		&iotago.RestrictedAddress{
			Address:             &iotago.Ed25519Address{0x52, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
			AllowedCapabilities: iotago.AddressCapabilitiesBitMask{0x55},
		},
		"atoi19qq99l0uquscye20zcl47ru6vgwh99txcax3qqmuf4amkpq8683vvjgp252sfghk",
		nil,
	},
	{
		"Restricted Account Address",
		iotago.PrefixDevnet,
		&iotago.RestrictedAddress{
			Address:             &iotago.AccountAddress{0x52, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
			AllowedCapabilities: iotago.AddressCapabilitiesBitMask{0x55},
		},
		"atoi19qy99l0uquscye20zcl47ru6vgwh99txcax3qqmuf4amkpq8683vvjgp25v80yzv",
		nil,
	},
	{
		"Restricted NFT Address",
		iotago.PrefixDevnet,
		&iotago.RestrictedAddress{
			Address:             &iotago.NFTAddress{0x52, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
			AllowedCapabilities: iotago.AddressCapabilitiesBitMask{0x55},
		},
		"atoi19qg99l0uquscye20zcl47ru6vgwh99txcax3qqmuf4amkpq8683vvjgp25xh9s5t",
		nil,
	},
	{
		"Restricted Multi Address",
		iotago.PrefixDevnet,
		&iotago.RestrictedAddress{
			Address: &iotago.MultiAddress{
				Addresses: []*iotago.AddressWithWeight{
					{
						Address: &iotago.Ed25519Address{0x52, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
						Weight:  1,
					},
					{
						Address: &iotago.Ed25519Address{0x53, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
						Weight:  1,
					},
					{
						Address: &iotago.Ed25519Address{0x54, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
						Weight:  1,
					},
					{
						Address: &iotago.AccountAddress{0x55, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
						Weight:  2,
					},
					{
						Address: &iotago.NFTAddress{0x56, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
						Weight:  3,
					},
				},
				Threshold: 2,
			},
			AllowedCapabilities: iotago.AddressCapabilitiesBitMask{0x55},
		},
		"atoi19qs25rxj2kzkhhcqxv2h7e470ny8q56mzu20hy4aqnc7avlv4qhyyecp258thqqw",
		iotago.ErrMultiAddrCannotBeReconstructedViaBech32,
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
			if tt.parseErr != nil {
				assert.ErrorIs(t, err, tt.parseErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.network, network)
			assert.Equal(t, tt.addr, addr)
		})
	}
}

func TestImplicitAccountCreationAddressCapabilities(t *testing.T) {
	address := iotago.ImplicitAccountCreationAddressFromPubKey(ed25519.PublicKey(tpkg.Rand32ByteArray()).ToEd25519())
	require.True(t, address.CannotReceiveNativeTokens())
	require.False(t, address.CannotReceiveMana())
	require.True(t, address.CannotReceiveOutputsWithTimelockUnlockCondition())
	require.True(t, address.CannotReceiveOutputsWithExpirationUnlockCondition())
	require.True(t, address.CannotReceiveOutputsWithStorageDepositReturnUnlockCondition())
	require.True(t, address.CannotReceiveAccountOutputs())
	require.True(t, address.CannotReceiveNFTOutputs())
	require.True(t, address.CannotReceiveDelegationOutputs())
}

func assertRestrictedAddresses(t *testing.T, addresses []*iotago.RestrictedAddress) {
	t.Helper()

	for i, addr := range addresses {
		// fmt.Println(addr.Bech32(iotago.PrefixMainnet))

		j, err := tpkg.TestAPI.JSONEncode(addr)
		_ = j
		require.NoError(t, err)
		// fmt.Println(string(j))

		b, err := tpkg.TestAPI.Encode(addr)
		require.NoError(t, err)
		// fmt.Println(hexutil.Encode(b))

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

		indexModuloTen := (i % 10)

		switch indexModuloTen {
		default:
			for checkIndex, check := range addrChecks {
				require.Equalf(t, check(), indexModuloTen != checkIndex, "index: %d", i)
			}
			require.Equalf(t, addr.AllowedCapabilitiesBitMask(), iotago.AddressCapabilitiesBitMask{0 | 1<<indexModuloTen}, "index: %d", i)
			require.Equalf(t, addr.AllowedCapabilitiesBitMask().Size(), 2, "index: %d", i)
		case 8:
			for _, check := range addrChecks {
				require.Falsef(t, check(), "index: %d", i)
			}
			require.Equalf(t, addr.AllowedCapabilitiesBitMask(), iotago.AddressCapabilitiesBitMask{0xFF}, "index: %d", i)
			require.Equalf(t, addr.AllowedCapabilitiesBitMask().Size(), 2, "index: %d", i)
		case 9:
			for _, check := range addrChecks {
				require.Truef(t, check(), "index: %d", i)
			}
			require.Equalf(t, addr.AllowedCapabilitiesBitMask(), iotago.AddressCapabilitiesBitMask{}, "index: %d", i)
			require.Equalf(t, addr.AllowedCapabilitiesBitMask().Size(), 1, "index: %d", i)
		}
	}
}

//nolint:dupl // we have a lot of similar tests
func TestRestrictedAddressCapabilities(t *testing.T) {
	edAddr := tpkg.RandEd25519Address()
	accountAddr := tpkg.RandAccountAddress()
	nftAddr := tpkg.RandNFTAddress()
	multiAddress := tpkg.RandMultiAddress()

	addresses := []*iotago.RestrictedAddress{
		iotago.RestrictedAddressWithCapabilities(edAddr, iotago.WithAddressCanReceiveNativeTokens(true)),
		iotago.RestrictedAddressWithCapabilities(edAddr, iotago.WithAddressCanReceiveMana(true)),
		iotago.RestrictedAddressWithCapabilities(edAddr, iotago.WithAddressCanReceiveOutputsWithTimelockUnlockCondition(true)),
		iotago.RestrictedAddressWithCapabilities(edAddr, iotago.WithAddressCanReceiveOutputsWithExpirationUnlockCondition(true)),
		iotago.RestrictedAddressWithCapabilities(edAddr, iotago.WithAddressCanReceiveOutputsWithStorageDepositReturnUnlockCondition(true)),
		iotago.RestrictedAddressWithCapabilities(edAddr, iotago.WithAddressCanReceiveAccountOutputs(true)),
		iotago.RestrictedAddressWithCapabilities(edAddr, iotago.WithAddressCanReceiveNFTOutputs(true)),
		iotago.RestrictedAddressWithCapabilities(edAddr, iotago.WithAddressCanReceiveDelegationOutputs(true)),
		iotago.RestrictedAddressWithCapabilities(edAddr, iotago.WithAddressHasNoLimitations()),
		iotago.RestrictedAddressWithCapabilities(edAddr),

		iotago.RestrictedAddressWithCapabilities(accountAddr, iotago.WithAddressCanReceiveNativeTokens(true)),
		iotago.RestrictedAddressWithCapabilities(accountAddr, iotago.WithAddressCanReceiveMana(true)),
		iotago.RestrictedAddressWithCapabilities(accountAddr, iotago.WithAddressCanReceiveOutputsWithTimelockUnlockCondition(true)),
		iotago.RestrictedAddressWithCapabilities(accountAddr, iotago.WithAddressCanReceiveOutputsWithExpirationUnlockCondition(true)),
		iotago.RestrictedAddressWithCapabilities(accountAddr, iotago.WithAddressCanReceiveOutputsWithStorageDepositReturnUnlockCondition(true)),
		iotago.RestrictedAddressWithCapabilities(accountAddr, iotago.WithAddressCanReceiveAccountOutputs(true)),
		iotago.RestrictedAddressWithCapabilities(accountAddr, iotago.WithAddressCanReceiveNFTOutputs(true)),
		iotago.RestrictedAddressWithCapabilities(accountAddr, iotago.WithAddressCanReceiveDelegationOutputs(true)),
		iotago.RestrictedAddressWithCapabilities(accountAddr, iotago.WithAddressHasNoLimitations()),
		iotago.RestrictedAddressWithCapabilities(accountAddr),

		iotago.RestrictedAddressWithCapabilities(nftAddr, iotago.WithAddressCanReceiveNativeTokens(true)),
		iotago.RestrictedAddressWithCapabilities(nftAddr, iotago.WithAddressCanReceiveMana(true)),
		iotago.RestrictedAddressWithCapabilities(nftAddr, iotago.WithAddressCanReceiveOutputsWithTimelockUnlockCondition(true)),
		iotago.RestrictedAddressWithCapabilities(nftAddr, iotago.WithAddressCanReceiveOutputsWithExpirationUnlockCondition(true)),
		iotago.RestrictedAddressWithCapabilities(nftAddr, iotago.WithAddressCanReceiveOutputsWithStorageDepositReturnUnlockCondition(true)),
		iotago.RestrictedAddressWithCapabilities(nftAddr, iotago.WithAddressCanReceiveAccountOutputs(true)),
		iotago.RestrictedAddressWithCapabilities(nftAddr, iotago.WithAddressCanReceiveNFTOutputs(true)),
		iotago.RestrictedAddressWithCapabilities(nftAddr, iotago.WithAddressCanReceiveDelegationOutputs(true)),
		iotago.RestrictedAddressWithCapabilities(nftAddr, iotago.WithAddressHasNoLimitations()),
		iotago.RestrictedAddressWithCapabilities(nftAddr),

		iotago.RestrictedAddressWithCapabilities(multiAddress, iotago.WithAddressCanReceiveNativeTokens(true)),
		iotago.RestrictedAddressWithCapabilities(multiAddress, iotago.WithAddressCanReceiveMana(true)),
		iotago.RestrictedAddressWithCapabilities(multiAddress, iotago.WithAddressCanReceiveOutputsWithTimelockUnlockCondition(true)),
		iotago.RestrictedAddressWithCapabilities(multiAddress, iotago.WithAddressCanReceiveOutputsWithExpirationUnlockCondition(true)),
		iotago.RestrictedAddressWithCapabilities(multiAddress, iotago.WithAddressCanReceiveOutputsWithStorageDepositReturnUnlockCondition(true)),
		iotago.RestrictedAddressWithCapabilities(multiAddress, iotago.WithAddressCanReceiveAccountOutputs(true)),
		iotago.RestrictedAddressWithCapabilities(multiAddress, iotago.WithAddressCanReceiveNFTOutputs(true)),
		iotago.RestrictedAddressWithCapabilities(multiAddress, iotago.WithAddressCanReceiveDelegationOutputs(true)),
		iotago.RestrictedAddressWithCapabilities(multiAddress, iotago.WithAddressHasNoLimitations()),
		iotago.RestrictedAddressWithCapabilities(multiAddress),
	}

	assertRestrictedAddresses(t, addresses)
}
