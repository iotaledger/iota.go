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
	// EventAPIMilestonesLatest is the name of the latest milestone event channel.
	EventAPIMilestonesLatest = "milestones/latest"
	// EventAPIMilestonesConfirmed is the name of the confirmed milestone event channel.
	EventAPIMilestonesConfirmed = "milestones/confirmed"

	// EventAPIMessages is the name of the received messages event channel.
	EventAPIMessages = "messages"
	// EventAPIMessagesReferenced is the name of the referenced messages metadata event channel.
	EventAPIMessagesReferenced = "messages/referenced"
	// EventAPIMessagesTransaction is the name of the transaction messages metadata event channel.
	EventAPIMessagesTransaction = "messages/transaction"
	// EventAPIMessagesTransactionTaggedData is the name of the transaction with tagged data messages metadata event channel.
	EventAPIMessagesTransactionTaggedData = "messages/transaction/taggedData"
	// EventAPIMessagesTransactionTaggedDataTag is the name of the transaction with tagged data with tag messages metadata event channel.
	EventAPIMessagesTransactionTaggedDataTag = "messages/transaction/taggedData/{tag}"
	// EventAPIMessagesMilestone is the name of the milestone messages metadata event channel.
	EventAPIMessagesMilestone = "messages/milestone"
	// EventAPIMessagesTaggedData is the name of the tagged data messages metadata event channel.
	EventAPIMessagesTaggedData = "messages/taggedData"
	// EventAPIMessagesTaggedDataTag is the name of the tagged data with tag messages metadata event channel.
	EventAPIMessagesTaggedDataTag = "messages/taggedData/{tag}"
	// EventAPIMessagesMetadata is the name of the message metadata event channel.
	EventAPIMessagesMetadata = "messages/{messageId}/metadata"

	// EventAPITransactionsIncludedMessage is the name of the included transaction message event channel.
	EventAPITransactionsIncludedMessage = "transactions/{transactionId}/included-message"

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
	UnlockConditionStorageReturn    EventAPIUnlockCondition = "storageReturn"
	UnlockConditionExpirationReturn EventAPIUnlockCondition = "expirationReturn"
	UnlockConditionStateController  EventAPIUnlockCondition = "stateController"
	UnlockConditionGovernor         EventAPIUnlockCondition = "governor"
	UnlockConditionImmutableAlias   EventAPIUnlockCondition = "immutableAlias"
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

func (eac *EventAPIClient) subscribeToMilestonesTopic(topic string) (<-chan *MilestonePointer, *EventAPIClientSubscription) {
	panicIfEventAPIClientInactive(eac)
	channel := make(chan *MilestonePointer)
	if token := eac.MQTTClient.Subscribe(topic, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		msPointer := &MilestonePointer{}
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

func (eac *EventAPIClient) subscribeToMilestoneMessagesTopic(topic string, deSeriParas *iotago.DeSerializationParameters) (<-chan *iotago.Message, *EventAPIClientSubscription) {
	panicIfEventAPIClientInactive(eac)
	channel := make(chan *iotago.Message)
	if token := eac.MQTTClient.Subscribe(topic, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		msPointer := &MilestonePointer{}
		if err := json.Unmarshal(mqttMsg.Payload(), msPointer); err != nil {
			sendErrOrDrop(eac.Errors, err)
			return
		}
		res, err := eac.Client.MilestoneByIndex(context.Background(), msPointer.Index)
		if err != nil {
			sendErrOrDrop(eac.Errors, err)
			return
		}
		msg, err := eac.Client.MessageByMessageID(context.Background(), iotago.MustMessageIDFromHexString(res.MessageID), deSeriParas)
		if err != nil {
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

// Messages returns a channel of newly received messages.
func (eac *EventAPIClient) Messages(deSeriParas *iotago.DeSerializationParameters) (<-chan *iotago.Message, *EventAPIClientSubscription) {
	return eac.subscribeToMessagesTopic(EventAPIMessages, deSeriParas)
}

// ReferencedMessagesMetadata returns a channel of message metadata of newly referenced messages.
func (eac *EventAPIClient) ReferencedMessagesMetadata() (<-chan *MessageMetadataResponse, *EventAPIClientSubscription) {
	return eac.subscribeToMessageMetadataTopic(EventAPIMessagesReferenced)
}

// ReferencedMessages returns a channel of newly referenced messages.
func (eac *EventAPIClient) ReferencedMessages(deSeriParas *iotago.DeSerializationParameters) (<-chan *iotago.Message, *EventAPIClientSubscription) {
	return eac.subscribeToMessageMetadataMessagesTopic(EventAPIMessagesReferenced, deSeriParas)
}

// TransactionMessagesMetadata returns a channel of message metadata of messages containing transactions.
func (eac *EventAPIClient) TransactionMessagesMetadata() (<-chan *MessageMetadataResponse, *EventAPIClientSubscription) {
	return eac.subscribeToMessageMetadataTopic(EventAPIMessagesTransaction)
}

// TransactionMessages returns a channel of messages containing transactions.
func (eac *EventAPIClient) TransactionMessages(deSeriParas *iotago.DeSerializationParameters) (<-chan *iotago.Message, *EventAPIClientSubscription) {
	return eac.subscribeToMessageMetadataMessagesTopic(EventAPIMessagesTransaction, deSeriParas)
}

// TransactionTaggedDataMessages returns a channel of messages containing transactions with tagged data.
func (eac *EventAPIClient) TransactionTaggedDataMessages(deSeriParas *iotago.DeSerializationParameters) (<-chan *iotago.Message, *EventAPIClientSubscription) {
	return eac.subscribeToMessageMetadataMessagesTopic(EventAPIMessagesTransactionTaggedData, deSeriParas)
}

// TransactionTaggedDataMessagesMetadata returns a channel of message metadata of messages containing transactions with tagged data.
func (eac *EventAPIClient) TransactionTaggedDataMessagesMetadata() (<-chan *MessageMetadataResponse, *EventAPIClientSubscription) {
	return eac.subscribeToMessageMetadataTopic(EventAPIMessagesTransactionTaggedData)
}

// TransactionTaggedDataWithTagMessages returns a channel of messages containing transactions with tagged data containing the given tag.
func (eac *EventAPIClient) TransactionTaggedDataWithTagMessages(tag []byte, deSeriParas *iotago.DeSerializationParameters) (<-chan *iotago.Message, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPIMessagesTransactionTaggedDataTag, "{tag}", iotago.EncodeHex(tag), 1)
	return eac.subscribeToMessageMetadataMessagesTopic(topic, deSeriParas)
}

// TransactionTaggedDataMessagesWithTagMetadata returns a channel of message metadata of messages containing transactions with tagged data containing the given tag.
func (eac *EventAPIClient) TransactionTaggedDataMessagesWithTagMetadata(tag []byte) (<-chan *MessageMetadataResponse, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPIMessagesTransactionTaggedDataTag, "{tag}", iotago.EncodeHex(tag), 1)
	return eac.subscribeToMessageMetadataTopic(topic)
}

// MilestoneMessagesMetadata returns a channel of message metadata of messages containing milestones.
func (eac *EventAPIClient) MilestoneMessagesMetadata() (<-chan *MessageMetadataResponse, *EventAPIClientSubscription) {
	return eac.subscribeToMessageMetadataTopic(EventAPIMessagesMilestone)
}

// MilestoneMessages returns a channel of messages containing milestones.
func (eac *EventAPIClient) MilestoneMessages(deSeriParas *iotago.DeSerializationParameters) (<-chan *iotago.Message, *EventAPIClientSubscription) {
	return eac.subscribeToMessageMetadataMessagesTopic(EventAPIMessagesMilestone, deSeriParas)
}

// TaggedDataMessagesMetadata returns a channel of message metadata of messages containing tagged data.
func (eac *EventAPIClient) TaggedDataMessagesMetadata() (<-chan *MessageMetadataResponse, *EventAPIClientSubscription) {
	return eac.subscribeToMessageMetadataTopic(EventAPIMessagesTaggedData)
}

// TaggedDataMessages returns a channel of messages containing tagged data containing the given tag.
func (eac *EventAPIClient) TaggedDataMessages(deSeriParas *iotago.DeSerializationParameters) (<-chan *iotago.Message, *EventAPIClientSubscription) {
	return eac.subscribeToMessageMetadataMessagesTopic(EventAPIMessagesTaggedData, deSeriParas)
}

// TaggedDataWithTagMessagesMetadata returns a channel of message metadata of messages containing tagged data containing the given tag.
func (eac *EventAPIClient) TaggedDataWithTagMessagesMetadata(tag []byte) (<-chan *MessageMetadataResponse, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPIMessagesTaggedDataTag, "{tag}", iotago.EncodeHex(tag), 1)
	return eac.subscribeToMessageMetadataTopic(topic)
}

// TaggedDataWithTagMessages returns a channel of messages containing tagged data.
func (eac *EventAPIClient) TaggedDataWithTagMessages(tag []byte, deSeriParas *iotago.DeSerializationParameters) (<-chan *iotago.Message, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPIMessagesTaggedDataTag, "{tag}", iotago.EncodeHex(tag), 1)
	return eac.subscribeToMessageMetadataMessagesTopic(topic, deSeriParas)
}

// MessageMetadataChange returns a channel of MessageMetadataResponse each time the given message's state changes.
func (eac *EventAPIClient) MessageMetadataChange(msgID iotago.MessageID) (<-chan *MessageMetadataResponse, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPIMessagesMetadata, "{messageId}", iotago.MessageIDToHexString(msgID), 1)
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
	panicIfEventAPIClientInactive(eac)
	channel := make(chan *iotago.ReceiptMilestoneOpt)
	if token := eac.MQTTClient.Subscribe(EventAPIReceipts, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
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
	return channel, newSubscription(eac.MQTTClient, EventAPIReceipts)
}

// MilestonePointer is an informative struct holding a milestone index and timestamp.
type MilestonePointer struct {
	Index     uint32 `json:"index"`
	Timestamp uint64 `json:"timestamp"`
}

// LatestMilestones returns a channel of newly seen latest milestones.
func (eac *EventAPIClient) LatestMilestones() (<-chan *MilestonePointer, *EventAPIClientSubscription) {
	return eac.subscribeToMilestonesTopic(EventAPIMilestonesLatest)
}

// LatestMilestoneMessages returns a channel of newly seen latest milestones messages.
func (eac *EventAPIClient) LatestMilestoneMessages(deSeriParas *iotago.DeSerializationParameters) (<-chan *iotago.Message, *EventAPIClientSubscription) {
	return eac.subscribeToMilestoneMessagesTopic(EventAPIMilestonesLatest, deSeriParas)
}

// ConfirmedMilestones returns a channel of newly confirmed milestones.
func (eac *EventAPIClient) ConfirmedMilestones() (<-chan *MilestonePointer, *EventAPIClientSubscription) {
	return eac.subscribeToMilestonesTopic(EventAPIMilestonesConfirmed)
}

// ConfirmedMilestoneMessages returns a channel of newly confirmed milestones messages.
func (eac *EventAPIClient) ConfirmedMilestoneMessages(deSeriParas *iotago.DeSerializationParameters) (<-chan *iotago.Message, *EventAPIClientSubscription) {
	return eac.subscribeToMilestoneMessagesTopic(EventAPIMilestonesConfirmed, deSeriParas)
}
