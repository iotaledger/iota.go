//nolint:dupl
package iotago

import (
	"context"
	"time"

	"github.com/iotaledger/hive.go/core/safemath"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
)

const (
	apiV3Version = 3
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func disallowImplicitAccountCreationAddress(address Address) error {
	if address.Type() == AddressImplicitAccountCreation {
		return ErrImplicitAccountCreationAddressInInvalidUnlockCondition
	}

	return nil
}

var (
	// Note that when UniquenessSliceFunc is set and the mode is no dups and lexical order, then both will use
	// the return value of UniquenessSliceFunc for those checks.
	nativeTokensV3ArrRules = &serix.ArrayRules{
		Min: MinNativeTokenCountPerOutput,
		Max: MaxNativeTokenCountPerOutput,
		// Uniqueness and lexical order is checked based on the Token ID.
		UniquenessSliceFunc: func(next []byte) []byte { return next[:NativeTokenIDLength] },
		ValidationMode:      serializer.ArrayValidationModeNoDuplicates | serializer.ArrayValidationModeLexicalOrdering,
	}

	basicOutputV3UnlockCondArrRules = &serix.ArrayRules{
		Min: 1,
		Max: 4,
		MustOccur: serializer.TypePrefixes{
			uint32(UnlockConditionAddress): struct{}{},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}
	basicOutputV3FeatBlocksArrRules = &serix.ArrayRules{
		Min: 0,
		Max: 8,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	accountOutputV3UnlockCondArrRules = &serix.ArrayRules{
		Min: 2, Max: 2,
		MustOccur: serializer.TypePrefixes{
			uint32(UnlockConditionStateControllerAddress): struct{}{},
			uint32(UnlockConditionGovernorAddress):        struct{}{},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	accountOutputV3FeatBlocksArrRules = &serix.ArrayRules{
		Min: 0,
		Max: 3,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	accountOutputV3BlockIssuerKeysArrRules = &serix.ArrayRules{
		Min: MinBlockIssuerKeysCount,
		Max: MaxBlockIssuerKeysCount,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering,
	}

	accountOutputV3ImmFeatBlocksArrRules = &serix.ArrayRules{
		Min: 0,
		Max: 2,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	foundryOutputV3UnlockCondArrRules = &serix.ArrayRules{
		Min: 1, Max: 1,
		MustOccur: serializer.TypePrefixes{
			uint32(UnlockConditionImmutableAccount): struct{}{},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	foundryOutputV3FeatBlocksArrRules = &serix.ArrayRules{
		Min: 0, Max: 1,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	foundryOutputV3ImmFeatBlocksArrRules = &serix.ArrayRules{
		Min: 0, Max: 1,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	nftOutputV3UnlockCondArrRules = &serix.ArrayRules{
		Min: 1, Max: 4,
		MustOccur: serializer.TypePrefixes{
			uint32(UnlockConditionAddress): struct{}{},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	nftOutputV3FeatBlocksArrRules = &serix.ArrayRules{
		Min: 0,
		Max: 3,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	nftOutputV3ImmFeatBlocksArrRules = &serix.ArrayRules{
		Min: 0,
		Max: 2,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	delegationOutputV3UnlockCondArrRules = &serix.ArrayRules{
		Min: 1, Max: 1,
		MustOccur: serializer.TypePrefixes{
			uint32(UnlockConditionAddress): struct{}{},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	txEssenceV3ContextInputsArrRules = &serix.ArrayRules{
		Min:            MinContextInputsCount,
		Max:            MaxContextInputsCount,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates,
	}

	txEssenceV3InputsArrRules = &serix.ArrayRules{
		Min:            MinInputsCount,
		Max:            MaxInputsCount,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates,
	}

	txEssenceV3OutputsArrRules = &serix.ArrayRules{
		Min:            MinOutputsCount,
		Max:            MaxOutputsCount,
		ValidationMode: serializer.ArrayValidationModeNone,
	}

	txEssenceV3AllotmentsArrRules = &serix.ArrayRules{
		Min: MinAllotmentCount,
		Max: MaxAllotmentCount,
		// Uniqueness and lexical order is checked based on the Account ID.
		UniquenessSliceFunc: func(next []byte) []byte { return next[:AccountIDLength] },
		ValidationMode:      serializer.ArrayValidationModeNoDuplicates | serializer.ArrayValidationModeLexicalOrdering,
	}

	txV3UnlocksArrRules = &serix.ArrayRules{
		Min: 1, Max: MaxInputsCount,
	}

	blockIDsArrRules = &serix.ArrayRules{
		ValidationMode: serializer.ArrayValidationModeNoDuplicates | serializer.ArrayValidationModeLexicalOrdering,
	}
)

// v3api implements the iota-core 1.0 protocol core models.
type v3api struct {
	serixAPI *serix.API

	protocolParameters        *V3ProtocolParameters
	timeProvider              *TimeProvider
	manaDecayProvider         *ManaDecayProvider
	livenessThresholdDuration time.Duration
	maxBlockWork              WorkScore
	computedInitialReward     uint64
	computedFinalReward       uint64
}

func (v *v3api) JSONEncode(obj any, opts ...serix.Option) ([]byte, error) {
	return v.serixAPI.JSONEncode(context.TODO(), obj, opts...)
}

func (v *v3api) JSONDecode(jsonData []byte, obj any, opts ...serix.Option) error {
	return v.serixAPI.JSONDecode(context.TODO(), jsonData, obj, opts...)
}

func (v *v3api) Underlying() *serix.API {
	return v.serixAPI
}

func (v *v3api) Version() Version {
	return v.protocolParameters.Version()
}

func (v *v3api) ProtocolParameters() ProtocolParameters {
	return v.protocolParameters
}

func (v *v3api) TimeProvider() *TimeProvider {
	return v.timeProvider
}

func (v *v3api) ManaDecayProvider() *ManaDecayProvider {
	return v.manaDecayProvider
}

func (v *v3api) LivenessThresholdDuration() time.Duration {
	return v.livenessThresholdDuration
}

func (v *v3api) MaxBlockWork() WorkScore {
	return v.maxBlockWork
}

func (v *v3api) ComputedInitialReward() uint64 {
	return v.computedInitialReward
}

func (v *v3api) ComputedFinalReward() uint64 {
	return v.computedFinalReward
}

func (v *v3api) Encode(obj interface{}, opts ...serix.Option) ([]byte, error) {
	return v.serixAPI.Encode(context.TODO(), obj, opts...)
}

func (v *v3api) Decode(b []byte, obj interface{}, opts ...serix.Option) (int, error) {
	return v.serixAPI.Decode(context.TODO(), b, obj, opts...)
}

// V3API instantiates an API instance with types registered conforming to protocol version 3 (iota-core 1.0) of the IOTA protocol.
func V3API(protoParams ProtocolParameters) API {
	api := CommonSerixAPI()

	timeProvider := protoParams.TimeProvider()

	maxBlockWork, err := protoParams.WorkScoreStructure().MaxBlockWork()
	must(err)

	initialReward, finalReward, err := calculateRewards(protoParams)
	must(err)

	//nolint:forcetypeassert // we can safely assume that these are V3ProtocolParameters
	v3 := &v3api{
		serixAPI:                  api,
		protocolParameters:        protoParams.(*V3ProtocolParameters),
		timeProvider:              timeProvider,
		manaDecayProvider:         protoParams.ManaDecayProvider(),
		livenessThresholdDuration: time.Duration(uint64(protoParams.LivenessThreshold())*uint64(timeProvider.SlotDurationSeconds())) * time.Second,
		maxBlockWork:              maxBlockWork,
		computedInitialReward:     initialReward,
		computedFinalReward:       finalReward,
	}

	must(api.RegisterTypeSettings(TaggedData{},
		serix.TypeSettings{}.WithObjectType(uint32(PayloadTaggedData))),
	)

	{
		must(api.RegisterTypeSettings(Ed25519Signature{},
			serix.TypeSettings{}.WithObjectType(uint8(SignatureEd25519))),
		)
		must(api.RegisterInterfaceObjects((*Signature)(nil), (*Ed25519Signature)(nil)))
	}

	{
		must(api.RegisterTypeSettings(IssuerFeature{},
			serix.TypeSettings{}.WithObjectType(uint8(FeatureIssuer))),
		)
		must(api.RegisterTypeSettings(MetadataFeature{},
			serix.TypeSettings{}.WithObjectType(uint8(FeatureMetadata))),
		)
		must(api.RegisterTypeSettings(SenderFeature{},
			serix.TypeSettings{}.WithObjectType(uint8(FeatureSender))),
		)
		must(api.RegisterTypeSettings(TagFeature{},
			serix.TypeSettings{}.WithObjectType(uint8(FeatureTag))),
		)
		must(api.RegisterTypeSettings(BlockIssuerFeature{},
			serix.TypeSettings{}.WithObjectType(uint8(FeatureBlockIssuer))),
		)
		must(api.RegisterTypeSettings(StakingFeature{},
			serix.TypeSettings{}.WithObjectType(uint8(FeatureStaking))),
		)
		must(api.RegisterInterfaceObjects((*Feature)(nil), (*IssuerFeature)(nil)))
		must(api.RegisterInterfaceObjects((*Feature)(nil), (*MetadataFeature)(nil)))
		must(api.RegisterInterfaceObjects((*Feature)(nil), (*SenderFeature)(nil)))
		must(api.RegisterInterfaceObjects((*Feature)(nil), (*TagFeature)(nil)))
		must(api.RegisterInterfaceObjects((*Feature)(nil), (*BlockIssuerFeature)(nil)))
		must(api.RegisterInterfaceObjects((*Feature)(nil), (*StakingFeature)(nil)))
	}

	{
		must(api.RegisterTypeSettings(AddressUnlockCondition{},
			serix.TypeSettings{}.WithObjectType(uint8(UnlockConditionAddress))),
		)
		must(api.RegisterTypeSettings(StorageDepositReturnUnlockCondition{},
			serix.TypeSettings{}.WithObjectType(uint8(UnlockConditionStorageDepositReturn))),
		)
		must(api.RegisterValidators(StorageDepositReturnUnlockCondition{}, nil,
			func(ctx context.Context, sdruc StorageDepositReturnUnlockCondition) error {
				return disallowImplicitAccountCreationAddress(sdruc.ReturnAddress)
			},
		))
		must(api.RegisterTypeSettings(TimelockUnlockCondition{},
			serix.TypeSettings{}.WithObjectType(uint8(UnlockConditionTimelock))),
		)
		must(api.RegisterTypeSettings(ExpirationUnlockCondition{},
			serix.TypeSettings{}.WithObjectType(uint8(UnlockConditionExpiration))),
		)
		must(api.RegisterValidators(ExpirationUnlockCondition{}, nil,
			func(ctx context.Context, exp ExpirationUnlockCondition) error {
				return disallowImplicitAccountCreationAddress(exp.ReturnAddress)
			},
		))
		must(api.RegisterTypeSettings(StateControllerAddressUnlockCondition{},
			serix.TypeSettings{}.WithObjectType(uint8(UnlockConditionStateControllerAddress))),
		)
		must(api.RegisterValidators(StateControllerAddressUnlockCondition{}, nil,
			func(ctx context.Context, stateController StateControllerAddressUnlockCondition) error {
				return disallowImplicitAccountCreationAddress(stateController.Address)
			},
		))
		must(api.RegisterTypeSettings(GovernorAddressUnlockCondition{},
			serix.TypeSettings{}.WithObjectType(uint8(UnlockConditionGovernorAddress))),
		)
		must(api.RegisterValidators(GovernorAddressUnlockCondition{}, nil,
			func(ctx context.Context, gov GovernorAddressUnlockCondition) error {
				return disallowImplicitAccountCreationAddress(gov.Address)
			},
		))
		must(api.RegisterTypeSettings(ImmutableAccountUnlockCondition{},
			serix.TypeSettings{}.WithObjectType(uint8(UnlockConditionImmutableAccount))),
		)
		must(api.RegisterInterfaceObjects((*UnlockCondition)(nil), (*AddressUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*UnlockCondition)(nil), (*StorageDepositReturnUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*UnlockCondition)(nil), (*TimelockUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*UnlockCondition)(nil), (*ExpirationUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*UnlockCondition)(nil), (*StateControllerAddressUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*UnlockCondition)(nil), (*GovernorAddressUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*UnlockCondition)(nil), (*ImmutableAccountUnlockCondition)(nil)))
	}

	{
		must(api.RegisterTypeSettings(SignatureUnlock{}, serix.TypeSettings{}.WithObjectType(uint8(UnlockSignature))))
		must(api.RegisterTypeSettings(ReferenceUnlock{}, serix.TypeSettings{}.WithObjectType(uint8(UnlockReference))))
		must(api.RegisterTypeSettings(AccountUnlock{}, serix.TypeSettings{}.WithObjectType(uint8(UnlockAccount))))
		must(api.RegisterTypeSettings(NFTUnlock{}, serix.TypeSettings{}.WithObjectType(uint8(UnlockNFT))))
		must(api.RegisterTypeSettings(MultiUnlock{}, serix.TypeSettings{}.WithObjectType(uint8(UnlockMulti))))
		must(api.RegisterTypeSettings(EmptyUnlock{}, serix.TypeSettings{}.WithObjectType(uint8(UnlockEmpty))))
		must(api.RegisterInterfaceObjects((*Unlock)(nil), (*SignatureUnlock)(nil)))
		must(api.RegisterInterfaceObjects((*Unlock)(nil), (*ReferenceUnlock)(nil)))
		must(api.RegisterInterfaceObjects((*Unlock)(nil), (*AccountUnlock)(nil)))
		must(api.RegisterInterfaceObjects((*Unlock)(nil), (*NFTUnlock)(nil)))
		must(api.RegisterInterfaceObjects((*Unlock)(nil), (*MultiUnlock)(nil)))
		must(api.RegisterInterfaceObjects((*Unlock)(nil), (*EmptyUnlock)(nil)))
	}

	{
		must(api.RegisterTypeSettings(NativeToken{}, serix.TypeSettings{}))
		must(api.RegisterTypeSettings(NativeTokens{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(nativeTokensV3ArrRules),
		))
	}

	{
		must(api.RegisterTypeSettings(BasicOutput{}, serix.TypeSettings{}.WithObjectType(uint8(OutputBasic))))

		must(api.RegisterTypeSettings(BasicOutputUnlockConditions{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(basicOutputV3UnlockCondArrRules),
		))

		must(api.RegisterInterfaceObjects((*basicOutputUnlockCondition)(nil), (*AddressUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*basicOutputUnlockCondition)(nil), (*StorageDepositReturnUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*basicOutputUnlockCondition)(nil), (*TimelockUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*basicOutputUnlockCondition)(nil), (*ExpirationUnlockCondition)(nil)))

		must(api.RegisterTypeSettings(BasicOutputFeatures{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(basicOutputV3FeatBlocksArrRules),
		))

		must(api.RegisterInterfaceObjects((*basicOutputFeature)(nil), (*SenderFeature)(nil)))
		must(api.RegisterInterfaceObjects((*basicOutputFeature)(nil), (*MetadataFeature)(nil)))
		must(api.RegisterInterfaceObjects((*basicOutputFeature)(nil), (*TagFeature)(nil)))
	}

	{
		must(api.RegisterTypeSettings(AccountOutput{}, serix.TypeSettings{}.WithObjectType(uint8(OutputAccount))))
		must(api.RegisterValidators(AccountOutput{}, nil, func(ctx context.Context, account AccountOutput) error {
			return account.syntacticallyValidate()
		}))

		must(api.RegisterTypeSettings(AccountOutputUnlockConditions{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(accountOutputV3UnlockCondArrRules),
		))

		must(api.RegisterInterfaceObjects((*accountOutputUnlockCondition)(nil), (*StateControllerAddressUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*accountOutputUnlockCondition)(nil), (*GovernorAddressUnlockCondition)(nil)))

		must(api.RegisterTypeSettings(AccountOutputFeatures{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(accountOutputV3FeatBlocksArrRules),
		))

		must(api.RegisterInterfaceObjects((*accountOutputFeature)(nil), (*SenderFeature)(nil)))
		must(api.RegisterInterfaceObjects((*accountOutputFeature)(nil), (*MetadataFeature)(nil)))
		must(api.RegisterInterfaceObjects((*accountOutputFeature)(nil), (*BlockIssuerFeature)(nil)))
		must(api.RegisterInterfaceObjects((*accountOutputFeature)(nil), (*StakingFeature)(nil)))

		must(api.RegisterTypeSettings(AccountOutputImmFeatures{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(accountOutputV3ImmFeatBlocksArrRules),
		))

		must(api.RegisterInterfaceObjects((*accountOutputImmFeature)(nil), (*IssuerFeature)(nil)))
		must(api.RegisterInterfaceObjects((*accountOutputImmFeature)(nil), (*MetadataFeature)(nil)))
	}

	{
		must(api.RegisterTypeSettings(FoundryOutput{},
			serix.TypeSettings{}.WithObjectType(uint8(OutputFoundry))),
		)

		must(api.RegisterTypeSettings(FoundryOutputUnlockConditions{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(foundryOutputV3UnlockCondArrRules),
		))

		must(api.RegisterInterfaceObjects((*foundryOutputUnlockCondition)(nil), (*ImmutableAccountUnlockCondition)(nil)))

		must(api.RegisterTypeSettings(FoundryOutputFeatures{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(foundryOutputV3FeatBlocksArrRules),
		))

		must(api.RegisterInterfaceObjects((*foundryOutputFeature)(nil), (*MetadataFeature)(nil)))

		must(api.RegisterTypeSettings(FoundryOutputImmFeatures{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(foundryOutputV3ImmFeatBlocksArrRules),
		))

		must(api.RegisterInterfaceObjects((*foundryOutputImmFeature)(nil), (*MetadataFeature)(nil)))

		must(api.RegisterTypeSettings(SimpleTokenScheme{}, serix.TypeSettings{}.WithObjectType(uint8(TokenSchemeSimple))))
		must(api.RegisterInterfaceObjects((*TokenScheme)(nil), (*SimpleTokenScheme)(nil)))
	}

	{
		must(api.RegisterTypeSettings(NFTOutput{},
			serix.TypeSettings{}.WithObjectType(uint8(OutputNFT))),
		)
		must(api.RegisterValidators(NFTOutput{}, nil, func(ctx context.Context, nft NFTOutput) error {
			return nft.syntacticallyValidate()
		}))

		must(api.RegisterTypeSettings(NFTOutputUnlockConditions{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(nftOutputV3UnlockCondArrRules),
		))

		must(api.RegisterInterfaceObjects((*nftOutputUnlockCondition)(nil), (*AddressUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*nftOutputUnlockCondition)(nil), (*StorageDepositReturnUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*nftOutputUnlockCondition)(nil), (*TimelockUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*nftOutputUnlockCondition)(nil), (*ExpirationUnlockCondition)(nil)))

		must(api.RegisterTypeSettings(NFTOutputFeatures{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(nftOutputV3FeatBlocksArrRules),
		))

		must(api.RegisterInterfaceObjects((*nftOutputFeature)(nil), (*SenderFeature)(nil)))
		must(api.RegisterInterfaceObjects((*nftOutputFeature)(nil), (*MetadataFeature)(nil)))
		must(api.RegisterInterfaceObjects((*nftOutputFeature)(nil), (*TagFeature)(nil)))

		must(api.RegisterTypeSettings(NFTOutputImmFeatures{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(nftOutputV3ImmFeatBlocksArrRules),
		))

		must(api.RegisterInterfaceObjects((*nftOutputImmFeature)(nil), (*IssuerFeature)(nil)))
		must(api.RegisterInterfaceObjects((*nftOutputImmFeature)(nil), (*MetadataFeature)(nil)))
	}

	{
		must(api.RegisterTypeSettings(DelegationOutput{}, serix.TypeSettings{}.WithObjectType(uint8(OutputDelegation))))
		must(api.RegisterValidators(DelegationOutput{}, nil, func(ctx context.Context, delegation DelegationOutput) error {
			return delegation.syntacticallyValidate()
		}))

		must(api.RegisterTypeSettings(DelegationOutputUnlockConditions{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(delegationOutputV3UnlockCondArrRules),
		))

		must(api.RegisterInterfaceObjects((*delegationOutputUnlockCondition)(nil), (*AddressUnlockCondition)(nil)))
	}

	{
		must(api.RegisterTypeSettings(Transaction{}, serix.TypeSettings{}.WithObjectType(TransactionNormal)))

		must(api.RegisterTypeSettings(CommitmentInput{},
			serix.TypeSettings{}.WithObjectType(uint8(InputCommitment))),
		)
		must(api.RegisterTypeSettings(BlockIssuanceCreditInput{},
			serix.TypeSettings{}.WithObjectType(uint8(InputBlockIssuanceCredit))),
		)
		must(api.RegisterTypeSettings(RewardInput{},
			serix.TypeSettings{}.WithObjectType(uint8(InputReward))),
		)

		must(api.RegisterTypeSettings(TxEssenceContextInputs{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsUint16).WithArrayRules(txEssenceV3ContextInputsArrRules),
		))

		must(api.RegisterInterfaceObjects((*txEssenceContextInput)(nil), (*CommitmentInput)(nil)))
		must(api.RegisterInterfaceObjects((*txEssenceContextInput)(nil), (*BlockIssuanceCreditInput)(nil)))
		must(api.RegisterInterfaceObjects((*txEssenceContextInput)(nil), (*RewardInput)(nil)))

		must(api.RegisterTypeSettings(UTXOInput{},
			serix.TypeSettings{}.WithObjectType(uint8(InputUTXO))),
		)

		must(api.RegisterTypeSettings(TxEssenceInputs{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsUint16).WithArrayRules(txEssenceV3InputsArrRules),
		))
		must(api.RegisterInterfaceObjects((*txEssenceInput)(nil), (*UTXOInput)(nil)))

		must(api.RegisterTypeSettings(TxEssenceOutputs{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsUint16).WithArrayRules(txEssenceV3OutputsArrRules),
		))

		must(api.RegisterTypeSettings(TxEssenceAllotments{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsUint16).WithArrayRules(txEssenceV3AllotmentsArrRules),
		))

		must(api.RegisterInterfaceObjects((*TxEssencePayload)(nil), (*TaggedData)(nil)))
		must(api.RegisterInterfaceObjects((*TxEssenceOutput)(nil), (*BasicOutput)(nil)))
		must(api.RegisterInterfaceObjects((*TxEssenceOutput)(nil), (*AccountOutput)(nil)))
		must(api.RegisterInterfaceObjects((*TxEssenceOutput)(nil), (*DelegationOutput)(nil)))
		must(api.RegisterInterfaceObjects((*TxEssenceOutput)(nil), (*FoundryOutput)(nil)))
		must(api.RegisterInterfaceObjects((*TxEssenceOutput)(nil), (*NFTOutput)(nil)))
	}

	{
		must(api.RegisterTypeSettings(SignedTransaction{}, serix.TypeSettings{}.WithObjectType(uint32(PayloadTransaction))))
		must(api.RegisterTypeSettings(Unlocks{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsUint16).WithArrayRules(txV3UnlocksArrRules),
		))
		must(api.RegisterValidators(SignedTransaction{}, nil, func(ctx context.Context, tx SignedTransaction) error {
			return tx.syntacticallyValidate(v3)
		}))
		must(api.RegisterInterfaceObjects((*TxEssencePayload)(nil), (*TaggedData)(nil)))
	}

	{
		must(api.RegisterTypeSettings(BlockIDs{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsUint32).WithArrayRules(blockIDsArrRules),
		))
	}

	{
		must(api.RegisterTypeSettings(BasicBlock{},
			serix.TypeSettings{}.WithObjectType(byte(BlockTypeBasic))),
		)
	}

	{
		must(api.RegisterTypeSettings(ValidationBlock{},
			serix.TypeSettings{}.WithObjectType(byte(BlockTypeValidation))),
		)
	}

	{
		must(api.RegisterInterfaceObjects((*Block)(nil), (*BasicBlock)(nil)))
		must(api.RegisterInterfaceObjects((*Block)(nil), (*ValidationBlock)(nil)))

		must(api.RegisterInterfaceObjects((*BlockPayload)(nil), (*SignedTransaction)(nil)))
		must(api.RegisterInterfaceObjects((*BlockPayload)(nil), (*TaggedData)(nil)))

		must(api.RegisterTypeSettings(ProtocolBlock{}, serix.TypeSettings{}))
		must(api.RegisterValidators(ProtocolBlock{}, func(ctx context.Context, bytes []byte) error {
			if len(bytes) > MaxBlockSize {
				return ierrors.Errorf("max size of a block is %d but got %d bytes", MaxBlockSize, len(bytes))
			}

			return nil
		}, func(ctx context.Context, protocolBlock ProtocolBlock) error {
			return protocolBlock.syntacticallyValidate(v3)
		}))
	}

	{
		must(api.RegisterTypeSettings(Attestation{}, serix.TypeSettings{}))
		must(api.RegisterTypeSettings(Attestations{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte),
		))
	}

	return v3
}

func calculateRewards(protoParams ProtocolParameters) (initialRewards, finalRewards uint64, err error) {
	manaStructure := protoParams.ManaDecayProvider()

	// final reward, after bootstrapping phase
	result, err := safemath.SafeMul(uint64(protoParams.TokenSupply()), protoParams.RewardsParameters().ManaShareCoefficient)
	if err != nil {
		return 0, 0, ierrors.Wrap(err, "failed to calculate target reward due to tokenSupply and RewardsManaShareCoefficient multiplication overflow")
	}

	result, err = safemath.SafeMul(result, manaStructure.generationRate)
	if err != nil {
		return 0, 0, ierrors.Wrapf(err, "failed to calculate target reward due to multiplication with generationRate overflow")
	}

	subExponent, err := safemath.SafeSub(manaStructure.generationRateExponent, uint64(protoParams.TimeProvider().SlotsPerEpochExponent()))
	if err != nil {
		return 0, 0, ierrors.Wrapf(err, "failed to calculate target reward due to generationRateExponent - slotsPerEpochExponent subtraction overflow")
	}

	finalRewards = result >> subExponent

	// initial reward for bootstrapping phase
	initialReward, err := safemath.SafeMul(finalRewards, protoParams.RewardsParameters().DecayBalancingConstant)
	if err != nil {
		return 0, 0, ierrors.Wrapf(err, "failed to calculate initial reward due to finalReward and DecayBalancingConstant multiplication overflow")
	}

	initialRewards = initialReward >> uint64(protoParams.RewardsParameters().DecayBalancingConstantExponent)

	return
}
