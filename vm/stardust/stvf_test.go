//nolint:forcetypeassert,dupl,nlreturn,scopelint
package stardust_test

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
	"github.com/iotaledger/iota.go/v4/vm"
)

type fieldMutations map[string]interface{}

//nolint:thelper
func copyObject(t *testing.T, source any, mutations fieldMutations) any {
	srcBytes, err := tpkg.TestAPI.Encode(source)
	require.NoError(t, err)

	ptrToCpyOfSrc := reflect.New(reflect.ValueOf(source).Elem().Type())

	cpySeri := ptrToCpyOfSrc.Interface()
	_, err = tpkg.TestAPI.Decode(srcBytes, cpySeri)
	require.NoError(t, err)

	for fieldName, newVal := range mutations {
		ptrToCpyOfSrc.Elem().FieldByName(fieldName).Set(reflect.ValueOf(newVal))
	}

	return cpySeri
}

func TestAccountOutput_ValidateStateTransition(t *testing.T) {
	exampleIssuer := tpkg.RandEd25519Address()
	exampleAccountID := tpkg.RandAccountAddress().AccountID()

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
		Conditions: iotago.FoundryOutputUnlockConditions{
			&iotago.ImmutableAccountUnlockCondition{Address: exampleAccountID.ToAddress().(*iotago.AccountAddress)},
		},
	}
	exampleExistingFoundryOutputID := exampleExistingFoundryOutput.MustID()

	currentEpoch := iotago.EpochIndex(20)
	currentSlot := tpkg.TestAPI.TimeProvider().EpochStart(currentEpoch)

	pubkey := iotago.Ed25519PublicKeyBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray())
	exampleBlockIssuerFeature := &iotago.BlockIssuerFeature{
		BlockIssuerKeys: iotago.BlockIssuerKeys{pubkey},
		ExpirySlot:      currentSlot + tpkg.TestAPI.ProtocolParameters().MaxCommittableAge(),
	}

	exampleBIC := map[iotago.AccountID]iotago.BlockIssuanceCredits{
		exampleAccountID: 100,
	}

	type test struct {
		name      string
		input     *vm.ChainOutputWithCreationSlot
		next      *iotago.AccountOutput
		nextMut   map[string]fieldMutations
		transType iotago.ChainTransitionType
		svCtx     *vm.Params
		wantErr   error
	}

	tests := []test{
		{
			name: "ok - genesis transition",
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: iotago.AccountID{},
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				ImmutableFeatures: iotago.AccountOutputImmFeatures{
					&iotago.IssuerFeature{Address: exampleIssuer},
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - block issuer genesis transition",
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: iotago.AccountID{},
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				ImmutableFeatures: iotago.AccountOutputImmFeatures{
					&iotago.IssuerFeature{Address: exampleIssuer},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
						ExpirySlot:      1000,
					},
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: 900,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							CreationSlot: 900,
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - block issuer genesis expiry too early",
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: iotago.AccountID{},
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				ImmutableFeatures: iotago.AccountOutputImmFeatures{
					&iotago.IssuerFeature{Address: exampleIssuer},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
						ExpirySlot:      1000,
					},
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: 10001,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							CreationSlot: 10001,
						},
					},
				},
			},
			wantErr: iotago.ErrInvalidBlockIssuerTransition,
		},
		{
			name: "fail - block issuer genesis expired but within MCA",
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: iotago.AccountID{},
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				ImmutableFeatures: iotago.AccountOutputImmFeatures{
					&iotago.IssuerFeature{Address: exampleIssuer},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
						ExpirySlot:      1000,
					},
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: 991,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							CreationSlot: 991,
						},
					},
				},
			},
			wantErr: iotago.ErrInvalidBlockIssuerTransition,
		},
		{
			name: "ok - staking genesis transition",
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: iotago.AccountID{},
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.StakingFeature{
						StakedAmount: 50,
						FixedCost:    5,
						StartEpoch:   currentEpoch,
						EndEpoch:     math.MaxUint32,
					},
					exampleBlockIssuerFeature,
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - staking genesis start epoch invalid",
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: iotago.AccountID{},
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.StakingFeature{
						StakedAmount: 50,
						FixedCost:    5,
						StartEpoch:   currentEpoch - 2,
						EndEpoch:     math.MaxUint32,
					},
					exampleBlockIssuerFeature,
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: iotago.ErrInvalidStakingStartEpoch,
		},
		{
			name: "fail - staking genesis end epoch too early",
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: iotago.AccountID{},
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.StakingFeature{
						StakedAmount: 50,
						FixedCost:    5,
						StartEpoch:   currentEpoch,
						EndEpoch:     currentEpoch + tpkg.TestAPI.ProtocolParameters().StakingUnbondingPeriod() - 1,
					},
					exampleBlockIssuerFeature,
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: iotago.ErrInvalidStakingEndEpochTooEarly,
		},
		{
			name: "fail - staking genesis staked amount higher than amount",
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: iotago.AccountID{},
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.StakingFeature{
						StakedAmount: 500,
						FixedCost:    5,
						StartEpoch:   currentEpoch,
						EndEpoch:     currentEpoch + tpkg.TestAPI.ProtocolParameters().StakingUnbondingPeriod(),
					},
					exampleBlockIssuerFeature,
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: iotago.ErrInvalidStakingAmountMismatch,
		},
		{
			name: "fail - staking feature without block issuer feature",
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: iotago.AccountID{},
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.StakingFeature{
						StakedAmount: 50,
						FixedCost:    5,
						StartEpoch:   currentEpoch,
						EndEpoch:     math.MaxUint32,
					},
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: iotago.ErrInvalidStakingBlockIssuerRequired,
		},
		{
			name: "ok - valid staking transition",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:     100,
					AccountID:  exampleAccountID,
					StateIndex: 50,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.StakingFeature{
							StakedAmount: 100,
							FixedCost:    50,
							StartEpoch:   currentEpoch,
							EndEpoch:     currentEpoch + 10000,
						},
						exampleBlockIssuerFeature,
					},
				},
			},
			next: &iotago.AccountOutput{
				Amount:     100,
				AccountID:  exampleAccountID,
				StateIndex: 51,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.StakingFeature{
						StakedAmount: 100,
						FixedCost:    50,
						StartEpoch:   currentEpoch,
						EndEpoch:     currentEpoch + tpkg.TestAPI.ProtocolParameters().StakingUnbondingPeriod(),
					},
					exampleBlockIssuerFeature,
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
						},
					},
					BIC: exampleBIC,
				},
			},
		},
		{
			name: "ok - adding staking feature in account state transition",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:     100,
					AccountID:  exampleAccountID,
					StateIndex: 50,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						exampleBlockIssuerFeature,
					},
				},
			},
			next: &iotago.AccountOutput{
				Amount:     100,
				AccountID:  exampleAccountID,
				StateIndex: 51,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.StakingFeature{
						StakedAmount: 100,
						FixedCost:    50,
						StartEpoch:   currentEpoch,
						EndEpoch:     currentEpoch + tpkg.TestAPI.ProtocolParameters().StakingUnbondingPeriod(),
					},
					exampleBlockIssuerFeature,
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					BIC: exampleBIC,
				},
			},
		},
		{
			name: "fail - adding staking feature in account state transition with start epoch set too early",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:     100,
					AccountID:  exampleAccountID,
					StateIndex: 50,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						exampleBlockIssuerFeature,
					},
				},
			},
			next: &iotago.AccountOutput{
				Amount:     100,
				AccountID:  exampleAccountID,
				StateIndex: 51,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.StakingFeature{
						StakedAmount: 100,
						FixedCost:    50,
						StartEpoch:   currentEpoch - 5,
						EndEpoch:     currentEpoch + tpkg.TestAPI.ProtocolParameters().StakingUnbondingPeriod(),
					},
					exampleBlockIssuerFeature,
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: iotago.ErrInvalidStakingStartEpoch,
		},
		{
			name: "fail - negative BIC during account state transition",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:     100,
					AccountID:  exampleAccountID,
					StateIndex: 50,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						exampleBlockIssuerFeature,
					},
				},
			},
			next: &iotago.AccountOutput{
				Amount:     100,
				AccountID:  exampleAccountID,
				StateIndex: 51,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				StateMetadata: []byte("1337"),
				Features: iotago.AccountOutputFeatures{
					exampleBlockIssuerFeature,
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
						},
					},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredits{
						exampleAccountID: -100,
					},
				},
			},
			wantErr: iotago.ErrAccountLocked,
		},
		{
			name: "fail - removing staking feature before end epoch",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:     100,
					AccountID:  exampleAccountID,
					StateIndex: 50,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.StakingFeature{
							StakedAmount: 100,
							FixedCost:    50,
							StartEpoch:   currentEpoch,
							EndEpoch:     currentEpoch + 10000,
						},
						exampleBlockIssuerFeature,
					},
				},
			},
			next: &iotago.AccountOutput{
				Amount:     100,
				AccountID:  exampleAccountID,
				StateIndex: 51,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				Features: iotago.AccountOutputFeatures{
					exampleBlockIssuerFeature,
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: iotago.ErrInvalidStakingBondedRemoval,
		},
		{
			name: "fail - changing staking feature's staked amount",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:     100,
					AccountID:  exampleAccountID,
					StateIndex: 50,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.StakingFeature{
							StakedAmount: 100,
							FixedCost:    50,
							StartEpoch:   currentEpoch,
							EndEpoch:     currentEpoch + 10000,
						},
						exampleBlockIssuerFeature,
					},
				},
			},
			next: &iotago.AccountOutput{
				Amount:     100,
				AccountID:  exampleAccountID,
				StateIndex: 51,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.StakingFeature{
						StakedAmount: 90,
						FixedCost:    50,
						StartEpoch:   currentEpoch,
						EndEpoch:     currentEpoch + 10000,
					},
					exampleBlockIssuerFeature,
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: iotago.ErrInvalidStakingBondedModified,
		},
		{
			name: "fail - reducing staking feature's end epoch by more than the unbonding period",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:     100,
					AccountID:  exampleAccountID,
					StateIndex: 50,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.StakingFeature{
							StakedAmount: 100,
							FixedCost:    50,
							StartEpoch:   currentEpoch,
							EndEpoch:     currentEpoch + 10000,
						},
						exampleBlockIssuerFeature,
					},
				},
			},
			next: &iotago.AccountOutput{
				Amount:     100,
				AccountID:  exampleAccountID,
				StateIndex: 51,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.StakingFeature{
						StakedAmount: 100,
						FixedCost:    50,
						StartEpoch:   currentEpoch,
						EndEpoch:     currentEpoch + tpkg.TestAPI.ProtocolParameters().StakingUnbondingPeriod() - 5,
					},
					exampleBlockIssuerFeature,
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: iotago.ErrInvalidStakingEndEpochTooEarly,
		},
		{
			name: "fail - account removes block issuer feature while having a staking feature",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:     100,
					AccountID:  exampleAccountID,
					StateIndex: 1,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.StakingFeature{
							StakedAmount: 50,
							FixedCost:    5,
							StartEpoch:   currentEpoch,
							EndEpoch:     math.MaxUint32,
						},
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
							ExpirySlot:      990,
						},
					},
				},
				CreationSlot: 1000,
			},
			next: &iotago.AccountOutput{
				Amount:     100,
				AccountID:  exampleAccountID,
				StateIndex: 1,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.StakingFeature{
						StakedAmount: 50,
						FixedCost:    5,
						StartEpoch:   currentEpoch,
						EndEpoch:     math.MaxUint32,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: iotago.ErrInvalidStakingBlockIssuerRequired,
		},
		{
			name: "fail - account removes staking feature in governance transition",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:     100,
					AccountID:  exampleAccountID,
					StateIndex: 1,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.StakingFeature{
							StakedAmount: 50,
							FixedCost:    5,
							StartEpoch:   currentEpoch,
							EndEpoch:     math.MaxUint32,
						},
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
							ExpirySlot:      990,
						},
					},
				},
				CreationSlot: 1000,
			},
			next: &iotago.AccountOutput{
				Amount:     100,
				AccountID:  exampleAccountID,
				StateIndex: 1,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				Features: iotago.AccountOutputFeatures{},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: iotago.ErrInvalidAccountGovernanceTransition,
		},
		{
			name: "fail - account removes block issuer feature in staking transition",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:     100,
					AccountID:  exampleAccountID,
					StateIndex: 1,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
							ExpirySlot:      990,
						},
					},
				},
				CreationSlot: 1000,
			},
			next: &iotago.AccountOutput{
				Amount:     100,
				AccountID:  exampleAccountID,
				StateIndex: 2,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				Features: iotago.AccountOutputFeatures{},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: iotago.ErrInvalidAccountStateTransition,
		},
		{
			name: "fail - expired staking feature removed without specifying reward input",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:     100,
					AccountID:  exampleAccountID,
					StateIndex: 1,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.StakingFeature{
							StakedAmount: 50,
							FixedCost:    5,
							StartEpoch:   currentEpoch - 10,
							EndEpoch:     currentEpoch - 5,
						},
						exampleBlockIssuerFeature,
					},
				},
				CreationSlot: 1000,
			},
			next: &iotago.AccountOutput{
				Amount:     100,
				AccountID:  exampleAccountID,
				StateIndex: 2,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				Features: iotago.AccountOutputFeatures{
					exampleBlockIssuerFeature,
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: iotago.ErrInvalidStakingRewardInputRequired,
		},
		{
			name: "fail - changing an expired staking feature without claiming",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:     100,
					AccountID:  exampleAccountID,
					StateIndex: 1,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.StakingFeature{
							StakedAmount: 50,
							FixedCost:    5,
							StartEpoch:   currentEpoch - 10,
							EndEpoch:     currentEpoch - 5,
						},
						exampleBlockIssuerFeature,
					},
				},
				CreationSlot: 1000,
			},
			next: &iotago.AccountOutput{
				Amount:     100,
				AccountID:  exampleAccountID,
				StateIndex: 2,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.StakingFeature{
						StakedAmount: 80,
						FixedCost:    5,
						StartEpoch:   currentEpoch - 10,
						EndEpoch:     currentEpoch - 5,
					},
					exampleBlockIssuerFeature,
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: iotago.ErrInvalidStakingRewardInputRequired,
		},
		{
			name: "fail - claiming rewards of an expired staking feature without resetting start epoch",
			input: &vm.ChainOutputWithCreationSlot{
				ChainID: exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:     100,
					AccountID:  exampleAccountID,
					StateIndex: 1,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.StakingFeature{
							StakedAmount: 50,
							FixedCost:    5,
							StartEpoch:   currentEpoch - 10,
							EndEpoch:     currentEpoch - 5,
						},
						exampleBlockIssuerFeature,
					},
				},
				CreationSlot: 1000,
			},
			next: &iotago.AccountOutput{
				Amount:     100,
				AccountID:  exampleAccountID,
				StateIndex: 2,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.StakingFeature{
						StakedAmount: 50,
						FixedCost:    5,
						StartEpoch:   currentEpoch - 10,
						EndEpoch:     currentEpoch + 10,
					},
					exampleBlockIssuerFeature,
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
						},
					},
					BIC: exampleBIC,
					Rewards: map[iotago.ChainID]iotago.Mana{
						exampleAccountID: 200,
					},
				},
			},
			wantErr: iotago.ErrInvalidStakingStartEpoch,
		},
		{
			name: "ok - destroy transition",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: tpkg.RandAccountAddress().AccountID(),
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
				},
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - destroy block issuer account with negative BIC",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
							ExpirySlot:      1000,
						},
					},
				},
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: 1001,
					},
					UnlockedIdents: vm.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							CreationSlot: 1001,
						},
					},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredits{
						exampleAccountID: -1,
					},
				},
			},
			wantErr: iotago.ErrAccountLocked,
		},
		{
			name: "fail - destroy block issuer account no BIC provided",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
							ExpirySlot:      1000,
						},
					},
				},
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: 1001,
					},
					UnlockedIdents: vm.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							CreationSlot: 1001,
						},
					},
				},
			},
			wantErr: iotago.ErrBlockIssuanceCreditInputRequired,
		},

		{
			name: "fail - non-expired block issuer destroy transition",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: tpkg.RandAccountAddress().AccountID(),
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
							ExpirySlot:      1000,
						},
					},
				},
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: 1000,
					},
					UnlockedIdents: vm.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							CreationSlot: 1000,
						},
					},
				},
			},
			wantErr: iotago.ErrInvalidBlockIssuerTransition,
		},
		{
			name: "ok - expired block issuer destroy transition",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
							ExpirySlot:      1000,
						},
					},
				},
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: 1001,
					},
					UnlockedIdents: vm.UnlockedIdentities{},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredits{
						exampleAccountID: 0,
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							Inputs:       nil,
							CreationSlot: 1001,
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "failed - remove non-expired block issuer feature transition",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
							ExpirySlot:      1000,
						},
					},
					StateIndex: 10,
				},
			},
			next: &iotago.AccountOutput{
				Amount:     100,
				AccountID:  exampleAccountID,
				StateIndex: 10,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				Features: iotago.AccountOutputFeatures{},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: 999,
					},
					UnlockedIdents: vm.UnlockedIdentities{},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredits{
						exampleAccountID: 0,
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							Inputs:       nil,
							CreationSlot: 999,
						},
					},
				},
			},
			wantErr: iotago.ErrInvalidBlockIssuerTransition,
		},
		{
			name: "ok - remove expired block issuer feature transition",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
							ExpirySlot:      1000,
						},
					},
					StateIndex: 10,
				},
			},
			next: &iotago.AccountOutput{
				Amount:     100,
				AccountID:  exampleAccountID,
				StateIndex: 10,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				Features: iotago.AccountOutputFeatures{},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: 1001,
					},
					UnlockedIdents: vm.UnlockedIdentities{},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredits{
						exampleAccountID: 0,
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							Inputs:       nil,
							CreationSlot: 1001,
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - gov transition",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					StateIndex: 10,
				},
			},
			next: &iotago.AccountOutput{
				Amount:     100,
				AccountID:  exampleAccountID,
				StateIndex: 10,
				// mutating controllers
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.SenderFeature{Address: exampleGovCtrl},
					&iotago.MetadataFeature{Data: []byte("1337")},
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
						ExpirySlot:      1015,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleGovCtrl.Key(): {UnlockedAt: 0},
					},
					Commitment: &iotago.Commitment{
						Index: 990,
					},
					BIC: exampleBIC,
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							CreationSlot: 900,
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - state transition",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					StateIndex:     10,
					FoundryCounter: 5,
				},
			},
			next: &iotago.AccountOutput{
				Amount:       200,
				NativeTokens: tpkg.RandSortNativeTokens(50),
				AccountID:    exampleAccountID,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				StateIndex:     11,
				StateMetadata:  []byte("1337"),
				FoundryCounter: 7,
				Features: iotago.AccountOutputFeatures{
					&iotago.SenderFeature{Address: exampleStateCtrl},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleStateCtrl.Key(): {UnlockedAt: 0},
					},
					InChains: map[iotago.ChainID]*vm.ChainOutputWithCreationSlot{
						// serial number 5
						exampleExistingFoundryOutputID: {
							Output: exampleExistingFoundryOutput,
						},
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							Inputs: nil,
							Outputs: iotago.TxEssenceOutputs{
								&iotago.FoundryOutput{
									Amount:       100,
									SerialNumber: 6,
									TokenScheme:  &iotago.SimpleTokenScheme{},
									Conditions: iotago.FoundryOutputUnlockConditions{
										&iotago.ImmutableAccountUnlockCondition{Address: exampleAccountID.ToAddress().(*iotago.AccountAddress)},
									},
								},
								&iotago.FoundryOutput{
									Amount:       100,
									SerialNumber: 7,
									TokenScheme:  &iotago.SimpleTokenScheme{},
									Conditions: iotago.FoundryOutputUnlockConditions{
										&iotago.ImmutableAccountUnlockCondition{Address: exampleAccountID.ToAddress().(*iotago.AccountAddress)},
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
			name: "ok - update expired block issuer feature without extending expiration after MCA",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
							ExpirySlot:      1000,
						},
					},
					StateIndex: 10,
				},
			},
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				StateIndex: 10,
				Features: iotago.AccountOutputFeatures{
					&iotago.SenderFeature{Address: exampleStateCtrl},
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
						ExpirySlot:      1000,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: 990,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleStateCtrl.Key(): {UnlockedAt: 0},
					},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredits{
						exampleAccountID: 10,
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							Inputs:       nil,
							CreationSlot: 990,
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - update account immutable features",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					ImmutableFeatures: iotago.AccountOutputImmFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
							ExpirySlot:      900,
						},
					},
					StateIndex: 10,
				},
			},
			next: &iotago.AccountOutput{
				Amount:    200,
				AccountID: exampleAccountID,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				StateIndex:    11,
				StateMetadata: []byte("1337"),
				ImmutableFeatures: iotago.AccountOutputImmFeatures{
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
						ExpirySlot:      999,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleStateCtrl.Key(): {UnlockedAt: 0},
					},
				},
			},
			wantErr: iotago.ErrInvalidAccountStateTransition,
		},
		{
			name: "fail - update expired block issuer feature with extending expiration before MCA",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
							ExpirySlot:      900,
						},
					},
					StateIndex: 10,
				},
			},
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				StateIndex: 10,
				Features: iotago.AccountOutputFeatures{
					&iotago.SenderFeature{Address: exampleStateCtrl},
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
						ExpirySlot:      999,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: 990,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleStateCtrl.Key(): {UnlockedAt: 0},
					},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredits{
						exampleAccountID: 10,
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							Inputs:       nil,
							CreationSlot: 990,
						},
					},
				},
			},
			wantErr: iotago.ErrInvalidBlockIssuerTransition,
		},
		{
			name: "fail - update expired block issuer feature with extending expiration to the past before MCA",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
							ExpirySlot:      1100,
						},
					},
					StateIndex: 10,
				},
			},
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				StateIndex: 10,
				Features: iotago.AccountOutputFeatures{
					&iotago.SenderFeature{Address: exampleStateCtrl},
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
						ExpirySlot:      999,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: 990,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleStateCtrl.Key(): {UnlockedAt: 0},
					},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredits{
						exampleAccountID: 10,
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							Inputs:       nil,
							CreationSlot: 990,
						},
					},
				},
			},
			wantErr: iotago.ErrInvalidBlockIssuerTransition,
		},
		{
			name: "fail - update block issuer account with negative BIC",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
							ExpirySlot:      1000,
						},
					},
					StateIndex: 10,
				},
			},
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				StateIndex: 10,
				Features: iotago.AccountOutputFeatures{
					&iotago.MetadataFeature{Data: []byte("1337")},
					&iotago.SenderFeature{Address: exampleStateCtrl},
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
						ExpirySlot:      1000,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: 900,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleStateCtrl.Key(): {UnlockedAt: 0},
					},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredits{
						exampleAccountID: -1,
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							Inputs:       nil,
							CreationSlot: 900,
						},
					},
				},
			},
			wantErr: iotago.ErrAccountLocked,
		},
		{
			name: "fail - update block issuer account without BIC provided",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
							ExpirySlot:      1000,
						},
					},
					StateIndex: 10,
				},
			},
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				StateIndex: 10,
				Features: iotago.AccountOutputFeatures{
					&iotago.SenderFeature{Address: exampleStateCtrl},
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
						ExpirySlot:      1000,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: 900,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleStateCtrl.Key(): {UnlockedAt: 0},
					},

					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							CreationSlot: 900,
						},
					},
				},
			},
			wantErr: iotago.ErrBlockIssuanceCreditInputRequired,
		},
		{
			name: "ok - update block issuer feature expiration to earlier slot",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
							ExpirySlot:      1000,
						},
					},
					StateIndex: 10,
				},
			},
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				StateIndex: 10,
				Features: iotago.AccountOutputFeatures{
					&iotago.MetadataFeature{Data: []byte("1337")},
					&iotago.SenderFeature{Address: exampleStateCtrl},
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
						ExpirySlot:      999,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: 900,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleStateCtrl.Key(): {UnlockedAt: 0},
					},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredits{
						exampleAccountID: 10,
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							CreationSlot: 900,
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - non-expired block issuer replace key",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
							ExpirySlot:      1000,
						},
					},
					StateIndex:     10,
					FoundryCounter: 5,
				},
			},
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				Conditions: iotago.AccountOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				StateIndex:     10,
				FoundryCounter: 5,
				Features: iotago.AccountOutputFeatures{
					&iotago.SenderFeature{Address: exampleStateCtrl},
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandomBlockIsssuerKeysEd25519(1),
						ExpirySlot:      1000,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Index: 0,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleStateCtrl.Key(): {UnlockedAt: 0},
					},
					InChains: map[iotago.ChainID]*vm.ChainOutputWithCreationSlot{
						// serial number 5
						exampleExistingFoundryOutputID: {
							Output: exampleExistingFoundryOutput,
						},
					},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredits{
						exampleAccountID: 10,
					},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							Inputs: nil,
							Outputs: iotago.TxEssenceOutputs{
								&iotago.FoundryOutput{
									Amount:       100,
									SerialNumber: 6,
									TokenScheme:  &iotago.SimpleTokenScheme{},
									Conditions: iotago.FoundryOutputUnlockConditions{
										&iotago.ImmutableAccountUnlockCondition{Address: exampleAccountID.ToAddress().(*iotago.AccountAddress)},
									},
								},
								&iotago.FoundryOutput{
									Amount:       100,
									SerialNumber: 7,
									TokenScheme:  &iotago.SimpleTokenScheme{},
									Conditions: iotago.FoundryOutputUnlockConditions{
										&iotago.ImmutableAccountUnlockCondition{Address: exampleAccountID.ToAddress().(*iotago.AccountAddress)},
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
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:     100,
					AccountID:  exampleAccountID,
					StateIndex: 10,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
				},
			},
			nextMut: map[string]fieldMutations{
				"amount": {
					"Amount": iotago.BaseToken(1337),
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
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
				},
			},
			wantErr: iotago.ErrInvalidAccountGovernanceTransition,
		},
		{
			name: "fail - state transition",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.AccountOutput{
					Amount:         100,
					AccountID:      exampleAccountID,
					StateIndex:     10,
					FoundryCounter: 5,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					ImmutableFeatures: iotago.AccountOutputImmFeatures{
						&iotago.IssuerFeature{Address: exampleIssuer},
					},
				},
			},
			nextMut: map[string]fieldMutations{
				"state_controller": {
					"StateIndex": uint32(11),
					"Conditions": iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
				},
				"governance_controller": {
					"StateIndex": uint32(11),
					"Conditions": iotago.AccountOutputUnlockConditions{
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
					"Features": iotago.AccountOutputFeatures{
						&iotago.MetadataFeature{Data: []byte("foo")},
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					InChains:       vm.ChainInputSet{},
					Tx: &iotago.Transaction{
						API:     tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{},
					},
				},
			},
			wantErr: iotago.ErrInvalidAccountStateTransition,
		},
	}

	for _, tt := range tests {

		if tt.nextMut != nil {
			for mutName, muts := range tt.nextMut {
				t.Run(fmt.Sprintf("%s_%s", tt.name, mutName), func(t *testing.T) {
					cpy := copyObject(t, tt.input.Output, muts).(*iotago.AccountOutput)

					if tt.input != nil {
						// create the working set for the test
						if tt.svCtx.WorkingSet.UTXOInputsWithCreationSlot == nil {
							tt.svCtx.WorkingSet.UTXOInputsWithCreationSlot = make(vm.InputSet)
						}

						tt.svCtx.WorkingSet.UTXOInputsWithCreationSlot[tpkg.RandOutputID(0)] = vm.OutputWithCreationSlot{
							Output:       tt.input.Output,
							CreationSlot: tt.input.CreationSlot,
						}
					}

					err := stardustVM.ChainSTVF(tt.transType, tt.input, cpy, tt.svCtx)
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
			if tt.input != nil {
				// create the working set for the test
				if tt.svCtx.WorkingSet.UTXOInputsWithCreationSlot == nil {
					tt.svCtx.WorkingSet.UTXOInputsWithCreationSlot = make(vm.InputSet)
				}

				tt.svCtx.WorkingSet.UTXOInputsWithCreationSlot[tpkg.RandOutputID(0)] = vm.OutputWithCreationSlot{
					Output:       tt.input.Output,
					CreationSlot: tt.input.CreationSlot,
				}
			}

			err := stardustVM.ChainSTVF(tt.transType, tt.input, tt.next, tt.svCtx)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestFoundryOutput_ValidateStateTransition(t *testing.T) {
	exampleAccountIdent := tpkg.RandAccountAddress()

	startingSupply := new(big.Int).SetUint64(100)
	exampleFoundry := &iotago.FoundryOutput{
		Amount:       100,
		SerialNumber: 6,
		TokenScheme: &iotago.SimpleTokenScheme{
			MintedTokens:  startingSupply,
			MeltedTokens:  big.NewInt(0),
			MaximumSupply: new(big.Int).SetUint64(1000),
		},
		Conditions: iotago.FoundryOutputUnlockConditions{
			&iotago.ImmutableAccountUnlockCondition{Address: exampleAccountIdent},
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
		Conditions: iotago.FoundryOutputUnlockConditions{
			&iotago.ImmutableAccountUnlockCondition{Address: exampleAccountIdent},
		},
	}

	type test struct {
		name      string
		input     *vm.ChainOutputWithCreationSlot
		next      *iotago.FoundryOutput
		nextMut   map[string]fieldMutations
		transType iotago.ChainTransitionType
		svCtx     *vm.Params
		wantErr   error
	}

	tests := []test{
		{
			name:      "ok - genesis transition",
			next:      exampleFoundry,
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							Outputs: iotago.TxEssenceOutputs{exampleFoundry},
						},
						Unlocks: nil,
					},
					InChains: vm.ChainInputSet{
						exampleAccountIdent.AccountID(): &vm.ChainOutputWithCreationSlot{
							Output: &iotago.AccountOutput{FoundryCounter: 5},
						},
					},
					OutChains: map[iotago.ChainID]iotago.ChainOutput{
						exampleAccountIdent.AccountID(): &iotago.AccountOutput{FoundryCounter: 6},
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
			next:      exampleFoundry,
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							Outputs: iotago.TxEssenceOutputs{exampleFoundry},
						},
						Unlocks: nil,
					},
					InChains: vm.ChainInputSet{
						exampleAccountIdent.AccountID(): &vm.ChainOutputWithCreationSlot{
							Output: &iotago.AccountOutput{FoundryCounter: 5},
						},
					},
					OutChains: map[iotago.ChainID]iotago.ChainOutput{
						exampleAccountIdent.AccountID(): &iotago.AccountOutput{FoundryCounter: 6},
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
			next:      exampleFoundry,
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							Outputs: iotago.TxEssenceOutputs{exampleFoundry},
						},
						Unlocks: nil,
					},
					InChains: vm.ChainInputSet{
						exampleAccountIdent.AccountID(): &vm.ChainOutputWithCreationSlot{
							Output: &iotago.AccountOutput{FoundryCounter: 6},
						},
					},
					OutChains: map[iotago.ChainID]iotago.ChainOutput{
						exampleAccountIdent.AccountID(): &iotago.AccountOutput{FoundryCounter: 7},
					},
					OutNativeTokens: map[iotago.NativeTokenID]*big.Int{},
				},
			},
			wantErr: &iotago.ChainTransitionError{},
		},
		{
			name:      "fail - genesis transition - foundries unsorted",
			next:      exampleFoundry,
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						API: tpkg.TestAPI,
						Essence: &iotago.TransactionEssence{
							Outputs: iotago.TxEssenceOutputs{
								&iotago.FoundryOutput{
									Amount: 100,
									// exampleFoundry has serial number 6
									SerialNumber: 7,
									TokenScheme: &iotago.SimpleTokenScheme{
										MintedTokens:  startingSupply,
										MeltedTokens:  big.NewInt(0),
										MaximumSupply: new(big.Int).SetUint64(1000),
									},
									Conditions: iotago.FoundryOutputUnlockConditions{
										&iotago.ImmutableAccountUnlockCondition{Address: exampleAccountIdent},
									},
								},
								exampleFoundry,
							},
						},
						Unlocks: nil,
					},
					InChains: vm.ChainInputSet{
						exampleAccountIdent.AccountID(): &vm.ChainOutputWithCreationSlot{
							Output: &iotago.AccountOutput{FoundryCounter: 5},
						},
					},
					OutChains: map[iotago.ChainID]iotago.ChainOutput{
						exampleAccountIdent.AccountID(): &iotago.AccountOutput{FoundryCounter: 7},
					},
					OutNativeTokens: map[iotago.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): startingSupply,
					},
				},
			},
			wantErr: &iotago.ChainTransitionError{},
		},
		{
			name: "ok - state transition - metadata feature",
			input: &vm.ChainOutputWithCreationSlot{
				Output: exampleFoundry,
			},
			nextMut: map[string]fieldMutations{
				"change_metadata": {
					"Features": iotago.FoundryOutputFeatures{
						&iotago.MetadataFeature{Data: tpkg.RandBytes(20)},
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{
					OutNativeTokens: map[iotago.NativeTokenID]*big.Int{},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - state transition - mint",
			input: &vm.ChainOutputWithCreationSlot{
				Output: exampleFoundry,
			},
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
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{
					OutNativeTokens: map[iotago.NativeTokenID]*big.Int{
						exampleFoundry.MustNativeTokenID(): new(big.Int).SetUint64(300),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - state transition - melt",
			input: &vm.ChainOutputWithCreationSlot{
				Output: exampleFoundry,
			},
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
			svCtx: &vm.Params{
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
			name: "ok - state transition - burn",
			input: &vm.ChainOutputWithCreationSlot{
				Output: exampleFoundry,
			},
			nextMut:   map[string]fieldMutations{},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
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
			name: "ok - state transition - melt complete supply",
			input: &vm.ChainOutputWithCreationSlot{
				Output: exampleFoundry,
			},
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
			svCtx: &vm.Params{
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
			name: "fail - state transition - mint (out: excess)",
			input: &vm.ChainOutputWithCreationSlot{
				Output: exampleFoundry,
			},
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
			svCtx: &vm.Params{
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
			name: "fail - state transition - mint (out: deficit)",
			input: &vm.ChainOutputWithCreationSlot{
				Output: exampleFoundry,
			},
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
			svCtx: &vm.Params{
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
			name: "fail - state transition - melt (out: excess)",
			input: &vm.ChainOutputWithCreationSlot{
				Output: exampleFoundry,
			},
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
			svCtx: &vm.Params{
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
			name: "fail - state transition",
			input: &vm.ChainOutputWithCreationSlot{
				Output: exampleFoundry,
			},
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
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{},
			},
			wantErr: &iotago.ChainTransitionError{},
		},
		{
			name: "ok - destroy transition",
			input: &vm.ChainOutputWithCreationSlot{
				Output: toBeDestoyedFoundry,
			},
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{
					InNativeTokens:  map[iotago.NativeTokenID]*big.Int{},
					OutNativeTokens: map[iotago.NativeTokenID]*big.Int{},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - destroy transition - foundry token unbalanced",
			input: &vm.ChainOutputWithCreationSlot{
				Output: exampleFoundry,
			},
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
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
					cpy := copyObject(t, tt.input.Output, muts).(*iotago.FoundryOutput)
					err := stardustVM.ChainSTVF(tt.transType, tt.input, cpy, tt.svCtx)
					if tt.wantErr != nil {
						//nolint:gosec // false positive
						require.ErrorAs(t, err, &tt.wantErr)
						return
					}
					require.NoError(t, err)
				})
			}
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			err := stardustVM.ChainSTVF(tt.transType, tt.input, tt.next, tt.svCtx)
			if tt.wantErr != nil {
				//nolint:gosec // false positive
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
		Conditions: iotago.NFTOutputUnlockConditions{
			&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
		},
		ImmutableFeatures: iotago.NFTOutputImmFeatures{
			&iotago.IssuerFeature{Address: exampleIssuer},
			&iotago.MetadataFeature{Data: []byte("some-ipfs-link")},
		},
	}

	type test struct {
		name      string
		input     *vm.ChainOutputWithCreationSlot
		next      *iotago.NFTOutput
		nextMut   map[string]fieldMutations
		transType iotago.ChainTransitionType
		svCtx     *vm.Params
		wantErr   error
	}

	tests := []test{
		{
			name:      "ok - genesis transition",
			next:      exampleCurrentNFTOutput,
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
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
			input: &vm.ChainOutputWithCreationSlot{
				Output: exampleCurrentNFTOutput,
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - state transition",
			input: &vm.ChainOutputWithCreationSlot{
				Output: exampleCurrentNFTOutput,
			},
			nextMut: map[string]fieldMutations{
				"amount": {
					"Amount": iotago.BaseToken(1337),
				},
				"address": {
					"Conditions": iotago.NFTOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				"native_tokens": {
					"NativeTokens": tpkg.RandSortNativeTokens(10),
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - state transition",
			input: &vm.ChainOutputWithCreationSlot{
				Output: exampleCurrentNFTOutput,
			},
			nextMut: map[string]fieldMutations{
				"immutable_metadata": {
					"ImmutableFeatures": iotago.NFTOutputImmFeatures{
						&iotago.MetadataFeature{Data: []byte("link-to-cat.gif")},
					},
				},
				"issuer": {
					"ImmutableFeatures": iotago.NFTOutputImmFeatures{
						&iotago.IssuerFeature{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
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
					cpy := copyObject(t, tt.input.Output, muts).(*iotago.NFTOutput)
					err := stardustVM.ChainSTVF(tt.transType, tt.input, cpy, tt.svCtx)
					if tt.wantErr != nil {
						//nolint:gosec // false positive
						require.ErrorAs(t, err, &tt.wantErr)
						return
					}
					require.NoError(t, err)
				})
			}
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			err := stardustVM.ChainSTVF(tt.transType, tt.input, tt.next, tt.svCtx)
			if tt.wantErr != nil {
				//nolint:gosec // false positive
				require.ErrorAs(t, err, &tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestDelegationOutput_ValidateStateTransition(t *testing.T) {
	currentEpoch := iotago.EpochIndex(20)
	epochStartSlot := tpkg.TestAPI.TimeProvider().EpochStart(currentEpoch)
	epochEndSlot := tpkg.TestAPI.TimeProvider().EpochEnd(currentEpoch)
	minCommittableAge := tpkg.TestAPI.ProtocolParameters().MinCommittableAge()
	maxCommittableAge := tpkg.TestAPI.ProtocolParameters().MaxCommittableAge()

	// Commitment indices that will always end up being in the current epoch no matter if
	// future or past bounded.
	epochStartCommitmentIndex := epochStartSlot - minCommittableAge
	epochEndCommitmentIndex := epochEndSlot - maxCommittableAge

	exampleDelegationID := iotago.DelegationIDFromOutputID(tpkg.RandOutputID(0))

	type test struct {
		name      string
		input     *vm.ChainOutputWithCreationSlot
		next      *iotago.DelegationOutput
		nextMut   map[string]fieldMutations
		transType iotago.ChainTransitionType
		svCtx     *vm.Params
		wantErr   error
	}

	tests := []test{
		{
			name: "ok - valid genesis",
			next: &iotago.DelegationOutput{
				Amount:           100,
				DelegatedAmount:  100,
				DelegationID:     iotago.EmptyDelegationID(),
				ValidatorAddress: tpkg.RandAccountAddress(),
				StartEpoch:       currentEpoch + 1,
				EndEpoch:         0,
				Conditions: iotago.DelegationOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Commitment: &iotago.Commitment{
						Index: epochStartCommitmentIndex,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - invalid genesis - non-zero delegation ID",
			next: &iotago.DelegationOutput{
				Amount:           100,
				DelegatedAmount:  100,
				DelegationID:     exampleDelegationID,
				ValidatorAddress: tpkg.RandAccountAddress(),
				StartEpoch:       currentEpoch + 1,
				EndEpoch:         0,
				Conditions: iotago.DelegationOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Commitment: &iotago.Commitment{
						Index: epochStartCommitmentIndex,
					},
				},
			},
			wantErr: iotago.ErrInvalidDelegationNonZeroedID,
		},
		{
			name: "fail - invalid genesis - delegated amount does not match amount",
			next: &iotago.DelegationOutput{
				Amount:           100,
				DelegatedAmount:  120,
				DelegationID:     iotago.EmptyDelegationID(),
				ValidatorAddress: tpkg.RandAccountAddress(),
				StartEpoch:       currentEpoch + 1,
				EndEpoch:         0,
				Conditions: iotago.DelegationOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Commitment: &iotago.Commitment{
						Index: epochStartCommitmentIndex,
					},
				},
			},
			wantErr: iotago.ErrInvalidDelegationAmount,
		},
		{
			name: "fail - invalid genesis - non-zero end epoch",
			next: &iotago.DelegationOutput{
				Amount:           100,
				DelegatedAmount:  100,
				DelegationID:     iotago.EmptyDelegationID(),
				ValidatorAddress: tpkg.RandAccountAddress(),
				StartEpoch:       currentEpoch + 1,
				EndEpoch:         currentEpoch + 5,
				Conditions: iotago.DelegationOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Commitment: &iotago.Commitment{
						Index: epochStartCommitmentIndex,
					},
				},
			},
			wantErr: iotago.ErrInvalidDelegationNonZeroEndEpoch,
		},
		{
			name: "fail - invalid transition - start epoch not set to expected epoch",
			next: &iotago.DelegationOutput{
				Amount:           100,
				DelegatedAmount:  100,
				DelegationID:     iotago.EmptyDelegationID(),
				ValidatorAddress: tpkg.RandAccountAddress(),
				StartEpoch:       currentEpoch - 3,
				EndEpoch:         0,
				Conditions: iotago.DelegationOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Commitment: &iotago.Commitment{
						Index: epochStartCommitmentIndex,
					},
				},
			},
			wantErr: iotago.ErrInvalidDelegationStartEpoch,
		},
		{
			name: "fail - invalid transition - non-zero delegation id on input",
			input: &vm.ChainOutputWithCreationSlot{
				Output: &iotago.DelegationOutput{
					Amount:           100,
					DelegatedAmount:  100,
					DelegationID:     tpkg.RandDelegationID(),
					ValidatorAddress: tpkg.RandAccountAddress(),
					StartEpoch:       currentEpoch + 1,
					EndEpoch:         0,
					Conditions: iotago.DelegationOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			next:      &iotago.DelegationOutput{},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Commitment: &iotago.Commitment{
						Index: epochStartCommitmentIndex,
					},
				},
			},
			wantErr: iotago.ErrInvalidDelegationNonZeroedID,
		},
		{
			name: "fail - invalid transition - modified delegated amount, start epoch and validator id",
			input: &vm.ChainOutputWithCreationSlot{
				ChainID: exampleDelegationID,
				Output: &iotago.DelegationOutput{
					Amount:           100,
					DelegatedAmount:  100,
					DelegationID:     iotago.EmptyDelegationID(),
					ValidatorAddress: tpkg.RandAccountAddress(),
					StartEpoch:       currentEpoch + 1,
					EndEpoch:         0,
					Conditions: iotago.DelegationOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			nextMut: map[string]fieldMutations{
				"delegated_amount_modified": {
					"DelegatedAmount": iotago.BaseToken(1337),
					"Amount":          iotago.BaseToken(5),
					"DelegationID":    exampleDelegationID,
					"EndEpoch":        currentEpoch,
				},
				"start_epoch_modified": {
					"StartEpoch":   iotago.EpochIndex(3),
					"DelegationID": exampleDelegationID,
					"EndEpoch":     currentEpoch,
				},
				"validator_address_modified": {
					"ValidatorAddress": tpkg.RandAccountAddress(),
					"DelegationID":     exampleDelegationID,
					"EndEpoch":         currentEpoch,
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Commitment: &iotago.Commitment{
						Index: epochStartCommitmentIndex,
					},
				},
			},
			wantErr: iotago.ErrInvalidDelegationModified,
		},
		{
			name: "fail - invalid pre-registration slot transition - end epoch not set to expected epoch",
			input: &vm.ChainOutputWithCreationSlot{
				ChainID: exampleDelegationID,
				Output: &iotago.DelegationOutput{
					Amount:           100,
					DelegatedAmount:  100,
					DelegationID:     iotago.EmptyDelegationID(),
					ValidatorAddress: tpkg.RandAccountAddress(),
					StartEpoch:       currentEpoch + 1,
					EndEpoch:         0,
					Conditions: iotago.DelegationOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			nextMut: map[string]fieldMutations{
				"end_epoch_-1": {
					"DelegationID": exampleDelegationID,
					"EndEpoch":     currentEpoch - 1,
				},
				"end_epoch_+1": {
					"DelegationID": exampleDelegationID,
					"EndEpoch":     currentEpoch + 1,
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Commitment: &iotago.Commitment{
						Index: epochStartCommitmentIndex,
					},
				},
			},
			wantErr: iotago.ErrInvalidDelegationEndEpoch,
		},
		{
			name: "fail - invalid post-registration slot transition - end epoch not set to expected epoch",
			input: &vm.ChainOutputWithCreationSlot{
				ChainID: exampleDelegationID,
				Output: &iotago.DelegationOutput{
					Amount:           100,
					DelegatedAmount:  100,
					DelegationID:     iotago.EmptyDelegationID(),
					ValidatorAddress: tpkg.RandAccountAddress(),
					StartEpoch:       currentEpoch + 1,
					EndEpoch:         0,
					Conditions: iotago.DelegationOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			nextMut: map[string]fieldMutations{
				"end_epoch_current": {
					"DelegationID": exampleDelegationID,
					"EndEpoch":     currentEpoch,
				},
				"end_epoch_+2": {
					"DelegationID": exampleDelegationID,
					"EndEpoch":     currentEpoch + 2,
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Commitment: &iotago.Commitment{
						Index: epochEndCommitmentIndex,
					},
				},
			},
			wantErr: iotago.ErrInvalidDelegationEndEpoch,
		},
		{
			name: "fail - invalid transition - cannot claim rewards during transition",
			input: &vm.ChainOutputWithCreationSlot{
				ChainID: exampleDelegationID,
				Output: &iotago.DelegationOutput{
					Amount:           100,
					DelegatedAmount:  100,
					DelegationID:     iotago.EmptyDelegationID(),
					ValidatorAddress: tpkg.RandAccountAddress(),
					StartEpoch:       currentEpoch + 1,
					EndEpoch:         0,
					Conditions: iotago.DelegationOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			next:      &iotago.DelegationOutput{},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Commitment: &iotago.Commitment{
						Index: epochStartCommitmentIndex,
					},
					Rewards: map[iotago.ChainID]iotago.Mana{
						exampleDelegationID: 1,
					},
				},
			},
			wantErr: iotago.ErrInvalidDelegationRewardsClaiming,
		},
		{
			name: "ok - valid destruction",
			input: &vm.ChainOutputWithCreationSlot{
				ChainID: exampleDelegationID,
				Output: &iotago.DelegationOutput{
					Amount:           100,
					DelegatedAmount:  100,
					DelegationID:     iotago.EmptyDelegationID(),
					ValidatorAddress: tpkg.RandAccountAddress(),
					StartEpoch:       currentEpoch + 1,
					EndEpoch:         0,
					Conditions: iotago.DelegationOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			nextMut:   nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Commitment: &iotago.Commitment{
						Index: epochStartCommitmentIndex,
					},
					Rewards: map[iotago.ChainID]iotago.Mana{
						exampleDelegationID: 0,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - invalid destruction - missing reward input",
			input: &vm.ChainOutputWithCreationSlot{
				ChainID: exampleDelegationID,
				Output: &iotago.DelegationOutput{
					Amount:           100,
					DelegatedAmount:  100,
					DelegationID:     iotago.EmptyDelegationID(),
					ValidatorAddress: tpkg.RandAccountAddress(),
					StartEpoch:       currentEpoch + 1,
					EndEpoch:         0,
					Conditions: iotago.DelegationOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			nextMut:   nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Commitment: &iotago.Commitment{
						Index: epochStartCommitmentIndex,
					},
				},
			},
			wantErr: iotago.ErrInvalidDelegationRewardsClaiming,
		},
		{
			name: "fail - invalid genesis - missing commitment input",
			next: &iotago.DelegationOutput{
				Amount:           100,
				DelegatedAmount:  100,
				DelegationID:     iotago.EmptyDelegationID(),
				ValidatorAddress: tpkg.RandAccountAddress(),
				StartEpoch:       currentEpoch + 1,
				EndEpoch:         0,
				Conditions: iotago.DelegationOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.TestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Rewards: map[iotago.ChainID]iotago.Mana{
						exampleDelegationID: 0,
					},
				},
			},
			wantErr: iotago.ErrDelegationCommitmentInputRequired,
		},
	}

	for _, tt := range tests {
		if tt.nextMut != nil {
			for mutName, muts := range tt.nextMut {
				t.Run(fmt.Sprintf("%s_%s", tt.name, mutName), func(t *testing.T) {
					cpy := copyObject(t, tt.input.Output, muts).(*iotago.DelegationOutput)

					if tt.input != nil {
						// create the working set for the test
						if tt.svCtx.WorkingSet.UTXOInputsWithCreationSlot == nil {
							tt.svCtx.WorkingSet.UTXOInputsWithCreationSlot = make(vm.InputSet)
						}

						tt.svCtx.WorkingSet.UTXOInputsWithCreationSlot[tpkg.RandOutputID(0)] = vm.OutputWithCreationSlot{
							Output:       tt.input.Output,
							CreationSlot: tt.input.CreationSlot,
						}
					}

					err := stardustVM.ChainSTVF(tt.transType, tt.input, cpy, tt.svCtx)
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
			if tt.input != nil {
				// create the working set for the test
				if tt.svCtx.WorkingSet.UTXOInputsWithCreationSlot == nil {
					tt.svCtx.WorkingSet.UTXOInputsWithCreationSlot = make(vm.InputSet)
				}

				tt.svCtx.WorkingSet.UTXOInputsWithCreationSlot[tpkg.RandOutputID(0)] = vm.OutputWithCreationSlot{
					Output:       tt.input.Output,
					CreationSlot: tt.input.CreationSlot,
				}
			}

			err := stardustVM.ChainSTVF(tt.transType, tt.input, tt.next, tt.svCtx)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
