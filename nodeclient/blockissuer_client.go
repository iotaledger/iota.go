package nodeclient

import (
	"context"
	"net/http"
	"strconv"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/blockissuer/pow"
	"github.com/iotaledger/iota.go/v4/builder"
	"github.com/iotaledger/iota.go/v4/nodeclient/apimodels"
)

const (
	HeaderBlockIssuerProofOfWorkNonce = "X-IOTA-BlockIssuer-PoW-Nonce"
	HeaderBlockIssuerCommitmentID     = "X-IOTA-BlockIssuer-Commitment-ID"
)

// BlockIssuer plugin routes.
var (
	BlockIssuerAPIRouteInfo         = RootAPI + "/" + BlockIssuerPluginName + "/info"
	BlockIssuerAPIRouteIssuePayload = RootAPI + "/" + BlockIssuerPluginName + "/issue"
)

type (
	// BlockIssuerClient is a client which queries the optional blockissuer functionality of a node.
	BlockIssuerClient interface {
		// Info returns the info of the block issuer.
		Info(ctx context.Context) (*apimodels.BlockIssuerInfo, error)
		// SendPayload sends a BlockPayload to the block issuer.
		SendPayload(ctx context.Context, payload iotago.BlockPayload, commitmentID iotago.CommitmentID, numPoWWorkers ...int) (*apimodels.BlockCreatedResponse, error)
		// SendPayloadWithTransactionBuilder automatically allots the needed mana and sends a BlockPayload to the block issuer.
		SendPayloadWithTransactionBuilder(ctx context.Context, builder *builder.TransactionBuilder, signer iotago.AddressSigner, storedManaOutputIndex int, numPoWWorkers ...int) (iotago.BlockPayload, *apimodels.BlockCreatedResponse, error)
	}

	blockIssuerClient struct {
		core *Client
	}
)

// Do executes a request against the endpoint.
// This function is only meant to be used for special routes not covered through the standard API.
func (client *blockIssuerClient) Do(ctx context.Context, method string, route string, reqObj interface{}, resObj interface{}) (*http.Response, error) {
	return client.core.Do(ctx, method, route, reqObj, resObj)
}

// DoWithRequestHeaderHook executes a request against the endpoint.
// This function is only meant to be used for special routes not covered through the standard API.
func (client *blockIssuerClient) DoWithRequestHeaderHook(ctx context.Context, method string, route string, requestHeaderHook RequestHeaderHook, reqObj interface{}, resObj interface{}) (*http.Response, error) {
	return client.core.DoWithRequestHeaderHook(ctx, method, route, requestHeaderHook, reqObj, resObj)
}

func (client *blockIssuerClient) Info(ctx context.Context) (*apimodels.BlockIssuerInfo, error) {
	res := new(apimodels.BlockIssuerInfo)

	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, BlockIssuerAPIRouteInfo, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (client *blockIssuerClient) mineNonceAndSendPayload(ctx context.Context, payload iotago.BlockPayload, commitmentID iotago.CommitmentID, powTargetTrailingZeros uint8, numPoWWorkers ...int) (*apimodels.BlockCreatedResponse, error) {
	payloadBytes, err := client.core.CommittedAPI().Encode(payload, serix.WithValidation())
	if err != nil {
		return nil, ierrors.Wrap(err, "failed to encode the payload")
	}

	powWorker := pow.New(numPoWWorkers...)
	nonce, err := powWorker.Mine(ctx, payloadBytes, int(powTargetTrailingZeros))
	if err != nil {
		return nil, ierrors.Wrap(err, "failed to mine the nonce for proof of work")
	}

	requestHeaderHook := func(header http.Header) {
		RequestHeaderHookContentTypeIOTASerializerV2(header)

		header.Set(HeaderBlockIssuerCommitmentID, commitmentID.ToHex())
		header.Set(HeaderBlockIssuerProofOfWorkNonce, strconv.FormatUint(nonce, 10))
	}

	req := &RawDataEnvelope{Data: payloadBytes}

	res := new(apimodels.BlockCreatedResponse)
	//nolint:bodyclose // false positive
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodPost, BlockIssuerAPIRouteIssuePayload, requestHeaderHook, req, res); err != nil {
		return nil, ierrors.Wrap(err, "failed to send the payload issuance request")
	}

	return res, nil
}

func (client *blockIssuerClient) SendPayload(ctx context.Context, payload iotago.BlockPayload, commitmentID iotago.CommitmentID, numPoWWorkers ...int) (*apimodels.BlockCreatedResponse, error) {
	// get the info from the block issuer
	blockIssuerInfo, err := client.Info(ctx)
	if err != nil {
		return nil, ierrors.Wrap(err, "failed to get the block issuer info")
	}

	return client.mineNonceAndSendPayload(ctx, payload, commitmentID, blockIssuerInfo.PowTargetTrailingZeros, numPoWWorkers...)
}

func (client *blockIssuerClient) SendPayloadWithTransactionBuilder(ctx context.Context, builder *builder.TransactionBuilder, signer iotago.AddressSigner, storedManaOutputIndex int, numPoWWorkers ...int) (iotago.BlockPayload, *apimodels.BlockCreatedResponse, error) {
	// get the info from the block issuer
	blockIssuerInfo, err := client.Info(ctx)
	if err != nil {
		return nil, nil, ierrors.Wrap(err, "failed to get the block issuer info")
	}

	// parse the block issuer address
	//nolint:contextcheck // false positive
	_, blockIssuerAddress, err := iotago.ParseBech32(blockIssuerInfo.BlockIssuerAddress)
	if err != nil {
		return nil, nil, ierrors.Wrap(err, "failed to parse the block issuer address")
	}

	// check if the block issuer address is an account address
	blockIssuerAccountAddress, isAccount := blockIssuerAddress.(*iotago.AccountAddress)
	if !isAccount {
		return nil, nil, ierrors.New("failed to parse the block issuer address")
	}

	// get the current commitmentID and reference mana cost to calculate
	// the correct value for the mana that needs to be alloted to the block issuer.
	blockIssuance, err := client.core.BlockIssuance(ctx)
	if err != nil {
		return nil, nil, ierrors.Wrap(err, "failed to get the latest block issuance infos")
	}

	// allot the required mana to the block issuer
	builder.AllotRequiredManaAndStoreRemainingManaInOutput(builder.CreationSlot(), blockIssuance.Commitment.ReferenceManaCost, blockIssuerAccountAddress.AccountID(), storedManaOutputIndex)

	// sign the transaction
	payload, err := builder.Build(signer)
	if err != nil {
		return nil, nil, ierrors.Wrap(err, "failed to build the signed transaction payload")
	}

	//nolint:contextcheck // false positive
	commitmentID, err := blockIssuance.Commitment.ID()
	if err != nil {
		return nil, nil, ierrors.Wrap(err, "failed to calculate the commitment ID")
	}

	blockCreatedResponse, err := client.mineNonceAndSendPayload(ctx, payload, commitmentID, blockIssuerInfo.PowTargetTrailingZeros, numPoWWorkers...)
	if err != nil {
		return nil, nil, err
	}

	return payload, blockCreatedResponse, nil
}
