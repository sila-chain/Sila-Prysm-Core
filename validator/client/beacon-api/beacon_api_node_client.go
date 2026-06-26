package beacon_api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/sila-chain/Sila-Consensus-Core/v7/api/rest"
	"github.com/sila-chain/Sila-Consensus-Core/v7/api/server/structs"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/validator/client/iface"
	"github.com/sila-chain/Sila/common/hexutil"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	_ = iface.NodeClient(&beaconApiNodeClient{})
)

type beaconApiNodeClient struct {
	fallbackClient  iface.NodeClient
	handler         rest.Handler
	genesisProvider GenesisProvider
}

func (c *beaconApiNodeClient) SyncStatus(ctx context.Context, _ *empty.Empty) (*silapb.SyncStatus, error) {
	syncingResponse := structs.SyncStatusResponse{}
	if err := c.handler.Get(ctx, "/sila/v1/node/syncing", &syncingResponse); err != nil {
		return nil, err
	}

	if syncingResponse.Data == nil {
		return nil, errors.New("syncing data is nil")
	}

	return &silapb.SyncStatus{
		Syncing: syncingResponse.Data.IsSyncing,
	}, nil
}

func (c *beaconApiNodeClient) Genesis(ctx context.Context, _ *empty.Empty) (*silapb.Genesis, error) {
	genesisJson, err := c.genesisProvider.Genesis(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get genesis")
	}

	genesisValidatorRoot, err := hexutil.Decode(genesisJson.GenesisValidatorsRoot)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode genesis validator root `%s`", genesisJson.GenesisValidatorsRoot)
	}

	genesisTime, err := strconv.ParseInt(genesisJson.GenesisTime, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse genesis time `%s`", genesisJson.GenesisTime)
	}

	silaDepositJson := structs.GetSilaDepositResponse{}
	if err = c.handler.Get(ctx, "/sila/v1/config/sila_deposit", &silaDepositJson); err != nil {
		return nil, err
	}

	if silaDepositJson.Data == nil {
		return nil, errors.New("sila deposit data is nil")
	}

	depositContactAddress, err := hexutil.Decode(silaDepositJson.Data.Address)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode sila deposit address `%s`", silaDepositJson.Data.Address)
	}

	return &silapb.Genesis{
		GenesisTime: &timestamppb.Timestamp{
			Seconds: genesisTime,
		},
		SilaDepositAddress: depositContactAddress,
		GenesisValidatorsRoot:  genesisValidatorRoot,
	}, nil
}

func (c *beaconApiNodeClient) Version(ctx context.Context, _ *empty.Empty) (*silapb.Version, error) {
	var versionResponse structs.GetVersionResponse
	if err := c.handler.Get(ctx, "/sila/v1/node/version", &versionResponse); err != nil {
		return nil, err
	}

	if versionResponse.Data == nil || versionResponse.Data.Version == "" {
		return nil, errors.New("empty version response")
	}

	return &silapb.Version{
		Version: versionResponse.Data.Version,
	}, nil
}

func (c *beaconApiNodeClient) Peers(ctx context.Context, in *empty.Empty) (*silapb.Peers, error) {
	if c.fallbackClient != nil {
		return c.fallbackClient.Peers(ctx, in)
	}

	// TODO: Implement me
	return nil, errors.New("beaconApiNodeClient.Peers is not implemented. To use a fallback client, pass a fallback client as the last argument of NewBeaconApiNodeClientWithFallback.")
}

// IsReady returns true only if the node is fully synced (200 OK).
// A 206 Partial Content response indicates the node is syncing and not ready.
func (c *beaconApiNodeClient) IsReady(ctx context.Context) bool {
	statusCode, err := c.handler.GetStatusCode(ctx, "/sila/v1/node/health")
	if err != nil {
		log.WithError(err).WithField("url", c.handler.Host()).Error("failed to get health of node")
		return false
	}
	// Only 200 OK means the node is fully synced and ready.
	// 206 Partial Content means syncing, 503 means unavailable.
	return statusCode == http.StatusOK
}

func NewNodeClientWithFallback(handler rest.Handler, fallbackClient iface.NodeClient) iface.NodeClient {
	b := &beaconApiNodeClient{
		handler:         handler,
		fallbackClient:  fallbackClient,
		genesisProvider: &beaconApiGenesisProvider{handler: handler},
	}
	return b
}
