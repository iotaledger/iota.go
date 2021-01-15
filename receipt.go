package iota

import (
	"encoding/hex"
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
	// Returned if the count of MigratedFundsEntry items is too small.
	ErrMinMigratedFundsEntriesNotReached = fmt.Errorf("min %d migrated fund entries are required within a receipt", MinMigratedFundsEntryCount)
	// Returned if the count of MigratedFundsEntry items is too big.
	ErrMaxMigratedFundsEntriesExceeded = fmt.Errorf("max %d migrated fund entries are allowed within a receipt", MaxMigratedFundsEntryCount)
	// Returned if the MigratedFundsEntry items are not in lexical order when serialized.
	ErrMigratedFundsEntriesOrderViolatesLexicalOrder = errors.New("migrated fund entries must be in their lexical order (byte wise) when serialized")

	migratedFundEntriesArrayRules = &ArrayRules{
		Min:                         MinMigratedFundsEntryCount,
		Max:                         MaxMigratedFundsEntryCount,
		MinErr:                      ErrMinMigratedFundsEntriesNotReached,
		MaxErr:                      ErrMaxMigratedFundsEntriesExceeded,
		ElementBytesLexicalOrder:    true,
		ElementBytesLexicalOrderErr: ErrMigratedFundsEntriesOrderViolatesLexicalOrder,
	}
)

// Receipt is a listing of migrated funds.
type Receipt struct {
	// The milestone index at which the funds were migrated in the legacy network.
	MigratedAt uint32
	// The funds which were migrated with this Receipt.
	Funds Serializables
}

// SortFunds sorts the funds within the receipt after their serialized binary form.
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
		Done()
}

func (r *Receipt) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	var migratedFundsEntriesWrittenConsumer WrittenObjectConsumer
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if migratedFundEntriesArrayRules.ElementBytesLexicalOrder {
			migratedFundEntriesLexicalOrderValidator := migratedFundEntriesArrayRules.LexicalOrderValidator()
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

// MigratedFundsEntry are funds which were migrated from a legacy network.
type MigratedFundsEntry struct {
	// The tail transaction hash of the migration bundle.
	TailTransactionHash [49]byte
	// The target address of the migrated funds.
	Address Serializable
	// The amount of the deposit.
	Deposit uint64
}

func (m *MigratedFundsEntry) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	return NewDeserializer(data).
		ReadArrayOf49Bytes(&m.TailTransactionHash, func(err error) error {
			return fmt.Errorf("unable to deserialize migrated funds entry tail transaction hash: %w", err)
		}).
		ReadObject(func(seri Serializable) { m.Address = seri }, deSeriMode, TypeDenotationByte, AddressSelector, func(err error) error {
			return fmt.Errorf("unable to deserialize address for migrated funds entry: %w", err)
		}).
		ReadNum(&m.Deposit, func(err error) error {
			return fmt.Errorf("unable to deserialize deposit for migrated funds entry: %w", err)
		}).
		Done()
}

func (m *MigratedFundsEntry) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	return NewSerializer().
		WriteBytes(m.TailTransactionHash[:], func(err error) error {
			return fmt.Errorf("unable to serialize migrated funds entry tail transaction hash: %w", err)
		}).
		WriteObject(m.Address, deSeriMode, func(err error) error {
			return fmt.Errorf("unable to serialize migrated funds entry address: %w", err)
		}).
		WriteNum(m.Deposit, func(err error) error {
			return fmt.Errorf("unable to serialize migrated funds entry deposit: %w", err)
		}).
		Serialize()
}

func (m *MigratedFundsEntry) MarshalJSON() ([]byte, error) {
	jsonMigratedFundsEntry := &jsonmigratedfundsentry{}
	jsonMigratedFundsEntry.TailTransactionHash = hex.EncodeToString(m.TailTransactionHash[:])
	addrJsonBytes, err := m.Address.MarshalJSON()
	if err != nil {
		return nil, err
	}
	jsonRawMsgAddr := json.RawMessage(addrJsonBytes)
	jsonMigratedFundsEntry.Address = &jsonRawMsgAddr
	jsonMigratedFundsEntry.Deposit = int(m.Deposit)

	return json.Marshal(jsonMigratedFundsEntry)
}

func (m *MigratedFundsEntry) UnmarshalJSON(bytes []byte) error {
	jsonMigratedFundsEntry := &jsonmigratedfundsentry{}
	if err := json.Unmarshal(bytes, jsonMigratedFundsEntry); err != nil {
		return err
	}
	seri, err := jsonMigratedFundsEntry.ToSerializable()
	if err != nil {
		return err
	}
	*m = *seri.(*MigratedFundsEntry)
	return nil
}

// jsonmigratedfundsentry defines the json representation of a MigratedFundsEntry.
type jsonmigratedfundsentry struct {
	TailTransactionHash string           `json:"tailTransactionHash"`
	Address             *json.RawMessage `json:"address"`
	Deposit             int              `json:"deposit"`
}

func (j *jsonmigratedfundsentry) ToSerializable() (Serializable, error) {
	payload := &MigratedFundsEntry{}
	tailTransactionHash, err := hex.DecodeString(j.TailTransactionHash)
	if err != nil {
		return nil, fmt.Errorf("can't decode tail transaction hash for migrated funds entry from JSON: %w", err)
	}
	copy(payload.TailTransactionHash[:], tailTransactionHash)
	payload.Deposit = uint64(j.Deposit)
	jsonAddr, err := DeserializeObjectFromJSON(j.Address, jsonaddressselector)
	if err != nil {
		return nil, fmt.Errorf("can't decode address type from JSON: %w", err)
	}

	payload.Address, err = jsonAddr.ToSerializable()
	if err != nil {
		return nil, err
	}
	return payload, nil
}
