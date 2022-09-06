package iotago_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

func TestAliasOutput_ValidateStateTransition(t *testing.T) {
	exampleIssuer := tpkg.RandEd25519Address()
	exampleAliasID := tpkg.RandAliasAddress().AliasID()

	exampleStateCtrl := tpkg.RandEd25519Address()
	exampleGovCtrl := tpkg.RandEd25519Address()

	exampleExistingFoundryOutput := &iotago.FoundryOutput{
		Amount:       100,
		SerialNumber: 5,
		TokenScheme: &iotago.SimpleTokenScheme{
			MintedTokens:  new(big.Int).SetInt64(1000),
			MeltedTokens:  big.NewInt(0),
			MaximumSupply: new(big.Int).SetInt64(10000),
		},
		Conditions: iotago.UnlockConditions{
			&iotago.ImmutableAliasUnlockCondition{Address: exampleAliasID.ToAddress().(*iotago.AliasAddress)},
		},
	}
	exampleExistingFoundryOutputID := exampleExistingFoundryOutput.MustID()

	type test struct {
		name      string
		current   *iotago.AliasOutput
		next      *iotago.AliasOutput
		nextMut   map[string]fieldMutations
		transType iotago.ChainTransitionType
		svCtx     *iotago.SemanticValidationContext
		wantErr   error
	}

	tests := []test{
		{
			name: "ok - genesis transition",
			current: &iotago.AliasOutput{
				Amount:  100,
				AliasID: iotago.AliasID{},
				Conditions: iotago.UnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				ImmutableFeatures: iotago.Features{
					&iotago.IssuerFeature{Address: exampleIssuer},
				},
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &iotago.SemanticValidationContext{
				ExtParas: &iotago.ExternalUnlockParameters{},
				WorkingSet: &iotago.SemValiContextWorkingSet{
					UnlockedIdents: iotago.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - destroy transition",
			current: &iotago.AliasOutput{
				Amount:  100,
				AliasID: tpkg.RandAliasAddress().AliasID(),
				Conditions: iotago.UnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &iotago.SemanticValidationContext{
				ExtParas: &iotago.ExternalUnlockParameters{},
				WorkingSet: &iotago.SemValiContextWorkingSet{
					UnlockedIdents: iotago.UnlockedIdentities{},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - gov transition",
			current: &iotago.AliasOutput{
				Amount:  100,
				AliasID: exampleAliasID,
				Conditions: iotago.UnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				StateIndex: 10,
			},
			next: &iotago.AliasOutput{
				Amount:     100,
				AliasID:    exampleAliasID,
				StateIndex: 10,
				// mutating controllers
				Conditions: iotago.UnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: iotago.Features{
					&iotago.SenderFeature{Address: exampleGovCtrl},
					&iotago.MetadataFeature{Data: []byte("1337")},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &iotago.SemanticValidationContext{
				ExtParas: &iotago.ExternalUnlockParameters{},
				WorkingSet: &iotago.SemValiContextWorkingSet{
					UnlockedIdents: iotago.UnlockedIdentities{
						exampleGovCtrl.Key(): {UnlockedAt: 0},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - state transition",
			current: &iotago.AliasOutput{
				Amount:  100,
				AliasID: exampleAliasID,
				Conditions: iotago.UnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				StateIndex:     10,
				FoundryCounter: 5,
			},
			next: &iotago.AliasOutput{
				Amount:       200,
				NativeTokens: tpkg.RandSortNativeTokens(50),
				AliasID:      exampleAliasID,
				Conditions: iotago.UnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				StateIndex:     11,
				StateMetadata:  []byte("1337"),
				FoundryCounter: 7,
				Features: iotago.Features{
					&iotago.SenderFeature{Address: exampleStateCtrl},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &iotago.SemanticValidationContext{
				ExtParas: &iotago.ExternalUnlockParameters{},
				WorkingSet: &iotago.SemValiContextWorkingSet{
					UnlockedIdents: iotago.UnlockedIdentities{
						exampleStateCtrl.Key(): {UnlockedAt: 0},
					},
					InChains: map[iotago.ChainID]iotago.ChainConstrainedOutput{
						// serial number 5
						exampleExistingFoundryOutputID: exampleExistingFoundryOutput,
					},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							Inputs: nil,
							Outputs: iotago.Outputs{
								&iotago.FoundryOutput{
									Amount:       100,
									SerialNumber: 6,
									TokenScheme:  &iotago.SimpleTokenScheme{},
									Conditions: iotago.UnlockConditions{
										&iotago.ImmutableAliasUnlockCondition{Address: exampleAliasID.ToAddress().(*iotago.AliasAddress)},
									},
								},
								&iotago.FoundryOutput{
									Amount:       100,
									SerialNumber: 7,
									TokenScheme:  &iotago.SimpleTokenScheme{},
									Conditions: iotago.UnlockConditions{
										&iotago.ImmutableAliasUnlockCondition{Address: exampleAliasID.ToAddress().(*iotago.AliasAddress)},
									},
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - gov transition",
			current: &iotago.AliasOutput{
				Amount:     100,
				AliasID:    exampleAliasID,
				StateIndex: 10,
				Conditions: iotago.UnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
			},
			nextMut: map[string]fieldMutations{
				"amount": {
					"Amount": uint64(1337),
				},
				"native_tokens": {
					"NativeTokens": tpkg.RandSortNativeTokens(10),
				},
				"state_metadata": {
					"StateMetadata": []byte("7331"),
				},
				"foundry_counter": {
					"FoundryCounter": uint32(1337),
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &iotago.SemanticValidationContext{
				ExtParas: &iotago.ExternalUnlockParameters{},
				WorkingSet: &iotago.SemValiContextWorkingSet{
					UnlockedIdents: iotago.UnlockedIdentities{},
				},
			},
			wantErr: iotago.ErrInvalidAliasGovernanceTransition,
		},
		{
			name: "fail - state transition",
			current: &iotago.AliasOutput{
				Amount:         100,
				AliasID:        exampleAliasID,
				StateIndex:     10,
				FoundryCounter: 5,
				Conditions: iotago.UnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				ImmutableFeatures: iotago.Features{
					&iotago.IssuerFeature{Address: exampleIssuer},
				},
			},
			nextMut: map[string]fieldMutations{
				"state_controller": {
					"StateIndex": uint32(11),
					"Conditions": iotago.UnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
				},
				"governance_controller": {
					"StateIndex": uint32(11),
					"Conditions": iotago.UnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				"foundry_counter_lower_than_current": {
					"FoundryCounter": uint32(4),
					"StateIndex":     uint32(11),
				},
				"state_index_lower": {
					"StateIndex": uint32(4),
				},
				"state_index_bigger_more_than_1": {
					"StateIndex": uint32(7),
				},
				"foundries_not_created": {
					"StateIndex":     uint32(11),
					"FoundryCounter": uint32(7),
				},
				"metadata_feature_added": {
					"StateIndex": uint32(11),
					"Features": iotago.Features{
						&iotago.MetadataFeature{Data: []byte("foo")},
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &iotago.SemanticValidationContext{
				ExtParas: &iotago.ExternalUnlockParameters{},
				WorkingSet: &iotago.SemValiContextWorkingSet{
					UnlockedIdents: iotago.UnlockedIdentities{},
					InChains:       map[iotago.ChainID]iotago.ChainConstrainedOutput{},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{},
					},
				},
			},
			wantErr: iotago.ErrInvalidAliasStateTransition,
		},
	}

	for _, tt := range tests {

		if tt.nextMut != nil {
			for mutName, muts := range tt.nextMut {
				t.Run(fmt.Sprintf("%s_%s", tt.name, mutName), func(t *testing.T) {
					cpy := copyObject(t, tt.current, muts, tpkg.TestProtoParas).(*iotago.AliasOutput)
					err := tt.current.ValidateStateTransition(tt.transType, cpy, tt.svCtx)
					if tt.wantErr != nil {
						require.ErrorIs(t, err, tt.wantErr)
						return
					}
					require.NoError(t, err)
				})
			}
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			err := tt.current.ValidateStateTransition(tt.transType, tt.next, tt.svCtx)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
