package iotago

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer"

	"golang.org/x/crypto/blake2b"
)

const (
	// TransactionIDLength defines the length of a Transaction ID.
	TransactionIDLength = blake2b.Size256

	// TransactionBinSerializedMinSize defines the minimum size of a serialized Transaction.
	TransactionBinSerializedMinSize = serializer.UInt32ByteSize
)

var (
	// ErrUnlockBlocksMustMatchInputCount gets returned if the count of unlock blocks doesn't match the count of inputs.
	ErrUnlockBlocksMustMatchInputCount = errors.New("the count of unlock blocks must match the inputs of the transaction")
	// ErrInvalidTransactionEssence gets returned if the transaction essence within a Transaction is invalid.
	ErrInvalidTransactionEssence = errors.New("transaction essence is invalid")
	// ErrMissingUTXO gets returned if an UTXO is missing to commence a certain operation.
	ErrMissingUTXO = errors.New("missing utxo")
	// ErrInputOutputSumMismatch gets returned if a transaction does not spend the entirety of the inputs to the outputs.
	ErrInputOutputSumMismatch = errors.New("inputs and outputs do not spend/deposit the same amount")
	// ErrInputSignatureUnlockBlockInvalid gets returned for errors where an input has a wrong companion signature unlock block.
	ErrInputSignatureUnlockBlockInvalid = errors.New("companion signature unlock block is invalid for input")
	// ErrSignatureAndAddrIncompatible gets returned if an address of an input has a companion signature unlock block with the wrong signature type.
	ErrSignatureAndAddrIncompatible = errors.New("address and signature type are not compatible")
	// ErrInvalidInputUnlock gets returned when an input unlock is invalid.
	ErrInvalidInputUnlock = errors.New("invalid input unlock")
	// ErrSenderFeatureBlockNotUnlocked gets returned when an output contains a SenderFeatureBlock with an ident which is not unlocked.
	ErrSenderFeatureBlockNotUnlocked = errors.New("sender feature block is not unlocked")
	// ErrIssuerFeatureBlockNotUnlocked gets returned when an output contains a IssuerFeatureBlock with an ident which is not unlocked.
	ErrIssuerFeatureBlockNotUnlocked = errors.New("issuer feature block is not unlocked")
)

// TransactionID is the ID of a Transaction.
type TransactionID = [TransactionIDLength]byte

// TransactionIDs are IDs of transactions.
type TransactionIDs []TransactionID

// Transaction is a transaction with its inputs, outputs and unlock blocks.
type Transaction struct {
	// The transaction essence, respectively the transfer part of a Transaction.
	Essence *TransactionEssence
	// The unlock blocks defining the unlocking data for the inputs within the Essence.
	UnlockBlocks UnlockBlocks
}

func (t *Transaction) PayloadType() PayloadType {
	return PayloadTransaction
}

// ID computes the ID of the Transaction.
func (t *Transaction) ID() (*TransactionID, error) {
	data, err := t.Serialize(serializer.DeSeriModeNoValidation)
	if err != nil {
		return nil, fmt.Errorf("can't compute transaction ID: %w", err)
	}
	h := blake2b.Sum256(data)
	return &h, nil
}

func (t *Transaction) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	unlockBlockArrayRules := &serializer.ArrayRules{}

	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(PayloadTransaction), serializer.TypeDenotationUint32, func(err error) error {
			return fmt.Errorf("unable to deserialize transaction: %w", err)
		}).
		ReadObject(&t.Essence, deSeriMode, serializer.TypeDenotationByte, TransactionEssenceSelector, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize transaction essence within transaction", err)
		}).
		Do(func() {
			inputCount := uint(len(t.Essence.Inputs))
			unlockBlockArrayRules.Min = inputCount
			unlockBlockArrayRules.Max = inputCount
		}).
		ReadSliceOfObjects(&t.UnlockBlocks, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationByte, UnlockBlockSelector, unlockBlockArrayRules, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize unlock blocks", err)
		}).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				return t.SyntacticallyValidate()
			}
			return nil
		}).
		Done()
}

func (t *Transaction) Serialize(deSeriMode serializer.DeSerializationMode) ([]byte, error) {
	return serializer.NewSerializer().
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				return t.SyntacticallyValidate()
			}
			return nil
		}).
		WriteNum(PayloadTransaction, func(err error) error {
			return fmt.Errorf("%w: unable to serialize transaction payload ID", err)
		}).
		WriteObject(t.Essence, deSeriMode, func(err error) error {
			return fmt.Errorf("%w: unable to serialize transaction's essence", err)
		}).
		WriteSliceOfObjects(&t.UnlockBlocks, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, nil, func(err error) error {
			return fmt.Errorf("%w: unable to serialize transaction's unlock blocks", err)
		}).
		Serialize()
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	jTransaction := &jsonTransaction{
		UnlockBlocks: make([]*json.RawMessage, len(t.UnlockBlocks)),
	}
	jTransaction.Type = int(PayloadTransaction)
	txJson, err := t.Essence.MarshalJSON()
	if err != nil {
		return nil, err
	}
	rawMsgTxJson := json.RawMessage(txJson)
	jTransaction.Essence = &rawMsgTxJson
	for i, ub := range t.UnlockBlocks {
		jsonUB, err := ub.MarshalJSON()
		if err != nil {
			return nil, err
		}
		rawMsgJsonUB := json.RawMessage(jsonUB)
		jTransaction.UnlockBlocks[i] = &rawMsgJsonUB
	}
	return json.Marshal(jTransaction)
}

func (t *Transaction) UnmarshalJSON(bytes []byte) error {
	jTransaction := &jsonTransaction{}
	if err := json.Unmarshal(bytes, jTransaction); err != nil {
		return err
	}
	seri, err := jTransaction.ToSerializable()
	if err != nil {
		return err
	}
	*t = *seri.(*Transaction)
	return nil
}

// SyntacticallyValidate syntactically validates the Transaction:
//	1. The TransactionEssence isn't nil
//	2. syntactic validation on the TransactionEssence
//	3. input and unlock blocks count must match
//	4. signatures are unique and ref. unlock blocks reference a previous unlock block.
func (t *Transaction) SyntacticallyValidate() error {
	switch {
	case t.Essence == nil:
		return fmt.Errorf("%w: transaction is nil", ErrInvalidTransactionEssence)
	case t.UnlockBlocks == nil:
		return fmt.Errorf("%w: unlock blocks are nil", ErrInvalidTransactionEssence)
	}

	if err := t.Essence.SyntacticallyValidate(); err != nil {
		return fmt.Errorf("%w: transaction essence part is invalid", err)
	}

	txID, err := t.ID()
	if err != nil {
		return err
	}

	inputCount := len(t.Essence.Inputs)
	unlockBlockCount := len(t.UnlockBlocks)
	if inputCount != unlockBlockCount {
		return fmt.Errorf("%w: num of inputs %d, num of unlock blocks %d", ErrUnlockBlocksMustMatchInputCount, inputCount, unlockBlockCount)
	}

	if err := ValidateOutputs(t.Essence.Outputs, OutputsPredicateAlias(txID), OutputsPredicateNFT(txID)); err != nil {
		return err
	}

	if err := ValidateUnlockBlocks(t.UnlockBlocks, UnlockBlocksSigUniqueAndRefValidator()); err != nil {
		return fmt.Errorf("%w: invalid unlock blocks", err)
	}

	return nil
}

// SemanticValidationFunc is a function which when called tells whether
// the transaction is passing a specific semantic validation rule or not.
type SemanticValidationFunc = func(t *Transaction, utxos InputSet) error

// SemanticValidationContext defines the context under which a semantic validation for a Transaction is happening.
type SemanticValidationContext struct {
	// The confirming milestone's index.
	ConfirmingMilestoneIndex uint32
	// The confirming milestone's unix seconds timestamp.
	ConfirmingMilestoneUnix uint64

	// The working set which is auto. populated during the semantic validation.
	WorkingSet *SemValiContextWorkingSet
}

// SemValiContextWorkingSet contains fields which get automatically populated
// by the library during the semantic validation of a Transaction.
type SemValiContextWorkingSet struct {
	// The identities which are successfully unlocked from the input side.
	UnlockedIdents UnlockedIdentities
	// The mapping of UTXOInputID to the actual Outputs.
	InputSet InputSet
	// The transaction for which this semantic validation happens.
	Tx *Transaction
	// The message which signatures are signing.
	EssenceMsgtoSign []byte
	// The inputs of the transaction mapped by type.
	InputsByType OutputsByType
	// The ChainConstrainedOutput(s) from the input side which are not new.
	InNonNewChains ChainConstrainedOutputsSet
	// The ChainConstrainedOutput(s) from the input side which are new.
	InNewChains ChainConstrainedOutputsSet
	// The set of ChainConstrainedOutput(s) occurring on the input side.
	InChains ChainConstrainedOutputsSet
	// The Outputs of the transaction mapped by type.
	OutputsByType OutputsByType
	// The ChainConstrainedOutput(s) at the output side which are not new.
	OutNonNewChains ChainConstrainedOutputsSet
	// The UnlockBlocks carried by the transaction mapped by type.
	UnlockBlocksByType UnlockBlocksByType
}

func featureBlockSetFromOutput(output ChainConstrainedOutput) (FeatureBlocksSet, error) {
	featureBlockOutput, is := output.(FeatureBlockOutput)
	if !is {
		return nil, nil
	}

	featureBlocks, err := featureBlockOutput.FeatureBlocks().Set()
	if err != nil {
		return nil, fmt.Errorf("unable to compute feature block set for output: %w", err)
	}
	return featureBlocks, nil
}

func NewSemValiContextWorkingSet(t *Transaction, inputs InputSet) (*SemValiContextWorkingSet, error) {
	var err error
	workingSet := &SemValiContextWorkingSet{}
	workingSet.UnlockedIdents = make(UnlockedIdentities)
	workingSet.InputSet = inputs
	workingSet.Tx = t
	workingSet.EssenceMsgtoSign, err = t.Essence.SigningMessage()
	if err != nil {
		return nil, err
	}
	workingSet.InputsByType = func() OutputsByType {
		slice := make(Outputs, len(inputs))
		var i int
		for _, output := range inputs {
			slice[i] = output
			i++
		}
		return slice.ToOutputsByType()
	}()
	workingSet.InNonNewChains, err = workingSet.InputsByType.NonNewChainConstrainedOutputsSet()
	if err != nil {
		return nil, fmt.Errorf("unable to compute non-new chains (input side): %w", err)
	}
	workingSet.InNewChains, err = workingSet.InputSet.NewChains(SideIn, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to compute new chains (input side): %w", err)
	}
	workingSet.InChains, err = workingSet.InNewChains.Merge(workingSet.InNonNewChains)
	if err != nil {
		return nil, fmt.Errorf("unable to compute input chains set: %w", err)
	}
	workingSet.OutputsByType = t.Essence.Outputs.ToOutputsByType()
	workingSet.OutNonNewChains, err = workingSet.OutputsByType.NonNewChainConstrainedOutputsSet()
	if err != nil {
		return nil, fmt.Errorf("unable to compute non-new chains (output side): %w", err)
	}
	workingSet.UnlockBlocksByType = t.UnlockBlocks.ToUnlockBlocksByType()
	return workingSet, nil
}

// SemanticallyValidate semantically validates the Transaction by checking that the semantic rules applied to the inputs
// and outputs are fulfilled. SyntacticallyValidate() should be called before SemanticallyValidate() to
// ensure that the essence part of the transaction is syntactically valid.
func (t *Transaction) SemanticallyValidate(svCtx *SemanticValidationContext, inputs InputSet, semValFuncs ...SemanticValidationFunc) error {
	var err error
	svCtx.WorkingSet, err = NewSemValiContextWorkingSet(t, inputs)
	if err != nil {
		return err
	}

	// TODO:
	//	- iota token sum balance
	// 	- max 256 native tokens in/out
	// 	- etc.

	// do not change the order of these functions as
	// some of them might depend on certain mutations
	// on the given SemanticValidationContext
	if err := runSemanticValidations(svCtx,
		TxSemanticInputUnlocks(),
		TxSemanticTimelock(),
		TxSemanticOutputsSender(),
		TxSemanticNativeTokens(),
		TxSemanticChainConstrainedSTVF()); err != nil {
		return err
	}

	return nil
}

func runSemanticValidations(svCtx *SemanticValidationContext, checks ...TxSemanticValidationFunc) error {
	for _, check := range checks {
		if err := check(svCtx); err != nil {
			return err
		}
	}
	return nil
}

// UnlockedIdentities defines a set of identities which are unlocked from the input side of a Transaction.
type UnlockedIdentities map[Address]uint16

// TxSemanticValidationFunc is a function which given the context, input, outputs and
// unlock blocks runs a specific semantic validation. The function might also modify the SemanticValidationContext
// in order to supply information to subsequent TxSemanticValidationFunc(s).
type TxSemanticValidationFunc func(svCtx *SemanticValidationContext) error

// TxSemanticInputUnlocks produces the UnlockedIdentities which will be set into the given SemanticValidationContext
// and verifies that inputs are correctly unlocked.
func TxSemanticInputUnlocks() TxSemanticValidationFunc {
	return func(svCtx *SemanticValidationContext) error {
		for inputIndex, inputRef := range svCtx.WorkingSet.Tx.Essence.Inputs {
			input, ok := svCtx.WorkingSet.InputSet[inputRef.(IndexedUTXOReferencer).Ref()]
			if !ok {
				return fmt.Errorf("%w: utxo for input %d not supplied", ErrMissingUTXO, inputIndex)
			}

			if err := unlockOutput(svCtx, input, inputIndex); err != nil {
				return err
			}
		}
		return nil
	}
}

func identToUnlock(svCtx *SemanticValidationContext, input Output, inputIndex int) (Address, error) {
	switch in := input.(type) {
	case SingleIdentOutput:
		return in.Ident()
	case MultiIdentOutput:
		return identToUnlockFromMultiIdentOutput(svCtx, in, inputIndex)
	default:
		panic("unknown ident output type in semantic unlocks")
	}
}

// TODO: abstract this all to work with MultiIdentOutput / ChainID
func identToUnlockFromMultiIdentOutput(svCtx *SemanticValidationContext, inputMultiIdentOutput MultiIdentOutput, inputIndex int) (Address, error) {
	inputAliasOutput, is := inputMultiIdentOutput.(*AliasOutput)
	if !is {
		// this can not happen because only AliasOutput implements MultiIdentOutput
		panic("non alias output is implementing multi ident output in semantic unlocks")
	}

	aliasID := inputAliasOutput.AliasID
	if aliasID.Empty() {
		aliasID = AliasIDFromOutputID(svCtx.WorkingSet.Tx.Essence.Inputs[inputIndex].(IndexedUTXOReferencer).Ref())
	}

	ident := inputAliasOutput.StateController

	// means a governance transition as either state did not change
	// or the alias output is being destroyed
	if outputAliasOutput, has := svCtx.WorkingSet.OutNonNewChains[aliasID]; !has ||
		inputAliasOutput.StateIndex == outputAliasOutput.(*AliasOutput).StateIndex {
		ident = inputAliasOutput.GovernanceController
	}

	return ident, nil
}

func checkSenderFeatureBlockIdentUnlock(svCtx *SemanticValidationContext, output Output) (Address, error) {
	featBlockOutput, is := output.(FeatureBlockOutput)
	if !is {
		return nil, nil
	}

	featBlockSet, err := featBlockOutput.FeatureBlocks().Set()
	if err != nil {
		return nil, err
	}

	featBlockExpMsIndex := featBlockSet[FeatureBlockExpirationMilestoneIndex]
	featBlockExpUnix := featBlockSet[FeatureBlockExpirationUnix]

	if featBlockExpMsIndex == nil && featBlockExpUnix == nil {
		return nil, nil
	}

	// existence guaranteed by syntactical validation
	featBlockSender := featBlockSet[FeatureBlockSender].(*SenderFeatureBlock)

	switch {
	case featBlockExpMsIndex != nil && featBlockExpUnix != nil:
		if featBlockExpMsIndex.(*ExpirationMilestoneIndexFeatureBlock).MilestoneIndex <= svCtx.ConfirmingMilestoneIndex &&
			featBlockExpUnix.(*ExpirationUnixFeatureBlock).UnixTime <= svCtx.ConfirmingMilestoneUnix {
			return featBlockSender.Address, nil
		}
	case featBlockExpMsIndex != nil:
		if featBlockExpMsIndex.(*ExpirationMilestoneIndexFeatureBlock).MilestoneIndex <= svCtx.ConfirmingMilestoneIndex {
			return featBlockSender.Address, nil
		}
	case featBlockExpUnix != nil:
		if featBlockExpUnix.(*ExpirationUnixFeatureBlock).UnixTime <= svCtx.ConfirmingMilestoneUnix {
			return featBlockSender.Address, nil
		}
	}

	return nil, nil
}

func unlockOutput(svCtx *SemanticValidationContext, output Output, inputIndex int) error {
	targetIdent, err := identToUnlock(svCtx, output, inputIndex)
	if err != nil {
		return fmt.Errorf("unable to retrieve ident to unlock of input %d: %w", inputIndex, err)
	}

	actualIdentToUnlock, err := checkSenderFeatureBlockIdentUnlock(svCtx, output)
	if err != nil {
		return err
	}
	if actualIdentToUnlock != nil {
		targetIdent = actualIdentToUnlock
	}

	unlockBlock := svCtx.WorkingSet.Tx.UnlockBlocks[inputIndex]

	switch ident := targetIdent.(type) {
	case ChainConstrainedAddress:
		referentialUnlockBlock, isReferentialUnlockBlock := unlockBlock.(ReferentialUnlockBlock)
		if !isReferentialUnlockBlock || !referentialUnlockBlock.Chainable() || !referentialUnlockBlock.SourceAllowed(targetIdent) {
			return fmt.Errorf("%w: input %d has an chain constrained address of %s but its corresponding unlock block is of type %s", ErrInvalidInputUnlock, inputIndex, AddressTypeToString(ident.Type()), UnlockBlockTypeToString(unlockBlock.Type()))
		}

		unlockedAtIndex, wasUnlocked := svCtx.WorkingSet.UnlockedIdents[ident]
		if !wasUnlocked || unlockedAtIndex != referentialUnlockBlock.Ref() {
			return fmt.Errorf("%w: input %d's chain constrained address is not unlocked through input %d's unlock block", ErrInvalidInputUnlock, inputIndex, referentialUnlockBlock.Ref())
		}

		// since this input is now unlocked, and it has a ChainConstrainedAddress, it becomes automatically unlocked
		if chainConstrainedOutput, isChainConstrainedOutput := output.(ChainConstrainedOutput); isChainConstrainedOutput && chainConstrainedOutput.Chain().Addressable() {
			svCtx.WorkingSet.UnlockedIdents[chainConstrainedOutput.Chain().ToAddress()] = uint16(inputIndex)
		}

	case DirectUnlockableAddress:
		switch uBlock := unlockBlock.(type) {
		case ReferentialUnlockBlock:
			if uBlock.Chainable() || !uBlock.SourceAllowed(targetIdent) {
				return fmt.Errorf("%w: input %d has not chain constrained address of %s but its corresponding unlock block is of type %s", ErrInvalidInputUnlock, inputIndex, AddressTypeToString(ident.Type()), UnlockBlockTypeToString(unlockBlock.Type()))
			}

			unlockedAtIndex, wasUnlocked := svCtx.WorkingSet.UnlockedIdents[ident]
			if !wasUnlocked || unlockedAtIndex != uBlock.Ref() {
				return fmt.Errorf("%w: input %d's address is not unlocked through input %d's unlock block", ErrInvalidInputUnlock, inputIndex, uBlock.Ref())
			}
		case *SignatureUnlockBlock:
			// ident must not be unlocked already
			if unlockedAtIndex, wasAlreadyUnlocked := svCtx.WorkingSet.UnlockedIdents[ident]; wasAlreadyUnlocked {
				return fmt.Errorf("%w: input %d's address is already unlocked through input %d's unlock block but the input uses a non referential unlock block", ErrInvalidInputUnlock, inputIndex, unlockedAtIndex)
			}

			if err := ident.Unlock(svCtx.WorkingSet.EssenceMsgtoSign, uBlock.Signature); err != nil {
				return fmt.Errorf("%w: input %d's address is not unlocked through its signature unlock block", err, inputIndex)
			}

			svCtx.WorkingSet.UnlockedIdents[ident] = uint16(inputIndex)
		}
	default:
		panic("unknown address in semantic unlocks")
	}
	return nil
}

// TxSemanticOutputsSender validates that for SenderFeatureBlock occurring on the output side,
// the given identity is unlocked on the input side.
func TxSemanticOutputsSender() TxSemanticValidationFunc {
	return func(svCtx *SemanticValidationContext) error {
		for outputIndex, output := range svCtx.WorkingSet.Tx.Essence.Outputs {
			featureBlockOutput, is := output.(FeatureBlockOutput)
			if !is {
				continue
			}

			featureBlocks, err := featureBlockOutput.FeatureBlocks().Set()
			if err != nil {
				return fmt.Errorf("unable to compute feature block set for output %d: %w", outputIndex, err)
			}

			senderFeatureBlock, has := featureBlocks[FeatureBlockSender]
			if !has {
				continue
			}

			// check unlocked
			sender := senderFeatureBlock.(*SenderFeatureBlock).Address
			if _, isUnlocked := svCtx.WorkingSet.UnlockedIdents[sender]; !isUnlocked {
				return fmt.Errorf("%w: output %d", ErrSenderFeatureBlockNotUnlocked, outputIndex)
			}
		}
		return nil
	}
}

// TxSemanticTimelock validates following rules regarding timelocked inputs:
//	- Inputs with a TimelockMilestone<Index,Unix>FeatureBlock can only be unlocked if the confirming milestone allows it.
func TxSemanticTimelock() TxSemanticValidationFunc {
	return func(svCtx *SemanticValidationContext) error {
		for inputIndex, input := range svCtx.WorkingSet.InputSet {
			inputWithFeatureBlocks, is := input.(FeatureBlockOutput)
			if !is {
				continue
			}
			for _, featureBlock := range inputWithFeatureBlocks.FeatureBlocks() {
				switch block := featureBlock.(type) {
				case *TimelockMilestoneIndexFeatureBlock:
					if svCtx.ConfirmingMilestoneIndex < block.MilestoneIndex {
						return fmt.Errorf("%w: input at index %d's milestone index timelock is not expired, at %d, current %d", ErrInvalidInputUnlock, inputIndex, block.MilestoneIndex, svCtx.ConfirmingMilestoneIndex)
					}
				case *TimelockUnixFeatureBlock:
					if svCtx.ConfirmingMilestoneUnix < block.UnixTime {
						return fmt.Errorf("%w: input at index %d's unix timelock is not expired, at %d, current %d", ErrInvalidInputUnlock, inputIndex, block.UnixTime, svCtx.ConfirmingMilestoneUnix)
					}
				}
			}
		}
		return nil
	}
}

// TxSemanticChainConstrainedSTVF validates following rules regarding ChainConstrainedOutput(s):
//	- For output AliasOutput(s) with non-zeroed AliasID, there must be a corresponding input AliasOutput where either
//	  its AliasID is zeroed and StateIndex and FoundryCounter are zero or an input AliasOutput with the same AliasID.
//	- On alias state transitions:
//		- The StateIndex must be incremented by 1
//		- Only Amount, NativeTokens, StateIndex, StateMetadata and FoundryCounter can be mutated
//	- On alias governance transition:
//		- Only StateController (must be mutated), GovernanceController and the MetadataBlock can be mutated
func TxSemanticChainConstrainedSTVF() TxSemanticValidationFunc {
	return func(svCtx *SemanticValidationContext) error {
		// for every non-new chain on the output side, there must be a corresponding chain on the input side (new or existing)
		if err := svCtx.WorkingSet.OutNonNewChains.Includes(svCtx.WorkingSet.InChains); err != nil {
			return fmt.Errorf("missing chain on the input side to satisfy non-new chains on the output side: %w", err)
		}

		// state change transitions
		if err := svCtx.WorkingSet.InChains.EveryTuple(svCtx.WorkingSet.OutNonNewChains, ValidateStateTransitionOnTuple(svCtx)); err != nil {
			// TODO: fmt.Errorf
			return err
		}

		// new state transitions
		for outputIndex, output := range svCtx.WorkingSet.Tx.Essence.Outputs {
			chainConstrainedOutput, is := output.(ChainConstrainedOutput)
			if !is {
				continue
			}

			isNewChain, err := chainConstrainedOutput.IsNewChain(SideOut, svCtx)
			if err != nil {
				// TODO: fmt.Errorf
				return err
			}

			if !isNewChain {
				continue
			}

			if err := chainConstrainedOutput.ValidateStateTransition(ChainTransitionTypeNew, nil, svCtx); err != nil {
				return fmt.Errorf("new chain constrained output %d state transition failed: %w", outputIndex, err)
			}
		}

		return nil
	}
}

// TxSemanticNativeTokens validates following rules regarding NativeTokens:
//	- The NativeTokens between Inputs / Outputs must be balanced in terms of circulating supply adjustments.
//	- Within transitioning FoundryOutput(s) only the circulating supply and contained NativeTokens can change.
//	- TODO: The newly created FoundryOutput(s) belonging to an alias account must be sorted on the output side according to their
//	  serial number and must fill the gap between the alias account's starting/closing(inclusive) foundry counter.
func TxSemanticNativeTokens() TxSemanticValidationFunc {
	return func(svCtx *SemanticValidationContext) error {
		inputNativeTokens, err := svCtx.WorkingSet.InputsByType.NativeTokenOutputs().Sum()
		if err != nil {
			return fmt.Errorf("invalid input native token set: %w", err)
		}
		outputNativeTokens, err := svCtx.WorkingSet.InputsByType.NativeTokenOutputs().Sum()
		if err != nil {
			return fmt.Errorf("invalid output native token set: %w", err)
		}

		// easy route, tokens must be balanced between both sets
		_, hasOutputFoundryOutput := svCtx.WorkingSet.OutputsByType[OutputFoundry]
		_, hasInputFoundryOutput := svCtx.WorkingSet.InputsByType[OutputFoundry]
		if !hasInputFoundryOutput && !hasOutputFoundryOutput {
			if err := inputNativeTokens.Balanced(outputNativeTokens); err != nil {
				return err
			}
		}

		inputFoundryOutputs, err := svCtx.WorkingSet.InputsByType.FoundryOutputsSet()
		if err != nil {
			return fmt.Errorf("invalid input foundry outputs: %w", err)
		}
		outputFoundryOutputs, err := svCtx.WorkingSet.OutputsByType.FoundryOutputsSet()
		if err != nil {
			return fmt.Errorf("invalid output foundry outputs: %w", err)
		}

		diffs, err := inputFoundryOutputs.Diff(outputFoundryOutputs)
		if err != nil {
			return fmt.Errorf("unable to compute foundry outputs diff: %w", err)
		}

		// TODO: diffs.CheckFoundryOutputsSerialNumber()

		if err := inputNativeTokens.BalancedWithDiffs(outputNativeTokens, diffs); err != nil {
			return err
		}

		return nil
	}
}

// jsonTransaction defines the json representation of a Transaction.
type jsonTransaction struct {
	Type         int                `json:"type"`
	Essence      *json.RawMessage   `json:"essence"`
	UnlockBlocks []*json.RawMessage `json:"unlockBlocks"`
}

func (jsontx *jsonTransaction) ToSerializable() (serializer.Serializable, error) {
	jsonTxEssence, err := DeserializeObjectFromJSON(jsontx.Essence, jsonTransactionEssenceSelector)
	if err != nil {
		return nil, fmt.Errorf("unable to decode transaction essence from JSON: %w", err)
	}

	txEssenceSeri, err := jsonTxEssence.ToSerializable()
	if err != nil {
		return nil, err
	}

	unlockBlocks, err := unlockBlocksFromJSONRawMsg(jsontx.UnlockBlocks)
	if err != nil {
		return nil, err
	}

	return &Transaction{Essence: txEssenceSeri.(*TransactionEssence), UnlockBlocks: unlockBlocks}, nil
}
