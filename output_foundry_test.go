package iotago_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/stretchr/testify/require"
)

func TestFoundryOutput_ValidateStateTransition(t *testing.T) {
	exampleIssuer := tpkg.RandEd25519Address()
	exampleAliasIdent := tpkg.RandAliasAddress()

	genesisFoundry := &iotago.FoundryOutput{
		Address:           exampleAliasIdent,
		Amount:            100,
		SerialNumber:      6,
		TokenTag:          tpkg.Rand12ByteArray(),
		CirculatingSupply: new(big.Int).SetUint64(100),
		MaximumSupply:     new(big.Int).SetUint64(1000),
		TokenScheme:       &iotago.SimpleTokenScheme{},
		Blocks: iotago.FeatureBlocks{
			&iotago.IssuerFeatureBlock{Address: exampleIssuer},
		},
	}

	type test struct {
		name      string
		current   *iotago.FoundryOutput
		next      *iotago.FoundryOutput
		nextMut   map[string]fieldMutations
		transType iotago.ChainTransitionType
		svCtx     *iotago.SemanticValidationContext
		wantErr   error
	}

	tests := []test{
		{
			name:      "ok - genesis transition",
			current:   genesisFoundry,
			next:      nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &iotago.SemanticValidationContext{
				ExtParas: &iotago.ExternalUnlockParameters{},
				WorkingSet: &iotago.SemValiContextWorkingSet{
					UnlockedIdents: map[string]iotago.UnlockedIndices{},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							Outputs: iotago.Outputs{
								genesisFoundry,
							},
						},
						UnlockBlocks: nil,
					},
					InChains: map[iotago.ChainID]iotago.ChainConstrainedOutput{
						exampleAliasIdent.AliasID(): &iotago.AliasOutput{
							FoundryCounter: 5,
						},
					},
					OutChains: map[iotago.ChainID]iotago.ChainConstrainedOutput{
						exampleAliasIdent.AliasID(): &iotago.AliasOutput{
							FoundryCounter: 6,
						},
					},
					OutNativeTokens: map[iotago.NativeTokenID]*big.Int{
						genesisFoundry.MustNativeTokenID(): new(big.Int).SetUint64(100),
					},
				},
			},
			wantErr: nil,
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
