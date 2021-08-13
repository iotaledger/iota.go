package iotagox

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/iotaledger/hive.go/serializer"
	iotago "github.com/iotaledger/iota.go/v2"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	// NodeEventMilestonesLatest is the name of the latest milestone event channel.
	NodeEventMilestonesLatest = "milestones/latest"
	// NodeEventMilestonesConfirmed is the name of the confirmed milestone event channel.
	NodeEventMilestonesConfirmed = "milestones/confirmed"

	// NodeEventMessages is the name of the received messages event channel.
	NodeEventMessages = "messages"
	// NodeEventMessagesReferenced is the name of the referenced messages metadata event channel.
	NodeEventMessagesReferenced = "messages/referenced"
	// NodeEventMessagesIndexation is the name of the indexed messages  event channel.
	NodeEventMessagesIndexation = "messages/indexation/{index}"
	// NodeEventMessagesMetadata is the name of the message metadata event channel.
	NodeEventMessagesMetadata = "messages/{messageId}/metadata"

	// NodeEventTransactionsIncludedMessage is the name of the included transaction message event channel.
	NodeEventTransactionsIncludedMessage = "transactions/{transactionId}/included-message"

	// NodeEventOutputs is the name of the outputs event channel.
	NodeEventOutputs = "outputs/{outputId}"

	// NodeEventReceipts is the name of the receipts event channel.
	NodeEventReceipts = "receipts"

	// NodeEventAddressesOutput is the name of the address outputs event channel.
	NodeEventAddressesOutput = "addresses/{address}/outputs"
	// NodeEventAddressesEd25519Output is the name of the ed25519 address outputs event channel.
	NodeEventAddressesEd25519Output = "addresses/ed25519/{address}/outputs"
)

var (
	// ErrNodeEventAPIClientInactive gets returned when a NodeEventAPIClient is inactive.
	ErrNodeEventAPIClientInactive = errors.New("node event api client is inactive")
)

func randMQTTClientID() string {
	return strconv.FormatInt(rand.NewSource(time.Now().UnixNano()).Int63(), 10)
}

// NewNodeEventAPIClient creates a new NodeEventAPIClient using the given broker URI and default MQTT client options.
func NewNodeEventAPIClient(brokerURI string) *NodeEventAPIClient {
	clientOpts := mqtt.NewClientOptions()
	clientOpts.Order = false
	clientOpts.ClientID = randMQTTClientID()
	clientOpts.AddBroker(brokerURI)
	errChan := make(chan error)
	clientOpts.OnConnectionLost = func(client mqtt.Client, err error) { sendErrOrDrop(errChan, err) }
	return &NodeEventAPIClient{
		MQTTClient: mqtt.NewClient(clientOpts),
		Errors:     errChan,
	}
}

// NodeEventAPIClient represents a handle to retrieve channels for node events.
// Any registration will panic if the NodeEventAPIClient.Ctx is done or the client isn't connected.
// Multiple calls to the same channel registration will override the previously created channel.
type NodeEventAPIClient struct {
	MQTTClient mqtt.Client
	// The context over the EventChannelsHandle.
	Ctx context.Context
	// A channel up on which errors are returned from within subscriptions or when the connection is lost.
	// It is the instantiater's job to ensure that the respective connection handlers are linked to this error channel
	// if the client was created without NewNodeEventAPIClient.
	// Errors are dropped silently if no receiver is listening for them or can consume them fast enough.
	Errors chan error
}

func panicIfNodeEventAPIClientInactive(neac *NodeEventAPIClient) {
	if err := neac.Ctx.Err(); err != nil {
		panic(fmt.Errorf("%w: context is cancelled/done", ErrNodeEventAPIClientInactive))
	}
	if !neac.MQTTClient.IsConnected() {
		panic(fmt.Errorf("%w: client is not connected", ErrNodeEventAPIClientInactive))
	}
}

func sendErrOrDrop(errChan chan error, err error) {
	select {
	case errChan <- err:
	default:
	}
}

// Connect connects the NodeEventAPIClient to the specified brokers.
// The NodeEventAPIClient remains active as long as the given context isn't done/cancelled.
func (neac *NodeEventAPIClient) Connect(ctx context.Context) error {
	neac.Ctx = ctx
	if token := neac.MQTTClient.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// Close disconnects the underlying MQTT client.
// Call this function to clean up any registered channels.
func (neac *NodeEventAPIClient) Close() {
	neac.MQTTClient.Disconnect(0)
}

// Messages returns a channel of newly received messages.
func (neac *NodeEventAPIClient) Messages() <-chan *iotago.Message {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *iotago.Message)
	neac.MQTTClient.Subscribe(NodeEventMessages, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		msg := &iotago.Message{}
		if _, err := msg.Deserialize(mqttMsg.Payload(), serializer.DeSeriModePerformValidation); err != nil {
			sendErrOrDrop(neac.Errors, err)
			return
		}
		select {
		case <-neac.Ctx.Done():
			return
		case channel <- msg:
		}
	})
	return channel
}

// ReferencedMessagesMetadata returns a channel of message metadata of newly referenced messages.
func (neac *NodeEventAPIClient) ReferencedMessagesMetadata() <-chan *iotago.MessageMetadataResponse {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *iotago.MessageMetadataResponse)
	neac.MQTTClient.Subscribe(NodeEventMessagesReferenced, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		metadataRes := &iotago.MessageMetadataResponse{}
		if err := json.Unmarshal(mqttMsg.Payload(), metadataRes); err != nil {
			sendErrOrDrop(neac.Errors, err)
			return
		}
		select {
		case <-neac.Ctx.Done():
			return
		case channel <- metadataRes:
		}
	})
	return channel
}

// ReferencedMessages returns a channel of newly referenced messages.
func (neac *NodeEventAPIClient) ReferencedMessages(nodeHTTPAPIClient *iotago.NodeHTTPAPIClient) <-chan *iotago.Message {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *iotago.Message)
	neac.MQTTClient.Subscribe(NodeEventMessagesReferenced, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		metadataRes := &iotago.MessageMetadataResponse{}
		if err := json.Unmarshal(mqttMsg.Payload(), metadataRes); err != nil {
			sendErrOrDrop(neac.Errors, err)
			return
		}

		msg, err := nodeHTTPAPIClient.MessageByMessageID(context.Background(), iotago.MustMessageIDFromHexString(metadataRes.MessageID))
		if err != nil {
			return
		}

		select {
		case <-neac.Ctx.Done():
			return
		case channel <- msg:
		}
	})
	return channel
}

// MessagesWithIndex returns a channel of newly received messages with the given index.
func (neac *NodeEventAPIClient) MessagesWithIndex(index string) <-chan *iotago.Message {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *iotago.Message)
	neac.MQTTClient.Subscribe(strings.Replace(NodeEventMessagesIndexation, "{index}", index, 1), 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		msg := &iotago.Message{}
		if _, err := msg.Deserialize(mqttMsg.Payload(), serializer.DeSeriModePerformValidation); err != nil {
			sendErrOrDrop(neac.Errors, err)
			return
		}
		select {
		case <-neac.Ctx.Done():
			return
		case channel <- msg:
		}
	})
	return channel
}

// MessageMetadataChange returns a channel of MessageMetadataResponse each time the given message's state changes.
func (neac *NodeEventAPIClient) MessageMetadataChange(msgID iotago.MessageID) <-chan *iotago.MessageMetadataResponse {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *iotago.MessageMetadataResponse)
	topic := strings.Replace(NodeEventMessagesMetadata, "{messageId}", iotago.MessageIDToHexString(msgID), 1)
	neac.MQTTClient.Subscribe(topic, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		metadataRes := &iotago.MessageMetadataResponse{}
		if err := json.Unmarshal(mqttMsg.Payload(), metadataRes); err != nil {
			sendErrOrDrop(neac.Errors, err)
			return
		}
		select {
		case <-neac.Ctx.Done():
			return
		case channel <- metadataRes:
		}
	})
	return channel
}

// AddressOutputs returns a channel of newly created or spent outputs on the given address.
func (neac *NodeEventAPIClient) AddressOutputs(addr iotago.Address, netPrefix iotago.NetworkPrefix) <-chan *iotago.NodeOutputResponse {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *iotago.NodeOutputResponse)
	topic := strings.Replace(NodeEventAddressesOutput, "{address}", addr.Bech32(netPrefix), 1)
	neac.MQTTClient.Subscribe(topic, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		res := &iotago.NodeOutputResponse{}
		if err := json.Unmarshal(mqttMsg.Payload(), res); err != nil {
			sendErrOrDrop(neac.Errors, err)
			return
		}
		select {
		case <-neac.Ctx.Done():
			return
		case channel <- res:
		}
	})
	return channel
}

// Ed25519AddressOutputs returns a channel of newly created or spent outputs on the given ed25519 address.
func (neac *NodeEventAPIClient) Ed25519AddressOutputs(addr *iotago.Ed25519Address) <-chan *iotago.NodeOutputResponse {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *iotago.NodeOutputResponse)
	topic := strings.Replace(NodeEventAddressesEd25519Output, "{address}", addr.String(), 1)
	neac.MQTTClient.Subscribe(topic, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		res := &iotago.NodeOutputResponse{}
		if err := json.Unmarshal(mqttMsg.Payload(), res); err != nil {
			sendErrOrDrop(neac.Errors, err)
			return
		}
		select {
		case <-neac.Ctx.Done():
			return
		case channel <- res:
		}
	})
	return channel
}

// TransactionIncludedMessage returns a channel of the included message which carries the transaction with the given ID.
func (neac *NodeEventAPIClient) TransactionIncludedMessage(txID iotago.TransactionID) <-chan *iotago.Message {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *iotago.Message)
	topic := strings.Replace(NodeEventTransactionsIncludedMessage, "{transactionId}", iotago.MessageIDToHexString(txID), 1)
	neac.MQTTClient.Subscribe(topic, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		msg := &iotago.Message{}
		if _, err := msg.Deserialize(mqttMsg.Payload(), serializer.DeSeriModePerformValidation); err != nil {
			sendErrOrDrop(neac.Errors, err)
			return
		}
		select {
		case <-neac.Ctx.Done():
			return
		case channel <- msg:
		}
	})
	return channel
}

// Output returns a channel which immediately returns the output with the given ID and afterwards when its state changes.
func (neac *NodeEventAPIClient) Output(outputID iotago.UTXOInputID) <-chan *iotago.NodeOutputResponse {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *iotago.NodeOutputResponse)
	topic := strings.Replace(NodeEventOutputs, "{outputId}", hex.EncodeToString(outputID[:]), 1)
	neac.MQTTClient.Subscribe(topic, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		res := &iotago.NodeOutputResponse{}
		if err := json.Unmarshal(mqttMsg.Payload(), res); err != nil {
			sendErrOrDrop(neac.Errors, err)
			return
		}
		select {
		case <-neac.Ctx.Done():
			return
		case channel <- res:
		}
	})
	return channel
}

// Receipts returns a channel which returns newly applied receipts.
func (neac *NodeEventAPIClient) Receipts() <-chan *iotago.Receipt {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *iotago.Receipt)
	neac.MQTTClient.Subscribe(NodeEventReceipts, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		receipt := &iotago.Receipt{}
		if err := json.Unmarshal(mqttMsg.Payload(), receipt); err != nil {
			sendErrOrDrop(neac.Errors, err)
			return
		}
		select {
		case <-neac.Ctx.Done():
			return
		case channel <- receipt:
		}
	})
	return channel
}

// MilestonePointer is an informative struct holding a milestone index and timestamp.
type MilestonePointer struct {
	Index     uint32 `json:"index"`
	Timestamp uint64 `json:"timestamp"`
}

// LatestMilestones returns a channel of newly seen latest milestones.
func (neac *NodeEventAPIClient) LatestMilestones() <-chan *MilestonePointer {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *MilestonePointer)
	neac.MQTTClient.Subscribe(NodeEventMilestonesLatest, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		msPointer := &MilestonePointer{}
		if err := json.Unmarshal(mqttMsg.Payload(), msPointer); err != nil {
			sendErrOrDrop(neac.Errors, err)
			return
		}
		select {
		case <-neac.Ctx.Done():
			return
		case channel <- msPointer:
		}
	})
	return channel
}

// LatestMilestoneMessages returns a channel of newly seen latest milestones messages.
func (neac *NodeEventAPIClient) LatestMilestoneMessages(nodeHTTPAPIClient *iotago.NodeHTTPAPIClient) <-chan *iotago.Message {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *iotago.Message)
	neac.MQTTClient.Subscribe(NodeEventMilestonesLatest, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		msPointer := &MilestonePointer{}
		if err := json.Unmarshal(mqttMsg.Payload(), msPointer); err != nil {
			sendErrOrDrop(neac.Errors, err)
			return
		}
		res, err := nodeHTTPAPIClient.MilestoneByIndex(context.Background(), msPointer.Index)
		if err != nil {
			sendErrOrDrop(neac.Errors, err)
			return
		}
		msg, err := nodeHTTPAPIClient.MessageByMessageID(context.Background(), iotago.MustMessageIDFromHexString(res.MessageID))
		if err != nil {
			sendErrOrDrop(neac.Errors, err)
			return
		}
		select {
		case <-neac.Ctx.Done():
			return
		case channel <- msg:
		}
	})
	return channel
}

// ConfirmedMilestones returns a channel of newly confirmed milestones.
func (neac *NodeEventAPIClient) ConfirmedMilestones() <-chan *MilestonePointer {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *MilestonePointer)
	neac.MQTTClient.Subscribe(NodeEventMilestonesConfirmed, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		msPointer := &MilestonePointer{}
		if err := json.Unmarshal(mqttMsg.Payload(), msPointer); err != nil {
			sendErrOrDrop(neac.Errors, err)
			return
		}
		select {
		case <-neac.Ctx.Done():
			return
		case channel <- msPointer:
		}
	})
	return channel
}

// ConfirmedMilestoneMessages returns a channel of newly confirmed milestones messages.
func (neac *NodeEventAPIClient) ConfirmedMilestoneMessages(nodeHTTPAPIClient *iotago.NodeHTTPAPIClient) <-chan *iotago.Message {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *iotago.Message)
	neac.MQTTClient.Subscribe(NodeEventMilestonesConfirmed, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		msPointer := &MilestonePointer{}
		if err := json.Unmarshal(mqttMsg.Payload(), msPointer); err != nil {
			sendErrOrDrop(neac.Errors, err)
			return
		}
		res, err := nodeHTTPAPIClient.MilestoneByIndex(context.Background(), msPointer.Index)
		if err != nil {
			sendErrOrDrop(neac.Errors, err)
			return
		}
		msg, err := nodeHTTPAPIClient.MessageByMessageID(context.Background(), iotago.MustMessageIDFromHexString(res.MessageID))
		if err != nil {
			sendErrOrDrop(neac.Errors, err)
			return
		}
		select {
		case <-neac.Ctx.Done():
			return
		case channel <- msg:
		}
	})
	return channel
}
