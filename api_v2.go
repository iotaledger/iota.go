package iotago

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"github.com/iancoleman/orderedmap"
	"github.com/iotaledger/hive.go/serializer"
	"github.com/iotaledger/hive.go/serix"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

var (
	nativeTokensV2ArrRules = &serix.ArrayRules{
		Min: MinNativeTokenCountPerOutput,
		Max: MaxNativeTokenCountPerOutput,
		// uniqueness must be checked only by examining the actual NativeTokenID bytes
		UniquenessSliceFunc: func(next []byte) []byte { return next[:NativeTokenIDLength] },
		ValidationMode:      serializer.ArrayValidationModeNoDuplicates | serializer.ArrayValidationModeLexicalOrdering,
	}

	basicOutputV2UnlockCondArrRules = &serix.ArrayRules{
		Min: 1,
		Max: 4,
		MustOccur: serializer.TypePrefixes{
			uint32(UnlockConditionAddress): struct{}{},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}
	basicOutputV2FeatBlocksArrRules = &serix.ArrayRules{
		Min: 0,
		Max: 8,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	aliasOutputV2UnlockCondArrRules = &serix.ArrayRules{
		Min: 2, Max: 2,
		MustOccur: serializer.TypePrefixes{
			uint32(UnlockConditionStateControllerAddress): struct{}{},
			uint32(UnlockConditionGovernorAddress):        struct{}{},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	aliasOutputV2FeatBlocksArrRules = &serix.ArrayRules{
		Min: 0,
		Max: 3,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	aliasOutputV2ImmFeatBlocksArrRules = &serix.ArrayRules{
		Min: 0,
		Max: 2,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	foundryOutputV2UnlockCondArrRules = &serix.ArrayRules{
		Min: 1, Max: 1,
		MustOccur: serializer.TypePrefixes{
			uint32(UnlockConditionImmutableAlias): struct{}{},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	foundryOutputV2FeatBlocksArrRules = &serix.ArrayRules{
		Min: 0, Max: 1,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	foundryOutputV2ImmFeatBlocksArrRules = &serix.ArrayRules{
		Min: 0, Max: 1,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	nftOutputV2UnlockCondArrRules = &serix.ArrayRules{
		Min: 1, Max: 4,
		MustOccur: serializer.TypePrefixes{
			uint32(UnlockConditionAddress): struct{}{},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	nftOutputV2FeatBlocksArrRules = &serix.ArrayRules{
		Min: 0,
		Max: 3,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	nftOutputV2ImmFeatBlocksArrRules = &serix.ArrayRules{
		Min: 0,
		Max: 2,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	txEssenceV2InputsArrRules = &serix.ArrayRules{
		Min:            MinInputsCount,
		Max:            MaxInputsCount,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates,
	}

	txEssenceV2OutputsArrRules = &serix.ArrayRules{
		Min:            MinOutputsCount,
		Max:            MaxOutputsCount,
		ValidationMode: serializer.ArrayValidationModeNone,
	}

	txV2UnlocksArrRules = &serix.ArrayRules{
		Min: 1, Max: MaxInputsCount,
	}

	msV2ParentsArrRules = &serix.ArrayRules{
		Min: BlockMinParents,
		Max: BlockMaxParents,

		ValidationMode: serializer.ArrayValidationModeNoDuplicates | serializer.ArrayValidationModeLexicalOrdering,
	}

	msV2OptsArrRules = &serix.ArrayRules{
		Min:            0,
		Max:            2,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates | serializer.ArrayValidationModeLexicalOrdering,
	}

	msV2SigsArrRules = &serix.ArrayRules{
		Min:                 MinSignaturesInAMilestone,
		Max:                 MaxSignaturesInAMilestone,
		UniquenessSliceFunc: func(next []byte) []byte { return next[:ed25519.PublicKeySize] },
		ValidationMode:      serializer.ArrayValidationModeNoDuplicates | serializer.ArrayValidationModeLexicalOrdering,
	}

	receiptV2MigArrRules = &serix.ArrayRules{
		Min:            MinMigratedFundsEntryCount,
		Max:            MaxMigratedFundsEntryCount,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates | serializer.ArrayValidationModeLexicalOrdering,
	}
)

type v2api struct {
	ctx      context.Context
	serixAPI *serix.API
}

func (v *v2api) Underlying() *serix.API {
	return v.serixAPI
}

func (v *v2api) MapDecode(m map[string]any, obj any, opts ...serix.Option) error {
	return v.serixAPI.MapDecode(v.ctx, m, obj, opts...)
}

func (v *v2api) MapEncode(obj interface{}, opts ...serix.Option) (*orderedmap.OrderedMap, error) {
	return v.serixAPI.MapEncode(v.ctx, obj, opts...)
}

func (v *v2api) Encode(obj interface{}, opts ...serix.Option) ([]byte, error) {
	return v.serixAPI.Encode(v.ctx, obj, opts...)
}

func (v *v2api) Decode(b []byte, obj interface{}, opts ...serix.Option) (int, error) {
	return v.serixAPI.Decode(v.ctx, b, obj, opts...)
}

// V2API instantiates an API instance with types registered conforming to protocol version 2 of the IOTA protocol.
func V2API(protoParas *ProtocolParameters) API {
	api := serix.NewAPI()

	must(api.RegisterTypeSettings(ProtocolParameters{}, serix.TypeSettings{}))
	must(api.RegisterTypeSettings(RentStructure{}, serix.TypeSettings{}))

	must(api.RegisterTypeSettings(TaggedData{}, serix.TypeSettings{}.WithObjectType(uint32(PayloadTaggedData))))

	{
		must(api.RegisterTypeSettings(Ed25519Signature{},
			serix.TypeSettings{}.WithObjectType(uint8(SignatureEd25519))),
		)
		must(api.RegisterInterfaceObjects((*Signature)(nil), (*Ed25519Signature)(nil)))
	}

	{
		must(api.RegisterTypeSettings(Ed25519Address{},
			serix.TypeSettings{}.WithObjectType(uint8(AddressEd25519)).WithMapKey("pubKeyHash")),
		)
		must(api.RegisterTypeSettings(AliasAddress{},
			serix.TypeSettings{}.WithObjectType(uint8(AddressAlias)).WithMapKey("aliasId")),
		)
		must(api.RegisterTypeSettings(NFTAddress{},
			serix.TypeSettings{}.WithObjectType(uint8(AddressNFT)).WithMapKey("nftId")),
		)
		must(api.RegisterInterfaceObjects((*Address)(nil), (*Ed25519Address)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*AliasAddress)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*NFTAddress)(nil)))
	}

	{
		must(api.RegisterTypeSettings(IssuerFeature{}, serix.TypeSettings{}.WithObjectType(uint8(FeatureIssuer))))
		must(api.RegisterTypeSettings(MetadataFeature{}, serix.TypeSettings{}.WithObjectType(uint8(FeatureMetadata))))
		must(api.RegisterTypeSettings(SenderFeature{}, serix.TypeSettings{}.WithObjectType(uint8(FeatureSender))))
		must(api.RegisterTypeSettings(TagFeature{}, serix.TypeSettings{}.WithObjectType(uint8(FeatureTag))))
		must(api.RegisterInterfaceObjects((*Feature)(nil), (*IssuerFeature)(nil)))
		must(api.RegisterInterfaceObjects((*Feature)(nil), (*MetadataFeature)(nil)))
		must(api.RegisterInterfaceObjects((*Feature)(nil), (*SenderFeature)(nil)))
		must(api.RegisterInterfaceObjects((*Feature)(nil), (*TagFeature)(nil)))
	}

	{
		must(api.RegisterTypeSettings(AddressUnlockCondition{}, serix.TypeSettings{}.WithObjectType(uint8(UnlockConditionAddress))))
		must(api.RegisterTypeSettings(StorageDepositReturnUnlockCondition{}, serix.TypeSettings{}.WithObjectType(uint8(UnlockConditionStorageDepositReturn))))
		must(api.RegisterTypeSettings(TimelockUnlockCondition{}, serix.TypeSettings{}.WithObjectType(uint8(UnlockConditionTimelock))))
		must(api.RegisterTypeSettings(ExpirationUnlockCondition{}, serix.TypeSettings{}.WithObjectType(uint8(UnlockConditionExpiration))))
		must(api.RegisterTypeSettings(StateControllerAddressUnlockCondition{}, serix.TypeSettings{}.WithObjectType(uint8(UnlockConditionStateControllerAddress))))
		must(api.RegisterTypeSettings(GovernorAddressUnlockCondition{}, serix.TypeSettings{}.WithObjectType(uint8(UnlockConditionGovernorAddress))))
		must(api.RegisterTypeSettings(ImmutableAliasUnlockCondition{}, serix.TypeSettings{}.WithObjectType(uint8(UnlockConditionImmutableAlias))))
		must(api.RegisterInterfaceObjects((*UnlockCondition)(nil), (*AddressUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*UnlockCondition)(nil), (*StorageDepositReturnUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*UnlockCondition)(nil), (*TimelockUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*UnlockCondition)(nil), (*ExpirationUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*UnlockCondition)(nil), (*StateControllerAddressUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*UnlockCondition)(nil), (*GovernorAddressUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*UnlockCondition)(nil), (*ImmutableAliasUnlockCondition)(nil)))
	}

	{
		must(api.RegisterTypeSettings(SignatureUnlock{}, serix.TypeSettings{}.WithObjectType(uint8(UnlockSignature))))
		must(api.RegisterTypeSettings(ReferenceUnlock{}, serix.TypeSettings{}.WithObjectType(uint8(UnlockReference))))
		must(api.RegisterTypeSettings(AliasUnlock{}, serix.TypeSettings{}.WithObjectType(uint8(UnlockAlias))))
		must(api.RegisterTypeSettings(NFTUnlock{}, serix.TypeSettings{}.WithObjectType(uint8(UnlockNFT))))
		must(api.RegisterInterfaceObjects((*Unlock)(nil), (*SignatureUnlock)(nil)))
		must(api.RegisterInterfaceObjects((*Unlock)(nil), (*ReferenceUnlock)(nil)))
		must(api.RegisterInterfaceObjects((*Unlock)(nil), (*AliasUnlock)(nil)))
		must(api.RegisterInterfaceObjects((*Unlock)(nil), (*NFTUnlock)(nil)))
	}

	{
		must(api.RegisterTypeSettings(NativeToken{}, serix.TypeSettings{}))
		must(api.RegisterTypeSettings(NativeTokens{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(nativeTokensV2ArrRules),
		))
	}

	{
		must(api.RegisterTypeSettings(BasicOutput{}, serix.TypeSettings{}.WithObjectType(uint8(OutputBasic))))

		must(api.RegisterTypeSettings(UnlockConditions[BasicOutputUnlockCondition]{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(basicOutputV2UnlockCondArrRules),
		))

		must(api.RegisterInterfaceObjects((*BasicOutputUnlockCondition)(nil), (*AddressUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*BasicOutputUnlockCondition)(nil), (*StorageDepositReturnUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*BasicOutputUnlockCondition)(nil), (*TimelockUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*BasicOutputUnlockCondition)(nil), (*ExpirationUnlockCondition)(nil)))

		must(api.RegisterTypeSettings(Features[BasicOutputFeature]{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(basicOutputV2FeatBlocksArrRules),
		))

		must(api.RegisterInterfaceObjects((*BasicOutputFeature)(nil), (*SenderFeature)(nil)))
		must(api.RegisterInterfaceObjects((*BasicOutputFeature)(nil), (*MetadataFeature)(nil)))
		must(api.RegisterInterfaceObjects((*BasicOutputFeature)(nil), (*TagFeature)(nil)))
	}

	{
		must(api.RegisterTypeSettings(AliasOutput{}, serix.TypeSettings{}.WithObjectType(uint8(OutputAlias))))

		must(api.RegisterTypeSettings(UnlockConditions[AliasUnlockCondition]{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(aliasOutputV2UnlockCondArrRules),
		))

		must(api.RegisterInterfaceObjects((*AliasUnlockCondition)(nil), (*StateControllerAddressUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*AliasUnlockCondition)(nil), (*GovernorAddressUnlockCondition)(nil)))

		must(api.RegisterTypeSettings(Features[AliasFeature]{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(aliasOutputV2FeatBlocksArrRules),
		))

		must(api.RegisterInterfaceObjects((*AliasFeature)(nil), (*SenderFeature)(nil)))
		must(api.RegisterInterfaceObjects((*AliasFeature)(nil), (*MetadataFeature)(nil)))

		must(api.RegisterTypeSettings(Features[AliasImmFeature]{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(aliasOutputV2ImmFeatBlocksArrRules),
		))

		must(api.RegisterInterfaceObjects((*AliasImmFeature)(nil), (*IssuerFeature)(nil)))
		must(api.RegisterInterfaceObjects((*AliasImmFeature)(nil), (*MetadataFeature)(nil)))
	}

	{
		must(api.RegisterTypeSettings(FoundryOutput{}, serix.TypeSettings{}.WithObjectType(uint8(OutputFoundry))))

		must(api.RegisterTypeSettings(UnlockConditions[FoundryUnlockCondition]{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(foundryOutputV2UnlockCondArrRules),
		))

		must(api.RegisterInterfaceObjects((*FoundryUnlockCondition)(nil), (*ImmutableAliasUnlockCondition)(nil)))

		must(api.RegisterTypeSettings(Features[FoundryFeature]{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(foundryOutputV2FeatBlocksArrRules),
		))

		must(api.RegisterInterfaceObjects((*FoundryFeature)(nil), (*MetadataFeature)(nil)))

		must(api.RegisterTypeSettings(Features[FoundryImmFeature]{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(foundryOutputV2ImmFeatBlocksArrRules),
		))

		must(api.RegisterInterfaceObjects((*FoundryImmFeature)(nil), (*MetadataFeature)(nil)))

		must(api.RegisterTypeSettings(SimpleTokenScheme{}, serix.TypeSettings{}.WithObjectType(uint8(TokenSchemeSimple))))
		must(api.RegisterInterfaceObjects((*TokenScheme)(nil), (*SimpleTokenScheme)(nil)))
	}

	{
		must(api.RegisterTypeSettings(NFTOutput{}, serix.TypeSettings{}.WithObjectType(uint8(OutputNFT))))

		must(api.RegisterTypeSettings(UnlockConditions[NFTUnlockCondition]{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(nftOutputV2UnlockCondArrRules),
		))

		must(api.RegisterInterfaceObjects((*NFTUnlockCondition)(nil), (*AddressUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*NFTUnlockCondition)(nil), (*StorageDepositReturnUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*NFTUnlockCondition)(nil), (*TimelockUnlockCondition)(nil)))
		must(api.RegisterInterfaceObjects((*NFTUnlockCondition)(nil), (*ExpirationUnlockCondition)(nil)))

		must(api.RegisterTypeSettings(Features[NFTFeature]{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(nftOutputV2FeatBlocksArrRules),
		))

		must(api.RegisterInterfaceObjects((*NFTFeature)(nil), (*SenderFeature)(nil)))
		must(api.RegisterInterfaceObjects((*NFTFeature)(nil), (*MetadataFeature)(nil)))
		must(api.RegisterInterfaceObjects((*NFTFeature)(nil), (*TagFeature)(nil)))

		must(api.RegisterTypeSettings(Features[NFTImmFeature]{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(nftOutputV2ImmFeatBlocksArrRules),
		))

		must(api.RegisterInterfaceObjects((*NFTImmFeature)(nil), (*IssuerFeature)(nil)))
		must(api.RegisterInterfaceObjects((*NFTImmFeature)(nil), (*MetadataFeature)(nil)))
	}

	{
		must(api.RegisterTypeSettings(TransactionEssence{}, serix.TypeSettings{}.WithObjectType(uint8(TransactionEssenceNormal))))

		must(api.RegisterTypeSettings(UTXOInput{}, serix.TypeSettings{}.WithObjectType(uint8(InputUTXO))))
		must(api.RegisterTypeSettings(Inputs[TxEssenceInput]{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsUint16).WithArrayRules(txEssenceV2InputsArrRules),
		))
		must(api.RegisterInterfaceObjects((*TxEssenceInput)(nil), (*UTXOInput)(nil)))

		must(api.RegisterTypeSettings(Outputs[TxEssenceOutput]{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsUint16).WithArrayRules(txEssenceV2OutputsArrRules),
		))
		must(api.RegisterInterfaceObjects((*TxEssencePayload)(nil), (*TaggedData)(nil)))
		must(api.RegisterInterfaceObjects((*TxEssenceOutput)(nil), (*BasicOutput)(nil)))
		must(api.RegisterInterfaceObjects((*TxEssenceOutput)(nil), (*AliasOutput)(nil)))
		must(api.RegisterInterfaceObjects((*TxEssenceOutput)(nil), (*FoundryOutput)(nil)))
		must(api.RegisterInterfaceObjects((*TxEssenceOutput)(nil), (*NFTOutput)(nil)))
	}

	{
		must(api.RegisterTypeSettings(Transaction{}, serix.TypeSettings{}.WithObjectType(uint32(PayloadTransaction))))
		must(api.RegisterTypeSettings(Unlocks{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsUint16).WithArrayRules(txV2UnlocksArrRules),
		))
		must(api.RegisterValidators(&Transaction{}, nil, func(ctx context.Context, tx *Transaction) error {
			// limit unlock block count = input count
			if len(tx.Unlocks) != len(tx.Essence.Inputs) {
				return fmt.Errorf("unlock block count must match inputs in essence, %d vs. %d", len(tx.Unlocks), len(tx.Essence.Inputs))
			}
			protoParas := ctx.Value(ProtocolAPIContextKey)
			if protoParas == nil {
				return fmt.Errorf("unable to validate transaction: %w", ErrMissingProtocolParas)
			}
			return tx.syntacticallyValidate(protoParas.(*ProtocolParameters))
		}))
		must(api.RegisterInterfaceObjects((*TxEssencePayload)(nil), (*TaggedData)(nil)))
	}

	{
		must(api.RegisterTypeSettings(Milestone{}, serix.TypeSettings{}.WithObjectType(uint32(PayloadMilestone))))
		must(api.RegisterTypeSettings(MilestoneEssence{}, serix.TypeSettings{}))
		must(api.RegisterTypeSettings(ReceiptMilestoneOpt{}, serix.TypeSettings{}.WithObjectType(uint8(MilestoneOptReceipt))))
		must(api.RegisterTypeSettings(ProtocolParamsMilestoneOpt{}, serix.TypeSettings{}.WithObjectType(uint8(MilestoneOptProtocolParams))))
		must(api.RegisterTypeSettings(MigratedFundsEntries{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsUint16).WithArrayRules(receiptV2MigArrRules)),
		)
		must(api.RegisterTypeSettings(MigratedFundsEntry{}, serix.TypeSettings{}))

		must(api.RegisterTypeSettings(TreasuryTransaction{}, serix.TypeSettings{}.WithObjectType(uint32(PayloadTreasuryTransaction))))
		must(api.RegisterTypeSettings(TreasuryInput{},
			serix.TypeSettings{}.WithObjectType(uint8(InputTreasury)).WithMapKey("milestoneId")),
		)
		must(api.RegisterTypeSettings(TreasuryOutput{}, serix.TypeSettings{}.WithObjectType(uint8(OutputTreasury))))

		must(api.RegisterTypeSettings(MilestoneOpts{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(msV2OptsArrRules),
		))
		must(api.RegisterInterfaceObjects((*MilestoneOpt)(nil), (*ReceiptMilestoneOpt)(nil)))
		must(api.RegisterInterfaceObjects((*MilestoneOpt)(nil), (*ProtocolParamsMilestoneOpt)(nil)))

		must(api.RegisterTypeSettings(MilestoneParentIDs{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(msV2ParentsArrRules),
		))

		must(api.RegisterTypeSettings(Signatures[MilestoneSignature]{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(msV2SigsArrRules),
		))
		must(api.RegisterInterfaceObjects((*MilestoneSignature)(nil), (*Ed25519Signature)(nil)))
	}

	{
		must(api.RegisterTypeSettings(Block{}, serix.TypeSettings{}))
		must(api.RegisterValidators(&Block{}, func(ctx context.Context, bytes []byte) error {
			if len(bytes) > MaxBlockSize {
				return fmt.Errorf("max size of a block is %d but got %d bytes", MaxBlockSize, len(bytes))
			}
			return nil
		}, nil))
		must(api.RegisterInterfaceObjects((*BlockPayload)(nil), (*Transaction)(nil)))
		must(api.RegisterInterfaceObjects((*BlockPayload)(nil), (*Milestone)(nil)))
		must(api.RegisterInterfaceObjects((*BlockPayload)(nil), (*TaggedData)(nil)))
	}

	return &v2api{
		ctx:      protoParas.AsSerixContext(),
		serixAPI: api,
	}
}
