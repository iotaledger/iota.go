package iotago

import (
	"context"
	"fmt"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
)

const (
	// SignedTransactionIDLength defines the length of a SignedTransactionID.
	SignedTransactionIDLength = SlotIdentifierLength

	// TransactionIDLength defines the length of a TransactionID.
	TransactionIDLength = SlotIdentifierLength
)

var (
	// ErrMissingUTXO gets returned if an UTXO is missing to commence a certain operation.
	ErrMissingUTXO = ierrors.New("missing utxo")
	// ErrInputOutputSumMismatch gets returned if a transaction does not spend the entirety of the inputs to the outputs.
	ErrInputOutputSumMismatch = ierrors.New("inputs and outputs do not spend/deposit the same amount")
	// ErrManaOverflow gets returned when there is an under- or overflow in Mana calculations.
	ErrManaOverflow = ierrors.New("under- or overflow in Mana calculations")
	// ErrUnknownSignatureType gets returned for unknown signature types.
	ErrUnknownSignatureType = ierrors.New("unknown signature type")
	// ErrSignatureAndAddrIncompatible gets returned if an address of an input has a companion signature unlock with the wrong signature type.
	ErrSignatureAndAddrIncompatible = ierrors.New("address and signature type are not compatible")
	// ErrInvalidInputUnlock gets returned when an input unlock is invalid.
	ErrInvalidInputUnlock = ierrors.New("invalid input unlock")
	// ErrSenderFeatureNotUnlocked gets returned when an output contains a SenderFeature with an ident which is not unlocked.
	ErrSenderFeatureNotUnlocked = ierrors.New("sender feature is not unlocked")
	// ErrIssuerFeatureNotUnlocked gets returned when an output contains a IssuerFeature with an ident which is not unlocked.
	ErrIssuerFeatureNotUnlocked = ierrors.New("issuer feature is not unlocked")
	// ErrReturnAmountNotFulFilled gets returned when a return amount in a transaction is not fulfilled by the output side.
	ErrReturnAmountNotFulFilled = ierrors.New("return amount not fulfilled")
	// ErrInputOutputManaMismatch gets returned if Mana is not balanced across inputs and outputs/allotments.
	ErrInputOutputManaMismatch = ierrors.New("inputs and outputs do not contain the same amount of Mana")
	// ErrInputCreationAfterTxCreation gets returned if an input has creation slot after the transaction creation slot.
	ErrInputCreationAfterTxCreation = ierrors.New("input creation slot after tx creation slot")
)

var (
	EmptySignedTransactionID = SignedTransactionID{}
	EmptyTransactionID       = TransactionID{}
)

type SignedTransactionID = SlotIdentifier

// SignedTransactionIDs are IDs of signed transactions.
type SignedTransactionIDs []SignedTransactionID

// SignedTransactionIDFromData returns a new SignedTransactionID for the given data by hashing it with blake2b and appending the creation slot index.
func SignedTransactionIDFromData(creationSlot SlotIndex, data []byte) SignedTransactionID {
	return SlotIdentifierRepresentingData(creationSlot, data)
}

type TransactionID = SlotIdentifier

// TransactionIDs are IDs of transactions.
type TransactionIDs []TransactionID

// TransactionIDFromData returns a new SignedTransactionID for the given data by hashing it with blake2b and appending the creation slot index.
func TransactionIDFromData(creationSlot SlotIndex, data []byte) TransactionID {
	return SlotIdentifierRepresentingData(creationSlot, data)
}

type TransactionContextInputs ContextInputs[Input]

// SignedTransaction is a transaction with its inputs, outputs and unlocks.
type SignedTransaction struct {
	API API
	// The transaction essence, respectively the transfer part of a SignedTransaction.
	Transaction *Transaction `serix:"0,mapKey=transaction"`
	// The unlocks defining the unlocking data for the inputs within the Transaction.
	Unlocks Unlocks `serix:"1,mapKey=unlocks"`
}

func (t *SignedTransaction) SetDeserializationContext(ctx context.Context) {
	t.API = APIFromContext(ctx)
}

func (t *SignedTransaction) Clone() Payload {
	return &SignedTransaction{
		API:         t.API,
		Transaction: t.Transaction.Clone(),
		Unlocks:     t.Unlocks.Clone(),
	}
}

func (t *SignedTransaction) PayloadType() PayloadType {
	return PayloadSignedTransaction
}

// OutputsSet returns an OutputSet from the SignedTransaction's outputs, mapped by their OutputID.
func (t *SignedTransaction) OutputsSet() (OutputSet, error) {
	txID, err := t.ID()
	if err != nil {
		return nil, err
	}
	set := make(OutputSet)
	for index, output := range t.Transaction.Outputs {
		set[OutputIDFromTransactionIDAndIndex(txID, uint16(index))] = output
	}

	return set, nil
}

// ID computes the ID of the SignedTransaction.
func (t *SignedTransaction) ID() (SignedTransactionID, error) {
	unlocksBytes, err := t.API.Encode(t.Unlocks)
	if err != nil {
		return SignedTransactionID{}, ierrors.Errorf("can't compute unlock bytes: %w", err)
	}

	transactionID, err := t.Transaction.ID()
	if err != nil {
		return SignedTransactionID{}, ierrors.Errorf("can't compute transaction ID: %w", err)
	}

	return SignedTransactionIDFromData(t.Transaction.CreationSlot, byteutils.ConcatBytes(lo.PanicOnErr(transactionID.Identifier().Bytes()), unlocksBytes)), nil
}

func (t *SignedTransaction) Inputs() ([]*UTXOInput, error) {
	references := make([]*UTXOInput, 0, len(t.Transaction.Inputs))
	for _, input := range t.Transaction.Inputs {
		switch castInput := input.(type) {
		case *UTXOInput:
			references = append(references, castInput)
		default:
			return nil, ErrUnknownInputType
		}
	}

	return references, nil
}

func (t *SignedTransaction) ContextInputs() (TransactionContextInputs, error) {
	references := make(TransactionContextInputs, 0, len(t.Transaction.ContextInputs))
	for _, input := range t.Transaction.ContextInputs {
		switch castInput := input.(type) {
		case *CommitmentInput, *BlockIssuanceCreditInput, *RewardInput:
			references = append(references, castInput)
		default:
			return nil, ErrUnknownContextInputType
		}
	}

	return references, nil
}

func (t *SignedTransaction) BICInputs() ([]*BlockIssuanceCreditInput, error) {
	references := make([]*BlockIssuanceCreditInput, 0, len(t.Transaction.ContextInputs))
	for _, input := range t.Transaction.ContextInputs {
		switch castInput := input.(type) {
		case *BlockIssuanceCreditInput:
			references = append(references, castInput)
		case *CommitmentInput, *RewardInput:
			// ignore this type
		default:
			return nil, ErrUnknownContextInputType
		}
	}

	return references, nil
}

func (t *SignedTransaction) RewardInputs() ([]*RewardInput, error) {
	references := make([]*RewardInput, 0, len(t.Transaction.ContextInputs))
	for _, input := range t.Transaction.ContextInputs {
		switch castInput := input.(type) {
		case *RewardInput:
			references = append(references, castInput)
		case *CommitmentInput, *BlockIssuanceCreditInput:
			// ignore this type
		default:
			return nil, ErrUnknownContextInputType
		}
	}

	return references, nil
}

// Returns the first commitment input in the transaction if it exists or nil.
func (t *SignedTransaction) CommitmentInput() *CommitmentInput {
	for _, input := range t.Transaction.ContextInputs {
		switch castInput := input.(type) {
		case *BlockIssuanceCreditInput, *RewardInput:
			// ignore this type
		case *CommitmentInput:
			return castInput
		default:
			return nil
		}
	}

	return nil
}

func (t *SignedTransaction) Size() int {
	// PayloadType
	return serializer.TypeDenotationByteSize +
		t.Transaction.Size() +
		t.Unlocks.Size()
}

func (t *SignedTransaction) String() string {
	// TODO: stringify for debugging purposes
	return fmt.Sprintf("SignedTransaction[%v, %v]", t.Transaction, t.Unlocks)
}

// syntacticallyValidate syntactically validates the SignedTransaction.
func (t *SignedTransaction) syntacticallyValidate() error {
	// limit unlock block count = input count
	if len(t.Unlocks) != len(t.Transaction.Inputs) {
		return ierrors.Errorf("unlock block count must match inputs in transaction, %d vs. %d", len(t.Unlocks), len(t.Transaction.Inputs))
	}

	if err := t.Transaction.syntacticallyValidate(t.API); err != nil {
		return ierrors.Errorf("transaction is invalid: %w", err)
	}

	if err := ValidateUnlocks(t.Unlocks,
		UnlocksSigUniqueAndRefValidator(t.API),
	); err != nil {
		return ierrors.Errorf("invalid unlocks: %w", err)
	}

	return nil
}

func (t *SignedTransaction) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
	// we account for the network traffic only on "Payload" level
	workScoreSignedTransactionData, err := workScoreStructure.DataByte.Multiply(t.Size())
	if err != nil {
		return 0, err
	}

	workScoreTransaction, err := t.Transaction.WorkScore(workScoreStructure)
	if err != nil {
		return 0, err
	}

	workScoreUnlocks, err := t.Unlocks.WorkScore(workScoreStructure)
	if err != nil {
		return 0, err
	}

	return workScoreSignedTransactionData.Add(workScoreTransaction, workScoreUnlocks)
}
