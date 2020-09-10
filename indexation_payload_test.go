package iota_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/iota.go"
	"github.com/stretchr/testify/assert"
)

func TestIndexationPayload_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target iota.Serializable
		err    error
	}
	tests := []test{
		func() test {
			indexationPayload, indexationPayloadData := randIndexationPayload()
			return test{"ok", indexationPayloadData, indexationPayload, nil}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indexationPayload := &iota.IndexationPayload{}
			bytesRead, err := indexationPayload.Deserialize(tt.source, iota.DeSeriModePerformValidation)
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

func TestIndexationPayload_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iota.IndexationPayload
		target []byte
	}
	tests := []test{
		func() test {
			indexationPayload, indexationPayloadData := randIndexationPayload()
			return test{"ok", indexationPayload, indexationPayloadData}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edData, err := tt.source.Serialize(iota.DeSeriModePerformValidation)
			assert.NoError(t, err)
			assert.Equal(t, tt.target, edData)
		})
	}
}
