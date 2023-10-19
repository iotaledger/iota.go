package iotago_test

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v4"
)

func TestCandidacyAnnouncmentDeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok",
			source: &iotago.CandidacyAnnouncement{},
			target: &iotago.CandidacyAnnouncement{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}
