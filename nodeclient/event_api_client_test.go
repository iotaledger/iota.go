package nodeclient_test

import (
	"context"
	"os"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/api"
	"github.com/iotaledger/iota.go/v4/nodeclient"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestMain(m *testing.M) { // call the tests
	os.Exit(m.Run())
}

func Test_EventAPIEnabled(t *testing.T) {
	defer gock.Off()

	originRoutes := &api.RoutesResponse{
		Routes: []string{api.MQTTPluginName},
	}
	mockGetJSON(api.RouteRoutes, 200, originRoutes)

	_, err := nodeClient(t).EventAPI(context.TODO())
	require.NoError(t, err)
}

func Test_EventAPIDisabled(t *testing.T) {
	defer gock.Off()

	originRoutes := &api.RoutesResponse{
		Routes: []string{"someplugin/v1"},
	}

	mockGetJSON(api.RouteRoutes, 200, originRoutes)

	_, err := nodeClient(t).EventAPI(context.TODO())
	require.ErrorIs(t, err, nodeclient.ErrMQTTPluginNotAvailable)
}

func Test_NewEventAPIClient(t *testing.T) {
	block := tpkg.RandBlock(tpkg.RandBasicBlock(tpkg.TestAPI, iotago.PayloadTaggedData), tpkg.TestAPI, 0)
	originBlockBytes, err := tpkg.TestAPI.Encode(block)
	require.NoError(t, err)
	mock := &mockMqttClient{payload: originBlockBytes}
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	eventAPIClient := &nodeclient.EventAPIClient{
		Client:     nodeClient(t),
		MQTTClient: mock,
		Errors:     make(chan error),
	}
	require.NoError(t, eventAPIClient.Connect(ctx))

	blockChan, sub := eventAPIClient.Blocks()
	require.NoError(t, sub.Error())
	require.Eventually(t, func() bool {
		select {
		case recBlock := <-blockChan:
			gottenBlockBytes, err := tpkg.TestAPI.Encode(recBlock)
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

func (m *mockMqttClient) Disconnect(_ uint) { panic("implement me") }

func (m *mockMqttClient) Publish(_ string, _ byte, _ bool, _ interface{}) mqtt.Token {
	panic("implement me")
}

func (m *mockMqttClient) Subscribe(_ string, _ byte, callback mqtt.MessageHandler) mqtt.Token {
	go callback(m, &mockMsg{payload: m.payload})

	return &mockToken{}
}

func (m *mockMqttClient) SubscribeMultiple(_ map[string]byte, _ mqtt.MessageHandler) mqtt.Token {
	panic("implement me")
}

func (m *mockMqttClient) Unsubscribe(...string) mqtt.Token {
	return &mockToken{}
}

func (m *mockMqttClient) AddRoute(_ string, _ mqtt.MessageHandler) { panic("implement me") }

func (m *mockMqttClient) OptionsReader() mqtt.ClientOptionsReader { panic("implement me") }
