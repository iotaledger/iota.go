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
		panic(ierrors.WithMessage(ErrEventAPIClientInactive, "context is canceled/done"))
	}
	if !neac.MQTTClient.IsConnected() {
		panic(ierrors.WithMessage(ErrEventAPIClientInactive, "client is not connected"))
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

func (eac *EventAPIClient) subscribeToCommitmentsTopicRaw(topic string) (<-chan *iotago.Commitment, *EventAPIClientSubscription) {
	return subscribeToTopic(eac, topic+api.EventAPITopicSuffixRaw, func(payload []byte) (*iotago.Commitment, error) {
		response := new(iotago.Commitment)
		if _, err := eac.Client.CommittedAPI().Decode(payload, response); err != nil {
			sendErrOrDrop(eac.Errors, err)
			return nil, err
		}

		return response, nil
	})
}

func (eac *EventAPIClient) subscribeToBlocksTopicRaw(topic string) (<-chan *iotago.Block, *EventAPIClientSubscription) {
	return subscribeToTopic(eac, topic+api.EventAPITopicSuffixRaw, func(payload []byte) (*iotago.Block, error) {
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

func (eac *EventAPIClient) subscribeToTransactionMetadataTopicRaw(topic string) (<-chan *api.TransactionMetadataResponse, *EventAPIClientSubscription) {
	return subscribeToTopic(eac, topic+api.EventAPITopicSuffixRaw, func(payload []byte) (*api.TransactionMetadataResponse, error) {
		response := new(api.TransactionMetadataResponse)
		if _, err := eac.Client.CommittedAPI().Decode(payload, response); err != nil {
			sendErrOrDrop(eac.Errors, err)
			return nil, err
		}

		return response, nil
	})
}

func (eac *EventAPIClient) subscribeToBlockMetadataTopicRaw(topic string) (<-chan *api.BlockMetadataResponse, *EventAPIClientSubscription) {
	return subscribeToTopic(eac, topic+api.EventAPITopicSuffixRaw, func(payload []byte) (*api.BlockMetadataResponse, error) {
		response := new(api.BlockMetadataResponse)
		if _, err := eac.Client.CommittedAPI().Decode(payload, response); err != nil {
			sendErrOrDrop(eac.Errors, err)
			return nil, err
		}

		return response, nil
	})
}

func (eac *EventAPIClient) subscribeToOutputsWithMetadataTopicRaw(topic string) (<-chan *api.OutputWithMetadataResponse, *EventAPIClientSubscription) {
	return subscribeToTopic(eac, topic+api.EventAPITopicSuffixRaw, func(payload []byte) (*api.OutputWithMetadataResponse, error) {
		response := new(api.OutputWithMetadataResponse)
		if _, err := eac.Client.CommittedAPI().Decode(payload, response); err != nil {
			sendErrOrDrop(eac.Errors, err)
			return nil, err
		}

		return response, nil
	})
}

func (eac *EventAPIClient) CommitmentsLatest() (<-chan *iotago.Commitment, *EventAPIClientSubscription) {
	return eac.subscribeToCommitmentsTopicRaw(api.EventAPITopicCommitmentsLatest)
}

func (eac *EventAPIClient) CommitmentsFinalized() (<-chan *iotago.Commitment, *EventAPIClientSubscription) {
	return eac.subscribeToCommitmentsTopicRaw(api.EventAPITopicCommitmentsFinalized)
}

// Blocks returns a channel of newly received blocks.
func (eac *EventAPIClient) Blocks() (<-chan *iotago.Block, *EventAPIClientSubscription) {
	return eac.subscribeToBlocksTopicRaw(api.EventAPITopicBlocks)
}

// BlocksValidation returns a channel of newly received validation blocks.
func (eac *EventAPIClient) BlocksValidation() (<-chan *iotago.Block, *EventAPIClientSubscription) {
	return eac.subscribeToBlocksTopicRaw(api.EventAPITopicBlocksValidation)
}

// BlocksBasic returns a channel of newly received basic blocks.
func (eac *EventAPIClient) BlocksBasic() (<-chan *iotago.Block, *EventAPIClientSubscription) {
	return eac.subscribeToBlocksTopicRaw(api.EventAPITopicBlocksBasic)
}

// BlocksBasicWithTaggedData returns a channel of blocks containing tagged data containing the given tag.
func (eac *EventAPIClient) BlocksBasicWithTaggedData() (<-chan *iotago.Block, *EventAPIClientSubscription) {
	return eac.subscribeToBlocksTopicRaw(api.EventAPITopicBlocksBasicTaggedData)
}

// BlocksBasicWithTaggedDataByTag returns a channel of blocks containing tagged data.
func (eac *EventAPIClient) BlocksBasicWithTaggedDataByTag(tag []byte) (<-chan *iotago.Block, *EventAPIClientSubscription) {
	topic := api.EndpointWithNamedParameterValue(api.EventAPITopicBlocksBasicTaggedDataTag, api.ParameterTag, hexutil.EncodeHex(tag))

	return eac.subscribeToBlocksTopicRaw(topic)
}

// BlocksBasicWithTransactions returns a channel of blocks containing transactions.
func (eac *EventAPIClient) BlocksBasicWithTransactions() (<-chan *iotago.Block, *EventAPIClientSubscription) {
	return eac.subscribeToBlocksTopicRaw(api.EventAPITopicBlocksBasicTransaction)
}

// BlocksBasicWithTransactionsWithTaggedData returns a channel of blocks containing transactions with tagged data.
func (eac *EventAPIClient) BlocksBasicWithTransactionsWithTaggedData() (<-chan *iotago.Block, *EventAPIClientSubscription) {
	return eac.subscribeToBlocksTopicRaw(api.EventAPITopicBlocksBasicTransactionTaggedData)
}

// BlocksBasicWithTransactionsWithTaggedDataByTag returns a channel of blocks containing transactions with tagged data containing the given tag.
func (eac *EventAPIClient) BlocksBasicWithTransactionsWithTaggedDataByTag(tag []byte) (<-chan *iotago.Block, *EventAPIClientSubscription) {
	topic := api.EndpointWithNamedParameterValue(api.EventAPITopicBlocksBasicTransactionTaggedDataTag, api.ParameterTag, hexutil.EncodeHex(tag))

	return eac.subscribeToBlocksTopicRaw(topic)
}

// BlockMetadataTransactionIncludedBlocksByTransactionID returns a channel of BlockMetadataResponse of blocks which carry the transaction with the given ID.
func (eac *EventAPIClient) BlockMetadataTransactionIncludedBlocksByTransactionID(txID iotago.TransactionID) (<-chan *api.BlockMetadataResponse, *EventAPIClientSubscription) {
	topic := api.EndpointWithNamedParameterValue(api.EventAPITopicTransactionsIncludedBlockMetadata, api.ParameterTransactionID, txID.ToHex())

	return eac.subscribeToBlockMetadataTopicRaw(topic)
}

// TransactionMetadataByTransactionID returns a channel of TransactionMetadataResponse each time the given transaction's state changes.
func (eac *EventAPIClient) TransactionMetadataByTransactionID(txID iotago.TransactionID) (<-chan *api.TransactionMetadataResponse, *EventAPIClientSubscription) {
	topic := api.EndpointWithNamedParameterValue(api.EventAPITopicTransactionMetadata, api.ParameterTransactionID, txID.ToHex())

	return eac.subscribeToTransactionMetadataTopicRaw(topic)
}

// BlockMetadataByBlockID returns a channel of BlockMetadataResponse each time the given block's state changes.
func (eac *EventAPIClient) BlockMetadataByBlockID(blockID iotago.BlockID) (<-chan *api.BlockMetadataResponse, *EventAPIClientSubscription) {
	topic := api.EndpointWithNamedParameterValue(api.EventAPITopicBlockMetadata, api.ParameterBlockID, blockID.ToHex())

	return eac.subscribeToBlockMetadataTopicRaw(topic)
}

// BlockMetadataAcceptedBlocks returns a channel of BlockMetadataResponse of newly accepted blocks.
func (eac *EventAPIClient) BlockMetadataAcceptedBlocks() (<-chan *api.BlockMetadataResponse, *EventAPIClientSubscription) {
	return eac.subscribeToBlockMetadataTopicRaw(api.EventAPITopicBlockMetadataAccepted)
}

// BlockMetadataConfirmedBlocks returns a channel of BlockMetadataResponse of newly confirmed blocks.
func (eac *EventAPIClient) BlockMetadataConfirmedBlocks() (<-chan *api.BlockMetadataResponse, *EventAPIClientSubscription) {
	return eac.subscribeToBlockMetadataTopicRaw(api.EventAPITopicBlockMetadataConfirmed)
}

// OutputWithMetadataByOutputID returns a channel which immediately returns the output with the given ID and afterward when its state changes.
func (eac *EventAPIClient) OutputWithMetadataByOutputID(outputID iotago.OutputID) (<-chan *api.OutputWithMetadataResponse, *EventAPIClientSubscription) {
	topic := api.EndpointWithNamedParameterValue(api.EventAPITopicOutputs, api.ParameterOutputID, outputID.ToHex())

	return eac.subscribeToOutputsWithMetadataTopicRaw(topic)
}

// OutputsWithMetadataByAccountID returns a channel of newly created outputs to track the chain mutations of a given Account.
func (eac *EventAPIClient) OutputsWithMetadataByAccountID(accountID iotago.AccountID) (<-chan *api.OutputWithMetadataResponse, *EventAPIClientSubscription) {
	topic := api.EndpointWithNamedParameterValue(api.EventAPITopicAccountOutputs, api.ParameterAccountAddress, accountID.ToAddress().Bech32(eac.Client.CommittedAPI().ProtocolParameters().Bech32HRP()))

	return eac.subscribeToOutputsWithMetadataTopicRaw(topic)
}

// OutputsWithMetadataByAnchorID returns a channel of newly created outputs to track the chain mutations of a given anchor ID.
func (eac *EventAPIClient) OutputsWithMetadataByAnchorID(anchorID iotago.AnchorID) (<-chan *api.OutputWithMetadataResponse, *EventAPIClientSubscription) {
	topic := api.EndpointWithNamedParameterValue(api.EventAPITopicAnchorOutputs, api.ParameterAnchorAddress, anchorID.ToAddress().Bech32(eac.Client.CommittedAPI().ProtocolParameters().Bech32HRP()))

	return eac.subscribeToOutputsWithMetadataTopicRaw(topic)
}

// OutputsWithMetadataByFoundryID returns a channel of newly created outputs to track the chain mutations of a given Foundry.
func (eac *EventAPIClient) OutputsWithMetadataByFoundryID(foundryID iotago.FoundryID) (<-chan *api.OutputWithMetadataResponse, *EventAPIClientSubscription) {
	topic := api.EndpointWithNamedParameterValue(api.EventAPITopicFoundryOutputs, api.ParameterFoundryID, foundryID.ToHex())

	return eac.subscribeToOutputsWithMetadataTopicRaw(topic)
}

// OutputsWithMetadataByNFTID returns a channel of newly created outputs to track the chain mutations of a given NFT.
func (eac *EventAPIClient) OutputsWithMetadataByNFTID(nftID iotago.NFTID) (<-chan *api.OutputWithMetadataResponse, *EventAPIClientSubscription) {
	topic := api.EndpointWithNamedParameterValue(api.EventAPITopicNFTOutputs, api.ParameterNFTAddress, nftID.ToAddress().Bech32(eac.Client.CommittedAPI().ProtocolParameters().Bech32HRP()))

	return eac.subscribeToOutputsWithMetadataTopicRaw(topic)
}

// OutputsWithMetadataByDelegationID returns a channel of newly created outputs to track the chain mutations of a given delegation ID.
func (eac *EventAPIClient) OutputsWithMetadataByDelegationID(delegationID iotago.DelegationID) (<-chan *api.OutputWithMetadataResponse, *EventAPIClientSubscription) {
	topic := api.EndpointWithNamedParameterValue(api.EventAPITopicDelegationOutputs, api.ParameterDelegationID, delegationID.ToHex())

	return eac.subscribeToOutputsWithMetadataTopicRaw(topic)
}

// OutputsWithMetadataByUnlockConditionAndAddress returns a channel of newly created outputs on the given unlock condition and address.
func (eac *EventAPIClient) OutputsWithMetadataByUnlockConditionAndAddress(condition api.EventAPIUnlockCondition, addr iotago.Address) (<-chan *api.OutputWithMetadataResponse, *EventAPIClientSubscription) {
	topic := api.EndpointWithNamedParameterValue(api.EventAPITopicOutputsByUnlockConditionAndAddress, api.ParameterCondition, string(condition))
	topic = api.EndpointWithNamedParameterValue(topic, api.ParameterAddress, addr.Bech32(eac.Client.CommittedAPI().ProtocolParameters().Bech32HRP()))

	return eac.subscribeToOutputsWithMetadataTopicRaw(topic)
}
