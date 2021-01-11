package iota

import (
	"encoding/json"
	"errors"
	"fmt"

	"golang.org/x/crypto/blake2b"
)

const (
	// Defines the transaction payload's type ID.
	TransactionPayloadTypeID uint32 = 0

	// Defines the length of a Transaction ID.
	TransactionIDLength = blake2b.Size256

	// Defines the minimum size of a serialized Transaction.
	TransactionBinSerializedMinSize = UInt32ByteSize

	// Defines the divisor used to compute the allowed dust outputs on an address.
	// The amount of dust outputs on an address is calculated by:
	//	sum(dust_allowance_output_deposit) / DustAllowanceDivisor
	// Example: 1_000_000 / 10_000 = 100 dust outputs
	DustAllowanceDivisor int64 = 10_000
)

var (
	// Returned if the count of unlock blocks doesn't match the count of inputs.
	ErrUnlockBlocksMustMatchInputCount = errors.New("the count of unlock blocks must match the inputs of the transaction")
	// Returned if the transaction essence within a Transaction is invalid.
	ErrInvalidTransactionEssence = errors.New("transaction essence is invalid")
	// Returned if an UTXO is missing to commence a certain operation.
	ErrMissingUTXO = errors.New("missing utxo")
	// Returned if a transaction does not spend the entirety of the inputs to the outputs.
	ErrInputOutputSumMismatch = errors.New("inputs and outputs do not spend/deposit the same amount")
	// Returned if an address of an input has a companion signature unlock block with the wrong signature type.
	ErrSignatureAndAddrIncompatible = errors.New("address and signature type are not compatible")
	// Returned for errors where the dust allowance is semantically invalid.
	ErrInvalidDustAllowance = errors.New("invalid dust allowance")
)

// TransactionID is the ID of a Transaction.
type TransactionID = [TransactionIDLength]byte

// TransactionIDs are IDs of transactions.
type TransactionIDs []TransactionID

// Transaction is a transaction with its inputs, outputs and unlock blocks.
type Transaction struct {
	// The transaction essence, respectively the transfer part of a Transaction.
	Essence Serializable
	// The unlock blocks defining the unlocking data for the inputs within the Essence.
	UnlockBlocks Serializables
}

// ID computes the ID of the Transaction.
func (t *Transaction) ID() (*TransactionID, error) {
	data, err := t.Serialize(DeSeriModeNoValidation)
	if err != nil {
		return nil, fmt.Errorf("can't compute transaction ID: %w", err)
	}
	h := blake2b.Sum256(data)
	return &h, nil
}

func (t *Transaction) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	unlockBlockArrayRules := &ArrayRules{
		MinErr: ErrUnlockBlocksMustMatchInputCount,
		MaxErr: ErrUnlockBlocksMustMatchInputCount,
	}

	return NewDeserializer(data).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(DeSeriModePerformValidation) {
				if err := checkMinByteLength(TransactionBinSerializedMinSize, len(data)); err != nil {
					return fmt.Errorf("invalid transaction bytes: %w", err)
				}
				if err := checkType(data, TransactionPayloadTypeID); err != nil {
					return fmt.Errorf("unable to deserialize transaction: %w", err)
				}
			}
			return nil
		}).
		Skip(TypeDenotationByteSize, func(err error) error {
			return fmt.Errorf("unable to skip transaction payload ID during deserialization: %w", err)
		}).
		ReadObject(func(seri Serializable) { t.Essence = seri }, deSeriMode, TypeDenotationByte, TransactionEssenceSelector, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize transaction essence within transaction", err)
		}).
		Do(func() {
			// TODO: tx must be a TransactionEssence but might be something else in the future
			inputCount := uint16(len(t.Essence.(*TransactionEssence).Inputs))
			unlockBlockArrayRules.Min = inputCount
			unlockBlockArrayRules.Max = inputCount
		}).
		ReadSliceOfObjects(func(seri Serializables) { t.UnlockBlocks = seri }, deSeriMode, TypeDenotationByte, UnlockBlockSelector, unlockBlockArrayRules, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize unlock blocks", err)
		}).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(DeSeriModePerformValidation) {
				return t.SyntacticallyValidate()
			}
			return nil
		}).
		Done()
}

func (t *Transaction) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	return NewSerializer().
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(DeSeriModePerformValidation) {
				return t.SyntacticallyValidate()
			}
			return nil
		}).
		WriteNum(TransactionPayloadTypeID, func(err error) error {
			return fmt.Errorf("%w: unable to serialize transaction payload ID", err)
		}).
		WriteObject(t.Essence, deSeriMode, func(err error) error {
			return fmt.Errorf("%w: unable to serialize transaction's essence", err)
		}).
		WriteSliceOfObjects(t.UnlockBlocks, deSeriMode, nil, func(err error) error {
			return fmt.Errorf("%w: unable to serialize transaction's unlock blocks", err)
		}).
		Serialize()
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	jsonSigTxPayload := &jsontransaction{
		UnlockBlocks: make([]*json.RawMessage, len(t.UnlockBlocks)),
	}
	jsonSigTxPayload.Type = int(TransactionPayloadTypeID)
	txJson, err := t.Essence.MarshalJSON()
	if err != nil {
		return nil, err
	}
	rawMsgTxJson := json.RawMessage(txJson)
	jsonSigTxPayload.Essence = &rawMsgTxJson
	for i, ub := range t.UnlockBlocks {
		jsonUB, err := ub.MarshalJSON()
		if err != nil {
			return nil, err
		}
		rawMsgJsonUB := json.RawMessage(jsonUB)
		jsonSigTxPayload.UnlockBlocks[i] = &rawMsgJsonUB
	}
	return json.Marshal(jsonSigTxPayload)
}

func (t *Transaction) UnmarshalJSON(bytes []byte) error {
	jsonSigTxPayload := &jsontransaction{}
	if err := json.Unmarshal(bytes, jsonSigTxPayload); err != nil {
		return err
	}
	seri, err := jsonSigTxPayload.ToSerializable()
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

	if t.Essence == nil {
		return fmt.Errorf("%w: transaction is nil", ErrInvalidTransactionEssence)
	}

	if t.UnlockBlocks == nil {
		return fmt.Errorf("%w: unlock blocks are nil", ErrInvalidTransactionEssence)
	}

	txEssence, ok := t.Essence.(*TransactionEssence)
	if !ok {
		return fmt.Errorf("%w: transaction essence is not *TransactionEssence", ErrInvalidTransactionEssence)
	}

	if err := txEssence.SyntacticallyValidate(); err != nil {
		return fmt.Errorf("%w: transaction essence part is invalid", err)
	}

	inputCount := len(txEssence.Inputs)
	unlockBlockCount := len(t.UnlockBlocks)
	if inputCount != unlockBlockCount {
		return fmt.Errorf("%w: num of inputs %d, num of unlock blocks %d", ErrUnlockBlocksMustMatchInputCount, inputCount, unlockBlockCount)
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
type SemanticValidationFunc = func(t *Transaction, utxos InputToOutputMapping) error

// DustAllowanceFunc returns the deposit sum of dust allowance outputs and amount of dust outputs on the given address.
type DustAllowanceFunc func(addr Serializable) (dustAllowanceSum uint64, amountDustOutputs int64, err error)

// NewDustSemanticValidation returns a SemanticValidationFunc which verifies whether
// a transaction fulfils the semantics regarding dust outputs:
//	A transaction:
//		- consuming a SigLockedDustAllowanceOutput on address A or
//		- creating a SigLockedSingleOutput with deposit amount < OutputSigLockedDustAllowanceOutputMinDeposit (dust output)
//	is only semantically valid, if after the transaction is booked, the number of dust outputs on address A does not exceed the allowed
//	threshold of the sum of S / div. Where S is the sum of deposits of all dust allowance outputs on address A.
func NewDustSemanticValidation(div int64, dustAllowanceFunc DustAllowanceFunc) SemanticValidationFunc {
	return func(t *Transaction, utxos InputToOutputMapping) error {
		essence := t.Essence.(*TransactionEssence)

		dustAllowanceAddrToBalance := make(map[Serializable]int64)
		dustAllowanceAddrToNumOfDustOutputs := make(map[Serializable]int64)

		for _, output := range essence.Outputs {
			switch out := output.(type) {
			case *SigLockedDustAllowanceOutput:
				dustAllowanceAddrToBalance[out.Address] += int64(out.Amount)
			case *SigLockedSingleOutput:
				if out.Amount < OutputSigLockedDustAllowanceOutputMinDeposit {
					dustAllowanceAddrToNumOfDustOutputs[out.Address] += 1
				}
			}
		}

		for i, x := range t.Essence.(*TransactionEssence).Inputs {
			utxoID := x.(*UTXOInput).ID()
			utxo, ok := utxos[utxoID]
			if !ok {
				return fmt.Errorf("%w: UTXO for ID %v is not provided (input at index %d)", ErrMissingUTXO, utxoID, i)
			}

			deposit, err := utxo.Deposit()
			if err != nil {
				return fmt.Errorf("unable to get deposit from UTXO %v (input at index %d): %w", utxoID, i, err)
			}

			target, err := utxo.Target()
			if err != nil {
				return fmt.Errorf("unable to get target of UTXO %v (input at index %d): %w", utxoID, i, err)
			}

			if deposit < OutputSigLockedDustAllowanceOutputMinDeposit {
				dustAllowanceAddrToNumOfDustOutputs[target] -= 1
				continue
			}

			if utxo.Type() == OutputSigLockedDustAllowanceOutput {
				dustAllowanceAddrToBalance[target] -= int64(deposit)
			}
		}

		addrToValidate := make(map[Serializable]struct{})
		for addr := range dustAllowanceAddrToBalance {
			addrToValidate[addr] = struct{}{}
		}
		for addr := range dustAllowanceAddrToNumOfDustOutputs {
			addrToValidate[addr] = struct{}{}
		}

		for addr := range addrToValidate {
			dustAllowanceDepositSumUint64, numDustOutputs, err := dustAllowanceFunc(addr)
			if err != nil {
				return fmt.Errorf("unable to fetch dust allowance information on address %v: %w", addr, err)
			}
			numDustOutputsPrev := numDustOutputs
			numDustOutputs += dustAllowanceAddrToNumOfDustOutputs[addr]

			var dustAllowanceDepositSum = int64(dustAllowanceDepositSumUint64)
			// Go integer division floors the value
			prevAllowed := dustAllowanceDepositSum / div
			allowed := (dustAllowanceDepositSum + dustAllowanceAddrToBalance[addr]) / div

			if numDustOutputs > allowed {
				addrHex := addr.(Address).String()
				short := numDustOutputs - allowed
				return fmt.Errorf("%w: addr %s, new num of dust outputs %d (previous %d), allowance deposit %d (previous %d), short %d", ErrInvalidDustAllowance, addrHex, numDustOutputs, numDustOutputsPrev, allowed, prevAllowed, short)
			}
		}

		return nil
	}
}

// InputToOutputMapping maps inputs to their origin UTXOs.
type InputToOutputMapping = map[UTXOInputID]Output

// SemanticallyValidate semantically validates the Transaction
// by checking that the given input UTXOs are spent entirely and the signatures
// provided are valid. SyntacticallyValidate() should be called before SemanticallyValidate() to
// ensure that the essence part of the transaction is syntactically valid.
func (t *Transaction) SemanticallyValidate(utxos InputToOutputMapping, semValFuncs ...SemanticValidationFunc) error {

	txEssence, ok := t.Essence.(*TransactionEssence)
	if !ok {
		return fmt.Errorf("%w: transaction is not *TransactionEssence", ErrInvalidTransactionEssence)
	}

	txEssenceBytes, err := txEssence.Serialize(DeSeriModeNoValidation)
	if err != nil {
		return err
	}

	inputSum, sigValidFuncs, err := t.SemanticallyValidateInputs(utxos, txEssence, txEssenceBytes)
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
		if err := semValFunc(t, utxos); err != nil {
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

// SemanticallyValidateInputs checks that every referenced UTXO is available, computes the input sum
// and returns functions which can be called to verify the signatures.
// This function should only be called from SemanticallyValidate().
func (t *Transaction) SemanticallyValidateInputs(utxos InputToOutputMapping, transaction *TransactionEssence, txEssenceBytes []byte) (uint64, []SigValidationFunc, error) {
	var sigValidFuncs []SigValidationFunc
	var inputSum uint64

	for i, input := range transaction.Inputs {
		// TODO: switch out with type switch or interface once more types are known
		in, ok := input.(*UTXOInput)
		if !ok {
			return 0, nil, fmt.Errorf("%w: unsupported input type at index %d", ErrUnknownInputType, i)
		}

		// check that we got the needed UTXO
		utxoID := in.ID()
		utxo, ok := utxos[utxoID]
		if !ok {
			return 0, nil, fmt.Errorf("%w: UTXO for ID %v is not provided (input at index %d)", ErrMissingUTXO, utxoID, i)
		}

		var err error
		deposit, err := utxo.Deposit()
		if err != nil {
			return 0, nil, fmt.Errorf("unable to get deposit from UTXO %v (input at index %d): %w", utxoID, i, err)
		}
		inputSum += deposit

		sigBlock, sigBlockIndex, err := t.signatureUnlockBlock(i)
		if err != nil {
			return 0, nil, err
		}

		// TODO: optimization: currently the same signature is validated "inputs-with-same-addr"-times instead of only once
		sigValidF, err := createSigValidationFunc(i, sigBlock.Signature, sigBlockIndex, txEssenceBytes, utxo)
		if err != nil {
			return 0, nil, err
		}

		sigValidFuncs = append(sigValidFuncs, sigValidF)
	}

	return inputSum, sigValidFuncs, nil
}

// retrieves the SignatureUnlockBlock at the given index or follows
// the reference of an ReferenceUnlockBlock to retrieve it.
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
func createSigValidationFunc(pos int, sig Serializable, sigBlockIndex int, txEssenceBytes []byte, utxo Output) (SigValidationFunc, error) {
	addr, err := utxo.Target()
	if err != nil {
		return nil, fmt.Errorf("unable to get target for UTXO: %w", err)
	}
	switch addr := addr.(type) {
	case *WOTSAddress:
		// TODO: implement
		return nil, fmt.Errorf("%w: unsupported address type at index %d", ErrWOTSNotImplemented, pos)
	case *Ed25519Address:
		return createEd25519SigValidationFunc(pos, sig, sigBlockIndex, addr, txEssenceBytes)
	default:
		return nil, fmt.Errorf("%w: unsupported address type at index %d", ErrUnknownAddrType, pos)
	}
}

// creates a SigValidationFunc validating the given Ed25519Signature against the Ed25519Address.
func createEd25519SigValidationFunc(pos int, sig Serializable, sigBlockIndex int, addr *Ed25519Address, essenceBytes []byte) (SigValidationFunc, error) {
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

// SemanticallyValidateOutputs accumulates the sum of all outputs.
// This function should only be called from SemanticallyValidate().
func (t *Transaction) SemanticallyValidateOutputs(transaction *TransactionEssence) (uint64, error) {
	var outputSum uint64
	for i, output := range transaction.Outputs {
		// TODO: switch out with type switch
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

// jsontransaction defines the json representation of a Transaction.
type jsontransaction struct {
	Type         int                `json:"type"`
	Essence      *json.RawMessage   `json:"essence"`
	UnlockBlocks []*json.RawMessage `json:"unlockBlocks"`
}

func (jsontx *jsontransaction) ToSerializable() (Serializable, error) {
	jsonTxEssence, err := DeserializeObjectFromJSON(jsontx.Essence, jsontransactionessenceselector)
	if err != nil {
		return nil, fmt.Errorf("unable to decode transaction essence from JSON: %w", err)
	}

	txEssenceSeri, err := jsonTxEssence.ToSerializable()
	if err != nil {
		return nil, err
	}

	unlockBlocks := make(Serializables, len(jsontx.UnlockBlocks))
	for i, ele := range jsontx.UnlockBlocks {
		jsonUnlockBlock, err := DeserializeObjectFromJSON(ele, jsonunlockblockselector)
		if err != nil {
			return nil, fmt.Errorf("unable to decode unlock block type from JSON, pos %d: %w", i, err)
		}
		unlockBlock, err := jsonUnlockBlock.ToSerializable()
		if err != nil {
			return nil, fmt.Errorf("pos %d: %w", i, err)
		}
		unlockBlocks[i] = unlockBlock
	}

	return &Transaction{Essence: txEssenceSeri, UnlockBlocks: unlockBlocks}, nil
}
