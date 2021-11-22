package iotago

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/iotaledger/hive.go/serializer"
)

const (
	// MinMigratedFundsEntryCount defines the minimum amount of MigratedFundsEntry items within a Receipt.
	MinMigratedFundsEntryCount = 1
	// MaxMigratedFundsEntryCount defines the maximum amount of MigratedFundsEntry items within a Receipt.
	MaxMigratedFundsEntryCount = 127
)

var (
	// ErrReceiptMustContainATreasuryTransaction gets returned if a Receipt does not contain a TreasuryTransaction.
	ErrReceiptMustContainATreasuryTransaction = errors.New("receipt must contain a treasury transaction")

	receiptPayloadGuard = serializer.SerializableGuard{
		ReadGuard: func(ty uint32) (serializer.Serializable, error) {
			if PayloadType(ty) != PayloadTreasuryTransaction {
				return nil, ErrTypeIsNotSupportedPayload
			}
			return PayloadSelector(ty)
		},
		WriteGuard: func(seri serializer.Serializable) error {
			if seri == nil {
				return ErrReceiptMustContainATreasuryTransaction
			}
			if _, is := seri.(*TreasuryTransaction); !is {
				return ErrTypeIsNotSupportedPayload
			}
			return nil
		},
	}

	migratedFundEntriesArrayRules = &serializer.ArrayRules{
		Min: MinMigratedFundsEntryCount,
		Max: MaxMigratedFundsEntryCount,
		Guards: serializer.SerializableGuard{
			ReadGuard:  func(_ uint32) (serializer.Serializable, error) { return &MigratedFundsEntry{}, nil },
			WriteGuard: nil,
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates | serializer.ArrayValidationModeLexicalOrdering,
	}
)

// MigratedFundEntriesArrayRules returns array rules defining the constraints of a slice of MigratedFundsEntry.
func MigratedFundEntriesArrayRules() serializer.ArrayRules {
	return *migratedFundEntriesArrayRules
}

// Receipt is a listing of migrated funds.
type Receipt struct {
	// The milestone index at which the funds were migrated in the legacy network.
	MigratedAt uint32
	// Whether this Receipt is the final one for a given migrated at index.
	Final bool
	// The funds which were migrated with this Receipt.
	Funds MigratedFundsEntries
	// The TreasuryTransaction used to fund the funds.
	Transaction *TreasuryTransaction
}

func (r *Receipt) PayloadType() PayloadType {
	return PayloadReceipt
}

// SortFunds sorts the funds within the receipt after their serialized binary form in lexical order.
func (r *Receipt) SortFunds() {
	seris := r.Funds.ToSerializables()
	sort.Sort(serializer.SortedSerializables(seris))
	r.Funds.FromSerializables(seris)
}

// Sum returns the sum of all MigratedFundsEntry items within the Receipt.
func (r *Receipt) Sum() uint64 {
	var sum uint64
	for _, item := range r.Funds {
		sum += item.Deposit
	}
	return sum
}

// Treasury returns the TreasuryTransaction within the receipt or nil if none is contained.
// This function panics if the Receipt.Transaction is not nil and not a TreasuryTransaction.
func (r *Receipt) Treasury() *TreasuryTransaction {
	return r.Transaction
}

func (r *Receipt) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(PayloadReceipt), serializer.TypeDenotationUint32, func(err error) error {
			return fmt.Errorf("unable to deserialize receipt: %w", err)
		}).
		ReadNum(&r.MigratedAt, func(err error) error {
			return fmt.Errorf("unable to deserialize receipt migrated index: %w", err)
		}).
		ReadBool(&r.Final, func(err error) error {
			return fmt.Errorf("unable to deserialize receipt final flag: %w", err)
		}).
		// special as the MigratedFundsEntry has no type denotation byte
		ReadSliceOfObjects(&r.Funds, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationNone, migratedFundEntriesArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize receipt migrated fund entries: %w", err)
		}).
		ReadPayload(&r.Transaction, deSeriMode, deSeriCtx, receiptPayloadGuard.ReadGuard, func(err error) error {
			return fmt.Errorf("unable to deserialize receipt transaction: %w", err)
		}).
		WithValidation(deSeriMode, func(_ []byte, err error) error {
			if r.Transaction == nil {
				return ErrReceiptMustContainATreasuryTransaction
			}
			return nil
		}).
		Done()
}

func (r *Receipt) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	if r.Transaction == nil {
		return nil, ErrReceiptMustContainATreasuryTransaction
	}
	return serializer.NewSerializer().
		WriteNum(PayloadReceipt, func(err error) error {
			return fmt.Errorf("unable to serialize receipt payload ID: %w", err)
		}).
		WriteNum(r.MigratedAt, func(err error) error {
			return fmt.Errorf("unable to serialize receipt payload ID: %w", err)
		}).
		WriteBool(r.Final, func(err error) error {
			return fmt.Errorf("unable to serialize receipt final flag: %w", err)
		}).
		WriteSliceOfObjects(&r.Funds, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsUint16, migratedFundEntriesArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize receipt funds: %w", err)
		}).
		WithValidation(deSeriMode, func(_ []byte, err error) error {
			if r.Transaction == nil {
				return ErrReceiptMustContainATreasuryTransaction
			}
			return nil
		}).
		WritePayload(r.Transaction, deSeriMode, deSeriCtx, receiptPayloadGuard.WriteGuard, func(err error) error {
			return fmt.Errorf("unable to serialize receipt transaction: %w", err)
		}).
		Serialize()
}

func (r *Receipt) MarshalJSON() ([]byte, error) {
	jReceipt := &jsonReceipt{}
	jReceipt.Type = int(PayloadReceipt)
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

	migratedFundsEntries := make(MigratedFundsEntries, len(j.Funds))
	for i, ele := range j.Funds {
		jsonMigratedFundsEntry, _ := DeserializeObjectFromJSON(ele, func(ty int) (JSONSerializable, error) {
			return &jsonMigratedFundsEntry{}, nil
		})
		migratedFundsEntry, err := jsonMigratedFundsEntry.ToSerializable()
		if err != nil {
			return nil, fmt.Errorf("pos %d: %w", i, err)
		}
		migratedFundsEntries[i] = migratedFundsEntry.(*MigratedFundsEntry)
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
	payload.Transaction = treasuryTransaction.(*TreasuryTransaction)
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
	for fIndex, entry := range receipt.Funds {
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
	newTreasury := treasuryTransaction.Output.Amount
	if prevTreasury-migratedFundsSum != newTreasury {
		return fmt.Errorf("%w: new treasury amount mismatch, prev %d, delta %d (migrated funds), new %d", ErrInvalidReceipt, prevTreasury, migratedFundsSum, newTreasury)
	}

	return nil
}
