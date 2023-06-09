package stardust_test

import (
	"crypto/ed25519"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
	"github.com/iotaledger/iota.go/v4/vm"
)

var (
	v3API = iotago.V3API(tpkg.TestProtoParams)
)

type fieldMutations map[string]interface{}

func copyObject(t *testing.T, source any, mutations fieldMutations) any {
	srcBytes, err := v3API.Encode(source)
	require.NoError(t, err)

	ptrToCpyOfSrc := reflect.New(reflect.ValueOf(source).Elem().Type())

	cpySeri := ptrToCpyOfSrc.Interface()
	_, err = v3API.Decode(srcBytes, cpySeri)
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

	type test struct {
		name      string
		input     *iotago.ChainOutputWithCreationTime
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
						BlockIssuerKeys: []ed25519.PublicKey{tpkg.RandEd25519PrivateKey().Public().(ed25519.PublicKey)},
						ExpirySlot:      1000,
					},
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				External: &iotago.ExternalUnlockParameters{
					ProtocolParameters: &iotago.ProtocolParameters{MaxCommitableAge: 10},
				},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							CreationTime: 900,
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
						BlockIssuerKeys: []ed25519.PublicKey{tpkg.RandEd25519PrivateKey().Public().(ed25519.PublicKey)},
						ExpirySlot:      1000,
					},
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				External: &iotago.ExternalUnlockParameters{
					ProtocolParameters: &iotago.ProtocolParameters{MaxCommitableAge: 10},
				},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							CreationTime: 10001,
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
						BlockIssuerKeys: []ed25519.PublicKey{tpkg.RandEd25519PrivateKey().Public().(ed25519.PublicKey)},
						ExpirySlot:      1000,
					},
				},
			},
			input:     nil,
			transType: iotago.ChainTransitionTypeGenesis,
			svCtx: &vm.Params{
				External: &iotago.ExternalUnlockParameters{
					ProtocolParameters: &iotago.ProtocolParameters{MaxCommitableAge: 10},
				},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleIssuer.Key(): {UnlockedAt: 0},
					},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							CreationTime: 991,
						},
					},
				},
			},
			wantErr: iotago.ErrInvalidBlockIssuerTransition,
		},
		{
			name: "ok - destroy transition",
			input: &iotago.ChainOutputWithCreationTime{
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
				External: &iotago.ExternalUnlockParameters{},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - destroy block issuer account with negative BIC",
			input: &iotago.ChainOutputWithCreationTime{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: []ed25519.PublicKey{tpkg.RandEd25519PrivateKey().Public().(ed25519.PublicKey)},
							ExpirySlot:      1000,
						},
					},
				},
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				External: &iotago.ExternalUnlockParameters{},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							CreationTime: 1001,
						},
					},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredit{
						exampleAccountID: {
							AccountID:    exampleAccountID,
							CommitmentID: iotago.CommitmentID{},
							Value:        -1,
						},
					},
				},
			},
			wantErr: iotago.ErrInvalidBlockIssuerTransition,
		},
		{
			name: "fail - destroy block issuer account no BIC provided",
			input: &iotago.ChainOutputWithCreationTime{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: []ed25519.PublicKey{tpkg.RandEd25519PrivateKey().Public().(ed25519.PublicKey)},
							ExpirySlot:      1000,
						},
					},
				},
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				External: &iotago.ExternalUnlockParameters{},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							CreationTime: 1001,
						},
					},
				},
			},
			wantErr: iotago.ErrInvalidBlockIssuerTransition,
		},
		{
			name: "fail - non-expired block issuer destroy transition",
			input: &iotago.ChainOutputWithCreationTime{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: tpkg.RandAccountAddress().AccountID(),
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: []ed25519.PublicKey{tpkg.RandEd25519PrivateKey().Public().(ed25519.PublicKey)},
							ExpirySlot:      1000,
						},
					},
				},
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				External: &iotago.ExternalUnlockParameters{},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							CreationTime: 1000,
						},
					},
				},
			},
			wantErr: iotago.ErrInvalidBlockIssuerTransition,
		},
		{
			name: "ok - expired block issuer destroy transition",
			input: &iotago.ChainOutputWithCreationTime{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: []ed25519.PublicKey{tpkg.RandEd25519PrivateKey().Public().(ed25519.PublicKey)},
							ExpirySlot:      1000,
						},
					},
				},
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				External: &iotago.ExternalUnlockParameters{},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredit{
						exampleAccountID: {
							AccountID:    exampleAccountID,
							CommitmentID: iotago.CommitmentID{},
							Value:        0,
						},
					},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							Inputs:       nil,
							CreationTime: 1001,
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - gov transition",
			input: &iotago.ChainOutputWithCreationTime{
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
						BlockIssuerKeys: []ed25519.PublicKey{tpkg.RandEd25519PrivateKey().Public().(ed25519.PublicKey)},
						ExpirySlot:      1000,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
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
			input: &iotago.ChainOutputWithCreationTime{
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
				External: &iotago.ExternalUnlockParameters{},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleStateCtrl.Key(): {UnlockedAt: 0},
					},
					InChains: map[iotago.ChainID]*iotago.ChainOutputWithCreationTime{
						// serial number 5
						exampleExistingFoundryOutputID: {
							Output: exampleExistingFoundryOutput,
						},
					},
					Tx: &iotago.Transaction{
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
			name: "fail - update expired account without extending expiration after MCA",
			input: &iotago.ChainOutputWithCreationTime{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: []ed25519.PublicKey{tpkg.RandEd25519PrivateKey().Public().(ed25519.PublicKey)},
							ExpirySlot:      1000,
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
				Features: iotago.AccountOutputFeatures{
					&iotago.SenderFeature{Address: exampleStateCtrl},
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: []ed25519.PublicKey{tpkg.RandEd25519PrivateKey().Public().(ed25519.PublicKey)},
						ExpirySlot:      1000,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				External: &iotago.ExternalUnlockParameters{
					DecayProvider:      iotago.NewDecayProvider(0, []float64{}, []float64{}),
					ProtocolParameters: &iotago.ProtocolParameters{MaxCommitableAge: 10},
				},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleStateCtrl.Key(): {UnlockedAt: 0},
					},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredit{
						exampleAccountID: {
							AccountID:    exampleAccountID,
							CommitmentID: iotago.CommitmentID{},
							Value:        10,
						},
					},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							Inputs:       nil,
							CreationTime: 990,
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - update account immutable features",
			input: &iotago.ChainOutputWithCreationTime{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					ImmutableFeatures: iotago.AccountOutputImmFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: []ed25519.PublicKey{tpkg.RandEd25519PrivateKey().Public().(ed25519.PublicKey)},
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
						BlockIssuerKeys: []ed25519.PublicKey{tpkg.RandEd25519PrivateKey().Public().(ed25519.PublicKey)},
						ExpirySlot:      999,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				External: &iotago.ExternalUnlockParameters{
					DecayProvider:      iotago.NewDecayProvider(0, []float64{}, []float64{}),
					ProtocolParameters: &iotago.ProtocolParameters{MaxCommitableAge: 10},
				},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleStateCtrl.Key(): {UnlockedAt: 0},
					},
				},
			},
			wantErr: iotago.ErrInvalidAccountStateTransition,
		},
		{
			name: "fail - update expired account with extending expiration before MCA",
			input: &iotago.ChainOutputWithCreationTime{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: []ed25519.PublicKey{tpkg.RandEd25519PrivateKey().Public().(ed25519.PublicKey)},
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
				Features: iotago.AccountOutputFeatures{
					&iotago.SenderFeature{Address: exampleStateCtrl},
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: []ed25519.PublicKey{tpkg.RandEd25519PrivateKey().Public().(ed25519.PublicKey)},
						ExpirySlot:      999,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				External: &iotago.ExternalUnlockParameters{
					DecayProvider:      iotago.NewDecayProvider(0, []float64{}, []float64{}),
					ProtocolParameters: &iotago.ProtocolParameters{MaxCommitableAge: 10},
				},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleStateCtrl.Key(): {UnlockedAt: 0},
					},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredit{
						exampleAccountID: {
							AccountID:    exampleAccountID,
							CommitmentID: iotago.CommitmentID{},
							Value:        10,
						},
					},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							Inputs:       nil,
							CreationTime: 990,
						},
					},
				},
			},
			wantErr: iotago.ErrInvalidBlockIssuerTransition,
		},
		{
			name: "fail - update expired account with extending expiration to the past before MCA",
			input: &iotago.ChainOutputWithCreationTime{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: []ed25519.PublicKey{tpkg.RandEd25519PrivateKey().Public().(ed25519.PublicKey)},
							ExpirySlot:      1100,
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
				Features: iotago.AccountOutputFeatures{
					&iotago.SenderFeature{Address: exampleStateCtrl},
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: []ed25519.PublicKey{tpkg.RandEd25519PrivateKey().Public().(ed25519.PublicKey)},
						ExpirySlot:      999,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				External: &iotago.ExternalUnlockParameters{
					DecayProvider:      iotago.NewDecayProvider(0, []float64{}, []float64{}),
					ProtocolParameters: &iotago.ProtocolParameters{MaxCommitableAge: 10},
				},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleStateCtrl.Key(): {UnlockedAt: 0},
					},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredit{
						exampleAccountID: {
							AccountID:    exampleAccountID,
							CommitmentID: iotago.CommitmentID{},
							Value:        10,
						},
					},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							Inputs:       nil,
							CreationTime: 990,
						},
					},
				},
			},
			wantErr: iotago.ErrInvalidBlockIssuerTransition,
		},
		{
			name: "fail - update block issuer account with negative BIC",
			input: &iotago.ChainOutputWithCreationTime{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: []ed25519.PublicKey{tpkg.RandEd25519PrivateKey().Public().(ed25519.PublicKey)},
							ExpirySlot:      1000,
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
				Features: iotago.AccountOutputFeatures{
					&iotago.SenderFeature{Address: exampleStateCtrl},
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: []ed25519.PublicKey{tpkg.RandEd25519PrivateKey().Public().(ed25519.PublicKey)},
						ExpirySlot:      1000,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				External: &iotago.ExternalUnlockParameters{
					DecayProvider:      iotago.NewDecayProvider(0, []float64{}, []float64{}),
					ProtocolParameters: &iotago.ProtocolParameters{MaxCommitableAge: 10},
				},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleStateCtrl.Key(): {UnlockedAt: 0},
					},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredit{
						exampleAccountID: {
							AccountID:    exampleAccountID,
							CommitmentID: iotago.CommitmentID{},
							Value:        -1,
						},
					},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							Inputs:       nil,
							CreationTime: 900,
						},
					},
				},
			},
			wantErr: iotago.ErrInvalidBlockIssuerTransition,
		},
		{
			name: "fail - update block issuer account without BIC provided",
			input: &iotago.ChainOutputWithCreationTime{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: []ed25519.PublicKey{tpkg.RandEd25519PrivateKey().Public().(ed25519.PublicKey)},
							ExpirySlot:      1000,
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
				Features: iotago.AccountOutputFeatures{
					&iotago.SenderFeature{Address: exampleStateCtrl},
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: []ed25519.PublicKey{tpkg.RandEd25519PrivateKey().Public().(ed25519.PublicKey)},
						ExpirySlot:      1000,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				External: &iotago.ExternalUnlockParameters{
					DecayProvider:      iotago.NewDecayProvider(0, []float64{}, []float64{}),
					ProtocolParameters: &iotago.ProtocolParameters{MaxCommitableAge: 10},
				},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleStateCtrl.Key(): {UnlockedAt: 0},
					},

					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							CreationTime: 900,
						},
					},
				},
			},
			wantErr: iotago.ErrInvalidBlockIssuerTransition,
		},
		{
			name: "ok - update expiration to earlier slot",
			input: &iotago.ChainOutputWithCreationTime{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: []ed25519.PublicKey{tpkg.RandEd25519PrivateKey().Public().(ed25519.PublicKey)},
							ExpirySlot:      1000,
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
				Features: iotago.AccountOutputFeatures{
					&iotago.SenderFeature{Address: exampleStateCtrl},
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: []ed25519.PublicKey{tpkg.RandEd25519PrivateKey().Public().(ed25519.PublicKey)},
						ExpirySlot:      999,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				External: &iotago.ExternalUnlockParameters{
					DecayProvider:      iotago.NewDecayProvider(0, []float64{}, []float64{}),
					ProtocolParameters: &iotago.ProtocolParameters{MaxCommitableAge: 10},
				},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleStateCtrl.Key(): {UnlockedAt: 0},
					},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredit{
						exampleAccountID: {
							AccountID:    exampleAccountID,
							CommitmentID: iotago.CommitmentID{},
							Value:        10,
						},
					},
					Tx: &iotago.Transaction{
						Essence: &iotago.TransactionEssence{
							CreationTime: 900,
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - non-expired block issuer replace key",
			input: &iotago.ChainOutputWithCreationTime{
				Output: &iotago.AccountOutput{
					Amount:    100,
					AccountID: exampleAccountID,
					Conditions: iotago.AccountOutputUnlockConditions{
						&iotago.StateControllerAddressUnlockCondition{Address: exampleStateCtrl},
						&iotago.GovernorAddressUnlockCondition{Address: exampleGovCtrl},
					},
					Features: iotago.AccountOutputFeatures{
						&iotago.BlockIssuerFeature{
							BlockIssuerKeys: []ed25519.PublicKey{tpkg.RandEd25519PrivateKey().Public().(ed25519.PublicKey)},
							ExpirySlot:      1000,
						},
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
					&iotago.BlockIssuerFeature{
						BlockIssuerKeys: []ed25519.PublicKey{tpkg.RandEd25519PrivateKey().Public().(ed25519.PublicKey)},
						ExpirySlot:      1000,
					},
				},
			},
			transType: iotago.ChainTransitionTypeStateChange,
			svCtx: &vm.Params{
				External: &iotago.ExternalUnlockParameters{
					DecayProvider: iotago.NewDecayProvider(0, []float64{}, []float64{}),
				},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{
						exampleStateCtrl.Key(): {UnlockedAt: 0},
					},
					InChains: map[iotago.ChainID]*iotago.ChainOutputWithCreationTime{
						// serial number 5
						exampleExistingFoundryOutputID: {
							Output: exampleExistingFoundryOutput,
						},
					},
					BIC: map[iotago.AccountID]iotago.BlockIssuanceCredit{
						exampleAccountID: {
							AccountID:    exampleAccountID,
							CommitmentID: iotago.CommitmentID{},
							Value:        10,
						},
					},
					Tx: &iotago.Transaction{
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
			input: &iotago.ChainOutputWithCreationTime{
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
			svCtx: &vm.Params{
				External: &iotago.ExternalUnlockParameters{},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
				},
			},
			wantErr: iotago.ErrInvalidAccountGovernanceTransition,
		},
		{
			name: "fail - state transition",
			input: &iotago.ChainOutputWithCreationTime{
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
				External: &iotago.ExternalUnlockParameters{},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
					InChains:       iotago.ChainInputSet{},
					Tx: &iotago.Transaction{
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
		input     *iotago.ChainOutputWithCreationTime
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
						Essence: &iotago.TransactionEssence{
							Outputs: iotago.TxEssenceOutputs{exampleFoundry},
						},
						Unlocks: nil,
					},
					InChains: iotago.ChainInputSet{
						exampleAccountIdent.AccountID(): &iotago.ChainOutputWithCreationTime{
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
						Essence: &iotago.TransactionEssence{
							Outputs: iotago.TxEssenceOutputs{exampleFoundry},
						},
						Unlocks: nil,
					},
					InChains: iotago.ChainInputSet{
						exampleAccountIdent.AccountID(): &iotago.ChainOutputWithCreationTime{
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
						Essence: &iotago.TransactionEssence{
							Outputs: iotago.TxEssenceOutputs{exampleFoundry},
						},
						Unlocks: nil,
					},
					InChains: iotago.ChainInputSet{
						exampleAccountIdent.AccountID(): &iotago.ChainOutputWithCreationTime{
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
					InChains: iotago.ChainInputSet{
						exampleAccountIdent.AccountID(): &iotago.ChainOutputWithCreationTime{
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
			input: &iotago.ChainOutputWithCreationTime{
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
			input: &iotago.ChainOutputWithCreationTime{
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
			input: &iotago.ChainOutputWithCreationTime{
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
			input: &iotago.ChainOutputWithCreationTime{
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
			input: &iotago.ChainOutputWithCreationTime{
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
			input: &iotago.ChainOutputWithCreationTime{
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
			input: &iotago.ChainOutputWithCreationTime{
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
			input: &iotago.ChainOutputWithCreationTime{
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
			input: &iotago.ChainOutputWithCreationTime{
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
			input: &iotago.ChainOutputWithCreationTime{
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
			input: &iotago.ChainOutputWithCreationTime{
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
		input     *iotago.ChainOutputWithCreationTime
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
			input: &iotago.ChainOutputWithCreationTime{
				Output: exampleCurrentNFTOutput,
			},
			next:      nil,
			transType: iotago.ChainTransitionTypeDestroy,
			svCtx: &vm.Params{
				External: &iotago.ExternalUnlockParameters{},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - state transition",
			input: &iotago.ChainOutputWithCreationTime{
				Output: exampleCurrentNFTOutput,
			},
			nextMut: map[string]fieldMutations{
				"amount": {
					"Amount": uint64(1337),
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
				External: &iotago.ExternalUnlockParameters{},
				WorkingSet: &vm.WorkingSet{
					UnlockedIdents: vm.UnlockedIdentities{},
				},
			},
			wantErr: nil,
		},
		{
			name: "fail - state transition",
			input: &iotago.ChainOutputWithCreationTime{
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
					cpy := copyObject(t, tt.input.Output, muts).(*iotago.NFTOutput)
					err := stardustVM.ChainSTVF(tt.transType, tt.input, cpy, tt.svCtx)
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
			err := stardustVM.ChainSTVF(tt.transType, tt.input, tt.next, tt.svCtx)
			if tt.wantErr != nil {
				require.ErrorAs(t, err, &tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
