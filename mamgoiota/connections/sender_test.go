package mamgoiota

import (
	"fmt"
	"testing"

	"github.com/iotaledger/mamgoiota/mamutils"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/giota"
)

type FakeSender struct {
	OnSendToApi func([]giota.Transfer) (giota.Bundle, error)
}

func (f *FakeSender) SendToApi(ts []giota.Transfer) (giota.Bundle, error) {
	return f.OnSendToApi(ts)
}

func TestSend(t *testing.T) {
	assert := assert.New(t)
	recipient, err := giota.NewAddress(giota.NewSeed(), 0, 1)
	assert.Nil(err)

	transactionId, err := Send(string(recipient), 1000, "ABC", &FakeSender{
		OnSendToApi: func(ts []giota.Transfer) (giota.Bundle, error) {
			encodedMessage, _ := mamutils.ToMAMTrytes("ABC")

			assert.Len(ts, 1)
			assert.EqualValues(giota.Transfer{
				Address: recipient,
				Value:   1000,
				Message: encodedMessage,
				Tag:     "",
			}, ts[0])

			return giota.Bundle{giota.Transaction{}}, err
		},
	})

	assert.EqualValues((&giota.Transaction{}).Hash(), transactionId)
	assert.Nil(err)
}
func TestSendErrorRecipient(t *testing.T) {
	assert := assert.New(t)

	transactionId, err := Send("123", 1000, "ABC", &FakeSender{})

	assert.EqualValues("", transactionId)
	assert.NotNil(err)
}
func TestSendErrorSendAPI(t *testing.T) {
	assert := assert.New(t)
	recipient, err := giota.NewAddress(giota.NewSeed(), 0, 1)
	if err != nil {
		t.Error(err)
	}

	sendError := fmt.Errorf("SOMESENDERROR")
	transactionId, err := Send(string(recipient), 1000, "ABC", &FakeSender{
		OnSendToApi: func(ts []giota.Transfer) (giota.Bundle, error) {
			return giota.Bundle{}, sendError
		},
	})

	assert.EqualValues("", transactionId)
	assert.EqualValues(sendError, err)
}
