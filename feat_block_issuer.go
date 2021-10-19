package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer"
)

// IssuerFeatureBlock is a feature block which associates an output
// with an issuer identity. Unlike the SenderFeatureBlock, the issuer identity
// only has to be unlocked when the state machine type of output is first created,
// afterwards, the issuer block must not change, meaning that subsequent outputs
// must always define the same issuer identity (the identity does not need to be unlocked anymore though).
type IssuerFeatureBlock struct {
	Address serializer.Serializable
}

func (s *IssuerFeatureBlock) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	return serializer.NewDeserializer(data).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := serializer.CheckTypeByte(data, FeatureBlockIssuer); err != nil {
					return fmt.Errorf("unable to deserialize issuer feature block: %w", err)
				}
			}
			return nil
		}).
		Skip(serializer.SmallTypeDenotationByteSize, func(err error) error {
			return fmt.Errorf("unable to skip issuer feature block type during deserialization: %w", err)
		}).
		ReadObject(func(seri serializer.Serializable) { s.Address = seri }, deSeriMode, serializer.TypeDenotationByte, AddressSelector, func(err error) error {
			return fmt.Errorf("unable to deserialize address for issuer feature block: %w", err)
		}).Done()
}

func (s *IssuerFeatureBlock) Serialize(deSeriMode serializer.DeSerializationMode) ([]byte, error) {
	return serializer.NewSerializer().
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := isValidAddrType(s.Address); err != nil {
					return fmt.Errorf("invalid address set in issuer feature block: %w", err)
				}
			}
			return nil
		}).
		WriteObject(s.Address, deSeriMode, func(err error) error {
			return fmt.Errorf("unable to serialize issuer feature block address: %w", err)
		}).
		Serialize()
}

func (s *IssuerFeatureBlock) MarshalJSON() ([]byte, error) {
	jIssuerFeatBlock := &jsonIssuerFeatureBlock{}

	addrJsonBytes, err := s.Address.MarshalJSON()
	if err != nil {
		return nil, err
	}
	jsonRawMsgAddr := json.RawMessage(addrJsonBytes)

	jIssuerFeatBlock.Type = int(FeatureBlockIssuer)
	jIssuerFeatBlock.Address = &jsonRawMsgAddr
	return json.Marshal(jIssuerFeatBlock)
}

func (s *IssuerFeatureBlock) UnmarshalJSON(bytes []byte) error {
	jIssuerFeatBlock := &jsonIssuerFeatureBlock{}
	if err := json.Unmarshal(bytes, jIssuerFeatBlock); err != nil {
		return err
	}
	seri, err := jIssuerFeatBlock.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*IssuerFeatureBlock)
	return nil
}

// jsonIssuerFeatureBlock defines the json representation of a IssuerFeatureBlock.
type jsonIssuerFeatureBlock struct {
	Type    int              `json:"type"`
	Address *json.RawMessage `json:"address"`
}

func (j *jsonIssuerFeatureBlock) ToSerializable() (serializer.Serializable, error) {
	dep := &IssuerFeatureBlock{}

	jsonAddr, err := DeserializeObjectFromJSON(j.Address, jsonAddressSelector)
	if err != nil {
		return nil, fmt.Errorf("can't decode address type from JSON: %w", err)
	}

	dep.Address, err = jsonAddr.ToSerializable()
	if err != nil {
		return nil, err
	}
	return dep, nil
}
