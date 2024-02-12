package nodeclient

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/iotaledger/hive.go/ierrors"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/api"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

const (
	// EventAPICommitmentInfoLatest is the name of the latest commitment info event channel.
	EventAPICommitmentInfoLatest = "commitment-info/latest"
	// EventAPICommitmentInfoFinalized is the name of the finalized commitment info event channel.
	EventAPICommitmentInfoFinalized = "commitment-info/finalized"
	// EventAPICommitments is the name of the commitment event channel.
	EventAPICommitments = "commitments"

	// EventAPIBlocks is the name of the received blocks event channel.
	EventAPIBlocks = "blocks"
	// EventAPIBlocksAccepted is the name of the accepted blocks event channel.
	EventAPIBlocksAccepted = "blocks/accepted"
	// EventAPIBlocksConfirmed is the name of the confirmed blocks event channel.
	EventAPIBlocksConfirmed = "blocks/confirmed"
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

	// EventAPIOutputs is the name of the outputs event channel.
	EventAPIOutputs = "outputs/{outputId}"
	// EventAPIOutputMetadata is the name of the outputs event channel.
	EventAPIOutputMetadata = "output-metadata/{outputId}"
	// EventAPINFTOutputs is the name of the NFT output event channel to retrieve NFT mutations by their ID.
	EventAPINFTOutputs = "outputs/nft/{nftId}"
	// EventAPIAccountOutputs is the name of the Account output event channel to retrieve Account mutations by their ID.
	EventAPIAccountOutputs = "outputs/account/{accountId}"
	// EventAPIFoundryOutputs is the name of the Foundry output event channel to retrieve Foundry mutations by their ID.
	EventAPIFoundryOutputs = "outputs/foundry/{foundryId}"
	// EventAPIOutputsByUnlockConditionAndAddress is the name of the outputs by unlock condition address event channel.
	EventAPIOutputsByUnlockConditionAndAddress = "outputs/unlock/{condition}/{address}"
	// EventAPISpentOutputsByUnlockConditionAndAddress is the name of the spent outputs by unlock condition address event channel.
	EventAPISpentOutputsByUnlockConditionAndAddress = "outputs/unlock/{condition}/{address}/spent"
)

var (
	// ErrEventAPIClientInactive gets returned when an EventAPIClient is inactive.
	ErrEventAPIClientInactive = ierrors.New("event api client is inactive")
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
	UnlockConditionImmutableAccount EventAPIUnlockCondition = "immutable-account"
)

func randMQTTClientID() string {
	return strconv.FormatInt(rand.NewSource(time.Now().UnixNano()).Int63(), 10)
}

func brokerURLFromClient(nc *Client) string {
	baseURL := nc.BaseURL
	baseURL = strings.Replace(baseURL, "https://", "wss://", 1)
	baseURL = strings.Replace(baseURL, "http://", "ws://", 1)

	return fmt.Sprintf("%s%s/%s", baseURL, api.APIRoot, api.MQTTPluginName)
}

func newEventAPIClient(nc *Client) *EventAPIClient {
	clientOpts := mqtt.NewClientOptions()
	clientOpts.Order = false
	clientOpts.ClientID = randMQTTClientID()
	clientOpts.AddBroker(brokerURLFromClient(nc))
	errChan := make(chan error)
	clientOpts.OnConnectionLost = func(_ mqtt.Client, err error) { sendErrOrDrop(errChan, err) }

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
	//nolint:containedctx
	ctx context.Context
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
	if err := neac.ctx.Err(); err != nil {
		panic(ierrors.Wrap(ErrEventAPIClientInactive, "context is canceled/done"))
	}
	if !neac.MQTTClient.IsConnected() {
		panic(ierrors.Wrap(ErrEventAPIClientInactive, "client is not connected"))
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
	eac.ctx = ctx
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

func subscribeToTopic[T any](eac *EventAPIClient, topic string, deseriFunc func(payload []byte) (T, error)) (<-chan T, *EventAPIClientSubscription) {
	panicIfEventAPIClientInactive(eac)
	channel := make(chan T)
	if token := eac.MQTTClient.Subscribe(topic, 2, func(_ mqtt.Client, mqttMsg mqtt.Message) {
		obj, err := deseriFunc(mqttMsg.Payload())
		if err != nil {
			sendErrOrDrop(eac.Errors, err)

			return
		}

		select {
		case <-eac.ctx.Done():
			return
		case channel <- obj:
		}
	}); token.Wait() && token.Error() != nil {
		return nil, newSubscriptionWithError(token.Error())
	}

	return channel, newSubscription(eac.MQTTClient, topic)
}

func (eac *EventAPIClient) subscribeToOutputsTopic(topic string) (<-chan iotago.Output, *EventAPIClientSubscription) {
	return subscribeToTopic(eac, topic, func(payload []byte) (iotago.Output, error) {
		var output iotago.TxEssenceOutput
		if err := eac.Client.CommittedAPI().JSONDecode(payload, &output); err != nil {
			sendErrOrDrop(eac.Errors, err)
			return nil, err
		}

		return output, nil
	})
}

func (eac *EventAPIClient) subscribeToBlockMetadataTopic(topic string) (<-chan *api.BlockMetadataResponse, *EventAPIClientSubscription) {
	return subscribeToTopic(eac, topic, func(payload []byte) (*api.BlockMetadataResponse, error) {
		response := new(api.BlockMetadataResponse)
		if err := eac.Client.CommittedAPI().JSONDecode(payload, response); err != nil {
			sendErrOrDrop(eac.Errors, err)
			return nil, err
		}

		return response, nil
	})
}

func (eac *EventAPIClient) subscribeToBlocksTopic(topic string) (<-chan *iotago.Block, *EventAPIClientSubscription) {
	return subscribeToTopic(eac, topic, func(payload []byte) (*iotago.Block, error) {
		version, _, err := iotago.VersionFromBytes(payload)
		if err != nil {
			sendErrOrDrop(eac.Errors, err)
			return nil, err
		}
		apiForVersion, err := eac.Client.apiProvider.APIForVersion(version)
		if err != nil {
			sendErrOrDrop(eac.Errors, err)
			return nil, err
		}

		block := new(iotago.Block)
		if _, err := apiForVersion.Decode(payload, block); err != nil {
			sendErrOrDrop(eac.Errors, err)
			return nil, err
		}

		return block, nil
	})
}

// Blocks returns a channel of newly received blocks.
func (eac *EventAPIClient) Blocks() (<-chan *iotago.Block, *EventAPIClientSubscription) {
	return eac.subscribeToBlocksTopic(EventAPIBlocks)
}

// AcceptedBlocks returns a channel of blocks of newly accepted blocks.
func (eac *EventAPIClient) AcceptedBlocks() (<-chan *iotago.Block, *EventAPIClientSubscription) {
	return eac.subscribeToBlocksTopic(EventAPIBlocksAccepted)
}

// ConfirmedBlocks returns a channel of blocks of newly confirmed blocks.
func (eac *EventAPIClient) ConfirmedBlocks() (<-chan *iotago.Block, *EventAPIClientSubscription) {
	return eac.subscribeToBlocksTopic(EventAPIBlocksConfirmed)
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
	topic := strings.Replace(EventAPIBlocksTransactionTaggedDataTag, "{tag}", hexutil.EncodeHex(tag), 1)

	return eac.subscribeToBlocksTopic(topic)
}

// TaggedDataBlocks returns a channel of blocks containing tagged data containing the given tag.
func (eac *EventAPIClient) TaggedDataBlocks() (<-chan *iotago.Block, *EventAPIClientSubscription) {
	return eac.subscribeToBlocksTopic(EventAPIBlocksTaggedData)
}

// TaggedDataWithTagBlocks returns a channel of blocks containing tagged data.
func (eac *EventAPIClient) TaggedDataWithTagBlocks(tag []byte) (<-chan *iotago.Block, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPIBlocksTaggedDataTag, "{tag}", hexutil.EncodeHex(tag), 1)

	return eac.subscribeToBlocksTopic(topic)
}

// BlockMetadataChange returns a channel of BlockMetadataResponse each time the given block's state changes.
func (eac *EventAPIClient) BlockMetadataChange(blockID iotago.BlockID) (<-chan *api.BlockMetadataResponse, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPIBlockMetadata, "{blockId}", blockID.ToHex(), 1)

	return eac.subscribeToBlockMetadataTopic(topic)
}

// NFTOutputsByID returns a channel of newly created outputs to track the chain mutations of a given NFT.
func (eac *EventAPIClient) NFTOutputsByID(nftID iotago.NFTID) (<-chan iotago.Output, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPINFTOutputs, "{nftId}", nftID.String(), 1)

	return eac.subscribeToOutputsTopic(topic)
}

// AccountOutputsByID returns a channel of newly created outputs to track the chain mutations of a given Account.
func (eac *EventAPIClient) AccountOutputsByID(accountID iotago.AccountID) (<-chan iotago.Output, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPIAccountOutputs, "{accountId}", accountID.String(), 1)

	return eac.subscribeToOutputsTopic(topic)
}

// FoundryOutputsByID returns a channel of newly created outputs to track the chain mutations of a given Foundry.
func (eac *EventAPIClient) FoundryOutputsByID(foundryID iotago.FoundryID) (<-chan iotago.Output, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPIFoundryOutputs, "{foundryId}", foundryID.String(), 1)

	return eac.subscribeToOutputsTopic(topic)
}

// OutputsByUnlockConditionAndAddress returns a channel of newly created outputs on the given unlock condition and address.
func (eac *EventAPIClient) OutputsByUnlockConditionAndAddress(addr iotago.Address, netPrefix iotago.NetworkPrefix, condition EventAPIUnlockCondition) (<-chan iotago.Output, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPIOutputsByUnlockConditionAndAddress, "{address}", addr.Bech32(netPrefix), 1)
	topic = strings.Replace(topic, "{condition}", string(condition), 1)

	return eac.subscribeToOutputsTopic(topic)
}

// SpentOutputsByUnlockConditionAndAddress returns a channel of newly spent outputs on the given unlock condition and address.
func (eac *EventAPIClient) SpentOutputsByUnlockConditionAndAddress(addr iotago.Address, netPrefix iotago.NetworkPrefix, condition EventAPIUnlockCondition) (<-chan iotago.Output, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPISpentOutputsByUnlockConditionAndAddress, "{address}", addr.Bech32(netPrefix), 1)
	topic = strings.Replace(topic, "{condition}", string(condition), 1)

	return eac.subscribeToOutputsTopic(topic)
}

// TransactionIncludedBlock returns a channel of the included block which carries the transaction with the given ID.
func (eac *EventAPIClient) TransactionIncludedBlock(txID iotago.TransactionID) (<-chan *iotago.Block, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPITransactionsIncludedBlock, "{transactionId}", txID.ToHex(), 1)

	return eac.subscribeToBlocksTopic(topic)
}

// Output returns a channel which immediately returns the output with the given ID and afterward when its state changes.
func (eac *EventAPIClient) Output(outputID iotago.OutputID) (<-chan iotago.Output, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPIOutputs, "{outputId}", hexutil.EncodeHex(outputID[:]), 1)

	return eac.subscribeToOutputsTopic(topic)
}

// OutputMetadata returns a channel which immediately returns the output metadata with the given ID and afterward when its state changes.
func (eac *EventAPIClient) OutputMetadata(outputID iotago.OutputID) (<-chan *api.OutputMetadata, *EventAPIClientSubscription) {
	topic := strings.Replace(EventAPIOutputMetadata, "{outputId}", hexutil.EncodeHex(outputID[:]), 1)

	return subscribeToTopic(eac, topic, func(payload []byte) (*api.OutputMetadata, error) {
		response := new(api.OutputMetadata)
		if err := eac.Client.CommittedAPI().JSONDecode(payload, response); err != nil {
			sendErrOrDrop(eac.Errors, err)
			return nil, err
		}

		return response, nil
	})
}
