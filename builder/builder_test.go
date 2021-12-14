package builder_test

/*
func TestTxBuilder(t *testing.T) {
	inputIDs := tpkg.RandOutputIDs(4)
	_, ident1, ident1AddrKeys := tpkg.RandEd25519Identity()
	_, ident2, ident2AddrKeys := tpkg.RandEd25519Identity()
	_, ident3, ident3AddrKeys := tpkg.RandEd25519Identity()
	alias1Ident := iotago.AliasIDFromOutputID(inputIDs[0])

	txBuilder := builder.NewTxBuilder(iotago.OutputSet{
		inputIDs[0]: &iotago.AliasOutput{
			Amount:               1000,
			AliasID:              iotago.AliasID{},
			StateController:      ident1,
			GovernanceController: ident2,
			StateIndex:           0,
			StateMetadata:        nil,
			FoundryCounter:       0,
			Blocks:               nil,
		},
		inputIDs[1]: &iotago.ExtendedOutput{
			Address: alias1Ident.ToAddress(),
			Amount:  1000,
		},
	})

	txBuilder.
		Transfer(alias1Ident, 1000, nil, nil, builder.TransferOpts{
			Lock: &builder.Lock{
				Milestone:     500,
				UnixTimestamp: 0,
			},
			Sender: &builder.Sender{
				Ident: ident1,
				ExpireOn: builder.Expiration{
					Milestone:     1000,
					UnixTimestamp: 0,
				},
				Return: 500,
			},
			Index:    []byte("purchase-23847258"),
			Metadata: []byte("token-transfer"),
		}).
		Transfer(alias1Ident, 0, nfTID, nil, builder.TransferOpts{
			Lock: &builder.Lock{
				Milestone:     500,
				UnixTimestamp: 0,
			},
			Sender: &builder.Sender{
				Ident: ident2,
				ExpireOn: builder.Expiration{
					Milestone:     1000,
					UnixTimestamp: 0,
				},
				Return: 500,
			},
			Index:    []byte("nft-transfer-1238573"),
			Metadata: []byte("nft-transfer"),
		}).
		Alias(alias1ID).GovTransition(builder.AliasGovOpts{
			StateController:      ident3,
			GovernanceController: ident3,
			Metadata:             []byte("gov transition"),
		}).
		Alias(alias2ID).StateTransition(builder.AliasStateOpts{
			Metadata:         []byte("alias metadata"),
			NewFoundries: builder.Foundries{
				{
					TokenTag:          tpkg.Rand12ByteArray(),
					CirculatingSupply: new(big.Int).SetUint64(100),
					MaximumSupply:     new(big.Int).SetUint64(1000),
				},
				{
					TokenTag:          tpkg.Rand12ByteArray(),
					CirculatingSupply: new(big.Int).SetUint64(500),
					MaximumSupply:     new(big.Int).SetUint64(10000),
				},
			},
		}).
		Alias(alias3ID).Destroy().
		MintNFT(alias2Ident, builder.NFTOpts{
			ImmutableMetadata: []byte("ipfs-link"),
			Issuer:            artistIdent1,
		}).
		Foundry(foundry1ID).Burn(big.NewInt(10)).
		Foundry(foundry2ID).Mint(big.NewInt(10)).
		Foundry(foundry2ID).Destroy().
		NFT(nftToBurnID).Destroy().
		Remainder(ident2).
		Build(keys)

}
 */
