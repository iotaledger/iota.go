package iota

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"golang.org/x/crypto/blake2b"
)

const (
	// Defines the signed transaction's payload ID.
	SignedTransactionPayloadID uint32 = 0

	// Defines the maximum amount of inputs within a transaction.
	MaxInputsCount = 126
	// Defines the minimum amount of inputs within a transaction.
	MinInputsCount = 1
	// Defines the maximum amount of outputs within a transaction.
	MaxOutputsCount = 126
	// Defines the minimum amount of inputs within a transaction.
	MinOutputsCount = 1

	// Defines the length of a signed transaction payload hash.
	SignedTransactionPayloadHashLength = blake2b.Size256

	// Defines the minimum size of a signed transaction payload.
	SignedTransactionPayloadMinSize = UInt32ByteSize
)

var (
	// Returned if the count of inputs is too small.
	ErrMinInputsNotReached = fmt.Errorf("min %d input(s) are required within a transaction", MinInputsCount)
	// Returned if the count of inputs is too big.
	ErrMaxInputsExceeded = fmt.Errorf("max %d input(s) are allowed within a transaction", MaxInputsCount)
	// Returned if the count of outputs is too small.
	ErrMinOutputsNotReached = fmt.Errorf("min %d output(s) are required within a transaction", MinOutputsCount)
	// Returned if the count of outputs is too big.
	ErrMaxOutputsExceeded = fmt.Errorf("max %d output(s) are allowed within a transaction", MaxOutputsCount)
	// Returned if the count of unlock blocks doesn't match the count of inputs.
	ErrUnlockBlocksMustMatchInputCount = errors.New("the count of unlock blocks must match the inputs of the transaction")
	// Returned if the transaction within a signed transaction payload is invalid.
	ErrInvalidTransaction = errors.New("transaction is invalid")
	// Returned if an UTXO is missing to commence a certain operation.
	ErrMissingUTXO = errors.New("missing utxo")
	// Returned if a transaction does not spend the entirety of the inputs to the outputs.
	ErrInputOutputSumMismatch = errors.New("inputs and outputs do not spend/deposit the same amount")
	// Returned if an address of an input has a companion signature unlock block with the wrong signature type.
	ErrSignatureAndAddrIncompatible = errors.New("address and signature type are not compatible")

	// restrictions around input within a transaction.
	inputsArrayBound = ArrayRules{
		Min:                         MinInputsCount,
		Max:                         MaxInputsCount,
		MinErr:                      ErrMinInputsNotReached,
		MaxErr:                      ErrMaxInputsExceeded,
		ElementBytesLexicalOrder:    true,
		ElementBytesLexicalOrderErr: ErrInputsOrderViolatesLexicalOrder,
	}

	// restrictions around outputs within a transaction.
	outputsArrayBound = ArrayRules{
		Min:                         MinInputsCount,
		Max:                         MaxInputsCount,
		MinErr:                      ErrMinOutputsNotReached,
		MaxErr:                      ErrMaxOutputsExceeded,
		ElementBytesLexicalOrder:    true,
		ElementBytesLexicalOrderErr: ErrOutputsOrderViolatesLexicalOrder,
	}
)

// SignedTransactionPayloadHash is the hash of a SignedTransactionPayload.
type SignedTransactionPayloadHash = [SignedTransactionPayloadHashLength]byte

// SignedTransactionPayload is a transaction with its inputs, outputs and unlock blocks.
type SignedTransactionPayload struct {
	// The transaction respectively transfer part of a signed transaction payload.
	Transaction Serializable `json:"transaction"`
	// The unlock blocks defining the unlocking data for the inputs within the transaction.
	UnlockBlocks Serializables `json:"unlock_blocks"`
}

// Hash computes the hash of the SignedTransactionPayload.
func (s *SignedTransactionPayload) Hash() (*SignedTransactionPayloadHash, error) {
	data, err := s.Serialize(DeSeriModeNoValidation)
	if err != nil {
		return nil, fmt.Errorf("can't compute signed transaction payload hash: %w", err)
	}
	h := blake2b.Sum256(data)
	return &h, nil
}

func (s *SignedTransactionPayload) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkMinByteLength(SignedTransactionPayloadMinSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid signed transaction payload bytes: %w", err)
		}
		if err := checkType(data, SignedTransactionPayloadID); err != nil {
			return 0, fmt.Errorf("unable to deserialize signed transaction payload: %w", err)
		}
	}

	// skip payload type
	bytesReadTotal := TypeDenotationByteSize
	data = data[TypeDenotationByteSize:]

	tx, txBytesRead, err := DeserializeObject(data, deSeriMode, TypeDenotationByte, TransactionSelector)
	if err != nil {
		return 0, fmt.Errorf("%w: unable to deserialize transaction within signed transaction payload", err)
	}
	bytesReadTotal += txBytesRead
	s.Transaction = tx

	// TODO: tx must be an unsigned tx but might be something else in the future
	inputCount := uint16(len(tx.(*UnsignedTransaction).Inputs))

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

func (s *SignedTransactionPayload) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := s.SyntacticallyValidate(); err != nil {
			return nil, err
		}
	}

	var b bytes.Buffer
	if err := binary.Write(&b, binary.LittleEndian, SignedTransactionPayloadID); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize signed transaction payload ID", err)
	}

	// write transaction
	txBytes, err := s.Transaction.Serialize(deSeriMode)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to serialize signed transaction payload's inner transaction", err)
	}
	if _, err := b.Write(txBytes); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize signed transaction payload's inner transaction to buffer", err)
	}

	// write unlock blocks and count
	if err := binary.Write(&b, binary.LittleEndian, uint16(len(s.UnlockBlocks))); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize signed transaction payload's unlock block count", err)
	}
	for i := range s.UnlockBlocks {
		unlockBlockSer, err := s.UnlockBlocks[i].Serialize(deSeriMode)
		if err != nil {
			return nil, fmt.Errorf("%w: unable to serialize signed transaction payload's unlock block at pos %d", err, i)
		}
		if _, err := b.Write(unlockBlockSer); err != nil {
			return nil, fmt.Errorf("%w: unable to serialize signed transaction payload's unlock block at pos %d to buffer", err, i)
		}
	}

	return b.Bytes(), nil
}

// SyntacticallyValidate syntactically validates the SignedTransactionPayload:
//	1. The UnsignedTransaction isn't nil
//	2. syntactic validation on the UnsignedTransaction
//	3. input and unlock blocks count must match
func (s *SignedTransactionPayload) SyntacticallyValidate() error {

	if s.Transaction == nil {
		return fmt.Errorf("%w: transaction is nil", ErrInvalidTransaction)
	}

	if s.UnlockBlocks == nil {
		return fmt.Errorf("%w: unlock blocks are nil", ErrInvalidTransaction)
	}

	unsignedPart, ok := s.Transaction.(*UnsignedTransaction)
	if !ok {
		return fmt.Errorf("%w: transaction is not *UnsignedTransaction", ErrInvalidTransaction)
	}

	if err := unsignedPart.SyntacticallyValidate(); err != nil {
		return fmt.Errorf("%w: unsigned transaction part is invalid", err)
	}

	inputCount := len(unsignedPart.Inputs)
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
// its signature verification compution was successful or not.
type SigValidationFunc = func() error

// InputToOutputMapping maps inputs to their origin UTXOs.
type InputToOutputMapping = map[UTXOInputID]SigLockedSingleDeposit

// SemanticallyValidate semantically validates the SignedTransactionPayload
// by checking that the given input UTXOs are spent entirely and the signatures
// provided are valid. SyntacticallyValidate() should be called before SemanticallyValidate() to
// ensure that the unsigned part of the transaction is syntactically valid.
func (s *SignedTransactionPayload) SemanticallyValidate(utxos InputToOutputMapping) error {

	unsignedPart, ok := s.Transaction.(*UnsignedTransaction)
	if !ok {
		return fmt.Errorf("%w: transaction is not *UnsignedTransaction", ErrInvalidTransaction)
	}

	unsignedPartBytes, err := unsignedPart.Serialize(DeSeriModeNoValidation)
	if err != nil {
		return err
	}

	inputSum, sigValidFuncs, err := s.SemanticallyValidateInputs(utxos, unsignedPart, unsignedPartBytes)
	if err != nil {
		return err
	}

	outputSum, err := s.SemanticallyValidateOutputs(unsignedPart)
	if err != nil {
		return err
	}

	if inputSum != outputSum {
		return fmt.Errorf("%w: inputs sum %d, outputs sum %d", ErrInputOutputSumMismatch, inputSum, outputSum)
	}

	// sig verifications run at the end as they are the most computationally expensive operation
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
func (s *SignedTransactionPayload) SemanticallyValidateInputs(utxos InputToOutputMapping, transaction *UnsignedTransaction, unsignedPartBytes []byte) (uint64, []SigValidationFunc, error) {
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
				if err := ed25519Sig.Valid(unsignedPartBytes, addr); err != nil {
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
func (s *SignedTransactionPayload) SemanticallyValidateOutputs(transaction *UnsignedTransaction) (uint64, error) {
	var outputSum uint64
	for i, output := range transaction.Outputs {
		// TODO: switch out with type switch
		out, ok := output.(*SigLockedSingleDeposit)
		if !ok {
			return 0, fmt.Errorf("%w: unsupported output type at index %d", ErrUnknownOutputType, i)
		}
		outputSum += out.Amount
	}

	return outputSum, nil
}
