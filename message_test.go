package iota_test

import (
	"errors"
	"testing"

	"github.com/iotaledger/iota.go"
	"github.com/stretchr/testify/assert"
)

func TestMessage_Deserialize(t *testing.T) {
	type test struct {
		name   string
		source []byte
		target iota.Serializable
		err    error
	}
	tests := []test{
		func() test {
			msgPayload, msgPayloadData := randMessage(iota.SignedTransactionPayloadID)
			return test{"ok", msgPayloadData, msgPayload, nil}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &iota.Message{}
			bytesRead, err := msg.Deserialize(tt.source, iota.DeSeriModePerformValidation)
			if tt.err != nil {
				assert.True(t, errors.Is(err, tt.err))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, len(tt.source), bytesRead)
			assert.EqualValues(t, tt.target, msg)
		})
	}
}

func TestMessage_Serialize(t *testing.T) {
	type test struct {
		name   string
		source *iota.Message
		target []byte
	}
	tests := []test{
		func() test {
			msgPayload, msgPayloadData := randMessage(iota.SignedTransactionPayloadID)
			return test{"ok - with signed transaction payload", msgPayload, msgPayloadData}
		}(),
		func() test {
			msgPayload, msgPayloadData := randMessage(1337)
			return test{"ok - without any payload", msgPayload, msgPayloadData}
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
