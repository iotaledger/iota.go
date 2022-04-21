package iotago

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	// ErrInvalidReceiptMilestoneOpt gets returned when a ReceiptMilestoneOpt is invalid.
	ErrInvalidReceiptMilestoneOpt = errors.New("invalid receipt")
)

const (
	// MinMigratedFundsEntryCount defines the minimum amount of MigratedFundsEntry items within a ReceiptMilestoneOpt.
	MinMigratedFundsEntryCount = 1
	// MaxMigratedFundsEntryCount defines the maximum amount of MigratedFundsEntry items within a ReceiptMilestoneOpt.
	MaxMigratedFundsEntryCount = 127
)

var (
	// ErrReceiptMustContainATreasuryTransaction gets returned if a ReceiptMilestoneOpt does not contain a TreasuryTransaction.
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

// ReceiptMilestoneOpt is a listing of migrated funds.
type ReceiptMilestoneOpt struct {
	// The milestone index at which the funds were migrated in the legacy network.
	MigratedAt uint32
	// Whether this ReceiptMilestoneOpt is the final one for a given migrated at index.
	Final bool
	// The funds which were migrated with this ReceiptMilestoneOpt.
	Funds MigratedFundsEntries
	// The TreasuryTransaction used to fund the funds.
	Transaction *TreasuryTransaction
}

func (r *ReceiptMilestoneOpt) Size() int {
	return serializer.OneByte + serializer.UInt32ByteSize + serializer.OneByte + r.Funds.Size() +
		// payloads have a 4 byte length prefix
		serializer.UInt32ByteSize + r.Transaction.Size()
}

func (r *ReceiptMilestoneOpt) Type() MilestoneOptType {
	return MilestoneOptReceipt
}

func (r *ReceiptMilestoneOpt) Clone() MilestoneOpt {
	return &ReceiptMilestoneOpt{
		MigratedAt:  r.MigratedAt,
		Final:       r.Final,
		Funds:       r.Funds.Clone(),
		Transaction: r.Transaction.Clone(),
	}
}

// SortFunds sorts the funds within the receipt after their serialized binary form in lexical order.
func (r *ReceiptMilestoneOpt) SortFunds() {
	seris := r.Funds.ToSerializables()
	sort.Sort(serializer.SortedSerializables(seris))
	r.Funds.FromSerializables(seris)
}

// Sum returns the sum of all MigratedFundsEntry items within the ReceiptMilestoneOpt.
func (r *ReceiptMilestoneOpt) Sum() uint64 {
	var sum uint64
	for _, item := range r.Funds {
		sum += item.Deposit
	}
	return sum
}

// Treasury returns the TreasuryTransaction within the receipt or nil if none is contained.
// This function panics if the ReceiptMilestoneOpt.Transaction is not nil and not a TreasuryTransaction.
func (r *ReceiptMilestoneOpt) Treasury() *TreasuryTransaction {
	return r.Transaction
}

func (r *ReceiptMilestoneOpt) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(MilestoneOptReceipt), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize receipt milestone option: %w", err)
		}).
		ReadNum(&r.MigratedAt, func(err error) error {
			return fmt.Errorf("unable to deserialize receipt milestone option migrated at index: %w", err)
		}).
		ReadBool(&r.Final, func(err error) error {
			return fmt.Errorf("unable to deserialize receipt milestone option final flag: %w", err)
		}).
		// special as the MigratedFundsEntry has no type denotation byte
		ReadSliceOfObjects(&r.Funds, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationNone, migratedFundEntriesArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize receipt milestone option migrated fund entries: %w", err)
		}).
		ReadPayload(&r.Transaction, deSeriMode, deSeriCtx, receiptPayloadGuard.ReadGuard, func(err error) error {
			return fmt.Errorf("unable to deserialize receipt milestone option transaction: %w", err)
		}).
		WithValidation(deSeriMode, func(_ []byte, err error) error {
			if r.Transaction == nil {
				return ErrReceiptMustContainATreasuryTransaction
			}
			return nil
		}).
		Done()
}

func (r *ReceiptMilestoneOpt) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	if r.Transaction == nil {
		return nil, ErrReceiptMustContainATreasuryTransaction
	}
	return serializer.NewSerializer().
		WriteNum(byte(MilestoneOptReceipt), func(err error) error {
			return fmt.Errorf("unable to serialize receipt milestone option type ID: %w", err)
		}).
		WriteNum(r.MigratedAt, func(err error) error {
			return fmt.Errorf("unable to serialize receipt milestone option migrated at index: %w", err)
		}).
		WriteBool(r.Final, func(err error) error {
			return fmt.Errorf("unable to serialize receipt milestone option final flag: %w", err)
		}).
		WriteSliceOfObjects(&r.Funds, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsUint16, migratedFundEntriesArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize receipt milestone option funds: %w", err)
		}).
		WithValidation(deSeriMode, func(_ []byte, err error) error {
			if r.Transaction == nil {
				return ErrReceiptMustContainATreasuryTransaction
			}
			return nil
		}).
		WritePayload(r.Transaction, deSeriMode, deSeriCtx, receiptPayloadGuard.WriteGuard, func(err error) error {
			return fmt.Errorf("unable to serialize receipt milestone option transaction: %w", err)
		}).
		Serialize()
}

func (r *ReceiptMilestoneOpt) MarshalJSON() ([]byte, error) {
	jReceipt := &jsonReceiptMilestoneOpt{}
	jReceipt.Type = int(MilestoneOptReceipt)
	jReceipt.MigratedAt = int(r.MigratedAt)

	jReceipt.Funds = make([]*json.RawMessage, len(r.Funds))
	for i, migratedFundsEntry := range r.Funds {
		jMigratedFundsEntry, err := migratedFundsEntry.MarshalJSON()
		if err != nil {
			return nil, err
		}
		rawMsgJsonMigratedFundsEntry := json.RawMessage(jMigratedFundsEntry)
		jReceipt.Funds[i] = &rawMsgJsonMigratedFundsEntry
	}

	jTreasuryTransaction, err := r.Transaction.MarshalJSON()
	if err != nil {
		return nil, err
	}
	rawMsgJsonTreasuryTransaction := json.RawMessage(jTreasuryTransaction)
	jReceipt.Transaction = &rawMsgJsonTreasuryTransaction

	jReceipt.Final = r.Final

	return json.Marshal(jReceipt)
}

func (r *ReceiptMilestoneOpt) UnmarshalJSON(bytes []byte) error {
	jReceipt := &jsonReceiptMilestoneOpt{}
	if err := json.Unmarshal(bytes, jReceipt); err != nil {
		return err
	}
	seri, err := jReceipt.ToSerializable()
	if err != nil {
		return err
	}
	*r = *seri.(*ReceiptMilestoneOpt)
	return nil
}

// jsonReceiptMilestoneOpt defines the json representation of a ReceiptMilestoneOpt.
type jsonReceiptMilestoneOpt struct {
	Type        int                `json:"type"`
	MigratedAt  int                `json:"migratedAt"`
	Funds       []*json.RawMessage `json:"funds"`
	Transaction *json.RawMessage   `json:"transaction"`
	Final       bool               `json:"final"`
}

func (j *jsonReceiptMilestoneOpt) ToSerializable() (serializer.Serializable, error) {
	payload := &ReceiptMilestoneOpt{}
	payload.MigratedAt = uint32(j.MigratedAt)

	migratedFundsEntries := make(MigratedFundsEntries, len(j.Funds))
	for i, ele := range j.Funds {
		jMigratedFundsEntry, _ := DeserializeObjectFromJSON(ele, func(ty int) (JSONSerializable, error) {
			return &jsonMigratedFundsEntry{}, nil
		})
		migratedFundsEntry, err := jMigratedFundsEntry.ToSerializable()
		if err != nil {
			return nil, fmt.Errorf("pos %d: %w", i, err)
		}
		migratedFundsEntries[i] = migratedFundsEntry.(*MigratedFundsEntry)
	}
	payload.Funds = migratedFundsEntries

	if j.Transaction == nil {
		return nil, fmt.Errorf("%w: JSON receipt must contain a treasury transaction", ErrInvalidJSON)
	}

	jTreasuryTransaction, _ := DeserializeObjectFromJSON(j.Transaction, func(ty int) (JSONSerializable, error) {
		return &jsonTreasuryTransaction{}, nil
	})

	treasuryTransaction, err := jTreasuryTransaction.ToSerializable()
	if err != nil {
		return nil, err
	}
	payload.Transaction = treasuryTransaction.(*TreasuryTransaction)
	payload.Final = j.Final

	return payload, nil
}

// ValidateReceipt validates whether given the following receipt:
//	- None of the MigratedFundsEntry objects deposits more than the max supply and deposits at least
//	  MinMigratedFundsEntryDeposit tokens.
//	- The sum of all migrated fund entries is not bigger than the total supply.
//	- The previous unspent TreasuryOutput minus the sum of all migrated funds
//    equals the amount of the new TreasuryOutput.
// This function panics if the receipt is nil, the receipt does not include any migrated fund entries or
// the given treasury output is nil.
func ValidateReceipt(receipt *ReceiptMilestoneOpt, prevTreasuryOutput *TreasuryOutput, totalSupply uint64) error {
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
			return fmt.Errorf("%w: tail transaction hash at index %d occurrs multiple times (previous %d)", ErrInvalidReceiptMilestoneOpt, fIndex, prevIndex)
		}
		seenTailTxHashes[entry.TailTransactionHash] = fIndex

		switch {
		case entry.Deposit < MinMigratedFundsEntryDeposit:
			return fmt.Errorf("%w: migrated fund entry at index %d deposits less than %d", ErrInvalidReceiptMilestoneOpt, fIndex, MinMigratedFundsEntryDeposit)
		case entry.Deposit > totalSupply:
			return fmt.Errorf("%w: migrated fund entry at index %d deposits more than total supply", ErrInvalidReceiptMilestoneOpt, fIndex)
		case entry.Deposit+migratedFundsSum > totalSupply:
			// this can't overflow because the previous case ensures that
			return fmt.Errorf("%w: migrated fund entry at index %d overflows total supply", ErrInvalidReceiptMilestoneOpt, fIndex)
		}

		migratedFundsSum += entry.Deposit
	}

	prevTreasury := prevTreasuryOutput.Amount
	newTreasury := treasuryTransaction.Output.Amount
	if prevTreasury-migratedFundsSum != newTreasury {
		return fmt.Errorf("%w: new treasury amount mismatch, prev %d, delta %d (migrated funds), new %d", ErrInvalidReceiptMilestoneOpt, prevTreasury, migratedFundsSum, newTreasury)
	}

	return nil
}
