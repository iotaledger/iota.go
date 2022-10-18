package nodeclient_test

import (
	"context"
	"math/rand"
	"os"
	"testing"
	"time"

	"gopkg.in/h2non/gock.v1"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())

	// call the tests
	os.Exit(m.Run())
}

func Test_EventAPIEnabled(t *testing.T) {
	defer gock.Off()

	originRoutes := &nodeclient.RoutesResponse{
		Routes: []string{"mqtt/v1"},
	}

	gock.New(nodeAPIUrl).
		Get(nodeclient.RouteRoutes).
		Reply(200).
		JSON(originRoutes)

	client := nodeclient.New(nodeAPIUrl)

	_, err := client.EventAPI(context.TODO())
	require.NoError(t, err)
}

func Test_EventAPIDisabled(t *testing.T) {
	defer gock.Off()

	originRoutes := &nodeclient.RoutesResponse{
		Routes: []string{"someplugin/v1"},
	}

	gock.New(nodeAPIUrl).
		Get(nodeclient.RouteRoutes).
		Reply(200).
		JSON(originRoutes)

	client := nodeclient.New(nodeAPIUrl)

	_, err := client.EventAPI(context.TODO())
	require.ErrorIs(t, err, nodeclient.ErrMQTTPluginNotAvailable)
}

func Test_NewEventAPIClient(t *testing.T) {

	block := tpkg.RandBlock(iotago.PayloadTaggedData)
	originBlockBytes, err := v2API.Encode(block)
	require.NoError(t, err)
	mock := &mockMqttClient{payload: originBlockBytes}
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	eventAPIClient := &nodeclient.EventAPIClient{
		Client:     nodeclient.New(nodeAPIUrl),
		MQTTClient: mock,
		Errors:     make(chan error),
	}
	require.NoError(t, eventAPIClient.Connect(ctx))

	blockChan, sub := eventAPIClient.Blocks()
	require.NoError(t, sub.Error())
	require.Eventually(t, func() bool {
		select {
		case recBlock := <-blockChan:
			gottenBlockBytes, err := v2API.Encode(recBlock)
			require.NoError(t, err)
			require.Equal(t, originBlockBytes, gottenBlockBytes)

			require.NoError(t, sub.Close())
			require.NoError(t, sub.Close())

			return true
		default:
			return false
		}
	}, 5*time.Second, 100*time.Millisecond)
}

type mockMqttClient struct {
	payload []byte
}

type mockToken struct{}

func (m *mockToken) Wait() bool {
	return false
}

func (m *mockToken) WaitTimeout(_ time.Duration) bool { panic("implement me") }

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
