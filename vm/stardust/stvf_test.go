package stardust_test

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/iota.go/v3/vm"
)

var (
	v2API = iotago.V2API(tpkg.TestProtoParas)
)

type fieldMutations map[string]interface{}

func copyObject(t *testing.T, source any, mutations fieldMutations) any {
	srcBytes, err := v2API.Encode(source)
	require.NoError(t, err)

	ptrToCpyOfSrc := reflect.New(reflect.ValueOf(source).Elem().Type())

	cpySeri := ptrToCpyOfSrc.Interface()
	_, err = v2API.Decode(srcBytes, cpySeri)
	require.NoError(t, err)

	for fieldName, newVal := range mutations {
		ptrToCpyOfSrc.Elem().FieldByName(fieldName).Set(reflect.ValueOf(newVal))
	}

	return cpySeri
}

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
		Conditions: iotago.UnlockConditions[iotago.FoundryUnlockCondition]{
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
		svCtx     *vm.Paras
		wantErr   error
	}

	tests := []test{
		{
			name: "ok - genesis transition",
			current: &iotago.AliasOutput{
				Amount:  100,
				AliasID: iotago.AliasID{},
				Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				ImmutableFeatures: iotago.Features[iotago.AliasImmFeature]{
					&iotago.IssuerFeature{Address: exampleIssuer},
				},
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Paras{
				External: &iotago.ExternalUnlockParameters{},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
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
				Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Paras{
				External: &iotago.ExternalUnlockParameters{},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - gov transition",
			current: &iotago.AliasOutput{
				Amount:  100,
				AliasID: exampleAliasID,
				Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
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
				Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: iotago.Features[iotago.AliasFeature]{
					&iotago.SenderFeature{Address: exampleGovCtrl},
					&iotago.MetadataFeature{Data: []byte("1337")},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Paras{
				External: &iotago.ExternalUnlockParameters{},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
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
				Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
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
				Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				StateIndex:     11,
				StateMetadata:  []byte("1337"),
				FoundryCounter: 7,
				Features: iotago.Features[iotago.AliasFeature]{
					&iotago.SenderFeature{Address: exampleStateCtrl},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Paras{
				External: &iotago.ExternalUnlockParameters{},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleStateCtrl.Key(): {UnlockedAt: 0},
					},
					InChains: map[iotago.ChainID]iotago.ChainOutput{
						// serial number 5
						exampleExistingFoundryOutputID: exampleExistingFoundryOutput,
					},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							Inputs: nil,
							Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
								&iotago.FoundryOutput{
									Amount:       100,
									SerialNumber: 6,
									TokenScheme:  &iotago.SimpleTokenScheme{},
									Conditions: iotago.UnlockConditions[iotago.FoundryUnlockCondition]{
										&iotago.ImmutableAliasUnlockCondition{Address: exampleAliasID.ToAddress().(*iotago.AliasAddress)},
									},
								},
								&iotago.FoundryOutput{
									Amount:       100,
									SerialNumber: 7,
									TokenScheme:  &iotago.SimpleTokenScheme{},
									Conditions: iotago.UnlockConditions[iotago.FoundryUnlockCondition]{
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
				Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
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
			svCtx: &vm.Paras{
				External: &iotago.ExternalUnlockParameters{},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
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
				Conditions: iotago.UnlockConditions[iotago.AliasUnlockCondition]{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				ImmutableFeatures: iotago.Features[iotago.AliasImmFeature]{
					&iotago.IssuerFeature{Address: exampleIssuer},
				},
			},
			nextMut: map[string]fieldMutations{
				"state_controller": {
					"StateIndex": uint32(11),
					"Conditions": iotago.UnlockConditions[iotago.AliasUnlockCondition]{
						&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
				},
				"governance_controller": {
					"StateIndex": uint32(11),
					"Conditions": iotago.UnlockConditions[iotago.AliasUnlockCondition]{
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
					"Features": iotago.Features[iotago.AliasFeature]{
						&iotago.MetadataFeature{Data: []byte("foo")},
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Paras{
				External: &iotago.ExternalUnlockParameters{},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					InChains:       map[iotago.ChainID]iotago.ChainOutput{},
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
					cpy := copyObject(t, tt.current, muts).(*iotago.AliasOutput)
					err := stardustVM.ChainSTVF(tt.transType, tt.current, cpy, tt.svCtx)
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
			err := stardustVM.ChainSTVF(tt.transType, tt.current, tt.next, tt.svCtx)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

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
		Conditions: iotago.UnlockConditions[iotago.FoundryUnlockCondition]{
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
		Conditions: iotago.UnlockConditions[iotago.FoundryUnlockCondition]{
			&iotago.ImmutableAliasUnlockCondition{Address: exampleAliasIdent},
		},
	}

	type test struct {
		name      string
		current   *iotago.FoundryOutput
		next      *iotago.FoundryOutput
		nextMut   map[string]fieldMutations
		transType iotago.ChainTransitionType
		svCtx     *vm.Paras
		wantErr   error
	}

	tests := []test{
		{
			name:      "ok - genesis transition",
			current:   exampleFoundry,
			next:      nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Paras{
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							Outputs: iotago.Outputs[iotago.TxEssenceOutput]{exampleFoundry},
						},
						Unlocks: nil,
					},
					InChains: map[iotago.ChainID]iotago.ChainOutput{
						exampleAliasIdent.AliasID(): &iotago.AliasOutput{FoundryCounter: 5},
					},
					OutChains: map[iotago.ChainID]iotago.ChainOutput{
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
			svCtx: &vm.Paras{
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							Outputs: iotago.Outputs[iotago.TxEssenceOutput]{exampleFoundry},
						},
						Unlocks: nil,
					},
					InChains: map[iotago.ChainID]iotago.ChainOutput{
						exampleAliasIdent.AliasID(): &iotago.AliasOutput{FoundryCounter: 5},
					},
					OutChains: map[iotago.ChainID]iotago.ChainOutput{
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
			svCtx: &vm.Paras{
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							Outputs: iotago.Outputs[iotago.TxEssenceOutput]{exampleFoundry},
						},
						Unlocks: nil,
					},
					InChains: map[iotago.ChainID]iotago.ChainOutput{
						exampleAliasIdent.AliasID(): &iotago.AliasOutput{FoundryCounter: 6},
					},
					OutChains: map[iotago.ChainID]iotago.ChainOutput{
						exampleAliasIdent.AliasID(): &iotago.AliasOutput{FoundryCounter: 7},
					},
					OutNativeTokens: map[iotago.NativeTokenID]*big.Int{},
				},
			},
			wantErr: &iotago.ChainTransitionError{},
		},
		{
			name:      "fail - genesis transition - foundries unsorted",
			current:   exampleFoundry,
			next:      nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Paras{
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
								&iotago.FoundryOutput{
									Amount: 100,
									// exampleFoundry has serial number 6
									SerialNumber: 7,
									TokenScheme: &iotago.SimpleTokenScheme{
										MintedTokens:  startingSupply,
										MeltedTokens:  big.NewInt(0),
										MaximumSupply: new(big.Int).SetUint64(1000),
									},
									Conditions: iotago.UnlockConditions[iotago.FoundryUnlockCondition]{
										&iotago.ImmutableAliasUnlockCondition{Address: exampleAliasIdent},
									},
								},
								exampleFoundry,
							},
						},
						Unlocks: nil,
					},
					InChains: map[iotago.ChainID]iotago.ChainOutput{
						exampleAliasIdent.AliasID(): &iotago.AliasOutput{FoundryCounter: 5},
					},
					OutChains: map[iotago.ChainID]iotago.ChainOutput{
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
					"Features": iotago.Features[iotago.FoundryFeature]{
						&iotago.MetadataFeature{Data: tpkg.RandBytes(20)},
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Paras{
				WorkingSet: &vm.WorkingSet{
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
			svCtx: &vm.Paras{
				WorkingSet: &vm.WorkingSet{
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
			svCtx: &vm.Paras{
				WorkingSet: &vm.WorkingSet{
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
			svCtx: &vm.Paras{
				WorkingSet: &vm.WorkingSet{
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
			svCtx: &vm.Paras{
				WorkingSet: &vm.WorkingSet{
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
			svCtx: &vm.Paras{
				WorkingSet: &vm.WorkingSet{
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
			svCtx: &vm.Paras{
				WorkingSet: &vm.WorkingSet{
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
			svCtx: &vm.Paras{
				WorkingSet: &vm.WorkingSet{
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
			svCtx: &vm.Paras{
				WorkingSet: &vm.WorkingSet{},
			},
			wantErr: &iotago.ChainTransitionError{},
		},
		{
			name:      "ok - destroy transition",
			current:   toBeDestoyedFoundry,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Paras{
				WorkingSet: &vm.WorkingSet{
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
			svCtx: &vm.Paras{
				WorkingSet: &vm.WorkingSet{
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
					err := stardustVM.ChainSTVF(tt.transType, tt.current, cpy, tt.svCtx)
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
			err := stardustVM.ChainSTVF(tt.transType, tt.current, tt.next, tt.svCtx)
			if tt.wantErr != nil {
				require.ErrorAs(t, err, &tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestNFTOutput_ValidateStateTransition(t *testing.T) {
	exampleIssuer := tpkg.RandEd25519Address()

	exampleCurrentNFTOutput := &iotago.NFTOutput{
		Amount: 100,
		NFTID:  iotago.NFTID{},
		Conditions: iotago.UnlockConditions[iotago.NFTUnlockCondition]{
			&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
		},
		ImmutableFeatures: iotago.Features[iotago.NFTImmFeature]{
			&iotago.IssuerFeature{Address: exampleIssuer},
			&iotago.MetadataFeature{Data: []byte("some-ipfs-link")},
		},
	}

	type test struct {
		name      string
		current   *iotago.NFTOutput
		next      *iotago.NFTOutput
		nextMut   map[string]fieldMutations
		transType iotago.ChainTransitionType
		svCtx     *vm.Paras
		wantErr   error
	}

	tests := []test{
		{
			name:      "ok - genesis transition",
			current:   exampleCurrentNFTOutput,
			next:      nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Paras{
				External: &iotago.ExternalUnlockParameters{},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
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
			svCtx: &vm.Paras{
				External: &iotago.ExternalUnlockParameters{},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
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
					"Conditions": iotago.UnlockConditions[iotago.NFTUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				"native_tokens": {
					"NativeTokens": tpkg.RandSortNativeTokens(10),
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Paras{
				External: &iotago.ExternalUnlockParameters{},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
				},
			},
			wantErr: nil,
		},
		{
			name:    "fail - state transition",
			current: exampleCurrentNFTOutput,
			nextMut: map[string]fieldMutations{
				"immutable_metadata": {
					"ImmutableFeatures": iotago.Features[iotago.NFTImmFeature]{
						&iotago.MetadataFeature{Data: []byte("link-to-cat.gif")},
					},
				},
				"issuer": {
					"ImmutableFeatures": iotago.Features[iotago.NFTImmFeature]{
						&iotago.IssuerFeature{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Paras{
				External: &iotago.ExternalUnlockParameters{},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
				},
			},
			wantErr: &iotago.ChainTransitionError{},
		},
	}

	for _, tt := range tests {

		if tt.nextMut != nil {
			for mutName, muts := range tt.nextMut {
				t.Run(fmt.Sprintf("%s_%s", tt.name, mutName), func(t *testing.T) {
					cpy := copyObject(t, tt.current, muts).(*iotago.NFTOutput)
					err := stardustVM.ChainSTVF(tt.transType, tt.current, cpy, tt.svCtx)
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
			err := stardustVM.ChainSTVF(tt.transType, tt.current, tt.next, tt.svCtx)
			if tt.wantErr != nil {
				require.ErrorAs(t, err, &tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
