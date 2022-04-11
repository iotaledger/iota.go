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

// ProtocolParamsMilestoneOpt is a MilestoneOpt defining changing protocol parameters.
type ProtocolParamsMilestoneOpt struct {
	// The next minimum PoW score to use after NextPoWScoreMilestoneIndex is hit.
	NextPoWScore uint32
	// The milestone index at which the PoW score changes to NextPoWScore.
	NextPoWScoreMilestoneIndex uint32
}

func (p *ProtocolParamsMilestoneOpt) Size() int {
	return serializer.OneByte + util.NumByteLen(p.NextPoWScore) + util.NumByteLen(p.NextPoWScoreMilestoneIndex)
}

func (p *ProtocolParamsMilestoneOpt) Type() MilestoneOptType {
	return MilestoneOptProtocolParams
}

func (p *ProtocolParamsMilestoneOpt) Clone() MilestoneOpt {
	return &ProtocolParamsMilestoneOpt{
		NextPoWScore:               p.NextPoWScore,
		NextPoWScoreMilestoneIndex: p.NextPoWScoreMilestoneIndex,
	}
}

func (p *ProtocolParamsMilestoneOpt) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, _ interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(MilestoneOptProtocolParams), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize protocol params milestone option: %w", err)
		}).
		ReadNum(&p.NextPoWScore, func(err error) error {
			return fmt.Errorf("unable to deserialize protocol params milestone option next pow score: %w", err)
		}).
		ReadNum(&p.NextPoWScoreMilestoneIndex, func(err error) error {
			return fmt.Errorf("unable to deserialize protocol params milestone option next pow score milestone index: %w", err)
		}).
		WithValidation(deSeriMode, func(_ []byte, err error) error {
			switch {
			case p.NextPoWScoreMilestoneIndex == 0:
				return fmt.Errorf("%w: next-pow-score-milestone-index is zero", ErrProtocolParamsMilestoneOptInvalid)
			}
			return nil
		}).
		Done()
}

func (p *ProtocolParamsMilestoneOpt) Serialize(deSeriMode serializer.DeSerializationMode, _ interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WithValidation(deSeriMode, func(_ []byte, err error) error {
			switch {
			case p.NextPoWScoreMilestoneIndex == 0:
				return fmt.Errorf("%w: next-pow-score-milestone-index is zero", ErrProtocolParamsMilestoneOptInvalid)
			}
			return nil
		}).
		WriteNum(byte(MilestoneOptProtocolParams), func(err error) error {
			return fmt.Errorf("unable to serialize protocol params milestone option type ID: %w", err)
		}).
		WriteNum(p.NextPoWScore, func(err error) error {
			return fmt.Errorf("unable to serialize protocol params milestone option next pow score: %w", err)
		}).
		WriteNum(p.NextPoWScoreMilestoneIndex, func(err error) error {
			return fmt.Errorf("unable to serialize protocol params milestone option next pow score milestone index: %w", err)
		}).
		Serialize()
}

func (p *ProtocolParamsMilestoneOpt) MarshalJSON() ([]byte, error) {
	jProtocolParamsMilestoneOpt := &jsonProtocolParamsMilestoneOpt{
		Type:                       int(MilestoneOptProtocolParams),
		NextPoWScore:               int(p.NextPoWScore),
		NextPoWScoreMilestoneIndex: int(p.NextPoWScoreMilestoneIndex),
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
	Type                       int `json:"type"`
	NextPoWScore               int `json:"nextPoWScore"`
	NextPoWScoreMilestoneIndex int `json:"nextPoWScoreMilestoneIndex"`
}

func (j *jsonProtocolParamsMilestoneOpt) ToSerializable() (serializer.Serializable, error) {
	return &ProtocolParamsMilestoneOpt{
		NextPoWScore:               uint32(j.NextPoWScore),
		NextPoWScoreMilestoneIndex: uint32(j.NextPoWScoreMilestoneIndex),
	}, nil
}
