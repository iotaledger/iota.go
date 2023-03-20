package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

var (
	senderFeatAddrGuard = serializer.SerializableGuard{
		ReadGuard:  AddressReadGuard(allAddressTypeSet),
		WriteGuard: AddressWriteGuard(allAddressTypeSet),
	}
)

// SenderFeature is a feature which associates an output
// with a sender identity. The sender identity needs to be unlocked in the transaction
// for the SenderFeature to be valid.
type SenderFeature struct {
	Address Address
}

func (s *SenderFeature) Clone() Feature {
	return &SenderFeature{Address: s.Address.Clone()}
}

func (s *SenderFeature) VBytes(rentStruct *RentStructure, f VBytesFunc) uint64 {
	if f != nil {
		return f(rentStruct)
	}
	return rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize) + s.Address.VBytes(rentStruct, nil)
}

func (s *SenderFeature) Equal(other Feature) bool {
	otherFeat, is := other.(*SenderFeature)
	if !is {
		return false
	}

	return s.Address.Equal(otherFeat.Address)
}

func (s *SenderFeature) Type() FeatureType {
	return FeatureSender
}

func (s *SenderFeature) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(FeatureSender), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize sender feature: %w", err)
		}).
		ReadObject(&s.Address, deSeriMode, deSeriCtx, serializer.TypeDenotationByte, senderFeatAddrGuard.ReadGuard, func(err error) error {
			return fmt.Errorf("unable to deserialize address for sender feature: %w", err)
		}).Done()
}

func (s *SenderFeature) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(byte(FeatureSender), func(err error) error {
			return fmt.Errorf("unable to serialize sender feature type ID: %w", err)
		}).
		WriteObject(s.Address, deSeriMode, deSeriCtx, senderFeatAddrGuard.WriteGuard, func(err error) error {
			return fmt.Errorf("unable to serialize sender feature address: %w", err)
		}).
		Serialize()
}

func (s *SenderFeature) Size() int {
	return util.NumByteLen(byte(FeatureSender)) + s.Address.Size()
}

func (s *SenderFeature) MarshalJSON() ([]byte, error) {
	jSenderFeat := &jsonSenderFeature{}

	addrJsonBytes, err := s.Address.MarshalJSON()
	if err != nil {
		return nil, err
	}
	jsonRawMsgAddr := json.RawMessage(addrJsonBytes)

	jSenderFeat.Type = int(FeatureSender)
	jSenderFeat.Address = &jsonRawMsgAddr
	return json.Marshal(jSenderFeat)
}

func (s *SenderFeature) UnmarshalJSON(bytes []byte) error {
	jSenderFeat := &jsonSenderFeature{}
	if err := json.Unmarshal(bytes, jSenderFeat); err != nil {
		return err
	}
	seri, err := jSenderFeat.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*SenderFeature)
	return nil
}

// jsonSenderFeature defines the json representation of a SenderFeature.
type jsonSenderFeature struct {
	Type    int              `json:"type"`
	Address *json.RawMessage `json:"address"`
}

func (j *jsonSenderFeature) ToSerializable() (serializer.Serializable, error) {
	dep := &SenderFeature{}

	jsonAddr, err := DeserializeObjectFromJSON(j.Address, jsonAddressSelector)
	if err != nil {
		return nil, fmt.Errorf("can't decode address type from JSON: %w", err)
	}

	dep.Address, err = jsonAddressToAddress(jsonAddr)
	if err != nil {
		return nil, err
	}
	return dep, nil
}
