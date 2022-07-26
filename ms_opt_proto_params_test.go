package iotago_test

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
)

func TestProtocolParamsMilestoneOpt_DeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name: "ok - protocol params milestone option",
			source: &iotago.ProtocolParamsMilestoneOpt{
				TargetMilestoneIndex: 1337,
				ProtocolVersion:      3,
				Params:               []byte{1, 2, 3, 4, 5, 6, 7, 8, 9},
			},
			target: &iotago.ProtocolParamsMilestoneOpt{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}
