package iotago

import (
	"encoding/json"
	"fmt"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

var (
	// ErrProtocolParamsMilestoneOptInvalid gets returned when a ProtocolParamsMilestoneOpt is invalid.
	ErrProtocolParamsMilestoneOptInvalid = fmt.Errorf("invalid protocol params milestone option")
)

const (
	// MaxParamsLength defines the max length of the data within a ProtocolParamsMilestoneOpt.
	MaxParamsLength = 8192
)

// ProtocolParamsMilestoneOpt is a MilestoneOpt defining changing protocol parameters.
type ProtocolParamsMilestoneOpt struct {
	// The milestone index at which these protocol parameters become active.
	TargetMilestoneIndex uint32
	// The protocol version.
	ProtocolVersion byte
	// The protocol parameters in binary form.
	Params []byte
}

func (p *ProtocolParamsMilestoneOpt) Size() int {
	return serializer.OneByte + util.NumByteLen(p.TargetMilestoneIndex) + util.NumByteLen(p.ProtocolVersion) +
		serializer.UInt16ByteSize + len(p.Params)
}

func (p *ProtocolParamsMilestoneOpt) Type() MilestoneOptType {
	return MilestoneOptProtocolParams
}

func (p *ProtocolParamsMilestoneOpt) Clone() MilestoneOpt {
	return &ProtocolParamsMilestoneOpt{
		TargetMilestoneIndex: p.TargetMilestoneIndex,
		ProtocolVersion:      p.ProtocolVersion,
		Params:               append([]byte{}, p.Params...),
	}
}

func (p *ProtocolParamsMilestoneOpt) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, _ interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(MilestoneOptProtocolParams), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize protocol params milestone option: %w", err)
		}).
		ReadNum(&p.TargetMilestoneIndex, func(err error) error {
			return fmt.Errorf("unable to deserialize protocol params milestone option target milestone index: %w", err)
		}).
		ReadNum(&p.ProtocolVersion, func(err error) error {
			return fmt.Errorf("unable to deserialize protocol params milestone option protocol version: %w", err)
		}).
		ReadVariableByteSlice(&p.Params, serializer.SeriLengthPrefixTypeAsUint16, func(err error) error {
			return fmt.Errorf("unable to deserialize protocol params milestone option parameters: %w", err)
		}, MaxParamsLength).
		Done()
}

func (p *ProtocolParamsMilestoneOpt) Serialize(deSeriMode serializer.DeSerializationMode, _ interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WithValidation(deSeriMode, func(_ []byte, err error) error {
			switch {
			case len(p.Params) == MaxParamsLength:
				return fmt.Errorf("%w: params bigger than %d", ErrProtocolParamsMilestoneOptInvalid, MaxParamsLength)
			}
			return nil
		}).
		WriteNum(byte(MilestoneOptProtocolParams), func(err error) error {
			return fmt.Errorf("unable to serialize protocol params milestone option type ID: %w", err)
		}).
		WriteNum(p.TargetMilestoneIndex, func(err error) error {
			return fmt.Errorf("unable to serialize protocol params milestone option target milestone index: %w", err)
		}).
		WriteNum(p.ProtocolVersion, func(err error) error {
			return fmt.Errorf("unable to serialize protocol params milestone option protocol version: %w", err)
		}).
		WriteVariableByteSlice(p.Params, serializer.SeriLengthPrefixTypeAsUint16, func(err error) error {
			return fmt.Errorf("unable to serialize protocol params milestone option parameters: %w", err)
		}).
		Serialize()
}

func (p *ProtocolParamsMilestoneOpt) MarshalJSON() ([]byte, error) {
	jProtocolParamsMilestoneOpt := &jsonProtocolParamsMilestoneOpt{
		Type:                 int(MilestoneOptProtocolParams),
		TargetMilestoneIndex: int(p.TargetMilestoneIndex),
		ProtocolVersion:      int(p.ProtocolVersion),
		Params:               EncodeHex(p.Params),
	}

	return json.Marshal(jProtocolParamsMilestoneOpt)
}

func (p *ProtocolParamsMilestoneOpt) UnmarshalJSON(bytes []byte) error {
	jProtocolParamsMilestoneOpt := &jsonProtocolParamsMilestoneOpt{}
	if err := json.Unmarshal(bytes, jProtocolParamsMilestoneOpt); err != nil {
		return err
	}
	seri, err := jProtocolParamsMilestoneOpt.ToSerializable()
	if err != nil {
		return err
	}
	*p = *seri.(*ProtocolParamsMilestoneOpt)
	return nil
}

// jsonProtocolParasMilestoneOpt defines the json representation of a ProtocolParamsMilestoneOpt.
type jsonProtocolParamsMilestoneOpt struct {
	Type                 int    `json:"type"`
	TargetMilestoneIndex int    `json:"targetMilestoneIndex"`
	ProtocolVersion      int    `json:"protocolVersion"`
	Params               string `json:"params"`
}

func (j *jsonProtocolParamsMilestoneOpt) ToSerializable() (serializer.Serializable, error) {
	params, err := DecodeHex(j.Params)
	if err != nil {
		return nil, fmt.Errorf("unable to decode json milestone protocol params option: %w", err)
	}
	return &ProtocolParamsMilestoneOpt{
		TargetMilestoneIndex: uint32(j.TargetMilestoneIndex),
		ProtocolVersion:      byte(j.ProtocolVersion),
		Params:               params,
	}, nil
}
