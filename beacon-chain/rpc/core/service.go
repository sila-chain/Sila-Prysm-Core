package core

import (
	"sync/atomic"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/blockchain"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/cache"
	opfeed "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed/operation"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/db"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/operations/synccommittee"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/p2p"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/stategen"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/sync"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"golang.org/x/sync/singleflight"
)

type Service struct {
	BeaconDB              db.ReadOnlyDatabase
	ChainInfoFetcher      blockchain.ChainInfoFetcher
	HeadFetcher           blockchain.HeadFetcher
	ForkchoiceFetcher     blockchain.ForkchoiceFetcher
	FinalizedFetcher      blockchain.FinalizationFetcher
	GenesisTimeFetcher    blockchain.TimeFetcher
	SyncChecker           sync.Checker
	Broadcaster           p2p.Broadcaster
	SyncCommitteePool     synccommittee.Pool
	OperationNotifier     opfeed.Notifier
	AttestationCache      *cache.AttestationDataCache
	StateGen              stategen.StateManager
	P2P                   p2p.Broadcaster
	ReplayerBuilder       stategen.ReplayerBuilder
	OptimisticModeFetcher blockchain.OptimisticModeFetcher

	payloadAttestationData   atomic.Pointer[silapb.PayloadAttestationData]
	payloadAttestationFlight singleflight.Group
}
