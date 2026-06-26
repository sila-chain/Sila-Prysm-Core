// Package node defines a gRPC node service implementation, providing
// useful endpoints for checking a node's sync status, peer info,
// genesis data, and version information.
package node

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/blockchain"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/db"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/p2p"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/silaexec"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/sync"
	"github.com/sila-chain/Sila-Consensus-Core/v7/io/logs"
	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/tracing/trace"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server defines a server implementation of the gRPC Node service,
// providing RPC endpoints for verifying a beacon node's sync status, genesis and
// version information, and services the node implements and runs.
type Server struct {
	LogsStreamer          logs.Streamer
	StreamLogsBufferSize  int
	SyncChecker           sync.Checker
	Server                *grpc.Server
	BeaconDB              db.ReadOnlyDatabase
	PeersFetcher          p2p.PeersProvider
	PeerManager           p2p.PeerManager
	GenesisTimeFetcher    blockchain.TimeFetcher
	GenesisFetcher        blockchain.GenesisFetcher
	POWChainInfoFetcher   silaexec.ChainInfoFetcher
	BeaconMonitoringHost  string
	BeaconMonitoringPort  int
	OptimisticModeFetcher blockchain.OptimisticModeFetcher
}

// Deprecated: The gRPC API will remain the default and fully supported through v8 (expected in 2026) but will be eventually removed in favor of REST API.
//
// GetHealth checks the health of the node
func (ns *Server) GetHealth(ctx context.Context, request *silapb.HealthRequest) (*empty.Empty, error) {
	ctx, span := trace.StartSpan(ctx, "node.GetHealth")
	defer span.End()

	// Set a timeout for the health check operation
	timeoutDuration := 10 * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeoutDuration)
	defer cancel() // Important to avoid a context leak

	// Check optimistic status - validators should not participate when optimistic
	isOptimistic, err := ns.OptimisticModeFetcher.IsOptimistic(ctx)
	if err != nil {
		return &empty.Empty{}, status.Errorf(codes.Internal, "Could not check optimistic status: %v", err)
	}

	if ns.SyncChecker.Synced() && !isOptimistic {
		return &empty.Empty{}, nil
	}
	if ns.SyncChecker.Syncing() || ns.SyncChecker.Initialized() {
		// Set header for REST API clients (via gRPC-gateway)
		if err := grpc.SetHeader(ctx, metadata.Pairs("x-http-code", strconv.FormatUint(http.StatusPartialContent, 10))); err != nil {
			return &empty.Empty{}, status.Errorf(codes.Internal, "Could not set status code header: %v", err)
		}
		return &empty.Empty{}, status.Error(codes.Unavailable, "node is syncing")
	}
	if isOptimistic {
		// Set header for REST API clients (via gRPC-gateway)
		if err := grpc.SetHeader(ctx, metadata.Pairs("x-http-code", strconv.FormatUint(http.StatusPartialContent, 10))); err != nil {
			return &empty.Empty{}, status.Errorf(codes.Internal, "Could not set status code header: %v", err)
		}
		return &empty.Empty{}, status.Error(codes.Unavailable, "node is optimistic")
	}
	return &empty.Empty{}, status.Errorf(codes.Unavailable, "service unavailable")
}

// Deprecated: The gRPC API will remain the default and fully supported through v8 (expected in 2026) but will be eventually removed in favor of REST API.
//
// GetSyncStatus checks the current network sync status of the node.
func (ns *Server) GetSyncStatus(_ context.Context, _ *empty.Empty) (*silapb.SyncStatus, error) {
	return &silapb.SyncStatus{
		Syncing: ns.SyncChecker.Syncing(),
	}, nil
}

// Deprecated: The gRPC API will remain the default and fully supported through v8 (expected in 2026) but will be eventually removed in favor of REST API.
//
// GetGenesis fetches genesis chain information of Sila. Returns unix timestamp 0
// if a genesis time has yet to be determined.
func (ns *Server) GetGenesis(ctx context.Context, _ *empty.Empty) (*silapb.Genesis, error) {
	contractAddr, err := ns.BeaconDB.SilaDepositAddress(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not retrieve contract address from db: %v", err)
	}
	genesisTime := ns.GenesisTimeFetcher.GenesisTime()
	var defaultGenesisTime time.Time
	var gt *timestamp.Timestamp
	if genesisTime == defaultGenesisTime {
		gt = timestamppb.New(time.Unix(0, 0))
	} else {
		gt = timestamppb.New(genesisTime)
	}

	genValRoot := ns.GenesisFetcher.GenesisValidatorsRoot()
	return &silapb.Genesis{
		GenesisTime:           gt,
		SilaDepositAddress:    contractAddr,
		GenesisValidatorsRoot: genValRoot[:],
	}, nil
}

// Deprecated: The gRPC API will remain the default and fully supported through v8 (expected in 2026) but will be eventually removed in favor of REST API.
//
// GetVersion checks the version information of the beacon node.
func (_ *Server) GetVersion(_ context.Context, _ *empty.Empty) (*silapb.Version, error) {
	return &silapb.Version{
		Version: version.Version(),
	}, nil
}

// Deprecated: The gRPC API will remain the default and fully supported through v8 (expected in 2026) but will be eventually removed in favor of REST API.
//
// ListImplementedServices lists the services implemented and enabled by this node.
//
// Any service not present in this list may return UNIMPLEMENTED or
// PERMISSION_DENIED. The server may also support fetching services by grpc
// reflection.
func (ns *Server) ListImplementedServices(_ context.Context, _ *empty.Empty) (*silapb.ImplementedServices, error) {
	serviceInfo := ns.Server.GetServiceInfo()
	serviceNames := make([]string, 0, len(serviceInfo))
	for svc := range serviceInfo {
		serviceNames = append(serviceNames, svc)
	}
	sort.Strings(serviceNames)
	return &silapb.ImplementedServices{
		Services: serviceNames,
	}, nil
}

// Deprecated: The gRPC API will remain the default and fully supported through v8 (expected in 2026) but will be eventually removed in favor of REST API.
//
// GetHost returns the p2p data on the current local and host peer.
func (ns *Server) GetHost(_ context.Context, _ *empty.Empty) (*silapb.HostData, error) {
	var stringAddr []string
	for _, addr := range ns.PeerManager.Host().Addrs() {
		stringAddr = append(stringAddr, addr.String())
	}
	record := ns.PeerManager.ENR()
	enr := ""
	var err error
	if record != nil {
		enr, err = p2p.SerializeENR(record)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Unable to serialize enr: %v", err)
		}
	}

	return &silapb.HostData{
		Addresses: stringAddr,
		PeerId:    ns.PeerManager.PeerID().String(),
		Enr:       enr,
	}, nil
}

// Deprecated: The gRPC API will remain the default and fully supported through v8 (expected in 2026) but will be eventually removed in favor of REST API.
//
// GetPeer returns the data known about the peer defined by the provided peer id.
func (ns *Server) GetPeer(_ context.Context, peerReq *silapb.PeerRequest) (*silapb.Peer, error) {
	pid, err := peer.Decode(peerReq.PeerId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to parse provided peer id: %v", err)
	}
	addr, err := ns.PeersFetcher.Peers().Address(pid)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Requested peer does not exist: %v", err)
	}
	dir, err := ns.PeersFetcher.Peers().Direction(pid)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Requested peer does not exist: %v", err)
	}
	pbDirection := silapb.PeerDirection_UNKNOWN
	switch dir {
	case network.DirInbound:
		pbDirection = silapb.PeerDirection_INBOUND
	case network.DirOutbound:
		pbDirection = silapb.PeerDirection_OUTBOUND
	}
	connState, err := ns.PeersFetcher.Peers().ConnectionState(pid)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Requested peer does not exist: %v", err)
	}
	record, err := ns.PeersFetcher.Peers().ENR(pid)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Requested peer does not exist: %v", err)
	}
	enr := ""
	if record != nil {
		enr, err = p2p.SerializeENR(record)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Unable to serialize enr: %v", err)
		}
	}
	return &silapb.Peer{
		Address:         addr.String(),
		Direction:       pbDirection,
		ConnectionState: silapb.ConnectionState(connState),
		PeerId:          peerReq.PeerId,
		Enr:             enr,
	}, nil
}

// Deprecated: The gRPC API will remain the default and fully supported through v8 (expected in 2026) but will be eventually removed in favor of REST API.
//
// ListPeers lists the peers connected to this node.
func (ns *Server) ListPeers(ctx context.Context, _ *empty.Empty) (*silapb.Peers, error) {
	peers := ns.PeersFetcher.Peers().Connected()
	res := make([]*silapb.Peer, 0, len(peers))
	for _, pid := range peers {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		multiaddr, err := ns.PeersFetcher.Peers().Address(pid)
		if err != nil {
			continue
		}
		direction, err := ns.PeersFetcher.Peers().Direction(pid)
		if err != nil {
			continue
		}
		record, err := ns.PeersFetcher.Peers().ENR(pid)
		if err != nil {
			continue
		}
		enr := ""
		if record != nil {
			enr, err = p2p.SerializeENR(record)
			if err != nil {
				continue
			}
		}
		multiAddrStr := "unknown"
		if multiaddr != nil {
			multiAddrStr = multiaddr.String()
		}
		address := fmt.Sprintf("%s/p2p/%s", multiAddrStr, pid.String())
		pbDirection := silapb.PeerDirection_UNKNOWN
		switch direction {
		case network.DirInbound:
			pbDirection = silapb.PeerDirection_INBOUND
		case network.DirOutbound:
			pbDirection = silapb.PeerDirection_OUTBOUND
		}
		res = append(res, &silapb.Peer{
			Address:         address,
			Direction:       pbDirection,
			ConnectionState: silapb.ConnectionState_CONNECTED,
			PeerId:          pid.String(),
			Enr:             enr,
		})
	}

	return &silapb.Peers{
		Peers: res,
	}, nil
}

// Deprecated: The gRPC API will remain the default and fully supported through v8 (expected in 2026) but will be eventually removed in favor of REST API.
//
// GetSilaExecutionConnectionStatus gets data about the SILAEXEC endpoints.
func (ns *Server) GetSilaExecutionConnectionStatus(_ context.Context, _ *empty.Empty) (*silapb.SilaExecutionConnectionStatus, error) {
	var currErr string
	err := ns.POWChainInfoFetcher.SilaClientConnectionErr()
	if err != nil {
		currErr = err.Error()
	}
	return &silapb.SilaExecutionConnectionStatus{
		CurrentAddress:         ns.POWChainInfoFetcher.SilaClientEndpoint(),
		CurrentConnectionError: currErr,
		Addresses:              []string{ns.POWChainInfoFetcher.SilaClientEndpoint()},
	}, nil
}

// Deprecated: The gRPC API will remain the default and fully supported through v8 (expected in 2026) but will be eventually removed in favor of REST API.
//
// StreamBeaconLogs from the beacon node via a gRPC server-side stream.
// DEPRECATED: This endpoint doesn't appear to be used and have been marked for deprecation.
func (ns *Server) StreamBeaconLogs(_ *empty.Empty, stream silapb.Health_StreamBeaconLogsServer) error {
	ch := make(chan []byte, ns.StreamLogsBufferSize)
	sub := ns.LogsStreamer.LogsFeed().Subscribe(ch)
	defer func() {
		sub.Unsubscribe()
		close(ch)
	}()

	recentLogs := ns.LogsStreamer.GetLastFewLogs()
	logStrings := make([]string, len(recentLogs))
	for i, log := range recentLogs {
		logStrings[i] = string(log)
	}
	if err := stream.Send(&silapb.LogsResponse{
		Logs: logStrings,
	}); err != nil {
		return status.Errorf(codes.Unavailable, "Could not send over stream: %v", err)
	}
	for {
		select {
		case log := <-ch:
			resp := &silapb.LogsResponse{
				Logs: []string{string(log)},
			}
			if err := stream.Send(resp); err != nil {
				return status.Errorf(codes.Unavailable, "Could not send over stream: %v", err)
			}
		case err := <-sub.Err():
			return status.Errorf(codes.Canceled, "Subscriber error, closing: %v", err)
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "Context canceled")
		}
	}
}
