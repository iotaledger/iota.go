package iotago_test

import (
	"math/big"
	"testing"

	"github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/stretchr/testify/require"
)

func TestAliasOutput_ValidateStateTransition(t *testing.T) {
	exampleIssuer := tpkg.RandEd25519Address()
	exampleAliasID := tpkg.RandAliasAddress().AliasID()

	exampleStateCtrl := tpkg.RandEd25519Address()
	exampleGovCtrl := tpkg.RandEd25519Address()

	exampleExistingFoundryOutput := &iotago.FoundryOutput{
		Address:           exampleAliasID.ToAddress(),
		Amount:            100,
		SerialNumber:      5,
		TokenTag:          iotago.TokenTag{},
		CirculatingSupply: new(big.Int).SetInt64(1000),
		MaximumSupply:     new(big.Int).SetInt64(10000),
		TokenScheme:       &iotago.SimpleTokenScheme{},
	}
	exampleExistingFoundryOutputID := exampleExistingFoundryOutput.MustID()

	type test struct {
		name      string
		current   *iotago.AliasOutput
		next      *iotago.AliasOutput
		transType iotago.ChainTransitionType
		svCtx     *iotago.SemanticValidationContext
		wantErr   error
	}
	tests := []test{
		{
			name: "ok - genesis transition",
			current: &iotago.AliasOutput{
				Amount:               100,
				AliasID:              iotago.AliasID{},
				StateController:      tpkg.RandEd25519Address(),
				GovernanceController: tpkg.RandEd25519Address(),
				Blocks: iotago.FeatureBlocks{
					&iotago.IssuerFeatureBlock{Address: exampleIssuer},
				},
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &iotago.SemanticValidationContext{
				ExtParas: &iotago.ExternalUnlockParameters{},
				WorkingSet: &iotago.SemValiContextWorkingSet{
					UnlockedIdents: map[string]iotago.UnlockedIndices{
						exampleIssuer.Key(): {0: {}},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - destroy transition",
			current: &iotago.AliasOutput{
				Amount:               100,
				AliasID:              tpkg.RandAliasAddress().AliasID(),
				StateController:      exampleStateCtrl,
				GovernanceController: exampleGovCtrl,
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &iotago.SemanticValidationContext{
				ExtParas: &iotago.ExternalUnlockParameters{},
				WorkingSet: &iotago.SemValiContextWorkingSet{
					UnlockedIdents: map[string]iotago.UnlockedIndices{},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - gov transition",
			current: &iotago.AliasOutput{
				Amount:               100,
				AliasID:              exampleAliasID,
				StateController:      exampleStateCtrl,
				GovernanceController: exampleGovCtrl,
				StateIndex:           10,
			},
			next: &iotago.AliasOutput{
				Amount:  100,
				AliasID: exampleAliasID,
				// mutating controllers
				StateController:      tpkg.RandEd25519Address(),
				GovernanceController: tpkg.RandEd25519Address(),
				StateIndex:           10,
				Blocks: iotago.FeatureBlocks{
					&iotago.MetadataFeatureBlock{Data: []byte("1337")},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &iotago.SemanticValidationContext{
				ExtParas: &iotago.ExternalUnlockParameters{},
				WorkingSet: &iotago.SemValiContextWorkingSet{
					UnlockedIdents: map[string]iotago.UnlockedIndices{},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - state transition",
			current: &iotago.AliasOutput{
				Amount:               100,
				AliasID:              exampleAliasID,
				StateController:      exampleStateCtrl,
				GovernanceController: exampleGovCtrl,
				StateIndex:           10,
				FoundryCounter:       5,
			},
			next: &iotago.AliasOutput{
				Amount:               200,
				NativeTokens:         tpkg.RandSortNativeTokens(50),
				AliasID:              exampleAliasID,
				StateController:      exampleStateCtrl,
				GovernanceController: exampleGovCtrl,
				StateIndex:           11,
				StateMetadata:        []byte("1337"),
				FoundryCounter:       7,
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &iotago.SemanticValidationContext{
				ExtParas: &iotago.ExternalUnlockParameters{},
				WorkingSet: &iotago.SemValiContextWorkingSet{
					UnlockedIdents: map[string]iotago.UnlockedIndices{},
					InChains: map[iotago.ChainID]iotago.ChainConstrainedOutput{
						// serial number 5
						exampleExistingFoundryOutputID: exampleExistingFoundryOutput,
					},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							Inputs: nil,
							Outputs: iotago.Outputs{
								&iotago.FoundryOutput{
									Address:           exampleAliasID.ToAddress(),
									Amount:            100,
									SerialNumber:      6,
									TokenTag:          tpkg.Rand12ByteArray(),
									CirculatingSupply: new(big.Int).SetInt64(1000),
									MaximumSupply:     new(big.Int).SetInt64(10000),
									TokenScheme:       &iotago.SimpleTokenScheme{},
								},
								&iotago.FoundryOutput{
									Address:           exampleAliasID.ToAddress(),
									Amount:            100,
									SerialNumber:      7,
									TokenTag:          tpkg.Rand12ByteArray(),
									CirculatingSupply: new(big.Int).SetInt64(1000),
									MaximumSupply:     new(big.Int).SetInt64(10000),
									TokenScheme:       &iotago.SimpleTokenScheme{},
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
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
