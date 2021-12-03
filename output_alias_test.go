package iotago_test

import (
	"testing"

	"github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

func TestAliasOutput_ValidateStateTransition(t *testing.T) {
	exampleIssuer := tpkg.RandEd25519Address()

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
			name: "ok - genesis transaction",
			current: &iotago.AliasOutput{
				Amount:               100,
				NativeTokens:         nil,
				AliasID:              iotago.AliasID{},
				StateController:      tpkg.RandEd25519Address(),
				GovernanceController: tpkg.RandEd25519Address(),
				StateIndex:           0,
				StateMetadata:        nil,
				FoundryCounter:       0,
				Blocks: iotago.FeatureBlocks{
					&iotago.IssuerFeatureBlock{Address: exampleIssuer},
				},
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &iotago.SemanticValidationContext{
				ExtParas: &iotago.ExternalUnlockParameters{},
				WorkingSet: &iotago.SemValiContextWorkingSet{
					UnlockedIdents:     nil,
					InputSet:           nil,
					InputIDToIndex:     nil,
					Tx:                 nil,
					EssenceMsgToSign:   nil,
					InputsByType:       nil,
					InChains:           nil,
					InNativeTokens:     nil,
					OutputsByType:      nil,
					OutChains:          nil,
					OutNativeTokens:    nil,
					UnlockBlocksByType: nil,
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			/*
			err := tt.current.ValidateStateTransition(tt.transType, tt.next, tt.svCtx)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			 */
		})
	}
}
