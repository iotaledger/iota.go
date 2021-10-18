package iotago

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer"

	"golang.org/x/crypto/blake2b"
)

const (
	// TransactionPayloadTypeID defines the transaction payload's type ID.
	TransactionPayloadTypeID uint32 = 0

	// TransactionIDLength defines the length of a Transaction ID.
	TransactionIDLength = blake2b.Size256

	// TransactionBinSerializedMinSize defines the minimum size of a serialized Transaction.
	TransactionBinSerializedMinSize = serializer.UInt32ByteSize

	// DustAllowanceDivisor defines the divisor used to compute the allowed dust outputs on an address.
	// The amount of dust outputs on an address is calculated by:
	//	min(sum(dust_allowance_output_deposit) / DustAllowanceDivisor, dustOutputCountLimit)
	DustAllowanceDivisor int64 = 100_000
	// MaxDustOutputsOnAddress defines the maximum amount of dust outputs allowed to "reside" on an address.
	MaxDustOutputsOnAddress = 100
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
	// ErrInvalidDustAllowance gets returned for errors where the dust allowance is semantically invalid.
	ErrInvalidDustAllowance = errors.New("invalid dust allowance")
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
	UnlockBlocks serializer.Serializables
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
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := serializer.CheckMinByteLength(TransactionBinSerializedMinSize, len(data)); err != nil {
					return fmt.Errorf("invalid transaction bytes: %w", err)
				}
				if err := serializer.CheckType(data, TransactionPayloadTypeID); err != nil {
					return fmt.Errorf("unable to deserialize transaction: %w", err)
				}
			}
			return nil
		}).
		Skip(serializer.TypeDenotationByteSize, func(err error) error {
			return fmt.Errorf("unable to skip transaction payload ID during deserialization: %w", err)
		}).
		ReadObject(func(seri serializer.Serializable) { t.Essence = seri }, deSeriMode, serializer.TypeDenotationByte, TransactionEssenceSelector, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize transaction essence within transaction", err)
		}).
		Do(func() {
			inputCount := uint(len(t.Essence.(*TransactionEssence).Inputs))
			unlockBlockArrayRules.Min = inputCount
			unlockBlockArrayRules.Max = inputCount
		}).
		ReadSliceOfObjects(func(seri serializer.Serializables) { t.UnlockBlocks = seri }, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationByte, UnlockBlockSelector, unlockBlockArrayRules, func(err error) error {
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
		WriteNum(TransactionPayloadTypeID, func(err error) error {
			return fmt.Errorf("%w: unable to serialize transaction payload ID", err)
		}).
		WriteObject(t.Essence, deSeriMode, func(err error) error {
			return fmt.Errorf("%w: unable to serialize transaction's essence", err)
		}).
		WriteSliceOfObjects(t.UnlockBlocks, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, nil, func(err error) error {
			return fmt.Errorf("%w: unable to serialize transaction's unlock blocks", err)
		}).
		Serialize()
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	jTransaction := &jsonTransaction{
		UnlockBlocks: make([]*json.RawMessage, len(t.UnlockBlocks)),
	}
	jTransaction.Type = int(TransactionPayloadTypeID)
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
type DustAllowanceFunc func(addr Address) (dustAllowanceSum uint64, amountDustOutputs int64, err error)

// NewDustSemanticValidation returns a SemanticValidationFunc which verifies whether
// a transaction fulfils the semantics regarding dust outputs:
//	A transaction:
//		- consuming a SigLockedDustAllowanceOutput on address A or
//		- creating a SigLockedSingleOutput with deposit amount < OutputSigLockedDustAllowanceOutputMinDeposit (dust output)
//	is only semantically valid, if after the transaction is booked, the number of dust outputs on address A does not exceed the allowed
//	threshold of the sum of min(S / div, dustOutputsCountLimit). Where S is the sum of deposits of all dust allowance outputs on address A.
func NewDustSemanticValidation(div int64, dustOutputsCountLimit int64, dustAllowanceFunc DustAllowanceFunc) SemanticValidationFunc {
	return func(t *Transaction, utxos InputToOutputMapping) error {
		essence := t.Essence.(*TransactionEssence)

		addrToValidate := make(map[string]Address)
		dustAllowanceAddrToBalance := make(map[string]int64)
		dustAllowanceAddrToNumOfDustOutputs := make(map[string]int64)

		for _, output := range essence.Outputs {
			switch out := output.(type) {
			case *SigLockedDustAllowanceOutput:
				addrToValidate[out.Address.(Address).String()] = out.Address.(Address)
				dustAllowanceAddrToBalance[out.Address.(Address).String()] += int64(out.Amount)
			case *SigLockedSingleOutput:
				if out.Amount < OutputSigLockedDustAllowanceOutputMinDeposit {
					addrToValidate[out.Address.(Address).String()] = out.Address.(Address)
					dustAllowanceAddrToNumOfDustOutputs[out.Address.(Address).String()] += 1
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
				addrToValidate[target.(Address).String()] = target.(Address)
				dustAllowanceAddrToNumOfDustOutputs[target.(Address).String()] -= 1
				continue
			}

			if utxo.Type() == OutputSigLockedDustAllowanceOutput {
				addrToValidate[target.(Address).String()] = target.(Address)
				dustAllowanceAddrToBalance[target.(Address).String()] -= int64(deposit)
			}
		}

		for addrKey, addr := range addrToValidate {
			dustAllowanceDepositSumUint64, numDustOutputs, err := dustAllowanceFunc(addr)
			if err != nil {
				return fmt.Errorf("unable to fetch dust allowance information on address %v: %w", addr, err)
			}
			numDustOutputsPrev := numDustOutputs
			numDustOutputs += dustAllowanceAddrToNumOfDustOutputs[addrKey]

			var dustAllowanceDepositSum = int64(dustAllowanceDepositSumUint64)
			// Go integer division floors the value
			prevAllowed := dustAllowanceDepositSum / div
			allowed := (dustAllowanceDepositSum + dustAllowanceAddrToBalance[addrKey]) / div

			// limit
			if allowed > dustOutputsCountLimit {
				allowed = dustOutputsCountLimit
			}

			if numDustOutputs > allowed {
				short := numDustOutputs - allowed
				return fmt.Errorf("%w: addr %s, new num of dust outputs %d (previous %d), allowance deposit %d (previous %d), short %d", ErrInvalidDustAllowance, addrKey, numDustOutputs, numDustOutputsPrev, allowed, prevAllowed, short)
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

	txEssenceBytes, err := txEssence.SigningMessage()
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
	seenInputAddr := make(map[string]int)

	for i, input := range transaction.Inputs {
		in, alreadySeen := input.(*UTXOInput)
		if !alreadySeen {
			return 0, nil, fmt.Errorf("%w: unsupported input type at index %d", ErrUnknownInputType, i)
		}

		// check that we got the needed UTXO
		utxoID := in.ID()
		utxo, has := utxos[utxoID]
		if !has {
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

		target, err := utxo.Target()
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
func createSigValidationFunc(pos int, sig serializer.Serializable, sigBlockIndex int, txEssenceBytes []byte, addr Address) (SigValidationFunc, error) {
	switch addr := addr.(type) {
	case *Ed25519Address:
		return createEd25519SigValidationFunc(pos, sig, sigBlockIndex, addr, txEssenceBytes)
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

	unlockBlocks := make(serializer.Serializables, len(jsontx.UnlockBlocks))
	for i, ele := range jsontx.UnlockBlocks {
		jsonUnlockBlock, err := DeserializeObjectFromJSON(ele, jsonUnlockBlockSelector)
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
