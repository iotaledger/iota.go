package builder

import (
	"github.com/iotaledger/hive.go/ierrors"
	iotago "github.com/iotaledger/iota.go/v4"
)

// NewNFTOutputBuilder creates a new NFTOutputBuilder with the address and base token amount.
func NewNFTOutputBuilder(targetAddr iotago.Address, amount iotago.BaseToken) *NFTOutputBuilder {
	return &NFTOutputBuilder{output: &iotago.NFTOutput{
		Amount: amount,
		Mana:   0,
		NFTID:  iotago.EmptyNFTID(),
		Conditions: iotago.NFTOutputUnlockConditions{
			&iotago.AddressUnlockCondition{Address: targetAddr},
		},
		Features:          iotago.NFTOutputFeatures{},
		ImmutableFeatures: iotago.NFTOutputImmFeatures{},
	}}
}

// NewNFTOutputBuilderFromPrevious creates a new NFTOutputBuilder starting from a copy of the previous iotago.NFTOutput.
func NewNFTOutputBuilderFromPrevious(previous *iotago.NFTOutput) *NFTOutputBuilder {
	return &NFTOutputBuilder{
		prev: previous,
		//nolint:forcetypeassert // we can safely assume that this is a NFTOutput
		output: previous.Clone().(*iotago.NFTOutput),
	}
}

// NFTOutputBuilder builds an iotago.NFTOutput.
type NFTOutputBuilder struct {
	prev   *iotago.NFTOutput
	output *iotago.NFTOutput
}

// Amount sets the base token amount of the output.
func (builder *NFTOutputBuilder) Amount(amount iotago.BaseToken) *NFTOutputBuilder {
	builder.output.Amount = amount

	return builder
}

// Amount sets the mana of the output.
func (builder *NFTOutputBuilder) Mana(mana iotago.Mana) *NFTOutputBuilder {
	builder.output.Mana = mana

	return builder
}

// Address sets/modifies an iotago.AddressUnlockCondition on the output.
func (builder *NFTOutputBuilder) Address(addr iotago.Address) *NFTOutputBuilder {
	builder.output.Conditions.Upsert(&iotago.AddressUnlockCondition{Address: addr})

	return builder
}

// NFTID sets the iotago.NFTID of this output.
// Do not call this function if the underlying iotago.NFTID is not new.
func (builder *NFTOutputBuilder) NFTID(nftID iotago.NFTID) *NFTOutputBuilder {
	builder.output.NFTID = nftID

	return builder
}

// StorageDepositReturn sets/modifies an iotago.StorageDepositReturnUnlockCondition on the output.
func (builder *NFTOutputBuilder) StorageDepositReturn(returnAddr iotago.Address, amount iotago.BaseToken) *NFTOutputBuilder {
	builder.output.Conditions.Upsert(&iotago.StorageDepositReturnUnlockCondition{ReturnAddress: returnAddr, Amount: amount})

	return builder
}

// Timelock sets/modifies an iotago.TimelockUnlockCondition on the output.
func (builder *NFTOutputBuilder) Timelock(untilSlot iotago.SlotIndex) *NFTOutputBuilder {
	builder.output.Conditions.Upsert(&iotago.TimelockUnlockCondition{Slot: untilSlot})

	return builder
}

// Expiration sets/modifies an iotago.ExpirationUnlockCondition on the output.
func (builder *NFTOutputBuilder) Expiration(returnAddr iotago.Address, expiredAfterSlot iotago.SlotIndex) *NFTOutputBuilder {
	builder.output.Conditions.Upsert(&iotago.ExpirationUnlockCondition{ReturnAddress: returnAddr, Slot: expiredAfterSlot})

	return builder
}

// Sender sets/modifies an iotago.SenderFeature on the output.
func (builder *NFTOutputBuilder) Sender(senderAddr iotago.Address) *NFTOutputBuilder {
	builder.output.Features.Upsert(&iotago.SenderFeature{Address: senderAddr})

	return builder
}

// Metadata sets/modifies an iotago.MetadataFeature on the output.
func (builder *NFTOutputBuilder) Metadata(data []byte) *NFTOutputBuilder {
	builder.output.Features.Upsert(&iotago.MetadataFeature{Data: data})

	return builder
}

// ImmutableMetadata sets/modifies an iotago.MetadataFeature as an immutable feature on the output.
// Only call this function on a new iotago.NFTOutput.
func (builder *NFTOutputBuilder) ImmutableMetadata(data []byte) *NFTOutputBuilder {
	builder.output.ImmutableFeatures.Upsert(&iotago.MetadataFeature{Data: data})

	return builder
}

// Tag sets/modifies an iotago.TagFeature on the output.
func (builder *NFTOutputBuilder) Tag(tag []byte) *NFTOutputBuilder {
	builder.output.Features.Upsert(&iotago.TagFeature{Tag: tag})

	return builder
}

// ImmutableIssuer sets/modifies an iotago.IssuerFeature as an immutable feature on the output.
// Only call this function on a new iotago.NFTOutput.
func (builder *NFTOutputBuilder) ImmutableIssuer(issuer iotago.Address) *NFTOutputBuilder {
	builder.output.ImmutableFeatures.Upsert(&iotago.IssuerFeature{Address: issuer})

	return builder
}

// Build builds the iotago.FoundryOutput.
func (builder *NFTOutputBuilder) Build() (*iotago.NFTOutput, error) {
	if builder.prev != nil {
		if !builder.prev.ImmutableFeatures.Equal(builder.output.ImmutableFeatures) {
			return nil, ierrors.New("immutable features are not allowed to be changed")
		}
	}

	builder.output.Conditions.Sort()
	builder.output.Features.Sort()
	builder.output.ImmutableFeatures.Sort()

	return builder.output, nil
}

// MustBuild works like Build() but panics if an error is encountered.
func (builder *NFTOutputBuilder) MustBuild() *iotago.NFTOutput {
	output, err := builder.Build()
	if err != nil {
		panic(err)
	}

	return output
}
