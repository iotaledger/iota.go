package iotagox_test

import (
	"context"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/iota.go/v3/x"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/stretchr/testify/require"
)

func TestNewNodeEventAPIClient(t *testing.T) {
	msg := tpkg.RandMessage(iotago.PayloadTaggedData)
	originMsgBytes, err := msg.Serialize(serializer.DeSeriModeNoValidation, nil)
	require.NoError(t, err)
	mock := &mockMqttClient{payload: originMsgBytes}
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	eventAPIClient := &iotagox.NodeEventAPIClient{
		MQTTClient: mock,
		Errors:     make(chan error),
	}
	require.NoError(t, eventAPIClient.Connect(ctx))

	msgChan, sub := eventAPIClient.Messages()
	require.NoError(t, sub.Error())
	require.Eventually(t, func() bool {
		select {
		case msg := <-msgChan:
			gottenMsgBytes, err := msg.Serialize(serializer.DeSeriModeNoValidation, nil)
			require.NoError(t, err)
			require.Equal(t, originMsgBytes, gottenMsgBytes)

			require.NoError(t, sub.Close())
			require.ErrorIs(t, sub.Close(), iotagox.ErrNodeEventAPIClientSubscriptionAlreadyClosed)

			return true
		default:
			return false
		}
	}, 5*time.Second, 100*time.Millisecond)
}

type mockMqttClient struct {
	payload []byte
	f       func()
}

type mockToken struct{}

func (m *mockToken) Wait() bool {
	return false
}

func (m *mockToken) WaitTimeout(duration time.Duration) bool { panic("implement me") }

func (m *mockToken) Done() <-chan struct{} { panic("implement me") }

func (m *mockToken) Error() error { panic("implement me") }

type mockMsg struct {
	payload []byte
}

func (m *mockMsg) Duplicate() bool { panic("implement me") }

func (m *mockMsg) Qos() byte { panic("implement me") }

func (m *mockMsg) Retained() bool { panic("implement me") }

func (m *mockMsg) Topic() string { panic("implement me") }

func (m *mockMsg) MessageID() uint16 { panic("implement me") }

func (m *mockMsg) Payload() []byte {
	return m.payload
}

func (m *mockMsg) Ack() { panic("implement me") }

func (m *mockMqttClient) IsConnected() bool { return true }

func (m *mockMqttClient) IsConnectionOpen() bool { panic("implement me") }

func (m *mockMqttClient) Connect() mqtt.Token { return &mockToken{} }

func (m *mockMqttClient) Disconnect(quiesce uint) { panic("implement me") }

func (m *mockMqttClient) Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	panic("implement me")
}

func (m *mockMqttClient) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
	go callback(m, &mockMsg{payload: m.payload})
	return &mockToken{}
}

func (m *mockMqttClient) SubscribeMultiple(filters map[string]byte, callback mqtt.MessageHandler) mqtt.Token {
	panic("implement me")
}

func (m *mockMqttClient) Unsubscribe(topics ...string) mqtt.Token {
	return &mockToken{}
}

func (m *mockMqttClient) AddRoute(topic string, callback mqtt.MessageHandler) { panic("implement me") }

func (m *mockMqttClient) OptionsReader() mqtt.ClientOptionsReader { panic("implement me") }
