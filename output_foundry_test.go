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
	exampleAliasIdent := tpkg.RandAliasAddress()

	startingSupply := new(big.Int).SetUint64(100)

	genesisFoundry := &iotago.FoundryOutput{
		Amount:        100,
		SerialNumber:  6,
		TokenTag:      tpkg.Rand12ByteArray(),
		MintedTokens:  big.NewInt(0),
		MeltedTokens:  big.NewInt(0),
		MaximumSupply: new(big.Int).SetUint64(1000),
		TokenScheme:   &iotago.SimpleTokenScheme{},
		Conditions: iotago.UnlockConditions{
			&iotago.ImmutableAliasUnlockCondition{Address: exampleAliasIdent},
		},
	}

	exampleFoundry := &iotago.FoundryOutput{
		Amount:        100,
		SerialNumber:  6,
		TokenTag:      tpkg.Rand12ByteArray(),
		MintedTokens:  startingSupply,
		MeltedTokens:  big.NewInt(0),
		MaximumSupply: new(big.Int).SetUint64(1000),
		TokenScheme:   &iotago.SimpleTokenScheme{},
		Conditions: iotago.UnlockConditions{
			&iotago.ImmutableAliasUnlockCondition{Address: exampleAliasIdent},
		},
	}

	toBeDestoyedFoundry := &iotago.FoundryOutput{
		Amount:        100,
		SerialNumber:  6,
		TokenTag:      tpkg.Rand12ByteArray(),
		MintedTokens:  startingSupply,
		MeltedTokens:  startingSupply,
		MaximumSupply: new(big.Int).SetUint64(1000),
		TokenScheme:   &iotago.SimpleTokenScheme{},
		Conditions: iotago.UnlockConditions{
			&iotago.ImmutableAliasUnlockCondition{Address: exampleAliasIdent},
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
				WorkingSet: &iotago.SemValiContextWorkingSet{
					UnlockedIdents: map[string]iotago.UnlockedIndices{},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							Outputs: iotago.Outputs{genesisFoundry},
						},
						UnlockBlocks: nil,
					},
					InChains: map[iotago.ChainID]iotago.ChainConstrainedOutput{
						exampleAliasIdent.AliasID(): &iotago.AliasOutput{FoundryCounter: 5},
					},
					OutChains: map[iotago.ChainID]iotago.ChainConstrainedOutput{
						exampleAliasIdent.AliasID(): &iotago.AliasOutput{FoundryCounter: 6},
					},
					OutNativeTokens: map[iotago.NativeTokenID]*big.Int{},
				},
			},
			wantErr: nil,
		},
		{
			name:      "fail - genesis transition - serial number not in interval",
			current:   exampleFoundry,
			next:      nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &iotago.SemanticValidationContext{
				WorkingSet: &iotago.SemValiContextWorkingSet{
					UnlockedIdents: map[string]iotago.UnlockedIndices{},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							Outputs: iotago.Outputs{exampleFoundry},
						},
						UnlockBlocks: nil,
					},
					InChains: map[iotago.ChainID]iotago.ChainConstrainedOutput{
						exampleAliasIdent.AliasID(): &iotago.AliasOutput{FoundryCounter: 6},
					},
					OutChains: map[iotago.ChainID]iotago.ChainConstrainedOutput{
						exampleAliasIdent.AliasID(): &iotago.AliasOutput{FoundryCounter: 7},
					},
					OutNativeTokens: map[iotago.NativeTokenID]*big.Int{},
				},
			},
			wantErr: iotago.ErrInvalidChainStateTransition,
		},
		{
			name:      "fail - genesis transition - non mint supply",
			current:   exampleFoundry,
			next:      nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &iotago.SemanticValidationContext{
				WorkingSet: &iotago.SemValiContextWorkingSet{
					UnlockedIdents: map[string]iotago.UnlockedIndices{},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							Outputs: iotago.Outputs{exampleFoundry},
						},
						UnlockBlocks: nil,
					},
					InChains: map[iotago.ChainID]iotago.ChainConstrainedOutput{
						exampleAliasIdent.AliasID(): &iotago.AliasOutput{FoundryCounter: 6},
					},
					OutChains: map[iotago.ChainID]iotago.ChainConstrainedOutput{
						exampleAliasIdent.AliasID(): &iotago.AliasOutput{FoundryCounter: 7},
					},
					OutNativeTokens: map[iotago.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): startingSupply,
					},
				},
			},
			wantErr: iotago.ErrInvalidChainStateTransition,
		},
		{
			name:    "ok - state transition - metadata feature block",
			current: exampleFoundry,
			nextMut: map[string]fieldMutations{
				"change_metadata": {
					"Blocks": iotago.FeatureBlocks{
						&iotago.MetadataFeatureBlock{Data: tpkg.RandBytes(20)},
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &iotago.SemanticValidationContext{
				WorkingSet: &iotago.SemValiContextWorkingSet{
					OutNativeTokens: map[iotago.NativeTokenID]*big.Int{},
				},
			},
			wantErr: nil,
		},
		{
			name:    "ok - state transition - mint",
			current: exampleFoundry,
			nextMut: map[string]fieldMutations{
				"+300": {
					"MintedTokens": big.NewInt(400),
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &iotago.SemanticValidationContext{
				WorkingSet: &iotago.SemValiContextWorkingSet{
					OutNativeTokens: map[iotago.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): new(big.Int).SetUint64(300),
					},
				},
			},
			wantErr: nil,
		},
		{
			name:    "ok - state transition - melt",
			current: exampleFoundry,
			nextMut: map[string]fieldMutations{
				"-50": {
					"MintedTokens": big.NewInt(100),
					"MeltedTokens": big.NewInt(50),
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &iotago.SemanticValidationContext{
				WorkingSet: &iotago.SemValiContextWorkingSet{
					InNativeTokens: map[iotago.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): startingSupply,
					},
					OutNativeTokens: map[iotago.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): new(big.Int).SetUint64(50),
					},
				},
			},
			wantErr: nil,
		},
		{
			name:      "ok - state transition - burn",
			current:   exampleFoundry,
			nextMut:   map[string]fieldMutations{},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &iotago.SemanticValidationContext{
				WorkingSet: &iotago.SemValiContextWorkingSet{
					InNativeTokens: map[iotago.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): startingSupply,
					},
					OutNativeTokens: map[iotago.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): new(big.Int).SetUint64(50),
					},
				},
			},
			wantErr: nil,
		},
		{
			name:    "ok - state transition - melt complete supply",
			current: exampleFoundry,
			nextMut: map[string]fieldMutations{
				"-100": {
					"MintedTokens": big.NewInt(100),
					"MeltedTokens": big.NewInt(100),
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &iotago.SemanticValidationContext{
				WorkingSet: &iotago.SemValiContextWorkingSet{
					InNativeTokens: map[iotago.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): startingSupply,
					},
					OutNativeTokens: map[iotago.NativeTokenID]*big.Int{},
				},
			},
			wantErr: nil,
		},
		{
			name:    "fail - state transition - mint (out: excess)",
			current: exampleFoundry,
			nextMut: map[string]fieldMutations{
				"+100": {
					"MintedTokens": big.NewInt(200),
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &iotago.SemanticValidationContext{
				WorkingSet: &iotago.SemValiContextWorkingSet{
					OutNativeTokens: map[iotago.NativeTokenID]*big.Int{
						// 100 excess
						exampleFoundry.MustNativeTokenID(): new(big.Int).SetUint64(200),
					},
				},
			},
			wantErr: iotago.ErrNativeTokenSumUnbalanced,
		},
		{
			name:    "fail - state transition - mint (out: deficit)",
			current: exampleFoundry,
			nextMut: map[string]fieldMutations{
				"+100": {
					"MintedTokens": big.NewInt(200),
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &iotago.SemanticValidationContext{
				WorkingSet: &iotago.SemValiContextWorkingSet{
					OutNativeTokens: map[iotago.NativeTokenID]*big.Int{
						// 50 deficit
						exampleFoundry.MustNativeTokenID(): new(big.Int).SetUint64(50),
					},
				},
			},
			wantErr: iotago.ErrNativeTokenSumUnbalanced,
		},
		{
			name:    "fail - state transition - melt (out: excess)",
			current: exampleFoundry,
			nextMut: map[string]fieldMutations{
				"-50": {
					"MintedTokens": big.NewInt(100),
					"MeltedTokens": big.NewInt(50),
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &iotago.SemanticValidationContext{
				WorkingSet: &iotago.SemValiContextWorkingSet{
					InNativeTokens: map[iotago.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): startingSupply,
					},
					OutNativeTokens: map[iotago.NativeTokenID]*big.Int{
						// 25 excess
						exampleFoundry.MustNativeTokenID(): new(big.Int).SetUint64(75),
					},
				},
			},
			wantErr: iotago.ErrNativeTokenSumUnbalanced,
		},
		{
			name:    "fail - state transition",
			current: exampleFoundry,
			nextMut: map[string]fieldMutations{
				"maximum_supply": {
					"MaximumSupply": big.NewInt(1337),
				},
				"token_tag": {
					"TokenTag": tpkg.Rand12ByteArray(),
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &iotago.SemanticValidationContext{
				WorkingSet: &iotago.SemValiContextWorkingSet{},
			},
			wantErr: iotago.ErrInvalidChainStateTransition,
		},
		{
			name:      "ok - destroy transition",
			current:   toBeDestoyedFoundry,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &iotago.SemanticValidationContext{
				WorkingSet: &iotago.SemValiContextWorkingSet{
					InNativeTokens:  map[iotago.NativeTokenID]*big.Int{},
					OutNativeTokens: map[iotago.NativeTokenID]*big.Int{},
				},
			},
			wantErr: nil,
		},
		{
			name:      "fail - destroy transition - foundry token unbalanced",
			current:   exampleFoundry,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &iotago.SemanticValidationContext{
				WorkingSet: &iotago.SemValiContextWorkingSet{
					InNativeTokens: map[iotago.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): startingSupply,
					},
					OutNativeTokens: map[iotago.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): new(big.Int).Mul(startingSupply, new(big.Int).SetUint64(2)),
					},
				},
			},
			wantErr: iotago.ErrNativeTokenSumUnbalanced,
		},
	}

	for _, tt := range tests {

		if tt.nextMut != nil {
			for mutName, muts := range tt.nextMut {
				t.Run(fmt.Sprintf("%s_%s", tt.name, mutName), func(t *testing.T) {
					cpy := copyObject(t, tt.current, muts).(*iotago.FoundryOutput)
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
