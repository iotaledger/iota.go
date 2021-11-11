package iotago

import (
	"encoding/json"
	"fmt"
	"github.com/iotaledger/hive.go/serializer"
)

// SenderFeatureBlock is a feature block which associates an output
// with a sender identity. The sender identity needs to be unlocked in the transaction
// for the SenderFeatureBlock block to be valid.
type SenderFeatureBlock struct {
	Address Address
}

func (s *SenderFeatureBlock) Equal(other FeatureBlock) bool {
	otherBlock, is := other.(*SenderFeatureBlock)
	if !is {
		return false
	}

	return s.Address.Equal(otherBlock.Address)
}

func (s *SenderFeatureBlock) Type() FeatureBlockType {
	return FeatureBlockSender
}

func (s *SenderFeatureBlock) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(FeatureBlockSender), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize sender feature block: %w", err)
		}).
		ReadObject(&s.Address, deSeriMode, serializer.TypeDenotationByte, AddressSelector, func(err error) error {
			return fmt.Errorf("unable to deserialize address for sender feature block: %w", err)
		}).Done()
}

func (s *SenderFeatureBlock) Serialize(deSeriMode serializer.DeSerializationMode) ([]byte, error) {
	return serializer.NewSerializer().
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := isValidAddrType(s.Address); err != nil {
					return fmt.Errorf("invalid address set in sender feature block: %w", err)
				}
			}
			return nil
		}).
		WriteNum(byte(FeatureBlockSender), func(err error) error {
			return fmt.Errorf("unable to serialize sender feature block type ID: %w", err)
		}).
		WriteObject(s.Address, deSeriMode, func(err error) error {
			return fmt.Errorf("unable to serialize sender feature block address: %w", err)
		}).
		Serialize()
}

func (s *SenderFeatureBlock) MarshalJSON() ([]byte, error) {
	jSenderFeatBlock := &jsonSenderFeatureBlock{}

	addrJsonBytes, err := s.Address.MarshalJSON()
	if err != nil {
		return nil, err
	}
	jsonRawMsgAddr := json.RawMessage(addrJsonBytes)

	jSenderFeatBlock.Type = int(FeatureBlockSender)
	jSenderFeatBlock.Address = &jsonRawMsgAddr
	return json.Marshal(jSenderFeatBlock)
}

func (s *SenderFeatureBlock) UnmarshalJSON(bytes []byte) error {
	jSenderFeatBlock := &jsonSenderFeatureBlock{}
	if err := json.Unmarshal(bytes, jSenderFeatBlock); err != nil {
		return err
	}
	seri, err := jSenderFeatBlock.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*SenderFeatureBlock)
	return nil
}

// jsonSenderFeatureBlock defines the json representation of a SenderFeatureBlock.
type jsonSenderFeatureBlock struct {
	Type    int              `json:"type"`
	Address *json.RawMessage `json:"address"`
}

func (j *jsonSenderFeatureBlock) ToSerializable() (serializer.Serializable, error) {
	dep := &SenderFeatureBlock{}

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
