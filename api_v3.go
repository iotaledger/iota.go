package iotago

import (
	"context"
	"crypto/ed25519"
	"fmt"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

var (
	nativeTokensV3ArrRules = &serix.ArrayRules{
		Min: MinNativeTokenCountPerOutput,
		Max: MaxNativeTokenCountPerOutput,
		// uniqueness must be checked only by examining the actual NativeTokenID bytes
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
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
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
		Min: MinAllottmentCount,
		Max: MaxAllottmentCount,
		// TODO should we define another type to check for duplicates inside of the allotments,
		// TODO maybe can use UniquenessSliceFunc?
		ValidationMode: serializer.ArrayValidationModeLexicalOrdering,
	}

	txV3UnlocksArrRules = &serix.ArrayRules{
		Min: 1, Max: MaxInputsCount,
	}

	blockV3StrongParentsArrRules = &serix.ArrayRules{
		Min: BlockMinStrongParents,
		Max: BlockMaxParents,

		ValidationMode: serializer.ArrayValidationModeNoDuplicates | serializer.ArrayValidationModeLexicalOrdering,
	}

	blockV3NonStrongParentsArrRules = &serix.ArrayRules{
		Min: BlockMinParents,
		Max: BlockMaxParents,

		ValidationMode: serializer.ArrayValidationModeNoDuplicates | serializer.ArrayValidationModeLexicalOrdering,
	}
)

// v3api implements the iota-core 1.0 protocol core models.
type v3api struct {
	ctx          context.Context
	serixAPI     *serix.API
	timeProvider *SlotTimeProvider
}

func (v *v3api) JSONEncode(obj any, opts ...serix.Option) ([]byte, error) {
	return v.serixAPI.JSONEncode(v.ctx, obj, opts...)
}

func (v *v3api) JSONDecode(jsonData []byte, obj any, opts ...serix.Option) error {
	return v.serixAPI.JSONDecode(v.ctx, jsonData, obj, opts...)
}

func (v *v3api) Underlying() *serix.API {
	return v.serixAPI
}

func (v *v3api) SlotTimeProvider() *SlotTimeProvider {
	return v.timeProvider
}

func (v *v3api) Encode(obj interface{}, opts ...serix.Option) ([]byte, error) {
	return v.serixAPI.Encode(v.ctx, obj, opts...)
}

func (v *v3api) Decode(b []byte, obj interface{}, opts ...serix.Option) (int, error) {
	return v.serixAPI.Decode(v.ctx, b, obj, opts...)
}

// V3API instantiates an API instance with types registered conforming to protocol version 3 (iota-core 1.0) of the IOTA protocol.
func V3API(protoParams *ProtocolParameters) API {
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
		must(api.RegisterTypeSettings(AccountAddress{},
			serix.TypeSettings{}.WithObjectType(uint8(AddressAccount)).WithMapKey("accountId")),
		)
		must(api.RegisterTypeSettings(NFTAddress{},
			serix.TypeSettings{}.WithObjectType(uint8(AddressNFT)).WithMapKey("nftId")),
		)
		must(api.RegisterInterfaceObjects((*Address)(nil), (*Ed25519Address)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*AccountAddress)(nil)))
		must(api.RegisterInterfaceObjects((*Address)(nil), (*NFTAddress)(nil)))
	}

	{
		must(api.RegisterTypeSettings(IssuerFeature{}, serix.TypeSettings{}.WithObjectType(uint8(FeatureIssuer))))
		must(api.RegisterTypeSettings(MetadataFeature{}, serix.TypeSettings{}.WithObjectType(uint8(FeatureMetadata))))
		must(api.RegisterTypeSettings(SenderFeature{}, serix.TypeSettings{}.WithObjectType(uint8(FeatureSender))))
		must(api.RegisterTypeSettings(TagFeature{}, serix.TypeSettings{}.WithObjectType(uint8(FeatureTag))))
		must(api.RegisterTypeSettings(BlockIssuerFeature{}, serix.TypeSettings{}.WithObjectType(uint8(FeatureBlockIssuer))))
		must(api.RegisterInterfaceObjects((*Feature)(nil), (*IssuerFeature)(nil)))
		must(api.RegisterInterfaceObjects((*Feature)(nil), (*MetadataFeature)(nil)))
		must(api.RegisterInterfaceObjects((*Feature)(nil), (*SenderFeature)(nil)))
		must(api.RegisterInterfaceObjects((*Feature)(nil), (*TagFeature)(nil)))
		must(api.RegisterInterfaceObjects((*Feature)(nil), (*BlockIssuerFeature)(nil)))
	}

	{
		must(api.RegisterTypeSettings(AddressUnlockCondition{}, serix.TypeSettings{}.WithObjectType(uint8(UnlockConditionAddress))))
		must(api.RegisterTypeSettings(StorageDepositReturnUnlockCondition{}, serix.TypeSettings{}.WithObjectType(uint8(UnlockConditionStorageDepositReturn))))
		must(api.RegisterTypeSettings(TimelockUnlockCondition{}, serix.TypeSettings{}.WithObjectType(uint8(UnlockConditionTimelock))))
		must(api.RegisterTypeSettings(ExpirationUnlockCondition{}, serix.TypeSettings{}.WithObjectType(uint8(UnlockConditionExpiration))))
		must(api.RegisterTypeSettings(StateControllerAddressUnlockCondition{}, serix.TypeSettings{}.WithObjectType(uint8(UnlockConditionStateControllerAddress))))
		must(api.RegisterTypeSettings(GovernorAddressUnlockCondition{}, serix.TypeSettings{}.WithObjectType(uint8(UnlockConditionGovernorAddress))))
		must(api.RegisterTypeSettings(ImmutableAccountUnlockCondition{}, serix.TypeSettings{}.WithObjectType(uint8(UnlockConditionImmutableAccount))))
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
		must(api.RegisterInterfaceObjects((*Unlock)(nil), (*SignatureUnlock)(nil)))
		must(api.RegisterInterfaceObjects((*Unlock)(nil), (*ReferenceUnlock)(nil)))
		must(api.RegisterInterfaceObjects((*Unlock)(nil), (*AccountUnlock)(nil)))
		must(api.RegisterInterfaceObjects((*Unlock)(nil), (*NFTUnlock)(nil)))
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

		must(api.RegisterTypeSettings(AccountOutputImmFeatures{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(accountOutputV3ImmFeatBlocksArrRules),
		))

		must(api.RegisterInterfaceObjects((*accountOutputImmFeature)(nil), (*IssuerFeature)(nil)))
		must(api.RegisterInterfaceObjects((*accountOutputImmFeature)(nil), (*MetadataFeature)(nil)))
		must(api.RegisterInterfaceObjects((*accountOutputImmFeature)(nil), (*BlockIssuerFeature)(nil)))
	}

	{
		must(api.RegisterTypeSettings(FoundryOutput{}, serix.TypeSettings{}.WithObjectType(uint8(OutputFoundry))))

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
		must(api.RegisterTypeSettings(NFTOutput{}, serix.TypeSettings{}.WithObjectType(uint8(OutputNFT))))

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
		must(api.RegisterTypeSettings(TransactionEssence{}, serix.TypeSettings{}.WithObjectType(TransactionEssenceNormal)))

		must(api.RegisterTypeSettings(UTXOInput{}, serix.TypeSettings{}.WithObjectType(uint8(InputUTXO))))
		must(api.RegisterTypeSettings(TxEssenceInputs{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsUint16).WithArrayRules(txEssenceV3InputsArrRules),
		))
		must(api.RegisterInterfaceObjects((*txEssenceInput)(nil), (*UTXOInput)(nil)))

		must(api.RegisterTypeSettings(TxEssenceOutputs{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsUint16).WithArrayRules(txEssenceV3OutputsArrRules),
		))

		// TODO should we also register allotment itself, check how to prevent duplicates based on allotment ID
		must(api.RegisterTypeSettings(TxEssenceAllotments{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsUint16).WithArrayRules(txEssenceV3AllotmentsArrRules),
		))

		must(api.RegisterInterfaceObjects((*TxEssencePayload)(nil), (*TaggedData)(nil)))
		must(api.RegisterInterfaceObjects((*TxEssenceOutput)(nil), (*BasicOutput)(nil)))
		must(api.RegisterInterfaceObjects((*TxEssenceOutput)(nil), (*AccountOutput)(nil)))
		must(api.RegisterInterfaceObjects((*TxEssenceOutput)(nil), (*FoundryOutput)(nil)))
		must(api.RegisterInterfaceObjects((*TxEssenceOutput)(nil), (*NFTOutput)(nil)))
	}

	{
		must(api.RegisterTypeSettings(Transaction{}, serix.TypeSettings{}.WithObjectType(uint32(PayloadTransaction))))
		must(api.RegisterTypeSettings(Unlocks{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsUint16).WithArrayRules(txV3UnlocksArrRules),
		))
		must(api.RegisterValidators(&Transaction{}, nil, func(ctx context.Context, tx *Transaction) error {
			// limit unlock block count = input count
			if len(tx.Unlocks) != len(tx.Essence.Inputs) {
				return fmt.Errorf("unlock block count must match inputs in essence, %d vs. %d", len(tx.Unlocks), len(tx.Essence.Inputs))
			}
			protoParams := ctx.Value(ProtocolAPIContextKey)
			if protoParams == nil {
				return fmt.Errorf("unable to validate transaction: %w", ErrMissingProtocolParams)
			}
			return tx.syntacticallyValidate(protoParams.(*ProtocolParameters))
		}))
		must(api.RegisterInterfaceObjects((*TxEssencePayload)(nil), (*TaggedData)(nil)))
	}

	{
		must(api.RegisterTypeSettings(Block{}, serix.TypeSettings{}))
		must(api.RegisterValidators(&Block{}, func(ctx context.Context, bytes []byte) error {
			if len(bytes) > MaxBlockSize {
				return fmt.Errorf("max size of a block is %d but got %d bytes", MaxBlockSize, len(bytes))
			}
			return nil
		}, func(ctx context.Context, block *Block) error {
			val := ctx.Value(ProtocolAPIContextKey)
			if val == nil {
				return fmt.Errorf("unable to validate block: %w", ErrMissingProtocolParams)
			}
			protoParams := val.(*ProtocolParameters)
			if protoParams.Version != block.ProtocolVersion {
				return fmt.Errorf("mismatched protocol version: wanted %d, got %d in block", protoParams.Version, block.ProtocolVersion)
			}

			if len(block.WeakParents) > 0 {
				// weak parents must be disjunct to the rest of the parents
				nonWeakParents := lo.KeyOnlyBy(append(block.StrongParents, block.ShallowLikeParents...), func(v BlockID) BlockID {
					return v
				})

				for _, parent := range block.WeakParents {
					if _, contains := nonWeakParents[parent]; contains {
						return fmt.Errorf("weak parents must be disjunct to the rest of the parents")
					}
				}
			}

			return nil
		}))

		must(api.RegisterTypeSettings(StrongParentsIDs{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(blockV3StrongParentsArrRules),
		))
		must(api.RegisterTypeSettings(WeakParentsIDs{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(blockV3NonStrongParentsArrRules),
		))
		must(api.RegisterTypeSettings(ShallowLikeParentIDs{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(blockV3NonStrongParentsArrRules),
		))

		must(api.RegisterInterfaceObjects((*BlockPayload)(nil), (*Transaction)(nil)))
		must(api.RegisterInterfaceObjects((*BlockPayload)(nil), (*TaggedData)(nil)))
	}

	{
		must(api.RegisterTypeSettings(Attestation{}, serix.TypeSettings{}))
	}

	{
		must(api.RegisterTypeSettings(BlockIssuerKeys{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte).WithArrayRules(accountOutputV3BlockIssuerKeysArrRules),
		))
		must(api.RegisterTypeSettings(ed25519.PublicKey{},
			serix.TypeSettings{}.WithLengthPrefixType(serix.LengthPrefixTypeAsByte),
		))
	}

	return &v3api{
		ctx:          protoParams.AsSerixContext(),
		serixAPI:     api,
		timeProvider: protoParams.SlotTimeProvider(),
	}
}
