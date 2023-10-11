package builder

import (
	"github.com/iotaledger/hive.go/ierrors"
	iotago "github.com/iotaledger/iota.go/v4"
)

// NewAnchorOutputBuilder creates a new AnchorOutputBuilder with the required state controller/governor addresses and base token amount.
func NewAnchorOutputBuilder(stateCtrl iotago.Address, govAddr iotago.Address, amount iotago.BaseToken) *AnchorOutputBuilder {
	return &AnchorOutputBuilder{output: &iotago.AnchorOutput{
		Amount:        amount,
		Mana:          0,
		AnchorID:      iotago.EmptyAnchorID(),
		StateIndex:    0,
		StateMetadata: []byte{},
		Conditions: iotago.AnchorOutputUnlockConditions{
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
	builder.stateCtrlReq = true

	return builder
}

// AnchorID sets the iotago.AnchorID of this output.
// Do not call this function if the underlying iotago.AnchorOutput is not new.
func (builder *AnchorOutputBuilder) AnchorID(anchorID iotago.AnchorID) *AnchorOutputBuilder {
	builder.output.AnchorID = anchorID

	return builder
}

// StateMetadata sets the state metadata of the output.
func (builder *AnchorOutputBuilder) StateMetadata(data []byte) *AnchorOutputBuilder {
	builder.output.StateMetadata = data
	builder.stateCtrlReq = true

	return builder
}

// StateController sets the iotago.StateControllerAddressUnlockCondition of the output.
func (builder *AnchorOutputBuilder) StateController(stateCtrl iotago.Address) *AnchorOutputBuilder {
	builder.output.Conditions.Upsert(&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl})
	builder.govCtrlReq = true

	return builder
}

// Governor sets the iotago.GovernorAddressUnlockCondition of the output.
func (builder *AnchorOutputBuilder) Governor(governor iotago.Address) *AnchorOutputBuilder {
	builder.output.Conditions.Upsert(&iotago.GovernorAddressUnlockCondition{Address: governor})
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
func (builder *AnchorOutputBuilder) Metadata(data []byte) *AnchorOutputBuilder {
	builder.output.Features.Upsert(&iotago.MetadataFeature{Data: data})
	builder.govCtrlReq = true

	return builder
}

// ImmutableMetadata sets/modifies an iotago.MetadataFeature as an immutable feature on the output.
// Only call this function on a new iotago.AnchorOutput.
func (builder *AnchorOutputBuilder) ImmutableMetadata(data []byte) *AnchorOutputBuilder {
	builder.output.ImmutableFeatures.Upsert(&iotago.MetadataFeature{Data: data})

	return builder
}

// Build builds the iotago.AnchorOutput.
func (builder *AnchorOutputBuilder) Build() (*iotago.AnchorOutput, error) {
	if builder.prev != nil && builder.govCtrlReq && builder.stateCtrlReq {
		return nil, ierrors.New("builder calls require both state and governor transitions which is not possible")
	}

	if builder.stateCtrlReq {
		builder.output.StateIndex++
	}

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
func (builder *AnchorOutputBuilder) MustBuild() *iotago.AnchorOutput {
	output, err := builder.Build()
	if err != nil {
		panic(err)
	}

	return output
}

type anchorStateTransition struct {
	builder *AnchorOutputBuilder
}

// StateTransition narrows the builder functions to the ones available for an anchor state transition.
//
//nolint:revive
func (builder *AnchorOutputBuilder) StateTransition() *anchorStateTransition {
	return &anchorStateTransition{builder: builder}
}

// Amount sets the base token amount of the output.
func (trans *anchorStateTransition) Amount(amount iotago.BaseToken) *anchorStateTransition {
	return trans.builder.Amount(amount).StateTransition()
}

// Mana sets the mana of the output.
func (trans *anchorStateTransition) Mana(mana iotago.Mana) *anchorStateTransition {
	return trans.builder.Mana(mana).StateTransition()
}

// StateMetadata sets the state metadata of the output.
func (trans *anchorStateTransition) StateMetadata(data []byte) *anchorStateTransition {
	return trans.builder.StateMetadata(data).StateTransition()
}

// Sender sets/modifies an iotago.SenderFeature as a mutable feature on the output.
func (trans *anchorStateTransition) Sender(senderAddr iotago.Address) *anchorStateTransition {
	return trans.builder.Sender(senderAddr).StateTransition()
}

// Builder returns the AnchorOutputBuilder.
func (trans *anchorStateTransition) Builder() *AnchorOutputBuilder {
	return trans.builder
}

type anchorGovernanceTransition struct {
	builder *AnchorOutputBuilder
}

// GovernanceTransition narrows the builder functions to the ones available for an anchor governance transition.
//
//nolint:revive
func (builder *AnchorOutputBuilder) GovernanceTransition() *anchorGovernanceTransition {
	return &anchorGovernanceTransition{builder: builder}
}

// StateController sets the iotago.StateControllerAddressUnlockCondition of the output.
func (trans *anchorGovernanceTransition) StateController(stateCtrl iotago.Address) *anchorGovernanceTransition {
	return trans.builder.StateController(stateCtrl).GovernanceTransition()
}

// Governor sets the iotago.GovernorAddressUnlockCondition of the output.
func (trans *anchorGovernanceTransition) Governor(governor iotago.Address) *anchorGovernanceTransition {
	return trans.builder.Governor(governor).GovernanceTransition()
}

// Sender sets/modifies an iotago.SenderFeature as a mutable feature on the output.
func (trans *anchorGovernanceTransition) Sender(senderAddr iotago.Address) *anchorGovernanceTransition {
	return trans.builder.Sender(senderAddr).GovernanceTransition()
}

// Metadata sets/modifies an iotago.MetadataFeature as a mutable feature on the output.
func (trans *anchorGovernanceTransition) Metadata(data []byte) *anchorGovernanceTransition {
	return trans.builder.Metadata(data).GovernanceTransition()
}

// Builder returns the AnchorOutputBuilder.
func (trans *anchorGovernanceTransition) Builder() *AnchorOutputBuilder {
	return trans.builder
}
