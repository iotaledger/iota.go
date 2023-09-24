//nolint:scopelint,dupl
package iotago_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
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

type bech32Test struct {
	name       string
	network    iotago.NetworkPrefix
	addr       iotago.Address
	targetAddr iotago.Address
	bech32     string
}

var bech32Tests = []*bech32Test{
	func() *bech32Test {
		addr := &iotago.Ed25519Address{0x52, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49}
		return &bech32Test{
			name:       "RFC example: Ed25519 mainnet",
			network:    iotago.PrefixMainnet,
			addr:       addr,
			targetAddr: addr,
			bech32:     "iota1qpf0mlq8yxpx2nck8a0slxnzr4ef2ek8f5gqxlzd0wasgp73utryj430ldu",
		}
	}(),
	func() *bech32Test {
		addr := &iotago.Ed25519Address{0x52, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49}
		return &bech32Test{
			name:       "RFC example: Ed25519 testnet",
			network:    iotago.PrefixTestnet,
			addr:       addr,
			targetAddr: addr,
			bech32:     "rms1qpf0mlq8yxpx2nck8a0slxnzr4ef2ek8f5gqxlzd0wasgp73utryjkxa9q5",
		}
	}(),
	func() *bech32Test {
		addr := &iotago.MultiAddress{
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
		}

		return &bech32Test{
			name:       "Multi Address",
			network:    iotago.PrefixTestnet,
			addr:       addr,
			targetAddr: iotago.NewMultiAddressReferenceFromMultiAddress(addr),
			bech32:     "rms1yz4qe5j4s44a7qpnz4lkd0nuepc9xkchznae90gy78ht8m9g9epxweejdxn",
		}
	}(),
	func() *bech32Test {
		addr := &iotago.RestrictedAddress{
			Address:             &iotago.Ed25519Address{0x52, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
			AllowedCapabilities: iotago.AddressCapabilitiesBitMask{0x55},
		}

		return &bech32Test{
			name:       "Restricted Ed25519 Address",
			network:    iotago.PrefixTestnet,
			addr:       addr,
			targetAddr: addr,
			bech32:     "rms19qq99l0uquscye20zcl47ru6vgwh99txcax3qqmuf4amkpq8683vvjgp2588z4ru",
		}
	}(),
	func() *bech32Test {
		addr := &iotago.RestrictedAddress{
			Address:             &iotago.AccountAddress{0x52, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
			AllowedCapabilities: iotago.AddressCapabilitiesBitMask{0x55},
		}

		return &bech32Test{
			name:       "Restricted Account Address",
			network:    iotago.PrefixTestnet,
			addr:       addr,
			targetAddr: addr,
			bech32:     "rms19qy99l0uquscye20zcl47ru6vgwh99txcax3qqmuf4amkpq8683vvjgp25psyekx",
		}
	}(),
	func() *bech32Test {
		addr := &iotago.RestrictedAddress{
			Address:             &iotago.NFTAddress{0x52, 0xfd, 0xfc, 0x07, 0x21, 0x82, 0x65, 0x4f, 0x16, 0x3f, 0x5f, 0x0f, 0x9a, 0x62, 0x1d, 0x72, 0x95, 0x66, 0xc7, 0x4d, 0x10, 0x03, 0x7c, 0x4d, 0x7b, 0xbb, 0x04, 0x07, 0xd1, 0xe2, 0xc6, 0x49},
			AllowedCapabilities: iotago.AddressCapabilitiesBitMask{0x55},
		}

		return &bech32Test{
			name:       "Restricted NFT Address",
			network:    iotago.PrefixTestnet,
			addr:       addr,
			targetAddr: addr,
			bech32:     "rms19qg99l0uquscye20zcl47ru6vgwh99txcax3qqmuf4amkpq8683vvjgp25tqwdqp",
		}
	}(),
	func() *bech32Test {
		multiAddr := &iotago.MultiAddress{
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
		}

		addr := &iotago.RestrictedAddress{
			Address:             multiAddr,
			AllowedCapabilities: iotago.AddressCapabilitiesBitMask{0x55},
		}

		return &bech32Test{
			name:    "Restricted Multi Address",
			network: iotago.PrefixTestnet,
			addr:    addr,
			targetAddr: &iotago.RestrictedAddress{
				Address:             iotago.NewMultiAddressReferenceFromMultiAddress(multiAddr),
				AllowedCapabilities: addr.AllowedCapabilities,
			},
			bech32: "rms19qs25rxj2kzkhhcqxv2h7e470ny8q56mzu20hy4aqnc7avlv4qhyyecp252uua5y",
		}
	}(),
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
			assert.Equal(t, tt.targetAddr.ID(), addr.ID(), "parsed bech32 address does not match the given target address: %s != %s", tt.targetAddr.Bech32(tt.network), addr.Bech32(tt.network))
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
		iotago.RestrictedAddressWithCapabilities(edAddr, iotago.WithAddressCanReceiveAnything()),
		iotago.RestrictedAddressWithCapabilities(edAddr),

		iotago.RestrictedAddressWithCapabilities(accountAddr, iotago.WithAddressCanReceiveNativeTokens(true)),
		iotago.RestrictedAddressWithCapabilities(accountAddr, iotago.WithAddressCanReceiveMana(true)),
		iotago.RestrictedAddressWithCapabilities(accountAddr, iotago.WithAddressCanReceiveOutputsWithTimelockUnlockCondition(true)),
		iotago.RestrictedAddressWithCapabilities(accountAddr, iotago.WithAddressCanReceiveOutputsWithExpirationUnlockCondition(true)),
		iotago.RestrictedAddressWithCapabilities(accountAddr, iotago.WithAddressCanReceiveOutputsWithStorageDepositReturnUnlockCondition(true)),
		iotago.RestrictedAddressWithCapabilities(accountAddr, iotago.WithAddressCanReceiveAccountOutputs(true)),
		iotago.RestrictedAddressWithCapabilities(accountAddr, iotago.WithAddressCanReceiveNFTOutputs(true)),
		iotago.RestrictedAddressWithCapabilities(accountAddr, iotago.WithAddressCanReceiveDelegationOutputs(true)),
		iotago.RestrictedAddressWithCapabilities(accountAddr, iotago.WithAddressCanReceiveAnything()),
		iotago.RestrictedAddressWithCapabilities(accountAddr),

		iotago.RestrictedAddressWithCapabilities(nftAddr, iotago.WithAddressCanReceiveNativeTokens(true)),
		iotago.RestrictedAddressWithCapabilities(nftAddr, iotago.WithAddressCanReceiveMana(true)),
		iotago.RestrictedAddressWithCapabilities(nftAddr, iotago.WithAddressCanReceiveOutputsWithTimelockUnlockCondition(true)),
		iotago.RestrictedAddressWithCapabilities(nftAddr, iotago.WithAddressCanReceiveOutputsWithExpirationUnlockCondition(true)),
		iotago.RestrictedAddressWithCapabilities(nftAddr, iotago.WithAddressCanReceiveOutputsWithStorageDepositReturnUnlockCondition(true)),
		iotago.RestrictedAddressWithCapabilities(nftAddr, iotago.WithAddressCanReceiveAccountOutputs(true)),
		iotago.RestrictedAddressWithCapabilities(nftAddr, iotago.WithAddressCanReceiveNFTOutputs(true)),
		iotago.RestrictedAddressWithCapabilities(nftAddr, iotago.WithAddressCanReceiveDelegationOutputs(true)),
		iotago.RestrictedAddressWithCapabilities(nftAddr, iotago.WithAddressCanReceiveAnything()),
		iotago.RestrictedAddressWithCapabilities(nftAddr),

		iotago.RestrictedAddressWithCapabilities(multiAddress, iotago.WithAddressCanReceiveNativeTokens(true)),
		iotago.RestrictedAddressWithCapabilities(multiAddress, iotago.WithAddressCanReceiveMana(true)),
		iotago.RestrictedAddressWithCapabilities(multiAddress, iotago.WithAddressCanReceiveOutputsWithTimelockUnlockCondition(true)),
		iotago.RestrictedAddressWithCapabilities(multiAddress, iotago.WithAddressCanReceiveOutputsWithExpirationUnlockCondition(true)),
		iotago.RestrictedAddressWithCapabilities(multiAddress, iotago.WithAddressCanReceiveOutputsWithStorageDepositReturnUnlockCondition(true)),
		iotago.RestrictedAddressWithCapabilities(multiAddress, iotago.WithAddressCanReceiveAccountOutputs(true)),
		iotago.RestrictedAddressWithCapabilities(multiAddress, iotago.WithAddressCanReceiveNFTOutputs(true)),
		iotago.RestrictedAddressWithCapabilities(multiAddress, iotago.WithAddressCanReceiveDelegationOutputs(true)),
		iotago.RestrictedAddressWithCapabilities(multiAddress, iotago.WithAddressCanReceiveAnything()),
		iotago.RestrictedAddressWithCapabilities(multiAddress),
	}

	assertRestrictedAddresses(t, addresses)
}

type outputsSyntacticalValidationTest struct {
	// the name of the testcase
	name string
	// the amount of randomly created ed25519 addresses with private keys
	ed25519AddrCnt int
	// used to create outputs for the test
	outputsFunc func(ed25519Addresses []iotago.Address) iotago.TxEssenceOutputs
	// expected error during serialization of the transaction
	wantErr error
}

func runOutputsSyntacticalValidationTest(t *testing.T, testAPI iotago.API, test *outputsSyntacticalValidationTest) {
	t.Helper()

	t.Run(test.name, func(t *testing.T) {
		// generate random ed25519 addresses
		ed25519Addresses, _ := tpkg.RandEd25519IdentitiesSortedByAddress(test.ed25519AddrCnt)

		_, err := testAPI.Encode(test.outputsFunc(ed25519Addresses), serix.WithValidation())
		if test.wantErr != nil {
			require.ErrorIs(t, err, test.wantErr)
			return
		}
		require.NoError(t, err)
	})
}

func TestRestrictedAddressSyntacticalValidation(t *testing.T) {

	defaultAmount := OneMi

	tests := []*outputsSyntacticalValidationTest{
		// ok - Valid address types nested inside of a RestrictedAddress
		func() *outputsSyntacticalValidationTest {
			return &outputsSyntacticalValidationTest{
				name:           "ok - Valid address types nested inside of a RestrictedAddress",
				ed25519AddrCnt: 2,
				outputsFunc: func(ed25519Addresses []iotago.Address) iotago.TxEssenceOutputs {
					return iotago.TxEssenceOutputs{
						&iotago.BasicOutput{
							Amount: defaultAmount,
							Conditions: iotago.BasicOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: &iotago.RestrictedAddress{
									Address:             ed25519Addresses[0],
									AllowedCapabilities: iotago.AddressCapabilitiesBitMask{},
								}},
							},
						},
						&iotago.BasicOutput{
							Amount: defaultAmount,
							Conditions: iotago.BasicOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: &iotago.RestrictedAddress{
									Address:             &iotago.AccountAddress{},
									AllowedCapabilities: iotago.AddressCapabilitiesBitMask{},
								}},
							},
						},
						&iotago.BasicOutput{
							Amount: defaultAmount,
							Conditions: iotago.BasicOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: &iotago.RestrictedAddress{
									Address:             &iotago.NFTAddress{},
									AllowedCapabilities: iotago.AddressCapabilitiesBitMask{},
								}},
							},
						},
						&iotago.BasicOutput{
							Amount: defaultAmount,
							Conditions: iotago.BasicOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: &iotago.RestrictedAddress{
									Address: &iotago.MultiAddress{
										Addresses: []*iotago.AddressWithWeight{
											{
												Address: ed25519Addresses[1],
												Weight:  1,
											},
										},
										Threshold: 1,
									},
									AllowedCapabilities: iotago.AddressCapabilitiesBitMask{},
								}},
							},
						},
					}
				},
				wantErr: nil,
			}
		}(),

		// fail - ImplicitAccountCreationAddress nested inside of a RestrictedAddress
		func() *outputsSyntacticalValidationTest {
			return &outputsSyntacticalValidationTest{
				name:           "fail - ImplicitAccountCreationAddress nested inside of a RestrictedAddress",
				ed25519AddrCnt: 0,
				outputsFunc: func(ed25519Addresses []iotago.Address) iotago.TxEssenceOutputs {
					return iotago.TxEssenceOutputs{
						&iotago.BasicOutput{
							Amount: defaultAmount,
							Conditions: iotago.BasicOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: &iotago.RestrictedAddress{
									Address: &iotago.ImplicitAccountCreationAddress{},
								}},
							},
						},
					}
				},
				wantErr: iotago.ErrInvalidNestedAddressType,
			}
		}(),

		// fail - RestrictedAddress nested inside of a RestrictedAddress
		func() *outputsSyntacticalValidationTest {
			return &outputsSyntacticalValidationTest{
				name:           "fail - RestrictedAddress nested inside of a RestrictedAddress",
				ed25519AddrCnt: 1,
				outputsFunc: func(ed25519Addresses []iotago.Address) iotago.TxEssenceOutputs {
					return iotago.TxEssenceOutputs{
						&iotago.BasicOutput{
							Amount: defaultAmount,
							Conditions: iotago.BasicOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: &iotago.RestrictedAddress{
									Address: &iotago.RestrictedAddress{
										Address:             ed25519Addresses[0],
										AllowedCapabilities: iotago.AddressCapabilitiesBitMask{},
									},
								}},
							},
						},
					}
				},
				wantErr: iotago.ErrInvalidNestedAddressType,
			}
		}(),
	}

	testAPI := iotago.V3API(iotago.NewV3ProtocolParameters(
		iotago.WithNetworkOptions("test", "test"),
	))

	for _, tt := range tests {
		runOutputsSyntacticalValidationTest(t, testAPI, tt)
	}
}

func TestMultiAddressSyntacticalValidation(t *testing.T) {

	defaultAmount := OneMi

	tests := []*outputsSyntacticalValidationTest{
		// fail - threshold > cumulativeWeight
		func() *outputsSyntacticalValidationTest {
			return &outputsSyntacticalValidationTest{
				name:           "fail - threshold > cumulativeWeight",
				ed25519AddrCnt: 2,
				outputsFunc: func(ed25519Addresses []iotago.Address) iotago.TxEssenceOutputs {
					return iotago.TxEssenceOutputs{
						&iotago.BasicOutput{
							Amount: defaultAmount,
							Conditions: iotago.BasicOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: &iotago.MultiAddress{
									Addresses: []*iotago.AddressWithWeight{
										{
											Address: ed25519Addresses[0],
											Weight:  1,
										},
										{
											Address: ed25519Addresses[1],
											Weight:  1,
										},
									},
									Threshold: 3,
								}},
							},
						},
					}
				},
				wantErr: iotago.ErrMultiAddressThresholdInvalid,
			}
		}(),

		// fail - threshold < 1
		func() *outputsSyntacticalValidationTest {
			return &outputsSyntacticalValidationTest{
				name:           "fail - threshold < 1",
				ed25519AddrCnt: 1,
				outputsFunc: func(ed25519Addresses []iotago.Address) iotago.TxEssenceOutputs {
					return iotago.TxEssenceOutputs{
						&iotago.BasicOutput{
							Amount: defaultAmount,
							Conditions: iotago.BasicOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: &iotago.MultiAddress{
									Addresses: []*iotago.AddressWithWeight{
										{
											Address: ed25519Addresses[0],
											Weight:  1,
										},
									},
									Threshold: 0,
								}},
							},
						},
					}
				},
				wantErr: iotago.ErrMultiAddressThresholdInvalid,
			}
		}(),

		// fail - address weight == 0
		func() *outputsSyntacticalValidationTest {
			return &outputsSyntacticalValidationTest{
				name:           "fail - address weight == 0",
				ed25519AddrCnt: 2,
				outputsFunc: func(ed25519Addresses []iotago.Address) iotago.TxEssenceOutputs {
					return iotago.TxEssenceOutputs{
						&iotago.BasicOutput{
							Amount: defaultAmount,
							Conditions: iotago.BasicOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: &iotago.MultiAddress{
									Addresses: []*iotago.AddressWithWeight{
										{
											Address: ed25519Addresses[0],
											Weight:  0,
										},
										{
											Address: ed25519Addresses[1],
											Weight:  1,
										},
									},
									Threshold: 1,
								}},
							},
						},
					}
				},
				wantErr: iotago.ErrMultiAddressWeightInvalid,
			}
		}(),

		// fail - empty MultiAddress
		func() *outputsSyntacticalValidationTest {
			return &outputsSyntacticalValidationTest{
				name:           "fail - empty MultiAddress",
				ed25519AddrCnt: 2,
				outputsFunc: func(ed25519Addresses []iotago.Address) iotago.TxEssenceOutputs {
					return iotago.TxEssenceOutputs{
						&iotago.BasicOutput{
							Amount: defaultAmount,
							Conditions: iotago.BasicOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: &iotago.MultiAddress{
									Addresses: []*iotago.AddressWithWeight{},
									Threshold: 1,
								}},
							},
						},
					}
				},
				wantErr: iotago.ErrMultiAddressThresholdInvalid,
			}
		}(),

		// fail - MultiAddress limit exceeded
		func() *outputsSyntacticalValidationTest {
			return &outputsSyntacticalValidationTest{
				name:           "fail - MultiAddress limit exceeded",
				ed25519AddrCnt: 13,
				outputsFunc: func(ed25519Addresses []iotago.Address) iotago.TxEssenceOutputs {
					return iotago.TxEssenceOutputs{
						&iotago.BasicOutput{
							Amount: defaultAmount,
							Conditions: iotago.BasicOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: &iotago.MultiAddress{
									Addresses: []*iotago.AddressWithWeight{
										{Address: ed25519Addresses[2], Weight: 1},
										{Address: ed25519Addresses[3], Weight: 1},
										{Address: ed25519Addresses[4], Weight: 1},
										{Address: ed25519Addresses[5], Weight: 1},
										{Address: ed25519Addresses[6], Weight: 1},
										{Address: ed25519Addresses[7], Weight: 1},
										{Address: ed25519Addresses[8], Weight: 1},
										{Address: ed25519Addresses[9], Weight: 1},
										{Address: ed25519Addresses[10], Weight: 1},
										{Address: ed25519Addresses[11], Weight: 1},
										{Address: ed25519Addresses[12], Weight: 1},
									},
									Threshold: 11,
								}},
							},
						},
					}
				},
				wantErr: serializer.ErrArrayValidationMaxElementsExceeded,
			}
		}(),

		// fail - raw address part of all addresses inside MultiAddress need to be unique
		func() *outputsSyntacticalValidationTest {
			return &outputsSyntacticalValidationTest{
				name:           "fail - raw address part of all addresses inside MultiAddress need to be unique",
				ed25519AddrCnt: 1,
				outputsFunc: func(ed25519Addresses []iotago.Address) iotago.TxEssenceOutputs {
					return iotago.TxEssenceOutputs{
						&iotago.BasicOutput{
							Amount: defaultAmount,
							Conditions: iotago.BasicOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: &iotago.MultiAddress{
									Addresses: []*iotago.AddressWithWeight{
										// both have the same pubKeyHash
										{
											Address: &iotago.Ed25519Address{},
											Weight:  1,
										},
										{
											Address: &iotago.Ed25519Address{},
											Weight:  1,
										},
									},
									Threshold: 1,
								}},
							},
						},
					}
				},
				wantErr: serializer.ErrArrayValidationViolatesUniqueness,
			}
		}(),

		// fail - ImplicitAccountCreationAddress nested inside of a MultiAddress
		func() *outputsSyntacticalValidationTest {
			return &outputsSyntacticalValidationTest{
				name:           "fail - ImplicitAccountCreationAddress nested inside of a MultiAddress",
				ed25519AddrCnt: 1,
				outputsFunc: func(ed25519Addresses []iotago.Address) iotago.TxEssenceOutputs {
					return iotago.TxEssenceOutputs{
						&iotago.BasicOutput{
							Amount: defaultAmount,
							Conditions: iotago.BasicOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: &iotago.RestrictedAddress{
									Address: &iotago.MultiAddress{
										Addresses: []*iotago.AddressWithWeight{
											{
												Address: &iotago.ImplicitAccountCreationAddress{},
												Weight:  1,
											},
										},
										Threshold: 1,
									},
								}},
							},
						},
					}
				},
				wantErr: iotago.ErrInvalidNestedAddressType,
			}
		}(),

		// fail - MultiAddress nested inside of a MultiAddress
		func() *outputsSyntacticalValidationTest {
			return &outputsSyntacticalValidationTest{
				name:           "fail - MultiAddress nested inside of a MultiAddress",
				ed25519AddrCnt: 2,
				outputsFunc: func(ed25519Addresses []iotago.Address) iotago.TxEssenceOutputs {
					return iotago.TxEssenceOutputs{
						&iotago.BasicOutput{
							Amount: defaultAmount,
							Conditions: iotago.BasicOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: &iotago.MultiAddress{
									Addresses: []*iotago.AddressWithWeight{
										{
											Address: &iotago.MultiAddress{
												Addresses: iotago.AddressesWithWeight{
													{
														Address: ed25519Addresses[1],
														Weight:  1,
													},
												},
												Threshold: 1,
											},
											Weight: 1,
										},
									},
									Threshold: 1,
								}},
							},
						},
					}
				},
				wantErr: iotago.ErrInvalidNestedAddressType,
			}
		}(),

		// fail - RestrictedAddress nested inside of a MultiAddress
		func() *outputsSyntacticalValidationTest {
			return &outputsSyntacticalValidationTest{
				name:           "fail - RestrictedAddress nested inside of a MultiAddress",
				ed25519AddrCnt: 1,
				outputsFunc: func(ed25519Addresses []iotago.Address) iotago.TxEssenceOutputs {
					return iotago.TxEssenceOutputs{
						&iotago.BasicOutput{
							Amount: defaultAmount,
							Conditions: iotago.BasicOutputUnlockConditions{
								&iotago.AddressUnlockCondition{Address: &iotago.MultiAddress{
									Addresses: []*iotago.AddressWithWeight{
										{
											Address: &iotago.RestrictedAddress{
												Address:             ed25519Addresses[0],
												AllowedCapabilities: iotago.AddressCapabilitiesBitMask{},
											},
											Weight: 1,
										},
									},
									Threshold: 1,
								}},
							},
						},
					}
				},
				wantErr: iotago.ErrInvalidNestedAddressType,
			}
		}(),
	}

	testAPI := iotago.V3API(iotago.NewV3ProtocolParameters(
		iotago.WithNetworkOptions("test", "test"),
	))

	for _, tt := range tests {
		runOutputsSyntacticalValidationTest(t, testAPI, tt)
	}
}
