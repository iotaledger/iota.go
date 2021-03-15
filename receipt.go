package iotago

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

const (
	// Defines the Receipt payload's ID.
	ReceiptPayloadTypeID uint32 = 3
	// Defines the minimum amount of MigratedFundsEntry items within a Receipt.
	MinMigratedFundsEntryCount = 1
	// Defines the maximum amount of MigratedFundsEntry items within a Receipt.
	MaxMigratedFundsEntryCount = 127
)

var (
	// Returned if a Receipt does not contain a TreasuryTransaction.
	ErrReceiptMustContainATreasuryTransaction = errors.New("receipt must contain a treasury transaction")

	migratedFundEntriesArrayRules = &ArrayRules{
		Min:            MinMigratedFundsEntryCount,
		Max:            MaxMigratedFundsEntryCount,
		ValidationMode: ArrayValidationModeNoDuplicates | ArrayValidationModeLexicalOrdering,
	}
)

// Receipt is a listing of migrated funds.
type Receipt struct {
	// The milestone index at which the funds were migrated in the legacy network.
	MigratedAt uint32
	// Whether this Receipt is the final one for a given migrated at index.
	Final bool
	// The funds which were migrated with this Receipt.
	Funds Serializables
	// The TreasuryTransaction used to fund the funds.
	Transaction Serializable
}

// SortFunds sorts the funds within the receipt after their serialized binary form in lexical order.
func (r *Receipt) SortFunds() {
	sort.Sort(SortedSerializables(r.Funds))
}

// Sum returns the sum of all MigratedFundsEntry items within the Receipt.
func (r *Receipt) Sum() uint64 {
	var sum uint64
	for _, item := range r.Funds {
		migrateFundEntry := item.(*MigratedFundsEntry)
		sum += migrateFundEntry.Deposit
	}
	return sum
}

// Treasury returns the TreasuryTransaction within the receipt or nil if none is contained.
// This function panics if the Receipt.Transaction is not nil and not a TreasuryTransaction.
func (r *Receipt) Treasury() *TreasuryTransaction {
	if r.Transaction == nil {
		return nil
	}
	t, ok := r.Transaction.(*TreasuryTransaction)
	if !ok {
		panic("receipt contains non treasury transaction")
	}
	return t
}

func (r *Receipt) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	return NewDeserializer(data).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(DeSeriModePerformValidation) {
				if err := checkType(data, ReceiptPayloadTypeID); err != nil {
					return fmt.Errorf("unable to deserialize receipt: %w", err)
				}
			}
			return nil
		}).
		Skip(TypeDenotationByteSize, func(err error) error {
			return fmt.Errorf("unable to skip receipt payload ID during deserialization: %w", err)
		}).
		ReadNum(&r.MigratedAt, func(err error) error {
			return fmt.Errorf("unable to deserialize receipt migrated index: %w", err)
		}).
		ReadBool(&r.Final, func(err error) error {
			return fmt.Errorf("unable to deserialize receipt final flag: %w", err)
		}).
		// special as the MigratedFundsEntry has no type denotation byte
		ReadSliceOfObjects(func(seri Serializables) { r.Funds = seri }, deSeriMode, TypeDenotationNone, func(_ uint32) (Serializable, error) {
			// there is no real selector, so we always return a fresh MigratedFundsEntry
			return &MigratedFundsEntry{}, nil
		}, migratedFundEntriesArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize receipt migrated fund entries: %w", err)
		}).
		ReadPayload(func(seri Serializable) { r.Transaction = seri }, deSeriMode, func(err error) error {
			return fmt.Errorf("unable to deserialize receipt transaction: %w", err)
		}, func(ty uint32) (Serializable, error) {
			if ty != TreasuryTransactionPayloadTypeID {
				return nil, fmt.Errorf("a receipt can only contain a treasury transaction but got type ID %d:  %w", ty, ErrUnknownPayloadType)
			}
			return PayloadSelector(ty)
		}).
		AbortIf(func(err error) error {
			if r.Transaction == nil {
				return ErrReceiptMustContainATreasuryTransaction
			}
			return nil
		}).
		Done()
}

func (r *Receipt) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	if r.Transaction == nil {
		return nil, ErrReceiptMustContainATreasuryTransaction
	}
	var migratedFundsEntriesWrittenConsumer WrittenObjectConsumer
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if migratedFundEntriesArrayRules.ValidationMode.HasMode(ArrayValidationModeLexicalOrdering) {
			migratedFundEntriesLexicalOrderValidator := migratedFundEntriesArrayRules.LexicalOrderWithoutDupsValidator()
			migratedFundsEntriesWrittenConsumer = func(index int, written []byte) error {
				if err := migratedFundEntriesLexicalOrderValidator(index, written); err != nil {
					return fmt.Errorf("%w: unable to serialize migrated fund entries of receipt since they are not in lexical order", err)
				}
				return nil
			}
		}
	}
	return NewSerializer().
		Do(func() {
			if deSeriMode.HasMode(DeSeriModePerformLexicalOrdering) {
				r.SortFunds()
			}
		}).
		WriteNum(ReceiptPayloadTypeID, func(err error) error {
			return fmt.Errorf("unable to serialize receipt payload ID: %w", err)
		}).
		WriteNum(r.MigratedAt, func(err error) error {
			return fmt.Errorf("unable to serialize receipt payload ID: %w", err)
		}).
		WriteBool(r.Final, func(err error) error {
			return fmt.Errorf("unable to serialize receipt final flag: %w", err)
		}).
		WriteSliceOfObjects(r.Funds, deSeriMode, migratedFundsEntriesWrittenConsumer, func(err error) error {
			return fmt.Errorf("unable to serialize receipt funds: %w", err)
		}).
		WritePayload(r.Transaction, deSeriMode, func(err error) error {
			return fmt.Errorf("unable to serialize receipt transaction: %w", err)
		}).
		Serialize()
}

func (r *Receipt) MarshalJSON() ([]byte, error) {
	jsonReceiptPayload := &jsonreceiptpayload{}
	jsonReceiptPayload.Type = int(ReceiptPayloadTypeID)
	jsonReceiptPayload.MigratedAt = int(r.MigratedAt)

	jsonReceiptPayload.Funds = make([]*json.RawMessage, len(r.Funds))
	for i, migratedFundsEntry := range r.Funds {
		jsonMigratedFundsEntry, err := migratedFundsEntry.MarshalJSON()
		if err != nil {
			return nil, err
		}
		rawMsgJsonMigratedFundsEntry := json.RawMessage(jsonMigratedFundsEntry)
		jsonReceiptPayload.Funds[i] = &rawMsgJsonMigratedFundsEntry
	}

	jsonTreasuryTransaction, err := r.Transaction.MarshalJSON()
	if err != nil {
		return nil, err
	}
	rawMsgJsonTreasuryTransaction := json.RawMessage(jsonTreasuryTransaction)
	jsonReceiptPayload.Transaction = &rawMsgJsonTreasuryTransaction

	jsonReceiptPayload.Final = r.Final

	return json.Marshal(jsonReceiptPayload)
}

func (r *Receipt) UnmarshalJSON(bytes []byte) error {
	jsonReceiptPayload := &jsonreceiptpayload{}
	if err := json.Unmarshal(bytes, jsonReceiptPayload); err != nil {
		return err
	}
	seri, err := jsonReceiptPayload.ToSerializable()
	if err != nil {
		return err
	}
	*r = *seri.(*Receipt)
	return nil
}

// jsonreceiptpayload defines the json representation of a Receipt.
type jsonreceiptpayload struct {
	Type        int                `json:"type"`
	MigratedAt  int                `json:"migratedAt"`
	Funds       []*json.RawMessage `json:"funds"`
	Transaction *json.RawMessage   `json:"transaction"`
	Final       bool               `json:"final"`
}

func (j *jsonreceiptpayload) ToSerializable() (Serializable, error) {
	payload := &Receipt{}
	payload.MigratedAt = uint32(j.MigratedAt)

	migratedFundsEntries := make(Serializables, len(j.Funds))
	for i, ele := range j.Funds {
		jsonMigratedFundsEntry, _ := DeserializeObjectFromJSON(ele, func(ty int) (JSONSerializable, error) {
			return &jsonmigratedfundsentry{}, nil
		})
		migratedFundsEntry, err := jsonMigratedFundsEntry.ToSerializable()
		if err != nil {
			return nil, fmt.Errorf("pos %d: %w", i, err)
		}
		migratedFundsEntries[i] = migratedFundsEntry
	}
	payload.Funds = migratedFundsEntries

	if j.Transaction == nil {
		return nil, fmt.Errorf("%w: JSON receipt must contain a treasury transaction", ErrInvalidJSON)
	}

	jsonTreasuryTransaction, _ := DeserializeObjectFromJSON(j.Transaction, func(ty int) (JSONSerializable, error) {
		return &jsontreasurytransaction{}, nil
	})

	treasuryTransaction, err := jsonTreasuryTransaction.ToSerializable()
	if err != nil {
		return nil, err
	}
	payload.Transaction = treasuryTransaction
	payload.Final = j.Final

	return payload, nil
}

var (
	// Returned when a receipt is invalid
	ErrInvalidReceipt = errors.New("invalid receipt")
)

// ValidateReceipt validates whether given the following receipt:
//	- None of the MigratedFundsEntry objects deposits more than the max supply and deposits at least
//	  MinMigratedFundsEntryDeposit tokens.
//	- The sum of all migrated fund entries is not bigger than the total supply.
//	- The previous unspent TreasuryOutput minus the sum of all migrated funds
//    equals the amount of the new TreasuryOutput.
// This function panics if the receipt is nil, the receipt does not include any migrated fund entries or
// the given treasury output is nil.
func ValidateReceipt(receipt *Receipt, prevTreasuryOutput *TreasuryOutput) error {
	switch {
	case prevTreasuryOutput == nil:
		panic("given previous treasury output is nil")
	}

	treasuryTransaction := receipt.Treasury()
	if treasuryTransaction == nil {
		return ErrReceiptMustContainATreasuryTransaction
	}

	if receipt.Funds == nil || len(receipt.Funds) == 0 {
		panic("receipt has no migrated funds")
	}

	seenTailTxHashes := make(map[LegacyTailTransactionHash]int)
	var migratedFundsSum uint64
	for fIndex, f := range receipt.Funds {
		entry := f.(*MigratedFundsEntry)
		if prevIndex, seen := seenTailTxHashes[entry.TailTransactionHash]; seen {
			return fmt.Errorf("%w: tail transaction hash at index %d occurrs multiple times (previous %d)", ErrInvalidReceipt, fIndex, prevIndex)
		}
		seenTailTxHashes[entry.TailTransactionHash] = fIndex

		switch {
		case entry.Deposit < MinMigratedFundsEntryDeposit:
			return fmt.Errorf("%w: migrated fund entry at index %d deposits less than %d", ErrInvalidReceipt, fIndex, MinMigratedFundsEntryDeposit)
		case entry.Deposit > TokenSupply:
			return fmt.Errorf("%w: migrated fund entry at index %d deposits more than total supply", ErrInvalidReceipt, fIndex)
		case entry.Deposit+migratedFundsSum > TokenSupply:
			// this can't overflow because the previous case ensures that
			return fmt.Errorf("%w: migrated fund entry at index %d overflows total supply", ErrInvalidReceipt, fIndex)
		}

		migratedFundsSum += entry.Deposit
	}

	prevTreasury := prevTreasuryOutput.Amount
	newTreasury := treasuryTransaction.Output.(*TreasuryOutput).Amount
	if prevTreasury-migratedFundsSum != newTreasury {
		return fmt.Errorf("%w: new treasury amount mismatch, prev %d, delta %d (migrated funds), new %d", ErrInvalidReceipt, prevTreasury, migratedFundsSum, newTreasury)
	}

	return nil
}
