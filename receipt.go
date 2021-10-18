package iotago

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/iotaledger/hive.go/serializer"
)

const (
	// ReceiptPayloadTypeID defines the Receipt payload's ID.
	ReceiptPayloadTypeID uint32 = 3
	// MinMigratedFundsEntryCount defines the minimum amount of MigratedFundsEntry items within a Receipt.
	MinMigratedFundsEntryCount = 1
	// MaxMigratedFundsEntryCount defines the maximum amount of MigratedFundsEntry items within a Receipt.
	MaxMigratedFundsEntryCount = 127
)

var (
	// ErrReceiptMustContainATreasuryTransaction gets returned if a Receipt does not contain a TreasuryTransaction.
	ErrReceiptMustContainATreasuryTransaction = errors.New("receipt must contain a treasury transaction")

	migratedFundEntriesArrayRules = &serializer.ArrayRules{
		Min:            MinMigratedFundsEntryCount,
		Max:            MaxMigratedFundsEntryCount,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates | serializer.ArrayValidationModeLexicalOrdering,
	}
)

// Receipt is a listing of migrated funds.
type Receipt struct {
	// The milestone index at which the funds were migrated in the legacy network.
	MigratedAt uint32
	// Whether this Receipt is the final one for a given migrated at index.
	Final bool
	// The funds which were migrated with this Receipt.
	Funds serializer.Serializables
	// The TreasuryTransaction used to fund the funds.
	Transaction serializer.Serializable
}

// SortFunds sorts the funds within the receipt after their serialized binary form in lexical order.
func (r *Receipt) SortFunds() {
	sort.Sort(serializer.SortedSerializables(r.Funds))
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

func (r *Receipt) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	return serializer.NewDeserializer(data).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := serializer.CheckType(data, ReceiptPayloadTypeID); err != nil {
					return fmt.Errorf("unable to deserialize receipt: %w", err)
				}
			}
			return nil
		}).
		Skip(serializer.TypeDenotationByteSize, func(err error) error {
			return fmt.Errorf("unable to skip receipt payload ID during deserialization: %w", err)
		}).
		ReadNum(&r.MigratedAt, func(err error) error {
			return fmt.Errorf("unable to deserialize receipt migrated index: %w", err)
		}).
		ReadBool(&r.Final, func(err error) error {
			return fmt.Errorf("unable to deserialize receipt final flag: %w", err)
		}).
		// special as the MigratedFundsEntry has no type denotation byte
		ReadSliceOfObjects(func(seri serializer.Serializables) { r.Funds = seri }, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationNone, func(_ uint32) (serializer.Serializable, error) {
			// there is no real selector, so we always return a fresh MigratedFundsEntry
			return &MigratedFundsEntry{}, nil
		}, migratedFundEntriesArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize receipt migrated fund entries: %w", err)
		}).
		ReadPayload(func(seri serializer.Serializable) { r.Transaction = seri }, deSeriMode, func(ty uint32) (serializer.Serializable, error) {
			if ty != TreasuryTransactionPayloadTypeID {
				return nil, fmt.Errorf("a receipt can only contain a treasury transaction but got type ID %d:  %w", ty, ErrUnknownPayloadType)
			}
			return PayloadSelector(ty)
		}, func(err error) error {
			return fmt.Errorf("unable to deserialize receipt transaction: %w", err)
		}).
		AbortIf(func(err error) error {
			if r.Transaction == nil {
				return ErrReceiptMustContainATreasuryTransaction
			}
			return nil
		}).
		Done()
}

func (r *Receipt) Serialize(deSeriMode serializer.DeSerializationMode) ([]byte, error) {
	if r.Transaction == nil {
		return nil, ErrReceiptMustContainATreasuryTransaction
	}
	var migratedFundsEntriesWrittenConsumer serializer.WrittenObjectConsumer
	if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
		if migratedFundEntriesArrayRules.ValidationMode.HasMode(serializer.ArrayValidationModeLexicalOrdering) {
			migratedFundEntriesLexicalOrderValidator := migratedFundEntriesArrayRules.LexicalOrderWithoutDupsValidator()
			migratedFundsEntriesWrittenConsumer = func(index int, written []byte) error {
				if err := migratedFundEntriesLexicalOrderValidator(index, written); err != nil {
					return fmt.Errorf("%w: unable to serialize migrated fund entries of receipt since they are not in lexical order", err)
				}
				return nil
			}
		}
	}
	return serializer.NewSerializer().
		Do(func() {
			if deSeriMode.HasMode(serializer.DeSeriModePerformLexicalOrdering) {
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
		WriteSliceOfObjects(r.Funds, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, migratedFundsEntriesWrittenConsumer, func(err error) error {
			return fmt.Errorf("unable to serialize receipt funds: %w", err)
		}).
		WritePayload(r.Transaction, deSeriMode, func(err error) error {
			return fmt.Errorf("unable to serialize receipt transaction: %w", err)
		}).
		Serialize()
}

func (r *Receipt) MarshalJSON() ([]byte, error) {
	jReceipt := &jsonReceipt{}
	jReceipt.Type = int(ReceiptPayloadTypeID)
	jReceipt.MigratedAt = int(r.MigratedAt)

	jReceipt.Funds = make([]*json.RawMessage, len(r.Funds))
	for i, migratedFundsEntry := range r.Funds {
		jsonMigratedFundsEntry, err := migratedFundsEntry.MarshalJSON()
		if err != nil {
			return nil, err
		}
		rawMsgJsonMigratedFundsEntry := json.RawMessage(jsonMigratedFundsEntry)
		jReceipt.Funds[i] = &rawMsgJsonMigratedFundsEntry
	}

	jsonTreasuryTransaction, err := r.Transaction.MarshalJSON()
	if err != nil {
		return nil, err
	}
	rawMsgJsonTreasuryTransaction := json.RawMessage(jsonTreasuryTransaction)
	jReceipt.Transaction = &rawMsgJsonTreasuryTransaction

	jReceipt.Final = r.Final

	return json.Marshal(jReceipt)
}

func (r *Receipt) UnmarshalJSON(bytes []byte) error {
	jReceipt := &jsonReceipt{}
	if err := json.Unmarshal(bytes, jReceipt); err != nil {
		return err
	}
	seri, err := jReceipt.ToSerializable()
	if err != nil {
		return err
	}
	*r = *seri.(*Receipt)
	return nil
}

// jsonReceipt defines the json representation of a Receipt.
type jsonReceipt struct {
	Type        int                `json:"type"`
	MigratedAt  int                `json:"migratedAt"`
	Funds       []*json.RawMessage `json:"funds"`
	Transaction *json.RawMessage   `json:"transaction"`
	Final       bool               `json:"final"`
}

func (j *jsonReceipt) ToSerializable() (serializer.Serializable, error) {
	payload := &Receipt{}
	payload.MigratedAt = uint32(j.MigratedAt)

	migratedFundsEntries := make(serializer.Serializables, len(j.Funds))
	for i, ele := range j.Funds {
		jsonMigratedFundsEntry, _ := DeserializeObjectFromJSON(ele, func(ty int) (JSONSerializable, error) {
			return &jsonMigratedFundsEntry{}, nil
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
		return &jsonTreasuryTransaction{}, nil
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
	// ErrInvalidReceipt gets returned when a receipt is invalid.
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
