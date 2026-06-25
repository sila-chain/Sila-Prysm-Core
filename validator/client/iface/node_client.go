package iface

import (
	"context"

	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/golang/protobuf/ptypes/empty"
)

type NodeClient interface {
	SyncStatus(ctx context.Context, in *empty.Empty) (*silapb.SyncStatus, error)
	Genesis(ctx context.Context, in *empty.Empty) (*silapb.Genesis, error)
	Version(ctx context.Context, in *empty.Empty) (*silapb.Version, error)
	Peers(ctx context.Context, in *empty.Empty) (*silapb.Peers, error)
	IsReady(ctx context.Context) bool
}
