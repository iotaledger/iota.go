package builder

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
)

// NewBasicOutputBuilder creates a new BasicOutputBuilder with the required target address and deposit amount.
func NewBasicOutputBuilder(targetAddr iotago.Address, deposit uint64) *BasicOutputBuilder {
	return &BasicOutputBuilder{output: &iotago.BasicOutput{
		Amount:       deposit,
		NativeTokens: iotago.NativeTokens{},
		Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
			&iotago.AddressUnlockCondition{Address: targetAddr},
		},
		Features: iotago.Features[iotago.BasicOutputFeature]{},
	}}
}

// NewBasicOutputBuilderFromPrevious creates a new BasicOutputBuilder starting from a copy of the previous iotago.BasicOutput.
func NewBasicOutputBuilderFromPrevious(previous *iotago.BasicOutput) *BasicOutputBuilder {
	return &BasicOutputBuilder{output: previous.Clone().(*iotago.BasicOutput)}
}

// BasicOutputBuilder builds an iotago.BasicOutput.
type BasicOutputBuilder struct {
	output *iotago.BasicOutput
}

// Deposit sets the deposit of the output.
func (builder *BasicOutputBuilder) Deposit(deposit uint64) *BasicOutputBuilder {
	builder.output.Amount = deposit
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
func (builder *BasicOutputBuilder) StorageDepositReturn(returnAddr iotago.Address, amount uint64) *BasicOutputBuilder {
	builder.output.Conditions.Upsert(&iotago.StorageDepositReturnUnlockCondition{ReturnAddress: returnAddr, Amount: amount})
	return builder
}

// Timelock sets/modifies an iotago.TimelockUnlockCondition on the output.
func (builder *BasicOutputBuilder) Timelock(untilUnixTimeSec int64) *BasicOutputBuilder {
	builder.output.Conditions.Upsert(&iotago.TimelockUnlockCondition{UnixTime: uint32(untilUnixTimeSec)})
	return builder
}

// Expiration sets/modifies an iotago.ExpirationUnlockCondition on the output.
func (builder *BasicOutputBuilder) Expiration(returnAddr iotago.Address, expiredAfterUnixTimeSec int64) *BasicOutputBuilder {
	builder.output.Conditions.Upsert(&iotago.ExpirationUnlockCondition{ReturnAddress: returnAddr, UnixTime: uint32(expiredAfterUnixTimeSec)})
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

// NewAliasOutputBuilder creates a new AliasOutputBuilder with the required state controller/governor addresses and deposit amount.
func NewAliasOutputBuilder(stateCtrl iotago.Address, govAddr iotago.Address, deposit uint64) *AliasOutputBuilder {
	return &AliasOutputBuilder{output: &iotago.AliasOutput{
		Amount:       deposit,
		NativeTokens: iotago.NativeTokens{},
		Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
			&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl},
			&iotago.GovernorAddressUnlockCondition{Address: govAddr},
		},
		Features:          iotago.Features[iotago.AliasFeature]{},
		ImmutableFeatures: iotago.Features[iotago.AliasImmFeature]{},
	}}
}

// NewAliasOutputBuilderFromPrevious creates a new AliasOutputBuilder starting from a copy of the previous iotago.AliasOutput.
func NewAliasOutputBuilderFromPrevious(previous *iotago.AliasOutput) *AliasOutputBuilder {
	return &AliasOutputBuilder{
		prev:   previous,
		output: previous.Clone().(*iotago.AliasOutput),
	}
}

// AliasOutputBuilder builds an iotago.AliasOutput.
type AliasOutputBuilder struct {
	prev         *iotago.AliasOutput
	output       *iotago.AliasOutput
	stateCtrlReq bool
	govCtrlReq   bool
}

// Deposit sets the deposit of the output.
func (builder *AliasOutputBuilder) Deposit(deposit uint64) *AliasOutputBuilder {
	builder.output.Amount = deposit
	builder.stateCtrlReq = true
	return builder
}

// AliasID sets the iotago.AliasID of this output.
// Do not call this function if the underlying iotago.AliasOutput is not new.
func (builder *AliasOutputBuilder) AliasID(aliasID iotago.AliasID) *AliasOutputBuilder {
	builder.output.AliasID = aliasID
	return builder
}

// StateMetadata sets the state metadata of the output.
func (builder *AliasOutputBuilder) StateMetadata(data []byte) *AliasOutputBuilder {
	builder.output.StateMetadata = data
	builder.stateCtrlReq = true
	return builder
}

// FoundriesToGenerate bumps the output's foundry counter by the amount of foundries to generate.
func (builder *AliasOutputBuilder) FoundriesToGenerate(count uint32) *AliasOutputBuilder {
	builder.output.FoundryCounter += count
	builder.stateCtrlReq = true
	return builder
}

// NativeToken adds/modifies a native token to/on the output.
func (builder *AliasOutputBuilder) NativeToken(nt *iotago.NativeToken) *AliasOutputBuilder {
	builder.output.NativeTokens.Upsert(nt)
	builder.stateCtrlReq = true
	return builder
}

// NativeTokens sets the native tokens held by the output.
func (builder *AliasOutputBuilder) NativeTokens(nts iotago.NativeTokens) *AliasOutputBuilder {
	builder.output.NativeTokens = nts
	builder.stateCtrlReq = true
	return builder
}

// StateController sets the iotago.StateControllerAddressUnlockCondition of the output.
func (builder *AliasOutputBuilder) StateController(stateCtrl iotago.Address) *AliasOutputBuilder {
	builder.output.Conditions.Upsert(&iotago.StateControllerAddressUnlockCondition{Address: stateCtrl})
	builder.govCtrlReq = true
	return builder
}

// Governor sets the iotago.GovernorAddressUnlockCondition of the output.
func (builder *AliasOutputBuilder) Governor(governor iotago.Address) *AliasOutputBuilder {
	builder.output.Conditions.Upsert(&iotago.GovernorAddressUnlockCondition{Address: governor})
	builder.govCtrlReq = true
	return builder
}

// Sender sets/modifies an iotago.SenderFeature as a mutable feature on the output.
func (builder *AliasOutputBuilder) Sender(senderAddr iotago.Address) *AliasOutputBuilder {
	builder.output.Features.Upsert(&iotago.SenderFeature{Address: senderAddr})
	builder.govCtrlReq = true
	return builder
}

// ImmutableSender sets/modifies an iotago.SenderFeature as an immutable feature on the output.
// Only call this function on a new iotago.AliasOutput.
func (builder *AliasOutputBuilder) ImmutableSender(senderAddr iotago.Address) *AliasOutputBuilder {
	builder.output.ImmutableFeatures.Upsert(&iotago.SenderFeature{Address: senderAddr})
	return builder
}

// Metadata sets/modifies an iotago.MetadataFeature on the output.
func (builder *AliasOutputBuilder) Metadata(data []byte) *AliasOutputBuilder {
	builder.output.Features.Upsert(&iotago.MetadataFeature{Data: data})
	builder.govCtrlReq = true
	return builder
}

// ImmutableMetadata sets/modifies an iotago.MetadataFeature as an immutable feature on the output.
// Only call this function on a new iotago.AliasOutput.
func (builder *AliasOutputBuilder) ImmutableMetadata(data []byte) *AliasOutputBuilder {
	builder.output.ImmutableFeatures.Upsert(&iotago.MetadataFeature{Data: data})
	return builder
}

// MustBuild works like Build() but panics if an error is encountered.
func (builder *AliasOutputBuilder) MustBuild() *iotago.AliasOutput {
	output, err := builder.Build()
	if err != nil {
		panic(err)
	}
	return output
}

// Build builds the iotago.AliasOutput.
func (builder *AliasOutputBuilder) Build() (*iotago.AliasOutput, error) {
	if builder.prev != nil && builder.govCtrlReq && builder.stateCtrlReq {
		return nil, fmt.Errorf("builder calls require both state and governor transitions which is not possible")
	}

	if builder.stateCtrlReq {
		builder.output.StateIndex++
	}

	if builder.prev != nil {
		if !builder.prev.ImmutableFeatures.Equal(builder.output.ImmutableFeatures) {
			return nil, fmt.Errorf("immutable features are not allowed to be changed")
		}
	}

	builder.output.Conditions.Sort()
	builder.output.Features.Sort()
	builder.output.ImmutableFeatures.Sort()
	builder.output.NativeTokens.Sort()

	return builder.output, nil
}

type aliasStateTransition struct {
	builder *AliasOutputBuilder
}

// StateTransition narrows the builder functions to the ones available for an alias state transition.
func (builder *AliasOutputBuilder) StateTransition() *aliasStateTransition {
	return &aliasStateTransition{builder: builder}
}

// Deposit sets the deposit of the output.
func (trans *aliasStateTransition) Deposit(deposit uint64) *aliasStateTransition {
	return trans.builder.Deposit(deposit).StateTransition()
}

// StateMetadata sets the state metadata of the output.
func (trans *aliasStateTransition) StateMetadata(data []byte) *aliasStateTransition {
	return trans.builder.StateMetadata(data).StateTransition()
}

// FoundriesToGenerate bumps the output's foundry counter by the amount of foundries to generate.
func (trans *aliasStateTransition) FoundriesToGenerate(count uint32) *aliasStateTransition {
	return trans.builder.FoundriesToGenerate(count).StateTransition()
}

// NativeToken adds/modifies a native token to/on the output.
func (trans *aliasStateTransition) NativeToken(nt *iotago.NativeToken) *aliasStateTransition {
	return trans.builder.NativeToken(nt).StateTransition()
}

// NativeTokens sets the native tokens held by the output.
func (trans *aliasStateTransition) NativeTokens(nts iotago.NativeTokens) *aliasStateTransition {
	return trans.builder.NativeTokens(nts).StateTransition()
}

// Sender sets/modifies an iotago.SenderFeature as a mutable feature on the output.
func (trans *aliasStateTransition) Sender(senderAddr iotago.Address) *aliasStateTransition {
	return trans.builder.Sender(senderAddr).StateTransition()
}

// Builder returns the AliasOutputBuilder.
func (trans *aliasStateTransition) Builder() *AliasOutputBuilder {
	return trans.builder
}

type aliasGovernanceTransition struct {
	builder *AliasOutputBuilder
}

// GovernanceTransition narrows the builder functions to the ones available for an alias governance transition.
func (builder *AliasOutputBuilder) GovernanceTransition() *aliasGovernanceTransition {
	return &aliasGovernanceTransition{builder: builder}
}

// StateController sets the iotago.StateControllerAddressUnlockCondition of the output.
func (trans *aliasGovernanceTransition) StateController(stateCtrl iotago.Address) *aliasGovernanceTransition {
	return trans.builder.StateController(stateCtrl).GovernanceTransition()
}

// Governor sets the iotago.GovernorAddressUnlockCondition of the output.
func (trans *aliasGovernanceTransition) Governor(governor iotago.Address) *aliasGovernanceTransition {
	return trans.builder.Governor(governor).GovernanceTransition()
}

// Sender sets/modifies an iotago.SenderFeature as a mutable feature on the output.
func (trans *aliasGovernanceTransition) Sender(senderAddr iotago.Address) *aliasGovernanceTransition {
	return trans.builder.Sender(senderAddr).GovernanceTransition()
}

// Metadata sets/modifies an iotago.MetadataFeature as a mutable feature on the output.
func (trans *aliasGovernanceTransition) Metadata(data []byte) *aliasGovernanceTransition {
	return trans.builder.Metadata(data).GovernanceTransition()
}

// Builder returns the AliasOutputBuilder.
func (trans *aliasGovernanceTransition) Builder() *AliasOutputBuilder {
	return trans.builder
}

// NewFoundryOutputBuilder creates a new FoundryOutputBuilder with the alias address, serial number, token scheme and deposit amount.
func NewFoundryOutputBuilder(aliasAddr *iotago.AliasAddress, tokenScheme iotago.TokenScheme, deposit uint64) *FoundryOutputBuilder {
	return &FoundryOutputBuilder{output: &iotago.FoundryOutput{
		Amount:       deposit,
		TokenScheme:  tokenScheme,
		NativeTokens: iotago.NativeTokens{},
		Conditions: iotago.UnlockConditions[iotago.FoundryUnlockCondition]{
			&iotago.ImmutableAliasUnlockCondition{Address: aliasAddr},
		},
		Features:          iotago.Features[iotago.FoundryFeature]{},
		ImmutableFeatures: iotago.Features[iotago.FoundryImmFeature]{},
	}}
}

// NewFoundryOutputBuilderFromPrevious creates a new FoundryOutputBuilder starting from a copy of the previous iotago.FoundryOutput.
func NewFoundryOutputBuilderFromPrevious(previous *iotago.FoundryOutput) *FoundryOutputBuilder {
	return &FoundryOutputBuilder{
		prev:   previous,
		output: previous.Clone().(*iotago.FoundryOutput),
	}
}

// FoundryOutputBuilder builds an iotago.FoundryOutput.
type FoundryOutputBuilder struct {
	prev   *iotago.FoundryOutput
	output *iotago.FoundryOutput
}

// Deposit sets the deposit of the output.
func (builder *FoundryOutputBuilder) Deposit(deposit uint64) *FoundryOutputBuilder {
	builder.output.Amount = deposit
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
// Only call this function on a new iotago.AliasOutput.
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
			return nil, fmt.Errorf("immutable features are not allowed to be changed")
		}
	}

	builder.output.Conditions.Sort()
	builder.output.Features.Sort()
	builder.output.ImmutableFeatures.Sort()
	builder.output.NativeTokens.Sort()

	return builder.output, nil
}

// NewNFTOutputBuilder creates a new NFTOutputBuilder with the address and deposit amount.
func NewNFTOutputBuilder(targetAddr iotago.Address, deposit uint64) *NFTOutputBuilder {
	return &NFTOutputBuilder{output: &iotago.NFTOutput{
		Amount:       deposit,
		NativeTokens: iotago.NativeTokens{},
		Conditions: iotago.UnlockConditions[iotago.NFTUnlockCondition]{
			&iotago.AddressUnlockCondition{Address: targetAddr},
		},
		Features:          iotago.Features[iotago.NFTFeature]{},
		ImmutableFeatures: iotago.Features[iotago.NFTImmFeature]{},
	}}
}

// NewNFTOutputBuilderFromPrevious creates a new NFTOutputBuilder starting from a copy of the previous iotago.NFTOutput.
func NewNFTOutputBuilderFromPrevious(previous *iotago.NFTOutput) *NFTOutputBuilder {
	return &NFTOutputBuilder{
		prev:   previous,
		output: previous.Clone().(*iotago.NFTOutput),
	}
}

// NFTOutputBuilder builds an iotago.NFTOutput.
type NFTOutputBuilder struct {
	prev   *iotago.NFTOutput
	output *iotago.NFTOutput
}

// Deposit sets the deposit of the output.
func (builder *NFTOutputBuilder) Deposit(deposit uint64) *NFTOutputBuilder {
	builder.output.Amount = deposit
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
func (builder *NFTOutputBuilder) StorageDepositReturn(returnAddr iotago.Address, amount uint64) *NFTOutputBuilder {
	builder.output.Conditions.Upsert(&iotago.StorageDepositReturnUnlockCondition{ReturnAddress: returnAddr, Amount: amount})
	return builder
}

// Timelock sets/modifies an iotago.TimelockUnlockCondition on the output.
func (builder *NFTOutputBuilder) Timelock(untilUnixTimeSec int64) *NFTOutputBuilder {
	builder.output.Conditions.Upsert(&iotago.TimelockUnlockCondition{UnixTime: uint32(untilUnixTimeSec)})
	return builder
}

// Expiration sets/modifies an iotago.ExpirationUnlockCondition on the output.
func (builder *NFTOutputBuilder) Expiration(returnAddr iotago.Address, expiredAfterUnixTimeSec int64) *NFTOutputBuilder {
	builder.output.Conditions.Upsert(&iotago.ExpirationUnlockCondition{ReturnAddress: returnAddr, UnixTime: uint32(expiredAfterUnixTimeSec)})
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
			return nil, fmt.Errorf("immutable features are not allowed to be changed")
		}
	}

	builder.output.Conditions.Sort()
	builder.output.Features.Sort()
	builder.output.ImmutableFeatures.Sort()
	builder.output.NativeTokens.Sort()

	return builder.output, nil
}
