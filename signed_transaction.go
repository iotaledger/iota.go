package iotago

import (
	"context"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
	"github.com/iotaledger/hive.go/stringify"
)

var (
	// ErrMissingUTXO gets returned if an UTXO is missing to commence a certain operation.
	ErrMissingUTXO = ierrors.New("missing utxo")
	// ErrInputOutputBaseTokenMismatch gets returned if a transaction does not spend the entirety of the inputs to the outputs.
	ErrInputOutputBaseTokenMismatch = ierrors.New("inputs and outputs do not spend/deposit the same amount of base tokens")
	// ErrManaOverflow gets returned when there is an under- or overflow in Mana calculations.
	ErrManaOverflow = ierrors.New("under- or overflow in Mana calculations")
	// ErrUnknownSignatureType gets returned for unknown signature types.
	ErrUnknownSignatureType = ierrors.New("unknown signature type")
	// ErrSignatureAndAddrIncompatible gets returned if an address of an input has a companion signature unlock with the wrong signature type.
	ErrSignatureAndAddrIncompatible = ierrors.New("address and signature type are not compatible")
	// ErrSenderFeatureNotUnlocked gets returned when an output contains a SenderFeature with an address which is not unlocked.
	ErrSenderFeatureNotUnlocked = ierrors.New("sender feature is not unlocked")
	// ErrIssuerFeatureNotUnlocked gets returned when an output contains a IssuerFeature with an address which is not unlocked.
	ErrIssuerFeatureNotUnlocked = ierrors.New("issuer feature is not unlocked")
	// ErrReturnAmountNotFulFilled gets returned when a return amount in a transaction is not fulfilled by the output side.
	ErrReturnAmountNotFulFilled = ierrors.New("return amount not fulfilled")
	// ErrInputOutputManaMismatch gets returned if Mana is not balanced across inputs and outputs/allotments.
	ErrInputOutputManaMismatch = ierrors.New("inputs and outputs do not contain the same amount of Mana")
	// ErrInputCreationAfterTxCreation gets returned if an input has creation slot after the transaction creation slot.
	ErrInputCreationAfterTxCreation = ierrors.New("input creation slot after tx creation slot")
)

type TransactionContextInputs ContextInputs[ContextInput]

// SignedTransaction is a transaction with its inputs, outputs and unlocks.
type SignedTransaction struct {
	API API
	// The transaction essence, respectively the transfer part of a SignedTransaction.
	Transaction *Transaction `serix:""`
	// The unlocks defining the unlocking data for the inputs within the Transaction.
	Unlocks Unlocks `serix:""`
}

// ID computes the ID of the SignedTransaction.
func (t *SignedTransaction) ID() (SignedTransactionID, error) {
	transactionBytes, err := t.API.Encode(t.Transaction)
	if err != nil {
		return EmptySignedTransactionID, ierrors.Errorf("can't compute unlock bytes: %w", err)
	}

	unlocksBytes, err := t.API.Encode(t.Unlocks)
	if err != nil {
		return EmptySignedTransactionID, ierrors.Errorf("can't compute unlock bytes: %w", err)
	}

	return SignedTransactionIDRepresentingData(t.Transaction.CreationSlot, byteutils.ConcatBytes(transactionBytes, unlocksBytes)), nil
}

func (t *SignedTransaction) Size() int {
	// PayloadType
	return serializer.SmallTypeDenotationByteSize +
		t.Transaction.Size() +
		t.Unlocks.Size()
}

func (t *SignedTransaction) PayloadType() PayloadType {
	return PayloadSignedTransaction
}

func (t *SignedTransaction) Clone() Payload {
	return &SignedTransaction{
		API:         t.API,
		Transaction: t.Transaction.Clone(),
		Unlocks:     t.Unlocks.Clone(),
	}
}

func (t *SignedTransaction) SetDeserializationContext(ctx context.Context) {
	t.API = APIFromContext(ctx)
}

// String returns a human readable version of the SignedTransaction.
func (t *SignedTransaction) String() string {
	return stringify.Struct("SignedTransaction",
		stringify.NewStructField("Transaction", t.Transaction),
		stringify.NewStructField("Unlocks", t.Unlocks),
	)
}

// syntacticallyValidate syntactically validates the SignedTransaction.
func (t *SignedTransaction) syntacticallyValidate() error {
	// limit unlock block count = input count
	inputs, err := t.Transaction.Inputs()
	if err != nil {
		return err
	}

	if len(t.Unlocks) != len(inputs) {
		return ierrors.Errorf("unlock block count must match inputs in transaction, %d vs. %d", len(t.Unlocks), len(inputs))
	}

	if err := t.Transaction.SyntacticallyValidate(t.API); err != nil {
		return ierrors.Errorf("transaction is invalid: %w", err)
	}

	if err := ValidateUnlocks(t.Unlocks,
		SignaturesUniqueAndReferenceUnlocksValidator(t.API),
	); err != nil {
		return ierrors.Errorf("invalid unlocks: %w", err)
	}

	return nil
}

func (t *SignedTransaction) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	// we account for the network traffic only on "Payload" level
	workScoreSignedTransactionData, err := workScoreParameters.DataByte.Multiply(t.Size())
	if err != nil {
		return 0, err
	}

	workScoreTransaction, err := t.Transaction.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	workScoreUnlocks, err := t.Unlocks.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	// we include the block offset in the payload WorkScore
	return workScoreParameters.Block.Add(workScoreSignedTransactionData, workScoreTransaction, workScoreUnlocks)
}
