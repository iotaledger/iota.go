package iota_test

import (
	"bytes"
	"encoding/json"
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
			msgPayload, msgPayloadData := randMessage(1337)
			return test{"ok - no payload", msgPayloadData, msgPayload, nil}
		}(),
		func() test {
			msgPayload, msgPayloadData := randMessage(iota.TransactionPayloadTypeID)
			return test{"ok - transaction payload", msgPayloadData, msgPayload, nil}
		}(),
		func() test {
			msgPayload, msgPayloadData := randMessage(iota.MilestonePayloadTypeID)
			return test{"ok - milestone payload", msgPayloadData, msgPayload, nil}
		}(),
		func() test {
			msgPayload, msgPayloadData := randMessage(iota.IndexationPayloadTypeID)
			return test{"ok - indexation payload", msgPayloadData, msgPayload, nil}
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
			msgPayload, msgPayloadData := randMessage(iota.TransactionPayloadTypeID)
			return test{"ok - with transaction payload", msgPayload, msgPayloadData}
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

func TestMessage_UnmarshalJSON(t *testing.T) {
	data := `
		{
		  "version": 1,
          "networkId": "1337133713371337",
		  "parent1MessageId": "f532a53545103276b46876c473846d98648ee418468bce76df4868648dd73e5d",
		  "parent2MessageId": "78d546b46aec4557872139a48f66bc567687e8413578a14323548732358914a2",
		  "payload": {
			"type": 0,
			"essence": {
			  "type": 0,
			  "inputs": [
				{
				  "type": 0,
				  "transactionId": "162863a2f4b134d352a886bf9cfb08788735499694864753ee686e02b3763d9d",
				  "transactionOutputIndex": 3
				}
			  ],
			  "outputs": [
				{
				  "type": 0,
				  "address": {
					"type": 1,
					"address": "5f24ebcb5d48acbbfe6e7401b502ba7bb93acb3591d55eda7d32c37306cc805f"
				  },
				  "amount": 5710
				}
			  ],
			  "payload": {
				"type": 2,
				"index": "allyourtritsbelongtous",
				"data": "a487f431d852b060b49427f513dca1d5288e697e8bd9eb062534d09e7cb337ac"
			  }
			},
			"unlockBlocks": [
			  {
				"type": 0,
				"signature": {
				  "type": 1,
				  "publicKey": "ed3c3f1a319ff4e909cf2771d79fece0ac9bd9fd2ee49ea6c0885c9cb3b1248c",
				  "signature": "651941eddb3e68cb1f6ef4ef5b04625dcf5c70de1fdc4b1c9eadb2c219c074e0ed3c3f1a319ff4e909cf2771d79fece0ac9bd9fd2ee49ea6c0885c9cb3b1248c"
				}
			  }
			]
		  },
		  "nonce": "133945865838"
		}`

	msg := &iota.Message{}
	assert.NoError(t, json.Unmarshal([]byte(data), msg))
	var emptyID = [32]byte{}
	assert.False(t, bytes.Equal(msg.Parent1[:], emptyID[:]))
	assert.False(t, bytes.Equal(msg.Parent2[:], emptyID[:]))

	msgJson, err := json.Marshal(msg)
	assert.NoError(t, err)

	msg2 := &iota.Message{}
	assert.NoError(t, json.Unmarshal(msgJson, msg2))

	assert.EqualValues(t, msg, msg2)

	minimal := `
		{
		  "payload": null
		}`
	msgMinimal := &iota.Message{}
	assert.NoError(t, json.Unmarshal([]byte(minimal), msgMinimal))
	assert.True(t, bytes.Equal(msgMinimal.Parent1[:], emptyID[:]))
	assert.True(t, bytes.Equal(msgMinimal.Parent2[:], emptyID[:]))
	assert.Nil(t, msgMinimal.Payload)
	assert.Equal(t, msgMinimal.Version, byte(iota.MessageVersion))
	assert.Equal(t, msgMinimal.Nonce, uint64(0))
}
