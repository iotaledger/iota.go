package iotago_test

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg/frameworks"
)

func TestCandidacyAnnouncmentDeSerialize(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name:   "ok",
			Source: &iotago.CandidacyAnnouncement{},
			Target: &iotago.CandidacyAnnouncement{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}
