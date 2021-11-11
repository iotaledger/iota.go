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
)

// TransactionID is the ID of a Transaction.
type TransactionID = [TransactionIDLength]byte

// TransactionIDs are IDs of transactions.
type TransactionIDs []TransactionID

// Transaction is a transaction with its inputs, outputs and unlock blocks.
type Transaction struct {
	// The transaction essence, respectively the transfer part of a Transaction.
	Essence serializer.Serializable
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
			inputCount := uint(len(t.Essence.(*TransactionEssence).Inputs))
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

	txEssence, ok := t.Essence.(*TransactionEssence)
	if !ok {
		return fmt.Errorf("%w: transaction essence is not *TransactionEssence", ErrInvalidTransactionEssence)
	}

	if err := txEssence.SyntacticallyValidate(); err != nil {
		return fmt.Errorf("%w: transaction essence part is invalid", err)
	}

	txID, err := t.ID()
	if err != nil {
		return err
	}

	inputCount := len(txEssence.Inputs)
	unlockBlockCount := len(t.UnlockBlocks)
	if inputCount != unlockBlockCount {
		return fmt.Errorf("%w: num of inputs %d, num of unlock blocks %d", ErrUnlockBlocksMustMatchInputCount, inputCount, unlockBlockCount)
	}

	if err := ValidateOutputs(txEssence.Outputs, OutputsPredicateAlias(txID), OutputsPredicateNFT(txID)); err != nil {
		return err
	}

	if err := ValidateUnlockBlocks(t.UnlockBlocks, UnlockBlocksSigUniqueAndRefValidator()); err != nil {
		return fmt.Errorf("%w: invalid unlock blocks", err)
	}

	return nil
}

// SigValidationFunc is a function which when called tells whether
// its signature verification computation was successful or not.
type SigValidationFunc = func() error

// SemanticValidationFunc is a function which when called tells whether
// the transaction is passing a specific semantic validation rule or not.
type SemanticValidationFunc = func(t *Transaction, utxos InputSet) error

// SemanticValidationContext defines the context under which a semantic validation for a Transaction is happening.
type SemanticValidationContext struct {
	// The confirming milestone's index.
	ConfirmingMilestoneIndex uint32
	// The confirming milestone's unix seconds timestamp.
	ConfirmingMilestoneUnix uint64

	// fields filled by iota.go
	unlockedIdents     UnlockedIdentities
	inputSet           InputSet
	tx                 *Transaction
	essence            *TransactionEssence
	essenceBytes       []byte
	inputsByType       OutputsByType
	outputsByType      OutputsByType
	unlockBlocksByType UnlockBlocksByType
}

// SemanticallyValidate semantically validates the Transaction by checking that the semantic rules applied to the inputs
// and outputs are fulfilled. SyntacticallyValidate() should be called before SemanticallyValidate() to
// ensure that the essence part of the transaction is syntactically valid.
func (t *Transaction) SemanticallyValidate(svCtx *SemanticValidationContext, inputs InputSet, semValFuncs ...SemanticValidationFunc) error {

	txEssence, ok := t.Essence.(*TransactionEssence)
	if !ok {
		return fmt.Errorf("%w: transaction is not *TransactionEssence", ErrInvalidTransactionEssence)
	}

	txEssenceBytes, err := txEssence.SigningMessage()
	if err != nil {
		return err
	}

	svCtx.unlockedIdents = make(UnlockedIdentities)
	svCtx.inputSet = inputs
	svCtx.tx = t
	svCtx.essence = txEssence
	svCtx.essenceBytes = txEssenceBytes
	svCtx.inputsByType = func() OutputsByType {
		slice := make(Outputs, len(inputs))
		var i int
		for _, output := range inputs {
			slice[i] = output
			i++
		}
		return slice.ToOutputsByType()
	}()
	svCtx.outputsByType = txEssence.Outputs.ToOutputsByType()
	svCtx.unlockBlocksByType = t.UnlockBlocks.ToUnlockBlocksByType()

	inputSum, sigValidFuncs, err := t.SemanticallyValidateInputs(inputs, txEssence, txEssenceBytes)
	if err != nil {
		return err
	}

	outputSum, err := t.SemanticallyValidateOutputs(txEssence)
	if err != nil {
		return err
	}

	if inputSum != outputSum {
		return fmt.Errorf("%w: inputs sum %d, outputs sum %d", ErrInputOutputSumMismatch, inputSum, outputSum)
	}

	for _, semValFunc := range semValFuncs {
		if err := semValFunc(t, inputs); err != nil {
			return err
		}
	}

	// sig verifications runs at the end as they are the most computationally expensive operation
	for _, f := range sigValidFuncs {
		if err := f(); err != nil {
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
		for inputIndex, inputRef := range svCtx.essence.Inputs {
			input, ok := svCtx.inputSet[inputRef.(IndexedUTXOReferencer).Ref()]
			if !ok {
				return fmt.Errorf("%w: utxo for input %d not supplied", ErrMissingUTXO, inputIndex)
			}

			switch in := input.(type) {
			case SingleIdentOutput:
				if err := unlockSingleIdentOutput(svCtx, in, inputIndex); err != nil {
					return err
				}
			case MultiIdentOutput:
				if err := unlockMultiIdentOutput(svCtx, in, inputIndex); err != nil {
					return err
				}
			default:
				panic("unknown ident output in semantic unlocks")
			}

		}
		return nil
	}
}

func unlockMultiIdentOutput(svCtx *SemanticValidationContext, multiIdentOutput MultiIdentOutput, inputIndex int) error {
	 svCtx.outputsByType.MultiIdentOutputsSet()
	outputNonNewAliases, err := svCtx.outputsByType.NonNewAliasOutputsSet()
	if err != nil {
		return err
	}
	accountID := multiIdentOutput.Account()
	outputNonNewAliases[]
}

func unlockSingleIdentOutput(svCtx *SemanticValidationContext, singleIdentInput SingleIdentOutput, inputIndex int) error {
	identToUnlock, err := singleIdentInput.Ident()
	if err != nil {
		return fmt.Errorf("unable to retrieve ident of input %d: %w", inputIndex, err)
	}

	// TODO: examine feature block constraints

	unlockBlock := svCtx.tx.UnlockBlocks[inputIndex]

	switch ident := identToUnlock.(type) {
	case AccountAddress:
		referentialUnlockBlock, isReferentialUnlockBlock := unlockBlock.(ReferentialUnlockBlock)
		if !isReferentialUnlockBlock || !referentialUnlockBlock.Chainable() || !referentialUnlockBlock.SourceAllowed(identToUnlock) {
			return fmt.Errorf("%w: input %d has an account address of %s but its corresponding unlock block is of type %s", ErrInvalidInputUnlock, inputIndex, AddressTypeToString(ident.Type()), UnlockBlockTypeToString(unlockBlock.Type()))
		}

		unlockedAtIndex, wasUnlocked := svCtx.unlockedIdents[ident]
		if !wasUnlocked || unlockedAtIndex != referentialUnlockBlock.Ref() {
			return fmt.Errorf("%w: input %d's account address is not unlocked through input %d's unlock block", ErrInvalidInputUnlock, inputIndex, referentialUnlockBlock.Ref())
		}

		// since this input is now unlocked, and it has an AccountAddress, it becomes automatically unlocked
		if accountOutput, isAccountOutput := singleIdentInput.(AccountOutput); isAccountOutput {
			svCtx.unlockedIdents[accountOutput.Account().ToAddress()] = uint16(inputIndex)
		}

	case DirectUnlockableAddress:
		switch uBlock := unlockBlock.(type) {
		case ReferentialUnlockBlock:
			if uBlock.Chainable() || !uBlock.SourceAllowed(identToUnlock) {
				return fmt.Errorf("%w: input %d has none account address of %s but its corresponding unlock block is of type %s", ErrInvalidInputUnlock, inputIndex, AddressTypeToString(ident.Type()), UnlockBlockTypeToString(unlockBlock.Type()))
			}

			unlockedAtIndex, wasUnlocked := svCtx.unlockedIdents[ident]
			if !wasUnlocked || unlockedAtIndex != uBlock.Ref() {
				return fmt.Errorf("%w: input %d's address is not unlocked through input %d's unlock block", ErrInvalidInputUnlock, inputIndex, uBlock.Ref())
			}
		case *SignatureUnlockBlock:
			// ident must not be unlocked already
			if _, wasAlreadyUnlocked := svCtx.unlockedIdents[ident]; wasAlreadyUnlocked {
				return fmt.Errorf("%w: input %d's address is already unlocked through input %d's unlock block but the input uses a non referential unlock block", ErrInvalidInputUnlock, inputIndex, uBlock.Ref())
			}

			if err := ident.Unlock(svCtx.essenceBytes, uBlock.Signature); err != nil {
				return fmt.Errorf("%w: input %d's address is not unlocked through its signature unlock block", err, inputIndex)
			}

			svCtx.unlockedIdents[ident] = uint16(inputIndex)
		}
	default:
		panic("unknown address in semantic unlocks")
	}
	return nil
}

// TxSemanticTimelock validates following rules regarding timelocked inputs:
//	- Inputs with a TimelockMilestone<Index,Unix>FeatureBlock can only be unlocked if the confirming milestone allows it.
func TxSemanticTimelock(TxSemanticValidationFunc) TxSemanticValidationFunc {
	return func(svCtx *SemanticValidationContext) error {
		for inputIndex, input := range svCtx.inputSet {
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

// TxSemanticAlias validates following rules regarding aliases:
//	- For output AliasOutput(s) with non-zeroed AliasID, there must be a corresponding input AliasOutput where either
//	  its AliasID is zeroed and StateIndex and FoundryCounter are zero or an input AliasOutput with the same AliasID.
//	- On alias state transitions:
//		- The StateIndex must be incremented by 1
//		- Only Amount, NativeTokens, StateIndex, StateMetadata and FoundryCounter can be mutated
//	- On alias governance transition:
//		- Only StateController (must be mutated), GovernanceController and the MetadataBlock can be mutated
func TxSemanticAlias() TxSemanticValidationFunc {
	return func(svCtx *SemanticValidationContext) error {
		inputsNewAliases := svCtx.inputSet.NewAliases()

		outputsNonNewAliases, err := svCtx.outputsByType.NonNewAliasOutputsSet()
		if err != nil {
			return fmt.Errorf("unable to compute non-new aliases (output side): %w", err)
		}

		inputsNonNewAliases, err := svCtx.inputsByType.NonNewAliasOutputsSet()
		if err != nil {
			return fmt.Errorf("unable to compute non-new aliases (input side): %w", err)
		}

		inputAliases, err := inputsNewAliases.Merge(inputsNonNewAliases)
		if err != nil {
			return fmt.Errorf("unable to compute alias input set: %w", err)
		}

		// for every non-new alias, there must be a corresponding alias on the input side (new or existing)
		if err := outputsNonNewAliases.Includes(inputAliases); err != nil {
			return fmt.Errorf("missing aliases on the input side to satisfy non-new aliases on the output side: %w", err)
		}

		if err := inputAliases.EveryTuple(outputsNonNewAliases, func(in *AliasOutput, out *AliasOutput) error {
			return in.TransitionWith(out)
		}); err != nil {
			return err
		}

		return nil
	}
}

// TxSemanticNativeTokens validates following rules regarding NativeTokens:
//	- The NativeTokens between Inputs / Outputs must be balanced in terms of circulating supply adjustments.
//	- Within transitioning FoundryOutput(s) only the circulating supply and contained NativeTokens can change.
//	- TODO: The newly created FoundryOutput(s) belonging to an alias account must be sorted on the output side according to their
//	  serial number and must fill the gap between the alias account's starting/closing(inclusive) foundry counter.
func TxSemanticNativeTokens(originInputs Outputs, originOutput Outputs) TxSemanticValidationFunc {
	return func(svCtx *SemanticValidationContext) error {
		inputNativeTokens, err := svCtx.inputsByType.NativeTokenOutputs().Sum()
		if err != nil {
			return fmt.Errorf("invalid input native token set: %w", err)
		}
		outputNativeTokens, err := svCtx.inputsByType.NativeTokenOutputs().Sum()
		if err != nil {
			return fmt.Errorf("invalid output native token set: %w", err)
		}

		// easy route, tokens must be balanced between both sets
		_, hasOutputFoundryOutput := svCtx.outputsByType[OutputFoundry]
		_, hasInputFoundryOutput := svCtx.inputsByType[OutputFoundry]
		if !hasInputFoundryOutput && !hasOutputFoundryOutput {
			if err := inputNativeTokens.Balanced(outputNativeTokens); err != nil {
				return err
			}
		}

		inputFoundryOutputs, err := svCtx.inputsByType.FoundryOutputsSet()
		if err != nil {
			return fmt.Errorf("invalid input foundry outputs: %w", err)
		}
		outputFoundryOutputs, err := svCtx.outputsByType.FoundryOutputsSet()
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

// SemanticallyValidateInputs checks that every referenced UTXO is available, computes the input sum
// and returns functions which can be called to verify the signatures.
// This function should only be called from SemanticallyValidate().
func (t *Transaction) SemanticallyValidateInputs(inputs InputSet, essence *TransactionEssence, txEssenceBytes []byte) (uint64, []SigValidationFunc, error) {
	var sigValidFuncs []SigValidationFunc
	var inputSum uint64
	seenInputAddr := make(map[string]int)

	for i, input := range essence.Inputs {
		utxoInput, isUTXOInput := input.(IndexedUTXOReferencer)
		if !isUTXOInput {
			return 0, nil, fmt.Errorf("%w: unsupported input type at index %d", ErrUnknownInputType, i)
		}

		// check that we got the needed UTXO
		utxoID := utxoInput.Ref()
		input, has := inputs[utxoID]
		if !has {
			return 0, nil, fmt.Errorf("%w: UTXO for ID %v is not provided (input at index %d)", ErrMissingUTXO, utxoID, i)
		}

		var err error
		deposit, err := input.Deposit()
		if err != nil {
			return 0, nil, fmt.Errorf("unable to get deposit from UTXO %v (input at index %d): %w", utxoID, i, err)
		}
		inputSum += deposit

		sigBlock, sigBlockIndex, err := t.signatureUnlockBlock(i)
		if err != nil {
			return 0, nil, err
		}

		target, err := input.Target()
		if err != nil {
			return 0, nil, fmt.Errorf("unable to get target for UTXO %v: %w", utxoID, err)
		}

		// change this logic here once we got tx output types without addrs
		addr, isAddr := target.(Address)
		if !isAddr {
			return 0, nil, fmt.Errorf("target for UTXO %v must be an address: %w", utxoID, err)
		}

		usedSigBlockIndex, alreadySeen := seenInputAddr[addr.String()]
		if alreadySeen {
			if usedSigBlockIndex != sigBlockIndex {
				return 0, nil, fmt.Errorf("%w: target for UTXO %v uses a different signature unlock block (%d) than a previous UTXO (%d) for the same address", ErrInputSignatureUnlockBlockInvalid, utxoID, sigBlockIndex, usedSigBlockIndex)
			}
			// we can skip here as we already created a sig validation func
			continue
		}

		sigValidF, err := createSigValidationFunc(i, sigBlock.Signature, sigBlockIndex, txEssenceBytes, addr)
		if err != nil {
			return 0, nil, err
		}

		seenInputAddr[addr.String()] = sigBlockIndex

		sigValidFuncs = append(sigValidFuncs, sigValidF)
	}

	return inputSum, sigValidFuncs, nil
}

// retrieves the SignatureUnlockBlock at the given index or follows
// the reference of a ReferenceUnlockBlock to retrieve it.
func (t *Transaction) signatureUnlockBlock(index int) (*SignatureUnlockBlock, int, error) {
	// indexation valid via SyntacticallyValidate()
	switch ub := t.UnlockBlocks[index].(type) {
	case *SignatureUnlockBlock:
		return ub, index, nil
	case *ReferenceUnlockBlock:
		// it is ensured by the syntactical validation that
		// the corresponding signature unlock block exists
		sigUBIndex := int(ub.Reference)
		return t.UnlockBlocks[sigUBIndex].(*SignatureUnlockBlock), sigUBIndex, nil
	default:
		return nil, 0, fmt.Errorf("%w: unsupported unlock block type at index %d", ErrUnknownUnlockBlockType, index)
	}
}

// creates a SigValidationFunc appropriate for the underlying signature type.
func createSigValidationFunc(pos int, sig serializer.Serializable, sigBlockIndex int, txEssenceBytes []byte, addr Address) (SigValidationFunc, error) {
	switch addr := addr.(type) {
	case *Ed25519Address:
		return createEd25519SigValidationFunc(pos, sig, sigBlockIndex, addr, txEssenceBytes)
	case *BLSAddress:
		return createBLSSigValidationFunc(pos, sig, sigBlockIndex, addr, txEssenceBytes)
	default:
		return nil, fmt.Errorf("%w: unsupported address type at index %d", ErrUnknownAddrType, pos)
	}
}

// creates a SigValidationFunc validating the given Ed25519Signature against the Ed25519Address.
func createEd25519SigValidationFunc(pos int, sig serializer.Serializable, sigBlockIndex int, addr *Ed25519Address, essenceBytes []byte) (SigValidationFunc, error) {
	ed25519Sig, isEd25519Sig := sig.(*Ed25519Signature)
	if !isEd25519Sig {
		return nil, fmt.Errorf("%w: UTXO at index %d has an Ed25519 address but its corresponding signature is of type %T (at index %d)", ErrSignatureAndAddrIncompatible, pos, sig, sigBlockIndex)
	}

	return func() error {
		if err := ed25519Sig.Valid(essenceBytes, addr); err != nil {
			return fmt.Errorf("%w: input at index %d, signature block at index %d", err, pos, sigBlockIndex)
		}
		return nil
	}, nil
}

// creates a SigValidationFunc validating the given BLSSignature against the BLSAddress.
func createBLSSigValidationFunc(pos int, sig serializer.Serializable, sigBlockIndex int, addr *BLSAddress, essenceBytes []byte) (SigValidationFunc, error) {
	blsSig, isBLSSig := sig.(*BLSSignature)
	if !isBLSSig {
		return nil, fmt.Errorf("%w: UTXO at index %d has a BLS address but its corresponding signature is of type %T (at index %d)", ErrSignatureAndAddrIncompatible, pos, sig, sigBlockIndex)
	}

	return func() error {
		if err := blsSig.Valid(essenceBytes, addr); err != nil {
			return fmt.Errorf("%w: input at index %d, signature block at index %d", err, pos, sigBlockIndex)
		}
		return nil
	}, nil
}

// SemanticallyValidateOutputs accumulates the sum of all outputs.
// This function should only be called from SemanticallyValidate().
func (t *Transaction) SemanticallyValidateOutputs(transaction *TransactionEssence) (uint64, error) {
	var outputSum uint64
	for i, output := range transaction.Outputs {
		out, ok := output.(Output)
		if !ok {
			return 0, fmt.Errorf("%w: unsupported output type at index %d", ErrUnknownOutputType, i)
		}
		deposit, err := out.Deposit()
		if err != nil {
			return 0, fmt.Errorf("unable to get deposit from output at index %d: %w", i, err)
		}
		outputSum += deposit
	}

	return outputSum, nil
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

	return &Transaction{Essence: txEssenceSeri, UnlockBlocks: unlockBlocks}, nil
}
