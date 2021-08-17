package iotago

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/iotaledger/hive.go/serializer"
)

const (
	// RefUTXOIndexMin is the minimum index of a referenced UTXO.
	RefUTXOIndexMin = 0
	// RefUTXOIndexMax is the maximum index of a referenced UTXO.
	RefUTXOIndexMax = 126

	// UTXOInputSize is the size of a UTXO input: input type + tx id + index
	UTXOInputSize = serializer.SmallTypeDenotationByteSize + TransactionIDLength + serializer.UInt16ByteSize
)

// UTXOInputID defines the identifier for an UTXO input which consists
// out of the referenced transaction ID and the given output index.
type UTXOInputID [TransactionIDLength + serializer.UInt16ByteSize]byte

// ToHex converts the UTXOInputID to its hex representation.
func (utxoInputID UTXOInputID) ToHex() string {
	return fmt.Sprintf("%x", utxoInputID)
}

// UTXOInputIDs is a slice of UTXOInputID.
type UTXOInputIDs []UTXOInputID

// ToHex converts all UTXOInput to their hex string representation.
func (utxoInputIDs UTXOInputIDs) ToHex() []string {
	ids := make([]string, len(utxoInputIDs))
	for i := range utxoInputIDs {
		ids[i] = fmt.Sprintf("%x", utxoInputIDs[i])
	}
	return ids
}

// UTXOInput references an unspent transaction output by the Transaction's ID and the corresponding index of the output.
type UTXOInput struct {
	// The transaction ID of the referenced transaction.
	TransactionID [TransactionIDLength]byte
	// The output index of the output on the referenced transaction.
	TransactionOutputIndex uint16
}

// ID returns the UTXOInputID.
func (u *UTXOInput) ID() UTXOInputID {
	var id UTXOInputID
	copy(id[:TransactionIDLength], u.TransactionID[:])
	binary.LittleEndian.PutUint16(id[TransactionIDLength:], u.TransactionOutputIndex)
	return id
}

func (u *UTXOInput) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	return serializer.NewDeserializer(data).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := serializer.CheckMinByteLength(UTXOInputSize, len(data)); err != nil {
					return fmt.Errorf("invalid UTXO input bytes: %w", err)
				}
				if err := serializer.CheckTypeByte(data, InputUTXO); err != nil {
					return fmt.Errorf("unable to deserialize UTXO input: %w", err)
				}
			}
			return nil
		}).
		Skip(serializer.SmallTypeDenotationByteSize, func(err error) error {
			return fmt.Errorf("unable to skip UTXO input type during deserialization: %w", err)
		}).
		ReadArrayOf32Bytes(&u.TransactionID, func(err error) error {
			return fmt.Errorf("unable to deserialize transaction ID in UTXO input: %w", err)
		}).
		ReadNum(&u.TransactionOutputIndex, func(err error) error {
			return fmt.Errorf("unable to deserialize transaction output index in UTXO input: %w", err)
		}).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := utxoInputRefBoundsValidator(-1, u); err != nil {
					return fmt.Errorf("%w: unable to deserialize UTXO input", err)
				}
			}
			return nil
		}).
		Done()
}

func (u *UTXOInput) Serialize(deSeriMode serializer.DeSerializationMode) (data []byte, err error) {
	return serializer.NewSerializer().
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := utxoInputRefBoundsValidator(-1, u); err != nil {
					return fmt.Errorf("%w: unable to serialize UTXO input", err)
				}
			}
			return nil
		}).
		WriteNum(InputUTXO, func(err error) error {
			return fmt.Errorf("unable to serialize UTXO input type ID: %w", err)
		}).
		WriteBytes(u.TransactionID[:], func(err error) error {
			return fmt.Errorf("unable to serialize UTXO input transaction ID: %w", err)
		}).
		WriteNum(u.TransactionOutputIndex, func(err error) error {
			return fmt.Errorf("unable to serialize UTXO input transaction output index: %w", err)
		}).Serialize()
}

func (u *UTXOInput) MarshalJSON() ([]byte, error) {
	jUTXOInput := &jsonUTXOInput{}
	jUTXOInput.TransactionID = hex.EncodeToString(u.TransactionID[:])
	jUTXOInput.TransactionOutputIndex = int(u.TransactionOutputIndex)
	jUTXOInput.Type = int(InputUTXO)
	return json.Marshal(jUTXOInput)
}

func (u *UTXOInput) UnmarshalJSON(bytes []byte) error {
	jUTXOInput := &jsonUTXOInput{}
	if err := json.Unmarshal(bytes, jUTXOInput); err != nil {
		return err
	}
	seri, err := jUTXOInput.ToSerializable()
	if err != nil {
		return err
	}
	*u = *seri.(*UTXOInput)
	return nil
}

// jsonUTXOInput defines the JSON representation of a UTXOInput.
type jsonUTXOInput struct {
	Type                   int    `json:"type"`
	TransactionID          string `json:"transactionId"`
	TransactionOutputIndex int    `json:"transactionOutputIndex"`
}

func (j *jsonUTXOInput) ToSerializable() (serializer.Serializable, error) {
	utxoInput := &UTXOInput{
		TransactionID:          [32]byte{},
		TransactionOutputIndex: uint16(j.TransactionOutputIndex),
	}
	transactionIDBytes, err := hex.DecodeString(j.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("unable to decode transaction ID from JSON for UTXO input: %w", err)
	}
	copy(utxoInput.TransactionID[:], transactionIDBytes)
	return utxoInput, nil
}
