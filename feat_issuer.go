package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

var (
	issuerFeatAddrGuard = serializer.SerializableGuard{
		ReadGuard:  AddressReadGuard(allAddressTypeSet),
		WriteGuard: AddressWriteGuard(allAddressTypeSet),
	}
)

// IssuerFeature is a feature which associates an output
// with an issuer identity. Unlike the SenderFeature, the issuer identity
// only has to be unlocked when the ChainConstrainedOutput is first created,
// afterwards, the issuer feature must not change, meaning that subsequent outputs
// must always define the same issuer identity (the identity does not need to be unlocked anymore though).
type IssuerFeature struct {
	Address Address
}

func (s *IssuerFeature) Clone() Feature {
	return &IssuerFeature{Address: s.Address.Clone()}
}

func (s *IssuerFeature) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	return rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize) +
		s.Address.VBytes(rentStruct, nil)
}

func (s *IssuerFeature) ByteSizeKey() uint64 {
	return 0 + s.Address.ByteSizeKey()
}

func (s *IssuerFeature) ByteSizeData() uint64 {
	return serializer.SmallTypeDenotationByteSize +
		s.Address.ByteSizeData()
}

func (s *IssuerFeature) Equal(other Feature) bool {
	otherFeat, is := other.(*IssuerFeature)
	if !is {
		return false
	}

	return s.Address.Equal(otherFeat.Address)
}

func (s *IssuerFeature) Type() FeatureType {
	return FeatureIssuer
}

func (s *IssuerFeature) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(FeatureIssuer), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize issuer feature: %w", err)
		}).
		ReadObject(&s.Address, deSeriMode, deSeriCtx, serializer.TypeDenotationByte, issuerFeatAddrGuard.ReadGuard, func(err error) error {
			return fmt.Errorf("unable to deserialize address for issuer feature: %w", err)
		}).Done()
}

func (s *IssuerFeature) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(byte(FeatureIssuer), func(err error) error {
			return fmt.Errorf("unable to serialize issuer feature type ID: %w", err)
		}).
		WriteObject(s.Address, deSeriMode, deSeriCtx, issuerFeatAddrGuard.WriteGuard, func(err error) error {
			return fmt.Errorf("unable to serialize issuer feature address: %w", err)
		}).
		Serialize()
}

func (s *IssuerFeature) Size() int {
	return util.NumByteLen(byte(FeatureIssuer)) + s.Address.Size()
}

func (s *IssuerFeature) MarshalJSON() ([]byte, error) {
	jIssuerFeat := &jsonIssuerFeature{}

	addrJsonBytes, err := s.Address.MarshalJSON()
	if err != nil {
		return nil, err
	}
	jsonRawMsgAddr := json.RawMessage(addrJsonBytes)

	jIssuerFeat.Type = int(FeatureIssuer)
	jIssuerFeat.Address = &jsonRawMsgAddr
	return json.Marshal(jIssuerFeat)
}

func (s *IssuerFeature) UnmarshalJSON(bytes []byte) error {
	jIssuerFeat := &jsonIssuerFeature{}
	if err := json.Unmarshal(bytes, jIssuerFeat); err != nil {
		return err
	}
	seri, err := jIssuerFeat.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*IssuerFeature)
	return nil
}

// jsonIssuerFeature defines the json representation of a IssuerFeature.
type jsonIssuerFeature struct {
	Type    int              `json:"type"`
	Address *json.RawMessage `json:"address"`
}

func (j *jsonIssuerFeature) ToSerializable() (serializer.Serializable, error) {
	dep := &IssuerFeature{}

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
