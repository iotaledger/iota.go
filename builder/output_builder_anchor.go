package builder

import (
	"github.com/iotaledger/hive.go/ierrors"
	iotago "github.com/iotaledger/iota.go/v4"
)

// NewAnchorOutputBuilder creates a new AnchorOutputBuilder with the required state controller/governor addresses and base token amount.
func NewAnchorOutputBuilder(stateCtrl iotago.Address, govAddr iotago.Address, amount iotago.BaseToken) *AnchorOutputBuilder {
	return &AnchorOutputBuilder{output: &iotago.AnchorOutput{
		Amount:     amount,
		Mana:       0,
		AnchorID:   iotago.EmptyAnchorID,
		StateIndex: 0,
		UnlockConditions: iotago.AnchorOutputUnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl},
			&iotago.GovernorAddressUnlockCondition{Address: govAddr},
		},
		Features:          iotago.AnchorOutputFeatures{},
		ImmutableFeatures: iotago.AnchorOutputImmFeatures{},
	}}
}

// NewAnchorOutputBuilderFromPrevious creates a new AnchorOutputBuilder starting from a copy of the previous iotago.AnchorOutput.
func NewAnchorOutputBuilderFromPrevious(previous *iotago.AnchorOutput) *AnchorOutputBuilder {
	return &AnchorOutputBuilder{
		prev: previous,
		//nolint:forcetypeassert // we can safely assume that this is an AnchorOutput
		output: previous.Clone().(*iotago.AnchorOutput),
	}
}

// AnchorOutputBuilder builds an iotago.AnchorOutput.
type AnchorOutputBuilder struct {
	prev         *iotago.AnchorOutput
	output       *iotago.AnchorOutput
	stateCtrlReq bool
	govCtrlReq   bool
}

// Amount sets the base token amount of the output.
func (builder *AnchorOutputBuilder) Amount(amount iotago.BaseToken) *AnchorOutputBuilder {
	builder.output.Amount = amount
	builder.stateCtrlReq = true

	return builder
}

// Mana sets the mana of the output.
func (builder *AnchorOutputBuilder) Mana(mana iotago.Mana) *AnchorOutputBuilder {
	builder.output.Mana = mana

	return builder
}

// AnchorID sets the iotago.AnchorID of this output.
// Do not call this function if the underlying iotago.AnchorOutput is not new.
func (builder *AnchorOutputBuilder) AnchorID(anchorID iotago.AnchorID) *AnchorOutputBuilder {
	builder.output.AnchorID = anchorID

	return builder
}

// StateController sets the iotago.StateControllerAddressUnlockCondition of the output.
func (builder *AnchorOutputBuilder) StateController(stateCtrl iotago.Address) *AnchorOutputBuilder {
	builder.output.UnlockConditions.Upsert(&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl})
	builder.govCtrlReq = true

	return builder
}

// Governor sets the iotago.GovernorAddressUnlockCondition of the output.
func (builder *AnchorOutputBuilder) Governor(governor iotago.Address) *AnchorOutputBuilder {
	builder.output.UnlockConditions.Upsert(&iotago.GovernorAddressUnlockCondition{Address: governor})
	builder.govCtrlReq = true

	return builder
}

// Sender sets/modifies an iotago.SenderFeature as a mutable feature on the output.
func (builder *AnchorOutputBuilder) Sender(senderAddr iotago.Address) *AnchorOutputBuilder {
	builder.output.Features.Upsert(&iotago.SenderFeature{Address: senderAddr})
	builder.govCtrlReq = true

	return builder
}

// ImmutableSender sets/modifies an iotago.SenderFeature as an immutable feature on the output.
// Only call this function on a new iotago.AnchorOutput.
func (builder *AnchorOutputBuilder) ImmutableSender(senderAddr iotago.Address) *AnchorOutputBuilder {
	builder.output.ImmutableFeatures.Upsert(&iotago.SenderFeature{Address: senderAddr})

	return builder
}

// Metadata sets/modifies an iotago.MetadataFeature on the output.
func (builder *AnchorOutputBuilder) Metadata(entries iotago.MetadataFeatureEntries) *AnchorOutputBuilder {
	builder.output.Features.Upsert(&iotago.MetadataFeature{Entries: entries})
	builder.stateCtrlReq = true

	return builder
}

// GovernorMetadata sets/modifies an iotago.GovernorMetadataFeature on the output.
func (builder *AnchorOutputBuilder) GovernorMetadata(entries iotago.GovernorMetadataFeatureEntries) *AnchorOutputBuilder {
	builder.output.Features.Upsert(&iotago.GovernorMetadataFeature{Entries: entries})
	builder.govCtrlReq = true

	return builder
}

// ImmutableMetadata sets/modifies an iotago.MetadataFeature as an immutable feature on the output.
// Only call this function on a new iotago.AnchorOutput.
func (builder *AnchorOutputBuilder) ImmutableMetadata(entries iotago.MetadataFeatureEntries) *AnchorOutputBuilder {
	builder.output.ImmutableFeatures.Upsert(&iotago.MetadataFeature{Entries: entries})

	return builder
}

// Build builds the iotago.AnchorOutput.
func (builder *AnchorOutputBuilder) Build() (*iotago.AnchorOutput, error) {
	if builder.prev != nil && builder.govCtrlReq && builder.stateCtrlReq {
		return nil, ierrors.New("builder calls require both state and governor transitions which is not possible")
	}

	if builder.prev != nil {
		if builder.stateCtrlReq {
			builder.output.StateIndex++
		}
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
func (builder *AnchorOutputBuilder) MustBuild() *iotago.AnchorOutput {
	output, err := builder.Build()
	if err != nil {
		panic(err)
	}

	return output
}

type AnchorStateTransition struct {
	builder *AnchorOutputBuilder
}

// StateTransition narrows the builder functions to the ones available for an anchor state transition.
//
//nolint:revive
func (builder *AnchorOutputBuilder) StateTransition() *AnchorStateTransition {
	return &AnchorStateTransition{builder: builder}
}

// Amount sets the base token amount of the output.
func (trans *AnchorStateTransition) Amount(amount iotago.BaseToken) *AnchorStateTransition {
	return trans.builder.Amount(amount).StateTransition()
}

// Mana sets the mana of the output.
func (trans *AnchorStateTransition) Mana(mana iotago.Mana) *AnchorStateTransition {
	return trans.builder.Mana(mana).StateTransition()
}

// Metadata sets/modifies an iotago.MetadataFeature as a mutable feature on the output.
func (trans *AnchorStateTransition) Metadata(entries iotago.MetadataFeatureEntries) *AnchorStateTransition {
	return trans.builder.Metadata(entries).StateTransition()
}

// Sender sets/modifies an iotago.SenderFeature as a mutable feature on the output.
func (trans *AnchorStateTransition) Sender(senderAddr iotago.Address) *AnchorStateTransition {
	return trans.builder.Sender(senderAddr).StateTransition()
}

// Builder returns the AnchorOutputBuilder.
func (trans *AnchorStateTransition) Builder() *AnchorOutputBuilder {
	return trans.builder
}

type AnchorGovernanceTransition struct {
	builder *AnchorOutputBuilder
}

// GovernanceTransition narrows the builder functions to the ones available for an anchor governance transition.
//
//nolint:revive
func (builder *AnchorOutputBuilder) GovernanceTransition() *AnchorGovernanceTransition {
	return &AnchorGovernanceTransition{builder: builder}
}

// StateController sets the iotago.StateControllerAddressUnlockCondition of the output.
func (trans *AnchorGovernanceTransition) StateController(stateCtrl iotago.Address) *AnchorGovernanceTransition {
	return trans.builder.StateController(stateCtrl).GovernanceTransition()
}

// Governor sets the iotago.GovernorAddressUnlockCondition of the output.
func (trans *AnchorGovernanceTransition) Governor(governor iotago.Address) *AnchorGovernanceTransition {
	return trans.builder.Governor(governor).GovernanceTransition()
}

// Sender sets/modifies an iotago.SenderFeature as a mutable feature on the output.
func (trans *AnchorGovernanceTransition) Sender(senderAddr iotago.Address) *AnchorGovernanceTransition {
	return trans.builder.Sender(senderAddr).GovernanceTransition()
}

// GovernorMetadata sets/modifies an iotago.GovernorMetadataFeature on the output.
func (trans *AnchorGovernanceTransition) GovernorMetadata(entries iotago.GovernorMetadataFeatureEntries) *AnchorGovernanceTransition {
	return trans.builder.GovernorMetadata(entries).GovernanceTransition()
}

// Builder returns the AnchorOutputBuilder.
func (trans *AnchorGovernanceTransition) Builder() *AnchorOutputBuilder {
	return trans.builder
}
