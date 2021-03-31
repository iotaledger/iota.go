package iotago

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

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
	// ErrNodeEventChannelsHandleInactive gets returned when a EventChannelsHandle is inactive.
	ErrNodeEventChannelsHandleInactive = errors.New("event channels handle is inactive")
)

// NewNodeEventAPIClient creates a new NodeEventAPIClient using the given broker URI and default MQTT client options.
func NewNodeEventAPIClient(brokerURI string) *NodeEventAPIClient {
	clientOpts := mqtt.NewClientOptions()
	clientOpts.Order = false
	clientOpts.AddBroker(brokerURI)
	return &NodeEventAPIClient{MQTTClient: mqtt.NewClient(clientOpts)}
}

// NodeEventAPIClient is a client for node events.
type NodeEventAPIClient struct {
	MQTTClient mqtt.Client
}

// EventChannelsHandle represents a handle to retrieve channels for events.
// Any registration will panic if the EventChannelsHandle.Ctx is done.
// Multiple calls to the same non parameterized registration will override the previously created channel.
type EventChannelsHandle struct {
	// The context over the EventChannelsHandle.
	Ctx        context.Context
	mqttClient mqtt.Client
}

func panicIfEventChannelsHandleInactive(ech *EventChannelsHandle) {
	if err := ech.Ctx.Err(); err != nil {
		panic(ErrNodeEventChannelsHandleInactive)
	}
}

// Connect connects to the node event API and returns a handle to subscribe to events.
// The returned EventChannelsHandle handle remains active as long as the given context isn't done/cancelled.
func (nea *NodeEventAPIClient) Connect(ctx context.Context) (*EventChannelsHandle, error) {
	if token := nea.MQTTClient.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return &EventChannelsHandle{mqttClient: nea.MQTTClient, Ctx: ctx}, nil
}

// Disconnect disconnects the underlying MQTT client.
// Call this function to clean up any registered channels.
func (ech *EventChannelsHandle) Disconnect() {
	ech.mqttClient.Disconnect(0)
}

// Messages returns a channel of newly received messages.
func (ech *EventChannelsHandle) Messages() <-chan *Message {
	panicIfEventChannelsHandleInactive(ech)
	channel := make(chan *Message)
	ech.mqttClient.Subscribe(NodeEventMessages, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		msg := &Message{}
		if _, err := msg.Deserialize(mqttMsg.Payload(), DeSeriModePerformValidation); err != nil {
			return
		}
		select {
		case <-ech.Ctx.Done():
			return
		case channel <- msg:
		}
	})
	return channel
}

// ReferencedMessagesMetadata returns a channel of message metadata of newly referenced messages.
func (ech *EventChannelsHandle) ReferencedMessagesMetadata() <-chan *MessageMetadataResponse {
	panicIfEventChannelsHandleInactive(ech)
	channel := make(chan *MessageMetadataResponse)
	ech.mqttClient.Subscribe(NodeEventMessagesReferenced, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		metadataRes := &MessageMetadataResponse{}
		if err := json.Unmarshal(mqttMsg.Payload(), metadataRes); err != nil {
			return
		}
		select {
		case <-ech.Ctx.Done():
			return
		case channel <- metadataRes:
		}
	})
	return channel
}

// ReferencedMessages returns a channel of newly referenced messages.
func (ech *EventChannelsHandle) ReferencedMessages(nodeHTTPAPIClient *NodeHTTPAPIClient) <-chan *Message {
	panicIfEventChannelsHandleInactive(ech)
	channel := make(chan *Message)
	ech.mqttClient.Subscribe(NodeEventMessagesReferenced, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		metadataRes := &MessageMetadataResponse{}
		if err := json.Unmarshal(mqttMsg.Payload(), metadataRes); err != nil {
			return
		}

		msg, err := nodeHTTPAPIClient.MessageByMessageID(MustMessageIDFromHexString(metadataRes.MessageID))
		if err != nil {
			return
		}

		select {
		case <-ech.Ctx.Done():
			return
		case channel <- msg:
		}
	})
	return channel
}

// MessagesWithIndex returns a channel of newly received messages with the given index.
func (ech *EventChannelsHandle) MessagesWithIndex(index string) <-chan *Message {
	panicIfEventChannelsHandleInactive(ech)
	channel := make(chan *Message)
	ech.mqttClient.Subscribe(strings.Replace(NodeEventMessagesIndexation, "{index}", index, 1), 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		msg := &Message{}
		if _, err := msg.Deserialize(mqttMsg.Payload(), DeSeriModePerformValidation); err != nil {
			return
		}
		select {
		case <-ech.Ctx.Done():
			return
		case channel <- msg:
		}
	})
	return channel
}

// MessageMetadataChange returns a channel of MessageMetadataResponse each time the given message's state changes.
func (ech *EventChannelsHandle) MessageMetadataChange(msgID MessageID) <-chan *MessageMetadataResponse {
	panicIfEventChannelsHandleInactive(ech)
	channel := make(chan *MessageMetadataResponse)
	topic := strings.Replace(NodeEventMessagesMetadata, "{messageId}", MessageIDToHexString(msgID), 1)
	ech.mqttClient.Subscribe(topic, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		metadataRes := &MessageMetadataResponse{}
		if err := json.Unmarshal(mqttMsg.Payload(), metadataRes); err != nil {
			return
		}
		select {
		case <-ech.Ctx.Done():
			return
		case channel <- metadataRes:
		}
	})
	return channel
}

// AddressOutputs returns a channel of newly created or spent outputs on the given address.
func (ech *EventChannelsHandle) AddressOutputs(addr Address, netPrefix NetworkPrefix) <-chan *NodeOutputResponse {
	panicIfEventChannelsHandleInactive(ech)
	channel := make(chan *NodeOutputResponse)
	topic := strings.Replace(NodeEventAddressesOutput, "{address}", addr.Bech32(netPrefix), 1)
	ech.mqttClient.Subscribe(topic, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		res := &NodeOutputResponse{}
		if err := json.Unmarshal(mqttMsg.Payload(), res); err != nil {
			return
		}
		select {
		case <-ech.Ctx.Done():
			return
		case channel <- res:
		}
	})
	return channel
}

// Ed25519AddressOutputs returns a channel of newly created or spent outputs on the given ed25519 address.
func (ech *EventChannelsHandle) Ed25519AddressOutputs(addr *Ed25519Address) <-chan *NodeOutputResponse {
	panicIfEventChannelsHandleInactive(ech)
	channel := make(chan *NodeOutputResponse)
	topic := strings.Replace(NodeEventAddressesEd25519Output, "{address}", addr.String(), 1)
	ech.mqttClient.Subscribe(topic, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		res := &NodeOutputResponse{}
		if err := json.Unmarshal(mqttMsg.Payload(), res); err != nil {
			return
		}
		select {
		case <-ech.Ctx.Done():
			return
		case channel <- res:
		}
	})
	return channel
}

// TransactionIncludedMessage returns a channel of the included message which carries the transaction with the given ID.
func (ech *EventChannelsHandle) TransactionIncludedMessage(txID TransactionID) <-chan *Message {
	panicIfEventChannelsHandleInactive(ech)
	channel := make(chan *Message)
	topic := strings.Replace(NodeEventTransactionsIncludedMessage, "{transactionId}", MessageIDToHexString(txID), 1)
	ech.mqttClient.Subscribe(topic, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		msg := &Message{}
		if _, err := msg.Deserialize(mqttMsg.Payload(), DeSeriModePerformValidation); err != nil {
			return
		}
		select {
		case <-ech.Ctx.Done():
			return
		case channel <- msg:
		}
	})
	return channel
}

// Output returns a channel which immediately returns the output with the given ID and afterwards when its state changes.
func (ech *EventChannelsHandle) Output(outputID UTXOInputID) <-chan *NodeOutputResponse {
	panicIfEventChannelsHandleInactive(ech)
	channel := make(chan *NodeOutputResponse)
	topic := strings.Replace(NodeEventOutputs, "{outputId}", hex.EncodeToString(outputID[:]), 1)
	ech.mqttClient.Subscribe(topic, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		res := &NodeOutputResponse{}
		if err := json.Unmarshal(mqttMsg.Payload(), res); err != nil {
			return
		}
		select {
		case <-ech.Ctx.Done():
			return
		case channel <- res:
		}
	})
	return channel
}

// Receipts returns a channel which returns newly applied receipts.
func (ech *EventChannelsHandle) Receipts() <-chan *Receipt {
	panicIfEventChannelsHandleInactive(ech)
	channel := make(chan *Receipt)
	ech.mqttClient.Subscribe(NodeEventReceipts, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		receipt := &Receipt{}
		if err := json.Unmarshal(mqttMsg.Payload(), receipt); err != nil {
			return
		}
		select {
		case <-ech.Ctx.Done():
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
func (ech *EventChannelsHandle) LatestMilestones() <-chan *MilestonePointer {
	panicIfEventChannelsHandleInactive(ech)
	channel := make(chan *MilestonePointer)
	ech.mqttClient.Subscribe(NodeEventMilestonesLatest, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		msPointer := &MilestonePointer{}
		if err := json.Unmarshal(mqttMsg.Payload(), msPointer); err != nil {
			return
		}
		select {
		case <-ech.Ctx.Done():
			return
		case channel <- msPointer:
		}
	})
	return channel
}

// LatestMilestoneMessages returns a channel of newly seen latest milestones messages.
func (ech *EventChannelsHandle) LatestMilestoneMessages(nodeHTTPAPIClient *NodeHTTPAPIClient) <-chan *Message {
	panicIfEventChannelsHandleInactive(ech)
	channel := make(chan *Message)
	ech.mqttClient.Subscribe(NodeEventMilestonesLatest, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		msPointer := &MilestonePointer{}
		if err := json.Unmarshal(mqttMsg.Payload(), msPointer); err != nil {
			return
		}
		res, err := nodeHTTPAPIClient.MilestoneByIndex(msPointer.Index)
		if err != nil {
			return
		}
		msg, err := nodeHTTPAPIClient.MessageByMessageID(MustMessageIDFromHexString(res.MessageID))
		if err != nil {
			return
		}
		select {
		case <-ech.Ctx.Done():
			return
		case channel <- msg:
		}
	})
	return channel
}

// ConfirmedMilestones returns a channel of newly confirmed milestones.
func (ech *EventChannelsHandle) ConfirmedMilestones() <-chan *MilestonePointer {
	panicIfEventChannelsHandleInactive(ech)
	channel := make(chan *MilestonePointer)
	ech.mqttClient.Subscribe(NodeEventMilestonesConfirmed, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		msPointer := &MilestonePointer{}
		if err := json.Unmarshal(mqttMsg.Payload(), msPointer); err != nil {
			return
		}
		select {
		case <-ech.Ctx.Done():
			return
		case channel <- msPointer:
		}
	})
	return channel
}

// ConfirmedMilestoneMessages returns a channel of newly confirmed milestones messages.
func (ech *EventChannelsHandle) ConfirmedMilestoneMessages(nodeHTTPAPIClient *NodeHTTPAPIClient) <-chan *Message {
	panicIfEventChannelsHandleInactive(ech)
	channel := make(chan *Message)
	ech.mqttClient.Subscribe(NodeEventMilestonesLatest, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		msPointer := &MilestonePointer{}
		if err := json.Unmarshal(mqttMsg.Payload(), msPointer); err != nil {
			return
		}
		res, err := nodeHTTPAPIClient.MilestoneByIndex(msPointer.Index)
		if err != nil {
			return
		}
		msg, err := nodeHTTPAPIClient.MessageByMessageID(MustMessageIDFromHexString(res.MessageID))
		if err != nil {
			return
		}
		select {
		case <-ech.Ctx.Done():
			return
		case channel <- msg:
		}
	})
	return channel
}
