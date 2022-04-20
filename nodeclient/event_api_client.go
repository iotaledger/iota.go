package nodeclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
)

const (
	// EventAPIMilestoneInfoLatest is the name of the latest milestone info event channel.
	EventAPIMilestoneInfoLatest = "milestone-info/latest"
	// EventAPIMilestoneInfoConfirmed is the name of the confirmed milestone info event channel.
	EventAPIMilestoneInfoConfirmed = "milestone-info/confirmed"
	// EventAPIMilestones is the name of the milestone event channel.
	EventAPIMilestones = "milestones"

	// EventAPIMessages is the name of the received messages event channel.
	EventAPIMessages = "messages"
	// EventAPIMessagesTransaction is the name of the messages containing transactions event channel.
	EventAPIMessagesTransaction = "messages/transaction"
	// EventAPIMessagesTransactionTaggedData is the name of the messages containing transaction with tagged data event channel.
	EventAPIMessagesTransactionTaggedData = "messages/transaction/tagged-data"
	// EventAPIMessagesTransactionTaggedDataTag is the name of the messages containing transaction with a specific tagged data event channel.
	EventAPIMessagesTransactionTaggedDataTag = "messages/transaction/tagged-data/{tag}"
	// EventAPIMessagesTaggedData is the name of the messages containing tagged data event channel.
	EventAPIMessagesTaggedData = "messages/tagged-data"
	// EventAPIMessagesTaggedDataTag is the name of the messages containing a specific tagged data event channel.
	EventAPIMessagesTaggedDataTag = "messages/tagged-data/{tag}"

	// EventAPITransactionsIncludedMessage is the name of the included transaction message event channel.
	EventAPITransactionsIncludedMessage = "transactions/{transactionId}/included-message"

	// EventAPIMessageMetadata is the name of the message metadata event channel.
	EventAPIMessageMetadata = "message-metadata/{messageId}"
	// EventAPIMessageMetadataReferenced is the name of the referenced messages metadata event channel.
	EventAPIMessageMetadataReferenced = "message-metadata/referenced"

	// EventAPIOutputs is the name of the outputs event channel.
	EventAPIOutputs = "outputs/{outputId}"
	// EventAPINFTOutputs is the name of the NFT output event channel to retrieve NFT mutations by their ID.
	EventAPINFTOutputs = "outputs/nfts/{nftId}"
	// EventAPIAliasOutputs is the name of the Alias output event channel to retrieve Alias mutations by their ID.
	EventAPIAliasOutputs = "outputs/aliases/{aliasId}"
	// EventAPIFoundryOutputs is the name of the Foundry output event channel to retrieve Foundry mutations by their ID.
	EventAPIFoundryOutputs = "outputs/foundries/{foundryId}"
	// EventAPIOutputsByUnlockConditionAndAddress is the name of the outputs by unlock condition address event channel.
	EventAPIOutputsByUnlockConditionAndAddress = "outputs/unlock/{condition}/{address}"
	// EventAPISpentOutputsByUnlockConditionAndAddress is the name of the spent outputs by unlock condition address event channel.
	EventAPISpentOutputsByUnlockConditionAndAddress = "outputs/unlock/{condition}/{address}/spent"

	// EventAPIReceipts is the name of the receipts event channel.
	EventAPIReceipts = "receipts"
)

var (
	// ErrEventAPIClientInactive gets returned when an EventAPIClient is inactive.
	ErrEventAPIClientInactive = errors.New("event api client is inactive")
)

// EventAPIUnlockCondition denotes the different unlock conditions.
type EventAPIUnlockCondition string

// Unlock conditions.
const (
	UnlockConditionAny              EventAPIUnlockCondition = "+"
	UnlockConditionAddress          EventAPIUnlockCondition = "address"
	UnlockConditionStorageReturn    EventAPIUnlockCondition = "storage-return"
	UnlockConditionExpirationReturn EventAPIUnlockCondition = "expiration"
	UnlockConditionStateController  EventAPIUnlockCondition = "state-controller"
	UnlockConditionGovernor         EventAPIUnlockCondition = "governor"
	UnlockConditionImmutableAlias   EventAPIUnlockCondition = "immutable-alias"
)

func randMQTTClientID() string {
	return strconv.FormatInt(rand.NewSource(time.Now().UnixNano()).Int63(), 10)
}

func brokerURLFromClient(nc *Client) string {
	baseURL := nc.BaseURL
	baseURL = strings.Replace(baseURL, "https://", "wss://", 1)
	baseURL = strings.Replace(baseURL, "http://", "ws://", 1)
	return fmt.Sprintf("%s/api/plugins/%s", baseURL, MQTTPluginName)
}

func newEventAPIClient(nc *Client) *EventAPIClient {
	clientOpts := mqtt.NewClientOptions()
	clientOpts.Order = false
	clientOpts.ClientID = randMQTTClientID()
	clientOpts.AddBroker(brokerURLFromClient(nc))
	errChan := make(chan error)
	clientOpts.OnConnectionLost = func(client mqtt.Client, err error) { sendErrOrDrop(errChan, err) }
	return &EventAPIClient{
		Client:     nc,
		MQTTClient: mqtt.NewClient(clientOpts),
		Errors:     errChan,
	}
}

// EventAPIClient represents a handle to retrieve channels for node events.
// Any registration will panic if the EventAPIClient.Ctx is done or the client isn't connected.
// Multiple calls to the same channel registration will override the previously created channel.
type EventAPIClient struct {
	Client *Client

	MQTTClient mqtt.Client
	// The context over the EventChannelsHandle.
	Ctx context.Context
	// A channel up on which errors are returned from within subscriptions or when the connection is lost.
	// Errors are dropped silently if no receiver is listening for them or can consume them fast enough.
	Errors chan error
}

// EventAPIClientSubscription holds any error that happened when trying to subscribe to an event.
// It also allows to close the subscription to cleanly unsubscribe from the node.
type EventAPIClientSubscription struct {
	mqttClient mqtt.Client
	topic      string
	error      error
}

func newSubscription(client mqtt.Client, topic string) *EventAPIClientSubscription {
	return &EventAPIClientSubscription{
		mqttClient: client,
		topic:      topic,
	}
}

func newSubscriptionWithError(err error) *EventAPIClientSubscription {
	return &EventAPIClientSubscription{
		error: err,
	}
}

// Error holds any error that happened when trying to subscribe.
func (s *EventAPIClientSubscription) Error() error {
	return s.error
}

// Close allows to close the subscription to cleanly unsubscribe from the node.
func (s *EventAPIClientSubscription) Close() error {
	if s.error != nil {
		return s.error
	}
	if token := s.mqttClient.Unsubscribe(s.topic); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func panicIfEventAPIClientInactive(neac *EventAPIClient) {
	if err := neac.Ctx.Err(); err != nil {
		panic(fmt.Errorf("%w: context is cancelled/done", ErrEventAPIClientInactive))
	}
	if !neac.MQTTClient.IsConnected() {
		panic(fmt.Errorf("%w: client is not connected", ErrEventAPIClientInactive))
	}
}

func sendErrOrDrop(errChan chan error, err error) {
	select {
	case errChan <- err:
	default:
	}
}

// Connect connects the EventAPIClient to the specified brokers.
// The EventAPIClient remains active as long as the given context isn't done/cancelled.
func (eac *EventAPIClient) Connect(ctx context.Context) error {
	eac.Ctx = ctx
	if token := eac.MQTTClient.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// Close disconnects the underlying MQTT client.
// Call this function to clean up any registered channels.
func (eac *EventAPIClient) Close() {
	eac.MQTTClient.Disconnect(0)
}

func (eac *EventAPIClient) subscribeToOutputsTopic(topic string) (<-chan *OutputResponse, *EventAPIClientSubscription) {
	panicIfEventAPIClientInactive(eac)
	channel := make(chan *OutputResponse)
	if token := eac.MQTTClient.Subscribe(topic, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		res := &OutputResponse{}
		if err := json.Unmarshal(mqttMsg.Payload(), res); err != nil {
			sendErrOrDrop(eac.Errors, err)
			return
		}
		select {
		case <-eac.Ctx.Done():
			return
		case channel <- res:
		}
	}); token.Wait() && token.Error() != nil {
		return nil, newSubscriptionWithError(token.Error())
	}
	return channel, newSubscription(eac.MQTTClient, topic)
}

func (eac *EventAPIClient) subscribeToMessageMetadataTopic(topic string) (<-chan *MessageMetadataResponse, *EventAPIClientSubscription) {
	panicIfEventAPIClientInactive(eac)
	channel := make(chan *MessageMetadataResponse)
	if token := eac.MQTTClient.Subscribe(topic, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		metadataRes := &MessageMetadataResponse{}
		if err := json.Unmarshal(mqttMsg.Payload(), metadataRes); err != nil {
			sendErrOrDrop(eac.Errors, err)
			return
		}
		select {
		case <-eac.Ctx.Done():
			return
		case channel <- metadataRes:
		}
	}); token.Wait() && token.Error() != nil {
		return nil, newSubscriptionWithError(token.Error())
	}
	return channel, newSubscription(eac.MQTTClient, topic)
}

func (eac *EventAPIClient) subscribeToMessageMetadataMessagesTopic(topic string, deSeriParas *iotago.DeSerializationParameters) (<-chan *iotago.Message, *EventAPIClientSubscription) {
	panicIfEventAPIClientInactive(eac)
	channel := make(chan *iotago.Message)
	if token := eac.MQTTClient.Subscribe(topic, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		metadataRes := &MessageMetadataResponse{}
		if err := json.Unmarshal(mqttMsg.Payload(), metadataRes); err != nil {
			sendErrOrDrop(eac.Errors, err)
			return
		}

		msg, err := eac.Client.MessageByMessageID(context.Background(), iotago.MustMessageIDFromHexString(metadataRes.MessageID), deSeriParas)
		if err != nil {
			return
		}

		select {
		case <-eac.Ctx.Done():
			return
		case channel <- msg:
		}
	}); token.Wait() && token.Error() != nil {
		return nil, newSubscriptionWithError(token.Error())
	}
	return channel, newSubscription(eac.MQTTClient, topic)
}

func (eac *EventAPIClient) subscribeToMessagesTopic(topic string, deSeriParas *iotago.DeSerializationParameters) (<-chan *iotago.Message, *EventAPIClientSubscription) {
	panicIfEventAPIClientInactive(eac)
	channel := make(chan *iotago.Message)
	if token := eac.MQTTClient.Subscribe(topic, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		msg := &iotago.Message{}
		if _, err := msg.Deserialize(mqttMsg.Payload(), serializer.DeSeriModeNoValidation, deSeriParas); err != nil {
			sendErrOrDrop(eac.Errors, err)
			return
		}
		select {
		case <-eac.Ctx.Done():
			return
		case channel <- msg:
		}
	}); token.Wait() && token.Error() != nil {
		return nil, newSubscriptionWithError(token.Error())
	}
	return channel, newSubscription(eac.MQTTClient, topic)
}

func (eac *EventAPIClient) subscribeToMilestoneInfoTopic(topic string) (<-chan *MilestoneInfo, *EventAPIClientSubscription) {
	panicIfEventAPIClientInactive(eac)
	channel := make(chan *MilestoneInfo)
	if token := eac.MQTTClient.Subscribe(topic, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		msPointer := &MilestoneInfo{}
		if err := json.Unmarshal(mqttMsg.Payload(), msPointer); err != nil {
			sendErrOrDrop(eac.Errors, err)
			return
		}
		select {
		case <-eac.Ctx.Done():
			return
		case channel <- msPointer:
		}
	}); token.Wait() && token.Error() != nil {
		return nil, newSubscriptionWithError(token.Error())
	}
	return channel, newSubscription(eac.MQTTClient, topic)
}

func (eac *EventAPIClient) subscribeToMilestoneTopic(topic string) (<-chan *iotago.Milestone, *EventAPIClientSubscription) {
	panicIfEventAPIClientInactive(eac)
	channel := make(chan *iotago.Milestone)
	if token := eac.MQTTClient.Subscribe(topic, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		msPayload := &iotago.Milestone{}
		if err := json.Unmarshal(mqttMsg.Payload(), msPayload); err != nil {
			sendErrOrDrop(eac.Errors, err)
			return
		}
		select {
		case <-eac.Ctx.Done():
			return
		case channel <- msPayload:
		}
	}); token.Wait() && token.Error() != nil {
		return nil, newSubscriptionWithError(token.Error())
	}
	return channel, newSubscription(eac.MQTTClient, topic)
}

func (eac *EventAPIClient) subscribeToReceiptsTopic(topic string) (<-chan *iotago.ReceiptMilestoneOpt, *EventAPIClientSubscription) {
	panicIfEventAPIClientInactive(eac)
	channel := make(chan *iotago.ReceiptMilestoneOpt)
	if token := eac.MQTTClient.Subscribe(topic, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		receipt := &iotago.ReceiptMilestoneOpt{}
		if err := json.Unmarshal(mqttMsg.Payload(), receipt); err != nil {
			sendErrOrDrop(eac.Errors, err)
			return
		}
		select {
		case <-eac.Ctx.Done():
			return
		case channel <- receipt:
		}
	}); token.Wait() && token.Error() != nil {
		return nil, newSubscriptionWithError(token.Error())
	}
	return channel, newSubscription(eac.MQTTClient, topic)
}

// Messages returns a channel of newly received messages.
func (eac *EventAPIClient) Messages(deSeriParas *iotago.DeSerializationParameters) (<-chan *iotago.Message, *EventAPIClientSubscription) {
	return eac.subscribeToMessagesTopic(EventAPIMessages, deSeriParas)
}

// ReferencedMessagesMetadata returns a channel of message metadata of newly referenced messages.
func (eac *EventAPIClient) ReferencedMessagesMetadata() (<-chan *MessageMetadataResponse, *EventAPIClientSubscription) {
	return eac.subscribeToMessageMetadataTopic(EventAPIMessageMetadataReferenced)
}

// ReferencedMessages returns a channel of newly referenced messages.
func (eac *EventAPIClient) ReferencedMessages(deSeriParas *iotago.DeSerializationParameters) (<-chan *iotago.Message, *EventAPIClientSubscription) {
	return eac.subscribeToMessageMetadataMessagesTopic(EventAPIMessageMetadataReferenced, deSeriParas)
}

// TransactionMessages returns a channel of messages containing transactions.
func (eac *EventAPIClient) TransactionMessages(deSeriParas *iotago.DeSerializationParameters) (<-chan *iotago.Message, *EventAPIClientSubscription) {
	return eac.subscribeToMessagesTopic(EventAPIMessagesTransaction, deSeriParas)
}

// TransactionTaggedDataMessages returns a channel of messages containing transactions with tagged data.
func (eac *EventAPIClient) TransactionTaggedDataMessages(deSeriParas *iotago.DeSerializationParameters) (<-chan *iotago.Message, *EventAPIClientSubscription) {
	return eac.subscribeToMessagesTopic(EventAPIMessagesTransactionTaggedData, deSeriParas)
}

// TransactionTaggedDataWithTagMessages returns a channel of messages containing transactions with tagged data containing the given tag.
func (eac *EventAPIClient) TransactionTaggedDataWithTagMessages(tag []byte, deSeriParas *iotago.DeSerializationParameters) (<-chan *iotago.Message, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPIMessagesTransactionTaggedDataTag, "{tag}", iotago.EncodeHex(tag), 1)
	return eac.subscribeToMessagesTopic(topic, deSeriParas)
}

// TaggedDataMessages returns a channel of messages containing tagged data containing the given tag.
func (eac *EventAPIClient) TaggedDataMessages(deSeriParas *iotago.DeSerializationParameters) (<-chan *iotago.Message, *EventAPIClientSubscription) {
	return eac.subscribeToMessagesTopic(EventAPIMessagesTaggedData, deSeriParas)
}

// TaggedDataWithTagMessages returns a channel of messages containing tagged data.
func (eac *EventAPIClient) TaggedDataWithTagMessages(tag []byte, deSeriParas *iotago.DeSerializationParameters) (<-chan *iotago.Message, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPIMessagesTaggedDataTag, "{tag}", iotago.EncodeHex(tag), 1)
	return eac.subscribeToMessagesTopic(topic, deSeriParas)
}

// MessageMetadataChange returns a channel of MessageMetadataResponse each time the given message's state changes.
func (eac *EventAPIClient) MessageMetadataChange(msgID iotago.MessageID) (<-chan *MessageMetadataResponse, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPIMessageMetadata, "{messageId}", iotago.MessageIDToHexString(msgID), 1)
	return eac.subscribeToMessageMetadataTopic(topic)
}

// NFTOutputsByID returns a channel of newly created outputs to track the chain mutations of a given NFT.
func (eac *EventAPIClient) NFTOutputsByID(nftID iotago.NFTID) (<-chan *OutputResponse, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPINFTOutputs, "{nftId}", nftID.String(), 1)
	return eac.subscribeToOutputsTopic(topic)
}

// AliasOutputsByID returns a channel of newly created outputs to track the chain mutations of a given Alias.
func (eac *EventAPIClient) AliasOutputsByID(aliasID iotago.AliasID) (<-chan *OutputResponse, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPIAliasOutputs, "{aliasId}", aliasID.String(), 1)
	return eac.subscribeToOutputsTopic(topic)
}

// FoundryOutputsByID returns a channel of newly created outputs to track the chain mutations of a given Foundry.
func (eac *EventAPIClient) FoundryOutputsByID(foundryID iotago.FoundryID) (<-chan *OutputResponse, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPIFoundryOutputs, "{foundryId}", foundryID.String(), 1)
	return eac.subscribeToOutputsTopic(topic)
}

// OutputsByUnlockConditionAndAddress returns a channel of newly created outputs on the given unlock condition and address.
func (eac *EventAPIClient) OutputsByUnlockConditionAndAddress(addr iotago.Address, netPrefix iotago.NetworkPrefix, condition EventAPIUnlockCondition) (<-chan *OutputResponse, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPIOutputsByUnlockConditionAndAddress, "{address}", addr.Bech32(netPrefix), 1)
	topic = strings.Replace(topic, "{condition}", string(condition), 1)
	return eac.subscribeToOutputsTopic(topic)
}

// SpentOutputsByUnlockConditionAndAddress returns a channel of newly spent outputs on the given unlock condition and address.
func (eac *EventAPIClient) SpentOutputsByUnlockConditionAndAddress(addr iotago.Address, netPrefix iotago.NetworkPrefix, condition EventAPIUnlockCondition) (<-chan *OutputResponse, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPISpentOutputsByUnlockConditionAndAddress, "{address}", addr.Bech32(netPrefix), 1)
	topic = strings.Replace(topic, "{condition}", string(condition), 1)
	return eac.subscribeToOutputsTopic(topic)
}

// TransactionIncludedMessage returns a channel of the included message which carries the transaction with the given ID.
func (eac *EventAPIClient) TransactionIncludedMessage(txID iotago.TransactionID, deSeriParas *iotago.DeSerializationParameters) (<-chan *iotago.Message, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPITransactionsIncludedMessage, "{transactionId}", iotago.MessageIDToHexString(txID), 1)
	return eac.subscribeToMessagesTopic(topic, deSeriParas)
}

// Output returns a channel which immediately returns the output with the given ID and afterwards when its state changes.
func (eac *EventAPIClient) Output(outputID iotago.OutputID) (<-chan *OutputResponse, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPIOutputs, "{outputId}", iotago.EncodeHex(outputID[:]), 1)
	return eac.subscribeToOutputsTopic(topic)
}

// Receipts returns a channel which returns newly applied receipts.
func (eac *EventAPIClient) Receipts() (<-chan *iotago.ReceiptMilestoneOpt, *EventAPIClientSubscription) {
	return eac.subscribeToReceiptsTopic(EventAPIReceipts)
}

// MilestoneInfo is an informative struct holding a milestone index, milestone ID and timestamp.
type MilestoneInfo struct {
	Index       uint32 `json:"index"`
	Timestamp   uint32 `json:"timestamp"`
	MilestoneID string `json:"milestoneId"`
}

// LatestMilestones returns a channel of infos about newly seen latest milestones.
func (eac *EventAPIClient) LatestMilestones() (<-chan *MilestoneInfo, *EventAPIClientSubscription) {
	return eac.subscribeToMilestoneInfoTopic(EventAPIMilestoneInfoLatest)
}

// ConfirmedMilestones returns a channel of infos about newly confirmed milestones.
func (eac *EventAPIClient) ConfirmedMilestones() (<-chan *MilestoneInfo, *EventAPIClientSubscription) {
	return eac.subscribeToMilestoneInfoTopic(EventAPIMilestoneInfoConfirmed)
}

// Milestones returns a channel of newly received milestones.
func (eac *EventAPIClient) Milestones() (<-chan *iotago.Milestone, *EventAPIClientSubscription) {
	return eac.subscribeToMilestoneTopic(EventAPIMilestones)
}
