package iotago_test

import (
	"fmt"
	"testing"

	"github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/stretchr/testify/require"
)

func TestNFTOutput_ValidateStateTransition(t *testing.T) {
	exampleIssuer := tpkg.RandEd25519Address()

	exampleCurrentNFTOutput := &iotago.NFTOutput{
		Address:           tpkg.RandEd25519Address(),
		Amount:            100,
		NFTID:             iotago.NFTID{},
		ImmutableMetadata: []byte("some-ipfs-link"),
		Blocks: iotago.FeatureBlocks{
			&iotago.IssuerFeatureBlock{Address: exampleIssuer},
		},
	}

	type test struct {
		name      string
		current   *iotago.NFTOutput
		next      *iotago.NFTOutput
		nextMut   map[string]fieldMutations
		transType iotago.ChainTransitionType
		svCtx     *iotago.SemanticValidationContext
		wantErr   error
	}

	tests := []test{
		{
			name:      "ok - genesis transition",
			current:   exampleCurrentNFTOutput,
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
			name:      "ok - destroy transition",
			current:   exampleCurrentNFTOutput,
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
			name:    "ok - state transition",
			current: exampleCurrentNFTOutput,
			nextMut: map[string]fieldMutations{
				"amount": {
					"Amount": uint64(1337),
				},
				"address": {
					"Address": tpkg.RandEd25519Address(),
				},
				"native_tokens": {
					"NativeTokens": tpkg.RandSortNativeTokens(10),
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
			name:    "fail - state transition",
			current: exampleCurrentNFTOutput,
			nextMut: map[string]fieldMutations{
				"immutable_metadata": {
					"ImmutableMetadata": []byte("link-to-cat.gif"),
				},
				"issuer": {
					"Blocks": iotago.FeatureBlocks{
						&iotago.IssuerFeatureBlock{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &iotago.SemanticValidationContext{
				ExtParas: &iotago.ExternalUnlockParameters{},
				WorkingSet: &iotago.SemValiContextWorkingSet{
					UnlockedIdents: map[string]iotago.UnlockedIndices{},
				},
			},
			wantErr: iotago.ErrInvalidChainStateTransition,
		},
	}

	for _, tt := range tests {

		if tt.nextMut != nil {
			for mutName, muts := range tt.nextMut {
				t.Run(fmt.Sprintf("%s_%s", tt.name, mutName), func(t *testing.T) {
					cpy := copyObject(t, tt.current, muts).(*iotago.NFTOutput)
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