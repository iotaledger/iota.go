//nolint:forcetypeassert,dupl,nlreturn,scopelint
package nova_test

import (
	"fmt"
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
func copyObjectAndMutate(t *testing.T, source any, mutations fieldMutations) any {
	srcBytes, err := tpkg.ZeroCostTestAPI.Encode(source)
	require.NoError(t, err)

	ptrToCpyOfSrc := reflect.New(reflect.ValueOf(source).Elem().Type())

	cpySeri := ptrToCpyOfSrc.Interface()
	_, err = tpkg.ZeroCostTestAPI.Decode(srcBytes, cpySeri)
	require.NoError(t, err)

	for fieldName, newVal := range mutations {
		ptrToCpyOfSrc.Elem().FieldByName(fieldName).Set(reflect.ValueOf(newVal))
	}

	return cpySeri
}

func TestAccountOutput_ValidateStateTransition(t *testing.T) {
	exampleIssuer := tpkg.RandEd25519Address()
	exampleAccountID := tpkg.RandAccountAddress().AccountID()

	exampleAddress := tpkg.RandEd25519Address()

	exampleExistingFoundryOutput := &iotago.FoundryOutput{
		Amount:       100,
		SerialNumber: 5,
		TokenScheme: &iotago.SimpleTokenScheme{
			MintedTokens:  new(big.Int).SetInt64(1000),
			MeltedTokens:  big.NewInt(0),
			MaximumSupply: new(big.Int).SetInt64(10000),
		},
		UnlockConditions: iotago.FoundryOutputUnlockConditions{
			&iotago.ImmutableAccountUnlockCondition{Address: exampleAccountID.ToAddress().(*iotago.AccountAddress)},
		},
	}
	exampleExistingFoundryOutputFoundryID := exampleExistingFoundryOutput.MustFoundryID()

	currentEpoch := iotago.EpochIndex(20)
	currentSlot := tpkg.ZeroCostTestAPI.TimeProvider().EpochStart(currentEpoch)

	blockIssuerPubKey := iotago.Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray())
	exampleBlockIssuerFeature := &iotago.BlockIssuerFeature{
		BlockIssuerKeys: iotago.NewBlockIssuerKeys(blockIssuerPubKey),
		ExpirySlot:      currentSlot + tpkg.ZeroCostTestAPI.ProtocolParameters().MaxCommittableAge(),
	}

	exampleBIC := map[iotago.AccountID]iotago.BlockIssuanceCredits{
		exampleAccountID: 100,
	}

	type test struct {
		name      string
		input     *vm.ChainOutputWithIDs
		next      *iotago.AccountOutput
		nextMut   map[string]fieldMutations
		transType iotago.ChainTransitionType
		svCtx     *vm.Params
		wantErr   error
	}

	tests := []*test{
		{
			name: "ok - genesis transition",
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: iotago.AccountID{},
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				ImmutableFeatures: iotago.AccountOutputImmFeatures{
					&iotago.IssuerFeature{Address: exampleIssuer},
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
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
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				ImmutableFeatures: iotago.AccountOutputImmFeatures{
					&iotago.IssuerFeature{Address: exampleIssuer},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
						ExpirySlot:      1000,
					},
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: 900,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: 900,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
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
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				ImmutableFeatures: iotago.AccountOutputImmFeatures{
					&iotago.IssuerFeature{Address: exampleIssuer},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
						ExpirySlot:      1000,
					},
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: 10001,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: 10001,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrBlockIssuerExpiryTooEarly,
		},
		{
			name: "fail - block issuer genesis expired but within MCA",
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: iotago.AccountID{},
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				ImmutableFeatures: iotago.AccountOutputImmFeatures{
					&iotago.IssuerFeature{Address: exampleIssuer},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
						ExpirySlot:      1000,
					},
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: 991,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: 991,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrBlockIssuerExpiryTooEarly,
		},
		{
			name: "ok - staking genesis transition",
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: iotago.AccountID{},
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.StakingFeature{
						StakedAmount: 50,
						FixedCost:    5,
						StartEpoch:   currentEpoch,
						EndEpoch:     iotago.MaxEpochIndex,
					},
					exampleBlockIssuerFeature,
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
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
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.StakingFeature{
						StakedAmount: 50,
						FixedCost:    5,
						StartEpoch:   currentEpoch - 2,
						EndEpoch:     iotago.MaxEpochIndex,
					},
					exampleBlockIssuerFeature,
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: iotago.ErrStakingStartEpochInvalid,
		},
		{
			name: "fail - staking genesis end epoch too early",
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: iotago.AccountID{},
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.StakingFeature{
						StakedAmount: 50,
						FixedCost:    5,
						StartEpoch:   currentEpoch,
						EndEpoch:     currentEpoch + tpkg.ZeroCostTestAPI.ProtocolParameters().StakingUnbondingPeriod() - 1,
					},
					exampleBlockIssuerFeature,
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: iotago.ErrStakingEndEpochTooEarly,
		},
		{
			name: "fail - staking feature without block issuer feature",
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: iotago.AccountID{},
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.StakingFeature{
						StakedAmount: 50,
						FixedCost:    5,
						StartEpoch:   currentEpoch,
						EndEpoch:     iotago.MaxEpochIndex,
					},
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: iotago.ErrStakingBlockIssuerFeatureMissing,
		},
		{
			name: "ok - valid staking transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
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
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.StakingFeature{
						StakedAmount: 100,
						FixedCost:    50,
						StartEpoch:   currentEpoch,
						EndEpoch:     currentEpoch + tpkg.ZeroCostTestAPI.ProtocolParameters().StakingUnbondingPeriod(),
					},
					exampleBlockIssuerFeature,
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
				},
			},
		},
		{
			name: "ok - adding staking feature in account state transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: iotago.AccountOutputFeatures{
						exampleBlockIssuerFeature,
					},
				},
			},
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.StakingFeature{
						StakedAmount: 100,
						FixedCost:    50,
						StartEpoch:   currentEpoch,
						EndEpoch:     currentEpoch + tpkg.ZeroCostTestAPI.ProtocolParameters().StakingUnbondingPeriod(),
					},
					exampleBlockIssuerFeature,
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
				},
			},
		},
		{
			name: "fail - adding staking feature in account state transition with start epoch set too early",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: iotago.AccountOutputFeatures{
						exampleBlockIssuerFeature,
					},
				},
			},
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.StakingFeature{
						StakedAmount: 100,
						FixedCost:    50,
						StartEpoch:   currentEpoch - 5,
						EndEpoch:     currentEpoch + tpkg.ZeroCostTestAPI.ProtocolParameters().StakingUnbondingPeriod(),
					},
					exampleBlockIssuerFeature,
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: iotago.ErrStakingStartEpochInvalid,
		},
		{
			name: "fail - negative BIC during account state transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: iotago.AccountOutputFeatures{
						exampleBlockIssuerFeature,
					},
				},
			},
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: iotago.AccountOutputFeatures{
					exampleBlockIssuerFeature,
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
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
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
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
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: iotago.AccountOutputFeatures{
					exampleBlockIssuerFeature,
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: iotago.ErrStakingFeatureRemovedBeforeUnbonding,
		},
		{
			name: "fail - changing staking feature's staked amount",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
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
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: exampleAddress},
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
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: iotago.ErrStakingFeatureModifiedBeforeUnbonding,
		},
		{
			name: "fail - reducing staking feature's end epoch by more than the unbonding period",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
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
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.StakingFeature{
						StakedAmount: 100,
						FixedCost:    50,
						StartEpoch:   currentEpoch,
						EndEpoch:     currentEpoch + tpkg.ZeroCostTestAPI.ProtocolParameters().StakingUnbondingPeriod() - 5,
					},
					exampleBlockIssuerFeature,
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: iotago.ErrStakingEndEpochTooEarly,
		},
		{
			name: "fail - account removes block issuer feature while having a staking feature",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(1000, 0),
				ChainID:  exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.StakingFeature{
							StakedAmount: 50,
							FixedCost:    5,
							StartEpoch:   currentEpoch,
							EndEpoch:     iotago.MaxEpochIndex,
						},
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      990,
						},
					},
				},
			},
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.StakingFeature{
						StakedAmount: 50,
						FixedCost:    5,
						StartEpoch:   currentEpoch,
						EndEpoch:     iotago.MaxEpochIndex,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: iotago.ErrStakingBlockIssuerFeatureMissing,
		},
		{
			name: "fail - expired staking feature removed without specifying reward input",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(1000, 0),
				ChainID:  exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
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
			},
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: iotago.AccountOutputFeatures{
					exampleBlockIssuerFeature,
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: iotago.ErrStakingRewardInputMissing,
		},
		{
			name: "fail - changing an expired staking feature without claiming",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(1000, 0),
				ChainID:  exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
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
			},
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.StakingFeature{
						StakedAmount: 80,
						FixedCost:    5,
						StartEpoch:   currentEpoch,
						EndEpoch:     currentEpoch + testProtoParams.StakingUnbondingPeriod(),
					},
					exampleBlockIssuerFeature,
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
				},
			},
			wantErr: iotago.ErrStakingRewardInputMissing,
		},
		{
			name: "fail - claiming rewards of an expired staking feature without resetting start epoch",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(1000, 0),
				ChainID:  exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
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
			},
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: exampleAddress},
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
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
					Rewards: map[iotago.ChainID]iotago.Mana{
						exampleAccountID: 200,
					},
				},
			},
			wantErr: iotago.ErrStakingStartEpochInvalid,
		},
		{
			name: "fail - claiming rewards without removing staking feature",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(1000, 0),
				ChainID:  exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
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
			},
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: exampleAddress},
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
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
					BIC: exampleBIC,
					Rewards: map[iotago.ChainID]iotago.Mana{
						exampleAccountID: 200,
					},
				},
			},
			wantErr: iotago.ErrStakingRewardClaimingInvalid,
		},
		{
			name: "fail - destroy account with expired staking feature but without claiming rewards",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(1000, 0),
				ChainID:  exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.StakingFeature{
							StakedAmount: 50,
							FixedCost:    5,
							StartEpoch:   currentEpoch - 10,
							EndEpoch:     currentEpoch - 5,
						},
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      currentSlot - 50,
						},
					},
				},
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: currentSlot,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
					BIC:     exampleBIC,
					Rewards: map[iotago.ChainID]iotago.Mana{},
				},
			},
			wantErr: iotago.ErrStakingRewardInputMissing,
		},
		{
			name: "ok - destroy transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: tpkg.RandAccountAddress().AccountID(),
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
					},
				},
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - destroy block issuer account with negative BIC",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      1000,
						},
					},
				},
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: 1001,
					},
					UnlockedIdents: vm.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: 1001,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
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
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      1000,
						},
					},
				},
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: 1001,
					},
					UnlockedIdents: vm.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: 1001,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrBlockIssuanceCreditInputMissing,
		},
		{
			name: "fail - non-expired block issuer destroy transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: tpkg.RandAccountAddress().AccountID(),
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      1000,
						},
					},
				},
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: 1000,
					},
					UnlockedIdents: vm.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: 1000,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrBlockIssuerNotExpired,
		},
		{
			name: "ok - expired block issuer destroy transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      1000,
						},
					},
				},
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: 1001,
					},
					UnlockedIdents: vm.UnlockedIdentities{},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredits{
						exampleAccountID: 0,
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							Inputs:       nil,
							CreationSlot: 1001,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "failed - remove non-expired block issuer feature transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      1000,
						},
					},
				},
			},
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: iotago.AccountOutputFeatures{},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: 999,
					},
					UnlockedIdents: vm.UnlockedIdentities{},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredits{
						exampleAccountID: 0,
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							Inputs:       nil,
							CreationSlot: 999,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrBlockIssuerNotExpired,
		},
		{
			name: "ok - remove expired block issuer feature transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      1000,
						},
					},
				},
			},
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: iotago.AccountOutputFeatures{},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: 1001,
					},
					UnlockedIdents: vm.UnlockedIdentities{},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredits{
						exampleAccountID: 0,
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							Inputs:       nil,
							CreationSlot: 1001,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - foundry counter increased by number of new foundries",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
					},
					FoundryCounter: 5,
				},
			},
			next: &iotago.AccountOutput{
				Amount:    200,
				AccountID: exampleAccountID,
				// mutating owner
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				FoundryCounter: 7,
				Features: iotago.AccountOutputFeatures{
					&iotago.SenderFeature{Address: exampleAddress},
					&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("1337")}},
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
						ExpirySlot:      1015,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleAddress.Key(): {UnlockedAt: 0},
					},
					Commitment: &iotago.Commitment{
						Slot: 990,
					},
					BIC: exampleBIC,
					InChains: map[iotago.ChainID]*vm.ChainOutputWithIDs{
						// serial number 5
						exampleExistingFoundryOutputFoundryID: {
							ChainID:  exampleExistingFoundryOutputFoundryID,
							OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
							Output:   exampleExistingFoundryOutput,
						},
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: 900,
							Inputs:       nil,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything())},
						Outputs: iotago.TxEssenceOutputs{
							&iotago.FoundryOutput{
								Amount:       100,
								SerialNumber: 6,
								TokenScheme:  &iotago.SimpleTokenScheme{},
								UnlockConditions: iotago.FoundryOutputUnlockConditions{
									&iotago.ImmutableAccountUnlockCondition{Address: exampleAccountID.ToAddress().(*iotago.AccountAddress)},
								},
							},
							&iotago.FoundryOutput{
								Amount:       100,
								SerialNumber: 7,
								TokenScheme:  &iotago.SimpleTokenScheme{},
								UnlockConditions: iotago.FoundryOutputUnlockConditions{
									&iotago.ImmutableAccountUnlockCondition{Address: exampleAccountID.ToAddress().(*iotago.AccountAddress)},
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
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      1000,
						},
					},
				},
			},
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.SenderFeature{Address: exampleAddress},
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
						ExpirySlot:      1000,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: 990,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleAddress.Key(): {UnlockedAt: 0},
					},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredits{
						exampleAccountID: 10,
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							Inputs:       nil,
							CreationSlot: 990,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - update account immutable features",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
					},
					ImmutableFeatures: iotago.AccountOutputImmFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      900,
						},
					},
				},
			},
			next: &iotago.AccountOutput{
				Amount:    200,
				AccountID: exampleAccountID,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: exampleAddress},
				},
				ImmutableFeatures: iotago.AccountOutputImmFeatures{
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
						ExpirySlot:      999,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleAddress.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrChainOutputImmutableFeaturesChanged,
		},
		{
			name: "fail - update expired block issuer feature with extending expiration before MCA",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      900,
						},
					},
				},
			},
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.SenderFeature{Address: exampleAddress},
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
						ExpirySlot:      999,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: 990,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleAddress.Key(): {UnlockedAt: 0},
					},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredits{
						exampleAccountID: 10,
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							Inputs:       nil,
							CreationSlot: 990,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrBlockIssuerExpiryTooEarly,
		},
		{
			name: "fail - update expired block issuer feature with extending expiration to the past before MCA",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      1100,
						},
					},
				},
			},
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.SenderFeature{Address: exampleAddress},
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
						ExpirySlot:      999,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: 990,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleAddress.Key(): {UnlockedAt: 0},
					},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredits{
						exampleAccountID: 10,
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							Inputs:       nil,
							CreationSlot: 990,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrBlockIssuerExpiryTooEarly,
		},
		{
			name: "fail - update block issuer account with negative BIC",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      1000,
						},
					},
				},
			},
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("1337")}},
					&iotago.SenderFeature{Address: exampleAddress},
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
						ExpirySlot:      1000,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: 900,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleAddress.Key(): {UnlockedAt: 0},
					},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredits{
						exampleAccountID: -1,
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							Inputs:       nil,
							CreationSlot: 900,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrAccountLocked,
		},
		{
			name: "fail - update block issuer account without BIC provided",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      1000,
						},
					},
				},
			},
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.SenderFeature{Address: exampleAddress},
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
						ExpirySlot:      1000,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: 900,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleAddress.Key(): {UnlockedAt: 0},
					},

					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: 900,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrBlockIssuanceCreditInputMissing,
		},
		{
			name: "ok - update block issuer feature expiration to earlier slot",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      1000,
						},
					},
				},
			},
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: exampleAddress},
				},
				Features: iotago.AccountOutputFeatures{
					&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("1337")}},
					&iotago.SenderFeature{Address: exampleAddress},
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
						ExpirySlot:      999,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: 900,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleAddress.Key(): {UnlockedAt: 0},
					},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredits{
						exampleAccountID: 10,
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: 900,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - non-expired block issuer replace key",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				ChainID:  exampleAccountID,
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
							ExpirySlot:      1000,
						},
					},
					FoundryCounter: 5,
				},
			},
			next: &iotago.AccountOutput{
				Amount:    100,
				AccountID: exampleAccountID,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: exampleAddress},
				},
				FoundryCounter: 5,
				Features: iotago.AccountOutputFeatures{
					&iotago.SenderFeature{Address: exampleAddress},
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: tpkg.RandBlockIssuerKeys(1),
						ExpirySlot:      1000,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					Commitment: &iotago.Commitment{
						Slot: 0,
					},
					UnlockedIdents: vm.UnlockedIdentities{
						exampleAddress.Key(): {UnlockedAt: 0},
					},
					InChains: map[iotago.ChainID]*vm.ChainOutputWithIDs{
						// serial number 5
						exampleExistingFoundryOutputFoundryID: {
							ChainID:  exampleExistingFoundryOutputFoundryID,
							OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
							Output:   exampleExistingFoundryOutput,
						},
					},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredits{
						exampleAccountID: 10,
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							Inputs:       nil,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything())},
						Outputs: iotago.TxEssenceOutputs{
							&iotago.FoundryOutput{
								Amount:       100,
								SerialNumber: 6,
								TokenScheme:  &iotago.SimpleTokenScheme{},
								UnlockConditions: iotago.FoundryOutputUnlockConditions{
									&iotago.ImmutableAccountUnlockCondition{Address: exampleAccountID.ToAddress().(*iotago.AccountAddress)},
								},
							},
							&iotago.FoundryOutput{
								Amount:       100,
								SerialNumber: 7,
								TokenScheme:  &iotago.SimpleTokenScheme{},
								UnlockConditions: iotago.FoundryOutputUnlockConditions{
									&iotago.ImmutableAccountUnlockCondition{Address: exampleAccountID.ToAddress().(*iotago.AccountAddress)},
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - invalid foundry counters",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &iotago.AccountOutput{
					Amount:         100,
					AccountID:      exampleAccountID,
					FoundryCounter: 5,
					UnlockConditions: iotago.AccountOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: exampleAddress},
					},
					ImmutableFeatures: iotago.AccountOutputImmFeatures{
						&iotago.IssuerFeature{Address: exampleIssuer},
					},
				},
			},
			nextMut: map[string]fieldMutations{
				"foundry_counter_lower_than_current": {
					"FoundryCounter": uint32(4),
				},
				"foundries_not_created": {
					"FoundryCounter": uint32(7),
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					InChains:       vm.ChainInputSet{},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrAccountInvalidFoundryCounter,
		},
	}

	for _, tt2 := range tests {
		tt := tt2

		if tt.nextMut != nil {
			for mutName, muts := range tt.nextMut {
				t.Run(fmt.Sprintf("%s_%s", tt.name, mutName), func(t *testing.T) {
					cpy := copyObjectAndMutate(t, tt.input.Output, muts).(*iotago.AccountOutput)

					createWorkingSet(t, tt.input, tt.svCtx.WorkingSet)

					err := novaVM.ChainSTVF(tt.svCtx, tt.transType, tt.input, cpy)
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
			createWorkingSet(t, tt.input, tt.svCtx.WorkingSet)

			err := novaVM.ChainSTVF(tt.svCtx, tt.transType, tt.input, tt.next)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestAnchorOutput_ValidateStateTransition(t *testing.T) {
	exampleIssuer := tpkg.RandEd25519Address()
	exampleAnchorID := tpkg.RandAnchorAddress().AnchorID()

	exampleStateCtrl := tpkg.RandEd25519Address()
	exampleGovCtrl := tpkg.RandEd25519Address()

	type test struct {
		name      string
		input     *vm.ChainOutputWithIDs
		next      *iotago.AnchorOutput
		nextMut   map[string]fieldMutations
		transType iotago.ChainTransitionType
		svCtx     *vm.Params
		wantErr   error
	}

	tests := []*test{
		{
			name: "ok - genesis transition",
			next: &iotago.AnchorOutput{
				Amount:   100,
				AnchorID: iotago.AnchorID{},
				UnlockConditions: iotago.AnchorOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				ImmutableFeatures: iotago.AnchorOutputImmFeatures{
					&iotago.IssuerFeature{Address: exampleIssuer},
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - destroy transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &iotago.AnchorOutput{
					Amount:   100,
					AnchorID: tpkg.RandAnchorAddress().AnchorID(),
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
				},
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - gov transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &iotago.AnchorOutput{
					Amount:   100,
					AnchorID: exampleAnchorID,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					StateIndex: 10,
					Features: iotago.AnchorOutputFeatures{
						&iotago.StateMetadataFeature{Entries: iotago.StateMetadataFeatureEntries{"data": []byte("1337")}},
					},
				},
			},
			next: &iotago.AnchorOutput{
				Amount:     100,
				AnchorID:   exampleAnchorID,
				StateIndex: 10,
				// mutating controllers
				UnlockConditions: iotago.AnchorOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: iotago.AnchorOutputFeatures{
					&iotago.SenderFeature{Address: exampleGovCtrl},
					&iotago.StateMetadataFeature{Entries: iotago.StateMetadataFeatureEntries{"data": []byte("1337")}},
					// adding metadata feature
					&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("1338")}},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleGovCtrl.Key(): {UnlockedAt: 0},
					},
					Commitment: &iotago.Commitment{
						Slot: 990,
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: 900,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - state transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &iotago.AnchorOutput{
					Amount:   100,
					AnchorID: exampleAnchorID,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					StateIndex: 10,
					Features: iotago.AnchorOutputFeatures{
						&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("1338")}},
					},
				},
			},
			next: &iotago.AnchorOutput{
				Amount:   200,
				AnchorID: exampleAnchorID,
				UnlockConditions: iotago.AnchorOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				StateIndex: 11,
				Features: iotago.AnchorOutputFeatures{
					&iotago.SenderFeature{Address: exampleStateCtrl},
					&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("1338")}},
					// adding state metadata feature
					&iotago.StateMetadataFeature{Entries: iotago.StateMetadataFeatureEntries{"data": []byte("1337")}},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleStateCtrl.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							Inputs:       nil,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - update anchor immutable features in gov transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &iotago.AnchorOutput{
					Amount:   100,
					AnchorID: exampleAnchorID,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					ImmutableFeatures: iotago.AnchorOutputImmFeatures{
						&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("1337")}},
					},
					StateIndex: 10,
				},
			},
			next: &iotago.AnchorOutput{
				Amount:   100,
				AnchorID: exampleAnchorID,
				UnlockConditions: iotago.AnchorOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				StateIndex: 10,
				ImmutableFeatures: iotago.AnchorOutputImmFeatures{
					&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("1338")}},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleStateCtrl.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrInvalidAnchorGovernanceTransition,
		},
		{
			name: "fail - update anchor immutable features in state transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &iotago.AnchorOutput{
					Amount:   100,
					AnchorID: exampleAnchorID,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					ImmutableFeatures: iotago.AnchorOutputImmFeatures{
						&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("1337")}},
					},
					StateIndex: 10,
				},
			},
			next: &iotago.AnchorOutput{
				Amount:   200,
				AnchorID: exampleAnchorID,
				UnlockConditions: iotago.AnchorOutputUnlockConditions{
					&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
					&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
				},
				StateIndex: 11,
				ImmutableFeatures: iotago.AnchorOutputImmFeatures{
					&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("1338")}},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleStateCtrl.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrInvalidAnchorStateTransition,
		},
		{
			name: "fail - gov transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &iotago.AnchorOutput{
					Amount:     100,
					AnchorID:   exampleAnchorID,
					StateIndex: 10,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AnchorOutputFeatures{
						&iotago.StateMetadataFeature{
							Entries: iotago.StateMetadataFeatureEntries{
								"data": []byte("foo"),
							},
						},
					},
				},
			},
			nextMut: map[string]fieldMutations{
				"amount": {
					"Amount": iotago.BaseToken(1337),
				},
				"state_metadata_feature_changed": {
					"Features": iotago.AnchorOutputFeatures{
						&iotago.StateMetadataFeature{
							Entries: iotago.StateMetadataFeatureEntries{
								"data": []byte("bar"),
							},
						},
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrInvalidAnchorGovernanceTransition,
		},
		{
			name: "fail - state transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &iotago.AnchorOutput{
					Amount:     100,
					AnchorID:   exampleAnchorID,
					StateIndex: 10,
					UnlockConditions: iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AnchorOutputFeatures{
						&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("foo")}},
					},
					ImmutableFeatures: iotago.AnchorOutputImmFeatures{
						&iotago.IssuerFeature{Address: exampleIssuer},
					},
				},
			},
			nextMut: map[string]fieldMutations{
				"state_controller": {
					"StateIndex": uint32(11),
					"UnlockConditions": iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
				},
				"governance_controller": {
					"StateIndex": uint32(11),
					"UnlockConditions": iotago.AnchorOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
				"state_index_lower": {
					"StateIndex": uint32(4),
				},
				"state_index_bigger_more_than_1": {
					"StateIndex": uint32(7),
				},
				"metadata_feature_changed": {
					"StateIndex": uint32(11),
					"Features": iotago.AnchorOutputFeatures{
						&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("bar")}},
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					InChains:       vm.ChainInputSet{},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrInvalidAnchorStateTransition,
		},
	}

	for _, tt2 := range tests {
		tt := tt2

		if tt.nextMut != nil {
			for mutName, muts := range tt.nextMut {
				t.Run(fmt.Sprintf("%s_%s", tt.name, mutName), func(t *testing.T) {
					cpy := copyObjectAndMutate(t, tt.input.Output, muts).(*iotago.AnchorOutput)

					createWorkingSet(t, tt.input, tt.svCtx.WorkingSet)

					err := novaVM.ChainSTVF(tt.svCtx, tt.transType, tt.input, cpy)
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
			createWorkingSet(t, tt.input, tt.svCtx.WorkingSet)

			err := novaVM.ChainSTVF(tt.svCtx, tt.transType, tt.input, tt.next)
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
		UnlockConditions: iotago.FoundryOutputUnlockConditions{
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
		UnlockConditions: iotago.FoundryOutputUnlockConditions{
			&iotago.ImmutableAccountUnlockCondition{Address: exampleAccountIdent},
		},
	}

	type test struct {
		name      string
		input     *vm.ChainOutputWithIDs
		next      *iotago.FoundryOutput
		nextMut   map[string]fieldMutations
		transType iotago.ChainTransitionType
		svCtx     *vm.Params
		wantErr   error
	}

	tests := []*test{
		{
			name:      "ok - genesis transition",
			next:      exampleFoundry,
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
						Outputs: iotago.TxEssenceOutputs{exampleFoundry},
					},
					InChains: vm.ChainInputSet{
						exampleAccountIdent.AccountID(): &vm.ChainOutputWithIDs{
							ChainID:  exampleAccountIdent.AccountID(),
							OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
							Output:   &iotago.AccountOutput{FoundryCounter: 5},
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
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
						Outputs: iotago.TxEssenceOutputs{exampleFoundry},
					},
					InChains: vm.ChainInputSet{
						exampleAccountIdent.AccountID(): &vm.ChainOutputWithIDs{
							ChainID:  exampleAccountIdent.AccountID(),
							OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
							Output:   &iotago.AccountOutput{FoundryCounter: 5},
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
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
						Outputs: iotago.TxEssenceOutputs{exampleFoundry},
					},
					InChains: vm.ChainInputSet{
						exampleAccountIdent.AccountID(): &vm.ChainOutputWithIDs{
							ChainID:  exampleAccountIdent.AccountID(),
							OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
							Output:   &iotago.AccountOutput{FoundryCounter: 6},
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
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
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
								UnlockConditions: iotago.FoundryOutputUnlockConditions{
									&iotago.ImmutableAccountUnlockCondition{Address: exampleAccountIdent},
								},
							},
							exampleFoundry,
						},
					},
					InChains: vm.ChainInputSet{
						exampleAccountIdent.AccountID(): &vm.ChainOutputWithIDs{
							ChainID:  exampleAccountIdent.AccountID(),
							OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
							Output:   &iotago.AccountOutput{FoundryCounter: 5},
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
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleFoundry,
			},
			nextMut: map[string]fieldMutations{
				"change_metadata": {
					"Features": iotago.FoundryOutputFeatures{
						&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": tpkg.RandBytes(20)}},
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{
					OutNativeTokens: map[iotago.NativeTokenID]*big.Int{},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - state transition - mint",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleFoundry,
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
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - state transition - melt",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleFoundry,
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
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - state transition - burn",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleFoundry,
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
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - state transition - melt complete supply",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleFoundry,
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
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - state transition - mint (out: excess)",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleFoundry,
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
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrNativeTokenSumUnbalanced,
		},
		{
			name: "fail - state transition - mint (out: deficit)",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleFoundry,
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
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrNativeTokenSumUnbalanced,
		},
		{
			name: "fail - state transition - melt (out: excess)",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleFoundry,
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
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrNativeTokenSumUnbalanced,
		},
		{
			name: "fail - state transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleFoundry,
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
				WorkingSet: &vm.WorkingSet{
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: &iotago.ChainTransitionError{},
		},
		{
			name: "ok - destroy transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   toBeDestoyedFoundry,
			},
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				WorkingSet: &vm.WorkingSet{
					InNativeTokens:  map[iotago.NativeTokenID]*big.Int{},
					OutNativeTokens: map[iotago.NativeTokenID]*big.Int{},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - destroy transition - foundry token unbalanced",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleFoundry,
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
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
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
					cpy := copyObjectAndMutate(t, tt.input.Output, muts).(*iotago.FoundryOutput)

					createWorkingSet(t, tt.input, tt.svCtx.WorkingSet)

					err := novaVM.ChainSTVF(tt.svCtx, tt.transType, tt.input, cpy)
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
			createWorkingSet(t, tt.input, tt.svCtx.WorkingSet)

			err := novaVM.ChainSTVF(tt.svCtx, tt.transType, tt.input, tt.next)
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
		UnlockConditions: iotago.NFTOutputUnlockConditions{
			&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
		},
		ImmutableFeatures: iotago.NFTOutputImmFeatures{
			&iotago.IssuerFeature{Address: exampleIssuer},
			&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("some-ipfs-link")}},
		},
	}

	type test struct {
		name      string
		input     *vm.ChainOutputWithIDs
		next      *iotago.NFTOutput
		nextMut   map[string]fieldMutations
		transType iotago.ChainTransitionType
		svCtx     *vm.Params
		wantErr   error
	}

	tests := []*test{
		{
			name:      "ok - genesis transition",
			next:      exampleCurrentNFTOutput,
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - destroy transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleCurrentNFTOutput,
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - state transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleCurrentNFTOutput,
			},
			nextMut: map[string]fieldMutations{
				"amount": {
					"Amount": iotago.BaseToken(1337),
				},
				"address": {
					"UnlockConditions": iotago.NFTOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - state transition",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output:   exampleCurrentNFTOutput,
			},
			nextMut: map[string]fieldMutations{
				"immutable_metadata": {
					"ImmutableFeatures": iotago.NFTOutputImmFeatures{
						&iotago.MetadataFeature{Entries: iotago.MetadataFeatureEntries{"data": []byte("link-to-cat.gif")}},
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
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: &iotago.ChainTransitionError{},
		},
	}

	for _, tt := range tests {

		if tt.nextMut != nil {
			for mutName, muts := range tt.nextMut {
				t.Run(fmt.Sprintf("%s_%s", tt.name, mutName), func(t *testing.T) {
					cpy := copyObjectAndMutate(t, tt.input.Output, muts).(*iotago.NFTOutput)

					createWorkingSet(t, tt.input, tt.svCtx.WorkingSet)

					err := novaVM.ChainSTVF(tt.svCtx, tt.transType, tt.input, cpy)
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
			createWorkingSet(t, tt.input, tt.svCtx.WorkingSet)

			err := novaVM.ChainSTVF(tt.svCtx, tt.transType, tt.input, tt.next)
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
	epochStartSlot := tpkg.ZeroCostTestAPI.TimeProvider().EpochStart(currentEpoch)
	epochEndSlot := tpkg.ZeroCostTestAPI.TimeProvider().EpochEnd(currentEpoch)
	minCommittableAge := tpkg.ZeroCostTestAPI.ProtocolParameters().MinCommittableAge()
	maxCommittableAge := tpkg.ZeroCostTestAPI.ProtocolParameters().MaxCommittableAge()

	// Commitment indices that will always end up being in the current epoch no matter if
	// future or past bounded.
	epochStartCommitmentIndex := epochStartSlot - minCommittableAge
	epochEndCommitmentIndex := epochEndSlot - maxCommittableAge

	exampleDelegationID := iotago.DelegationIDFromOutputID(tpkg.RandOutputID(0))

	type test struct {
		name      string
		input     *vm.ChainOutputWithIDs
		next      *iotago.DelegationOutput
		nextMut   map[string]fieldMutations
		transType iotago.ChainTransitionType
		svCtx     *vm.Params
		wantErr   error
	}

	tests := []*test{
		{
			name: "ok - valid genesis",
			next: &iotago.DelegationOutput{
				Amount:           100,
				DelegatedAmount:  100,
				DelegationID:     iotago.EmptyDelegationID(),
				ValidatorAddress: tpkg.RandAccountAddress(),
				StartEpoch:       currentEpoch + 1,
				EndEpoch:         0,
				UnlockConditions: iotago.DelegationOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Commitment: &iotago.Commitment{
						Slot: epochStartCommitmentIndex,
					},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
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
				UnlockConditions: iotago.DelegationOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Commitment: &iotago.Commitment{
						Slot: epochStartCommitmentIndex,
					},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrNewChainOutputHasNonZeroedID,
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
				UnlockConditions: iotago.DelegationOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Commitment: &iotago.Commitment{
						Slot: epochStartCommitmentIndex,
					},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrDelegationAmountMismatch,
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
				UnlockConditions: iotago.DelegationOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Commitment: &iotago.Commitment{
						Slot: epochStartCommitmentIndex,
					},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrDelegationEndEpochNotZero,
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
				UnlockConditions: iotago.DelegationOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Commitment: &iotago.Commitment{
						Slot: epochStartCommitmentIndex,
					},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrDelegationStartEpochInvalid,
		},
		{
			name: "fail - invalid transition - non-zero delegation id on input",
			input: &vm.ChainOutputWithIDs{
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &iotago.DelegationOutput{
					Amount:           100,
					DelegatedAmount:  100,
					DelegationID:     tpkg.RandDelegationID(),
					ValidatorAddress: tpkg.RandAccountAddress(),
					StartEpoch:       currentEpoch + 1,
					EndEpoch:         0,
					UnlockConditions: iotago.DelegationOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			next:      &iotago.DelegationOutput{},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Commitment: &iotago.Commitment{
						Slot: epochStartCommitmentIndex,
					},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrDelegationOutputTransitionedTwice,
		},
		{
			name: "fail - invalid transition - modified delegated amount, start epoch and validator id",
			input: &vm.ChainOutputWithIDs{
				ChainID:  exampleDelegationID,
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &iotago.DelegationOutput{
					Amount:           100,
					DelegatedAmount:  100,
					DelegationID:     iotago.EmptyDelegationID(),
					ValidatorAddress: tpkg.RandAccountAddress(),
					StartEpoch:       currentEpoch + 1,
					EndEpoch:         0,
					UnlockConditions: iotago.DelegationOutputUnlockConditions{
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
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Commitment: &iotago.Commitment{
						Slot: epochStartCommitmentIndex,
					},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrDelegationModified,
		},
		{
			name: "fail - invalid pre-registration slot transition - end epoch not set to expected epoch",
			input: &vm.ChainOutputWithIDs{
				ChainID:  exampleDelegationID,
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &iotago.DelegationOutput{
					Amount:           100,
					DelegatedAmount:  100,
					DelegationID:     iotago.EmptyDelegationID(),
					ValidatorAddress: tpkg.RandAccountAddress(),
					StartEpoch:       currentEpoch + 1,
					EndEpoch:         0,
					UnlockConditions: iotago.DelegationOutputUnlockConditions{
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
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Commitment: &iotago.Commitment{
						Slot: epochStartCommitmentIndex,
					},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrDelegationEndEpochInvalid,
		},
		{
			name: "fail - invalid post-registration slot transition - end epoch not set to expected epoch",
			input: &vm.ChainOutputWithIDs{
				ChainID:  exampleDelegationID,
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &iotago.DelegationOutput{
					Amount:           100,
					DelegatedAmount:  100,
					DelegationID:     iotago.EmptyDelegationID(),
					ValidatorAddress: tpkg.RandAccountAddress(),
					StartEpoch:       currentEpoch + 1,
					EndEpoch:         0,
					UnlockConditions: iotago.DelegationOutputUnlockConditions{
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
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Commitment: &iotago.Commitment{
						Slot: epochEndCommitmentIndex,
					},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrDelegationEndEpochInvalid,
		},
		{
			name: "fail - invalid transition - cannot claim rewards during transition",
			input: &vm.ChainOutputWithIDs{
				ChainID:  exampleDelegationID,
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &iotago.DelegationOutput{
					Amount:           100,
					DelegatedAmount:  100,
					DelegationID:     iotago.EmptyDelegationID(),
					ValidatorAddress: tpkg.RandAccountAddress(),
					StartEpoch:       currentEpoch + 1,
					EndEpoch:         0,
					UnlockConditions: iotago.DelegationOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			next:      &iotago.DelegationOutput{},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Commitment: &iotago.Commitment{
						Slot: epochStartCommitmentIndex,
					},
					Rewards: map[iotago.ChainID]iotago.Mana{
						exampleDelegationID: 1,
					},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrDelegationRewardsClaimingInvalid,
		},
		{
			name: "ok - valid destruction",
			input: &vm.ChainOutputWithIDs{
				ChainID:  exampleDelegationID,
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &iotago.DelegationOutput{
					Amount:           100,
					DelegatedAmount:  100,
					DelegationID:     iotago.EmptyDelegationID(),
					ValidatorAddress: tpkg.RandAccountAddress(),
					StartEpoch:       currentEpoch + 1,
					EndEpoch:         0,
					UnlockConditions: iotago.DelegationOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			nextMut:   nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Commitment: &iotago.Commitment{
						Slot: epochStartCommitmentIndex,
					},
					Rewards: map[iotago.ChainID]iotago.Mana{
						exampleDelegationID: 0,
					},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - invalid destruction - missing reward input",
			input: &vm.ChainOutputWithIDs{
				ChainID:  exampleDelegationID,
				OutputID: tpkg.RandOutputIDWithCreationSlot(0, 0),
				Output: &iotago.DelegationOutput{
					Amount:           100,
					DelegatedAmount:  100,
					DelegationID:     iotago.EmptyDelegationID(),
					ValidatorAddress: tpkg.RandAccountAddress(),
					StartEpoch:       currentEpoch + 1,
					EndEpoch:         0,
					UnlockConditions: iotago.DelegationOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
					},
				},
			},
			nextMut:   nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Commitment: &iotago.Commitment{
						Slot: epochStartCommitmentIndex,
					},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrDelegationRewardInputMissing,
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
				UnlockConditions: iotago.DelegationOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
			},
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Rewards: map[iotago.ChainID]iotago.Mana{
						exampleDelegationID: 0,
					},
					Tx: &iotago.Transaction{
						TransactionEssence: &iotago.TransactionEssence{
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrDelegationCommitmentInputMissing,
		},
	}

	for _, tt := range tests {
		if tt.nextMut != nil {
			for mutName, muts := range tt.nextMut {
				t.Run(fmt.Sprintf("%s_%s", tt.name, mutName), func(t *testing.T) {
					cpy := copyObjectAndMutate(t, tt.input.Output, muts).(*iotago.DelegationOutput)

					createWorkingSet(t, tt.input, tt.svCtx.WorkingSet)

					err := novaVM.ChainSTVF(tt.svCtx, tt.transType, tt.input, cpy)
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
			createWorkingSet(t, tt.input, tt.svCtx.WorkingSet)

			err := novaVM.ChainSTVF(tt.svCtx, tt.transType, tt.input, tt.next)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestImplicitAccountOutput_ValidateStateTransition(t *testing.T) {
	exampleIssuer := tpkg.RandEd25519Address()
	exampleAccountID := tpkg.RandAccountAddress().AccountID()

	currentEpoch := iotago.EpochIndex(20)
	currentSlot := tpkg.ZeroCostTestAPI.TimeProvider().EpochStart(currentEpoch)
	blockIssuerPubKey := iotago.Ed25519PublicKeyHashBlockIssuerKeyFromPublicKey(tpkg.Rand32ByteArray())
	exampleBlockIssuerFeature := &iotago.BlockIssuerFeature{
		BlockIssuerKeys: iotago.NewBlockIssuerKeys(blockIssuerPubKey),
		ExpirySlot:      currentSlot + tpkg.ZeroCostTestAPI.ProtocolParameters().MaxCommittableAge(),
	}

	exampleBIC := map[iotago.AccountID]iotago.BlockIssuanceCredits{
		exampleAccountID: 100,
	}

	type test struct {
		name      string
		input     *vm.ChainOutputWithIDs
		next      *iotago.AccountOutput
		transType iotago.ChainTransitionType
		svCtx     *vm.Params
		wantErr   error
	}

	implicitAccountCreationAddr := iotago.ImplicitAccountCreationAddressFromPubKey(tpkg.RandEd25519Signature().PublicKey[:])
	exampleAmount := iotago.BaseToken(100_000)

	tests := []*test{
		{
			name: "ok - implicit account conversion transition",
			next: &iotago.AccountOutput{
				Amount:    exampleAmount,
				AccountID: exampleAccountID,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: iotago.AccountOutputFeatures{
					exampleBlockIssuerFeature,
				},
			},
			input: &vm.ChainOutputWithIDs{
				ChainID: exampleAccountID,
				Output: &vm.ImplicitAccountOutput{
					BasicOutput: &iotago.BasicOutput{
						Amount: exampleAmount,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{
								Address: implicitAccountCreationAddr,
							},
						},
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{

				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					BIC: exampleBIC,
					Commitment: &iotago.Commitment{
						Slot: currentSlot,
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - explicit account lacks block issuer feature",
			next: &iotago.AccountOutput{
				Amount:    exampleAmount,
				AccountID: exampleAccountID,
				UnlockConditions: iotago.AccountOutputUnlockConditions{
					&iotago.AddressUnlockCondition{Address: tpkg.RandEd25519Address()},
				},
				Features: iotago.AccountOutputFeatures{},
			},
			input: &vm.ChainOutputWithIDs{
				ChainID: exampleAccountID,
				Output: &vm.ImplicitAccountOutput{
					BasicOutput: &iotago.BasicOutput{
						Amount: exampleAmount,
						UnlockConditions: iotago.BasicOutputUnlockConditions{
							&iotago.AddressUnlockCondition{
								Address: implicitAccountCreationAddr,
							},
						},
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				API: tpkg.ZeroCostTestAPI,
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					BIC: exampleBIC,
					Commitment: &iotago.Commitment{
						Slot: currentSlot,
					},
					Tx: &iotago.Transaction{
						API: tpkg.ZeroCostTestAPI,
						TransactionEssence: &iotago.TransactionEssence{
							CreationSlot: currentSlot,
							Capabilities: iotago.TransactionCapabilitiesBitMaskWithCapabilities(iotago.WithTransactionCanDoAnything()),
						},
					},
				},
			},
			wantErr: iotago.ErrBlockIssuerNotExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createWorkingSet(t, tt.input, tt.svCtx.WorkingSet)

			err := novaVM.ChainSTVF(tt.svCtx, tt.transType, tt.input, tt.next)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func createWorkingSet(t *testing.T, input *vm.ChainOutputWithIDs, workingSet *vm.WorkingSet) {
	t.Helper()

	if input != nil {
		// create the working set for the test
		if workingSet.UTXOInputsSet == nil {
			workingSet.UTXOInputsSet = make(vm.InputSet)
		}
		workingSet.UTXOInputsSet[input.OutputID] = input.Output

		totalManaIn, err := vm.TotalManaIn(
			tpkg.ZeroCostTestAPI.ManaDecayProvider(),
			tpkg.ZeroCostTestAPI.StorageScoreStructure(),
			workingSet.Tx.CreationSlot,
			workingSet.UTXOInputsSet,
			workingSet.Rewards,
		)
		require.NoError(t, err)
		workingSet.TotalManaIn = totalManaIn

		totalManaOut, err := vm.TotalManaOut(
			workingSet.Tx.Outputs,
			workingSet.Tx.Allotments,
		)
		require.NoError(t, err)
		workingSet.TotalManaOut = totalManaOut
	}
}
