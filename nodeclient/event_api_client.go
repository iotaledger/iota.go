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

var (
	// ErrEventAPIClientInactive gets returned when an EventAPIClient is inactive.
	ErrEventAPIClientInactive = ierrors.New("event api client is inactive")
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

func (eac *EventAPIClient) subscribeToOutputsWithMetadataTopic(topic string) (<-chan *api.OutputWithMetadataResponse, *EventAPIClientSubscription) {
	return subscribeToTopic(eac, topic, func(payload []byte) (*api.OutputWithMetadataResponse, error) {
		response := new(api.OutputWithMetadataResponse)
		if err := eac.Client.CommittedAPI().JSONDecode(payload, response); err != nil {
			sendErrOrDrop(eac.Errors, err)
			return nil, err
		}

		return response, nil
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

func (eac *EventAPIClient) subscribeToTransactionMetadataTopic(topic string) (<-chan *api.TransactionMetadataResponse, *EventAPIClientSubscription) {
	return subscribeToTopic(eac, topic, func(payload []byte) (*api.TransactionMetadataResponse, error) {
		response := new(api.TransactionMetadataResponse)
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
	return eac.subscribeToBlocksTopic(api.TopicBlocks)
}

// ValidationBlocks returns a channel of newly received validation blocks.
func (eac *EventAPIClient) ValidationBlocks() (<-chan *iotago.Block, *EventAPIClientSubscription) {
	return eac.subscribeToBlocksTopic(api.TopicBlocksValidation)
}

// BasicBlocks returns a channel of newly received basic blocks.
func (eac *EventAPIClient) BasicBlocks() (<-chan *iotago.Block, *EventAPIClientSubscription) {
	return eac.subscribeToBlocksTopic(api.TopicBlocksBasic)
}

// TaggedDataBlocks returns a channel of blocks containing tagged data containing the given tag.
func (eac *EventAPIClient) TaggedDataBlocks() (<-chan *iotago.Block, *EventAPIClientSubscription) {
	return eac.subscribeToBlocksTopic(api.TopicBlocksBasicTaggedData)
}

// TaggedDataWithTagBlocks returns a channel of blocks containing tagged data.
func (eac *EventAPIClient) TaggedDataWithTagBlocks(tag []byte) (<-chan *iotago.Block, *EventAPIClientSubscription) {
	topic := api.EndpointWithNamedParameterValue(api.TopicBlocksBasicTaggedDataTag, api.ParameterTag, hexutil.EncodeHex(tag))

	return eac.subscribeToBlocksTopic(topic)
}

// TransactionBlocks returns a channel of blocks containing transactions.
func (eac *EventAPIClient) TransactionBlocks() (<-chan *iotago.Block, *EventAPIClientSubscription) {
	return eac.subscribeToBlocksTopic(api.TopicBlocksBasicTransaction)
}

// TransactionTaggedDataBlocks returns a channel of blocks containing transactions with tagged data.
func (eac *EventAPIClient) TransactionTaggedDataBlocks() (<-chan *iotago.Block, *EventAPIClientSubscription) {
	return eac.subscribeToBlocksTopic(api.TopicBlocksBasicTransactionTaggedData)
}

// TransactionTaggedDataWithTagBlocks returns a channel of blocks containing transactions with tagged data containing the given tag.
func (eac *EventAPIClient) TransactionTaggedDataWithTagBlocks(tag []byte) (<-chan *iotago.Block, *EventAPIClientSubscription) {
	topic := api.EndpointWithNamedParameterValue(api.TopicBlocksBasicTransactionTaggedDataTag, api.ParameterTag, hexutil.EncodeHex(tag))

	return eac.subscribeToBlocksTopic(topic)
}

// TransactionIncludedBlock returns a channel of the included block which carries the transaction with the given ID.
func (eac *EventAPIClient) TransactionIncludedBlock(txID iotago.TransactionID) (<-chan *iotago.Block, *EventAPIClientSubscription) {
	topic := api.EndpointWithNamedParameterValue(api.TopicTransactionsIncludedBlock, api.ParameterTransactionID, txID.ToHex())

	return eac.subscribeToBlocksTopic(topic)
}

// TransactionMetadataChange returns a channel of TransactionMetadataResponse each time the given transaction's state changes.
func (eac *EventAPIClient) TransactionMetadataChange(txID iotago.TransactionID) (<-chan *api.TransactionMetadataResponse, *EventAPIClientSubscription) {
	topic := api.EndpointWithNamedParameterValue(api.TopicTransactionMetadata, api.ParameterTransactionID, txID.ToHex())

	return eac.subscribeToTransactionMetadataTopic(topic)
}

// BlockMetadataChange returns a channel of BlockMetadataResponse each time the given block's state changes.
func (eac *EventAPIClient) BlockMetadataChange(blockID iotago.BlockID) (<-chan *api.BlockMetadataResponse, *EventAPIClientSubscription) {
	topic := api.EndpointWithNamedParameterValue(api.TopicBlockMetadata, api.ParameterBlockID, blockID.ToHex())

	return eac.subscribeToBlockMetadataTopic(topic)
}

// AcceptedBlocksMetadata returns a channel of BlockMetadataResponse of newly accepted blocks.
func (eac *EventAPIClient) AcceptedBlocksMetadata() (<-chan *api.BlockMetadataResponse, *EventAPIClientSubscription) {
	return eac.subscribeToBlockMetadataTopic(api.TopicBlockMetadataAccepted)
}

// ConfirmedBlocksMetadata returns a channel of BlockMetadataResponse of newly confirmed blocks.
func (eac *EventAPIClient) ConfirmedBlocksMetadata() (<-chan *api.BlockMetadataResponse, *EventAPIClientSubscription) {
	return eac.subscribeToBlockMetadataTopic(api.TopicBlockMetadataConfirmed)
}

// Output returns a channel which immediately returns the output with the given ID and afterward when its state changes.
func (eac *EventAPIClient) Output(outputID iotago.OutputID) (<-chan *api.OutputWithMetadataResponse, *EventAPIClientSubscription) {
	topic := api.EndpointWithNamedParameterValue(api.TopicOutputs, api.ParameterOutputID, outputID.ToHex())

	return eac.subscribeToOutputsWithMetadataTopic(topic)
}

// NFTOutputsByID returns a channel of newly created outputs to track the chain mutations of a given NFT.
func (eac *EventAPIClient) NFTOutputsByID(nftID iotago.NFTID, hrp iotago.NetworkPrefix) (<-chan *api.OutputWithMetadataResponse, *EventAPIClientSubscription) {
	topic := api.EndpointWithNamedParameterValue(api.TopicNFTOutputs, api.ParameterNFTAddress, nftID.ToAddress().Bech32(hrp))

	return eac.subscribeToOutputsWithMetadataTopic(topic)
}

// AccountOutputsByID returns a channel of newly created outputs to track the chain mutations of a given Account.
func (eac *EventAPIClient) AccountOutputsByID(accountID iotago.AccountID, hrp iotago.NetworkPrefix) (<-chan *api.OutputWithMetadataResponse, *EventAPIClientSubscription) {
	topic := api.EndpointWithNamedParameterValue(api.TopicAccountOutputs, api.ParameterAccountAddress, accountID.ToAddress().Bech32(hrp))

	return eac.subscribeToOutputsWithMetadataTopic(topic)
}

// AnchorOutputsByID returns a channel of newly created outputs to track the chain mutations of a given anchor ID.
func (eac *EventAPIClient) AnchorOutputsByID(anchorID iotago.AnchorID, hrp iotago.NetworkPrefix) (<-chan *api.OutputWithMetadataResponse, *EventAPIClientSubscription) {
	topic := api.EndpointWithNamedParameterValue(api.TopicAnchorOutputs, api.ParameterAnchorAddress, anchorID.ToAddress().Bech32(hrp))

	return eac.subscribeToOutputsWithMetadataTopic(topic)
}

// FoundryOutputsByID returns a channel of newly created outputs to track the chain mutations of a given Foundry.
func (eac *EventAPIClient) FoundryOutputsByID(foundryID iotago.FoundryID) (<-chan *api.OutputWithMetadataResponse, *EventAPIClientSubscription) {
	topic := api.EndpointWithNamedParameterValue(api.TopicFoundryOutputs, api.ParameterFoundryID, foundryID.ToHex())

	return eac.subscribeToOutputsWithMetadataTopic(topic)
}

// DelegationOutputsByID returns a channel of newly created outputs to track the chain mutations of a given delegation ID.
func (eac *EventAPIClient) DelegationOutputsByID(delegationID iotago.DelegationID) (<-chan *api.OutputWithMetadataResponse, *EventAPIClientSubscription) {
	topic := api.EndpointWithNamedParameterValue(api.TopicDelegationOutputs, api.ParameterDelegationID, delegationID.ToHex())

	return eac.subscribeToOutputsWithMetadataTopic(topic)
}

// OutputsByUnlockConditionAndAddress returns a channel of newly created outputs on the given unlock condition and address.
func (eac *EventAPIClient) OutputsByUnlockConditionAndAddress(addr iotago.Address, hrp iotago.NetworkPrefix, condition api.EventAPIUnlockCondition) (<-chan *api.OutputWithMetadataResponse, *EventAPIClientSubscription) {
	topic := api.EndpointWithNamedParameterValue(api.TopicOutputsByUnlockConditionAndAddress, api.ParameterCondition, string(condition))
	topic = api.EndpointWithNamedParameterValue(topic, api.ParameterAddress, addr.Bech32(hrp))

	return eac.subscribeToOutputsWithMetadataTopic(topic)
}
