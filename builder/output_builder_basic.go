package builder

import (
	iotago "github.com/iotaledger/iota.go/v4"
)

// NewBasicOutputBuilder creates a new BasicOutputBuilder with the required target address and base token amount.
func NewBasicOutputBuilder(targetAddr iotago.Address, amount iotago.BaseToken) *BasicOutputBuilder {
	return &BasicOutputBuilder{output: &iotago.BasicOutput{
		Amount: amount,
		Mana:   0,
		UnlockConditions: iotago.BasicOutputUnlockConditions{
			&iotago.AddressUnlockCondition{Address: targetAddr},
		},
		Features: iotago.BasicOutputFeatures{},
	}}
}

// NewBasicOutputBuilderFromPrevious creates a new BasicOutputBuilder starting from a copy of the previous iotago.BasicOutput.
func NewBasicOutputBuilderFromPrevious(previous *iotago.BasicOutput) *BasicOutputBuilder {
	//nolint:forcetypeassert // we can safely assume that this is a BasicOutput
	return &BasicOutputBuilder{output: previous.Clone().(*iotago.BasicOutput)}
}

// BasicOutputBuilder builds an iotago.BasicOutput.
type BasicOutputBuilder struct {
	output *iotago.BasicOutput
}

// Amount sets the base token amount of the output.
func (builder *BasicOutputBuilder) Amount(amount iotago.BaseToken) *BasicOutputBuilder {
	builder.output.Amount = amount

	return builder
}

// Mana sets the mana of the output.
func (builder *BasicOutputBuilder) Mana(mana iotago.Mana) *BasicOutputBuilder {
	builder.output.Mana = mana

	return builder
}

// Address sets/modifies an iotago.AddressUnlockCondition on the output.
func (builder *BasicOutputBuilder) Address(addr iotago.Address) *BasicOutputBuilder {
	builder.output.UnlockConditions.Upsert(&iotago.AddressUnlockCondition{Address: addr})

	return builder
}

// StorageDepositReturn sets/modifies an iotago.StorageDepositReturnUnlockCondition on the output.
func (builder *BasicOutputBuilder) StorageDepositReturn(returnAddr iotago.Address, amount iotago.BaseToken) *BasicOutputBuilder {
	builder.output.UnlockConditions.Upsert(&iotago.StorageDepositReturnUnlockCondition{ReturnAddress: returnAddr, Amount: amount})

	return builder
}

// Timelock sets/modifies an iotago.TimelockUnlockCondition on the output.
func (builder *BasicOutputBuilder) Timelock(untilSlot iotago.SlotIndex) *BasicOutputBuilder {
	builder.output.UnlockConditions.Upsert(&iotago.TimelockUnlockCondition{Slot: untilSlot})

	return builder
}

// Expiration sets/modifies an iotago.ExpirationUnlockCondition on the output.
func (builder *BasicOutputBuilder) Expiration(returnAddr iotago.Address, expiredAfterSlot iotago.SlotIndex) *BasicOutputBuilder {
	builder.output.UnlockConditions.Upsert(&iotago.ExpirationUnlockCondition{ReturnAddress: returnAddr, Slot: expiredAfterSlot})

	return builder
}

// Sender sets/modifies an iotago.SenderFeature on the output.
func (builder *BasicOutputBuilder) Sender(senderAddr iotago.Address) *BasicOutputBuilder {
	builder.output.Features.Upsert(&iotago.SenderFeature{Address: senderAddr})

	return builder
}

// Metadata sets/modifies an iotago.MetadataFeature on the output.
func (builder *BasicOutputBuilder) Metadata(entries iotago.MetadataFeatureEntries) *BasicOutputBuilder {
	builder.output.Features.Upsert(&iotago.MetadataFeature{Entries: entries})

	return builder
}

// Tag sets/modifies an iotago.TagFeature on the output.
func (builder *BasicOutputBuilder) Tag(tag []byte) *BasicOutputBuilder {
	builder.output.Features.Upsert(&iotago.TagFeature{Tag: tag})

	return builder
}

// NativeToken adds/modifies a native token to/on the output.
func (builder *BasicOutputBuilder) NativeToken(nt *iotago.NativeTokenFeature) *BasicOutputBuilder {
	builder.output.Features.Upsert(nt)

	return builder
}

// Build builds the iotago.BasicOutput.
func (builder *BasicOutputBuilder) Build() (*iotago.BasicOutput, error) {
	builder.output.UnlockConditions.Sort()
	builder.output.Features.Sort()

	return builder.output, nil
}

// MustBuild works like Build() but panics if an error is encountered.
func (builder *BasicOutputBuilder) MustBuild() *iotago.BasicOutput {
	output, err := builder.Build()
	if err != nil {
		panic(err)
	}

	return output
}
