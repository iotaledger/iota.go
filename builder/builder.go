package builder

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
)

func NewTxBuilder(inputs iotago.OutputSet) *TxBuilder {
	return &TxBuilder{
		inputs:   inputs,
		inChains: inputs.ChainConstrainedOutputSet(),
	}
}

type TxBuilder struct {
	inputs   iotago.OutputSet
	inChains iotago.ChainConstrainedOutputsSet
}

type (
	TransferOpts struct {
		Lock     *Lock
		Sender   *Sender
		Index    []byte
		Metadata []byte
	}

	Lock struct {
		Milestone     uint32
		UnixTimestamp uint64
	}

	Expiration struct {
		Milestone     uint32
		UnixTimestamp uint64
	}

	Sender struct {
		Ident    iotago.Address
		ExpireOn Expiration
		Return   uint64
	}

	NFTOpts struct {
		ImmutableMetadata []byte
		Issuer            iotago.Address
	}
)

func (txBuilder *TxBuilder) Transfer(addr iotago.Address, iotaTokens uint64, nfts iotago.NFTIDs, nativeTokens iotago.NativeTokensSet, opts TransferOpts) *TxBuilder {

	return txBuilder
}

func (txBuilder *TxBuilder) MintNFT(addr iotago.Address, opts NFTOpts) *TxBuilder {

	return txBuilder
}

type (
	AliasBuilder struct {
		txBuilder *TxBuilder
	}

	AliasGovOpts struct {
		StateController      iotago.Address
		GovernanceController iotago.Address
		Metadata             []byte
	}

	Foundry struct {
		TokenTag          iotago.TokenTag
		CirculatingSupply *big.Int
		MaximumSupply     *big.Int
	}

	Foundries []Foundry

	AliasStateOpts struct {
		NewFoundries Foundries
		Metadata     []byte
	}
)

func (aliasBuilder *AliasBuilder) GovTransition(opts AliasGovOpts) *TxBuilder {
	return aliasBuilder.txBuilder
}

func (aliasBuilder *AliasBuilder) StateTransition(opts AliasStateOpts) *TxBuilder {
	return aliasBuilder.txBuilder
}

func (aliasBuilder *AliasBuilder) Destroy() *TxBuilder {
	return aliasBuilder.txBuilder
}

func (txBuilder *TxBuilder) Alias(id iotago.AliasID) *AliasBuilder {
	return &AliasBuilder{txBuilder: txBuilder}
}

type (
	FoundryBuilder struct {
		txBuilder *TxBuilder
	}
)

func (foundryBuilder *FoundryBuilder) Burn(tokensToBurn *big.Int) *TxBuilder {
	return foundryBuilder.txBuilder
}

func (foundryBuilder *FoundryBuilder) Mint(tokensToMint *big.Int) *TxBuilder {
	return foundryBuilder.txBuilder
}

func (foundryBuilder *FoundryBuilder) Destroy() *TxBuilder {
	return foundryBuilder.txBuilder
}

func (txBuilder *TxBuilder) Foundry(id iotago.FoundryID) *FoundryBuilder {
	return &FoundryBuilder{txBuilder: txBuilder}
}

type (
	NFTBuilder struct {
		txBuilder *TxBuilder
	}
)

func (nftBuilder *NFTBuilder) Destroy() *TxBuilder {
	return nftBuilder.txBuilder
}

func (txBuilder *TxBuilder) NFT(nftID iotago.NFTID) *NFTBuilder {
	return &NFTBuilder{txBuilder: txBuilder}
}
