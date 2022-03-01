package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	migratedFundEntryFeatBlockAddrGuard = serializer.SerializableGuard{
		ReadGuard:  addrReadGuard(AddressTypeSet{AddressEd25519: struct{}{}}),
		WriteGuard: addrWriteGuard(AddressTypeSet{AddressEd25519: struct{}{}}),
	}
)

const (
	// MinMigratedFundsEntryDeposit defines the minimum amount a MigratedFundsEntry must deposit.
	MinMigratedFundsEntryDeposit = 1_000_000
)

// LegacyTailTransactionHash represents the bytes of a T5B1 encoded legacy tail transaction hash.
type LegacyTailTransactionHash = [49]byte

// MigratedFundsEntries is a slice of MigratedFundsEntry.
type MigratedFundsEntries []*MigratedFundsEntry

func (o MigratedFundsEntries) ToSerializables() serializer.Serializables {
	seris := make(serializer.Serializables, len(o))
	for i, x := range o {
		seris[i] = x
	}
	return seris
}

func (o *MigratedFundsEntries) FromSerializables(seris serializer.Serializables) {
	*o = make(MigratedFundsEntries, len(seris))
	for i, seri := range seris {
		(*o)[i] = seri.(*MigratedFundsEntry)
	}
}

// MigratedFundsEntry are funds which were migrated from a legacy network.
type MigratedFundsEntry struct {
	// The tail transaction hash of the migration bundle.
	TailTransactionHash LegacyTailTransactionHash
	// The target address of the migrated funds.
	Address Address
	// The amount of the deposit.
	Deposit uint64
}

func (m *MigratedFundsEntry) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		ReadArrayOf49Bytes(&m.TailTransactionHash, func(err error) error {
			return fmt.Errorf("unable to deserialize migrated funds entry tail transaction hash: %w", err)
		}).
		ReadObject(&m.Address, deSeriMode, deSeriCtx, serializer.TypeDenotationByte, migratedFundEntryFeatBlockAddrGuard.ReadGuard, func(err error) error {
			return fmt.Errorf("unable to deserialize address for migrated funds entry: %w", err)
		}).
		ReadNum(&m.Deposit, func(err error) error {
			return fmt.Errorf("unable to deserialize deposit for migrated funds entry: %w", err)
		}).
		Done()
}

func (m *MigratedFundsEntry) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteBytes(m.TailTransactionHash[:], func(err error) error {
			return fmt.Errorf("unable to serialize migrated funds entry tail transaction hash: %w", err)
		}).
		WriteObject(m.Address, deSeriMode, deSeriCtx, migratedFundEntryFeatBlockAddrGuard.WriteGuard, func(err error) error {
			return fmt.Errorf("unable to serialize migrated funds entry address: %w", err)
		}).
		WriteNum(m.Deposit, func(err error) error {
			return fmt.Errorf("unable to serialize migrated funds entry deposit: %w", err)
		}).
		Serialize()
}

func (m *MigratedFundsEntry) MarshalJSON() ([]byte, error) {
	jMigratedFundsEntry := &jsonMigratedFundsEntry{}
	jMigratedFundsEntry.TailTransactionHash = EncodeHex(m.TailTransactionHash[:])
	addrJsonBytes, err := m.Address.MarshalJSON()
	if err != nil {
		return nil, err
	}
	jsonRawMsgAddr := json.RawMessage(addrJsonBytes)
	jMigratedFundsEntry.Address = &jsonRawMsgAddr
	jMigratedFundsEntry.Deposit = int(m.Deposit)

	return json.Marshal(jMigratedFundsEntry)
}

func (m *MigratedFundsEntry) UnmarshalJSON(bytes []byte) error {
	jMigratedFundsEntry := &jsonMigratedFundsEntry{}
	if err := json.Unmarshal(bytes, jMigratedFundsEntry); err != nil {
		return err
	}
	seri, err := jMigratedFundsEntry.ToSerializable()
	if err != nil {
		return err
	}
	*m = *seri.(*MigratedFundsEntry)
	return nil
}

// jsonMigratedFundsEntry defines the json representation of a MigratedFundsEntry.
type jsonMigratedFundsEntry struct {
	TailTransactionHash string           `json:"tailTransactionHash"`
	Address             *json.RawMessage `json:"address"`
	Deposit             int              `json:"deposit"`
}

func (j *jsonMigratedFundsEntry) ToSerializable() (serializer.Serializable, error) {
	payload := &MigratedFundsEntry{}
	tailTransactionHash, err := DecodeHex(j.TailTransactionHash)
	if err != nil {
		return nil, fmt.Errorf("can't decode tail transaction hash for migrated funds entry from JSON: %w", err)
	}
	copy(payload.TailTransactionHash[:], tailTransactionHash)
	payload.Deposit = uint64(j.Deposit)

	payload.Address, err = addressFromJSONRawMsg(j.Address)
	if err != nil {
		return nil, err
	}
	return payload, nil
}
