// Package beacon defines a gRPC beacon service implementation,
// following the official API standards https://ethereum.github.io/beacon-apis/#/.
// This package includes the beacon and config endpoints.
package beacon

import (
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/blockchain"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/cache"
	blockfeed "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed/block"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed/operation"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/db"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/execution"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/operations/attestations"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/operations/blstoexec"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/operations/payloadattestation"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/operations/slashings"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/operations/voluntaryexits"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/p2p"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/rpc/core"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/rpc/lookup"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/stategen"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/sync"
	eth "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

// Server defines a server implementation of the gRPC Beacon Chain service,
// providing RPC endpoints to access data relevant to the Sila Beacon Chain.
type Server struct {
	BeaconDB                db.ReadOnlyDatabase
	ChainInfoFetcher        blockchain.ChainInfoFetcher
	GenesisTimeFetcher      blockchain.TimeFetcher
	BlockReceiver           blockchain.BlockReceiver
	BlockNotifier           blockfeed.Notifier
	OperationNotifier       operation.Notifier
	Broadcaster             p2p.Broadcaster
	DataColumnReceiver      blockchain.DataColumnReceiver
	AttestationCache        *cache.AttestationCache
	AttestationsPool        attestations.Pool
	SlashingsPool           slashings.PoolManager
	VoluntaryExitsPool      voluntaryexits.PoolManager
	StateGenService         stategen.StateManager
	Stater                  lookup.Stater
	Blocker                 lookup.Blocker
	HeadFetcher             blockchain.HeadFetcher
	TimeFetcher             blockchain.TimeFetcher
	OptimisticModeFetcher   blockchain.OptimisticModeFetcher
	V1Alpha1ValidatorServer eth.BeaconNodeValidatorServer
	SyncChecker             sync.Checker
	CanonicalHistory        *stategen.CanonicalHistory
	ExecutionReconstructor  execution.Reconstructor
	FinalizationFetcher     blockchain.FinalizationFetcher
	BLSChangesPool          blstoexec.PoolManager
	PayloadAttestationPool  payloadattestation.PoolManager
	ForkchoiceFetcher       blockchain.ForkchoiceFetcher
	CoreService             *core.Service
	AttestationStateFetcher blockchain.AttestationStateFetcher
	// ExecutionPayloadEnvelopeCache reconstructs the full envelope in the blinded publish flow.
	ExecutionPayloadEnvelopeCache *cache.ExecutionPayloadEnvelopeCache
}
