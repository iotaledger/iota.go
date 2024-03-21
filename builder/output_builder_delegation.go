package builder

import (
	iotago "github.com/iotaledger/iota.go/v4"
)

// NewDelegationOutputBuilder creates a new DelegationOutputBuilder with the account address, serial number, token scheme and base token amount.
func NewDelegationOutputBuilder(validatorAddress *iotago.AccountAddress, addr iotago.Address, amount iotago.BaseToken) *DelegationOutputBuilder {
	return &DelegationOutputBuilder{output: &iotago.DelegationOutput{
		Amount:           amount,
		DelegatedAmount:  0,
		DelegationID:     iotago.DelegationID{},
		ValidatorAddress: validatorAddress,
		StartEpoch:       0,
		EndEpoch:         0,
		UnlockConditions: iotago.DelegationOutputUnlockConditions{
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
	builder.output.UnlockConditions.Upsert(&iotago.AddressUnlockCondition{Address: addr})

	return builder
}

// Build builds the iotago.DelegationOutput.
func (builder *DelegationOutputBuilder) Build() (*iotago.DelegationOutput, error) {
	builder.output.UnlockConditions.Sort()

	return builder.output, nil
}

// MustBuild works like Build() but panics if an error is encountered.
func (builder *DelegationOutputBuilder) MustBuild() *iotago.DelegationOutput {
	output, err := builder.Build()
	if err != nil {
		panic(err)
	}

	return output
}
