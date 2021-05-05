package iotago_test

import (
	"errors"
	test2 "github.com/iotaledger/iota.go/v2/test"
	"testing"

	"github.com/iotaledger/iota.go/v2"
	"github.com/stretchr/testify/assert"
)

func TestIndexation_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target iotago.Serializable
		err    error
	}
	tests := []test{
		func() test {
			indexationPayload, indexationPayloadData := test2.RandIndexation()
			return test{"ok", indexationPayloadData, indexationPayload, nil}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indexationPayload := &iotago.Indexation{}
			bytesRead, err := indexationPayload.Deserialize(tt.source, iotago.DeSeriModePerformValidation)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, len(tt.source), bytesRead)
			assert.EqualValues(t, tt.target, indexationPayload)
		})
	}
}

func TestIndexation_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iotago.Indexation
		target []byte
	}
	tests := []test{
		func() test {
			indexationPayload, indexationPayloadData := test2.RandIndexation()
			return test{"ok", indexationPayload, indexationPayloadData}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edData, err := tt.source.Serialize(iotago.DeSeriModePerformValidation)
			assert.NoError(t, err)
			assert.Equal(t, tt.target, edData)
		})
	}
}
