package iotago_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/iota.go/v3/tpkg"

	"github.com/stretchr/testify/assert"

	iotago "github.com/iotaledger/iota.go/v3"
)

func TestTransactionEssenceSelector(t *testing.T) {
	_, err := iotago.TransactionEssenceSelector(100)
	assert.True(t, errors.Is(err, iotago.ErrUnknownTransactionEssenceType))
}

func TestTransactionEssence_DeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok",
			source: tpkg.RandTransactionEssence(),
			target: &iotago.TransactionEssence{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}
