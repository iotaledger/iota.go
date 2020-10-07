package iota

import (
	"bytes"
	"encoding/binary"
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
func (s *Transaction) ID() (*TransactionID, error) {
	data, err := s.Serialize(DeSeriModeNoValidation)
	if err != nil {
		return nil, fmt.Errorf("can't compute transaction ID: %w", err)
	}
	h := blake2b.Sum256(data)
	return &h, nil
}

func (s *Transaction) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkMinByteLength(TransactionBinSerializedMinSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid transaction bytes: %w", err)
		}
		if err := checkType(data, TransactionPayloadTypeID); err != nil {
			return 0, fmt.Errorf("unable to deserialize transaction: %w", err)
		}
	}

	// skip payload type
	bytesReadTotal := TypeDenotationByteSize
	data = data[TypeDenotationByteSize:]

	tx, txBytesRead, err := DeserializeObject(data, deSeriMode, TypeDenotationByte, TransactionEssenceSelector)
	if err != nil {
		return 0, fmt.Errorf("%w: unable to deserialize transaction essence within transaction", err)
	}
	bytesReadTotal += txBytesRead
	s.Essence = tx

	// TODO: tx must be a TransactionEssence but might be something else in the future
	inputCount := uint16(len(tx.(*TransactionEssence).Inputs))

	// advance to unlock blocks
	data = data[txBytesRead:]
	unlockBlocks, unlockBlocksByteRead, err := DeserializeArrayOfObjects(data, deSeriMode, TypeDenotationByte, UnlockBlockSelector, &ArrayRules{
		Min:    inputCount,
		Max:    inputCount,
		MinErr: ErrUnlockBlocksMustMatchInputCount,
		MaxErr: ErrUnlockBlocksMustMatchInputCount,
	})
	if err != nil {
		return 0, fmt.Errorf("%w: unable to deserialize unlock blocks", err)
	}
	bytesReadTotal += unlockBlocksByteRead

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := ValidateUnlockBlocks(unlockBlocks, UnlockBlocksSigUniqueAndRefValidator()); err != nil {
			return 0, err
		}
	}

	s.UnlockBlocks = unlockBlocks

	return bytesReadTotal, nil
}

func (s *Transaction) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := s.SyntacticallyValidate(); err != nil {
			return nil, err
		}
	}

	var b bytes.Buffer
	if err := binary.Write(&b, binary.LittleEndian, TransactionPayloadTypeID); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize transaction ID", err)
	}

	// write transaction
	txBytes, err := s.Essence.Serialize(deSeriMode)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to serialize transaction's essence", err)
	}
	if _, err := b.Write(txBytes); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize transaction's essence to buffer", err)
	}

	// write unlock blocks and count
	if err := binary.Write(&b, binary.LittleEndian, uint16(len(s.UnlockBlocks))); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize transaction's unlock block count", err)
	}
	for i := range s.UnlockBlocks {
		unlockBlockSer, err := s.UnlockBlocks[i].Serialize(deSeriMode)
		if err != nil {
			return nil, fmt.Errorf("%w: unable to serialize transaction's unlock block at pos %d", err, i)
		}
		if _, err := b.Write(unlockBlockSer); err != nil {
			return nil, fmt.Errorf("%w: unable to serialize transaction's unlock block at pos %d to buffer", err, i)
		}
	}

	return b.Bytes(), nil
}

func (s *Transaction) MarshalJSON() ([]byte, error) {
	jsonSigTxPayload := &jsontransaction{
		UnlockBlocks: make([]*json.RawMessage, len(s.UnlockBlocks)),
	}
	jsonSigTxPayload.Type = int(TransactionPayloadTypeID)
	txJson, err := s.Essence.MarshalJSON()
	if err != nil {
		return nil, err
	}
	rawMsgTxJson := json.RawMessage(txJson)
	jsonSigTxPayload.Essence = &rawMsgTxJson
	for i, ub := range s.UnlockBlocks {
		jsonUB, err := ub.MarshalJSON()
		if err != nil {
			return nil, err
		}
		rawMsgJsonUB := json.RawMessage(jsonUB)
		jsonSigTxPayload.UnlockBlocks[i] = &rawMsgJsonUB
	}
	return json.Marshal(jsonSigTxPayload)
}

func (s *Transaction) UnmarshalJSON(bytes []byte) error {
	jsonSigTxPayload := &jsontransaction{}
	if err := json.Unmarshal(bytes, jsonSigTxPayload); err != nil {
		return err
	}
	seri, err := jsonSigTxPayload.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*Transaction)
	return nil
}

// SyntacticallyValidate syntactically validates the Transaction:
//	1. The TransactionEssence isn't nil
//	2. syntactic validation on the TransactionEssence
//	3. input and unlock blocks count must match
func (s *Transaction) SyntacticallyValidate() error {

	if s.Essence == nil {
		return fmt.Errorf("%w: transaction is nil", ErrInvalidTransactionEssence)
	}

	if s.UnlockBlocks == nil {
		return fmt.Errorf("%w: unlock blocks are nil", ErrInvalidTransactionEssence)
	}

	txEssence, ok := s.Essence.(*TransactionEssence)
	if !ok {
		return fmt.Errorf("%w: transaction essence is not *TransactionEssence", ErrInvalidTransactionEssence)
	}

	if err := txEssence.SyntacticallyValidate(); err != nil {
		return fmt.Errorf("%w: transaction essence part is invalid", err)
	}

	inputCount := len(txEssence.Inputs)
	unlockBlockCount := len(s.UnlockBlocks)
	if inputCount != unlockBlockCount {
		return fmt.Errorf("%w: num of inputs %d, num of unlock blocks %d", ErrUnlockBlocksMustMatchInputCount, inputCount, unlockBlockCount)
	}

	if err := ValidateUnlockBlocks(s.UnlockBlocks, UnlockBlocksSigUniqueAndRefValidator()); err != nil {
		return fmt.Errorf("%w: invalid unlock blocks", err)
	}

	return nil
}

// SigValidationFunc is a function which when called tells whether
// its signature verification computation was successful or not.
type SigValidationFunc = func() error

// InputToOutputMapping maps inputs to their origin UTXOs.
type InputToOutputMapping = map[UTXOInputID]SigLockedSingleOutput

// SemanticallyValidate semantically validates the Transaction
// by checking that the given input UTXOs are spent entirely and the signatures
// provided are valid. SyntacticallyValidate() should be called before SemanticallyValidate() to
// ensure that the essence part of the transaction is syntactically valid.
func (s *Transaction) SemanticallyValidate(utxos InputToOutputMapping) error {

	txEssence, ok := s.Essence.(*TransactionEssence)
	if !ok {
		return fmt.Errorf("%w: transaction is not *TransactionEssence", ErrInvalidTransactionEssence)
	}

	txEssenceBytes, err := txEssence.Serialize(DeSeriModeNoValidation)
	if err != nil {
		return err
	}

	inputSum, sigValidFuncs, err := s.SemanticallyValidateInputs(utxos, txEssence, txEssenceBytes)
	if err != nil {
		return err
	}

	outputSum, err := s.SemanticallyValidateOutputs(txEssence)
	if err != nil {
		return err
	}

	if inputSum != outputSum {
		return fmt.Errorf("%w: inputs sum %d, outputs sum %d", ErrInputOutputSumMismatch, inputSum, outputSum)
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
func (s *Transaction) SemanticallyValidateInputs(utxos InputToOutputMapping, transaction *TransactionEssence, txEssenceBytes []byte) (uint64, []SigValidationFunc, error) {
	var sigValidFuncs []SigValidationFunc
	var inputSum uint64

	for i, input := range transaction.Inputs {
		// TODO: switch out with type switch
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

		inputSum += utxo.Amount

		var sigBlock *SignatureUnlockBlock
		var refSigBlockIndex int
		switch ub := s.UnlockBlocks[i].(type) {
		case *SignatureUnlockBlock:
			sigBlock = ub
			refSigBlockIndex = i
		case *ReferenceUnlockBlock:
			// it is ensured by the syntactical validation that
			// the corresponding signature unlock block exists
			refSigBlockIndex = int(ub.Reference)
			sigBlock = s.UnlockBlocks[refSigBlockIndex].(*SignatureUnlockBlock)
		}

		switch addr := utxo.Address.(type) {
		case *WOTSAddress:
			// TODO: implement
		case *Ed25519Address:
			ed25519Sig, isEd25519Sig := sigBlock.Signature.(*Ed25519Signature)
			if !isEd25519Sig {
				return 0, nil, fmt.Errorf("%w: UTXO at index %d has an Ed25519 address but its corresponding signature is of type %T (at index %d)", ErrSignatureAndAddrIncompatible, i, sigBlock.Signature, refSigBlockIndex)
			}

			sigValidFuncs = append(sigValidFuncs, func() error {
				if err := ed25519Sig.Valid(txEssenceBytes, addr); err != nil {
					return fmt.Errorf("%w: input at index %d, signature block at index %d", err, i, refSigBlockIndex)
				}
				return nil
			})

		default:
			return 0, nil, fmt.Errorf("%w: unsupported address type at index %d", ErrUnknownAddrType, i)
		}

	}

	return inputSum, sigValidFuncs, nil
}

// SemanticallyValidateOutputs accumulates the sum of all outputs.
// This function should only be called from SemanticallyValidate().
func (s *Transaction) SemanticallyValidateOutputs(transaction *TransactionEssence) (uint64, error) {
	var outputSum uint64
	for i, output := range transaction.Outputs {
		// TODO: switch out with type switch
		out, ok := output.(*SigLockedSingleOutput)
		if !ok {
			return 0, fmt.Errorf("%w: unsupported output type at index %d", ErrUnknownOutputType, i)
		}
		outputSum += out.Amount
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
