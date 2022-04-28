package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

// ProtocolParamsMilestoneOpt is a MilestoneOpt defining changing protocol parameters.
type ProtocolParamsMilestoneOpt struct {
	// The milestone index at which these protocol parameters become active.
	TargetMilestoneIndex MilestoneIndex `serix:"0,mapKey=targetMilestoneIndex"`
	// The protocol version.
	ProtocolVersion byte `serix:"1,mapKey=protocolVersion"`
	// The protocol parameters in binary form.
	Params []byte `serix:"2,lengthPrefixType=uint16,mapKey=params,maxLen=8192"`
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
