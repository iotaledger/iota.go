package builder

import (
	"github.com/iotaledger/hive.go/ierrors"
	iotago "github.com/iotaledger/iota.go/v4"
)

// NewBasicOutputBuilder creates a new BasicOutputBuilder with the required target address and base token amount.
func NewBasicOutputBuilder(targetAddr iotago.Address, amount iotago.BaseToken) *BasicOutputBuilder {
	return &BasicOutputBuilder{output: &iotago.BasicOutput{
		Amount:       amount,
		NativeTokens: iotago.NativeTokens{},
		Conditions: iotago.BasicOutputUnlockConditions{
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

// Address sets/modifies an iotago.AddressUnlockCondition on the output.
func (builder *BasicOutputBuilder) Address(addr iotago.Address) *BasicOutputBuilder {
	builder.output.Conditions.Upsert(&iotago.AddressUnlockCondition{Address: addr})

	return builder
}

// NativeToken adds/modifies a native token to/on the output.
func (builder *BasicOutputBuilder) NativeToken(nt *iotago.NativeToken) *BasicOutputBuilder {
	builder.output.NativeTokens.Upsert(nt)

	return builder
}

// NativeTokens sets the native tokens held by the output.
func (builder *BasicOutputBuilder) NativeTokens(nts iotago.NativeTokens) *BasicOutputBuilder {
	builder.output.NativeTokens = nts

	return builder
}

// StorageDepositReturn sets/modifies an iotago.StorageDepositReturnUnlockCondition on the output.
func (builder *BasicOutputBuilder) StorageDepositReturn(returnAddr iotago.Address, amount iotago.BaseToken) *BasicOutputBuilder {
	builder.output.Conditions.Upsert(&iotago.StorageDepositReturnUnlockCondition{ReturnAddress: returnAddr, Amount: amount})

	return builder
}

// Timelock sets/modifies an iotago.TimelockUnlockCondition on the output.
func (builder *BasicOutputBuilder) Timelock(untilSlotIndex iotago.SlotIndex) *BasicOutputBuilder {
	builder.output.Conditions.Upsert(&iotago.TimelockUnlockCondition{SlotIndex: untilSlotIndex})

	return builder
}

// Expiration sets/modifies an iotago.ExpirationUnlockCondition on the output.
func (builder *BasicOutputBuilder) Expiration(returnAddr iotago.Address, expiredAfterSlotIndex iotago.SlotIndex) *BasicOutputBuilder {
	builder.output.Conditions.Upsert(&iotago.ExpirationUnlockCondition{ReturnAddress: returnAddr, SlotIndex: expiredAfterSlotIndex})

	return builder
}

// Sender sets/modifies an iotago.SenderFeature on the output.
func (builder *BasicOutputBuilder) Sender(senderAddr iotago.Address) *BasicOutputBuilder {
	builder.output.Features.Upsert(&iotago.SenderFeature{Address: senderAddr})

	return builder
}

// Metadata sets/modifies an iotago.MetadataFeature on the output.
func (builder *BasicOutputBuilder) Metadata(data []byte) *BasicOutputBuilder {
	builder.output.Features.Upsert(&iotago.MetadataFeature{Data: data})

	return builder
}

// Tag sets/modifies an iotago.TagFeature on the output.
func (builder *BasicOutputBuilder) Tag(tag []byte) *BasicOutputBuilder {
	builder.output.Features.Upsert(&iotago.TagFeature{Tag: tag})

	return builder
}

// MustBuild works like Build() but panics if an error is encountered.
func (builder *BasicOutputBuilder) MustBuild() *iotago.BasicOutput {
	output, err := builder.Build()
	if err != nil {
		panic(err)
	}

	return output
}

// Build builds the iotago.BasicOutput.
func (builder *BasicOutputBuilder) Build() (*iotago.BasicOutput, error) {
	builder.output.Conditions.Sort()
	builder.output.Features.Sort()
	builder.output.NativeTokens.Sort()

	return builder.output, nil
}

// NewAccountOutputBuilder creates a new AccountOutputBuilder with the required state controller/governor addresses and base token amount.
func NewAccountOutputBuilder(stateCtrl iotago.Address, govAddr iotago.Address, amount iotago.BaseToken) *AccountOutputBuilder {
	return &AccountOutputBuilder{output: &iotago.AccountOutput{
		Amount:       amount,
		NativeTokens: iotago.NativeTokens{},
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
func (builder *AccountOutputBuilder) Mana(amount iotago.Mana) *AccountOutputBuilder {
	builder.output.Mana = amount

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

// NativeToken adds/modifies a native token to/on the output.
func (builder *AccountOutputBuilder) NativeToken(nt *iotago.NativeToken) *AccountOutputBuilder {
	builder.output.NativeTokens.Upsert(nt)
	builder.stateCtrlReq = true

	return builder
}

// NativeTokens sets the native tokens held by the output.
func (builder *AccountOutputBuilder) NativeTokens(nts iotago.NativeTokens) *AccountOutputBuilder {
	builder.output.NativeTokens = nts
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

// MustBuild works like Build() but panics if an error is encountered.
func (builder *AccountOutputBuilder) MustBuild() *iotago.AccountOutput {
	output, err := builder.Build()
	if err != nil {
		panic(err)
	}

	return output
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
	builder.output.NativeTokens.Sort()

	return builder.output, nil
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

// StateMetadata sets the state metadata of the output.
func (trans *accountStateTransition) StateMetadata(data []byte) *accountStateTransition {
	return trans.builder.StateMetadata(data).StateTransition()
}

// FoundriesToGenerate bumps the output's foundry counter by the amount of foundries to generate.
func (trans *accountStateTransition) FoundriesToGenerate(count uint32) *accountStateTransition {
	return trans.builder.FoundriesToGenerate(count).StateTransition()
}

// NativeToken adds/modifies a native token to/on the output.
func (trans *accountStateTransition) NativeToken(nt *iotago.NativeToken) *accountStateTransition {
	return trans.builder.NativeToken(nt).StateTransition()
}

// NativeTokens sets the native tokens held by the output.
func (trans *accountStateTransition) NativeTokens(nts iotago.NativeTokens) *accountStateTransition {
	return trans.builder.NativeTokens(nts).StateTransition()
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
func (trans *blockIssuerTransition) ExpirySlot(index iotago.SlotIndex) *blockIssuerTransition {
	trans.feature.ExpirySlot = index

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
func (trans *stakingTransition) StartEpoch(index iotago.EpochIndex) *stakingTransition {
	trans.feature.StartEpoch = index

	return trans
}

// EndEpoch sets the EndEpoch of iotago.StakingFeature.
func (trans *stakingTransition) EndEpoch(index iotago.EpochIndex) *stakingTransition {
	trans.feature.EndEpoch = index

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

// NewFoundryOutputBuilder creates a new FoundryOutputBuilder with the account address, serial number, token scheme and base token amount.
func NewFoundryOutputBuilder(accountAddr *iotago.AccountAddress, tokenScheme iotago.TokenScheme, amount iotago.BaseToken) *FoundryOutputBuilder {
	return &FoundryOutputBuilder{output: &iotago.FoundryOutput{
		Amount:       amount,
		NativeTokens: iotago.NativeTokens{},
		TokenScheme:  tokenScheme,
		Conditions: iotago.FoundryOutputUnlockConditions{
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
func (builder *FoundryOutputBuilder) NativeToken(nt *iotago.NativeToken) *FoundryOutputBuilder {
	builder.output.NativeTokens.Upsert(nt)

	return builder
}

// NativeTokens sets the native tokens held by the output.
func (builder *FoundryOutputBuilder) NativeTokens(nts iotago.NativeTokens) *FoundryOutputBuilder {
	builder.output.NativeTokens = nts

	return builder
}

// Metadata sets/modifies an iotago.MetadataFeature on the output.
func (builder *FoundryOutputBuilder) Metadata(data []byte) *FoundryOutputBuilder {
	builder.output.Features.Upsert(&iotago.MetadataFeature{Data: data})

	return builder
}

// ImmutableMetadata sets/modifies an iotago.MetadataFeature as an immutable feature on the output.
// Only call this function on a new iotago.AccountOutput.
func (builder *FoundryOutputBuilder) ImmutableMetadata(data []byte) *FoundryOutputBuilder {
	builder.output.ImmutableFeatures.Upsert(&iotago.MetadataFeature{Data: data})

	return builder
}

// MustBuild works like Build() but panics if an error is encountered.
func (builder *FoundryOutputBuilder) MustBuild() *iotago.FoundryOutput {
	output, err := builder.Build()

	if err != nil {
		panic(err)
	}

	return output
}

// Build builds the iotago.FoundryOutput.
func (builder *FoundryOutputBuilder) Build() (*iotago.FoundryOutput, error) {
	if builder.prev != nil {
		if !builder.prev.ImmutableFeatures.Equal(builder.output.ImmutableFeatures) {
			return nil, ierrors.New("immutable features are not allowed to be changed")
		}
	}

	builder.output.Conditions.Sort()
	builder.output.Features.Sort()
	builder.output.ImmutableFeatures.Sort()
	builder.output.NativeTokens.Sort()

	return builder.output, nil
}

// NewNFTOutputBuilder creates a new NFTOutputBuilder with the address and base token amount.
func NewNFTOutputBuilder(targetAddr iotago.Address, amount iotago.BaseToken) *NFTOutputBuilder {
	return &NFTOutputBuilder{output: &iotago.NFTOutput{
		Amount:       amount,
		NativeTokens: iotago.NativeTokens{},
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

// Address sets/modifies an iotago.AddressUnlockCondition on the output.
func (builder *NFTOutputBuilder) Address(addr iotago.Address) *NFTOutputBuilder {
	builder.output.Conditions.Upsert(&iotago.AddressUnlockCondition{Address: addr})

	return builder
}

// NativeToken adds/modifies a native token to/on the output.
func (builder *NFTOutputBuilder) NativeToken(nt *iotago.NativeToken) *NFTOutputBuilder {
	builder.output.NativeTokens.Upsert(nt)

	return builder
}

// NativeTokens sets the native tokens held by the output.
func (builder *NFTOutputBuilder) NativeTokens(nts iotago.NativeTokens) *NFTOutputBuilder {
	builder.output.NativeTokens = nts

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
func (builder *NFTOutputBuilder) Timelock(untilSlotIndex iotago.SlotIndex) *NFTOutputBuilder {
	builder.output.Conditions.Upsert(&iotago.TimelockUnlockCondition{SlotIndex: untilSlotIndex})

	return builder
}

// Expiration sets/modifies an iotago.ExpirationUnlockCondition on the output.
func (builder *NFTOutputBuilder) Expiration(returnAddr iotago.Address, expiredAfterSlotIndex iotago.SlotIndex) *NFTOutputBuilder {
	builder.output.Conditions.Upsert(&iotago.ExpirationUnlockCondition{ReturnAddress: returnAddr, SlotIndex: expiredAfterSlotIndex})

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

// MustBuild works like Build() but panics if an error is encountered.
func (builder *NFTOutputBuilder) MustBuild() *iotago.NFTOutput {
	output, err := builder.Build()
	if err != nil {
		panic(err)
	}

	return output
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
	builder.output.NativeTokens.Sort()

	return builder.output, nil
}

// NewDelegationOutputBuilder creates a new DelegationOutputBuilder with the account address, serial number, token scheme and base token amount.
func NewDelegationOutputBuilder(validatorAddress *iotago.AccountAddress, addr iotago.Address, amount iotago.BaseToken) *DelegationOutputBuilder {
	return &DelegationOutputBuilder{output: &iotago.DelegationOutput{
		Amount:           amount,
		DelegatedAmount:  0,
		DelegationID:     iotago.DelegationID{},
		ValidatorAddress: validatorAddress,
		StartEpoch:       0,
		EndEpoch:         0,
		Conditions: iotago.DelegationOutputUnlockConditions{
			&iotago.AddressUnlockCondition{Address: addr},
		},
	}}
}

// NewDelegationOutputBuilderFromPrevious creates a new DelegationOutputBuilder starting from a copy of the previous iotago.DelegationOutput.
func NewDelegationOutputBuilderFromPrevious(previous *iotago.DelegationOutput) *DelegationOutputBuilder {
	return &DelegationOutputBuilder{
		//nolint:forcetypeassert // we can safely assume that this is a DelegationOutput
		output: previous.Clone().(*iotago.DelegationOutput),
	}
}

// DelegationOutputBuilder builds an iotago.DelegationOutput.
type DelegationOutputBuilder struct {
	output *iotago.DelegationOutput
}

// Amount sets the base token amount of the output.
func (builder *DelegationOutputBuilder) Amount(amount iotago.BaseToken) *DelegationOutputBuilder {
	builder.output.Amount = amount

	return builder
}

// DelegatedAmount sets the delegated amount of the output.
func (builder *DelegationOutputBuilder) DelegatedAmount(delegatedAmount iotago.BaseToken) *DelegationOutputBuilder {
	builder.output.DelegatedAmount = delegatedAmount

	return builder
}

// ValidatorAddress sets the validator address of the output.
func (builder *DelegationOutputBuilder) ValidatorAddress(validatorAddress *iotago.AccountAddress) *DelegationOutputBuilder {
	builder.output.ValidatorAddress = validatorAddress

	return builder
}

// DelegationID sets the delegation ID of the output.
func (builder *DelegationOutputBuilder) DelegationID(delegationID iotago.DelegationID) *DelegationOutputBuilder {
	builder.output.DelegationID = delegationID

	return builder
}

// StartEpoch sets the delegation start epoch.
func (builder *DelegationOutputBuilder) StartEpoch(startEpoch iotago.EpochIndex) *DelegationOutputBuilder {
	builder.output.StartEpoch = startEpoch

	return builder
}

// EndEpoch sets the delegation end epoch.
func (builder *DelegationOutputBuilder) EndEpoch(endEpoch iotago.EpochIndex) *DelegationOutputBuilder {
	builder.output.EndEpoch = endEpoch

	return builder
}

// Address sets/modifies an iotago.AddressUnlockCondition on the output.
func (builder *DelegationOutputBuilder) Address(addr iotago.Address) *DelegationOutputBuilder {
	builder.output.Conditions.Upsert(&iotago.AddressUnlockCondition{Address: addr})

	return builder
}

// MustBuild works like Build() but panics if an error is encountered.
func (builder *DelegationOutputBuilder) MustBuild() *iotago.DelegationOutput {
	output, err := builder.Build()
	if err != nil {
		panic(err)
	}

	return output
}

// Build builds the iotago.DelegationOutput.
func (builder *DelegationOutputBuilder) Build() (*iotago.DelegationOutput, error) {

	builder.output.Conditions.Sort()

	return builder.output, nil
}
