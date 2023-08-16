package iotago

import (
	"fmt"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	// TransactionIDLength defines the length of a Transaction ID.
	TransactionIDLength = IdentifierLength
)

var (
	// ErrMissingUTXO gets returned if an UTXO is missing to commence a certain operation.
	ErrMissingUTXO = ierrors.New("missing utxo")
	// ErrInputOutputSumMismatch gets returned if a transaction does not spend the entirety of the inputs to the outputs.
	ErrInputOutputSumMismatch = ierrors.New("inputs and outputs do not spend/deposit the same amount")
	// ErrManaOverflow gets returned when there is an under- or overflow in Mana calculations.
	ErrManaOverflow = ierrors.New("under- or overflow in Mana calculations")
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
	// ErrUnknownTransactionEssenceType gets returned for unknown transaction essence types.
	ErrUnknownTransactionEssenceType = ierrors.New("unknown transaction essence type")
)

// TransactionID is the ID of a Transaction.
type TransactionID = Identifier

// TransactionIDs are IDs of transactions.
type TransactionIDs []TransactionID

type TransactionContextInputs ContextInputs[Input]

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
func (t *Transaction) OutputsSet(api API) (OutputSet, error) {
	txID, err := t.ID(api)
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
func (t *Transaction) ID(api API) (TransactionID, error) {
	data, err := api.Encode(t)
	if err != nil {
		return TransactionID{}, ierrors.Errorf("can't compute transaction ID: %w", err)
	}

	return IdentifierFromData(data), nil
}

func (t *Transaction) Inputs() ([]*UTXOInput, error) {
	references := make([]*UTXOInput, 0, len(t.Essence.Inputs))
	for _, input := range t.Essence.Inputs {
		switch castInput := input.(type) {
		case *UTXOInput:
			references = append(references, castInput)
		default:
			return nil, ErrUnknownInputType
		}
	}

	return references, nil
}

func (t *Transaction) ContextInputs() (TransactionContextInputs, error) {
	references := make(TransactionContextInputs, 0, len(t.Essence.ContextInputs))
	for _, input := range t.Essence.ContextInputs {
		switch castInput := input.(type) {
		case *CommitmentInput, *BlockIssuanceCreditInput, *RewardInput:
			references = append(references, castInput)
		default:
			return nil, ErrUnknownContextInputType
		}
	}

	return references, nil
}

func (t *Transaction) BICInputs() ([]*BlockIssuanceCreditInput, error) {
	references := make([]*BlockIssuanceCreditInput, 0, len(t.Essence.ContextInputs))
	for _, input := range t.Essence.ContextInputs {
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

func (t *Transaction) RewardInputs() ([]*RewardInput, error) {
	references := make([]*RewardInput, 0, len(t.Essence.ContextInputs))
	for _, input := range t.Essence.ContextInputs {
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
func (t *Transaction) CommitmentInput() *CommitmentInput {
	for _, input := range t.Essence.ContextInputs {
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

func (t *Transaction) Size() int {
	// PayloadType
	return serializer.UInt32ByteSize +
		t.Essence.Size() +
		t.Unlocks.Size()
}

func (t *Transaction) String() string {
	// TODO: stringify for debugging purposes
	return fmt.Sprintf("Transaction[%v, %v]", t.Essence, t.Unlocks)
}

// syntacticallyValidate syntactically validates the Transaction.
func (t *Transaction) syntacticallyValidate(api API) error {
	// limit unlock block count = input count
	if len(t.Unlocks) != len(t.Essence.Inputs) {
		return ierrors.Errorf("unlock block count must match inputs in essence, %d vs. %d", len(t.Unlocks), len(t.Essence.Inputs))
	}

	if err := t.Essence.syntacticallyValidate(api.ProtocolParameters()); err != nil {
		return ierrors.Errorf("transaction essence is invalid: %w", err)
	}

	if err := ValidateUnlocks(t.Unlocks,
		UnlocksSigUniqueAndRefValidator(api),
	); err != nil {
		return ierrors.Errorf("invalid unlocks: %w", err)
	}

	return nil
}

func (t *Transaction) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
	workScoreEssence, err := t.Essence.WorkScore(workScoreStructure)
	if err != nil {
		return 0, err
	}

	workScoreUnlocks, err := t.Unlocks.WorkScore(workScoreStructure)
	if err != nil {
		return 0, err
	}

	return workScoreEssence.Add(workScoreUnlocks)
}
