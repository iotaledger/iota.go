package iotago_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

func TestFoundryOutput_ValidateStateTransition(t *testing.T) {
	exampleAliasIdent := tpkg.RandAliasAddress()

	startingSupply := new(big.Int).SetUint64(100)
	exampleFoundry := &iotago.FoundryOutput{
		Amount:       100,
		SerialNumber: 6,
		TokenScheme: &iotago.SimpleTokenScheme{
			MintedTokens:  startingSupply,
			MeltedTokens:  big.NewInt(0),
			MaximumSupply: new(big.Int).SetUint64(1000),
		},
		Conditions: iotago.UnlockConditions{
			&iotago.ImmutableAliasUnlockCondition{Address: exampleAliasIdent},
		},
	}

	toBeDestoyedFoundry := &iotago.FoundryOutput{
		Amount:       100,
		SerialNumber: 6,
		TokenScheme: &iotago.SimpleTokenScheme{
			MintedTokens:  startingSupply,
			MeltedTokens:  startingSupply,
			MaximumSupply: new(big.Int).SetUint64(1000),
		},
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
			current:   exampleFoundry,
			next:      nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &iotago.SemanticValidationContext{
				WorkingSet: &iotago.SemValiContextWorkingSet{
					UnlockedIdents: iotago.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							Outputs: iotago.Outputs{exampleFoundry},
						},
						Unlocks: nil,
					},
					InChains: map[iotago.ChainID]iotago.ChainConstrainedOutput{
						exampleAliasIdent.AliasID(): &iotago.AliasOutput{FoundryCounter: 5},
					},
					OutChains: map[iotago.ChainID]iotago.ChainConstrainedOutput{
						exampleAliasIdent.AliasID(): &iotago.AliasOutput{FoundryCounter: 6},
					},
					OutNativeTokens: map[iotago.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): startingSupply,
					},
				},
			},
			wantErr: nil,
		},
		{
			name:      "fail - genesis transition - mint supply not equal to out",
			current:   exampleFoundry,
			next:      nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &iotago.SemanticValidationContext{
				WorkingSet: &iotago.SemValiContextWorkingSet{
					UnlockedIdents: iotago.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							Outputs: iotago.Outputs{exampleFoundry},
						},
						Unlocks: nil,
					},
					InChains: map[iotago.ChainID]iotago.ChainConstrainedOutput{
						exampleAliasIdent.AliasID(): &iotago.AliasOutput{FoundryCounter: 5},
					},
					OutChains: map[iotago.ChainID]iotago.ChainConstrainedOutput{
						exampleAliasIdent.AliasID(): &iotago.AliasOutput{FoundryCounter: 6},
					},
					OutNativeTokens: map[iotago.NativeTokenID]*big.Int{
						// absent but should be there
					},
				},
			},
			wantErr: &iotago.ChainTransitionError{},
		},
		{
			name:      "fail - genesis transition - serial number not in interval",
			current:   exampleFoundry,
			next:      nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &iotago.SemanticValidationContext{
				WorkingSet: &iotago.SemValiContextWorkingSet{
					UnlockedIdents: iotago.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							Outputs: iotago.Outputs{exampleFoundry},
						},
						Unlocks: nil,
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
			wantErr: &iotago.ChainTransitionError{},
		},
		{
			name:      "fail - genesis transition - foundries unsorted",
			current:   exampleFoundry,
			next:      nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &iotago.SemanticValidationContext{
				WorkingSet: &iotago.SemValiContextWorkingSet{
					UnlockedIdents: iotago.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							Outputs: iotago.Outputs{
								&iotago.FoundryOutput{
									Amount: 100,
									// exampleFoundry has serial number 6
									SerialNumber: 7,
									TokenScheme: &iotago.SimpleTokenScheme{
										MintedTokens:  startingSupply,
										MeltedTokens:  big.NewInt(0),
										MaximumSupply: new(big.Int).SetUint64(1000),
									},
									Conditions: iotago.UnlockConditions{
										&iotago.ImmutableAliasUnlockCondition{Address: exampleAliasIdent},
									},
								},
								exampleFoundry,
							},
						},
						Unlocks: nil,
					},
					InChains: map[iotago.ChainID]iotago.ChainConstrainedOutput{
						exampleAliasIdent.AliasID(): &iotago.AliasOutput{FoundryCounter: 5},
					},
					OutChains: map[iotago.ChainID]iotago.ChainConstrainedOutput{
						exampleAliasIdent.AliasID(): &iotago.AliasOutput{FoundryCounter: 7},
					},
					OutNativeTokens: map[iotago.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): startingSupply,
					},
				},
			},
			wantErr: &iotago.ChainTransitionError{},
		},

		{
			name:    "ok - state transition - metadata feature",
			current: exampleFoundry,
			nextMut: map[string]fieldMutations{
				"change_metadata": {
					"Features": iotago.Features{
						&iotago.MetadataFeature{Data: tpkg.RandBytes(20)},
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
					"TokenScheme": &iotago.SimpleTokenScheme{
						MintedTokens:  big.NewInt(400),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
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
					"TokenScheme": &iotago.SimpleTokenScheme{
						MintedTokens:  big.NewInt(100),
						MeltedTokens:  big.NewInt(50),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
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
					"TokenScheme": &iotago.SimpleTokenScheme{
						MintedTokens:  big.NewInt(100),
						MeltedTokens:  big.NewInt(100),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
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
					"TokenScheme": &iotago.SimpleTokenScheme{
						MintedTokens:  big.NewInt(200),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
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
					"TokenScheme": &iotago.SimpleTokenScheme{
						MintedTokens:  big.NewInt(200),
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
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
					"TokenScheme": &iotago.SimpleTokenScheme{
						MintedTokens:  big.NewInt(100),
						MeltedTokens:  big.NewInt(50),
						MaximumSupply: new(big.Int).SetUint64(1000),
					},
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
					"TokenScheme": &iotago.SimpleTokenScheme{
						MintedTokens:  startingSupply,
						MeltedTokens:  big.NewInt(0),
						MaximumSupply: big.NewInt(1337),
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &iotago.SemanticValidationContext{
				WorkingSet: &iotago.SemValiContextWorkingSet{},
			},
			wantErr: &iotago.ChainTransitionError{},
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
					cpy := copyObject(t, tt.current, muts, tpkg.TestProtoParas).(*iotago.FoundryOutput)
					err := tt.current.ValidateStateTransition(tt.transType, cpy, tt.svCtx)
					if tt.wantErr != nil {
						require.ErrorAs(t, err, &tt.wantErr)
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
				require.ErrorAs(t, err, &tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
