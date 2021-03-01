package iotago

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
)

const (
	// Defines the minimum amount a MigratedFundEntry must deposit.
	MinMigratedFundsEntryDeposit = 1_000_000
)

// LegacyTailTransactionHash represents the bytes of a T5B1 encoded legacy tail transaction hash.
type LegacyTailTransactionHash = [49]byte

// MigratedFundsEntry are funds which were migrated from a legacy network.
type MigratedFundsEntry struct {
	// The tail transaction hash of the migration bundle.
	TailTransactionHash LegacyTailTransactionHash
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
