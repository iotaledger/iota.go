package builder

import (
	"github.com/iotaledger/hive.go/ierrors"
	iotago "github.com/iotaledger/iota.go/v4"
)

// NewFoundryOutputBuilder creates a new FoundryOutputBuilder with the account address, serial number, token scheme and base token amount.
func NewFoundryOutputBuilder(accountAddr *iotago.AccountAddress, amount iotago.BaseToken, serialNumber uint32, tokenScheme iotago.TokenScheme) *FoundryOutputBuilder {
	return &FoundryOutputBuilder{output: &iotago.FoundryOutput{
		Amount:       amount,
		SerialNumber: serialNumber,
		TokenScheme:  tokenScheme,
		UnlockConditions: iotago.FoundryOutputUnlockConditions{
			&iotago.ImmutableAccountUnlockCondition{Address: accountAddr},
		},
		Features:          iotago.FoundryOutputFeatures{},
		ImmutableFeatures: iotago.FoundryOutputImmFeatures{},
	}}
}

// NewFoundryOutputBuilderFromPrevious creates a new FoundryOutputBuilder starting from a copy of the previous iotago.FoundryOutput.
func NewFoundryOutputBuilderFromPrevious(previous *iotago.FoundryOutput) *FoundryOutputBuilder {
	return &FoundryOutputBuilder{
		prev: previous,
		//nolint:forcetypeassert // we can safely assume that this is a FoundryOutput
		output: previous.Clone().(*iotago.FoundryOutput),
	}
}

// FoundryOutputBuilder builds an iotago.FoundryOutput.
type FoundryOutputBuilder struct {
	prev   *iotago.FoundryOutput
	output *iotago.FoundryOutput
}

// Amount sets the base token amount of the output.
func (builder *FoundryOutputBuilder) Amount(amount iotago.BaseToken) *FoundryOutputBuilder {
	builder.output.Amount = amount

	return builder
}

// NativeToken adds/modifies a native token to/on the output.
func (builder *FoundryOutputBuilder) NativeToken(nt *iotago.NativeTokenFeature) *FoundryOutputBuilder {
	builder.output.Features.Upsert(nt)

	return builder
}

// Metadata sets/modifies an iotago.MetadataFeature on the output.
func (builder *FoundryOutputBuilder) Metadata(entries iotago.MetadataFeatureEntries) *FoundryOutputBuilder {
	builder.output.Features.Upsert(&iotago.MetadataFeature{Entries: entries})

	return builder
}

// ImmutableMetadata sets/modifies an iotago.MetadataFeature as an immutable feature on the output.
// Only call this function on a new iotago.FoundryOutput.
func (builder *FoundryOutputBuilder) ImmutableMetadata(entries iotago.MetadataFeatureEntries) *FoundryOutputBuilder {
	builder.output.ImmutableFeatures.Upsert(&iotago.MetadataFeature{Entries: entries})

	return builder
}

// Build builds the iotago.FoundryOutput.
func (builder *FoundryOutputBuilder) Build() (*iotago.FoundryOutput, error) {
	if builder.prev != nil {
		if !builder.prev.ImmutableFeatures.Equal(builder.output.ImmutableFeatures) {
			return nil, ierrors.New("immutable features are not allowed to be changed")
		}
	}

	builder.output.UnlockConditions.Sort()
	builder.output.Features.Sort()
	builder.output.ImmutableFeatures.Sort()

	return builder.output, nil
}

// MustBuild works like Build() but panics if an error is encountered.
func (builder *FoundryOutputBuilder) MustBuild() *iotago.FoundryOutput {
	output, err := builder.Build()

	if err != nil {
		panic(err)
	}

	return output
}
