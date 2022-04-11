package iotago_test

import (
	"github.com/iotaledger/iota.go/v3"
	"testing"
)

func TestProtocolParamsMilestoneOpt_DeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name: "ok - protocol params milestone option",
			source: &iotago.ProtocolParamsMilestoneOpt{
				NextPoWScore:               666,
				NextPoWScoreMilestoneIndex: 1337,
			},
			target: &iotago.ProtocolParamsMilestoneOpt{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}
