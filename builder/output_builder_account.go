package builder

import (
	"github.com/iotaledger/hive.go/ierrors"
	iotago "github.com/iotaledger/iota.go/v4"
)

// NewAccountOutputBuilder creates a new AccountOutputBuilder with the required state controller/governor addresses and base token amount.
func NewAccountOutputBuilder(stateCtrl iotago.Address, govAddr iotago.Address, amount iotago.BaseToken) *AccountOutputBuilder {
	return &AccountOutputBuilder{output: &iotago.AccountOutput{
		Amount:         amount,
		Mana:           0,
		AccountID:      iotago.EmptyAccountID(),
		StateIndex:     0,
		StateMetadata:  []byte{},
		FoundryCounter: 0,
		Conditions: iotago.AccountOutputUnlockConditions{
			&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl},
			&iotago.GovernorAddressUnlockCondition{Address: govAddr},
		},
		Features:          iotago.AccountOutputFeatures{},
		ImmutableFeatures: iotago.AccountOutputImmFeatures{},
	}}
}

// NewAccountOutputBuilderFromPrevious creates a new AccountOutputBuilder starting from a copy of the previous iotago.AccountOutput.
func NewAccountOutputBuilderFromPrevious(previous *iotago.AccountOutput) *AccountOutputBuilder {
	return &AccountOutputBuilder{
		prev: previous,
		//nolint:forcetypeassert // we can safely assume that this is an AccountOutput
		output: previous.Clone().(*iotago.AccountOutput),
	}
}

// AccountOutputBuilder builds an iotago.AccountOutput.
type AccountOutputBuilder struct {
	prev         *iotago.AccountOutput
	output       *iotago.AccountOutput
	stateCtrlReq bool
	govCtrlReq   bool
}

// Amount sets the base token amount of the output.
func (builder *AccountOutputBuilder) Amount(amount iotago.BaseToken) *AccountOutputBuilder {
	builder.output.Amount = amount
	builder.stateCtrlReq = true

	return builder
}

// Mana sets the mana of the output.
func (builder *AccountOutputBuilder) Mana(mana iotago.Mana) *AccountOutputBuilder {
	builder.output.Mana = mana

	return builder
}

// AccountID sets the iotago.AccountID of this output.
// Do not call this function if the underlying iotago.AccountOutput is not new.
func (builder *AccountOutputBuilder) AccountID(accountID iotago.AccountID) *AccountOutputBuilder {
	builder.output.AccountID = accountID

	return builder
}

// StateMetadata sets the state metadata of the output.
func (builder *AccountOutputBuilder) StateMetadata(data []byte) *AccountOutputBuilder {
	builder.output.StateMetadata = data
	builder.stateCtrlReq = true

	return builder
}

// FoundriesToGenerate bumps the output's foundry counter by the amount of foundries to generate.
func (builder *AccountOutputBuilder) FoundriesToGenerate(count uint32) *AccountOutputBuilder {
	builder.output.FoundryCounter += count
	builder.stateCtrlReq = true

	return builder
}

// StateController sets the iotago.StateControllerAddressUnlockCondition of the output.
func (builder *AccountOutputBuilder) StateController(stateCtrl iotago.Address) *AccountOutputBuilder {
	builder.output.Conditions.Upsert(&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl})
	builder.govCtrlReq = true

	return builder
}

// Governor sets the iotago.GovernorAddressUnlockCondition of the output.
func (builder *AccountOutputBuilder) Governor(governor iotago.Address) *AccountOutputBuilder {
	builder.output.Conditions.Upsert(&iotago.GovernorAddressUnlockCondition{Address: governor})
	builder.govCtrlReq = true

	return builder
}

// Staking sets/modifies an iotago.StakingFeature as a mutable feature on the output.
func (builder *AccountOutputBuilder) Staking(amount iotago.BaseToken, fixedCost iotago.Mana, startEpoch iotago.EpochIndex, optEndEpoch ...iotago.EpochIndex) *AccountOutputBuilder {
	endEpoch := iotago.MaxEpochIndex
	if len(optEndEpoch) > 0 {
		endEpoch = optEndEpoch[0]
	}

	builder.output.Features.Upsert(&iotago.StakingFeature{
		StakedAmount: amount,
		FixedCost:    fixedCost,
		StartEpoch:   startEpoch,
		EndEpoch:     endEpoch,
	})
	builder.govCtrlReq = true

	return builder
}

// BlockIssuer sets/modifies an iotago.BlockIssuerFeature as a mutable feature on the output.
func (builder *AccountOutputBuilder) BlockIssuer(keys iotago.BlockIssuerKeys, expirySlot iotago.SlotIndex) *AccountOutputBuilder {
	builder.output.Features.Upsert(&iotago.BlockIssuerFeature{
		BlockIssuerKeys: keys,
		ExpirySlot:      expirySlot,
	})
	builder.govCtrlReq = true

	return builder
}

// Sender sets/modifies an iotago.SenderFeature as a mutable feature on the output.
func (builder *AccountOutputBuilder) Sender(senderAddr iotago.Address) *AccountOutputBuilder {
	builder.output.Features.Upsert(&iotago.SenderFeature{Address: senderAddr})
	builder.govCtrlReq = true

	return builder
}

// ImmutableSender sets/modifies an iotago.SenderFeature as an immutable feature on the output.
// Only call this function on a new iotago.AccountOutput.
func (builder *AccountOutputBuilder) ImmutableSender(senderAddr iotago.Address) *AccountOutputBuilder {
	builder.output.ImmutableFeatures.Upsert(&iotago.SenderFeature{Address: senderAddr})

	return builder
}

// Metadata sets/modifies an iotago.MetadataFeature on the output.
func (builder *AccountOutputBuilder) Metadata(data []byte) *AccountOutputBuilder {
	builder.output.Features.Upsert(&iotago.MetadataFeature{Data: data})
	builder.govCtrlReq = true

	return builder
}

// ImmutableMetadata sets/modifies an iotago.MetadataFeature as an immutable feature on the output.
// Only call this function on a new iotago.AccountOutput.
func (builder *AccountOutputBuilder) ImmutableMetadata(data []byte) *AccountOutputBuilder {
	builder.output.ImmutableFeatures.Upsert(&iotago.MetadataFeature{Data: data})

	return builder
}

// Build builds the iotago.AccountOutput.
func (builder *AccountOutputBuilder) Build() (*iotago.AccountOutput, error) {
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
func (builder *AccountOutputBuilder) MustBuild() *iotago.AccountOutput {
	output, err := builder.Build()
	if err != nil {
		panic(err)
	}

	return output
}

type accountStateTransition struct {
	builder *AccountOutputBuilder
}

// StateTransition narrows the builder functions to the ones available for an account state transition.
//
//nolint:revive
func (builder *AccountOutputBuilder) StateTransition() *accountStateTransition {
	return &accountStateTransition{builder: builder}
}

// Amount sets the base token amount of the output.
func (trans *accountStateTransition) Amount(amount iotago.BaseToken) *accountStateTransition {
	return trans.builder.Amount(amount).StateTransition()
}

// Mana sets the mana of the output.
func (trans *accountStateTransition) Mana(mana iotago.Mana) *accountStateTransition {
	return trans.builder.Mana(mana).StateTransition()
}

// StateMetadata sets the state metadata of the output.
func (trans *accountStateTransition) StateMetadata(data []byte) *accountStateTransition {
	return trans.builder.StateMetadata(data).StateTransition()
}

// FoundriesToGenerate bumps the output's foundry counter by the amount of foundries to generate.
func (trans *accountStateTransition) FoundriesToGenerate(count uint32) *accountStateTransition {
	return trans.builder.FoundriesToGenerate(count).StateTransition()
}

// Sender sets/modifies an iotago.SenderFeature as a mutable feature on the output.
func (trans *accountStateTransition) Sender(senderAddr iotago.Address) *accountStateTransition {
	return trans.builder.Sender(senderAddr).StateTransition()
}

// Builder returns the AccountOutputBuilder.
func (trans *accountStateTransition) Builder() *AccountOutputBuilder {
	return trans.builder
}

type accountGovernanceTransition struct {
	builder *AccountOutputBuilder
}

// GovernanceTransition narrows the builder functions to the ones available for an account governance transition.
//
//nolint:revive
func (builder *AccountOutputBuilder) GovernanceTransition() *accountGovernanceTransition {
	return &accountGovernanceTransition{builder: builder}
}

// StateController sets the iotago.StateControllerAddressUnlockCondition of the output.
func (trans *accountGovernanceTransition) StateController(stateCtrl iotago.Address) *accountGovernanceTransition {
	return trans.builder.StateController(stateCtrl).GovernanceTransition()
}

// Governor sets the iotago.GovernorAddressUnlockCondition of the output.
func (trans *accountGovernanceTransition) Governor(governor iotago.Address) *accountGovernanceTransition {
	return trans.builder.Governor(governor).GovernanceTransition()
}

// BlockIssuerTransition narrows the builder functions to the ones available for an iotago.BlockIssuerFeature transition.
// If BlockIssuerFeature does not exist, it creates and sets an empty feature.
func (trans *accountGovernanceTransition) BlockIssuerTransition() *blockIssuerTransition {
	blockIssuerFeature := trans.builder.output.FeatureSet().BlockIssuer()
	if blockIssuerFeature == nil {
		blockIssuerFeature = &iotago.BlockIssuerFeature{
			BlockIssuerKeys: iotago.NewBlockIssuerKeys(),
			ExpirySlot:      0,
		}
	}

	return &blockIssuerTransition{
		feature: blockIssuerFeature,
		builder: trans.builder,
	}
}

// BlockIssuer sets/modifies an iotago.BlockIssuerFeature as a mutable feature on the output.
func (trans *accountGovernanceTransition) BlockIssuer(keys iotago.BlockIssuerKeys, expirySlot iotago.SlotIndex) *accountGovernanceTransition {
	return trans.builder.BlockIssuer(keys, expirySlot).GovernanceTransition()
}

// StakingTransition narrows the builder functions to the ones available for an iotago.StakingFeature transition.
// If StakingFeature does not exist, it creates and sets an empty feature.
func (trans *accountGovernanceTransition) StakingTransition() *stakingTransition {
	stakingFeature := trans.builder.output.FeatureSet().Staking()
	if stakingFeature == nil {
		stakingFeature = &iotago.StakingFeature{
			StakedAmount: 0,
			FixedCost:    0,
			StartEpoch:   0,
			EndEpoch:     0,
		}
	}

	return &stakingTransition{
		feature: stakingFeature,
		builder: trans.builder,
	}

}

// Staking sets/modifies an iotago.StakingFeature as a mutable feature on the output.
func (trans *accountGovernanceTransition) Staking(amount iotago.BaseToken, fixedCost iotago.Mana, startEpoch iotago.EpochIndex, optEndEpoch ...iotago.EpochIndex) *accountGovernanceTransition {
	return trans.builder.Staking(amount, fixedCost, startEpoch, optEndEpoch...).GovernanceTransition()
}

// Sender sets/modifies an iotago.SenderFeature as a mutable feature on the output.
func (trans *accountGovernanceTransition) Sender(senderAddr iotago.Address) *accountGovernanceTransition {
	return trans.builder.Sender(senderAddr).GovernanceTransition()
}

// Metadata sets/modifies an iotago.MetadataFeature as a mutable feature on the output.
func (trans *accountGovernanceTransition) Metadata(data []byte) *accountGovernanceTransition {
	return trans.builder.Metadata(data).GovernanceTransition()
}

// Builder returns the AccountOutputBuilder.
func (trans *accountGovernanceTransition) Builder() *AccountOutputBuilder {
	return trans.builder
}

type blockIssuerTransition struct {
	feature *iotago.BlockIssuerFeature
	builder *AccountOutputBuilder
}

// AddKeys adds the keys of the BlockIssuerFeature.
func (trans *blockIssuerTransition) AddKeys(keys ...iotago.BlockIssuerKey) *blockIssuerTransition {
	for _, key := range keys {
		blockIssuerKey := key
		trans.feature.BlockIssuerKeys.Add(blockIssuerKey)
	}

	return trans
}

// RemoveKey deletes the key of the iotago.BlockIssuerFeature.
func (trans *blockIssuerTransition) RemoveKey(keyToDelete iotago.BlockIssuerKey) *blockIssuerTransition {
	trans.feature.BlockIssuerKeys.Remove(keyToDelete)

	return trans
}

// Keys sets the keys of the iotago.BlockIssuerFeature.
func (trans *blockIssuerTransition) Keys(keys iotago.BlockIssuerKeys) *blockIssuerTransition {
	trans.feature.BlockIssuerKeys = keys

	return trans
}

// ExpirySlot sets the ExpirySlot of iotago.BlockIssuerFeature.
func (trans *blockIssuerTransition) ExpirySlot(slot iotago.SlotIndex) *blockIssuerTransition {
	trans.feature.ExpirySlot = slot

	return trans
}

// GovernanceTransition returns the accountGovernanceTransition.
func (trans *blockIssuerTransition) GovernanceTransition() *accountGovernanceTransition {
	return trans.builder.GovernanceTransition()
}

// Builder returns the AccountOutputBuilder.
func (trans *blockIssuerTransition) Builder() *AccountOutputBuilder {
	return trans.builder
}

type stakingTransition struct {
	feature *iotago.StakingFeature
	builder *AccountOutputBuilder
}

// StakedAmount sets the StakedAmount of iotago.StakingFeature.
func (trans *stakingTransition) StakedAmount(amount iotago.BaseToken) *stakingTransition {
	trans.feature.StakedAmount = amount

	return trans
}

// FixedCost sets the FixedCost of iotago.StakingFeature.
func (trans *stakingTransition) FixedCost(fixedCost iotago.Mana) *stakingTransition {
	trans.feature.FixedCost = fixedCost

	return trans
}

// StartEpoch sets the StartEpoch of iotago.StakingFeature.
func (trans *stakingTransition) StartEpoch(epoch iotago.EpochIndex) *stakingTransition {
	trans.feature.StartEpoch = epoch

	return trans
}

// EndEpoch sets the EndEpoch of iotago.StakingFeature.
func (trans *stakingTransition) EndEpoch(epoch iotago.EpochIndex) *stakingTransition {
	trans.feature.EndEpoch = epoch

	return trans
}

// GovernanceTransition returns the accountGovernanceTransition.
func (trans *stakingTransition) GovernanceTransition() *accountGovernanceTransition {
	return trans.builder.GovernanceTransition()
}

// Builder returns the AccountOutputBuilder.
func (trans *stakingTransition) Builder() *AccountOutputBuilder {
	return trans.builder
}
