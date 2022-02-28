package iotago_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/stretchr/testify/assert"
)

func TestMessage_DeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok - no payload",
			source: tpkg.RandMessage(1337),
			target: &iotago.Message{},
		},
		{
			name:   "ok - transaction",
			source: tpkg.RandMessage(iotago.PayloadTransaction),
			target: &iotago.Message{},
		},
		{
			name:   "ok - milestone",
			source: tpkg.RandMessage(iotago.PayloadMilestone),
			target: &iotago.Message{},
		},
		{
			name:   "ok - tagged data",
			source: tpkg.RandMessage(iotago.PayloadTaggedData),
			target: &iotago.Message{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}

func TestMessage_UnmarshalJSON(t *testing.T) {
	data := `
		{
		  "protocolVersion": 1,
		  "parentMessageIds": ["f532a53545103276b46876c473846d98648ee418468bce76df4868648dd73e5d", "78d546b46aec4557872139a48f66bc567687e8413578a14323548732358914a2"],
		  "payload": {
			"type": 6,
			"essence": {
			  "type": 0,
              "networkId": "1337133713371337",
			  "inputs": [
				{
				  "type": 0,
				  "transactionId": "162863a2f4b134d352a886bf9cfb08788735499694864753ee686e02b3763d9d",
				  "transactionOutputIndex": 3
				}
			  ],
			  "outputs": [
				{
				  "type": 3,
				  "address": {
					"type": 0,
					"address": "5f24ebcb5d48acbbfe6e7401b502ba7bb93acb3591d55eda7d32c37306cc805f"
				  },
				  "amount": 5710
				}
			  ],
			  "payload": {
				"type": 5,
				"tag": "616c6c796f7572747269747362656c6f6e67746f7573",
				"data": "a487f431d852b060b49427f513dca1d5288e697e8bd9eb062534d09e7cb337ac"
			  }
			},
			"unlockBlocks": [
			  {
				"type": 0,
				"signature": {
				  "type": 0,
				  "publicKey": "ed3c3f1a319ff4e909cf2771d79fece0ac9bd9fd2ee49ea6c0885c9cb3b1248c",
				  "signature": "651941eddb3e68cb1f6ef4ef5b04625dcf5c70de1fdc4b1c9eadb2c219c074e0ed3c3f1a319ff4e909cf2771d79fece0ac9bd9fd2ee49ea6c0885c9cb3b1248c"
				}
			  }
			]
		  },
		  "nonce": "133945865838"
		}`

	msg := &iotago.Message{}
	assert.NoError(t, json.Unmarshal([]byte(data), msg))

	var emptyID = [32]byte{}
	for _, parent := range msg.Parents {
		assert.False(t, bytes.Equal(parent[:], emptyID[:]))
	}

	msgJson, err := json.Marshal(msg)
	assert.NoError(t, err)

	msg2 := &iotago.Message{}
	assert.NoError(t, json.Unmarshal(msgJson, msg2))

	assert.EqualValues(t, msg, msg2)

	minimal := `
		{
		  "parentMessageIds": ["0000000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000000"],
		  "payload": null
		}`
	msgMinimal := &iotago.Message{}
	assert.NoError(t, json.Unmarshal([]byte(minimal), msgMinimal))

	assert.Len(t, msgMinimal.Parents, 2)
	for _, parent := range msgMinimal.Parents {
		assert.True(t, bytes.Equal(parent[:], emptyID[:]))
	}

	assert.Nil(t, msgMinimal.Payload)
	assert.Equal(t, msgMinimal.Nonce, uint64(0))
}
