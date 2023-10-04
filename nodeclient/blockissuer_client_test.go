package nodeclient_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/nodeclient"
	"github.com/iotaledger/iota.go/v4/nodeclient/apimodels"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestBlockIssuerClient_Enabled(t *testing.T) {
	defer gock.Off()

	originRoutes := &apimodels.RoutesResponse{
		Routes: []string{"blockissuer/v1"},
	}

	mockGetJSON(nodeclient.RouteRoutes, 200, originRoutes)

	client := nodeClient(t)

	_, err := client.BlockIssuer(context.TODO())
	require.NoError(t, err)
}

func TestBlockIssuerClient_Disabled(t *testing.T) {
	defer gock.Off()

	originRoutes := &apimodels.RoutesResponse{
		Routes: []string{"someplugin/v1"},
	}

	mockGetJSON(nodeclient.RouteRoutes, 200, originRoutes)

	client := nodeClient(t)

	_, err := client.BlockIssuer(context.TODO())
	require.Error(t, err, nodeclient.ErrBlockIssuerPluginNotAvailable)
}

func TestBlockIssuerClient_Info(t *testing.T) {
	defer gock.Off()

	infoResponse := &nodeclient.BlockIssuerInfo{
		BlockIssuerAddress:     tpkg.RandAccountAddress().Bech32(iotago.PrefixTestnet),
		PowTargetTrailingZeros: 25,
	}

	originRoutes := &apimodels.RoutesResponse{
		Routes: []string{"blockissuer/v1"},
	}

	mockGetJSON(nodeclient.RouteRoutes, 200, originRoutes)
	mockGetJSON(nodeclient.BlockIssuerAPIRouteInfo, 200, infoResponse)

	client := nodeClient(t)

	blockIssuer, err := client.BlockIssuer(context.TODO())
	require.NoError(t, err)

	result, err := blockIssuer.Info(context.TODO())
	require.NoError(t, err)

	require.Equal(t, infoResponse, result)
}
