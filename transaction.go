package iotago

import (
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/iota.go/v4/util"
)

const (
	// TransactionIDLength defines the length of a Transaction ID.
	TransactionIDLength = IdentifierLength
)

var (
	// ErrMissingUTXO gets returned if an UTXO is missing to commence a certain operation.
	ErrMissingUTXO = errors.New("missing utxo")
	// ErrInputOutputSumMismatch gets returned if a transaction does not spend the entirety of the inputs to the outputs.
	ErrInputOutputSumMismatch = errors.New("inputs and outputs do not spend/deposit the same amount")
	// ErrSignatureAndAddrIncompatible gets returned if an address of an input has a companion signature unlock with the wrong signature type.
	ErrSignatureAndAddrIncompatible = errors.New("address and signature type are not compatible")
	// ErrInvalidInputUnlock gets returned when an input unlock is invalid.
	ErrInvalidInputUnlock = errors.New("invalid input unlock")
	// ErrSenderFeatureNotUnlocked gets returned when an output contains a SenderFeature with an ident which is not unlocked.
	ErrSenderFeatureNotUnlocked = errors.New("sender feature is not unlocked")
	// ErrIssuerFeatureNotUnlocked gets returned when an output contains a IssuerFeature with an ident which is not unlocked.
	ErrIssuerFeatureNotUnlocked = errors.New("issuer feature is not unlocked")
	// ErrReturnAmountNotFulFilled gets returned when a return amount in a transaction is not fulfilled by the output side.
	ErrReturnAmountNotFulFilled = errors.New("return amount not fulfilled")
	// ErrInputOutputManaMismatch gets returned if Mana is not balanced accross inputs and outputs/allotments.
	ErrInputOutputManaMismatch = errors.New("inputs and outputs do not contain the same amount of Mana")
)

// TransactionID is the ID of a Transaction.
type TransactionID = Identifier

// TransactionIDs are IDs of transactions.
type TransactionIDs []TransactionID

// Transaction is a transaction with its inputs, outputs and unlocks.
type Transaction struct {
	// The transaction essence, respectively the transfer part of a Transaction.
	Essence *TransactionEssence `serix:"0,mapKey=essence"`
	// The unlocks defining the unlocking data for the inputs within the Essence.
	Unlocks Unlocks `serix:"1,mapKey=unlocks"`
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
		set[OutputIDFromTransactionIDAndIndex(txID, uint16(index))] = output
	}
	return set, nil
}

// ID computes the ID of the Transaction.
func (t *Transaction) ID() (TransactionID, error) {
	data, err := internalEncode(t)
	if err != nil {
		return TransactionID{}, fmt.Errorf("can't compute transaction ID: %w", err)
	}
	return IdentifierFromData(data), nil
}

func (t *Transaction) Inputs() ([]IndexedUTXOReferencer, error) {
	references := make([]IndexedUTXOReferencer, len(t.Essence.Inputs))
	for i, input := range t.Essence.Inputs {
		switch castInput := input.(type) {
		case *BICInput:
			// ignore this type
		case IndexedUTXOReferencer:
			references[i] = castInput
		default:
			return nil, ErrUnexpectedUnderlyingType
		}
	}

	return references, nil
}

func (t *Transaction) BICInputs() ([]*BICInput, error) {
	references := make([]*BICInput, len(t.Essence.Inputs))
	for i, input := range t.Essence.Inputs {
		switch castInput := input.(type) {
		case *BICInput:
			references[i] = castInput
		case IndexedUTXOReferencer:
			// ignore this type
		default:
			return nil, ErrUnexpectedUnderlyingType
		}
	}

	return references, nil
}

func (t *Transaction) Size() int {
	return util.NumByteLen(uint32(PayloadTransaction)) +
		t.Essence.Size() +
		t.Unlocks.Size()
}

func (t *Transaction) String() string {
	return "iotago.Transaction(" + lo.PanicOnErr(t.ID()).ToHex() + ")"
}

// syntacticallyValidate syntactically validates the Transaction.
func (t *Transaction) syntacticallyValidate(protoParams *ProtocolParameters) error {
	if err := t.Essence.syntacticallyValidate(protoParams); err != nil {
		return fmt.Errorf("transaction essence is invalid: %w", err)
	}

	if err := ValidateUnlocks(t.Unlocks,
		UnlocksSigUniqueAndRefValidator(),
	); err != nil {
		return fmt.Errorf("invalid unlocks: %w", err)
	}

	return nil
}
