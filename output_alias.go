package iotago

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/serializer"
)

const (
	// AliasIDLength is the byte length of an AliasID.
	AliasIDLength = 20
)

var (
	// ErrNonUniqueAliasOutputs gets returned when multiple AliasOutputs(s) with the same AliasID exist within sets.
	ErrNonUniqueAliasOutputs = errors.New("non unique aliases within outputs")
	// ErrInvalidAliasStateTransition gets returned when an alias is doing an invalid state transition.
	ErrInvalidAliasStateTransition = errors.New("invalid alias state transition")
	// ErrInvalidAliasGovernanceTransition gets returned when an alias is doing an invalid governance transition.
	ErrInvalidAliasGovernanceTransition = errors.New("invalid alias governance transition")
	// ErrAliasMissing gets returned when an alias is missing
	ErrAliasMissing = errors.New("alias is missing")
	emptyAliasID    = [AliasIDLength]byte{}

	aliasOutputAddrGuard = &serializer.SerializableGuard{
		ReadGuard:  addrReadGuard(AddressTypeSet{AddressAlias: struct{}{}, AddressEd25519: struct{}{}}),
		WriteGuard: addrWriteGuard(AddressTypeSet{AddressAlias: struct{}{}, AddressEd25519: struct{}{}}),
	}
	aliasOutputFeatBlockArrayRules = &serializer.ArrayRules{
		Min: 0,
		Max: 2,
		Guards: serializer.SerializableGuard{
			ReadGuard: func(ty uint32) (serializer.Serializable, error) {
				switch ty {
				case uint32(FeatureBlockIssuer):
				case uint32(FeatureBlockMetadata):
				default:
					return nil, fmt.Errorf("%w: unable to deserialize alias output, unsupported feature block type %s", ErrUnsupportedFeatureBlockType, FeatureBlockTypeToString(FeatureBlockType(ty)))
				}
				return FeatureBlockSelector(ty)
			},
			WriteGuard: func(seri serializer.Serializable) error {
				switch seri.(type) {
				case *IssuerFeatureBlock:
				case *MetadataFeatureBlock:
				default:
					return fmt.Errorf("%w: in alias output", ErrUnsupportedFeatureBlockType)
				}
				return nil
			},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}
)

// AliasID is the identifier for an alias account.
// It is computed as the Blake2b-160 hash of the OutputID of the output which created the account.
type AliasID [AliasIDLength]byte

func (id AliasID) Addressable() bool {
	return true
}

func (id AliasID) Key() interface{} {
	return id.String()
}

func (id AliasID) FromUTXOInputID(in UTXOInputID) ChainID {
	aliasID := AliasIDFromOutputID(in)
	return &aliasID
}

func (id AliasID) Empty() bool {
	return id == emptyAliasID
}

func (id AliasID) String() string {
	return hex.EncodeToString(id[:])
}

func (id AliasID) Matches(other ChainID) bool {
	otherAliasID, isAliasID := other.(AliasID)
	if !isAliasID {
		return false
	}
	return id == otherAliasID
}

func (id AliasID) ToAddress() ChainConstrainedAddress {
	var addr AliasAddress
	copy(addr[:], id[:])
	return &addr
}

// AliasIDFromOutputID returns the AliasID computed from a given UTXOInputID.
func AliasIDFromOutputID(outputID UTXOInputID) AliasID {
	// TODO: maybe use pkg with Sum160 exposed
	blake2b160, _ := blake2b.New(20, nil)
	var aliasID AliasID
	copy(aliasID[:], blake2b160.Sum(outputID[:]))
	return aliasID
}

// AliasOutputs is a slice of AliasOutput(s).
type AliasOutputs []*AliasOutput

// Every checks whether every element passes f.
// Returns either -1 if all elements passed f or the index of the first element which didn't
func (outputs AliasOutputs) Every(f func(output *AliasOutput) bool) int {
	for i, output := range outputs {
		if !f(output) {
			return i
		}
	}
	return -1
}

// AliasOutputsSet is a set of AliasOutput(s).
type AliasOutputsSet map[AliasID]*AliasOutput

// Includes checks whether all aliases included in other exist in this set.
func (set AliasOutputsSet) Includes(other AliasOutputsSet) error {
	for aliasID := range other {
		if _, has := set[aliasID]; !has {
			return fmt.Errorf("%w: %s missing in source", ErrAliasMissing, aliasID)
		}
	}
	return nil
}

// EveryTuple runs f for every key which exists in both this set and other.
func (set AliasOutputsSet) EveryTuple(other AliasOutputsSet, f func(in *AliasOutput, out *AliasOutput) error) error {
	for k, v := range set {
		v2, has := other[k]
		if !has {
			continue
		}
		if err := f(v, v2); err != nil {
			return err
		}
	}
	return nil
}

// Merge merges other with this set in a new set.
// Returns an error if an alias isn't unique across both sets.
func (set AliasOutputsSet) Merge(other AliasOutputsSet) (AliasOutputsSet, error) {
	newSet := make(AliasOutputsSet)
	for k, v := range set {
		newSet[k] = v
	}
	for k, v := range other {
		if _, has := newSet[k]; has {
			return nil, fmt.Errorf("%w: alias %s exists in both sets", ErrNonUniqueAliasOutputs, k)
		}
		newSet[k] = v
	}
	return newSet, nil
}

// AliasOutput is an output type which represents an alias account.
type AliasOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount uint64
	// The native tokens held by the output.
	NativeTokens NativeTokens
	// The identifier for this alias account.
	AliasID AliasID
	// The entity which is allowed to control this alias account state.
	StateController Address
	// The entity which is allowed to govern this alias account.
	GovernanceController Address
	// The index of the state.
	StateIndex uint32
	// The state of the alias account which can only be mutated by the state controller.
	StateMetadata []byte
	// The counter that denotes the number of foundries created by this alias account.
	FoundryCounter uint32
	// The feature blocks which modulate the constraints on the output.
	Blocks FeatureBlocks
}

func (a *AliasOutput) VByteCost(costStruct *RentStructure, override VByteCostFunc) uint64 {
	return costStruct.VBFactorKey.Multiply(UTXOIDLength) +
		costStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+serializer.UInt64ByteSize) +
		a.NativeTokens.VByteCost(costStruct, nil) +
		costStruct.VBFactorKey.With(costStruct.VBFactorData).Multiply(AliasIDLength) +
		a.StateController.VByteCost(costStruct, nil) +
		a.GovernanceController.VByteCost(costStruct, nil) +
		costStruct.VBFactorData.Multiply(uint64(serializer.UInt32ByteSize+serializer.UInt32ByteSize+len(a.StateMetadata)+serializer.UInt32ByteSize)) +
		a.Blocks.VByteCost(costStruct, nil)
}

//	TODO: document transitions
//	- For output AliasOutput(s) with non-zeroed AliasID, there must be a corresponding input AliasOutput where either
//	  its AliasID is zeroed and StateIndex and FoundryCounter are zero or an input AliasOutput with the same AliasID.
//	- On alias state transitions:
//		- The StateIndex must be incremented by 1
//		- Only Amount, NativeTokens, StateIndex, StateMetadata and FoundryCounter can be mutated
//	- On alias governance transition:
//		- Only StateController (must be mutated), GovernanceController and the MetadataBlock can be mutated
func (a *AliasOutput) ValidateStateTransition(transType ChainTransitionType, next ChainConstrainedOutput, semValCtx *SemanticValidationContext) error {
	switch transType {
	case ChainTransitionTypeGenesis:
		if !a.AliasID.Empty() {
			return fmt.Errorf("%w: AliasOutput's ID is not zeroed even though it is new", ErrInvalidChainStateTransition)
		}
		return IsIssuerOnOutputUnlocked(a, semValCtx.WorkingSet.UnlockedIdents)
	case ChainTransitionTypeStateChange:
		nextAliasOutput, is := next.(*AliasOutput)
		if !is {
			return fmt.Errorf("%w: AliasOutput can only state transition to another alias output", ErrInvalidChainStateTransition)
		}
		if err := IssuerBlockUnchanged(a, nextAliasOutput); err != nil {
			return err
		}
		if a.StateIndex == nextAliasOutput.StateIndex {
			return a.GovernanceSTVF(nextAliasOutput, semValCtx)
		}
		return a.StateSTVF(nextAliasOutput, semValCtx)
	case ChainTransitionTypeDestroy:
		return nil
	default:
		panic("unknown chain transition type in AliasOutput")
	}
}

func (a *AliasOutput) Ident(nextState MultiIdentOutput) (Address, error) {
	otherAliasOutput, isAliasOutput := nextState.(*AliasOutput)
	if !isAliasOutput {
		return nil, fmt.Errorf("%w: expected AliasOutput but got %s for ident computation", ErrMultiIdentOutputMismatch, OutputTypeToString(nextState.Type()))
	}
	switch {
	case a.StateIndex == otherAliasOutput.StateIndex:
		return a.GovernanceController, nil
	case a.StateIndex+1 == otherAliasOutput.StateIndex:
		return a.StateController, nil
	default:
		return nil, fmt.Errorf("%w: can not compute right ident for alias output as state index delta is invalid", ErrMultiIdentOutputMismatch)
	}
}

func (a *AliasOutput) Chain() ChainID {
	return a.AliasID
}

// GovernanceSTVF checks whether the governance transition with other is valid.
// Under a governance transition, only the StateController, GovernanceController and MetadataFeatureBlock can change.
func (a *AliasOutput) GovernanceSTVF(nextAliasOutput *AliasOutput, semValCtx *SemanticValidationContext) error {
	switch {
	case a.Amount != nextAliasOutput.Amount:
		return fmt.Errorf("%w: amount changed, in %d / out %d ", ErrInvalidAliasGovernanceTransition, a.Amount, nextAliasOutput.Amount)
	case !a.NativeTokens.Equal(nextAliasOutput.NativeTokens):
		return fmt.Errorf("%w: native tokens changed, in %v / out %v", ErrInvalidAliasGovernanceTransition, a.NativeTokens, nextAliasOutput.NativeTokens)
	case a.StateIndex != nextAliasOutput.StateIndex:
		return fmt.Errorf("%w: state index changed, in %d / out %d", ErrInvalidAliasGovernanceTransition, a.StateIndex, nextAliasOutput.StateIndex)
	case !bytes.Equal(a.StateMetadata, nextAliasOutput.StateMetadata):
		return fmt.Errorf("%w: state metadata changed, in %v / out %v", ErrInvalidAliasGovernanceTransition, a.StateMetadata, nextAliasOutput.StateMetadata)
	case a.FoundryCounter != nextAliasOutput.FoundryCounter:
		return fmt.Errorf("%w: foundry counter changed, in %d / out %d", ErrInvalidAliasGovernanceTransition, a.FoundryCounter, nextAliasOutput.FoundryCounter)
	}
	return nil
}

// StateSTVF checks whether the state transition with other is valid.
// Under a state transition, only Amount, NativeTokens, StateIndex, StateMetadata and FoundryCounter can change.
func (a *AliasOutput) StateSTVF(nextAliasOutput *AliasOutput, semValCtx *SemanticValidationContext) error {
	switch {
	case !a.StateController.Equal(nextAliasOutput.StateController):
		return fmt.Errorf("%w: state controller changed, in %v / out %v", ErrInvalidAliasStateTransition, a.StateController, nextAliasOutput.StateController)
	case !a.GovernanceController.Equal(nextAliasOutput.GovernanceController):
		return fmt.Errorf("%w: governance controller changed, in %v / out %v", ErrInvalidAliasStateTransition, a.StateController, nextAliasOutput.StateController)
	case !a.FeatureBlocks().Equal(nextAliasOutput.FeatureBlocks()):
		return fmt.Errorf("%w: feature blocks changed, in %v / out %v", ErrInvalidAliasStateTransition, a.StateController, nextAliasOutput.StateController)
	case a.FoundryCounter > nextAliasOutput.FoundryCounter:
		return fmt.Errorf("%w: foundry counter of next state is less than previous, in %d / out %d", ErrInvalidAliasStateTransition, a.FoundryCounter, nextAliasOutput.FoundryCounter)
	case a.StateIndex+1 != nextAliasOutput.StateIndex:
		return fmt.Errorf("%w: state index %d on the input side but %d on the output side", ErrInvalidAliasStateTransition, a.StateIndex, nextAliasOutput.StateIndex)
	}

	// check that for a foundry counter change, X amount of foundries were actually created
	if a.FoundryCounter == nextAliasOutput.FoundryCounter {
		return nil
	}

	var seenNewFoundriesOfAlias uint32
	for _, output := range semValCtx.WorkingSet.Tx.Essence.Outputs {
		foundryOutput, is := output.(*FoundryOutput)
		if !is {
			continue
		}
		foundryAliasID := foundryOutput.Address.(*AliasAddress).Chain()
		if foundryAliasID != a.AliasID {
			continue
		}
		if _, notNew := semValCtx.WorkingSet.InChains[foundryAliasID]; notNew {
			continue
		}
	}

	expectedNewFoundriesCount := nextAliasOutput.FoundryCounter - a.FoundryCounter
	if expectedNewFoundriesCount != seenNewFoundriesOfAlias {
		return fmt.Errorf("%w: %d new foundries were created but the alias output's foundry counter changed by %d", ErrInvalidAliasStateTransition, seenNewFoundriesOfAlias, expectedNewFoundriesCount)
	}

	return nil
}

func (a *AliasOutput) AliasEmpty() bool {
	return a.AliasID == emptyAliasID
}

func (a *AliasOutput) NativeTokenSet() NativeTokens {
	return a.NativeTokens
}

func (a *AliasOutput) FeatureBlocks() FeatureBlocks {
	return a.Blocks
}

func (a *AliasOutput) Deposit() (uint64, error) {
	return a.Amount, nil
}

func (a *AliasOutput) Target() (serializer.Serializable, error) {
	addr := new(AliasAddress)
	copy(addr[:], a.AliasID[:])
	return addr, nil
}

func (a *AliasOutput) Type() OutputType {
	return OutputAlias
}

func (a *AliasOutput) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(OutputAlias), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize alias output: %w", err)
		}).
		ReadNum(&a.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for alias output: %w", err)
		}).
		ReadSliceOfObjects(&a.NativeTokens, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationNone, nativeTokensArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize native tokens for alias output: %w", err)
		}).
		ReadBytesInPlace(a.AliasID[:], func(err error) error {
			return fmt.Errorf("unable to deserialize alias ID for alias output: %w", err)
		}).
		ReadObject(&a.StateController, deSeriMode, serializer.TypeDenotationByte, aliasOutputAddrGuard.ReadGuard, func(err error) error {
			return fmt.Errorf("unable to deserialize state controller for alias output: %w", err)
		}).
		ReadObject(&a.GovernanceController, deSeriMode, serializer.TypeDenotationByte, aliasOutputAddrGuard.ReadGuard, func(err error) error {
			return fmt.Errorf("unable to deserialize governance controller for alias output: %w", err)
		}).
		ReadNum(&a.StateIndex, func(err error) error {
			return fmt.Errorf("unable to deserialize state index for alias output: %w", err)
		}).
		ReadVariableByteSlice(&a.StateMetadata, serializer.SeriLengthPrefixTypeAsUint32, func(err error) error {
			// TODO: replace max read with actual variable
			return fmt.Errorf("unable to deserialize state metadata for alias output: %w", err)
		}, MaxMetadataLength).
		ReadNum(&a.FoundryCounter, func(err error) error {
			return fmt.Errorf("unable to deserialize foundry counter for alias output: %w", err)
		}).
		ReadSliceOfObjects(&a.Blocks, deSeriMode, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationByte, aliasOutputFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize feature blocks for NFT output: %w", err)
		}).
		Done()
}

func (a *AliasOutput) Serialize(deSeriMode serializer.DeSerializationMode) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(OutputAlias, func(err error) error {
			return fmt.Errorf("unable to serialize alias output type ID: %w", err)
		}).
		WriteNum(a.Amount, func(err error) error {
			return fmt.Errorf("unable to serialize alias output amount: %w", err)
		}).
		WriteSliceOfObjects(&a.NativeTokens, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, nativeTokensArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize alias output native tokens: %w", err)
		}).
		WriteBytes(a.AliasID[:], func(err error) error {
			return fmt.Errorf("unable to serialize alias output alias ID: %w", err)
		}).
		WriteObject(a.StateController, deSeriMode, aliasOutputAddrGuard.WriteGuard, func(err error) error {
			return fmt.Errorf("unable to serialize alias output state controller: %w", err)
		}).
		WriteObject(a.GovernanceController, deSeriMode, aliasOutputAddrGuard.WriteGuard, func(err error) error {
			return fmt.Errorf("unable to serialize alias output governance controller: %w", err)
		}).
		WriteNum(a.StateIndex, func(err error) error {
			return fmt.Errorf("unable to serialize alias output state index: %w", err)
		}).
		WriteVariableByteSlice(a.StateMetadata, serializer.SeriLengthPrefixTypeAsUint32, func(err error) error {
			return fmt.Errorf("unable to serialize alias output state metadata: %w", err)
		}).
		WriteNum(a.FoundryCounter, func(err error) error {
			return fmt.Errorf("unable to serialize alias output foundry counter: %w", err)
		}).
		WriteSliceOfObjects(&a.Blocks, deSeriMode, serializer.SeriLengthPrefixTypeAsByte, aliasOutputFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize alias output feature blocks: %w", err)
		}).
		Serialize()
}

func (a *AliasOutput) MarshalJSON() ([]byte, error) {
	var err error
	jAliasOutput := &jsonAliasOutput{
		Type:           int(OutputAlias),
		Amount:         int(a.Amount),
		StateIndex:     int(a.StateIndex),
		FoundryCounter: int(a.FoundryCounter),
	}

	jAliasOutput.NativeTokens, err = serializablesToJSONRawMsgs(a.NativeTokens.ToSerializables())
	if err != nil {
		return nil, err
	}

	jAliasOutput.AliasID = hex.EncodeToString(a.AliasID[:])

	jAliasOutput.StateController, err = addressToJSONRawMsg(a.StateController)
	if err != nil {
		return nil, err
	}

	jAliasOutput.GovernanceController, err = addressToJSONRawMsg(a.GovernanceController)
	if err != nil {
		return nil, err
	}

	jAliasOutput.StateMetadata = hex.EncodeToString(a.StateMetadata)

	jAliasOutput.Blocks, err = serializablesToJSONRawMsgs(a.Blocks.ToSerializables())
	if err != nil {
		return nil, err
	}

	return json.Marshal(jAliasOutput)
}

func (a *AliasOutput) UnmarshalJSON(bytes []byte) error {
	jAliasOutput := &jsonAliasOutput{}
	if err := json.Unmarshal(bytes, jAliasOutput); err != nil {
		return err
	}
	seri, err := jAliasOutput.ToSerializable()
	if err != nil {
		return err
	}
	*a = *seri.(*AliasOutput)
	return nil
}

// jsonAliasOutput defines the json representation of an AliasOutput.
type jsonAliasOutput struct {
	Type                 int                `json:"type"`
	Amount               int                `json:"amount"`
	NativeTokens         []*json.RawMessage `json:"nativeTokens"`
	AliasID              string             `json:"aliasId"`
	StateController      *json.RawMessage   `json:"stateController"`
	GovernanceController *json.RawMessage   `json:"governanceController"`
	StateIndex           int                `json:"stateIndex"`
	StateMetadata        string             `json:"stateMetadata"`
	FoundryCounter       int                `json:"foundryCounter"`
	Blocks               []*json.RawMessage `json:"blocks"`
}

func (j *jsonAliasOutput) ToSerializable() (serializer.Serializable, error) {
	var err error
	e := &AliasOutput{
		Amount:         uint64(j.Amount),
		StateIndex:     uint32(j.StateIndex),
		FoundryCounter: uint32(j.FoundryCounter),
	}

	e.NativeTokens, err = nativeTokensFromJSONRawMsg(j.NativeTokens)
	if err != nil {
		return nil, err
	}

	aliasIDSlice, err := hex.DecodeString(j.AliasID)
	if err != nil {
		return nil, err
	}
	copy(e.AliasID[:], aliasIDSlice)

	e.StateController, err = addressFromJSONRawMsg(j.StateController)
	if err != nil {
		return nil, err
	}

	e.GovernanceController, err = addressFromJSONRawMsg(j.GovernanceController)
	if err != nil {
		return nil, err
	}

	e.StateMetadata, err = hex.DecodeString(j.StateMetadata)
	if err != nil {
		return nil, err
	}

	e.Blocks, err = featureBlocksFromJSONRawMsg(j.Blocks)
	if err != nil {
		return nil, err
	}

	return e, nil
}
