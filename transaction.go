package iotago

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"

	"golang.org/x/crypto/blake2b"
)

const (
	// TransactionIDLength defines the length of a Transaction ID.
	TransactionIDLength = blake2b.Size256
)

var (
	// ErrMissingUTXO gets returned if an UTXO is missing to commence a certain operation.
	ErrMissingUTXO = errors.New("missing utxo")
	// ErrInputOutputSumMismatch gets returned if a transaction does not spend the entirety of the inputs to the outputs.
	ErrInputOutputSumMismatch = errors.New("inputs and outputs do not spend/deposit the same amount")
	// ErrSignatureAndAddrIncompatible gets returned if an address of an input has a companion signature unlock block with the wrong signature type.
	ErrSignatureAndAddrIncompatible = errors.New("address and signature type are not compatible")
	// ErrInvalidInputUnlock gets returned when an input unlock is invalid.
	ErrInvalidInputUnlock = errors.New("invalid input unlock")
	// ErrSenderFeatureBlockNotUnlocked gets returned when an output contains a SenderFeatureBlock with an ident which is not unlocked.
	ErrSenderFeatureBlockNotUnlocked = errors.New("sender feature block is not unlocked")
	// ErrIssuerFeatureBlockNotUnlocked gets returned when an output contains a IssuerFeatureBlock with an ident which is not unlocked.
	ErrIssuerFeatureBlockNotUnlocked = errors.New("issuer feature block is not unlocked")
	// ErrReturnAmountNotFulFilled gets returned when a return amount in a transaction is not fulfilled by the output side.
	ErrReturnAmountNotFulFilled = errors.New("return amount not fulfilled")
	// ErrTypeIsNotSupportedEssence gets returned when a serializable was found to not be a supported essence.
	ErrTypeIsNotSupportedEssence = errors.New("serializable is not a supported essence")

	txEssenceGuard = serializer.SerializableGuard{
		ReadGuard: func(ty uint32) (serializer.Serializable, error) {
			return TransactionEssenceSelector(ty)
		},
		WriteGuard: func(seri serializer.Serializable) error {
			if seri == nil {
				return fmt.Errorf("%w: because nil", ErrTypeIsNotSupportedEssence)
			}
			if _, is := seri.(*TransactionEssence); !is {
				return fmt.Errorf("%w: because not *TransactionEssence", ErrTypeIsNotSupportedEssence)
			}
			return nil
		},
	}
	txUnlockBlockArrayRules = serializer.ArrayRules{
		// min/max filled out in serialize/deserialize
		Guards: serializer.SerializableGuard{
			ReadGuard:  UnlockBlockSelector,
			WriteGuard: unlockBlockWriteGuard(),
		},
	}
)

// TransactionUnlockBlocksArrayRules returns array rules defining the constraints on UnlockBlocks within a Transaction.
func TransactionUnlockBlocksArrayRules() serializer.ArrayRules {
	return txUnlockBlockArrayRules
}

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

// OutputsSet returns an OutputSet from the Transaction's outputs, mapped by their OutputID.
func (t *Transaction) OutputsSet() (OutputSet, error) {
	txID, err := t.ID()
	if err != nil {
		return nil, err
	}
	set := make(OutputSet)
	for index, output := range t.Essence.Outputs {
		set[OutputIDFromTransactionIDAndIndex(*txID, uint16(index))] = output
	}
	return set, nil
}

// ID computes the ID of the Transaction.
func (t *Transaction) ID() (*TransactionID, error) {
	data, err := t.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return nil, fmt.Errorf("can't compute transaction ID: %w", err)
	}
	h := blake2b.Sum256(data)
	return &h, nil
}

func (t *Transaction) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	unlockBlockArrayRulesCopy := txUnlockBlockArrayRules
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(PayloadTransaction), serializer.TypeDenotationUint32, func(err error) error {
			return fmt.Errorf("unable to deserialize transaction: %w", err)
		}).
		ReadObject(&t.Essence, deSeriMode, deSeriCtx, serializer.TypeDenotationByte, txEssenceGuard.ReadGuard, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize transaction essence within transaction", err)
		}).
		Do(func() {
			inputCount := uint(len(t.Essence.Inputs))
			unlockBlockArrayRulesCopy.Min = inputCount
			unlockBlockArrayRulesCopy.Max = inputCount
		}).
		ReadSliceOfObjects(&t.UnlockBlocks, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationByte, &unlockBlockArrayRulesCopy, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize unlock blocks", err)
		}).
		WithValidation(deSeriMode, txDeSeriValidation(t, deSeriCtx)).
		Done()
}

func (t *Transaction) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	unlockBlockArrayRulesCopy := txUnlockBlockArrayRules
	inputCount := uint(len(t.Essence.Inputs))
	unlockBlockArrayRulesCopy.Min = inputCount
	unlockBlockArrayRulesCopy.Max = inputCount
	return serializer.NewSerializer().
		WriteNum(PayloadTransaction, func(err error) error {
			return fmt.Errorf("%w: unable to serialize transaction payload ID", err)
		}).
		WriteObject(t.Essence, deSeriMode, deSeriCtx, txEssenceGuard.WriteGuard, func(err error) error {
			return fmt.Errorf("%w: unable to serialize transaction's essence", err)
		}).
		WriteSliceOfObjects(&t.UnlockBlocks, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsUint16, &unlockBlockArrayRulesCopy, func(err error) error {
			return fmt.Errorf("%w: unable to serialize transaction's unlock blocks", err)
		}).
		WithValidation(deSeriMode, txDeSeriValidation(t, deSeriCtx)).
		Serialize()
}

func (t *Transaction) Size() int {
	return util.NumByteLen(uint32(PayloadTransaction)) +
		t.Essence.Size() +
		2 + t.UnlockBlocks.Size() //2 bytes length prefix
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

func txDeSeriValidation(tx *Transaction, deSeriCtx interface{}) serializer.ErrProducerWithRWBytes {
	return func(readBytes []byte, err error) error {
		deSeriParas, ok := deSeriCtx.(*DeSerializationParameters)
		if !ok || deSeriParas == nil {
			return fmt.Errorf("unable to validate transaction: %w", ErrMissingDeSerializationParas)
		}
		return tx.syntacticallyValidate(readBytes, deSeriParas.RentStructure)
	}
}

// syntacticallyValidate syntactically validates the Transaction.
func (t *Transaction) syntacticallyValidate(readBytes []byte, rentStruct *RentStructure) error {
	if err := t.Essence.syntacticallyValidate(rentStruct); err != nil {
		return fmt.Errorf("transaction essence is invalid: %w", err)
	}

	txID := blake2b.Sum256(readBytes)
	if err := ValidateOutputs(t.Essence.Outputs,
		OutputsSyntacticalAlias(&txID),
		OutputsSyntacticalNFT(&txID),
	); err != nil {
		return err
	}

	if err := ValidateUnlockBlocks(t.UnlockBlocks,
		UnlockBlocksSigUniqueAndRefValidator(),
	); err != nil {
		return fmt.Errorf("invalid unlock blocks: %w", err)
	}

	return nil
}

// SemanticValidationContext defines the context under which a semantic validation for a Transaction is happening.
type SemanticValidationContext struct {
	ExtParas *ExternalUnlockParameters

	// The working set which is auto. populated during the semantic validation.
	WorkingSet *SemValiContextWorkingSet
}

// SemValiContextWorkingSet contains fields which get automatically populated
// by the library during the semantic validation of a Transaction.
type SemValiContextWorkingSet struct {
	// The identities which are successfully unlocked from the input side.
	UnlockedIdents UnlockedIdentities
	// The mapping of OutputID to the actual Outputs.
	InputSet OutputSet
	// The mapping of inputs' OutputID to the index.
	InputIDToIndex map[OutputID]uint16
	// The transaction for which this semantic validation happens.
	Tx *Transaction
	// The message which signatures are signing.
	EssenceMsgToSign []byte
	// The inputs of the transaction mapped by type.
	InputsByType OutputsByType
	// The ChainConstrainedOutput(s) at the input side.
	InChains ChainConstrainedOutputsSet
	// The sum of NativeTokens at the input side.
	InNativeTokens NativeTokenSum
	// The Outputs of the transaction mapped by type.
	OutputsByType OutputsByType
	// The ChainConstrainedOutput(s) at the output side.
	OutChains ChainConstrainedOutputsSet
	// The sum of NativeTokens at the output side.
	OutNativeTokens NativeTokenSum
	// The UnlockBlocks carried by the transaction mapped by type.
	UnlockBlocksByType UnlockBlocksByType
}

// UTXOInputAtIndex retrieves the UTXOInput at the given index.
// Caller must ensure that the index is valid.
func (workingSet *SemValiContextWorkingSet) UTXOInputAtIndex(inputIndex uint16) *UTXOInput {
	return workingSet.Tx.Essence.Inputs[inputIndex].(*UTXOInput)
}

func NewSemValiContextWorkingSet(t *Transaction, inputs OutputSet) (*SemValiContextWorkingSet, error) {
	var err error
	workingSet := &SemValiContextWorkingSet{}
	workingSet.Tx = t
	workingSet.UnlockedIdents = make(UnlockedIdentities)
	workingSet.InputSet = inputs
	workingSet.InputIDToIndex = func() map[OutputID]uint16 {
		m := make(map[OutputID]uint16)
		for inputIndex, inputRef := range workingSet.Tx.Essence.Inputs {
			ref := inputRef.(IndexedUTXOReferencer).Ref()
			m[ref] = uint16(inputIndex)
		}
		return m
	}()
	workingSet.EssenceMsgToSign, err = t.Essence.SigningMessage()
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

	txID, err := workingSet.Tx.ID()
	if err != nil {
		return nil, err
	}

	workingSet.InChains = workingSet.InputSet.ChainConstrainedOutputSet()
	workingSet.OutputsByType = t.Essence.Outputs.ToOutputsByType()
	workingSet.OutChains = workingSet.Tx.Essence.Outputs.ChainConstrainedOutputSet(*txID)

	workingSet.UnlockBlocksByType = t.UnlockBlocks.ToUnlockBlocksByType()
	return workingSet, nil
}

// SemanticallyValidate semantically validates the Transaction by checking that the semantic rules applied to the inputs
// and outputs are fulfilled. Semantic validation must only be executed on Transaction(s) which are syntactically valid.
func (t *Transaction) SemanticallyValidate(svCtx *SemanticValidationContext, inputs OutputSet, semValFuncs ...TxSemanticValidationFunc) error {
	var err error
	svCtx.WorkingSet, err = NewSemValiContextWorkingSet(t, inputs)
	if err != nil {
		return err
	}

	if len(semValFuncs) > 0 {
		if err := runSemanticValidations(svCtx, semValFuncs...); err != nil {
			return err
		}
		return nil
	}

	// do not change the order of these functions as
	// some of them might depend on certain mutations
	// on the given SemanticValidationContext
	if err := runSemanticValidations(svCtx,
		TxSemanticTimelock(),
		TxSemanticInputUnlocks(),
		TxSemanticOutputsSender(),
		TxSemanticDeposit(),
		TxSemanticNativeTokens(),
		TxSemanticSTVFOnChains()); err != nil {
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
// The value represent the index of the unlock block which unlocked the identity.
type UnlockedIdentities map[string]UnlockedIndices

// AddInputUnlockedBy adds an entry defining that the given identity unlocked an input at the given index.
func (unlockedIdents UnlockedIdentities) AddInputUnlockedBy(identKey string, inputIndex uint16) {
	unlockedIndices, ok := unlockedIdents[identKey]
	if !ok {
		unlockedIndices = make(UnlockedIndices)
	}
	unlockedIndices[inputIndex] = struct{}{}
	unlockedIdents[identKey] = unlockedIndices
}

// UnlockedIndices is a set of unlocked indices of outputs an identity unlocked.
type UnlockedIndices map[uint16]struct{}

// Unlocked tells whether x is in this set.
func (indices UnlockedIndices) Unlocked(x uint16) bool {
	_, has := indices[x]
	return has
}

// TxSemanticValidationFunc is a function which given the context, input, outputs and
// unlock blocks runs a specific semantic validation. The function might also modify the SemanticValidationContext
// in order to supply information to subsequent TxSemanticValidationFunc(s).
type TxSemanticValidationFunc func(svCtx *SemanticValidationContext) error

// TxSemanticInputUnlocks produces the UnlockedIdentities which will be set into the given SemanticValidationContext
// and verifies that inputs are correctly unlocked and that the inputs commitment matches.
func TxSemanticInputUnlocks() TxSemanticValidationFunc {
	return func(svCtx *SemanticValidationContext) error {
		var inputs Outputs

		// it is important that the inputs are checked in order as referential unlocks
		// check against previous unlocks
		for inputIndex, inputRef := range svCtx.WorkingSet.Tx.Essence.Inputs {
			input, ok := svCtx.WorkingSet.InputSet[inputRef.(IndexedUTXOReferencer).Ref()]
			if !ok {
				return fmt.Errorf("%w: utxo for input %d not supplied", ErrMissingUTXO, inputIndex)
			}
			inputs = append(inputs, input)
		}

		actualInputCommitment, err := inputs.Commitment()
		if err != nil {
			return fmt.Errorf("unable to compute hash of inputs: %w", err)
		}

		expectedInputCommitment := svCtx.WorkingSet.Tx.Essence.InputsCommitment[:]
		if !bytes.Equal(expectedInputCommitment, actualInputCommitment) {
			return fmt.Errorf("%w: specified %v but got %v", ErrInvalidInputsCommitment, expectedInputCommitment, actualInputCommitment)
		}

		for inputIndex, input := range inputs {
			if err := unlockOutput(svCtx, input, uint16(inputIndex)); err != nil {
				return err
			}

			// since this input is now unlocked, and it is a ChainConstrainedOutput, the chain's address becomes automatically unlocked
			if chainConstrOutput, is := input.(ChainConstrainedOutput); is && chainConstrOutput.Chain().Addressable() {
				chainID := chainConstrOutput.Chain()
				if chainID.Empty() {
					chainID = chainID.(UTXOIDChainID).FromOutputID(svCtx.WorkingSet.UTXOInputAtIndex(uint16(inputIndex)).Ref())
				}
				svCtx.WorkingSet.UnlockedIdents.AddInputUnlockedBy(chainID.ToAddress().Key(), uint16(inputIndex))
			}
		}

		return nil
	}
}

func identToUnlock(svCtx *SemanticValidationContext, input Output, inputIndex uint16) (Address, error) {
	switch in := input.(type) {

	case TransIndepIdentOutput:
		return in.Ident(), nil

	case TransDepIdentOutput:
		chainID := in.Chain()
		if chainID.Empty() {
			utxoChainID, is := chainID.(UTXOIDChainID)
			if !is {
				return nil, ErrTransDepIdentOutputNonUTXOChainID
			}
			chainID = utxoChainID.FromOutputID(svCtx.WorkingSet.Tx.Essence.Inputs[inputIndex].(IndexedUTXOReferencer).Ref())
		}

		next := svCtx.WorkingSet.OutChains[chainID]
		if next == nil {
			return in.Ident(nil)
		}

		nextTransDepIdentOutput, ok := next.(TransDepIdentOutput)
		if !ok {
			return nil, ErrTransDepIdentOutputNextInvalid
		}

		return in.Ident(nextTransDepIdentOutput)

	default:
		panic("unknown ident output type in semantic unlocks")
	}
}

func checkExpiredForReceiver(svCtx *SemanticValidationContext, output Output) Address {
	unlockConditionOutput, is := output.(UnlockConditionOutput)
	if !is {
		return nil
	}

	unlockCondSet := unlockConditionOutput.UnlockConditions().MustSet()
	if ok, returnIdent := unlockCondSet.returnIdentCanUnlock(svCtx.ExtParas); ok {
		return returnIdent
	}

	return nil
}

func unlockOutput(svCtx *SemanticValidationContext, output Output, inputIndex uint16) error {
	ownerIdent, err := identToUnlock(svCtx, output, inputIndex)
	if err != nil {
		return fmt.Errorf("unable to retrieve ident to unlock of input %d: %w", inputIndex, err)
	}

	if actualIdentToUnlock := checkExpiredForReceiver(svCtx, output); actualIdentToUnlock != nil {
		ownerIdent = actualIdentToUnlock
	}

	unlockBlock := svCtx.WorkingSet.Tx.UnlockBlocks[inputIndex]

	switch owner := ownerIdent.(type) {
	case ChainConstrainedAddress:
		referentialUnlockBlock, isReferentialUnlockBlock := unlockBlock.(ReferentialUnlockBlock)
		if !isReferentialUnlockBlock || !referentialUnlockBlock.Chainable() || !referentialUnlockBlock.SourceAllowed(ownerIdent) {
			return fmt.Errorf("%w: input %d has a chain constrained address (%T) but its corresponding unlock block is of type %T", ErrInvalidInputUnlock, inputIndex, owner, unlockBlock)
		}

		unlockedIndices, wasUnlocked := svCtx.WorkingSet.UnlockedIdents[owner.Key()]
		if !wasUnlocked || !unlockedIndices.Unlocked(referentialUnlockBlock.Ref()) {
			return fmt.Errorf("%w: input %d's chain constrained address (%T) is not unlocked through input %d's unlock block", ErrInvalidInputUnlock, inputIndex, owner, referentialUnlockBlock.Ref())
		}

	case DirectUnlockableAddress:
		switch uBlock := unlockBlock.(type) {
		case ReferentialUnlockBlock:
			if uBlock.Chainable() || !uBlock.SourceAllowed(ownerIdent) {
				return fmt.Errorf("%w: input %d has none chain constrained address of %s but its corresponding unlock block is of type %s", ErrInvalidInputUnlock, inputIndex, owner.Type(), unlockBlock.Type())
			}

			unlockedIndices, wasUnlocked := svCtx.WorkingSet.UnlockedIdents[owner.Key()]
			if !wasUnlocked || !unlockedIndices.Unlocked(uBlock.Ref()) {
				return fmt.Errorf("%w: input %d's address is not unlocked through input %d's unlock block", ErrInvalidInputUnlock, inputIndex, uBlock.Ref())
			}

		case *SignatureUnlockBlock:
			// owner must not be unlocked already
			if unlockedAtIndex, wasAlreadyUnlocked := svCtx.WorkingSet.UnlockedIdents[owner.Key()]; wasAlreadyUnlocked {
				return fmt.Errorf("%w: input %d's address is already unlocked through input %d's unlock block but the input uses a non referential unlock block", ErrInvalidInputUnlock, inputIndex, unlockedAtIndex)
			}

			if err := owner.Unlock(svCtx.WorkingSet.EssenceMsgToSign, uBlock.Signature); err != nil {
				return fmt.Errorf("%w: input %d's address is not unlocked through its signature unlock block", err, inputIndex)
			}
		}
	default:
		panic("unknown address in semantic unlocks")
	}

	svCtx.WorkingSet.UnlockedIdents.AddInputUnlockedBy(ownerIdent.Key(), inputIndex)

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

			senderFeatureBlock := featureBlockOutput.FeatureBlocks().MustSet().SenderFeatureBlock()
			if senderFeatureBlock == nil {
				continue
			}

			// check unlocked
			sender := senderFeatureBlock.Address
			if _, isUnlocked := svCtx.WorkingSet.UnlockedIdents[sender.Key()]; !isUnlocked {
				return fmt.Errorf("%w: output %d", ErrSenderFeatureBlockNotUnlocked, outputIndex)
			}
		}
		return nil
	}
}

// TxSemanticDeposit validates that the IOTA tokens are balanced from the input/output side.
// It additionally also incorporates the check whether return amounts via DustDepositReturnUnlockCondition(s) for specified identities
// are fulfilled from the output side.
func TxSemanticDeposit() TxSemanticValidationFunc {
	return func(svCtx *SemanticValidationContext) error {
		// note that due to syntactic validation of outputs, input and output deposit sums
		// are always within bounds of the total token supply
		var in, out uint64
		inputSumReturnAmountPerIdent := make(map[string]uint64)
		for inputID, input := range svCtx.WorkingSet.InputSet {
			in += input.Deposit()
			unlockConditionOutput, is := input.(UnlockConditionOutput)
			if !is {
				continue
			}

			unlockCondSet := unlockConditionOutput.UnlockConditions().MustSet()
			dustDepositReturnUnlockCondition := unlockCondSet.DustDepositReturn()
			if dustDepositReturnUnlockCondition == nil {
				continue
			}

			if !unlockCondSet.HasExpirationCondition() {
				continue
			}

			dustDepositReturnIdentKey := dustDepositReturnUnlockCondition.ReturnAddress.Key()

			// if the return ident unlocked this input, then the return amount does
			// not have to be fulfilled
			unlockedIndices, has := svCtx.WorkingSet.UnlockedIdents[dustDepositReturnIdentKey]
			if has && unlockedIndices.Unlocked(svCtx.WorkingSet.InputIDToIndex[inputID]) {
				continue
			}

			inputSumReturnAmountPerIdent[dustDepositReturnIdentKey] += dustDepositReturnUnlockCondition.Amount
		}

		outputSimpleTransfersPerIdent := make(map[string]uint64)
		for _, output := range svCtx.WorkingSet.Tx.Essence.Outputs {
			outDeposit := output.Deposit()
			out += outDeposit

			// accumulate simple transfers for DustDepositReturnUnlockCondition checks
			if basicOutput, is := output.(*BasicOutput); is {
				if len(basicOutput.FeatureBlocks()) > 0 {
					continue
				}
				outputSimpleTransfersPerIdent[basicOutput.Ident().Key()] += outDeposit
			}
		}

		if in != out {
			return fmt.Errorf("%w: in %d, out %d", ErrInputOutputSumMismatch, in, out)
		}

		for ident, returnSum := range inputSumReturnAmountPerIdent {
			outSum, has := outputSimpleTransfersPerIdent[ident]
			if !has {
				return fmt.Errorf("%w: return amount of %d not fulfilled as there is no output for %s", ErrReturnAmountNotFulFilled, returnSum, ident)
			}
			if outSum < returnSum {
				return fmt.Errorf("%w: return amount of %d not fulfilled as output is only %d for %s", ErrReturnAmountNotFulFilled, returnSum, outSum, ident)
			}
		}

		return nil
	}
}

// TxSemanticTimelock validates that the inputs' timelocks are expired.
func TxSemanticTimelock() TxSemanticValidationFunc {
	return func(svCtx *SemanticValidationContext) error {
		for inputIndex, input := range svCtx.WorkingSet.InputSet {
			unlockConditionOutput, is := input.(UnlockConditionOutput)
			if !is {
				continue
			}

			if err := unlockConditionOutput.UnlockConditions().MustSet().TimelocksExpired(svCtx.ExtParas); err != nil {
				return fmt.Errorf("%w: input at index %d's timelocks are not expired", err, inputIndex)
			}
		}
		return nil
	}
}

// TxSemanticSTVFOnChains executes StateTransitionValidationFunc(s) on ChainConstrainedOutput(s).
func TxSemanticSTVFOnChains() TxSemanticValidationFunc {
	return func(svCtx *SemanticValidationContext) error {
		for chainID, inputChain := range svCtx.WorkingSet.InChains {
			nextState := svCtx.WorkingSet.OutChains[chainID]
			if nextState == nil {
				if err := inputChain.ValidateStateTransition(ChainTransitionTypeDestroy, nil, svCtx); err != nil {
					return fmt.Errorf("input chain %s (%T) destruction transition failed: %w", chainID, inputChain, err)
				}
				continue
			}
			if err := inputChain.ValidateStateTransition(ChainTransitionTypeStateChange, nextState, svCtx); err != nil {
				return fmt.Errorf("chain %s (%T) state transition failed: %w", chainID, inputChain, err)
			}
		}

		for chainID, outputChain := range svCtx.WorkingSet.OutChains {
			previousState := svCtx.WorkingSet.InChains[chainID]
			if previousState == nil {
				if err := outputChain.ValidateStateTransition(ChainTransitionTypeGenesis, nil, svCtx); err != nil {
					return fmt.Errorf("new chain %s (%T) state transition failed: %w", chainID, outputChain, err)
				}
			}
		}

		return nil
	}
}

// TxSemanticNativeTokens validates following rules regarding NativeTokens:
//	- The NativeTokens between Inputs / Outputs must be balanced in terms of circulating supply adjustments if
//	  there is no foundry state transition for a given NativeToken.
// 	- Max MaxNativeTokensCount native tokens within inputs + outputs
func TxSemanticNativeTokens() TxSemanticValidationFunc {
	return func(svCtx *SemanticValidationContext) error {
		// native token set creates handle overflows
		var err error
		var inNTCount, outNTCount int
		svCtx.WorkingSet.InNativeTokens, inNTCount, err = svCtx.WorkingSet.InputsByType.NativeTokenOutputs().Sum()
		if err != nil {
			return fmt.Errorf("invalid input native token set: %w", err)
		}

		if inNTCount > MaxNativeTokensCount {
			return fmt.Errorf("%w: inputs native token count %d exceeds max of %d", ErrMaxNativeTokensCountExceeded, inNTCount, MaxNativeTokensCount)
		}

		svCtx.WorkingSet.OutNativeTokens, outNTCount, err = svCtx.WorkingSet.OutputsByType.NativeTokenOutputs().Sum()
		if err != nil {
			return fmt.Errorf("invalid output native token set: %w", err)
		}

		if inNTCount+outNTCount > MaxNativeTokensCount {
			return fmt.Errorf("%w: native token count (in %d + out %d) exceeds max of %d", ErrMaxNativeTokensCountExceeded, inNTCount, outNTCount, MaxNativeTokensCount)
		}

		// easy route, tokens must be balanced between both sets
		if svCtx.WorkingSet.OutputsByType[OutputFoundry] == nil && svCtx.WorkingSet.InputsByType[OutputFoundry] == nil {
			if err := svCtx.WorkingSet.InNativeTokens.Balanced(svCtx.WorkingSet.OutNativeTokens); err != nil {
				return err
			}
			return nil
		}

		// check for the input and output side whether we have the transitioning foundry
		// in case either side is missing its companion sum or the tokens are unbalanced by
		// just looking at both sides' sums

		for nativeTokenID, inSum := range svCtx.WorkingSet.InNativeTokens {
			outSum := svCtx.WorkingSet.OutNativeTokens[nativeTokenID]
			_, foundryIsTransitioning := svCtx.WorkingSet.OutChains[nativeTokenID.FoundryID()]

			if foundryIsTransitioning {
				continue
			}

			switch {
			case outSum == nil:
				return fmt.Errorf("%w: native token %d exists on input but not output side and the foundry is not transitioning", ErrNativeTokenSumUnbalanced, nativeTokenID)
			case inSum.Cmp(outSum) != 0:
				return fmt.Errorf("%w: native token %d is unbalanced between input (%d) and output (%d) but the foundry is not transitioning", ErrNativeTokenSumUnbalanced, inSum, outSum, nativeTokenID)
			}
		}

		for nativeTokenID := range svCtx.WorkingSet.OutNativeTokens {
			inSum := svCtx.WorkingSet.InNativeTokens[nativeTokenID]
			_, foundryIsTransitioning := svCtx.WorkingSet.OutChains[nativeTokenID.FoundryID()]

			// just need to check whether the foundry is transitioning, since the balancing
			// between in and out is already given from the previous check
			if inSum == nil && !foundryIsTransitioning {
				return fmt.Errorf("%w: native token %s is new on the output side but the foundry is not transitioning", ErrNativeTokenSumUnbalanced, nativeTokenID)
			}
		}

		// from here the native tokens balancing is handled by the foundry's STVF

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
