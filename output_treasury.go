package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

// TreasuryOutput is an output which holds the treasury of a network.
type TreasuryOutput struct {
	// The currently residing funds in the treasury.
	Amount uint64
}

func (t *TreasuryOutput) NativeTokenSet() NativeTokens {
	return nil
}

func (t *TreasuryOutput) UnlockConditionsSet() UnlockConditionsSet {
	return nil
}

func (t *TreasuryOutput) FeaturesSet() FeaturesSet {
	return nil
}

func (t *TreasuryOutput) Clone() Output {
	return &TreasuryOutput{Amount: t.Amount}
}

func (t *TreasuryOutput) VBytes(_ *RentStructure, _ VBytesFunc) uint64 {
	return 0
}

func (t *TreasuryOutput) Deposit() uint64 {
	return t.Amount
}

func (t *TreasuryOutput) Type() OutputType {
	return OutputTreasury
}

func (t *TreasuryOutput) Deserialize(data []byte, _ serializer.DeSerializationMode, _ interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(OutputTreasury), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize treasury output: %w", err)
		}).
		ReadNum(&t.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for treasury output: %w", err)
		}).
		Done()
}

func (t *TreasuryOutput) Serialize(_ serializer.DeSerializationMode, _ interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(byte(OutputTreasury), func(err error) error {
			return fmt.Errorf("unable to serialize treasury output type ID: %w", err)
		}).
		WriteNum(t.Amount, func(err error) error {
			return fmt.Errorf("unable to serialize treasury output amount: %w", err)
		}).Serialize()
}

func (t *TreasuryOutput) Size() int {
	return util.NumByteLen(byte(OutputTreasury)) + util.NumByteLen(t.Amount)
}

func (t *TreasuryOutput) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsonTreasuryOutput{
		Type:   int(OutputTreasury),
		Amount: EncodeUint64(t.Amount),
	})
}

func (t *TreasuryOutput) UnmarshalJSON(bytes []byte) error {
	jTreasuryOutput := &jsonTreasuryOutput{}
	if err := json.Unmarshal(bytes, jTreasuryOutput); err != nil {
		return err
	}
	seri, err := jTreasuryOutput.ToSerializable()
	if err != nil {
		return err
	}
	*t = *seri.(*TreasuryOutput)
	return nil
}

// jsonTreasuryOutput defines the json representation of a TreasuryOutput.
type jsonTreasuryOutput struct {
	Type   int    `json:"type"`
	Amount string `json:"amount"`
}

func (j *jsonTreasuryOutput) ToSerializable() (serializer.Serializable, error) {
	var err error
	t := &TreasuryOutput{}
	t.Amount, err = DecodeUint64(j.Amount)
	if err != nil {
		return nil, err
	}
	return t, nil
}
