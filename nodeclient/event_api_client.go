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

	iotago "github.com/iotaledger/iota.go/v3"
)

const (
	// EventAPIMilestoneInfoLatest is the name of the latest milestone info event channel.
	EventAPIMilestoneInfoLatest = "milestone-info/latest"
	// EventAPIMilestoneInfoConfirmed is the name of the confirmed milestone info event channel.
	EventAPIMilestoneInfoConfirmed = "milestone-info/confirmed"
	// EventAPIMilestones is the name of the milestone event channel.
	EventAPIMilestones = "milestones"

	// EventAPIBlocks is the name of the received blocks event channel.
	EventAPIBlocks = "blocks"
	// EventAPIBlocksTransaction is the name of the blocks containing transactions event channel.
	EventAPIBlocksTransaction = "blocks/transaction"
	// EventAPIBlocksTransactionTaggedData is the name of the blocks containing transaction with tagged data event channel.
	EventAPIBlocksTransactionTaggedData = "blocks/transaction/tagged-data"
	// EventAPIBlocksTransactionTaggedDataTag is the name of the blocks containing transaction with a specific tagged data event channel.
	EventAPIBlocksTransactionTaggedDataTag = "blocks/transaction/tagged-data/{tag}"
	// EventAPIBlocksTaggedData is the name of the blocks containing tagged data event channel.
	EventAPIBlocksTaggedData = "blocks/tagged-data"
	// EventAPIBlocksTaggedDataTag is the name of the blocks containing a specific tagged data event channel.
	EventAPIBlocksTaggedDataTag = "blocks/tagged-data/{tag}"

	// EventAPITransactionsIncludedBlock is the name of the included transaction block event channel.
	EventAPITransactionsIncludedBlock = "transactions/{transactionId}/included-block"

	// EventAPIBlockMetadata is the name of the block metadata event channel.
	EventAPIBlockMetadata = "block-metadata/{blockId}"
	// EventAPIBlockMetadataReferenced is the name of the referenced blocks metadata event channel.
	EventAPIBlockMetadataReferenced = "block-metadata/referenced"

	// EventAPIOutputs is the name of the outputs event channel.
	EventAPIOutputs = "outputs/{outputId}"
	// EventAPINFTOutputs is the name of the NFT output event channel to retrieve NFT mutations by their ID.
	EventAPINFTOutputs = "outputs/nft/{nftId}"
	// EventAPIAliasOutputs is the name of the Alias output event channel to retrieve Alias mutations by their ID.
	EventAPIAliasOutputs = "outputs/alias/{aliasId}"
	// EventAPIFoundryOutputs is the name of the Foundry output event channel to retrieve Foundry mutations by their ID.
	EventAPIFoundryOutputs = "outputs/foundry/{foundryId}"
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
	return fmt.Sprintf("%s/api/%s", baseURL, MQTTPluginName)
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
		panic(fmt.Errorf("%w: context is canceled/done", ErrEventAPIClientInactive))
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
// The EventAPIClient remains active as long as the given context isn't done/canceled.
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

func jsonDeserializer[T any](payload []byte) (*T, error) {
	var inst T
	if err := json.Unmarshal(payload, &inst); err != nil {
		return nil, err
	}
	return &inst, nil
}

func subscribeToTopic[T any](eac *EventAPIClient, topic string, deseriFunc func(payload []byte) (*T, error)) (<-chan *T, *EventAPIClientSubscription) {
	panicIfEventAPIClientInactive(eac)
	channel := make(chan *T)
	if token := eac.MQTTClient.Subscribe(topic, 2, func(client mqtt.Client, mqttMsg mqtt.Message) {
		obj, err := deseriFunc(mqttMsg.Payload())
		if err != nil {
			sendErrOrDrop(eac.Errors, err)
			return
		}

		select {
		case <-eac.Ctx.Done():
			return
		case channel <- obj:
		}
	}); token.Wait() && token.Error() != nil {
		return nil, newSubscriptionWithError(token.Error())
	}
	return channel, newSubscription(eac.MQTTClient, topic)
}

func (eac *EventAPIClient) subscribeToOutputsTopic(topic string) (<-chan *OutputResponse, *EventAPIClientSubscription) {
	return subscribeToTopic[OutputResponse](eac, topic, jsonDeserializer[OutputResponse])
}

func (eac *EventAPIClient) subscribeToBlockMetadataTopic(topic string) (<-chan *BlockMetadataResponse, *EventAPIClientSubscription) {
	return subscribeToTopic[BlockMetadataResponse](eac, topic, jsonDeserializer[BlockMetadataResponse])
}

func (eac *EventAPIClient) subscribeToBlockMetadataBlockTopic(topic string) (<-chan *iotago.Block, *EventAPIClientSubscription) {
	return subscribeToTopic[iotago.Block](eac, topic, func(payload []byte) (*iotago.Block, error) {
		metadataRes := &BlockMetadataResponse{}
		if err := json.Unmarshal(payload, metadataRes); err != nil {
			sendErrOrDrop(eac.Errors, err)
			return nil, err
		}

		return eac.Client.BlockByBlockID(context.Background(), iotago.MustBlockIDFromHexString(metadataRes.BlockID))
	})
}

func (eac *EventAPIClient) subscribeToBlocksTopic(topic string) (<-chan *iotago.Block, *EventAPIClientSubscription) {
	return subscribeToTopic[iotago.Block](eac, topic, func(payload []byte) (*iotago.Block, error) {
		block := &iotago.Block{}
		if _, err := eac.Client.opts.iotagoAPI.Decode(payload, block); err != nil {
			return nil, err
		}
		return block, nil
	})
}

func (eac *EventAPIClient) subscribeToMilestoneInfoTopic(topic string) (<-chan *MilestoneInfo, *EventAPIClientSubscription) {
	return subscribeToTopic[MilestoneInfo](eac, topic, jsonDeserializer[MilestoneInfo])
}

func (eac *EventAPIClient) subscribeToMilestoneTopic(topic string) (<-chan *iotago.Milestone, *EventAPIClientSubscription) {
	return subscribeToTopic[iotago.Milestone](eac, topic, jsonDeserializer[iotago.Milestone])
}

func (eac *EventAPIClient) subscribeToReceiptsTopic(topic string) (<-chan *iotago.ReceiptMilestoneOpt, *EventAPIClientSubscription) {
	return subscribeToTopic[iotago.ReceiptMilestoneOpt](eac, topic, jsonDeserializer[iotago.ReceiptMilestoneOpt])
}

// Blocks returns a channel of newly received blocks.
func (eac *EventAPIClient) Blocks() (<-chan *iotago.Block, *EventAPIClientSubscription) {
	return eac.subscribeToBlocksTopic(EventAPIBlocks)
}

// ReferencedBlocksMetadata returns a channel of block metadata of newly referenced blocks.
func (eac *EventAPIClient) ReferencedBlocksMetadata() (<-chan *BlockMetadataResponse, *EventAPIClientSubscription) {
	return eac.subscribeToBlockMetadataTopic(EventAPIBlockMetadataReferenced)
}

// ReferencedBlocks returns a channel of newly referenced blocks.
func (eac *EventAPIClient) ReferencedBlocks() (<-chan *iotago.Block, *EventAPIClientSubscription) {
	return eac.subscribeToBlockMetadataBlockTopic(EventAPIBlockMetadataReferenced)
}

// TransactionBlocks returns a channel of blocks containing transactions.
func (eac *EventAPIClient) TransactionBlocks() (<-chan *iotago.Block, *EventAPIClientSubscription) {
	return eac.subscribeToBlocksTopic(EventAPIBlocksTransaction)
}

// TransactionTaggedDataBlocks returns a channel of blocks containing transactions with tagged data.
func (eac *EventAPIClient) TransactionTaggedDataBlocks() (<-chan *iotago.Block, *EventAPIClientSubscription) {
	return eac.subscribeToBlocksTopic(EventAPIBlocksTransactionTaggedData)
}

// TransactionTaggedDataWithTagBlocks returns a channel of blocks containing transactions with tagged data containing the given tag.
func (eac *EventAPIClient) TransactionTaggedDataWithTagBlocks(tag []byte) (<-chan *iotago.Block, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPIBlocksTransactionTaggedDataTag, "{tag}", iotago.EncodeHex(tag), 1)
	return eac.subscribeToBlocksTopic(topic)
}

// TaggedDataBlocks returns a channel of blocks containing tagged data containing the given tag.
func (eac *EventAPIClient) TaggedDataBlocks() (<-chan *iotago.Block, *EventAPIClientSubscription) {
	return eac.subscribeToBlocksTopic(EventAPIBlocksTaggedData)
}

// TaggedDataWithTagBlocks returns a channel of blocks containing tagged data.
func (eac *EventAPIClient) TaggedDataWithTagBlocks(tag []byte) (<-chan *iotago.Block, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPIBlocksTaggedDataTag, "{tag}", iotago.EncodeHex(tag), 1)
	return eac.subscribeToBlocksTopic(topic)
}

// BlockMetadataChange returns a channel of BlockMetadataResponse each time the given block's state changes.
func (eac *EventAPIClient) BlockMetadataChange(blockID iotago.BlockID) (<-chan *BlockMetadataResponse, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPIBlockMetadata, "{blockId}", blockID.ToHex(), 1)
	return eac.subscribeToBlockMetadataTopic(topic)
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

// TransactionIncludedBlock returns a channel of the included block which carries the transaction with the given ID.
func (eac *EventAPIClient) TransactionIncludedBlock(txID iotago.TransactionID) (<-chan *iotago.Block, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPITransactionsIncludedBlock, "{transactionId}", txID.ToHex(), 1)
	return eac.subscribeToBlocksTopic(topic)
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
	Index       iotago.MilestoneIndex `json:"index"`
	Timestamp   uint32                `json:"timestamp"`
	MilestoneID string                `json:"milestoneId"`
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
