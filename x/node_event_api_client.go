package iotagox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"

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
	// NodeEventOutputsByUnlockConditionAndAddress is the name of the outputs by unlock condition address event channel.
	NodeEventOutputsByUnlockConditionAndAddress = "outputs/unlock/{condition}/{address}"
	// NodeEventSpentOutputsByUnlockConditionAndAddress is the name of the spent outputs by unlock condition address event channel.
	NodeEventSpentOutputsByUnlockConditionAndAddress = "outputs/unlock/{condition}/{address}/spent"

	// NodeEventReceipts is the name of the receipts event channel.
	NodeEventReceipts = "receipts"
)

var (
	// ErrNodeEventAPIClientInactive gets returned when a NodeEventAPIClient is inactive.
	ErrNodeEventAPIClientInactive = errors.New("node event api client is inactive")
)

// NodeEventUnlockCondition denotes the different unlock conditions.
type NodeEventUnlockCondition string

// Unlock conditions.
const (
	UnlockConditionAny              NodeEventUnlockCondition = "+"
	UnlockConditionAddress          NodeEventUnlockCondition = "address"
	UnlockConditionStorageReturn    NodeEventUnlockCondition = "storageReturn"
	UnlockConditionExpirationReturn NodeEventUnlockCondition = "expirationReturn"
	UnlockConditionStateController  NodeEventUnlockCondition = "stateController"
	UnlockConditionGovernor         NodeEventUnlockCondition = "governor"
	UnlockConditionImmutableAlias   NodeEventUnlockCondition = "immutableAlias"
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
		if _, err := msg.Deserialize(mqttMsg.Payload(), serializer.DeSeriModeNoValidation, nil); err != nil {
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
func (neac *NodeEventAPIClient) ReferencedMessagesMetadata() (<-chan *nodeclient.MessageMetadataResponse, error) {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *nodeclient.MessageMetadataResponse)
	if token := neac.MQTTClient.Subscribe(NodeEventMessagesReferenced, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		metadataRes := &nodeclient.MessageMetadataResponse{}
		if err := json.Unmarshal(mqttMsg.Payload(), metadataRes); err != nil {
			sendErrOrDrop(neac.Errors, err)
			return
		}
		select {
		case <-neac.Ctx.Done():
			return
		case channel <- metadataRes:
		}
	}); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return channel, nil
}

// ReferencedMessages returns a channel of newly referenced messages.
func (neac *NodeEventAPIClient) ReferencedMessages(nodeHTTPAPIClient *nodeclient.Client) (<-chan *iotago.Message, error) {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *iotago.Message)
	if token := neac.MQTTClient.Subscribe(NodeEventMessagesReferenced, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		metadataRes := &nodeclient.MessageMetadataResponse{}
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
	}); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return channel, nil
}

// MessagesWithIndex returns a channel of newly received messages with the given index.
func (neac *NodeEventAPIClient) MessagesWithIndex(index string) (<-chan *iotago.Message, error) {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *iotago.Message)
	if token := neac.MQTTClient.Subscribe(strings.Replace(NodeEventMessagesIndexation, "{index}", index, 1), 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		msg := &iotago.Message{}
		if _, err := msg.Deserialize(mqttMsg.Payload(), serializer.DeSeriModeNoValidation, nil); err != nil {
			sendErrOrDrop(neac.Errors, err)
			return
		}
		select {
		case <-neac.Ctx.Done():
			return
		case channel <- msg:
		}
	}); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return channel, nil
}

// MessageMetadataChange returns a channel of MessageMetadataResponse each time the given message's state changes.
func (neac *NodeEventAPIClient) MessageMetadataChange(msgID iotago.MessageID) (<-chan *nodeclient.MessageMetadataResponse, error) {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *nodeclient.MessageMetadataResponse)
	topic := strings.Replace(NodeEventMessagesMetadata, "{messageId}", iotago.MessageIDToHexString(msgID), 1)
	if token := neac.MQTTClient.Subscribe(topic, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		metadataRes := &nodeclient.MessageMetadataResponse{}
		if err := json.Unmarshal(mqttMsg.Payload(), metadataRes); err != nil {
			sendErrOrDrop(neac.Errors, err)
			return
		}
		select {
		case <-neac.Ctx.Done():
			return
		case channel <- metadataRes:
		}
	}); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return channel, nil
}

// OutputsByUnlockConditionAndAddress returns a channel of newly created outputs on the given unlock condition and address.
func (neac *NodeEventAPIClient) OutputsByUnlockConditionAndAddress(addr iotago.Address, netPrefix iotago.NetworkPrefix, condition NodeEventUnlockCondition) (<-chan *nodeclient.OutputResponse, error) {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *nodeclient.OutputResponse)
	topic := strings.Replace(NodeEventOutputsByUnlockConditionAndAddress, "{address}", addr.Bech32(netPrefix), 1)
	topic = strings.Replace(topic, "{condition}", string(condition), 1)
	if token := neac.MQTTClient.Subscribe(topic, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		res := &nodeclient.OutputResponse{}
		if err := json.Unmarshal(mqttMsg.Payload(), res); err != nil {
			sendErrOrDrop(neac.Errors, err)
			return
		}
		select {
		case <-neac.Ctx.Done():
			return
		case channel <- res:
		}
	}); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return channel, nil
}

// SpentOutputsByUnlockConditionAndAddress returns a channel of newly spent outputs on the given unlock condition and address.
func (neac *NodeEventAPIClient) SpentOutputsByUnlockConditionAndAddress(addr iotago.Address, netPrefix iotago.NetworkPrefix, condition NodeEventUnlockCondition) (<-chan *nodeclient.OutputResponse, error) {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *nodeclient.OutputResponse)
	topic := strings.Replace(NodeEventSpentOutputsByUnlockConditionAndAddress, "{address}", addr.Bech32(netPrefix), 1)
	topic = strings.Replace(topic, "{condition}", string(condition), 1)
	if token := neac.MQTTClient.Subscribe(topic, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		res := &nodeclient.OutputResponse{}
		if err := json.Unmarshal(mqttMsg.Payload(), res); err != nil {
			sendErrOrDrop(neac.Errors, err)
			return
		}
		select {
		case <-neac.Ctx.Done():
			return
		case channel <- res:
		}
	}); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return channel, nil
}

// TransactionIncludedMessage returns a channel of the included message which carries the transaction with the given ID.
func (neac *NodeEventAPIClient) TransactionIncludedMessage(txID iotago.TransactionID) (<-chan *iotago.Message, error) {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *iotago.Message)
	topic := strings.Replace(NodeEventTransactionsIncludedMessage, "{transactionId}", iotago.MessageIDToHexString(txID), 1)
	if token := neac.MQTTClient.Subscribe(topic, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		msg := &iotago.Message{}
		if _, err := msg.Deserialize(mqttMsg.Payload(), serializer.DeSeriModePerformValidation, nil); err != nil {
			sendErrOrDrop(neac.Errors, err)
			return
		}
		select {
		case <-neac.Ctx.Done():
			return
		case channel <- msg:
		}
	}); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return channel, nil
}

// Output returns a channel which immediately returns the output with the given ID and afterwards when its state changes.
func (neac *NodeEventAPIClient) Output(outputID iotago.OutputID) (<-chan *nodeclient.OutputResponse, error) {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *nodeclient.OutputResponse)
	topic := strings.Replace(NodeEventOutputs, "{outputId}", iotago.EncodeHex(outputID[:]), 1)
	if token := neac.MQTTClient.Subscribe(topic, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		res := &nodeclient.OutputResponse{}
		if err := json.Unmarshal(mqttMsg.Payload(), res); err != nil {
			sendErrOrDrop(neac.Errors, err)
			return
		}
		select {
		case <-neac.Ctx.Done():
			return
		case channel <- res:
		}
	}); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return channel, nil
}

// Receipts returns a channel which returns newly applied receipts.
func (neac *NodeEventAPIClient) Receipts() (<-chan *iotago.Receipt, error) {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *iotago.Receipt)
	if token := neac.MQTTClient.Subscribe(NodeEventReceipts, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
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
	}); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return channel, nil
}

// MilestonePointer is an informative struct holding a milestone index and timestamp.
type MilestonePointer struct {
	Index     uint32 `json:"index"`
	Timestamp uint64 `json:"timestamp"`
}

// LatestMilestones returns a channel of newly seen latest milestones.
func (neac *NodeEventAPIClient) LatestMilestones() (<-chan *MilestonePointer, error) {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *MilestonePointer)
	if token := neac.MQTTClient.Subscribe(NodeEventMilestonesLatest, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
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
	}); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return channel, nil
}

// LatestMilestoneMessages returns a channel of newly seen latest milestones messages.
func (neac *NodeEventAPIClient) LatestMilestoneMessages(nodeHTTPAPIClient *nodeclient.Client) (<-chan *iotago.Message, error) {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *iotago.Message)
	if token := neac.MQTTClient.Subscribe(NodeEventMilestonesLatest, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
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
	}); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return channel, nil
}

// ConfirmedMilestones returns a channel of newly confirmed milestones.
func (neac *NodeEventAPIClient) ConfirmedMilestones() (<-chan *MilestonePointer, error) {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *MilestonePointer)
	if token := neac.MQTTClient.Subscribe(NodeEventMilestonesConfirmed, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
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
	}); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return channel, nil
}

// ConfirmedMilestoneMessages returns a channel of newly confirmed milestones messages.
func (neac *NodeEventAPIClient) ConfirmedMilestoneMessages(nodeHTTPAPIClient *nodeclient.Client) (<-chan *iotago.Message, error) {
	panicIfNodeEventAPIClientInactive(neac)
	channel := make(chan *iotago.Message)
	if token := neac.MQTTClient.Subscribe(NodeEventMilestonesConfirmed, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
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
	}); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return channel, nil
}
