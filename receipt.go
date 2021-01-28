package iota

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
	// The funds which were migrated with this Receipt.
	Funds Serializables
	// The TreasuryTransaction used to fund the funds. Might be nil.
	// A non nil Receipt.Transaction field indicates that this receipt
	// is the last one for the given MigratedAt milestone index.
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
		Done()
}

func (r *Receipt) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
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
	jsonReceiptPayload.Type = int(MilestonePayloadTypeID)
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
	Type       int                `json:"type"`
	MigratedAt int                `json:"migratedAt"`
	Funds      []*json.RawMessage `json:"funds"`
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
	return payload, nil
}

var (
	// Returned when a set of receipts are invalid.
	ErrInvalidReceiptsSet = errors.New("invalid receipt set")
)

// ValidateReceipts validates whether given the following receipts:
//	- They all share the same Receipt.MigratedAt index
//	- All MigratedFundsEntry objects are unique.
//	- None of the MigratedFundsEntry objects deposits more than the max supply or zero and minimum
//	  MinMigratedFundsEntryDeposit tokens.
//	- The sum of all migrated fund entries is not bigger than the total supply.
//	- There is at most one which includes a TreasuryTransaction.
//	- The previous unspent TreasuryOutput minus the sum of all migrated funds
//    equals the amount of the new TreasuryOutput.
// This function panics if the receipt slice is nil or empty and if any receipt
// does not include any migrated fund entries. The order of the receipts does not matter.
func ValidateReceipts(receipts []*Receipt, prevTreasuryOutput *TreasuryOutput) error {
	switch {
	case receipts == nil || len(receipts) == 0:
		panic("no receipts passed for validation")
	case prevTreasuryOutput == nil:
		panic("given previous treasury output is nil")
	}

	var migratedAt uint32
	var migratedFundsSum uint64
	var treasuryTransaction *TreasuryTransaction
	type tailloc struct {
		receipt int
		index   int
	}
	seenTailTxHashes := make(map[LegacyTailTransactionHash]tailloc)
	for rIndex, r := range receipts {
		if tt := r.Treasury(); tt != nil {
			if treasuryTransaction != nil {
				return fmt.Errorf("%w: only one receipt can contain a treasury transaction", ErrInvalidReceiptsSet)
			}
			treasuryTransaction = tt
		}

		if migratedAt == 0 {
			migratedAt = r.MigratedAt
		}

		if r.MigratedAt != migratedAt {
			return fmt.Errorf("%w: the migrated at index must be the same across all receipts", ErrInvalidReceiptsSet)
		}

		if r.Funds == nil || len(r.Funds) == 0 {
			panic("receipt includes not migrated funds")
		}

		for fIndex, f := range r.Funds {
			entry := f.(*MigratedFundsEntry)
			if tailLoc, seen := seenTailTxHashes[entry.TailTransactionHash]; seen {
				return fmt.Errorf("%w: same legacy tail transaction occurs multiple times, seen in receipt %d (index %d) and receipt %d (index %d)", ErrInvalidReceiptsSet, tailLoc.receipt, tailLoc.index, rIndex, fIndex)
			}
			seenTailTxHashes[entry.TailTransactionHash] = tailloc{rIndex, fIndex}
			switch {
			case entry.Deposit == 0:
				return fmt.Errorf("%w: migrated fund entry at receipt %d (index %d) deposits zero", ErrInvalidReceiptsSet, rIndex, fIndex)
			case entry.Deposit < MinMigratedFundsEntryDeposit:
				return fmt.Errorf("%w: migrated fund entry at receipt %d (index %d) deposits less than %d", ErrInvalidReceiptsSet, rIndex, fIndex, MinMigratedFundsEntryDeposit)
			case entry.Deposit > TokenSupply:
				return fmt.Errorf("%w: migrated fund entry at receipt %d (index %d) deposits more than total supply", ErrInvalidReceiptsSet, rIndex, fIndex)
			case entry.Deposit+migratedFundsSum > TokenSupply:
				return fmt.Errorf("%w: migrated fund entry at receipt %d (index %d) deposits overflows total supply", ErrInvalidReceiptsSet, rIndex, fIndex)
			}

			migratedFundsSum += entry.Deposit
		}
	}

	if treasuryTransaction == nil {
		return fmt.Errorf("%w: no treasury transaction was included in any receipt", ErrInvalidReceiptsSet)
	}

	prevTreasury := prevTreasuryOutput.Amount
	newTreasury := treasuryTransaction.Output.(*TreasuryOutput).Amount
	if prevTreasury-migratedFundsSum != newTreasury {
		return fmt.Errorf("%w: new treasury amount mismatch, prev %d, delta %d (migrated funds), new %d", ErrInvalidReceiptsSet, prevTreasury, migratedFundsSum, newTreasury)
	}

	return nil
}
